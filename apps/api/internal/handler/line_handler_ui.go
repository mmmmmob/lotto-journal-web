package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"

	"lotto-journal/api/internal/localization"
	"lotto-journal/api/internal/models"
	"lotto-journal/api/internal/service"
)

func (h *LineHandler) sendOcrConflictWarning(replyToken string, user *models.User, session *models.OcrSession) {
	dict := localization.GetDictionary(user.Language)
	quickReplies := &messaging_api.QuickReply{
		Items: []messaging_api.QuickReplyItem{
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrShowPending,
					Data:        fmt.Sprintf("action=show_pending&session_id=%s", session.ID),
					DisplayText: dict.OcrShowPending,
				},
			},
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrConfirmPending,
					Data:        fmt.Sprintf("action=confirm&session_id=%s", session.ID),
					DisplayText: dict.OcrConfirmPending,
				},
			},
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrCancelPending,
					Data:        fmt.Sprintf("action=cancel&session_id=%s", session.ID),
					DisplayText: dict.OcrCancelPending,
				},
			},
		},
	}
	h.replyAndLogTextWithQuickReplies(replyToken, user, models.NotifTypeHelpAdd, nil, dict.OcrConflictWarning, quickReplies)
}

func (h *LineHandler) sendOcrConfirmation(replyToken string, user *models.User, session *models.OcrSession) {
	var tickets []service.ParsedTicket
	if err := json.Unmarshal(session.Tickets, &tickets); err != nil {
		log.Printf("[ocr_confirm] failed to unmarshal tickets: %v", err)
		h.replyMaintenanceLocalized(replyToken, user.Language)
		return
	}

	var warnings []string
	if err := json.Unmarshal(session.Warnings, &warnings); err != nil {
		log.Printf("[ocr_confirm] failed to unmarshal warnings: %v", err)
	}

	dict := localization.GetDictionary(user.Language)

	var ticketLines []string
	for _, t := range tickets {
		if t.Quantity > 1 {
			ticketLines = append(ticketLines, fmt.Sprintf("  • %s x%d (%s)", t.Number, t.Quantity, t.Type))
		} else {
			ticketLines = append(ticketLines, fmt.Sprintf("  • %s (%s)", t.Number, t.Type))
		}
	}
	ticketsText := strings.Join(ticketLines, "\n")

	var warningLines []string
	for _, w := range warnings {
		warningLines = append(warningLines, "  • "+localization.GetOcrWarningMessage(w, user.Language))
	}
	warningsText := ""
	if len(warningLines) > 0 {
		warningsText = fmt.Sprintf(dict.OcrWarningHeader, strings.Join(warningLines, "\n"))
	}

	messageText := fmt.Sprintf(dict.OcrConfirmHeader, ticketsText) + warningsText

	quickReplies := &messaging_api.QuickReply{
		Items: []messaging_api.QuickReplyItem{
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrConfirmSave,
					Data:        fmt.Sprintf("action=confirm&session_id=%s", session.ID),
					DisplayText: dict.OcrConfirmSave,
				},
			},
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrEdit,
					Data:        fmt.Sprintf("action=edit_list&session_id=%s", session.ID),
					DisplayText: dict.OcrEdit,
				},
			},
			{
				Type: "action",
				Action: &messaging_api.PostbackAction{
					Label:       dict.OcrCancel,
					Data:        fmt.Sprintf("action=cancel&session_id=%s", session.ID),
					DisplayText: dict.OcrCancel,
				},
			},
		},
	}

	h.replyAndLogTextWithQuickReplies(replyToken, user, models.NotifTypeTicketSubmitted, nil, messageText, quickReplies)
}

func (h *LineHandler) replyText(replyToken, text string) {
	if _, err := h.bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []messaging_api.MessageInterface{
			messaging_api.TextMessage{Text: text},
		},
	}); err != nil {
		log.Printf("[reply] error: %v", err)
	}
}

func (h *LineHandler) replyAndLogText(replyToken string, user *models.User, notifType models.NotificationType, drawID *uuid.UUID, text string) {
	h.replyAndLogTextWithQuickReplies(replyToken, user, notifType, drawID, text, nil)
}

func (h *LineHandler) replyAndLogTextWithQuickReplies(replyToken string, user *models.User, notifType models.NotificationType, drawID *uuid.UUID, text string, quickReplies *messaging_api.QuickReply) {
	var errStr *string
	status := models.NotifStatusSuccess

	msg := messaging_api.TextMessage{Text: text}
	if quickReplies != nil {
		msg.QuickReply = quickReplies
	}

	if _, err := h.bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []messaging_api.MessageInterface{
			msg,
		},
	}); err != nil {
		status = models.NotifStatusFailed
		errMsg := err.Error()
		errStr = &errMsg
		log.Printf("[reply] error: %v", err)
	}

	if user == nil {
		log.Printf("[reply] cannot log notification: user is nil")
		return
	}

	if logErr := h.notificationSvc.LogNotification(user.ID, user.LineUserID, notifType, drawID, status, errStr); logErr != nil {
		log.Printf("[reply] failed to write notification log: %v", logErr)
	}
}

func (h *LineHandler) showLoading(chatID string, loadingSeconds int32) {
	if chatID == "" {
		return
	}
	if loadingSeconds < 5 {
		loadingSeconds = 5
	}
	// LINE requires loadingSeconds to be in 5-second increments.
	if rem := loadingSeconds % 5; rem != 0 {
		loadingSeconds += 5 - rem
	}
	if loadingSeconds > 60 {
		loadingSeconds = 60
	}

	if _, err := h.bot.ShowLoadingAnimation(&messaging_api.ShowLoadingAnimationRequest{
		ChatId:         chatID,
		LoadingSeconds: loadingSeconds,
	}); err != nil {
		log.Printf("[loading] error: %v", err)
	}
}

func buildWelcomeMessage(displayName string, lang string, isNew bool) string {
	dict := localization.GetDictionary(lang)
	var greeting string
	if displayName != "" {
		greeting = fmt.Sprintf(dict.WelcomeGreetingPersonal, displayName)
	} else {
		greeting = dict.WelcomeGreetingGeneric
	}

	var template string
	if isNew {
		template = dict.WelcomeFirstTime
	} else {
		template = dict.WelcomeReturning
	}

	return greeting + "\n\n" + template
}

// buildReply constructs the confirmation (or error) text for a ticket submission.
func buildReply(saved []service.ParsedTicket, invalid []string, lang string) string {
	dict := localization.GetDictionary(lang)
	if len(saved) == 0 && len(invalid) == 0 {
		return buildWelcomeMessage("", lang, true)
	}

	if len(saved) == 0 {
		return fmt.Sprintf(dict.SubmitInvalid, strings.Join(invalid, ", "))
	}

	var ticketsListLines []string
	for _, t := range saved {
		if t.Quantity > 1 {
			ticketsListLines = append(ticketsListLines, fmt.Sprintf("  • %s x%d (%s)", t.Number, t.Quantity, t.Type))
		} else {
			ticketsListLines = append(ticketsListLines, fmt.Sprintf("  • %s (%s)", t.Number, t.Type))
		}
	}
	ticketsList := strings.Join(ticketsListLines, "\n")

	if len(invalid) > 0 {
		return fmt.Sprintf(dict.SubmitMixed, ticketsList, strings.Join(invalid, ", "))
	}
	return fmt.Sprintf(dict.SubmitConfirm, ticketsList)
}

func buildTicketListReply(tickets []*models.Ticket, drawDate time.Time, lang string) string {
	dict := localization.GetDictionary(lang)
	if len(tickets) == 0 {
		return dict.ListEmpty
	}

	dateStr := drawDate.Format("02/01/2006")
	lines := []string{fmt.Sprintf(dict.ListHeader, dateStr)}

	for _, t := range tickets {
		if t.Quantity > 1 {
			lines = append(lines, fmt.Sprintf("  • %s x%d (%s)", t.Number, t.Quantity, t.Type))
		} else {
			lines = append(lines, fmt.Sprintf("  • %s (%s)", t.Number, t.Type))
		}
	}
	return strings.Join(lines, "\n")
}

func (h *LineHandler) replyMaintenance(replyToken string) {
	h.replyText(replyToken, localization.DbMaintenanceMessage)
}

func (h *LineHandler) replyMaintenanceLocalized(replyToken string, lang string) {
	dict := localization.GetDictionary(lang)
	h.replyText(replyToken, dict.DbMaintenance)
}
