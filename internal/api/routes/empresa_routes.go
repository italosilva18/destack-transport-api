// internal/api/routes/empresa_routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/empresa"
	"gorm.io/gorm"
)

// setupEmpresaRoutes configura as rotas de empresas
func setupEmpresaRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de empresa
	empresaHandler := empresa.NewEmpresaHandler(db)

	// Grupo de rotas de empresa
	empresaRoutes := router.Group("/empresas")
	{
		empresaRoutes.POST("", empresaHandler.CreateEmpresa)
		empresaRoutes.GET("", empresaHandler.ListEmpresas)
		empresaRoutes.GET("/:id", empresaHandler.GetEmpresa)
		empresaRoutes.PUT("/:id", empresaHandler.UpdateEmpresa)
		empresaRoutes.DELETE("/:id", empresaHandler.DeleteEmpresa)
	}
}
