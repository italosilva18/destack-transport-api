package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/manutencao"
	"gorm.io/gorm"
)

// setupManutencaoRoutes configura as rotas de manutenção
func setupManutencaoRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de manutenção
	manutencaoHandler := manutencao.NewManutencaoHandler(db)

	// Grupo de rotas de manutenção
	manutencaoRoutes := router.Group("/manutencoes")
	{
		manutencaoRoutes.POST("", manutencaoHandler.CreateManutencao)
		manutencaoRoutes.GET("", manutencaoHandler.ListManutencoes)
		manutencaoRoutes.GET("/:id", manutencaoHandler.GetManutencao)
		manutencaoRoutes.PUT("/:id", manutencaoHandler.UpdateManutencao)
		manutencaoRoutes.DELETE("/:id", manutencaoHandler.DeleteManutencao)
		manutencaoRoutes.GET("/estatisticas", manutencaoHandler.GetEstatisticas)
	}
}
