package dashboard

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// DashboardHandler contém os handlers para o dashboard
type DashboardHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewDashboardHandler cria uma nova instância de DashboardHandler
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// DashboardRequest representa os parâmetros de request para o dashboard
type DashboardRequest struct {
	Periodo    string `form:"periodo" binding:"omitempty,oneof=mes trimestre ano 7dias 30dias personalizado"`
	DataInicio string `form:"data_inicio" binding:"omitempty"`
	DataFim    string `form:"data_fim" binding:"omitempty"`
}

// GetDashboardCards retorna os dados dos cards do dashboard
func (h *DashboardHandler) GetDashboardCards(c *gin.Context) {
	var req DashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := h.getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Estatísticas de CT-e
	var totalCTe int64
	var valorTotalCTe float64
	var valorCIF float64
	var valorFOB float64

	// Contar total de CT-es no período
	query := h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false)

	err = query.Count(&totalCTe).Error
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar CT-es")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados"})
		return
	}

	// Somar valor total
	err = query.Select("COALESCE(SUM(valor_total), 0)").Scan(&valorTotalCTe).Error
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular valor total de CT-es")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados"})
		return
	}

	// Valor CIF
	err = h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false).
		Where("modalidade_frete = ?", "CIF").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&valorCIF).Error
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular valor CIF")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados"})
		return
	}

	// Valor FOB
	err = h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false).
		Where("modalidade_frete = ?", "FOB").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&valorFOB).Error
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular valor FOB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados"})
		return
	}

	// Estatísticas de MDF-e
	var totalMDFe int64
	err = h.db.Model(&models.MDFE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false).
		Count(&totalMDFe).Error
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar MDF-es")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_ctes":      totalCTe,
		"valor_total_cte": valorTotalCTe,
		"valor_cif":       valorCIF,
		"valor_fob":       valorFOB,
		"total_mdfe":      totalMDFe,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
	})
}

// Helper para obter datas de início e fim baseado no período
func (h *DashboardHandler) getPeriodDates(periodo, dataInicioStr, dataFimStr string) (time.Time, time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// Se as datas forem fornecidas diretamente
	if periodo == "personalizado" && dataInicioStr != "" && dataFimStr != "" {
		dataInicio, err := time.Parse("2006-01-02", dataInicioStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		dataFim, err := time.Parse("2006-01-02", dataFimStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// Continuação do método getPeriodDates no dashboard_handler.go
		// Ajustar hora final para o final do dia
		dataFim = time.Date(dataFim.Year(), dataFim.Month(), dataFim.Day(), 23, 59, 59, 0, dataFim.Location())
		return dataInicio, dataFim, nil
	}

	// Calcular período com base na opção selecionada
	var dataInicio time.Time

	switch periodo {
	case "mes", "": // Mês atual é o padrão
		dataInicio = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "trimestre":
		currentQuarter := (int(now.Month()) - 1) / 3
		startMonth := time.Month(currentQuarter*3 + 1)
		dataInicio = time.Date(now.Year(), startMonth, 1, 0, 0, 0, 0, now.Location())
	case "ano":
		dataInicio = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case "7dias":
		dataInicio = today.AddDate(0, 0, -7)
	case "30dias":
		dataInicio = today.AddDate(0, 0, -30)
	default:
		// Padrão para último mês
		dataInicio = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}

	return dataInicio, today, nil
}

// GetUltimosLancamentos retorna os últimos lançamentos para o dashboard
func (h *DashboardHandler) GetUltimosLancamentos(c *gin.Context) {
	// Definir limite de registros
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Estrutura para armazenar os resultados
	type Lancamento struct {
		ID          string    `json:"id"`
		Chave       string    `json:"chave"`
		Tipo        string    `json:"tipo"`
		Numero      int       `json:"numero"`
		DataEmissao time.Time `json:"data_emissao"`
		ValorTotal  float64   `json:"valor_total"`
		Origem      string    `json:"origem"`
		Destino     string    `json:"destino"`
		Status      string    `json:"status"`
	}

	var lancamentos []Lancamento

	// Buscar últimos CT-es
	var ctes []models.CTE
	if err := h.db.Preload("Emitente").Order("data_emissao DESC").Limit(limit).Find(&ctes).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar últimos CT-es")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar lançamentos"})
		return
	}

	// Converter CT-es para formato de lançamento
	for _, cte := range ctes {
		lancamentos = append(lancamentos, Lancamento{
			ID:          cte.ID.String(),
			Chave:       cte.Chave,
			Tipo:        "CTE",
			Numero:      cte.Numero,
			DataEmissao: cte.DataEmissao,
			ValorTotal:  cte.ValorTotal,
			Origem:      cte.MunicipioInicio,
			Destino:     cte.MunicipioFim,
			Status:      cte.Status,
		})
	}

	// Buscar últimos MDF-es (similar ao CT-e)
	// Por simplicidade, omitimos o código de busca de MDF-es

	// Ordenar todos os lançamentos por data
	sort.Slice(lancamentos, func(i, j int) bool {
		return lancamentos[i].DataEmissao.After(lancamentos[j].DataEmissao)
	})

	// Limitar ao número solicitado
	if len(lancamentos) > limit {
		lancamentos = lancamentos[:limit]
	}

	c.JSON(http.StatusOK, lancamentos)
}

// GetCifFobData retorna dados para o gráfico de evolução CIF/FOB
func (h *DashboardHandler) GetCifFobData(c *gin.Context) {
	// Similar ao GetDashboardCards, mas com agrupamento por mês
	var req DashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := h.getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Estrutura para armazenar os resultados
	type DadosMensais struct {
		Mes        int     `json:"mes"`
		Ano        int     `json:"ano"`
		ValorCIF   float64 `json:"valor_cif"`
		ValorFOB   float64 `json:"valor_fob"`
		ValorTotal float64 `json:"valor_total"`
	}

	var dados []DadosMensais

	// Consulta SQL para agrupar por mês
	query := `
        SELECT 
            EXTRACT(YEAR FROM data_emissao) AS ano,
            EXTRACT(MONTH FROM data_emissao) AS mes,
            SUM(CASE WHEN modalidade_frete = 'CIF' THEN valor_total ELSE 0 END) AS valor_cif,
            SUM(CASE WHEN modalidade_frete = 'FOB' THEN valor_total ELSE 0 END) AS valor_fob,
            SUM(valor_total) AS valor_total
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
        GROUP BY ano, mes
        ORDER BY ano, mes
    `

	if err := h.db.Raw(query, dataInicio, dataFim).Scan(&dados).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar dados CIF/FOB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados para o gráfico"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dados": dados,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
	})
}
