package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Setup: Criar arquivo .env temporário para testes
	envContent := `
ENVIRONMENT=test
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=test
DB_PASSWORD=test
DB_NAME=test
DB_SSLMODE=disable
JWT_SECRET=test_secret_key_for_testing_only
JWT_EXPIRES_IN=24
`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(".env")
	os.Exit(code)
}

func generateTestToken(secret string, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Teste 1: Requisição sem header Authorization
	t.Run("No_Authorization_Header", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header is required")
	})

	// Teste 2: Header Authorization com formato inválido
	t.Run("Invalid_Authorization_Format", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header format must be")
	})

	// Teste 3: Token JWT inválido
	t.Run("Invalid_JWT_Token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	// Teste 4: Token JWT expirado
	t.Run("Expired_JWT_Token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Gerar token expirado
		claims := jwt.MapClaims{
			"user_id":  "123",
			"username": "testuser",
			"role":     "user",
			"exp":      time.Now().Add(-time.Hour).Unix(), // Expirado há 1 hora
		}
		token, _ := generateTestToken("test_secret_key_for_testing_only", claims)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	// Teste 5: Token JWT válido
	t.Run("Valid_JWT_Token", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			// Verificar se os claims foram adicionados ao contexto
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")
			role, _ := c.Get("user_role")

			c.JSON(200, gin.H{
				"message":  "success",
				"user_id":  userID,
				"username": username,
				"role":     role,
			})
		})

		// Gerar token válido
		claims := jwt.MapClaims{
			"user_id":  "123",
			"username": "testuser",
			"role":     "admin",
			"exp":      time.Now().Add(time.Hour).Unix(), // Expira em 1 hora
		}
		token, _ := generateTestToken("test_secret_key_for_testing_only", claims)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
		assert.Contains(t, w.Body.String(), "123")
		assert.Contains(t, w.Body.String(), "testuser")
		assert.Contains(t, w.Body.String(), "admin")
	})

	// Teste 6: Token com algoritmo de assinatura incorreto
	t.Run("Invalid_Signing_Method", func(t *testing.T) {
		router := gin.New()
		router.Use(AuthMiddleware())
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// Gerar token com algoritmo diferente
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
			"user_id":  "123",
			"username": "testuser",
			"role":     "user",
			"exp":      time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})
}
