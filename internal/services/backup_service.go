package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"github.com/erick_nilson/mysql_auto_backup/internal/repository"
)

type BackupService struct {
	connSvc        *ConnectionService
	dump           *DumpService
	r2             *R2Service
	logRepo        *repository.BackupLogRepository
	maxConcurrency int
}

func NewBackupService(
	connSvc *ConnectionService,
	dump *DumpService,
	r2 *R2Service,
	logRepo *repository.BackupLogRepository,
	maxConcurrency int,
) *BackupService {
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	return &BackupService{
		connSvc: connSvc, dump: dump, r2: r2,
		logRepo: logRepo, maxConcurrency: maxConcurrency,
	}
}

// RunForAll executa o backup de todas as conexões ativas.
func (s *BackupService) RunForAll(ctx context.Context) {
	conns, err := s.connSvc.ListActiveDecrypted()
	if err != nil {
		log.Printf("[backup] erro listando conexões ativas: %v", err)
		return
	}
	s.runBatch(ctx, conns)
}

// RunForIDs executa o backup somente das conexões ativas dentre os IDs informados.
func (s *BackupService) RunForIDs(ctx context.Context, ids []uint64) {
	conns, err := s.connSvc.FindActiveDecryptedByIDs(ids)
	if err != nil {
		log.Printf("[backup] erro buscando conexões: %v", err)
		return
	}
	s.runBatch(ctx, conns)
}

func (s *BackupService) runBatch(ctx context.Context, conns []DecryptedConnection) {
	if len(conns) == 0 {
		return
	}
	sem := make(chan struct{}, s.maxConcurrency)
	var wg sync.WaitGroup
	for _, c := range conns {
		c := c
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			s.backupOne(ctx, c)
		}()
	}
	wg.Wait()
}

func (s *BackupService) backupOne(ctx context.Context, c DecryptedConnection) {
	sentAt := time.Now().UTC()
	objectKey, err := s.execute(ctx, c)
	logEntry := &models.BackupLog{
		ConnectionID: c.ID,
		FileName:     objectKey,
		SentAt:       sentAt,
	}
	if err != nil {
		log.Printf("[backup] conexão %d (%s) falhou: %v", c.ID, c.Name, err)
		logEntry.Status = models.BackupStatusFailed
		logEntry.ErrorMessage = err.Error()
	} else {
		logEntry.Status = models.BackupStatusSuccess
		log.Printf("[backup] conexão %d (%s) concluída: %s", c.ID, c.Name, objectKey)
	}
	if err := s.logRepo.Create(logEntry); err != nil {
		log.Printf("[backup] não foi possível gravar log para %d: %v", c.ID, err)
	}
}

func (s *BackupService) execute(ctx context.Context, c DecryptedConnection) (string, error) {
	dumpPath, err := s.dump.Dump(c.Name, c.Creds)
	if err != nil {
		return "", err
	}
	defer os.Remove(dumpPath)

	f, err := os.Open(dumpPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	objectKey := fmt.Sprintf("backups/%s/%s", sanitize(c.Name), filepath.Base(dumpPath))
	if err := s.r2.Upload(ctx, objectKey, f); err != nil {
		return "", err
	}
	return objectKey, nil
}
