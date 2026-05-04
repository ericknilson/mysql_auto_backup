package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"github.com/erick_nilson/mysql_auto_backup/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// _modelsRef garante que o swag (parser por arquivo) enxergue o pacote models
// para resolver referências como models.Connection nas anotações @Success.
var _ = models.Connection{}

type ConnectionHandler struct {
	svc *services.ConnectionService
}

func NewConnectionHandler(svc *services.ConnectionService) *ConnectionHandler {
	return &ConnectionHandler{svc: svc}
}

// Create godoc
//
//	@Summary		Cadastra uma nova conexão MySQL
//	@Description	Recebe credenciais em claro, valida com Ping e, em caso de sucesso, criptografa (AES-256-GCM) e persiste. Se a conexão não puder ser estabelecida, nada é gravado.
//	@Tags			connections
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			payload	body		CreateConnectionRequest	true	"Dados da conexão"
//	@Success		201		{object}	models.Connection
//	@Failure		400		{object}	ErrorResponse	"payload inválido"
//	@Failure		401		{object}	ErrorResponse	"API key ausente ou inválida"
//	@Failure		422		{object}	ErrorResponse	"não foi possível conectar ao banco com as credenciais fornecidas"
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/connections [post]
func (h *ConnectionHandler) Create(c *gin.Context) {
	var req CreateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	conn, err := h.svc.Create(c.Request.Context(), services.CreateInput{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		IsActive: req.IsActive,
	})
	if err != nil {
		if errors.Is(err, services.ErrConnectionUnreachable) {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, conn)
}

// Update godoc
//
//	@Summary		Atualiza uma conexão existente
//	@Description	Atualização parcial. Campos omitidos preservam o valor atual. Antes de persistir, a conexão é re-validada com as credenciais resultantes — se o Ping falhar, nada é alterado.
//	@Tags			connections
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		int						true	"ID da conexão"
//	@Param			payload	body		UpdateConnectionRequest	true	"Campos a atualizar (todos opcionais)"
//	@Success		200		{object}	models.Connection
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		422		{object}	ErrorResponse	"não foi possível conectar ao banco com as credenciais fornecidas"
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/connections/{id} [put]
func (h *ConnectionHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	var req UpdateConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	conn, err := h.svc.Update(c.Request.Context(), id, services.UpdateInput{
		Name:     req.Name,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		IsActive: req.IsActive,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "conexão não encontrada"})
			return
		}
		if errors.Is(err, services.ErrConnectionUnreachable) {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, conn)
}

// Get godoc
//
//	@Summary		Detalha uma conexão
//	@Description	Retorna os metadados de uma conexão. As credenciais nunca são expostas pela API.
//	@Tags			connections
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		int	true	"ID da conexão"
//	@Success		200	{object}	models.Connection
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/api/connections/{id} [get]
func (h *ConnectionHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	conn, err := h.svc.Get(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "conexão não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, conn)
}

// List godoc
//
//	@Summary		Lista todas as conexões
//	@Description	Retorna todas as conexões não excluídas (soft-deleted são automaticamente filtradas pelo GORM).
//	@Tags			connections
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{array}		models.Connection
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/connections [get]
func (h *ConnectionHandler) List(c *gin.Context) {
	conns, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, conns)
}

// Delete godoc
//
//	@Summary		Remove uma conexão (soft delete)
//	@Description	A conexão é marcada como excluída populando deleted_at. Logs de backup históricos relacionados continuam acessíveis via /api/backups.
//	@Tags			connections
//	@Security		ApiKeyAuth
//	@Param			id	path	int	true	"ID da conexão"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/api/connections/{id} [delete]
func (h *ConnectionHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}
