package empresa

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// EmpresaHandler contém os handlers para empresas
type EmpresaHandler struct {
	db     *gorm.DB
	logger zerolog.Logger
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
	CNPJ         *string `json:"cnpj" binding:"omitempty,len=14"`
	CPF          *string `json:"cpf" binding:"omitempty,len=11"`
	RazaoSocial  string  `json:"razao_social" binding:"required"`
	NomeFantasia *string `json:"nome_fantasia"`
	UF           string  `json:"uf" binding:"required,len=2"`
}

// UpdateEmpresaRequest representa os dados para atualizar uma empresa
type UpdateEmpresaRequest struct {
	RazaoSocial  string  `json:"razao_social"`
	NomeFantasia *string `json:"nome_fantasia"`
	UF           string  `json:"uf" binding:"omitempty,len=2"`
}

// ListEmpresasRequest representa os parâmetros para listar empresas
type ListEmpresasRequest struct {
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Limit   int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search" binding:"omitempty"`
	UF      string `form:"uf" binding:"omitempty,len=2"`
	TipoDoc string `form:"tipo_doc" binding:"omitempty,oneof=CNPJ CPF"`
}

// CreateEmpresa cria uma nova empresa
func (h *EmpresaHandler) CreateEmpresa(c *gin.Context) {
	var req CreateEmpresaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar que tem CNPJ ou CPF
	if req.CNPJ == nil && req.CPF == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ ou CPF é obrigatório"})
		return
	}

	// Verificar duplicidade
	var count int64
	query := h.db.Model(&models.Empresa{})
	if req.CNPJ != nil {
		query = query.Where("cnpj = ?", *req.CNPJ)
	} else {
		query = query.Where("cpf = ?", *req.CPF)
	}
	query.Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Empresa já cadastrada"})
		return
	}

	// Criar empresa
	empresa := models.Empresa{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
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

	// Preparar atualizações
	updates := map[string]interface{}{}

	if req.RazaoSocial != "" {
		updates["razao_social"] = req.RazaoSocial
	}

	if req.NomeFantasia != nil {
		updates["nome_fantasia"] = req.NomeFantasia
	}

	if req.UF != "" {
		updates["uf"] = req.UF
	}

	// Aplicar atualizações
	if err := h.db.Model(&empresa).Updates(updates).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao atualizar empresa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar empresa"})
		return
	}

	// Buscar empresa atualizada
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

	// Verificar se existem CT-es ou MDF-es vinculados
	var countCTEs int64
	h.db.Model(&models.CTE{}).Where("emitente_id = ? OR destinatario_id = ? OR remetente_id = ?", id, id, id).Count(&countCTEs)

	if countCTEs > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empresa possui documentos fiscais vinculados e não pode ser excluída",
		})
		return
	}

	if err := h.db.Delete(&empresa).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao excluir empresa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir empresa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Empresa excluída com sucesso"})
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

	// Buscar estatísticas relacionadas
	var stats struct {
		TotalCTEsEmitidos  int64      `json:"total_ctes_emitidos"`
		TotalCTEsRecebidos int64      `json:"total_ctes_recebidos"`
		ValorTotalEmitido  float64    `json:"valor_total_emitido"`
		ValorTotalRecebido float64    `json:"valor_total_recebido"`
		UltimaMovimentacao *time.Time `json:"ultima_movimentacao"`
	}

	// Total de CT-es emitidos
	h.db.Model(&models.CTE{}).Where("emitente_id = ?", id).Count(&stats.TotalCTEsEmitidos)

	// Total de CT-es recebidos
	h.db.Model(&models.CTE{}).Where("destinatario_id = ?", id).Count(&stats.TotalCTEsRecebidos)

	// Valor total emitido
	h.db.Model(&models.CTE{}).
		Where("emitente_id = ?", id).
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&stats.ValorTotalEmitido)

	// Valor total recebido
	h.db.Model(&models.CTE{}).
		Where("destinatario_id = ?", id).
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&stats.ValorTotalRecebido)

	// Última movimentação
	var ultimoCTE models.CTE
	if err := h.db.Where("emitente_id = ? OR destinatario_id = ?", id, id).
		Order("data_emissao DESC").
		First(&ultimoCTE).Error; err == nil {
		stats.UltimaMovimentacao = &ultimoCTE.DataEmissao
	}

	c.JSON(http.StatusOK, gin.H{
		"empresa":      empresa,
		"estatisticas": stats,
	})
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
	if req.Search != "" {
		searchWildcard := "%" + req.Search + "%"
		query = query.Where("razao_social ILIKE ? OR nome_fantasia ILIKE ? OR cnpj LIKE ? OR cpf LIKE ?",
			searchWildcard, searchWildcard, req.Search, req.Search)
	}

	if req.UF != "" {
		query = query.Where("uf = ?", req.UF)
	}

	if req.TipoDoc == "CNPJ" {
		query = query.Where("cnpj IS NOT NULL")
	} else if req.TipoDoc == "CPF" {
		query = query.Where("cpf IS NOT NULL")
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

// SearchEmpresas busca empresas para autocomplete
func (h *EmpresaHandler) SearchEmpresas(c *gin.Context) {
	query := c.Query("q")
	if query == "" || len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query deve ter pelo menos 2 caracteres"})
		return
	}

	tipo := c.Query("tipo") // emitente, destinatario, remetente
	limit := 10

	var empresas []models.Empresa
	searchWildcard := "%" + query + "%"

	dbQuery := h.db.Select("id", "cnpj", "cpf", "razao_social", "nome_fantasia", "uf").
		Where("razao_social ILIKE ? OR nome_fantasia ILIKE ? OR cnpj LIKE ? OR cpf LIKE ?",
			searchWildcard, searchWildcard, query, query).
		Limit(limit)

	// Filtrar por tipo se especificado
	if tipo != "" {
		// Aqui podemos adicionar lógica específica se necessário
		// Por exemplo, filtrar apenas empresas que já foram emitentes
	}

	if err := dbQuery.Find(&empresas).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar empresas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar empresas"})
		return
	}

	// Formatar resposta para autocomplete
	var results []gin.H
	for _, empresa := range empresas {
		doc := ""
		if empresa.CNPJ != nil {
			doc = *empresa.CNPJ
		} else if empresa.CPF != nil {
			doc = *empresa.CPF
		}

		label := empresa.RazaoSocial
		if doc != "" {
			label += " - " + doc
		}

		results = append(results, gin.H{
			"id":           empresa.ID,
			"label":        label,
			"razao_social": empresa.RazaoSocial,
			"documento":    doc,
			"uf":           empresa.UF,
		})
	}

	c.JSON(http.StatusOK, results)
}
