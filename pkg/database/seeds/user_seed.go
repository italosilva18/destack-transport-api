package seeds

import (
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// SeedUsers cria usuários iniciais se não existirem
func SeedUsers(db *gorm.DB) error {
	log := logger.GetLogger()
	log.Info().Msg("Verificando necessidade de seed de usuários...")

	var count int64
	db.Model(&models.User{}).Count(&count)

	if count > 0 {
		log.Info().Msg("Usuários já existem, pulando seed")
		return nil
	}

	log.Info().Msg("Criando usuário admin padrão")

	// Criar usuário admin
	admin := models.User{
		Name:     "Administrador",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "admin123", // Será hasheado pelo hook BeforeSave
		Role:     "admin",
		Active:   true,
	}

	result := db.Create(&admin)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Erro ao criar usuário admin")
		return result.Error
	}

	log.Info().Str("username", admin.Username).Msg("Usuário admin criado com sucesso")
	return nil
}
