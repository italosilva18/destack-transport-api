package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/configs"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/auth"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/cte"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/dashboard"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/financeiro"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/geografico"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/manutencao"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/mdfe"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/upload"
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

// setupAuthRoutes configura as rotas de autenticação
func setupAuthRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Carregar configurações
	config, _ := configs.LoadConfig(".")

	// Criar handler de autenticação
	authHandler := auth.NewAuthHandler(db, config.JWTSecret, config.JWTExpiresIn)

	// Grupo de rotas de autenticação
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.POST("/logout", authHandler.Logout)

		// Rota protegida pelo middleware de autenticação
		authProtected := authRoutes.Group("/")
		authProtected.Use(middlewares.AuthMiddleware())
		{
			authProtected.GET("/profile", authHandler.Profile)
		}
	}
}

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

// setupUploadRoutes configura as rotas de upload
func setupUploadRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Criar handler de upload
	uploadHandler := upload.NewUploadHandler(db)

	// Grupo de rotas de upload
	uploadRoutes := router.Group("/upload")
	{
		uploadRoutes.POST("/single", uploadHandler.UploadSingle)
		uploadRoutes.POST("/batch", uploadHandler.UploadBatch)
	}

	// Rotas para gestão de uploads
	uploadsRoutes := router.Group("/uploads")
	{
		uploadsRoutes.GET("", uploadHandler.ListUploads)
		uploadsRoutes.GET("/:id", uploadHandler.GetUpload)
		uploadsRoutes.DELETE("/:id", uploadHandler.DeleteUpload)
	}
}

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
	}
}

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
