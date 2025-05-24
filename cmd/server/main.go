package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/configs"
	"github.com/italosilva18/destack-transport-api/internal/api/routes"
	"github.com/italosilva18/destack-transport-api/pkg/database"
	"github.com/italosilva18/destack-transport-api/pkg/database/seeds"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
)

func main() {
	// Verificar se é apenas health check
	if len(os.Args) > 1 && os.Args[1] == "health" {
		checkHealth()
		return
	}

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

	// Aguardar banco de dados ficar disponível (importante para Docker)
	waitForDatabase(config.DBConfig)

	// Inicializar conexão com o banco de dados
	db, err := database.InitDB(config.DBConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao banco de dados")
	}

	// Auto-migração dos modelos
	log.Info().Msg("Executando auto-migração dos modelos...")
	err = database.MigrateModels(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro na migração do banco de dados")
	}

	// Executar seeds se necessário
	if config.Environment == "development" {
		log.Info().Msg("Executando seeds de desenvolvimento...")
		if err := seeds.SeedUsers(db); err != nil {
			log.Error().Err(err).Msg("Erro ao executar seed de usuários")
		}
	}

	// Criar o router Gin
	router := gin.New()

	// Aplicar middleware padrão do Gin
	router.Use(gin.Recovery())

	// Configurar rotas
	routes.SetupRoutes(router, db)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar o servidor em uma goroutine
	go func() {
		log.Info().Msgf("Servidor HTTP iniciado na porta %s", config.ServerPort)
		log.Info().Msgf("Ambiente: %s", config.Environment)

		if config.Environment == "development" {
			log.Info().Msg("Usuário padrão: admin / admin123")
		}

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

	// Fechar conexão com o banco
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Info().Msg("Servidor parado com sucesso")
}

// waitForDatabase aguarda o banco de dados ficar disponível
func waitForDatabase(config configs.DBConfig) {
	log := logger.GetLogger()
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err := database.InitDB(config)
		if err == nil {
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					sqlDB.Close()
					log.Info().Msg("Banco de dados está disponível")
					return
				}
			}
		}

		log.Warn().Msgf("Aguardando banco de dados... tentativa %d/%d", i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	log.Fatal().Msg("Tempo limite excedido aguardando o banco de dados")
}

// checkHealth verifica se a API está respondendo
func checkHealth() {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Println("API não está respondendo")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("API está saudável")
		os.Exit(0)
	} else {
		fmt.Println("API retornou status não saudável")
		os.Exit(1)
	}
}
