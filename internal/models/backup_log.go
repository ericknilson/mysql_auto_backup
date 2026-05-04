package models

import "time"

const (
	BackupStatusSuccess = "success"
	BackupStatusFailed  = "failed"
)

type BackupLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"     json:"id"`
	ConnectionID uint64    `gorm:"not null;index"               json:"connection_id"`
	FileName     string    `gorm:"size:512;not null"            json:"file_name"`
	Status       string    `gorm:"size:20;not null"             json:"status"`
	ErrorMessage string    `gorm:"type:text"                    json:"error_message,omitempty"`
	SentAt       time.Time `gorm:"not null;index"               json:"sent_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Connection *Connection `gorm:"foreignKey:ConnectionID" json:"connection,omitempty"`
}

func (BackupLog) TableName() string { return "backup_logs" }
