package mdfe

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// MDFEHandler contém os handlers para MDFEs
type MDFEHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMDFEHandler cria uma nova instância de MDFEHandler
func NewMDFEHandler(db *gorm.DB) *MDFEHandler {
	return &MDFEHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// ListMDFEsRequest representa os parâmetros de request para listar MDFEs
type ListMDFEsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	DataInicio string `form:"data_inicio" binding:"omitempty"`
	DataFim    string `form:"data_fim" binding:"omitempty"`
	Placa      string `form:"placa" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty"`
	UFInicio   string `form:"uf_inicio" binding:"omitempty"`
	UFFim      string `form:"uf_fim" binding:"omitempty"`
}

// ListMDFEs lista os MDFEs com filtros e paginação
func (h *MDFEHandler) ListMDFEs(c *gin.Context) {
	var req ListMDFEsRequest
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
	query := h.db.Model(&models.MDFE{}).Preload("Emitente")

	// Aplicar filtros
	if req.DataInicio != "" && req.DataFim != "" {
		dataInicio, err := time.Parse("2006-01-02", req.DataInicio)
		if err == nil {
			dataFim, err := time.Parse("2006-01-02", req.DataFim)
			if err == nil {
				// Ajustar hora final para o final do dia
				dataFim = time.Date(dataFim.Year(), dataFim.Month(), dataFim.Day(), 23, 59, 59, 0, dataFim.Location())
				query = query.Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim)
			}
		}
	}

	if req.Placa != "" {
		// Assumindo que temos um relacionamento para Veiculo aqui
		query = query.Joins("JOIN veiculos v ON mdfe.veiculo_tracao_id = v.id").
			Where("v.placa LIKE ?", "%"+req.Placa+"%")
	}

	if req.Status != "" {
		if req.Status == "encerrado" {
			query = query.Where("encerrado = ?", true)
		} else if req.Status == "cancelado" {
			query = query.Where("cancelado = ?", true)
		} else {
			query = query.Where("status = ?", req.Status)
		}
	}

	if req.UFInicio != "" {
		query = query.Where("uf_inicio = ?", req.UFInicio)
	}

	if req.UFFim != "" {
		query = query.Where("uf_destino = ?", req.UFFim)
	}

	// Contar total para paginação
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar MDFEs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar MDFEs"})
		return
	}

	// Buscar MDFEs com paginação
	var mdfes []models.MDFE
	if err := query.Offset(offset).Limit(limit).Order("data_emissao DESC").Find(&mdfes).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao listar MDFEs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar MDFEs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": mdfes,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetMDFE obtém um MDFE pelo ID ou chave
func (h *MDFEHandler) GetMDFE(c *gin.Context) {
	chave := c.Param("chave")

	var mdfe models.MDFE
	result := h.db.Preload("Emitente").Where("chave = ?", chave).First(&mdfe)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("chave", chave).Msg("MDFE não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "MDFE não encontrado"})
		return
	}

	c.JSON(http.StatusOK, mdfe)
}

// DownloadXML baixa o XML de um MDFE
func (h *MDFEHandler) DownloadXML(c *gin.Context) {
	chave := c.Param("chave")

	// Em um sistema real, recuperaríamos o XML armazenado
	// Aqui, simulamos uma resposta com um XML simples
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<MDFe xmlns="http://www.portalfiscal.inf.br/mdfe">
  <infMDFe Id="MDFe` + chave + `">
    <!-- conteúdo do MDF-e aqui -->
  </infMDFe>
</MDFe>`

	c.Header("Content-Disposition", "attachment; filename=mdfe_"+chave+".xml")
	c.Data(http.StatusOK, "application/xml", []byte(xml))
}

// GerarDAMDFE gera um DAMDFE em PDF
func (h *MDFEHandler) GerarDAMDFE(c *gin.Context) {
	chave := c.Param("chave")

	// Em um sistema real, geraria o PDF
	// Aqui, simulamos uma resposta com um PDF simples
	c.Header("Content-Disposition", "attachment; filename=damdfe_"+chave+".pdf")
	c.Data(http.StatusOK, "application/pdf", []byte("Simulação de PDF do DAMDFE"))
}

// Reprocessar reprocessa um MDFE
func (h *MDFEHandler) Reprocessar(c *gin.Context) {
	chave := c.Param("chave")

	var mdfe models.MDFE
	result := h.db.Where("chave = ?", chave).First(&mdfe)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("chave", chave).Msg("MDFE não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "MDFE não encontrado"})
		return
	}

	// Em um sistema real, faria o reprocessamento
	// Aqui, apenas atualizamos um campo para simular
	if err := h.db.Model(&mdfe).Update("data_processamento", time.Now()).Error; err != nil {
		h.logger.Error().Err(err).Str("chave", chave).Msg("Erro ao reprocessar MDFE")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao reprocessar MDFE"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "MDFE reprocessado com sucesso",
		"chave":   chave,
	})
}

// Encerrar encerra um MDFE
func (h *MDFEHandler) Encerrar(c *gin.Context) {
	chave := c.Param("chave")

	var mdfe models.MDFE
	result := h.db.Where("chave = ?", chave).First(&mdfe)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("chave", chave).Msg("MDFE não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "MDFE não encontrado"})
		return
	}

	// Verificar se já está encerrado
	if mdfe.Encerrado {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MDFE já está encerrado"})
		return
	}

	// Em um sistema real, enviaria o evento de encerramento para a SEFAZ
	// Aqui, apenas atualizamos o status
	now := time.Now()
	updates := map[string]interface{}{
		"encerrado":         true,
		"data_encerramento": now,
	}

	if err := h.db.Model(&mdfe).Updates(updates).Error; err != nil {
		h.logger.Error().Err(err).Str("chave", chave).Msg("Erro ao encerrar MDFE")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encerrar MDFE"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "MDFE encerrado com sucesso",
		"chave":             chave,
		"data_encerramento": now,
	})
}

// GetDocumentosVinculados retorna os documentos vinculados a um MDFE
func (h *MDFEHandler) GetDocumentosVinculados(c *gin.Context) {
	chave := c.Param("chave")

	// Em um sistema real, buscaríamos a relação entre MDF-e e CT-e
	// Aqui, simulamos alguns dados
	documentos := []gin.H{
		{
			"tipo":  "CTE",
			"chave": "31234567890123456789012345678901234567890123",
			"valor": 1500.50,
		},
		{
			"tipo":  "CTE",
			"chave": "31234567890123456789012345678901234567890124",
			"valor": 2500.75,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"mdfe_chave": chave,
		"documentos": documentos,
	})
}

// PainelMDFEResponse representa a resposta para o painel de MDF-e
type PainelMDFEResponse struct {
	TotalMDFEs       int64              `json:"total_mdfes"`
	TotalAutorizados int64              `json:"total_autorizados"`
	TotalEncerrados  int64              `json:"total_encerrados"`
	TotalCancelados  int64              `json:"total_cancelados"`
	TotalCTEsPeriodo int64              `json:"total_ctes_periodo"`
	Eficiencia       float64            `json:"eficiencia"`
	TopVeiculos      []TopVeiculo       `json:"top_veiculos"`
	DistribuicaoCTEs []DistribuicaoCTEs `json:"distribuicao_ctes"`
}

// Continuação do MDFEHandler...

// TopVeiculo representa um veículo no ranking
type TopVeiculo struct {
	ID              string `json:"id"`
	Placa           string `json:"placa"`
	TotalMDFEs      int64  `json:"total_mdfes"`
	TotalDocumentos int64  `json:"total_documentos"`
}

// DistribuicaoCTEs representa a distribuição de CT-es por MDF-e
type DistribuicaoCTEs struct {
	MDFEChave      string `json:"mdfe_chave"`
	Numero         int    `json:"numero"`
	QuantidadeCTEs int64  `json:"quantidade_ctes"`
}

// GetPainelMDFE retorna os dados para o painel de MDF-e
func (h *MDFEHandler) GetPainelMDFE(c *gin.Context) {
	// Obter parâmetros de filtro
	dataInicio := c.Query("data_inicio")
	dataFim := c.Query("data_fim")

	// Parsear datas
	var dataInicioTime, dataFimTime time.Time
	var err error

	if dataInicio != "" {
		dataInicioTime, err = time.Parse("2006-01-02", dataInicio)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
			return
		}
	} else {
		// Padrão: primeiro dia do mês atual
		now := time.Now()
		dataInicioTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	if dataFim != "" {
		dataFimTime, err = time.Parse("2006-01-02", dataFim)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
			return
		}
		// Ajustar hora final para o final do dia
		dataFimTime = time.Date(dataFimTime.Year(), dataFimTime.Month(), dataFimTime.Day(), 23, 59, 59, 0, dataFimTime.Location())
	} else {
		// Padrão: hoje
		now := time.Now()
		dataFimTime = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	}

	// Preparar a resposta
	response := PainelMDFEResponse{}

	// Filtro base
	baseQuery := h.db.Model(&models.MDFE{}).Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime)

	// Total de MDF-es
	baseQuery.Count(&response.TotalMDFEs)

	// Total por status
	h.db.Model(&models.MDFE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ? AND cancelado = ? AND encerrado = ?", "100", false, false).
		Count(&response.TotalAutorizados)

	h.db.Model(&models.MDFE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("encerrado = ?", true).
		Count(&response.TotalEncerrados)

	h.db.Model(&models.MDFE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("cancelado = ?", true).
		Count(&response.TotalCancelados)

	// Total de CT-es no período
	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Count(&response.TotalCTEsPeriodo)

	// Eficiência (simulação de cálculo)
	response.Eficiencia = 85.5 // Em um sistema real, calcularíamos com base no banco

	// Top Veículos
	// Em um sistema real, teríamos uma tabela de relação entre MDF-e e veículos
	// Aqui, simulamos alguns dados
	response.TopVeiculos = []TopVeiculo{
		{
			ID:              "1",
			Placa:           "ABC1234",
			TotalMDFEs:      12,
			TotalDocumentos: 42,
		},
		{
			ID:              "2",
			Placa:           "DEF5678",
			TotalMDFEs:      8,
			TotalDocumentos: 31,
		},
	}

	// Distribuição de CT-es por MDF-e
	// Em um sistema real, teríamos uma tabela de relação entre MDF-e e CT-e
	// Aqui, simulamos alguns dados
	response.DistribuicaoCTEs = []DistribuicaoCTEs{
		{
			MDFEChave:      "31234567890123456789012345678901234567890123",
			Numero:         123456,
			QuantidadeCTEs: 8,
		},
		{
			MDFEChave:      "31234567890123456789012345678901234567890124",
			Numero:         123457,
			QuantidadeCTEs: 5,
		},
	}

	c.JSON(http.StatusOK, response)
}
