package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/middlewares"
	"gorm.io/gorm"
)

// SetupRoutes configura todas as rotas da API
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Configuração do middleware CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Middleware de logging
	router.Use(middlewares.LoggerMiddleware())

	// Rota de saúde
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "API Destack Transport em execução",
		})
	})

	// Grupo de rotas da API
	api := router.Group("/api")

	// Rotas públicas
	setupAuthRoutes(api, db)

	// Middleware de autenticação
	api.Use(middlewares.AuthMiddleware())

	// Rotas protegidas
	setupCTeRoutes(api, db)
	setupMDFeRoutes(api, db)
	setupUploadRoutes(api, db)
	setupDashboardRoutes(api, db)
	setupFinanceiroRoutes(api, db)
	setupGeograficoRoutes(api, db)
	setupManutencaoRoutes(api, db)
	setupAlertasRoutes(api, db)
	setupRelatoriosRoutes(api, db)
	setupConfiguracoesRoutes(api, db)
}

// Funções para configurar grupos de rotas específicos
func setupAuthRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de autenticação
}

func setupCTeRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de CT-e
}

func setupMDFeRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de MDF-e
}

func setupUploadRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de upload
}

func setupDashboardRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas do dashboard
}

func setupFinanceiroRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas financeiras
}

func setupGeograficoRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas geográficas
}

func setupManutencaoRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de manutenção
}

func setupAlertasRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de alertas
}

func setupRelatoriosRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de relatórios
}

func setupConfiguracoesRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// TODO: Implementar rotas de configurações
}
