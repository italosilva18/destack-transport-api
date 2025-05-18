package configs

import (
	"github.com/spf13/viper"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	Environment  string   `mapstructure:"ENVIRONMENT"`
	ServerPort   string   `mapstructure:"SERVER_PORT"`
	DBConfig     DBConfig `mapstructure:",squash"`
	JWTSecret    string   `mapstructure:"JWT_SECRET"`
	JWTExpiresIn int      `mapstructure:"JWT_EXPIRES_IN"`
}

// DBConfig armazena configurações do banco de dados
type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

// LoadConfig carrega as configurações do arquivo ou variáveis de ambiente
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	// Ignorar erro se o arquivo de configuração não existir
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)

	// Valores padrão
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}

	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.JWTExpiresIn == 0 {
		config.JWTExpiresIn = 24 // 24 horas
	}

	return
}
