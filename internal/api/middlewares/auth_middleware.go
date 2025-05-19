package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger()

		// Obter o token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// O token deve começar com "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Authorization header format must be 'Bearer {token}'"})
			c.Abort()
			return
		}

		// TODO: Implementar a verificação do token JWT
		// Este é um exemplo básico, você deve obter a chave secreta das configurações
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar o método de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// TODO: Obter chave secreta das configurações
			return []byte("sua_chave_secreta"), nil
		})

		if err != nil {
			log.Error().Err(err).Msg("Erro ao validar token JWT")
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Verificar se o token é válido
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Adicionar claims ao contexto
			c.Set("user_id", claims["user_id"])
			c.Set("user_role", claims["role"])
			c.Next()
		} else {
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}
