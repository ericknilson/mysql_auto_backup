package router

import (
	"net/http"

	"github.com/erick_nilson/mysql_auto_backup/internal/handlers"
	"github.com/erick_nilson/mysql_auto_backup/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

func New(apiKey string, connHandler *handlers.ConnectionHandler, backupHandler *handlers.BackupHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	api := r.Group("/api", middleware.APIKey(apiKey))
	{
		conns := api.Group("/connections")
		{
			conns.POST("", connHandler.Create)
			conns.GET("", connHandler.List)
			conns.GET("/:id", connHandler.Get)
			conns.PUT("/:id", connHandler.Update)
			conns.DELETE("/:id", connHandler.Delete)
		}

		backups := api.Group("/backups")
		{
			backups.POST("/run", backupHandler.Run)
			backups.GET("", backupHandler.List)
		}
	}

	return r
}
