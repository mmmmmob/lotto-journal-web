package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"lotto-journal/api/internal/models"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(file *models.File) error {
	return r.db.Create(file).Error
}

func (r *FileRepository) FindByID(id uuid.UUID) (*models.File, error) {
	var file models.File
	if err := r.db.First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) Update(file *models.File) error {
	return r.db.Save(file).Error
}
