package repository

import (
	"time"

	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"gorm.io/gorm"
)

type BackupLogRepository struct {
	db *gorm.DB
}

func NewBackupLogRepository(db *gorm.DB) *BackupLogRepository {
	return &BackupLogRepository{db: db}
}

func (r *BackupLogRepository) Create(l *models.BackupLog) error {
	return r.db.Create(l).Error
}

type ListFilter struct {
	ConnectionID *uint64
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PerPage      int
}

func (r *BackupLogRepository) List(f ListFilter) ([]models.BackupLog, int64, error) {
	q := r.db.Model(&models.BackupLog{})
	if f.ConnectionID != nil {
		q = q.Where("connection_id = ?", *f.ConnectionID)
	}
	if f.StartDate != nil {
		q = q.Where("sent_at >= ?", *f.StartDate)
	}
	if f.EndDate != nil {
		q = q.Where("sent_at <= ?", *f.EndDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.PerPage <= 0 {
		f.PerPage = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PerPage

	var out []models.BackupLog
	if err := q.Order("sent_at DESC").Limit(f.PerPage).Offset(offset).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
