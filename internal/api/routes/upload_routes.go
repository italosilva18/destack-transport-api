package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/upload"
	"gorm.io/gorm"
)

// setupUploadRoutes configura as rotas de upload
func setupUploadRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de upload
	uploadHandler := upload.NewUploadHandler(db)

	// Grupo de rotas de upload
	uploadRoutes := router.Group("/upload")
	{
		uploadRoutes.POST("/single", uploadHandler.UploadSingle)
		// Implementar outros endpoints de upload conforme necessário
	}

	// Rotas para gestão de uploads
	uploadsRoutes := router.Group("/uploads")
	{
		uploadsRoutes.GET("", uploadHandler.ListUploads)
		uploadsRoutes.GET("/:id", uploadHandler.GetUpload)
	}
}
