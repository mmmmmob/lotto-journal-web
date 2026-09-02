package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"lotto-journal/api/internal/models"
)

type UserRepositoryInterface interface {
	Create(user *models.User) error
	FindByLineUserID(lineUserID string) (*models.User, error)
	FindOrCreate(lineUserID string) (*models.User, bool, error)
	UpdateStatus(lineUserID string, status string) error
}

type TicketRepositoryInterface interface {
	Create(ticket *models.Ticket) error
	List(drawID uuid.UUID, userID uuid.UUID) ([]*models.Ticket, error)
	FindUnchecked(drawID uuid.UUID) ([]*models.Ticket, error)
	FindUncheckedInTransaction(tx *gorm.DB, drawID uuid.UUID) ([]*models.Ticket, error)
	MarkCheckedInTransaction(tx *gorm.DB, ticketIDs []uuid.UUID) error
	ResetCheckedStatusByDrawIDInTransaction(tx *gorm.DB, drawID uuid.UUID) error
	FindDrawTicketsWithOwners(drawID uuid.UUID) ([]DrawTicketWithOwner, error)
}

type DrawRepositoryInterface interface {
	FindNextDraw(fromDate time.Time) (*models.Draw, error)
	FindByDate(date time.Time) (*models.Draw, error)
	FindOrCreate(date time.Time) (*models.Draw, error)
	FindLatestUnverified(date time.Time) (*models.Draw, error)
	MarkVerifiedInTransaction(tx *gorm.DB, drawID uuid.UUID) error
}

type DrawResultRepositoryInterface interface {
	CreateInBatches(results []*models.DrawResult) error
	CreateInBatchesInTransaction(tx *gorm.DB, results []*models.DrawResult) error
	DeleteByDrawIDInTransaction(tx *gorm.DB, drawID uuid.UUID) error
	FindSpecialResultByDrawID(drawID uuid.UUID) (*models.DrawResult, error)
}

type UserWinningRepositoryInterface interface {
	CreateInBatches(winnings []*models.UserWinning) error
	CreateInBatchesInTransaction(tx *gorm.DB, winnings []*models.UserWinning) error
	DeleteByDrawIDInTransaction(tx *gorm.DB, drawID uuid.UUID) error
	FindDrawWinnings(drawID uuid.UUID) ([]DrawWinningDetail, error)
}

type FileRepositoryInterface interface {
	Create(file *models.File) error
	FindByID(id uuid.UUID) (*models.File, error)
	Update(file *models.File) error
}

type OcrSessionRepositoryInterface interface {
	Create(session *models.OcrSession) error
	FindByID(id uuid.UUID) (*models.OcrSession, error)
	FindPendingByUserID(userID uuid.UUID) (*models.OcrSession, error)
	Update(session *models.OcrSession) error
	DeleteExpiredPendingByUserID(userID uuid.UUID) error
}
