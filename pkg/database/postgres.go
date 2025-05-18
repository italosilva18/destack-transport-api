package database

import (
	"fmt"
	"time"

	"github.com/italosilva18/destack-transport-api/configs"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DB é a instância global do banco de dados
var DB *gorm.DB

// InitDB inicializa a conexão com o banco de dados
func InitDB(config configs.DBConfig) (*gorm.DB, error) {
	log := logger.GetLogger()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	log.Info().Msg("Conectando ao banco de dados PostgreSQL...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		return nil, err
	}

	// Configuração do pool de conexões
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Define o número máximo de conexões abertas
	sqlDB.SetMaxOpenConns(100)

	// Define o número máximo de conexões no pool
	sqlDB.SetMaxIdleConns(10)

	// Define o tempo máximo de vida de uma conexão
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verifica se a conexão está funcionando
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Conexão com o banco de dados estabelecida com sucesso")

	// Define a instância global
	DB = db

	return db, nil
}

// GetDB retorna a instância atual do banco de dados
func GetDB() *gorm.DB {
	return DB
}
