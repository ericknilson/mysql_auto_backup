package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/erick_nilson/mysql_auto_backup/internal/config"
	"github.com/erick_nilson/mysql_auto_backup/internal/crypto"
	"github.com/erick_nilson/mysql_auto_backup/internal/database"
	"github.com/erick_nilson/mysql_auto_backup/internal/handlers"
	"github.com/erick_nilson/mysql_auto_backup/internal/migrations"
	"github.com/erick_nilson/mysql_auto_backup/internal/repository"
	"github.com/erick_nilson/mysql_auto_backup/internal/router"
	"github.com/erick_nilson/mysql_auto_backup/internal/scheduler"
	"github.com/erick_nilson/mysql_auto_backup/internal/services"

	_ "github.com/erick_nilson/mysql_auto_backup/docs"
)

//	@title			MySQL Auto Backup API
//	@version		1.0
//	@description	Sistema de backup automático de bancos MySQL para Cloudflare R2.
//	@description	Permite cadastrar conexões (com credenciais criptografadas em repouso via AES-256-GCM), agenda backups diários às 03:00 (America/Sao_Paulo) e expõe disparo manual + listagem de logs.
//	@description	Todos os endpoints sob /api requerem o header `X-API-Key`.
//
//	@contact.name	Erick Nilson
//
//	@host		localhost:8080
//	@BasePath	/
//	@schemes	http https
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-API-Key
//	@description				Chave de API estática definida em API_KEY no .env do servidor.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := migrations.Run(cfg); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Printf("[migrations] aplicadas com sucesso")

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	cipher, err := crypto.New(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}

	connRepo := repository.NewConnectionRepository(db)
	logRepo := repository.NewBackupLogRepository(db)

	connSvc := services.NewConnectionService(connRepo, cipher)
	dumpSvc := services.NewDumpService()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r2Svc, err := services.NewR2Service(ctx, cfg)
	if err != nil {
		log.Fatalf("r2: %v", err)
	}

	backupSvc := services.NewBackupService(connSvc, dumpSvc, r2Svc, logRepo, cfg.MaxConcurrentBackups)

	sched, err := scheduler.New(cfg.ScheduleTZ, cfg.ScheduleCron, backupSvc)
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	sched.Start()
	defer sched.Stop()
	log.Printf("[scheduler] agendado: cron=%q tz=%q", cfg.ScheduleCron, cfg.ScheduleTZ)

	connHandler := handlers.NewConnectionHandler(connSvc)
	backupHandler := handlers.NewBackupHandler(backupSvc, logRepo)
	engine := router.New(cfg.APIKey, connHandler, backupHandler)

	addr := ":" + cfg.AppPort
	go func() {
		log.Printf("[http] ouvindo em %s", addr)
		if err := engine.Run(addr); err != nil {
			log.Fatalf("http: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("encerrando...")
}
