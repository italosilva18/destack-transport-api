package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/mdfe"
	"gorm.io/gorm"
)

// setupMDFeRoutes configura as rotas de MDF-e
func setupMDFeRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de MDF-e
	mdfeHandler := mdfe.NewMDFEHandler(db)

	// Grupo de rotas de MDF-e
	mdfeRoutes := router.Group("/mdfes")
	{
		mdfeRoutes.GET("", mdfeHandler.ListMDFEs)
		mdfeRoutes.GET("/:chave", mdfeHandler.GetMDFE)
		mdfeRoutes.GET("/:chave/download-xml", mdfeHandler.DownloadXML)
		mdfeRoutes.GET("/:chave/damdfe", mdfeHandler.GerarDAMDFE)
		mdfeRoutes.POST("/:chave/reprocess", mdfeHandler.Reprocessar)
		mdfeRoutes.POST("/:chave/encerrar", mdfeHandler.Encerrar)
		mdfeRoutes.GET("/:chave/documentos", mdfeHandler.GetDocumentosVinculados)
	}

	// Rota para o painel de MDF-e
	painelRoutes := router.Group("/paineis")
	{
		painelRoutes.GET("/mdfe", mdfeHandler.GetPainelMDFE)
	}
}
