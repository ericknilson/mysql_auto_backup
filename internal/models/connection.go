package models

import (
	"time"

	"gorm.io/gorm"
)

type Connection struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:255;not null"        json:"name"`
	Host      string         `gorm:"type:text;not null"       json:"-"`
	Port      string         `gorm:"type:text;not null"       json:"-"`
	User      string         `gorm:"column:user;type:text;not null"     json:"-"`
	Password  string         `gorm:"type:text;not null"       json:"-"`
	Database  string         `gorm:"column:database;type:text;not null" json:"-"`
	IsActive  bool           `gorm:"not null;default:true"    json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Connection) TableName() string { return "connections" }
