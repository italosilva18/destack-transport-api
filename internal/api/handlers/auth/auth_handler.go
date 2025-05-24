package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	// Usar SQLite em memória para testes
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrar os modelos necessários
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return nil, err
	}

	// Criar usuário de teste
	user := models.User{
		Name:     "Test User",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123", // Será hasheado pelo BeforeSave
		Role:     "user",
		Active:   true,
	}
	db.Create(&user)

	return db, nil
}

func TestLogin(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)

	handler := NewAuthHandler(db, "test_secret", 24)
	router := gin.New()
	router.POST("/login", handler.Login)

	// Teste 1: Login com sucesso
	t.Run("Login_Success", func(t *testing.T) {
		loginData := LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, "testuser", response.User.Username)
	})

	// Teste 2: Login com credenciais inválidas
	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		loginData := LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Teste 3: Login com usuário não existente
	t.Run("Login_UserNotFound", func(t *testing.T) {
		loginData := LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Teste 4: Login com dados inválidos
	t.Run("Login_InvalidData", func(t *testing.T) {
		invalidData := `{"username": ""}`

		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(invalidData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)

	handler := NewAuthHandler(db, "test_secret", 24)
	router := gin.New()
	router.POST("/logout", handler.Logout)

	// Teste: Logout sempre retorna sucesso
	t.Run("Logout_Success", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Logout realizado com sucesso", response["message"])
	})
}

func TestProfile(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)

	// Buscar usuário criado no setup
	var user models.User
	db.First(&user, "username = ?", "testuser")

	handler := NewAuthHandler(db, "test_secret", 24)
	router := gin.New()

	// Simular middleware de autenticação
	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID.String())
		c.Next()
	})

	router.GET("/profile", handler.Profile)

	// Teste 1: Buscar perfil com sucesso
	t.Run("Profile_Success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", response["username"])
		assert.Equal(t, "test@example.com", response["email"])
	})
}

func TestProfileWithoutAuth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := setupTestDB()
	assert.NoError(t, err)

	handler := NewAuthHandler(db, "test_secret", 24)
	router := gin.New()
	router.GET("/profile", handler.Profile)

	// Teste: Buscar perfil sem autenticação
	t.Run("Profile_Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
