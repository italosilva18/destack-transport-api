// internal/api/handlers/empresa/empresa_handler.go
package empresa

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// EmpresaHandler contém os handlers para empresas
type EmpresaHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewEmpresaHandler cria uma nova instância de EmpresaHandler
func NewEmpresaHandler(db *gorm.DB) *EmpresaHandler {
	return &EmpresaHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// CreateEmpresaRequest representa os dados para criar uma empresa
type CreateEmpresaRequest struct {
	CNPJ         *string `json:"cnpj" binding:"omitempty"`
	CPF          *string `json:"cpf" binding:"omitempty"`
	RazaoSocial  string  `json:"razao_social" binding:"required"`
	NomeFantasia *string `json:"nome_fantasia"`
	UF           string  `json:"uf" binding:"required,len=2"`
}

// CreateEmpresa cria uma nova empresa
func (h *EmpresaHandler) CreateEmpresa(c *gin.Context) {
	var req CreateEmpresaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar que pelo menos um dos campos CNPJ ou CPF foi preenchido
	if req.CNPJ == nil && req.CPF == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "É necessário informar CNPJ ou CPF"})
		return
	}

	// Verificar se já existe empresa com mesmo CNPJ ou CPF
	if req.CNPJ != nil {
		var count int64
		h.db.Model(&models.Empresa{}).Where("cnpj = ?", req.CNPJ).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe uma empresa com este CNPJ"})
			return
		}
	}

	if req.CPF != nil {
		var count int64
		h.db.Model(&models.Empresa{}).Where("cpf = ?", req.CPF).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe uma empresa com este CPF"})
			return
		}
	}

	// Criar a empresa
	empresa := models.Empresa{
		CNPJ:         req.CNPJ,
		CPF:          req.CPF,
		RazaoSocial:  req.RazaoSocial,
		NomeFantasia: req.NomeFantasia,
		UF:           req.UF,
	}

	if err := h.db.Create(&empresa).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao criar empresa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar empresa"})
		return
	}

	c.JSON(http.StatusCreated, empresa)
}

// GetEmpresa obtém uma empresa pelo ID
func (h *EmpresaHandler) GetEmpresa(c *gin.Context) {
	id := c.Param("id")

	var empresa models.Empresa
	result := h.db.First(&empresa, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Empresa não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, empresa)
}

// UpdateEmpresaRequest representa os dados para atualizar uma empresa
type UpdateEmpresaRequest struct {
	CNPJ         *string `json:"cnpj"`
	CPF          *string `json:"cpf"`
	RazaoSocial  string  `json:"razao_social"`
	NomeFantasia *string `json:"nome_fantasia"`
	UF           string  `json:"uf" binding:"omitempty,len=2"`
}

// UpdateEmpresa atualiza uma empresa existente
func (h *EmpresaHandler) UpdateEmpresa(c *gin.Context) {
	id := c.Param("id")

	var empresa models.Empresa
	result := h.db.First(&empresa, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Empresa não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	var req UpdateEmpresaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar CNPJ único
	if req.CNPJ != nil && *req.CNPJ != *empresa.CNPJ {
		var count int64
		h.db.Model(&models.Empresa{}).Where("cnpj = ? AND id != ?", req.CNPJ, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe uma empresa com este CNPJ"})
			return
		}
	}

	// Verificar CPF único
	if req.CPF != nil && *req.CPF != *empresa.CPF {
		var count int64
		h.db.Model(&models.Empresa{}).Where("cpf = ? AND id != ?", req.CPF, id).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe uma empresa com este CPF"})
			return
		}
	}

	// Atualizar campos
	updates := map[string]interface{}{}

	if req.CNPJ != nil {
		updates["cnpj"] = req.CNPJ
	}

	if req.CPF != nil {
		updates["cpf"] = req.CPF
	}

	if req.RazaoSocial != "" {
		updates["razao_social"] = req.RazaoSocial
	}

	if req.NomeFantasia != nil {
		updates["nome_fantasia"] = req.NomeFantasia
	}

	if req.UF != "" {
		updates["uf"] = req.UF
	}

	if err := h.db.Model(&empresa).Updates(updates).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao atualizar empresa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar empresa"})
		return
	}

	// Recarregar a empresa com os dados atualizados
	h.db.First(&empresa, "id = ?", id)

	c.JSON(http.StatusOK, empresa)
}

// DeleteEmpresa exclui uma empresa
func (h *EmpresaHandler) DeleteEmpresa(c *gin.Context) {
	id := c.Param("id")

	var empresa models.Empresa
	result := h.db.First(&empresa, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Empresa não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	// Verificar se a empresa está sendo usada em outros registros
	var countCTEs int64
	h.db.Model(&models.CTE{}).Where("emitente_id = ? OR remetente_id = ? OR destinatario_id = ?", id, id, id).Count(&countCTEs)

	if countCTEs > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não é possível excluir esta empresa pois ela está associada a CT-es"})
		return
	}

	if err := h.db.Delete(&empresa).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao excluir empresa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir empresa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Empresa excluída com sucesso"})
}

// ListEmpresasRequest representa os parâmetros para listar empresas
type ListEmpresasRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	UF         string `form:"uf" binding:"omitempty,len=2"`
	CNPJ       string `form:"cnpj" binding:"omitempty"`
	CPF        string `form:"cpf" binding:"omitempty"`
	SearchText string `form:"search_text" binding:"omitempty"`
}

// ListEmpresas lista as empresas com filtros e paginação
func (h *EmpresaHandler) ListEmpresas(c *gin.Context) {
	var req ListEmpresasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores padrão para paginação
	page := 1
	if req.Page > 0 {
		page = req.Page
	}

	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}

	offset := (page - 1) * limit

	// Construir query
	query := h.db.Model(&models.Empresa{})

	// Aplicar filtros
	if req.UF != "" {
		query = query.Where("uf = ?", req.UF)
	}

	if req.CNPJ != "" {
		query = query.Where("cnpj LIKE ?", "%"+req.CNPJ+"%")
	}

	if req.CPF != "" {
		query = query.Where("cpf LIKE ?", "%"+req.CPF+"%")
	}

	if req.SearchText != "" {
		searchWildcard := "%" + req.SearchText + "%"
		query = query.Where("razao_social LIKE ? OR nome_fantasia LIKE ?", searchWildcard, searchWildcard)
	}

	// Contar total para paginação
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar empresas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar empresas"})
		return
	}

	// Buscar empresas com paginação
	var empresas []models.Empresa
	if err := query.Offset(offset).Limit(limit).Order("razao_social ASC").Find(&empresas).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao listar empresas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar empresas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": empresas,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}
