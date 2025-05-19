package financeiro

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// FinanceiroHandler contém os handlers relacionados ao financeiro
type FinanceiroHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewFinanceiroHandler cria uma nova instância de FinanceiroHandler
func NewFinanceiroHandler(db *gorm.DB) *FinanceiroHandler {
	return &FinanceiroHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// FinanceiroRequest representa os parâmetros de request para dados financeiros
type FinanceiroRequest struct {
	Periodo     string `form:"periodo" binding:"omitempty,oneof=mes trimestre ano 7dias 30dias personalizado"`
	DataInicio  string `form:"data_inicio" binding:"omitempty"`
	DataFim     string `form:"data_fim" binding:"omitempty"`
	Agrupamento string `form:"agrupamento" binding:"omitempty,oneof=cliente veiculo distribuidora"`
}

// GetDadosFinanceiros retorna os dados do painel financeiro
func (h *FinanceiroHandler) GetDadosFinanceiros(c *gin.Context) {
	var req FinanceiroRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valores padrão
	if req.Agrupamento == "" {
		req.Agrupamento = "cliente"
	}

	// Estatísticas principais
	var totalFaturamento float64
	var totalCTEs int64
	var valorCIF, valorFOB float64

	// Total de faturamento e CT-es
	query := h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false)

	if err := query.Select("COALESCE(SUM(valor_total), 0)").Scan(&totalFaturamento).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular faturamento total")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular dados financeiros"})
		return
	}

	if err := query.Count(&totalCTEs).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar total de CT-es")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular dados financeiros"})
		return
	}

	// Valores CIF e FOB
	if err := h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false).
		Where("modalidade_frete = ?", "CIF").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&valorCIF).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular valor CIF")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular dados financeiros"})
		return
	}

	if err := h.db.Model(&models.CTE{}).
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
		Where("cancelado = ?", false).
		Where("modalidade_frete = ?", "FOB").
		Select("COALESCE(SUM(valor_total), 0)").
		Scan(&valorFOB).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao calcular valor FOB")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular dados financeiros"})
		return
	}

	// Ticket médio
	ticketMedio := float64(0)
	if totalCTEs > 0 {
		ticketMedio = totalFaturamento / float64(totalCTEs)
	}

	// Percentuais CIF/FOB
	percentCIF, percentFOB := 0.0, 0.0
	if totalFaturamento > 0 {
		percentCIF = (valorCIF / totalFaturamento) * 100
		percentFOB = (valorFOB / totalFaturamento) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"faturamento_total": totalFaturamento,
		"total_ctes":        totalCTEs,
		"ticket_medio":      ticketMedio,
		"valor_cif":         valorCIF,
		"valor_fob":         valorFOB,
		"percent_cif":       percentCIF,
		"percent_fob":       percentFOB,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
	})
}

// GetFaturamentoMensal retorna o faturamento mensal para o gráfico
func (h *FinanceiroHandler) GetFaturamentoMensal(c *gin.Context) {
	var req FinanceiroRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Estrutura para armazenar os resultados
	type DadosMensais struct {
		Mes         int     `json:"mes"`
		Ano         int     `json:"ano"`
		ValorCIF    float64 `json:"valor_cif"`
		ValorFOB    float64 `json:"valor_fob"`
		ValorTotal  float64 `json:"valor_total"`
		QtdEntregas int64   `json:"qtd_entregas"`
	}

	var dados []DadosMensais

	// Consulta SQL para agrupar por mês
	query := `
        SELECT 
            EXTRACT(YEAR FROM data_emissao) AS ano,
            EXTRACT(MONTH FROM data_emissao) AS mes,
            SUM(CASE WHEN modalidade_frete = 'CIF' THEN valor_total ELSE 0 END) AS valor_cif,
            SUM(CASE WHEN modalidade_frete = 'FOB' THEN valor_total ELSE 0 END) AS valor_fob,
            SUM(valor_total) AS valor_total,
            COUNT(*) AS qtd_entregas
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
        GROUP BY ano, mes
        ORDER BY ano, mes
    `

	if err := h.db.Raw(query, dataInicio, dataFim).Scan(&dados).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar dados de faturamento mensal")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados de faturamento"})
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

// GetDadosAgrupados retorna dados agrupados conforme o tipo selecionado
func (h *FinanceiroHandler) GetDadosAgrupados(c *gin.Context) {
	var req FinanceiroRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Valores padrão
	if req.Agrupamento == "" {
		req.Agrupamento = "cliente"
	}

	// Estrutura base para os resultados
	type ResultadoAgrupado struct {
		ID          string  `json:"id"`
		Nome        string  `json:"nome"`
		Total       float64 `json:"total"`
		QtdCTEs     int64   `json:"qtd_ctes"`
		TicketMedio float64 `json:"ticket_medio"`
	}

	var resultados []ResultadoAgrupado
	var total int64
	var query string
	var countQuery string

	// Construir query baseada no tipo de agrupamento
	switch req.Agrupamento {
	case "cliente":
		// Agrupar por cliente (destinatário)
		query = `
            SELECT 
                e.id,
                e.razao_social AS nome,
                SUM(c.valor_total) AS total,
                COUNT(c.id) AS qtd_ctes
            FROM ctes c
            JOIN empresas e ON c.destinatario_id = e.id
            WHERE c.data_emissao BETWEEN ? AND ?
            AND c.cancelado = false
            GROUP BY e.id, e.razao_social
            ORDER BY total DESC
            LIMIT ? OFFSET ?
        `
		countQuery = `
            SELECT COUNT(DISTINCT c.destinatario_id)
            FROM ctes c
            WHERE c.data_emissao BETWEEN ? AND ?
            AND c.cancelado = false
        `
	case "veiculo":
		// Agrupar por veículo
		// Em um sistema real, teríamos uma relação entre CT-e e veículo
		// Aqui, simulamos uma consulta
		query = `
            SELECT 
                v.id,
                v.placa AS nome,
                SUM(c.valor_total) AS total,
                COUNT(c.id) AS qtd_ctes
            FROM ctes c
            JOIN veiculos v ON v.id = ? -- simulado
            WHERE c.data_emissao BETWEEN ? AND ?
            AND c.cancelado = false
            GROUP BY v.id, v.placa
            ORDER BY total DESC
            LIMIT ? OFFSET ?
        `
		countQuery = `
            SELECT COUNT(DISTINCT v.id)
            FROM veiculos v
        `
	case "distribuidora":
		// Agrupar por distribuidora (emitente)
		query = `
            SELECT 
                e.id,
                e.razao_social AS nome,
                SUM(c.valor_total) AS total,
                COUNT(c.id) AS qtd_ctes
            FROM ctes c
            JOIN empresas e ON c.emitente_id = e.id
            WHERE c.data_emissao BETWEEN ? AND ?
            AND c.cancelado = false
            GROUP BY e.id, e.razao_social
            ORDER BY total DESC
            LIMIT ? OFFSET ?
        `
		countQuery = `
            SELECT COUNT(DISTINCT c.emitente_id)
            FROM ctes c
            WHERE c.data_emissao BETWEEN ? AND ?
            AND c.cancelado = false
        `
	}

	// Executar consulta paginada
	if req.Agrupamento == "veiculo" {
		// Para a simulação de veículo, usamos um ID fake
		if err := h.db.Raw(query, "veiculo-id-fake", dataInicio, dataFim, limit, offset).Scan(&resultados).Error; err != nil {
			h.logger.Error().Err(err).Msg("Erro ao buscar dados agrupados")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados agrupados"})
			return
		}
		total = 10 // Valor simulado
	} else {
		if err := h.db.Raw(query, dataInicio, dataFim, limit, offset).Scan(&resultados).Error; err != nil {
			h.logger.Error().Err(err).Msg("Erro ao buscar dados agrupados")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados agrupados"})
			return
		}

		if err := h.db.Raw(countQuery, dataInicio, dataFim).Scan(&total).Error; err != nil {
			h.logger.Error().Err(err).Msg("Erro ao contar total de registros")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar total de registros"})
			return
		}
	}

	// Calcular ticket médio para cada resultado
	for i := range resultados {
		if resultados[i].QtdCTEs > 0 {
			resultados[i].TicketMedio = resultados[i].Total / float64(resultados[i].QtdCTEs)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resultados,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
		"agrupamento": req.Agrupamento,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
	})
}

// GetDetalheItem retorna detalhes de um item específico do agrupamento
func (h *FinanceiroHandler) GetDetalheItem(c *gin.Context) {
	id := c.Param("id")
	tipo := c.Param("tipo")

	var req FinanceiroRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates(req.Periodo, req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Diferentes detalhes baseados no tipo
	switch tipo {
	case "cliente":
		// Buscar detalhes do cliente
		var cliente models.Empresa
		if err := h.db.First(&cliente, "id = ?", id).Error; err != nil {
			h.logger.Error().Err(err).Str("id", id).Msg("Cliente não encontrado")
			c.JSON(http.StatusNotFound, gin.H{"error": "Cliente não encontrado"})
			return
		}

		// Buscar CT-es do cliente
		var ctes []models.CTE
		if err := h.db.Where("destinatario_id = ?", id).
			Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
			Where("cancelado = ?", false).
			Order("data_emissao DESC").
			Limit(20).
			Find(&ctes).Error; err != nil {
			h.logger.Error().Err(err).Str("id", id).Msg("Erro ao buscar CT-es do cliente")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar detalhes"})
			return
		}

		// Calcular valores
		var totalCTEs int64
		var valorTotal, valorCIF, valorFOB float64

		h.db.Model(&models.CTE{}).
			Where("destinatario_id = ?", id).
			Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
			Where("cancelado = ?", false).
			Count(&totalCTEs)

		h.db.Model(&models.CTE{}).
			Where("destinatario_id = ?", id).
			Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
			Where("cancelado = ?", false).
			Select("COALESCE(SUM(valor_total), 0)").
			Scan(&valorTotal)

		h.db.Model(&models.CTE{}).
			Where("destinatario_id = ?", id).
			Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
			Where("cancelado = ?", false).
			Where("modalidade_frete = ?", "CIF").
			Select("COALESCE(SUM(valor_total), 0)").
			Scan(&valorCIF)

		h.db.Model(&models.CTE{}).
			Where("destinatario_id = ?", id).
			Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim).
			Where("cancelado = ?", false).
			Where("modalidade_frete = ?", "FOB").
			Select("COALESCE(SUM(valor_total), 0)").
			Scan(&valorFOB)

		c.JSON(http.StatusOK, gin.H{
			"cliente": gin.H{
				"id":           cliente.ID,
				"razao_social": cliente.RazaoSocial,
				"cnpj":         cliente.CNPJ,
				"cpf":          cliente.CPF,
				"uf":           cliente.UF,
			},
			"ctes":        ctes,
			"total_ctes":  totalCTEs,
			"valor_total": valorTotal,
			"valor_cif":   valorCIF,
			"valor_fob":   valorFOB,
			"periodo": gin.H{
				"data_inicio": dataInicio.Format("2006-01-02"),
				"data_fim":    dataFim.Format("2006-01-02"),
			},
		})

	case "veiculo":
		// Consultar detalhes do veículo (simulado para exemplo)
		c.JSON(http.StatusOK, gin.H{
			"veiculo": gin.H{
				"id":    id,
				"placa": "ABC1234",
				"tipo":  "PROPRIO",
			},
			"total_ctes":    15,
			"valor_total":   25000.50,
			"km_percorrido": 1200,
			"periodo": gin.H{
				"data_inicio": dataInicio.Format("2006-01-02"),
				"data_fim":    dataFim.Format("2006-01-02"),
			},
		})

	case "distribuidora":
		// Consultar detalhes da distribuidora (similar ao cliente)
		c.JSON(http.StatusOK, gin.H{
			"distribuidora": gin.H{
				"id":           id,
				"razao_social": "Distribuidora Exemplo",
				"cnpj":         "12345678901234",
			},
			"total_ctes":  10,
			"valor_total": 15000.75,
			"periodo": gin.H{
				"data_inicio": dataInicio.Format("2006-01-02"),
				"data_fim":    dataFim.Format("2006-01-02"),
			},
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo inválido"})
	}
}

// Helper para obter datas de início e fim baseado no período
func getPeriodDates(periodo, dataInicioStr, dataFimStr string) (time.Time, time.Time, error) {
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
