package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/geografico"
	"gorm.io/gorm"
)

// setupGeograficoRoutes configura as rotas geográficas
func setupGeograficoRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler geográfico
	geograficoHandler := geografico.NewGeograficoHandler(db)

	// Grupo de rotas geográficas
	geograficoRoutes := router.Group("/geografico")
	{
		geograficoRoutes.GET("", geograficoHandler.GetDadosGeograficos)
		geograficoRoutes.GET("/origens", geograficoHandler.GetTopOrigens)
		geograficoRoutes.GET("/destinos", geograficoHandler.GetTopDestinos)
		geograficoRoutes.GET("/rotas", geograficoHandler.GetRotasFrequentes)
		geograficoRoutes.GET("/fluxo-ufs", geograficoHandler.GetFluxoUFs)
	}
}
