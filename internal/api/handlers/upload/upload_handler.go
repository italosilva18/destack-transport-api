package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/internal/services"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// UploadHandler contém os handlers para upload de arquivos
type UploadHandler struct {
	db     *gorm.DB
	logger zerolog.Logger
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

// UploadBatchResponse representa a resposta do upload em lote
type UploadBatchResponse struct {
	Message       string                 `json:"message"`
	TotalRecebido int                    `json:"total_recebido"`
	Uploads       []UploadSingleResponse `json:"uploads"`
}

// UploadBatch recebe múltiplos arquivos XML
func (h *UploadHandler) UploadBatch(c *gin.Context) {
	// Obter o formulário multipart
	form, err := c.MultipartForm()
	if err != nil {
		h.logger.Error().Err(err).Msg("Erro ao receber formulário multipart")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário"})
		return
	}

	// Obter arquivos do campo "arquivos_xml"
	files := form.File["arquivos_xml"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo foi enviado"})
		return
	}

	// Limite de arquivos por upload
	const maxFiles = 100
	if len(files) > maxFiles {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Limite máximo de arquivos excedido",
			"max":   maxFiles,
		})
		return
	}

	response := UploadBatchResponse{
		TotalRecebido: len(files),
		Uploads:       make([]UploadSingleResponse, 0, len(files)),
	}

	// Processar cada arquivo
	var wg sync.WaitGroup
	uploadsChan := make(chan UploadSingleResponse, len(files))
	errorsChan := make(chan error, len(files))

	for _, fileHeader := range files {
		// Validar o tipo do arquivo
		if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".xml") {
			h.logger.Warn().Str("filename", fileHeader.Filename).Msg("Arquivo ignorado - não é XML")
			continue
		}

		wg.Add(1)
		go func(fh *multipart.FileHeader) {
			defer wg.Done()

			// Abrir arquivo
			file, err := fh.Open()
			if err != nil {
				errorsChan <- err
				return
			}
			defer file.Close()

			// Ler conteúdo
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, file); err != nil {
				errorsChan <- err
				return
			}

			// Criar registro de upload
			uploadID := uuid.New()
			upload := models.Upload{
				BaseModel: models.BaseModel{
					ID: uploadID,
				},
				NomeArquivo:           fh.Filename,
				Status:                "PENDENTE",
				DataUpload:            time.Now(),
				ChaveDocProcessado:    nil,
				DetalhesProcessamento: "",
			}

			// Salvar registro no banco
			if err := h.db.Create(&upload).Error; err != nil {
				errorsChan <- err
				return
			}

			// Iniciar processamento assíncrono
			go func(id uuid.UUID, content []byte) {
				result, err := services.ProcessarXML(h.db, id.String(), content)
				if err != nil {
					h.logger.Error().Err(err).Str("upload_id", id.String()).Msg("Erro ao processar XML")
					h.db.Model(&models.Upload{}).Where("id = ?", id).Updates(map[string]interface{}{
						"status":                 "ERRO",
						"detalhes_processamento": err.Error(),
					})
					return
				}

				// Atualizar status após processamento
				h.db.Model(&models.Upload{}).Where("id = ?", id).Updates(map[string]interface{}{
					"status":               "CONCLUIDO",
					"chave_doc_processado": &result.Chave,
				})
			}(uploadID, buf.Bytes())

			uploadsChan <- UploadSingleResponse{
				ID:      uploadID.String(),
				Message: "Upload recebido. Processamento iniciado.",
			}
		}(fileHeader)
	}

	// Aguardar todos os uploads
	wg.Wait()
	close(uploadsChan)
	close(errorsChan)

	// Coletar resultados
	for upload := range uploadsChan {
		response.Uploads = append(response.Uploads, upload)
	}

	// Verificar erros
	var hasErrors bool
	for err := range errorsChan {
		if err != nil {
			hasErrors = true
			h.logger.Error().Err(err).Msg("Erro durante upload em lote")
		}
	}

	if hasErrors {
		response.Message = "Upload em lote concluído com alguns erros"
	} else {
		response.Message = "Upload em lote concluído com sucesso"
	}

	c.JSON(http.StatusAccepted, response)
}

// DeleteUpload exclui um upload
func (h *UploadHandler) DeleteUpload(c *gin.Context) {
	id := c.Param("id")

	// Buscar upload
	var upload models.Upload
	result := h.db.First(&upload, "id = ?", id)
	if result.Error != nil {
		h.logger.Error().Err(result.Error).Str("id", id).Msg("Upload não encontrado")
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload não encontrado"})
		return
	}

	// Verificar se o upload ainda está pendente
	if upload.Status == "PENDENTE" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Não é possível excluir um upload em processamento",
		})
		return
	}

	// Deletar upload (soft delete)
	if err := h.db.Delete(&upload).Error; err != nil {
		h.logger.Error().Err(err).Str("id", id).Msg("Erro ao excluir upload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload excluído com sucesso",
		"id":      id,
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
