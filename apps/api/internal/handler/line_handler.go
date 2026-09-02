package handler

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"lotto-journal/api/internal/repository"
	"lotto-journal/api/internal/service"
)

// LineHandler handles all LINE Messaging API webhook events.
type LineHandler struct {
	channelSecret   string
	bot             *messaging_api.MessagingApiAPI
	blobBot         *messaging_api.MessagingApiBlobAPI
	userSvc         service.UserServiceInterface
	ticketSvc       service.TicketServiceInterface
	notificationSvc service.NotificationServiceInterface
	webhookRepo     *repository.WebhookEventRepository
	storageSvc      service.StorageServiceInterface
	ocrSvc          service.OcrServiceInterface
	ocrSessionSvc   service.OcrSessionServiceInterface
	fileRepo        *repository.FileRepository
}

func NewLineHandler(
	channelSecret string,
	bot *messaging_api.MessagingApiAPI,
	blobBot *messaging_api.MessagingApiBlobAPI,
	userSvc service.UserServiceInterface,
	ticketSvc service.TicketServiceInterface,
	notificationSvc service.NotificationServiceInterface,
	webhookRepo *repository.WebhookEventRepository,
	storageSvc service.StorageServiceInterface,
	ocrSvc service.OcrServiceInterface,
	ocrSessionSvc service.OcrSessionServiceInterface,
	fileRepo *repository.FileRepository,
) *LineHandler {
	return &LineHandler{
		channelSecret:   channelSecret,
		bot:             bot,
		blobBot:         blobBot,
		userSvc:         userSvc,
		ticketSvc:       ticketSvc,
		notificationSvc: notificationSvc,
		webhookRepo:     webhookRepo,
		storageSvc:      storageSvc,
		ocrSvc:          ocrSvc,
		ocrSessionSvc:   ocrSessionSvc,
		fileRepo:        fileRepo,
	}
}

// Handle is the Fiber route handler for POST /webhook.
//
// @Summary LINE Webhook Receiver
// @Description Handles incoming LINE Messaging API events (follow, unfollow, message).
// @Tags LINE
// @Accept json
// @Produce json
// @Param X-Line-Signature header string true "Signature header for LINE payload validation"
// @Param body body string true "LINE Event Payload"
// @Success 200 "OK"
// @Failure 400 "Invalid signature or request payload"
// @Failure 500 "Internal Server Error"
// @Router /webhook [post]
func (h *LineHandler) Handle(c fiber.Ctx) error {
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"X-Line-Signature": []string{c.Get("X-Line-Signature")},
			"Content-Type":     []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(c.Body())),
	}

	cb, err := webhook.ParseRequest(h.channelSecret, req)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			log.Println("[webhook] invalid signature")
			return c.Status(fiber.StatusBadRequest).SendString("invalid signature")
		}
		log.Printf("[webhook] parse error: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("failed to parse request")
	}

	for _, event := range cb.Events {
		h.dispatch(event)
	}

	return c.SendStatus(fiber.StatusOK)
}
