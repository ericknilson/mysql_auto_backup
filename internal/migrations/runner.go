package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/erick_nilson/mysql_auto_backup/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var migrationFS embed.FS

func Run(cfg *config.Config) error {
	src, err := iofs.New(migrationFS, "sql")
	if err != nil {
		return fmt.Errorf("init iofs source: %w", err)
	}

	dsn := fmt.Sprintf(
		"mysql://%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
