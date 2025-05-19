package database

import (
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// MigrateModels executa a auto-migração dos modelos
func MigrateModels(db *gorm.DB) error {
	log := logger.GetLogger()
	log.Info().Msg("Iniciando migração dos modelos...")

	// Lista de modelos para migrar
	err := db.AutoMigrate(
		&models.User{},
		&models.Empresa{},
		&models.Veiculo{},
		&models.CTE{},
		&models.MDFE{},
		&models.Upload{},
		&models.Manutencao{},
		// Adicionar outros modelos conforme necessário
	)

	if err != nil {
		log.Error().Err(err).Msg("Erro durante a migração")
		return err
	}

	log.Info().Msg("Migração dos modelos concluída com sucesso")
	return nil
}
