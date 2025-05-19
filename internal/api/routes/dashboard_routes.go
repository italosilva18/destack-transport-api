package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/dashboard"
	"gorm.io/gorm"
)

// setupDashboardRoutes configura as rotas do dashboard
func setupDashboardRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler do dashboard
	dashboardHandler := dashboard.NewDashboardHandler(db)

	// Grupo de rotas do dashboard
	dashboardRoutes := router.Group("/dashboard")
	{
		dashboardRoutes.GET("/cards", dashboardHandler.GetDashboardCards)
		dashboardRoutes.GET("/lancamentos", dashboardHandler.GetUltimosLancamentos)
		dashboardRoutes.GET("/cif-fob", dashboardHandler.GetCifFobData)
		// Adicionar mais endpoints conforme necessário
	}
}
