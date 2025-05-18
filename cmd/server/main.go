package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seu-usuario/destack-transport-api/configs"
	"github.com/seu-usuario/destack-transport-api/internal/api/routes"
	"github.com/seu-usuario/destack-transport-api/pkg/database"
	"github.com/seu-usuario/destack-transport-api/pkg/logger"
)

func main() {
	// Inicializar o logger
	logger.InitLogger()
	log := logger.GetLogger()
	log.Info().Msg("Iniciando o servidor da API Destack Transport")
	
	// Carregar configurações
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível carregar as configurações")
	}
	
	// Definir modo de execução do Gin (release/debug)
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Inicializar conexão com o banco de dados
	db, err := database.InitDB(config.DBConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao banco de dados")
	}
	
	// Criar o router Gin
	router := gin.New()
	
	// Aplicar middleware padrão do Gin
	router.Use(gin.Recovery())
	
	// Configurar rotas
	routes.SetupRoutes(router, db)
	
	// Configurar servidor HTTP
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.ServerPort),
		Handler: router,
	}
	
	// Iniciar o servidor em uma goroutine
	go func() {
		log.Info().Msgf("Servidor HTTP iniciado na porta %s", config.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Falha ao iniciar o servidor HTTP")
		}
	}()
	
	// Esperar por sinais de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Desligando o servidor...")
	
	// Definir um timeout para o desligamento
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Erro ao desligar o servidor")
	}
	
	log.Info().Msg("Servidor parado com sucesso")
}