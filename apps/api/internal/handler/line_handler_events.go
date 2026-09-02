package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"lotto-journal/api/internal/localization"
	"lotto-journal/api/internal/models"
	"lotto-journal/api/internal/service"
)

// dispatch deduplicates and routes a single LINE event to the appropriate handler.
func (h *LineHandler) dispatch(event webhook.EventInterface) {
	eventID := webhookEventID(event)
	if eventID == "" {
		// webhookEventID only extracts IDs for event types we support.
		// Anything else (postback, join, leave, …) is unsupported and skipped.
		log.Printf("[webhook] unsupported event type %T — skipping", event)
		return
	}

	isNew, err := h.webhookRepo.MarkProcessed(eventID)
	if err != nil {
		log.Printf("[webhook] MarkProcessed %s: %v", eventID, err)
		if replyToken := eventReplyToken(event); replyToken != "" {
			h.replyMaintenance(replyToken)
		}
		return
	}
	if !isNew {
		log.Printf("[webhook] duplicate event %s — skipping", eventID)
		return
	}

	switch e := event.(type) {
	case webhook.FollowEvent:
		h.handleFollow(e)
	case webhook.UnfollowEvent:
		h.handleUnfollow(e)
	case webhook.MessageEvent:
		h.handleMessage(e)
	case webhook.PostbackEvent:
		h.handlePostback(e)
		// No default: unsupported types never reach here — webhookEventID returns "" for them.
	}
}

func (h *LineHandler) handleFollow(e webhook.FollowEvent) {
	lineUserID := sourceUserID(e.Source)
	if lineUserID == "" {
		log.Println("[follow] no userId in source")
		return
	}

	user, isNew, err := h.userSvc.FindOrCreate(lineUserID)
	if err != nil {
		log.Printf("[follow] FindOrCreate %s: %v", lineUserID, err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	var displayName string
	var detectedLanguage string

	profile, err := h.bot.GetProfile(lineUserID)
	if err == nil && profile != nil {
		displayName = strings.TrimSpace(profile.DisplayName)
		detectedLanguage = strings.TrimSpace(profile.Language)
	} else {
		log.Printf("[profile] GetProfile %s: %v", lineUserID, err)
	}

	if isNew {
		log.Printf("[follow] new user created: %s", lineUserID)
		lang := "en"
		if strings.ToLower(detectedLanguage) == "th" {
			lang = "th"
		}
		user.Language = lang
		if err := h.userSvc.UpdateLanguage(lineUserID, lang); err != nil {
			log.Printf("[follow] UpdateLanguage %s to %s: %v", lineUserID, lang, err)
		}
	} else {
		// User was previously inactive (unfollowed) — restore active status.
		if err := h.userSvc.Reactivate(lineUserID); err != nil {
			log.Printf("[follow] Reactivate %s: %v", lineUserID, err)
		}
		log.Printf("[follow] existing user re-followed: %s", lineUserID)
	}

	h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeWelcome, nil, buildWelcomeMessage(displayName, user.Language, isNew), localization.GetQuickReplies(user.Language))
}

func (h *LineHandler) handleUnfollow(e webhook.UnfollowEvent) {
	lineUserID := sourceUserID(e.Source)
	if lineUserID == "" {
		log.Println("[unfollow] no userId in source")
		return
	}

	if err := h.userSvc.Deactivate(lineUserID); err != nil {
		log.Printf("[unfollow] Deactivate %s: %v", lineUserID, err)
		return
	}
	log.Printf("[unfollow] user marked inactive: %s", lineUserID)
	// No reply — LINE does not allow replying to unfollow events.
}

func (h *LineHandler) handleMessage(e webhook.MessageEvent) {
	switch msg := e.Message.(type) {
	case webhook.TextMessageContent:
		h.handleTextMessage(e, msg)
	case webhook.ImageMessageContent:
		h.handleImageMessage(e, msg)
	default:
		log.Printf("[message] unsupported content type %T — ignoring", e.Message)
	}
}

func (h *LineHandler) handleTextMessage(e webhook.MessageEvent, textMsg webhook.TextMessageContent) {
	lineUserID := sourceUserID(e.Source)
	if lineUserID == "" {
		log.Println("[message] no userId in source")
		return
	}

	// Ensure the user record exists (edge case: message before follow event).
	user, isNew, err := h.userSvc.FindOrCreate(lineUserID)
	if err != nil {
		log.Printf("[message] FindOrCreate %s: %v", lineUserID, err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	if isNew {
		log.Printf("[message] new user created: %s", lineUserID)
		lang := "en"
		profile, err := h.bot.GetProfile(lineUserID)
		if err == nil && profile != nil {
			if strings.ToLower(profile.Language) == "th" {
				lang = "th"
			}
		}
		user.Language = lang
		if err := h.userSvc.UpdateLanguage(lineUserID, lang); err != nil {
			log.Printf("[message] UpdateLanguage %s to %s: %v", lineUserID, lang, err)
		}
	}

	msgText := textMsg.Text

	// Check if user has an active editing OCR session
	pendingSession, err := h.ocrSessionSvc.GetPendingSession(user.ID)
	if err == nil && pendingSession != nil && pendingSession.CurrentEditNumber != nil && *pendingSession.CurrentEditNumber != "" {
		// Active editing session! User sent a text correction.
		updatedSession, err := h.ocrSessionSvc.SubmitCorrection(pendingSession.ID, msgText)
		if err != nil {
			// Correction was invalid (e.g. invalid format)
			h.replyText(e.ReplyToken, err.Error())
			return
		}
		// Successfully corrected! Reply with updated list.
		h.sendOcrConfirmation(e.ReplyToken, user, updatedSession)
		return
	}

	// Best-effort loading indicator to show user we're processing the request.
	// Delay the indicator slightly and cancel it for fast paths so users don't
	// see a 5-second spinner for quick replies.
	processingDone := make(chan struct{})
	defer close(processingDone)
	go func(chatID string, done <-chan struct{}) {
		timer := time.NewTimer(700 * time.Millisecond)
		defer timer.Stop()

		select {
		case <-done:
			return
		case <-timer.C:
			h.showLoading(chatID, 5)
		}
	}(lineUserID, processingDone)

	if isThaiSwitchCmd(msgText) {
		if err := h.userSvc.UpdateLanguage(lineUserID, "th"); err != nil {
			log.Printf("[message] UpdateLanguage to th for %s: %v", lineUserID, err)
			h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
			return
		}
		user.Language = "th"
		dict := localization.GetDictionary("th")
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeLanguageChanged, nil, dict.LanguageSwitched, localization.GetQuickReplies("th"))
		return
	}

	if isEnglishSwitchCmd(msgText) {
		if err := h.userSvc.UpdateLanguage(lineUserID, "en"); err != nil {
			log.Printf("[message] UpdateLanguage to en for %s: %v", lineUserID, err)
			h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
			return
		}
		user.Language = "en"
		dict := localization.GetDictionary("en")
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeLanguageChanged, nil, dict.LanguageSwitched, localization.GetQuickReplies("en"))
		return
	}

	if isTicketListCmd(msgText) {
		userTickets, drawDate, err := h.ticketSvc.ListTickets(user.ID)
		if err != nil {
			log.Printf("[message] error retrieving %s tickets: %v", user.ID, err)
			h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
			return
		}
		var drawIDPtr *uuid.UUID
		if len(userTickets) > 0 {
			drawIDPtr = &userTickets[0].DrawID
		}
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketList, drawIDPtr, buildTicketListReply(userTickets, drawDate, user.Language), localization.GetQuickReplies(user.Language))
		return
	}

	if isAddHelpCmd(msgText) {
		dict := localization.GetDictionary(user.Language)
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeHelpAdd, nil, dict.AddHelp, localization.GetQuickReplies(user.Language))
		return
	}

	if isNotifyHelpCmd(msgText) {
		dict := localization.GetDictionary(user.Language)
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeHelpNotify, nil, dict.NotifyHelp, localization.GetQuickReplies(user.Language))
		return
	}

	// Default to parsing tickets
	saved, invalid, drawID, err := h.ticketSvc.SubmitTickets(user.ID, msgText)
	if err != nil {
		log.Printf("[message] SubmitTickets for %s: %v", lineUserID, err)
		h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
		return
	}
	var drawIDPtr *uuid.UUID
	if drawID != uuid.Nil {
		drawIDPtr = &drawID
	}
	h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketSubmitted, drawIDPtr, buildReply(saved, invalid, user.Language), localization.GetQuickReplies(user.Language))
}

func (h *LineHandler) handleImageMessage(e webhook.MessageEvent, imgMsg webhook.ImageMessageContent) {
	lineUserID := sourceUserID(e.Source)
	if lineUserID == "" {
		log.Println("[image] no userId in source")
		return
	}

	user, _, err := h.userSvc.FindOrCreate(lineUserID)
	if err != nil {
		log.Printf("[image] FindOrCreate %s: %v", lineUserID, err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	// Clean up expired pending sessions first
	if err := h.ocrSessionSvc.CleanExpiredSessions(user.ID); err != nil {
		log.Printf("[image] CleanExpiredSessions: %v", err)
	}

	// Check if user already has an active pending OCR session
	pendingSession, err := h.ocrSessionSvc.GetPendingSession(user.ID)
	if err != nil {
		log.Printf("[image] GetPendingSession %s: %v", user.ID, err)
	}
	if pendingSession != nil {
		h.sendOcrConflictWarning(e.ReplyToken, user, pendingSession)
		return
	}

	// Show loading animation (10 seconds)
	h.showLoading(lineUserID, 10)

	// Download image binary from LINE Content API
	resp, err := h.blobBot.GetMessageContent(imgMsg.Id)
	if err != nil {
		log.Printf("[image] GetMessageContent %s: %v", imgMsg.Id, err)
		h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
		return
	}
	defer resp.Body.Close()

	// Resize, encode, upload to R2, and get base64 Data URL
	fileID := uuid.New()
	storageKey := fmt.Sprintf("tickets/%s/%s.jpg", user.ID, fileID)

	base64DataURL, err := h.storageSvc.UploadAndResizeImage(context.Background(), resp.Body, storageKey)
	if err != nil {
		log.Printf("[image] UploadAndResizeImage: %v", err)
		h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
		return
	}

	// Create File record in database
	fileRecord := &models.File{
		ID:               fileID,
		OwnerID:          &user.ID,
		StorageKey:       storageKey,
		FileType:         ptr("image/jpeg"),
		ExtractionStatus: ptr("pending"),
	}
	if err := h.fileRepo.Create(fileRecord); err != nil {
		log.Printf("[image] CreateFileRecord %s: %v", fileID, err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	// Run OpenAI OCR
	ocrResult, rawJSONResponse, err := h.ocrSvc.PerformOCR(context.Background(), base64DataURL)
	if err != nil {
		log.Printf("[image] PerformOCR: %v", err)

		fileRecord.ExtractionStatus = ptr("failure")
		if rawJSONResponse != nil {
			fileRecord.ExtractedResponse = rawJSONResponse
		}
		fileRecord.ExtractionModel = ptr("gpt-4o-mini")
		_ = h.fileRepo.Update(fileRecord)

		h.replyMaintenanceLocalized(e.ReplyToken, user.Language)
		return
	}

	fileRecord.ExtractionStatus = ptr("success")
	fileRecord.ExtractedResponse = rawJSONResponse
	fileRecord.ExtractionModel = ptr("gpt-4o-mini")
	_ = h.fileRepo.Update(fileRecord)

	if len(ocrResult.Tickets) == 0 {
		h.replyText(e.ReplyToken, localization.GetDictionary(user.Language).SubmitInvalid)
		return
	}

	// Create OCR session in DB
	var parsedTickets []service.ParsedTicket
	for _, t := range ocrResult.Tickets {
		ticketType := "L6"
		if len(t.Number) == 3 {
			ticketType = "N3"
		}
		parsedTickets = append(parsedTickets, service.ParsedTicket{
			Number:   t.Number,
			Quantity: t.Quantity,
			Type:     ticketType,
		})
	}

	session, err := h.ocrSessionSvc.CreateSession(user.ID, &fileID, parsedTickets, ocrResult.Warnings)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Duplicate session occurred right after our check
			existing, getErr := h.ocrSessionSvc.GetPendingSession(user.ID)
			if getErr == nil && existing != nil {
				h.sendOcrConflictWarning(e.ReplyToken, user, existing)
				return
			}
		}
		log.Printf("[image] CreateSession: %v", err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	h.sendOcrConfirmation(e.ReplyToken, user, session)
}

func (h *LineHandler) handlePostback(e webhook.PostbackEvent) {
	lineUserID := sourceUserID(e.Source)
	if lineUserID == "" {
		log.Println("[postback] no userId in source")
		return
	}

	user, _, err := h.userSvc.FindOrCreate(lineUserID)
	if err != nil {
		log.Printf("[postback] FindOrCreate %s: %v", lineUserID, err)
		h.replyMaintenance(e.ReplyToken)
		return
	}

	data := e.Postback.Data
	params := parseQueryParams(data)
	action := params["action"]
	sessionIDStr := params["session_id"]
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		log.Printf("[postback] invalid session_id %s: %v", sessionIDStr, err)
		h.replyText(e.ReplyToken, localization.GetDictionary(user.Language).GenericError)
		return
	}

	dict := localization.GetDictionary(user.Language)

	switch action {
	case "confirm":
		saved, err := h.ocrSessionSvc.ConfirmSession(sessionID)
		if err != nil {
			log.Printf("[postback] ConfirmSession: %v", err)
			h.replyText(e.ReplyToken, err.Error())
			return
		}

		var ticketLines []string
		for _, t := range saved {
			if t.Quantity > 1 {
				ticketLines = append(ticketLines, fmt.Sprintf("  • %s x%d (%s)", t.Number, t.Quantity, t.Type))
			} else {
				ticketLines = append(ticketLines, fmt.Sprintf("  • %s (%s)", t.Number, t.Type))
			}
		}
		ticketsList := strings.Join(ticketLines, "\n")

		var drawIDPtr *uuid.UUID
		if len(saved) > 0 {
			drawIDPtr = &saved[0].DrawID
		}

		confirmMsg := fmt.Sprintf(dict.SubmitConfirm, ticketsList)
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketSubmitted, drawIDPtr, confirmMsg, localization.GetQuickReplies(user.Language))

	case "cancel":
		if err := h.ocrSessionSvc.CancelSession(sessionID); err != nil {
			log.Printf("[postback] CancelSession: %v", err)
		}
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketSubmitted, nil, dict.OcrCancelled, localization.GetQuickReplies(user.Language))

	case "show_pending":
		session, err := h.ocrSessionSvc.GetPendingSession(user.ID)
		if err != nil || session == nil {
			h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketSubmitted, nil, dict.OcrCancelled, localization.GetQuickReplies(user.Language))
			return
		}
		h.sendOcrConfirmation(e.ReplyToken, user, session)

	case "edit_list":
		session, err := h.ocrSessionSvc.GetPendingSession(user.ID)
		if err != nil || session == nil {
			h.replyText(e.ReplyToken, dict.GenericError)
			return
		}
		var tickets []service.ParsedTicket
		if err := json.Unmarshal(session.Tickets, &tickets); err != nil {
			h.replyText(e.ReplyToken, dict.GenericError)
			return
		}

		var quickReplyItems []messaging_api.QuickReplyItem
		for _, t := range tickets {
			label := fmt.Sprintf("%s x%d", t.Number, t.Quantity)
			quickReplyItems = append(quickReplyItems, messaging_api.QuickReplyItem{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       label,
					Data:        fmt.Sprintf("action=edit_select&session_id=%s&number=%s", session.ID, t.Number),
					DisplayText: fmt.Sprintf("Edit %s", label),
				},
			})
		}
		quickReplyItems = append(quickReplyItems, messaging_api.QuickReplyItem{
			Type: "action",
			Action: &messaging_api.PostbackAction{
				Label:       dict.OcrBack,
				Data:        fmt.Sprintf("action=show_pending&session_id=%s", session.ID),
				DisplayText: dict.OcrBack,
			},
		})

		quickReplies := &messaging_api.QuickReply{Items: quickReplyItems}
		h.replyAndLogTextWithQuickReplies(e.ReplyToken, user, models.NotifTypeTicketSubmitted, nil, dict.OcrEditInstruction, quickReplies)

	case "edit_select":
		number := params["number"]
		if err := h.ocrSessionSvc.StartEditing(sessionID, number); err != nil {
			log.Printf("[postback] StartEditing: %v", err)
			h.replyText(e.ReplyToken, dict.GenericError)
			return
		}
		promptMsg := fmt.Sprintf(dict.OcrPromptCorrection, number)
		h.replyText(e.ReplyToken, promptMsg)
	}
}
