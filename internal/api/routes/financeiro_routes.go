package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/financeiro"
	"gorm.io/gorm"
)

// setupFinanceiroRoutes configura as rotas financeiras
func setupFinanceiroRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler financeiro
	financeiroHandler := financeiro.NewFinanceiroHandler(db)

	// Grupo de rotas financeiras
	financeiroRoutes := router.Group("/financeiro")
	{
		financeiroRoutes.GET("", financeiroHandler.GetDadosFinanceiros)
		financeiroRoutes.GET("/faturamento-mensal", financeiroHandler.GetFaturamentoMensal)
		financeiroRoutes.GET("/agrupado", financeiroHandler.GetDadosAgrupados)
		financeiroRoutes.GET("/detalhes/:tipo/:id", financeiroHandler.GetDetalheItem)
	}
}
