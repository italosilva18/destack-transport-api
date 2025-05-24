package cte

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// CTEHandler contém os handlers para CTEs
type CTEHandler struct {
	db     *gorm.DB
	logger zerolog.Logger
}

// NewCTEHandler cria uma nova instância de CTEHandler
func NewCTEHandler(db *gorm.DB) *CTEHandler {
	return &CTEHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// ListCTEsRequest representa os parâmetros de request para listar CTEs
type ListCTEsRequest struct {
	Page           int    `form:"page" binding:"omitempty,min=1"`
	Limit          int    `form:"limit" binding:"omitempty,min=1,max=100"`
	DataInicio     string `form:"data_inicio" binding:"omitempty"`
	DataFim        string `form:"data_fim" binding:"omitempty"`
	Modalidade     string `form:"modalidade" binding:"omitempty,oneof=CIF FOB"`
	Status         string `form:"status" binding:"omitempty"`
	NumeroDoc      string `form:"numero_doc" binding:"omitempty"`
	EmitenteID     string `form:"emitente_id" binding:"omitempty"`
	DestinatarioID string `form:"destinatario_id" binding:"omitempty"`
}

// ListCTEs lista os CTEs com filtros e paginação
func (h *CTEHandler) ListCTEs(c *gin.Context) {
	var req ListCTEsRequest
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
	query := h.db.Model(&models.CTE{}).Preload("Emitente").Preload("Destinatario").Preload("Remetente")

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

	if req.Modalidade != "" {
		query = query.Where("modalidade_frete = ?", req.Modalidade)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.NumeroDoc != "" {
		query = query.Where("numero = ?", req.NumeroDoc)
	}

	if req.EmitenteID != "" {
		query = query.Where("emitente_id = ?", req.EmitenteID)
	}

	if req.DestinatarioID != "" {
		query = query.Where("destinatario_id = ?", req.DestinatarioID)
	}

	// Contar total para paginação
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar CTEs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar CTEs"})
		return
	}

	// Buscar CTEs com paginação
	var ctes []models.CTE
	if err := query.Offset(offset).Limit(limit).Order("data_emissao DESC").Find(&ctes).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao listar CTEs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar CTEs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ctes,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetCTE obtém um CTE pelo ID ou chave
func (h *CTEHandler) GetCTE(c *gin.Context) {
	chave := c.Param("chave")

	var cte models.CTE
	result := h.db.Preload("Emitente").Preload("Destinatario").Preload("Remetente").Preload("Tomador").Where("chave = ?", chave).First(&cte)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("chave", chave).Msg("CTE não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "CTE não encontrado"})
		return
	}

	c.JSON(http.StatusOK, cte)
}

// DownloadXML baixa o XML de um CTE
func (h *CTEHandler) DownloadXML(c *gin.Context) {
	chave := c.Param("chave")

	// Em um sistema real, recuperaríamos o XML armazenado
	// Aqui, simulamos uma resposta com um XML simples
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<CTe xmlns="http://www.portalfiscal.inf.br/cte">
  <infCte Id="CTe` + chave + `">
    <!-- conteúdo do CT-e aqui -->
  </infCte>
</CTe>`

	c.Header("Content-Disposition", "attachment; filename=cte_"+chave+".xml")
	c.Data(http.StatusOK, "application/xml", []byte(xml))
}

// GerarDACTE gera um DACTE em PDF
func (h *CTEHandler) GerarDACTE(c *gin.Context) {
	chave := c.Param("chave")

	// Em um sistema real, geraria o PDF
	// Aqui, simulamos uma resposta com um PDF simples
	c.Header("Content-Disposition", "attachment; filename=dacte_"+chave+".pdf")
	c.Data(http.StatusOK, "application/pdf", []byte("Simulação de PDF do DACTE"))
}

// Reprocessar reprocessa um CTE
func (h *CTEHandler) Reprocessar(c *gin.Context) {
	chave := c.Param("chave")

	var cte models.CTE
	result := h.db.Where("chave = ?", chave).First(&cte)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("chave", chave).Msg("CTE não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "CTE não encontrado"})
		return
	}

	// Em um sistema real, faria o reprocessamento
	// Aqui, apenas atualizamos um campo para simular
	if err := h.db.Model(&cte).Update("data_processamento", time.Now()).Error; err != nil {
		h.logger.Error().Err(err).Str("chave", chave).Msg("Erro ao reprocessar CTE")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao reprocessar CTE"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CTE reprocessado com sucesso",
		"chave":   chave,
	})
}

// PainelCTEResponse representa a resposta para o painel de CT-e
type PainelCTEResponse struct {
	TotalCTEs          int64              `json:"total_ctes"`
	ValorTotal         float64            `json:"valor_total"`
	ValorCIF           float64            `json:"valor_cif"`
	ValorFOB           float64            `json:"valor_fob"`
	TotalAutorizados   int64              `json:"total_autorizados"`
	TotalCancelados    int64              `json:"total_cancelados"`
	TotalRejeitados    int64              `json:"total_rejeitados"`
	TopClientes        []TopCliente       `json:"top_clientes"`
	DistribuicaoCIFFOB DistribuicaoCIFFOB `json:"distribuicao_cif_fob"`
}

// TopCliente representa um cliente no ranking
type TopCliente struct {
	ID             string  `json:"id"`
	Nome           string  `json:"nome"`
	CNPJ           *string `json:"cnpj"`
	CPF            *string `json:"cpf"`
	QuantidadeCTEs int64   `json:"quantidade_ctes"`
	ValorTotal     float64 `json:"valor_total"`
	TicketMedio    float64 `json:"ticket_medio"`
}

// DistribuicaoCIFFOB representa a distribuição entre CIF e FOB
type DistribuicaoCIFFOB struct {
	ValorCIF      float64 `json:"valor_cif"`
	ValorFOB      float64 `json:"valor_fob"`
	PercentualCIF float64 `json:"percentual_cif"`
	PercentualFOB float64 `json:"percentual_fob"`
}

// GetPainelCTE retorna os dados para o painel de CT-e
func (h *CTEHandler) GetPainelCTE(c *gin.Context) {
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
	response := PainelCTEResponse{}

	// Filtro base
	baseQuery := h.db.Model(&models.CTE{}).Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime)

	// Total de CT-es
	baseQuery.Count(&response.TotalCTEs)

	// Valor total
	baseQuery.Select("COALESCE(SUM(valor_total), 0)").Scan(&response.ValorTotal)

	// Valor CIF
	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("modalidade_frete = ?", "CIF").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&response.ValorCIF)

	// Valor FOB
	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("modalidade_frete = ?", "FOB").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&response.ValorFOB)

	// Status
	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "100").
		Count(&response.TotalAutorizados)

	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("cancelado = ?", true).
		Count(&response.TotalCancelados)

	h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status NOT IN (?, ?)", "100", "").
		Where("cancelado = ?", false).
		Count(&response.TotalRejeitados)

	// Top Clientes (Destinatários)
	type TopClienteQuery struct {
		ID             string
		Nome           string
		CNPJ           *string
		CPF            *string
		QuantidadeCTEs int64
		ValorTotal     float64
	}

	var topClientesQuery []TopClienteQuery
	h.db.Raw(`
		SELECT 
			e.id,
			e.razao_social AS nome,
			e.cnpj,
			e.cpf,
			COUNT(c.id) AS quantidade_ctes,
			COALESCE(SUM(c.valor_total), 0) AS valor_total
		FROM ctes c
		JOIN empresas e ON c.destinatario_id = e.id
		WHERE c.data_emissao BETWEEN ? AND ?
		GROUP BY e.id, e.razao_social, e.cnpj, e.cpf
		ORDER BY valor_total DESC
		LIMIT 10
	`, dataInicioTime, dataFimTime).Scan(&topClientesQuery)

	// Converter para a estrutura de resposta
	response.TopClientes = make([]TopCliente, len(topClientesQuery))
	for i, cliente := range topClientesQuery {
		ticketMedio := float64(0)
		if cliente.QuantidadeCTEs > 0 {
			ticketMedio = cliente.ValorTotal / float64(cliente.QuantidadeCTEs)
		}

		response.TopClientes[i] = TopCliente{
			ID:             cliente.ID,
			Nome:           cliente.Nome,
			CNPJ:           cliente.CNPJ,
			CPF:            cliente.CPF,
			QuantidadeCTEs: cliente.QuantidadeCTEs,
			ValorTotal:     cliente.ValorTotal,
			TicketMedio:    ticketMedio,
		}
	}

	// Distribuição CIF/FOB
	totalValor := response.ValorCIF + response.ValorFOB
	response.DistribuicaoCIFFOB = DistribuicaoCIFFOB{
		ValorCIF:      response.ValorCIF,
		ValorFOB:      response.ValorFOB,
		PercentualCIF: 0,
		PercentualFOB: 0,
	}

	if totalValor > 0 {
		response.DistribuicaoCIFFOB.PercentualCIF = (response.ValorCIF / totalValor) * 100
		response.DistribuicaoCIFFOB.PercentualFOB = (response.ValorFOB / totalValor) * 100
	}

	c.JSON(http.StatusOK, response)
}
