package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/configs"
	"github.com/italosilva18/destack-transport-api/internal/api/handlers/auth"
	"github.com/italosilva18/destack-transport-api/internal/api/middlewares"
	"gorm.io/gorm"
)

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
