package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erick_nilson/mysql_auto_backup/internal/services"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func New(tz, expr string, backupSvc *services.BackupService) (*Scheduler, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("timezone %q inválido: %w", tz, err)
	}
	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc(expr, func() {
		log.Printf("[scheduler] tick em %s — iniciando backup completo", time.Now().In(loc).Format(time.RFC3339))
		backupSvc.RunForAll(context.Background())
	}); err != nil {
		return nil, fmt.Errorf("registrar cron %q: %w", expr, err)
	}
	return &Scheduler{cron: c}, nil
}

func (s *Scheduler) Start() { s.cron.Start() }

func (s *Scheduler) Stop() { s.cron.Stop() }
