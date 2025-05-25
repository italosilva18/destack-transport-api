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
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
	User      UserInfo  `json:"user"`
}

// UserInfo informações básicas do usuário
type UserInfo struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `json:"name" example:"Administrador"`
	Username string `json:"username" example:"admin"`
	Email    string `json:"email" example:"admin@destack.com.br"`
	Role     string `json:"role" example:"admin"`
}

// ErrorResponse resposta de erro padrão
type ErrorResponse struct {
	Error string `json:"error" example:"Credenciais inválidas"`
}

// MessageResponse resposta com mensagem
type MessageResponse struct {
	Message string `json:"message" example:"Operação realizada com sucesso"`
}

// ProfileResponse resposta do perfil do usuário
type ProfileResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `json:"name" example:"Administrador"`
	Username string `json:"username" example:"admin"`
	Email    string `json:"email" example:"admin@destack.com.br"`
	Role     string `json:"role" example:"admin"`
	Active   bool   `json:"active" example:"true"`
}

// Login autentica um usuário e retorna um token JWT
// @Summary Login do usuário
// @Description Autentica um usuário com username e senha
// @Tags Autenticação
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Credenciais de login"
// @Success 200 {object} LoginResponse "Login realizado com sucesso"
// @Failure 400 {object} ErrorResponse "Dados inválidos"
// @Failure 401 {object} ErrorResponse "Credenciais inválidas"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	var user models.User
	result := h.db.Where("username = ? AND active = ?", req.Username, true).First(&user)
	if result.Error != nil {
		h.logger.Error().Str("username", req.Username).Msg("Usuário não encontrado ou inativo")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Credenciais inválidas"})
		return
	}

	if err := user.CheckPassword(req.Password); err != nil {
		h.logger.Error().Str("username", req.Username).Msg("Senha incorreta")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Credenciais inválidas"})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Erro ao gerar token"})
		return
	}

	// Preparar resposta
	response := LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime,
		User: UserInfo{
			ID:       user.ID.String(),
			Name:     user.Name,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Logout realiza o logout do usuário (opcional, já que JWT é stateless)
// @Summary Logout do usuário
// @Description Realiza o logout do usuário (informativo, pois JWT é stateless)
// @Tags Autenticação
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} MessageResponse "Logout realizado com sucesso"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Para um logout efetivo em um sistema baseado em JWT,
	// seria necessário implementar uma lista negra de tokens ou usar refresh tokens
	// Aqui apenas retornamos uma resposta de sucesso
	c.JSON(http.StatusOK, MessageResponse{Message: "Logout realizado com sucesso"})
}

// Profile retorna o perfil do usuário autenticado
// @Summary Perfil do usuário
// @Description Retorna os dados do perfil do usuário autenticado
// @Tags Autenticação
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} ProfileResponse "Dados do perfil"
// @Failure 401 {object} ErrorResponse "Usuário não autenticado"
// @Failure 500 {object} ErrorResponse "Erro interno do servidor"
// @Router /auth/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Usuário não autenticado"})
		return
	}

	var user models.User
	result := h.db.First(&user, "id = ?", userID)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("user_id", userID.(string)).Msg("Erro ao buscar perfil do usuário")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Erro ao buscar perfil do usuário"})
		return
	}

	response := ProfileResponse{
		ID:       user.ID.String(),
		Name:     user.Name,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Active:   user.Active,
	}

	c.JSON(http.StatusOK, response)
}
