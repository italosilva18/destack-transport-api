package manutencao

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// ManutencaoHandler contém os handlers para manutenções
type ManutencaoHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewManutencaoHandler cria uma nova instância de ManutencaoHandler
func NewManutencaoHandler(db *gorm.DB) *ManutencaoHandler {
	return &ManutencaoHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// CreateManutencaoRequest representa os dados para criar uma manutenção
type CreateManutencaoRequest struct {
	VeiculoID        string  `json:"veiculo_id" binding:"required"`
	DataServico      string  `json:"data_servico" binding:"required"`
	ServicoRealizado string  `json:"servico_realizado" binding:"required"`
	Oficina          string  `json:"oficina"`
	Quilometragem    *int    `json:"quilometragem"`
	PecaUtilizada    string  `json:"peca_utilizada"`
	NotaFiscal       string  `json:"nota_fiscal"`
	ValorPeca        float64 `json:"valor_peca"`
	ValorMaoObra     float64 `json:"valor_mao_obra"`
	Status           string  `json:"status" binding:"required,oneof=PENDENTE AGENDADO CONCLUIDO PAGO CANCELADO"`
	Observacoes      string  `json:"observacoes"`
}

// CreateManutencao cria uma nova manutenção
func (h *ManutencaoHandler) CreateManutencao(c *gin.Context) {
	var req CreateManutencaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se o veículo existe
	var veiculo models.Veiculo
	result := h.db.First(&veiculo, "id = ?", req.VeiculoID)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("veiculo_id", req.VeiculoID).Msg("Veículo não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "Veículo não encontrado"})
		return
	}

	// Parsear data
	dataServico, err := time.Parse("2006-01-02", req.DataServico)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
		return
	}

	// Criar manutenção
	manutencao := models.Manutencao{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		VeiculoID:        req.VeiculoID,
		DataServico:      dataServico,
		ServicoRealizado: req.ServicoRealizado,
		Oficina:          req.Oficina,
		Quilometragem:    req.Quilometragem,
		PecaUtilizada:    req.PecaUtilizada,
		NotaFiscal:       req.NotaFiscal,
		ValorPeca:        req.ValorPeca,
		ValorMaoObra:     req.ValorMaoObra,
		Status:           req.Status,
		Observacoes:      req.Observacoes,
	}

	if err := h.db.Create(&manutencao).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao criar manutenção")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar manutenção"})
		return
	}

	c.JSON(http.StatusCreated, manutencao)
}

// UpdateManutencaoRequest representa os dados para atualizar uma manutenção
type UpdateManutencaoRequest struct {
	VeiculoID        string  `json:"veiculo_id"`
	DataServico      string  `json:"data_servico"`
	ServicoRealizado string  `json:"servico_realizado"`
	Oficina          string  `json:"oficina"`
	Quilometragem    *int    `json:"quilometragem"`
	PecaUtilizada    string  `json:"peca_utilizada"`
	NotaFiscal       string  `json:"nota_fiscal"`
	ValorPeca        float64 `json:"valor_peca"`
	ValorMaoObra     float64 `json:"valor_mao_obra"`
	Status           string  `json:"status" binding:"omitempty,oneof=PENDENTE AGENDADO CONCLUIDO PAGO CANCELADO"`
	Observacoes      string  `json:"observacoes"`
}

// UpdateManutencao atualiza uma manutenção existente
func (h *ManutencaoHandler) UpdateManutencao(c *gin.Context) {
	id := c.Param("id")

	var manutencao models.Manutencao
	result := h.db.First(&manutencao, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Manutenção não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Manutenção não encontrada"})
		return
	}

	var req UpdateManutencaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preparar atualizações
	updates := map[string]interface{}{}

	if req.VeiculoID != "" {
		// Verificar se o veículo existe
		var veiculo models.Veiculo
		result := h.db.First(&veiculo, "id = ?", req.VeiculoID)
		if result.Error != nil {
			h.logger.Error().Err(result.Error).Str("veiculo_id", req.VeiculoID).Msg("Veículo não encontrado")
			c.JSON(http.StatusNotFound, gin.H{"error": "Veículo não encontrado"})
			return
		}
		updates["veiculo_id"] = req.VeiculoID
	}

	if req.DataServico != "" {
		// Parsear data
		dataServico, err := time.Parse("2006-01-02", req.DataServico)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido"})
			return
		}
		updates["data_servico"] = dataServico
	}

	if req.ServicoRealizado != "" {
		updates["servico_realizado"] = req.ServicoRealizado
	}

	if req.Oficina != "" {
		updates["oficina"] = req.Oficina
	}

	if req.Quilometragem != nil {
		updates["quilometragem"] = req.Quilometragem
	}

	// Continuação do arquivo internal/api/handlers/manutencao/manutencao_handler.go

	if req.PecaUtilizada != "" {
		updates["peca_utilizada"] = req.PecaUtilizada
	}

	if req.NotaFiscal != "" {
		updates["nota_fiscal"] = req.NotaFiscal
	}

	// Valores podem ser 0, então verificamos presença no JSON
	if c.Request.Method == http.MethodPut || req.ValorPeca != 0 {
		updates["valor_peca"] = req.ValorPeca
	}

	if c.Request.Method == http.MethodPut || req.ValorMaoObra != 0 {
		updates["valor_mao_obra"] = req.ValorMaoObra
	}

	if req.Status != "" {
		updates["status"] = req.Status
	}

	// Observações podem ser vazias
	if c.Request.Method == http.MethodPut || req.Observacoes != "" {
		updates["observacoes"] = req.Observacoes
	}

	// Aplicar atualizações
	if err := h.db.Model(&manutencao).Updates(updates).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao atualizar manutenção")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar manutenção"})
		return
	}

	// Buscar manutenção atualizada
	h.db.First(&manutencao, "id = ?", id)

	c.JSON(http.StatusOK, manutencao)
}

// DeleteManutencao exclui uma manutenção
func (h *ManutencaoHandler) DeleteManutencao(c *gin.Context) {
	id := c.Param("id")

	var manutencao models.Manutencao
	result := h.db.First(&manutencao, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Manutenção não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Manutenção não encontrada"})
		return
	}

	if err := h.db.Delete(&manutencao).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao excluir manutenção")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir manutenção"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manutenção excluída com sucesso"})
}

// GetManutencao obtém uma manutenção pelo ID
func (h *ManutencaoHandler) GetManutencao(c *gin.Context) {
	id := c.Param("id")

	var manutencao models.Manutencao
	result := h.db.First(&manutencao, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Manutenção não encontrada")
		c.JSON(http.StatusNotFound, gin.H{"error": "Manutenção não encontrada"})
		return
	}

	// Carregar veículo relacionado
	var veiculo models.Veiculo
	h.db.First(&veiculo, "id = ?", manutencao.VeiculoID)

	// Calcular valor total (em um sistema real, poderia estar no modelo)
	valorTotal := manutencao.ValorPeca + manutencao.ValorMaoObra

	c.JSON(http.StatusOK, gin.H{
		"manutencao":  manutencao,
		"veiculo":     veiculo,
		"valor_total": valorTotal,
	})
}

// ListManutencoesRequest representa os parâmetros para listar manutenções
type ListManutencoesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	DataInicio string `form:"data_inicio" binding:"omitempty"`
	DataFim    string `form:"data_fim" binding:"omitempty"`
	VeiculoID  string `form:"veiculo_id" binding:"omitempty"`
	Placa      string `form:"placa" binding:"omitempty"`
	Status     string `form:"status" binding:"omitempty,oneof=PENDENTE AGENDADO CONCLUIDO PAGO CANCELADO"`
	SearchText string `form:"search_text" binding:"omitempty"`
}

// ListManutencoes lista as manutenções com filtros e paginação
func (h *ManutencaoHandler) ListManutencoes(c *gin.Context) {
	var req ListManutencoesRequest
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
	query := h.db.Model(&models.Manutencao{})

	// Aplicar filtros
	if req.DataInicio != "" && req.DataFim != "" {
		dataInicio, err := time.Parse("2006-01-02", req.DataInicio)
		if err == nil {
			dataFim, err := time.Parse("2006-01-02", req.DataFim)
			if err == nil {
				// Ajustar hora final para o final do dia
				dataFim = time.Date(dataFim.Year(), dataFim.Month(), dataFim.Day(), 23, 59, 59, 0, dataFim.Location())
				query = query.Where("data_servico BETWEEN ? AND ?", dataInicio, dataFim)
			}
		}
	}

	if req.VeiculoID != "" {
		query = query.Where("veiculo_id = ?", req.VeiculoID)
	}

	if req.Placa != "" {
		query = query.Joins("JOIN veiculos v ON manutencoes.veiculo_id = v.id").
			Where("v.placa LIKE ?", "%"+req.Placa+"%")
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.SearchText != "" {
		searchWildcard := "%" + req.SearchText + "%"
		query = query.Where("servico_realizado LIKE ? OR oficina LIKE ? OR observacoes LIKE ?",
			searchWildcard, searchWildcard, searchWildcard)
	}

	// Contar total para paginação
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao contar manutenções")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao contar manutenções"})
		return
	}

	// Buscar manutenções com paginação
	var manutencoes []models.Manutencao
	if err := query.Preload("Veiculo").Offset(offset).Limit(limit).Order("data_servico DESC").Find(&manutencoes).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao listar manutenções")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar manutenções"})
		return
	}

	// Calcular valores totais
	var custoPecasTotal, custoMaoObraTotal float64
	for _, m := range manutencoes {
		custoPecasTotal += m.ValorPeca
		custoMaoObraTotal += m.ValorMaoObra
	}

	c.JSON(http.StatusOK, gin.H{
		"data": manutencoes,
		"totais": gin.H{
			"custo_pecas":    custoPecasTotal,
			"custo_mao_obra": custoMaoObraTotal,
			"custo_total":    custoPecasTotal + custoMaoObraTotal,
		},
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetEstatisticas retorna estatísticas para o painel de manutenção
func (h *ManutencaoHandler) GetEstatisticas(c *gin.Context) {
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

	// Estatísticas principais
	var totalManutencoes int64
	var custoPecas, custoMaoObra float64

	// Total de manutenções
	baseQuery := h.db.Model(&models.Manutencao{}).Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime)
	baseQuery.Count(&totalManutencoes)

	// Custos
	baseQuery.Select("COALESCE(SUM(valor_peca), 0)").Scan(&custoPecas)
	baseQuery.Select("COALESCE(SUM(valor_mao_obra), 0)").Scan(&custoMaoObra)

	// Contagem por status
	var countPorStatus struct {
		Pendentes  int64 `json:"pendentes"`
		Agendados  int64 `json:"agendados"`
		Concluidos int64 `json:"concluidos"`
		Pagos      int64 `json:"pagos"`
		Cancelados int64 `json:"cancelados"`
	}

	h.db.Model(&models.Manutencao{}).
		Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "PENDENTE").
		Count(&countPorStatus.Pendentes)

	h.db.Model(&models.Manutencao{}).
		Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "AGENDADO").
		Count(&countPorStatus.Agendados)

	h.db.Model(&models.Manutencao{}).
		Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "CONCLUIDO").
		Count(&countPorStatus.Concluidos)

	h.db.Model(&models.Manutencao{}).
		Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "PAGO").
		Count(&countPorStatus.Pagos)

	h.db.Model(&models.Manutencao{}).
		Where("data_servico BETWEEN ? AND ?", dataInicioTime, dataFimTime).
		Where("status = ?", "CANCELADO").
		Count(&countPorStatus.Cancelados)

	// Top veículos por custo
	type CustoPorVeiculo struct {
		VeiculoID    string  `json:"veiculo_id"`
		Placa        string  `json:"placa"`
		CustoPecas   float64 `json:"custo_pecas"`
		CustoMaoObra float64 `json:"custo_mao_obra"`
		CustoTotal   float64 `json:"custo_total"`
	}

	var topVeiculosPorCusto []CustoPorVeiculo
	h.db.Raw(`
        SELECT 
            m.veiculo_id,
            v.placa,
            SUM(m.valor_peca) AS custo_pecas,
            SUM(m.valor_mao_obra) AS custo_mao_obra,
            SUM(m.valor_peca) + SUM(m.valor_mao_obra) AS custo_total
        FROM manutencoes m
        JOIN veiculos v ON m.veiculo_id = v.id
        WHERE m.data_servico BETWEEN ? AND ?
        GROUP BY m.veiculo_id, v.placa
        ORDER BY custo_total DESC
        LIMIT 8
    `, dataInicioTime, dataFimTime).Scan(&topVeiculosPorCusto)

	c.JSON(http.StatusOK, gin.H{
		"total_manutencoes": totalManutencoes,
		"custo_pecas":       custoPecas,
		"custo_mao_obra":    custoMaoObra,
		"custo_total":       custoPecas + custoMaoObra,
		"por_status":        countPorStatus,
		"top_veiculos":      topVeiculosPorCusto,
		"periodo": gin.H{
			"data_inicio": dataInicioTime.Format("2006-01-02"),
			"data_fim":    dataFimTime.Format("2006-01-02"),
		},
	})
}
