package handlers

import "github.com/erick_nilson/mysql_auto_backup/internal/models"

// ErrorResponse é o corpo retornado quando uma requisição falha.
type ErrorResponse struct {
	Error string `json:"error" example:"unauthorized"`
}

// CreateConnectionRequest é o payload para cadastro de uma nova conexão MySQL.
type CreateConnectionRequest struct {
	Name     string `json:"name"      binding:"required" example:"prod-mysql-01"`
	Host     string `json:"host"      binding:"required" example:"db.exemplo.com"`
	Port     string `json:"port"      binding:"required" example:"3306"`
	User     string `json:"user"      binding:"required" example:"backup_user"`
	Password string `json:"password"  binding:"required" example:"s3nh4-d0-banc0"`
	Database string `json:"database"  binding:"required" example:"loja"`
	IsActive *bool  `json:"is_active" example:"true"`
}

// UpdateConnectionRequest é o payload para atualização parcial. Todos os
// campos são opcionais — os omitidos preservam o valor atual. Se quaisquer
// dados sensíveis forem alterados, a conexão é revalidada via Ping antes do
// update ser persistido.
type UpdateConnectionRequest struct {
	Name     *string `json:"name,omitempty"      example:"prod-mysql-01"`
	Host     *string `json:"host,omitempty"      example:"novo.exemplo.com"`
	Port     *string `json:"port,omitempty"      example:"3306"`
	User     *string `json:"user,omitempty"      example:"backup_user"`
	Password *string `json:"password,omitempty"  example:"s3nh4-nov4"`
	Database *string `json:"database,omitempty"  example:"loja"`
	IsActive *bool   `json:"is_active,omitempty" example:"true"`
}

// RunBackupRequest é o payload (opcional) para o disparo manual de backup.
// Quando ConnectionIDs é vazio ou ausente, o backup é executado para todas
// as conexões ativas.
type RunBackupRequest struct {
	ConnectionIDs []uint64 `json:"connection_ids,omitempty" example:"1,2,3"`
}

// RunBackupResponse confirma que o backup foi enfileirado em background.
type RunBackupResponse struct {
	Message       string   `json:"message"                  example:"backup disparado para todas as conexões ativas"`
	ConnectionIDs []uint64 `json:"connection_ids,omitempty" example:"1,2,3"`
}

// ListBackupsResponse é o envelope de listagem de logs de backup.
type ListBackupsResponse struct {
	Data  []models.BackupLog `json:"data"`
	Total int64              `json:"total" example:"42"`
}
