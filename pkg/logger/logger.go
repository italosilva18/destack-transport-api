package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger inicializa o logger
func InitLogger() {
	// Configurar formato e saída do log
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Use arquivos para logs em produção
	if os.Getenv("ENVIRONMENT") == "production" {
		// Criar arquivo de log se não existir
		logFile, err := os.OpenFile(
			"logs/api.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0666,
		)
		if err != nil {
			log.Error().Err(err).Msg("Não foi possível criar arquivo de log")
		} else {
			// Multiescritor para terminal e arquivo
			multi := zerolog.MultiLevelWriter(output, logFile)
			log.Logger = zerolog.New(multi).With().Timestamp().Logger()
		}
	} else {
		// Em desenvolvimento, só escreve no terminal
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	}
}

// GetLogger retorna a instância do logger
func GetLogger() zerolog.Logger {
	return log.Logger
}

// GetLogWriter retorna um writer para o log, útil para middlewares Gin
func GetLogWriter() io.Writer {
	return log.Logger
}
