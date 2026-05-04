package database

import (
	"github.com/erick_nilson/mysql_auto_backup/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.AppDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}
