package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"github.com/erick_nilson/mysql_auto_backup/internal/repository"
	"github.com/erick_nilson/mysql_auto_backup/internal/services"
	"github.com/gin-gonic/gin"
)

// _modelsRef garante que o swag (parser por arquivo) enxergue o pacote models
// para resolver referências como models.BackupLog nas anotações.
var _ = models.BackupLog{}

type BackupHandler struct {
	backupSvc *services.BackupService
	logRepo   *repository.BackupLogRepository
}

func NewBackupHandler(backupSvc *services.BackupService, logRepo *repository.BackupLogRepository) *BackupHandler {
	return &BackupHandler{backupSvc: backupSvc, logRepo: logRepo}
}

// Run godoc
//
//	@Summary		Dispara backup imediato
//	@Description	Inicia o backup em background e responde 202 imediatamente. O resultado individual é gravado em backup_logs e pode ser consultado em GET /api/backups. Se connection_ids for omitido ou vazio, executa para todas as conexões ativas. Caso contrário, executa apenas para os IDs informados que estiverem com is_active = true.
//	@Tags			backups
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			payload	body		RunBackupRequest	false	"IDs específicos (opcional)"
//	@Success		202		{object}	RunBackupResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/api/backups/run [post]
func (h *BackupHandler) Run(c *gin.Context) {
	var req RunBackupRequest
	// Body é opcional. Ignoramos erro de bind quando vazio.
	_ = c.ShouldBindJSON(&req)

	ctx := context.Background()
	if len(req.ConnectionIDs) > 0 {
		ids := req.ConnectionIDs
		go h.backupSvc.RunForIDs(ctx, ids)
		c.JSON(http.StatusAccepted, RunBackupResponse{
			Message:       "backup disparado para os IDs informados (somente os ativos serão processados)",
			ConnectionIDs: ids,
		})
		return
	}
	go h.backupSvc.RunForAll(ctx)
	c.JSON(http.StatusAccepted, RunBackupResponse{
		Message: "backup disparado para todas as conexões ativas",
	})
}

// List godoc
//
//	@Summary		Lista logs de backup
//	@Description	Retorna o histórico de execuções (sucesso e falha) com paginação. Aceita filtros por conexão e por intervalo de data (com base em sent_at).
//	@Tags			backups
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			connection_id	query		int		false	"Filtrar por ID de conexão"
//	@Param			start_date		query		string	false	"Data inicial inclusiva (YYYY-MM-DD)"	example(2026-05-01)
//	@Param			end_date		query		string	false	"Data final inclusiva (YYYY-MM-DD)"		example(2026-05-31)
//	@Param			page			query		int		false	"Página (default 1)"					default(1)
//	@Param			per_page		query		int		false	"Itens por página (default 50)"			default(50)
//	@Success		200				{object}	ListBackupsResponse
//	@Failure		400				{object}	ErrorResponse
//	@Failure		401				{object}	ErrorResponse
//	@Failure		500				{object}	ErrorResponse
//	@Router			/api/backups [get]
func (h *BackupHandler) List(c *gin.Context) {
	filter := repository.ListFilter{}

	if v := c.Query("connection_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "connection_id inválido"})
			return
		}
		filter.ConnectionID = &id
	}
	if v := c.Query("start_date"); v != "" {
		t, err := parseDate(v, false)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "start_date inválido (formato YYYY-MM-DD)"})
			return
		}
		filter.StartDate = &t
	}
	if v := c.Query("end_date"); v != "" {
		t, err := parseDate(v, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "end_date inválido (formato YYYY-MM-DD)"})
			return
		}
		filter.EndDate = &t
	}
	if v := c.Query("page"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.Page = n
	}
	if v := c.Query("per_page"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.PerPage = n
	}

	logs, total, err := h.logRepo.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, ListBackupsResponse{Data: logs, Total: total})
}

func parseDate(s string, endOfDay bool) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return t, nil
}
