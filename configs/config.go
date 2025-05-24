package configs

import (
	"os"
	"strconv"

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
	// Priorizar variáveis de ambiente
	viper.AutomaticEnv()

	// Tentar ler arquivo de configuração se existir
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Se o arquivo existir, ler dele
	if err := viper.ReadInConfig(); err != nil {
		// Se não existir arquivo, usar apenas variáveis de ambiente
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Se for outro erro, retornar
			return config, err
		}
	}

	// Unmarshal das configurações
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	// Verificar variáveis de ambiente diretamente se estiverem vazias
	if config.DBConfig.Host == "" {
		config.DBConfig.Host = os.Getenv("DB_HOST")
	}
	if config.DBConfig.Port == "" {
		config.DBConfig.Port = os.Getenv("DB_PORT")
	}
	if config.DBConfig.User == "" {
		config.DBConfig.User = os.Getenv("DB_USER")
	}
	if config.DBConfig.Password == "" {
		config.DBConfig.Password = os.Getenv("DB_PASSWORD")
	}
	if config.DBConfig.DBName == "" {
		config.DBConfig.DBName = os.Getenv("DB_NAME")
	}
	if config.DBConfig.SSLMode == "" {
		config.DBConfig.SSLMode = os.Getenv("DB_SSLMODE")
	}

	// Valores padrão
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	if config.JWTExpiresIn == 0 {
		jwtExp := os.Getenv("JWT_EXPIRES_IN")
		if jwtExp != "" {
			if exp, err := strconv.Atoi(jwtExp); err == nil {
				config.JWTExpiresIn = exp
			} else {
				config.JWTExpiresIn = 24
			}
		} else {
			config.JWTExpiresIn = 24
		}
	}
	if config.JWTSecret == "" {
		config.JWTSecret = os.Getenv("JWT_SECRET")
	}

	return config, nil
}
