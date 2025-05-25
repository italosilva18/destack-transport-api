package configs

import (
	"os"
	"strconv"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	Environment  string
	ServerPort   string
	DBConfig     DBConfig
	JWTSecret    string
	JWTExpiresIn int
}

// DBConfig armazena configurações do banco de dados
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadConfig carrega as configurações usando apenas variáveis de ambiente
func LoadConfig(path string) (Config, error) {
	config := Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DBConfig: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "destack_transport"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWTSecret:    getEnv("JWT_SECRET", "default_jwt_secret_change_in_production"),
		JWTExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 24),
	}

	return config, nil
}

// getEnv obtém variável de ambiente ou retorna um valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt obtém variável de ambiente como int ou retorna um valor padrão
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
