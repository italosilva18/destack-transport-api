package geografico

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// GeograficoHandler contém os handlers relacionados a análises geográficas
type GeograficoHandler struct {
	db     *gorm.DB
	logger zerolog.Logger
}

// NewGeograficoHandler cria uma nova instância de GeograficoHandler
func NewGeograficoHandler(db *gorm.DB) *GeograficoHandler {
	return &GeograficoHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// GeograficoRequest representa os parâmetros de request para dados geográficos
type GeograficoRequest struct {
	DataInicio string `form:"data_inicio" binding:"omitempty"`
	DataFim    string `form:"data_fim" binding:"omitempty"`
	UF         string `form:"uf" binding:"omitempty,len=2"`
}

// GetDadosGeograficos retorna os dados do painel geográfico
func (h *GeograficoHandler) GetDadosGeograficos(c *gin.Context) {
	var req GeograficoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates("", req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Filtro base
	baseQuery := h.db.Model(&models.CTE{}).Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim)

	// Filtro adicional por UF se fornecido
	if req.UF != "" {
		baseQuery = baseQuery.Where("uf_inicio = ? OR uf_destino = ?", req.UF, req.UF)
	}

	// Contar origens, destinos e rotas únicas
	var totalOrigens, totalDestinos, totalRotas int64

	// Origem: combinação única de UF_inicio e municipio_inicio
	queryOrigens := h.db.Table("ctes").
		Select("DISTINCT uf_inicio, municipio_inicio").
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim)

	if req.UF != "" {
		queryOrigens = queryOrigens.Where("uf_inicio = ?", req.UF)
	}

	queryOrigens.Count(&totalOrigens)

	// Destino: combinação única de UF_destino e municipio_fim
	queryDestinos := h.db.Table("ctes").
		Select("DISTINCT uf_destino, municipio_fim").
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim)

	if req.UF != "" {
		queryDestinos = queryDestinos.Where("uf_destino = ?", req.UF)
	}

	queryDestinos.Count(&totalDestinos)

	// Rotas: combinação única de origem e destino
	queryRotas := h.db.Table("ctes").
		Select("DISTINCT uf_inicio, municipio_inicio, uf_destino, municipio_fim").
		Where("data_emissao BETWEEN ? AND ?", dataInicio, dataFim)

	if req.UF != "" {
		queryRotas = queryRotas.Where("uf_inicio = ? OR uf_destino = ?", req.UF, req.UF)
	}

	queryRotas.Count(&totalRotas)

	c.JSON(http.StatusOK, gin.H{
		"total_origens":  totalOrigens,
		"total_destinos": totalDestinos,
		"total_rotas":    totalRotas,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
		"filtro_uf": req.UF,
	})
}

// TopOrigem representa uma origem no ranking
type TopOrigem struct {
	UF         string  `json:"uf"`
	Municipio  string  `json:"municipio"`
	Quantidade int64   `json:"quantidade"`
	ValorTotal float64 `json:"valor_total"`
}

// GetTopOrigens retorna as principais origens
func (h *GeograficoHandler) GetTopOrigens(c *gin.Context) {
	var req GeograficoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates("", req.DataInicio, req.DataFim)
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

	// Consulta para top origens
	query := `
        SELECT 
            uf_inicio AS uf,
            municipio_inicio AS municipio,
            COUNT(*) AS quantidade,
            COALESCE(SUM(valor_total), 0) AS valor_total
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
    `

	countQuery := `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT uf_inicio, municipio_inicio
            FROM ctes
            WHERE data_emissao BETWEEN ? AND ?
            AND cancelado = false
    `

	params := []interface{}{dataInicio, dataFim}

	// Adicionar filtro de UF se fornecido
	if req.UF != "" {
		query += " AND uf_inicio = ?"
		countQuery += " AND uf_inicio = ?"
		params = append(params, req.UF)
	}

	query += `
        GROUP BY uf_inicio, municipio_inicio
        ORDER BY quantidade DESC, valor_total DESC
        LIMIT ? OFFSET ?
    `
	countQuery += ") AS t"

	// Adicionar parâmetros para paginação
	params = append(params, limit, offset)

	var origens []TopOrigem
	if err := h.db.Raw(query, params...).Scan(&origens).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar top origens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar top origens"})
		return
	}

	// Contar total para paginação
	var total int64
	if err := h.db.Raw(countQuery, params[:len(params)-2]...).Scan(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar total de origens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar total de origens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": origens,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
		"filtro_uf": req.UF,
	})
}

// TopDestino representa um destino no ranking
type TopDestino struct {
	UF         string  `json:"uf"`
	Municipio  string  `json:"municipio"`
	Quantidade int64   `json:"quantidade"`
	ValorTotal float64 `json:"valor_total"`
}

// GetTopDestinos retorna os principais destinos
func (h *GeograficoHandler) GetTopDestinos(c *gin.Context) {
	var req GeograficoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates("", req.DataInicio, req.DataFim)
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

	// Consulta para top destinos
	query := `
        SELECT 
            uf_destino AS uf,
            municipio_fim AS municipio,
            COUNT(*) AS quantidade,
            COALESCE(SUM(valor_total), 0) AS valor_total
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
    `

	countQuery := `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT uf_destino, municipio_fim
            FROM ctes
            WHERE data_emissao BETWEEN ? AND ?
            AND cancelado = false
    `

	params := []interface{}{dataInicio, dataFim}

	// Adicionar filtro de UF se fornecido
	if req.UF != "" {
		query += " AND uf_destino = ?"
		countQuery += " AND uf_destino = ?"
		params = append(params, req.UF)
	}

	query += `
        GROUP BY uf_destino, municipio_fim
        ORDER BY quantidade DESC, valor_total DESC
        LIMIT ? OFFSET ?
    `
	countQuery += ") AS t"

	// Adicionar parâmetros para paginação
	params = append(params, limit, offset)

	var destinos []TopDestino
	if err := h.db.Raw(query, params...).Scan(&destinos).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar top destinos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar top destinos"})
		return
	}

	// Contar total para paginação
	var total int64
	if err := h.db.Raw(countQuery, params[:len(params)-2]...).Scan(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar total de destinos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar total de destinos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": destinos,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
		"filtro_uf": req.UF,
	})
}

// RotaFrequente representa uma rota no ranking
type RotaFrequente struct {
	UFOrigem         string  `json:"uf_origem"`
	MunicipioOrigem  string  `json:"municipio_origem"`
	UFDestino        string  `json:"uf_destino"`
	MunicipioDestino string  `json:"municipio_destino"`
	Quantidade       int64   `json:"quantidade"`
	ValorTotal       float64 `json:"valor_total"`
	KMTotal          float64 `json:"km_total"`
}

// GetRotasFrequentes retorna as rotas mais frequentes
func (h *GeograficoHandler) GetRotasFrequentes(c *gin.Context) {
	var req GeograficoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates("", req.DataInicio, req.DataFim)
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

	// Consulta para rotas frequentes
	query := `
        SELECT 
            uf_inicio AS uf_origem,
            municipio_inicio AS municipio_origem,
            uf_destino AS uf_destino,
            municipio_fim AS municipio_destino,
            COUNT(*) AS quantidade,
            COALESCE(SUM(valor_total), 0) AS valor_total,
            COUNT(*) * 100 AS km_total
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
    `

	countQuery := `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT uf_inicio, municipio_inicio, uf_destino, municipio_fim
            FROM ctes
            WHERE data_emissao BETWEEN ? AND ?
            AND cancelado = false
    `

	params := []interface{}{dataInicio, dataFim}

	// Adicionar filtro de UF se fornecido
	if req.UF != "" {
		query += " AND (uf_inicio = ? OR uf_destino = ?)"
		countQuery += " AND (uf_inicio = ? OR uf_destino = ?)"
		params = append(params, req.UF, req.UF)
	}

	query += `
        GROUP BY uf_inicio, municipio_inicio, uf_destino, municipio_fim
        ORDER BY quantidade DESC, valor_total DESC
        LIMIT ? OFFSET ?
    `
	countQuery += ") AS t"

	// Adicionar parâmetros para paginação
	params = append(params, limit, offset)

	var rotas []RotaFrequente
	if err := h.db.Raw(query, params...).Scan(&rotas).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar rotas frequentes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar rotas frequentes"})
		return
	}

	// Contar total para paginação
	var total int64
	if err := h.db.Raw(countQuery, params[:len(params)-2]...).Scan(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar total de rotas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar total de rotas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rotas,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
		"filtro_uf": req.UF,
	})
}

// FluxoUF representa o fluxo entre UFs
type FluxoUF struct {
	UFOrigem   string  `json:"uf_origem"`
	UFDestino  string  `json:"uf_destino"`
	Quantidade int64   `json:"quantidade"`
	ValorTotal float64 `json:"valor_total"`
}

// GetFluxoUFs retorna o fluxo entre UFs para o gráfico
func (h *GeograficoHandler) GetFluxoUFs(c *gin.Context) {
	var req GeograficoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determinar período
	dataInicio, dataFim, err := getPeriodDates("", req.DataInicio, req.DataFim)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Consulta para fluxo entre UFs
	query := `
        SELECT 
            uf_inicio AS uf_origem,
            uf_destino AS uf_destino,
            COUNT(*) AS quantidade,
            COALESCE(SUM(valor_total), 0) AS valor_total
        FROM ctes
        WHERE data_emissao BETWEEN ? AND ?
        AND cancelado = false
    `

	params := []interface{}{dataInicio, dataFim}

	// Adicionar filtro de UF se fornecido
	if req.UF != "" {
		query += " AND (uf_inicio = ? OR uf_destino = ?)"
		params = append(params, req.UF, req.UF)
	}

	query += `
        GROUP BY uf_inicio, uf_destino
        ORDER BY quantidade DESC
    `

	var fluxos []FluxoUF
	if err := h.db.Raw(query, params...).Scan(&fluxos).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao buscar fluxo entre UFs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo entre UFs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": fluxos,
		"periodo": gin.H{
			"data_inicio": dataInicio.Format("2006-01-02"),
			"data_fim":    dataFim.Format("2006-01-02"),
		},
		"filtro_uf": req.UF,
	})
}

// Helper para obter datas de início e fim baseado no período
func getPeriodDates(periodo, dataInicioStr, dataFimStr string) (time.Time, time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// Se as datas forem fornecidas diretamente
	if dataInicioStr != "" && dataFimStr != "" {
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

	// Padrão: último mês
	dataInicio := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	dataInicio = dataInicio.AddDate(0, -1, 0) // Mês anterior

	return dataInicio, today, nil
}
