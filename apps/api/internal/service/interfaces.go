package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"lotto-journal/api/internal/models"
)

type UserServiceInterface interface {
	FindOrCreate(lineUserID string) (*models.User, bool, error)
	Deactivate(lineUserID string) error
	Reactivate(lineUserID string) error
	UpdateLanguage(lineUserID string, language string) error
}

type TicketServiceInterface interface {
	SubmitTickets(ownerID uuid.UUID, text string) ([]ParsedTicket, []string, uuid.UUID, error)
	ListTickets(ownerID uuid.UUID) ([]*models.Ticket, time.Time, error)
}

type NotificationServiceInterface interface {
	SendDrawNotifications(ctx context.Context, drawID uuid.UUID, drawDateStr string) error
	LogNotification(userID uuid.UUID, lineUserID string, notifType models.NotificationType, drawID *uuid.UUID, status models.NotificationStatus, errStr *string) error
}

type StorageServiceInterface interface {
	UploadAndResizeImage(ctx context.Context, imageReader io.Reader, storageKey string) (string, error)
}

type OcrServiceInterface interface {
	PerformOCR(ctx context.Context, base64DataURL string) (*OCRResponse, []byte, error)
}

type OcrSessionServiceInterface interface {
	CreateSession(userID uuid.UUID, fileID *uuid.UUID, tickets []ParsedTicket, warnings []string) (*models.OcrSession, error)
	GetPendingSession(userID uuid.UUID) (*models.OcrSession, error)
	ConfirmSession(sessionID uuid.UUID) ([]models.Ticket, error)
	CancelSession(sessionID uuid.UUID) error
	StartEditing(sessionID uuid.UUID, number string) error
	SubmitCorrection(sessionID uuid.UUID, text string) (*models.OcrSession, error)
	CleanExpiredSessions(userID uuid.UUID) error
}
