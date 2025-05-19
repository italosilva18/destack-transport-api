package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
)

// LoggerMiddleware registra informações sobre as requisições
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger()

		// Tempo antes de processar a requisição
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Processar requisição
		c.Next()

		// Tempo depois de processar a requisição
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Status da resposta
		status := c.Writer.Status()

		// Registrar log
		logEvent := log.Info()

		// Se for um erro, mudar o nível do log
		if status >= 400 {
			logEvent = log.Error()
		}

		// Adicionar campos ao log
		logEvent.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user-agent", c.Request.UserAgent()).
			Msg("Request")
	}
}
