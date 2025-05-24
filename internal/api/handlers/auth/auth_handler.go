package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// AuthHandler contém os handlers relacionados à autenticação
type AuthHandler struct {
	db     *gorm.DB
	logger zerolog.Logger
	config struct {
		JWTSecret    string
		JWTExpiresIn int
	}
}

// NewAuthHandler cria uma nova instância de AuthHandler
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpiresIn int) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger.GetLogger(),
		config: struct {
			JWTSecret    string
			JWTExpiresIn int
		}{
			JWTSecret:    jwtSecret,
			JWTExpiresIn: jwtExpiresIn,
		},
	}
}

// LoginRequest representa os dados necessários para login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

// Login autentica um usuário e retorna um token JWT
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := h.db.Where("username = ? AND active = ?", req.Username, true).First(&user)
	if result.Error != nil {
		h.logger.Error().Str("username", req.Username).Msg("Usuário não encontrado ou inativo")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	if err := user.CheckPassword(req.Password); err != nil {
		h.logger.Error().Str("username", req.Username).Msg("Senha incorreta")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	// Criar token JWT
	expirationTime := time.Now().Add(time.Duration(h.config.JWTExpiresIn) * time.Hour)
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao gerar token JWT")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	// Preparar resposta
	response := LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime,
	}
	response.User.ID = user.ID.String()
	response.User.Name = user.Name
	response.User.Username = user.Username
	response.User.Email = user.Email
	response.User.Role = user.Role

	c.JSON(http.StatusOK, response)
}

// Logout realiza o logout do usuário (opcional, já que JWT é stateless)
func (h *AuthHandler) Logout(c *gin.Context) {
	// Para um logout efetivo em um sistema baseado em JWT,
	// seria necessário implementar uma lista negra de tokens ou usar refresh tokens
	// Aqui apenas retornamos uma resposta de sucesso
	c.JSON(http.StatusOK, gin.H{"message": "Logout realizado com sucesso"})
}

// Profile retorna o perfil do usuário autenticado
func (h *AuthHandler) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var user models.User
	result := h.db.First(&user, "id = ?", userID)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("user_id", userID.(string)).Msg("Erro ao buscar perfil do usuário")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar perfil do usuário"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"active":   user.Active,
	})
}
