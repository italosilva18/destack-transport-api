package upload

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/internal/services"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// UploadHandler contém os handlers para upload de arquivos
type UploadHandler struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewUploadHandler cria uma nova instância de UploadHandler
func NewUploadHandler(db *gorm.DB) *UploadHandler {
	return &UploadHandler{
		db:     db,
		logger: logger.GetLogger(),
	}
}

// UploadSingleResponse representa a resposta do upload
type UploadSingleResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// UploadSingle recebe um único arquivo XML
func (h *UploadHandler) UploadSingle(c *gin.Context) {
	// Obter o arquivo do request
	file, header, err := c.Request.FormFile("arquivo_xml")
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao receber arquivo")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado ou inválido"})
		return
	}
	defer file.Close()

	// Validar o tipo do arquivo
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xml") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Apenas arquivos XML são permitidos"})
		return
	}

	// Ler o conteúdo do arquivo
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		h.logger.Error().Err(err).Msg("Erro ao ler arquivo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar arquivo"})
		return
	}

	// Criar registro de upload
	uploadID := uuid.New()
	upload := models.Upload{
		BaseModel: models.BaseModel{
			ID: uploadID,
		},
		NomeArquivo:           header.Filename,
		Status:                "PENDENTE",
		DataUpload:            time.Now(),
		ChaveDocProcessado:    nil,
		DetalhesProcessamento: "",
	}

	// Salvar registro no banco
	if err := h.db.Create(&upload).Error; err != nil {
		h.logger.Error().Err(err).Msg("Erro ao salvar registro de upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao registrar upload"})
		return
	}

	// Iniciar processamento assíncrono
	go func() {
		// Aqui usaríamos um serviço de processamento real
		// Por enquanto, apenas simulamos
		result, err := services.ProcessarXML(h.db, uploadID.String(), buf.Bytes())
		if err != nil {
			h.logger.Error().Err(err).Str("upload_id", uploadID.String()).Msg("Erro ao processar XML")
			h.db.Model(&models.Upload{}).Where("id = ?", uploadID).Updates(map[string]interface{}{
				"status":                 "ERRO",
				"detalhes_processamento": err.Error(),
			})
			return
		}

		// Atualizar status após processamento
		h.db.Model(&models.Upload{}).Where("id = ?", uploadID).Updates(map[string]interface{}{
			"status":               "CONCLUIDO",
			"chave_doc_processado": &result.Chave,
		})
	}()

	// Retornar resposta de sucesso
	c.JSON(http.StatusAccepted, UploadSingleResponse{
		ID:      uploadID.String(),
		Message: "Upload recebido. Processamento iniciado.",
	})
}

// ListUploads retorna a lista de uploads
func (h *UploadHandler) ListUploads(c *gin.Context) {
	// Implementação básica de paginação
	page := 1
	limit := 10

	// Contagem total
	var total int64
	h.db.Model(&models.Upload{}).Count(&total)

	// Buscar uploads com paginação
	var uploads []models.Upload
	result := h.db.Order("data_upload DESC").Offset((page - 1) * limit).Limit(limit).Find(&uploads)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Msg("Erro ao buscar uploads")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar uploads"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": uploads,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUpload busca um upload pelo ID
func (h *UploadHandler) GetUpload(c *gin.Context) {
	id := c.Param("id")

	var upload models.Upload
	result := h.db.First(&upload, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Upload não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload não encontrado"})
		return
	}

	c.JSON(http.StatusOK, upload)
}
