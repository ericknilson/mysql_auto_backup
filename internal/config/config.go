package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	APIKey  string

	EncryptionKey string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string

	ScheduleTZ           string
	ScheduleCron         string
	MaxConcurrentBackups int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:              getEnv("APP_PORT", "8080"),
		APIKey:               os.Getenv("API_KEY"),
		EncryptionKey:        os.Getenv("ENCRYPTION_KEY"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               getEnv("DB_NAME", "mysql_auto_backup"),
		R2AccountID:          os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:        os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:    os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:             os.Getenv("R2_BUCKET"),
		ScheduleTZ:           getEnv("SCHEDULE_TZ", "America/Sao_Paulo"),
		ScheduleCron:         getEnv("SCHEDULE_CRON", "0 3 * * *"),
		MaxConcurrentBackups: getEnvInt("MAX_CONCURRENT_BACKUPS", 5),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"API_KEY":              c.APIKey,
		"ENCRYPTION_KEY":       c.EncryptionKey,
		"DB_NAME":              c.DBName,
		"R2_ACCOUNT_ID":        c.R2AccountID,
		"R2_ACCESS_KEY_ID":     c.R2AccessKeyID,
		"R2_SECRET_ACCESS_KEY": c.R2SecretAccessKey,
		"R2_BUCKET":            c.R2Bucket,
	}
	for k, v := range required {
		if v == "" {
			return fmt.Errorf("variável de ambiente obrigatória ausente: %s", k)
		}
	}
	if c.MaxConcurrentBackups < 1 {
		c.MaxConcurrentBackups = 1
	}
	return nil
}

func (c *Config) AppDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
