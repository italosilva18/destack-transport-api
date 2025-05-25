package main

import (
	"context"
	"fmt"
	"math"
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
	"gorm.io/gorm"
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

	// Aguardar banco de dados ficar disponível com backoff exponencial
	db, err := waitForDatabaseWithBackoff(config.DBConfig)
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
	if config.Environment == "development" || shouldRunSeeds() {
		log.Info().Msg("Executando seeds...")
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

// waitForDatabaseWithBackoff aguarda o banco com backoff exponencial
func waitForDatabaseWithBackoff(config configs.DBConfig) (*gorm.DB, error) {
	log := logger.GetLogger()

	maxRetries := 10
	baseDelay := 1 * time.Second
	maxDelay := 30 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Info().Msgf("Tentando conectar ao banco de dados... tentativa %d/%d", i+1, maxRetries)

		db, err := database.InitDB(config)
		if err == nil {
			// Testar a conexão
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					log.Info().Msg("Conexão com banco de dados estabelecida com sucesso!")
					return db, nil
				}
			}
		}

		// Calcular delay com backoff exponencial
		delay := time.Duration(math.Min(float64(baseDelay)*math.Pow(2, float64(i)), float64(maxDelay)))
		log.Warn().Err(err).Msgf("Tentativa %d falhou. Aguardando %v antes de tentar novamente...", i+1, delay)

		time.Sleep(delay)
	}

	return nil, fmt.Errorf("não foi possível conectar ao banco de dados após %d tentativas", maxRetries)
}

// shouldRunSeeds verifica se deve executar seeds
func shouldRunSeeds() bool {
	// Verificar variável de ambiente
	return os.Getenv("RUN_SEEDS") == "true"
}
