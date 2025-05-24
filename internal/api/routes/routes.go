package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/middlewares"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// SetupRoutes configura todas as rotas da API
func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	log := logger.GetLogger()
	log.Info().Msg("Configurando rotas da API")

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

	// Middleware de autenticação para rotas protegidas
	protected := api.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	// Rotas protegidas
	setupCTeRoutes(protected, db)
	setupMDFeRoutes(protected, db)
	setupUploadRoutes(protected, db)
	setupDashboardRoutes(protected, db)
	setupFinanceiroRoutes(protected, db)
	setupGeograficoRoutes(protected, db)
	setupManutencaoRoutes(protected, db)
	setupAlertasRoutes(protected, db)
	setupRelatoriosRoutes(protected, db)
	setupConfiguracoesRoutes(protected, db)

	log.Info().Msg("Rotas da API configuradas com sucesso")
}

// setupAlertasRoutes configura as rotas de alertas
func setupAlertasRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Placeholder para implementação futura
	router.GET("/alertas", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Funcionalidade de alertas em desenvolvimento",
		})
	})
}

// setupRelatoriosRoutes configura as rotas de relatórios
func setupRelatoriosRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Placeholder para implementação futura
	router.GET("/relatorios", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Funcionalidade de relatórios em desenvolvimento",
		})
	})
}

// setupConfiguracoesRoutes configura as rotas de configurações
func setupConfiguracoesRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Placeholder para implementação futura
	configRoutes := router.Group("/configuracoes")
	{
		configRoutes.GET("/empresa", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Funcionalidade de configurações de empresa em desenvolvimento",
			})
		})

		configRoutes.GET("/parametros", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Funcionalidade de parâmetros do sistema em desenvolvimento",
			})
		})

		configRoutes.GET("/usuarios", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Funcionalidade de gerenciamento de usuários em desenvolvimento",
			})
		})
	}
}
