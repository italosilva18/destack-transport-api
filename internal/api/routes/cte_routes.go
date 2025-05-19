package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/cte"
	"gorm.io/gorm"
)

// setupCTeRoutes configura as rotas de CT-e
func setupCTeRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de CT-e
	cteHandler := cte.NewCTEHandler(db)

	// Grupo de rotas de CT-e
	cteRoutes := router.Group("/ctes")
	{
		cteRoutes.GET("", cteHandler.ListCTEs)
		cteRoutes.GET("/:chave", cteHandler.GetCTE)
		cteRoutes.GET("/:chave/download-xml", cteHandler.DownloadXML)
		cteRoutes.GET("/:chave/dacte", cteHandler.GerarDACTE)
		cteRoutes.POST("/:chave/reprocess", cteHandler.Reprocessar)
	}

	// Rota para o painel de CT-e
	painelRoutes := router.Group("/paineis")
	{
		painelRoutes.GET("/cte", cteHandler.GetPainelCTE)
	}
}
