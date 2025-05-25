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

	// Lista de modelos para migrar na ordem correta (respeitando dependências)
	err := db.AutoMigrate(
		// Entidades base
		&models.User{},
		&models.Empresa{},
		&models.Veiculo{},
		&models.Upload{},

		// Documentos fiscais
		&models.CTE{},
		&models.MDFE{},

		// Outras entidades
		&models.Manutencao{},
	)

	if err != nil {
		log.Error().Err(err).Msg("Erro durante a migração")
		return err
	}

	// Criar índices adicionais se necessário
	createAdditionalIndexes(db)

	log.Info().Msg("Migração dos modelos concluída com sucesso")
	return nil
}

// createAdditionalIndexes cria índices compostos e especiais
func createAdditionalIndexes(db *gorm.DB) {
	log := logger.GetLogger()

	// Índices compostos para CT-e
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ctes_emitente_data ON ctes(emitente_id, data_emissao DESC)").Error; err != nil {
		log.Warn().Err(err).Msg("Erro ao criar índice idx_ctes_emitente_data")
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ctes_modalidade_data ON ctes(modalidade_frete, data_emissao DESC)").Error; err != nil {
		log.Warn().Err(err).Msg("Erro ao criar índice idx_ctes_modalidade_data")
	}

	// Índices compostos para MDF-e
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_mdfes_veiculo_data ON mdfes(veiculo_tracao_id, data_emissao DESC)").Error; err != nil {
		log.Warn().Err(err).Msg("Erro ao criar índice idx_mdfes_veiculo_data")
	}

	// Índice para busca de texto em observações
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ctes_obs_gin ON ctes USING gin(to_tsvector('portuguese', obs_gerais))").Error; err != nil {
		log.Warn().Err(err).Msg("Erro ao criar índice GIN para observações")
	}
}
