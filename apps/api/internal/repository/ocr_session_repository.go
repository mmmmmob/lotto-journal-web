package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"lotto-journal/api/internal/models"
)

type OcrSessionRepository struct {
	db *gorm.DB
}

func NewOcrSessionRepository(db *gorm.DB) *OcrSessionRepository {
	return &OcrSessionRepository{db: db}
}

func (r *OcrSessionRepository) Create(session *models.OcrSession) error {
	return r.db.Create(session).Error
}

func (r *OcrSessionRepository) FindByID(id uuid.UUID) (*models.OcrSession, error) {
	var session models.OcrSession
	if err := r.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *OcrSessionRepository) FindPendingByUserID(userID uuid.UUID) (*models.OcrSession, error) {
	var session models.OcrSession
	err := r.db.Where("user_id = ? AND status = 'pending' AND expires_at > ?", userID, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *OcrSessionRepository) Update(session *models.OcrSession) error {
	return r.db.Save(session).Error
}

func (r *OcrSessionRepository) DeleteExpiredPendingByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ? AND status = 'pending' AND expires_at < ?", userID, time.Now()).Delete(&models.OcrSession{}).Error
}
