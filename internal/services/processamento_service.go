package services

import (
	"errors"
	"strings"

	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/internal/parsers"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"gorm.io/gorm"
)

// DocumentoProcessado representa o resultado do processamento de um XML
type DocumentoProcessado struct {
	Chave    string
	Tipo     string
	Status   string
	Mensagem string
}

// ProcessarXML processa um arquivo XML
func ProcessarXML(db *gorm.DB, uploadID string, xmlContent []byte) (*DocumentoProcessado, error) {
	log := logger.GetLogger()
	log.Info().Str("upload_id", uploadID).Msg("Iniciando processamento de XML")

	// Detectar tipo de documento
	tipoDoc, err := detectarTipoDocumento(xmlContent)
	if err != nil {
		log.Error().Err(err).Msg("Erro ao detectar tipo de documento")
		return nil, err
	}

	// Processar conforme o tipo
	var resultado *DocumentoProcessado

	switch tipoDoc {
	case "CTE":
		resultado, err = processarCTe(db, xmlContent)
	case "MDFE":
		resultado, err = processarMDFe(db, xmlContent)
	case "EVENTO":
		resultado, err = processarEvento(db, xmlContent)
	default:
		err = errors.New("tipo de documento não suportado")
	}

	if err != nil {
		log.Error().Err(err).Str("tipo", tipoDoc).Msg("Erro ao processar documento")
		return nil, err
	}

	log.Info().Str("upload_id", uploadID).Str("tipo", tipoDoc).Str("chave", resultado.Chave).Msg("Processamento concluído com sucesso")

	return resultado, nil
}

// detectarTipoDocumento detecta o tipo de documento a partir do conteúdo XML
func detectarTipoDocumento(xmlContent []byte) (string, error) {
	// Simplificação: verificar strings características no XML
	content := string(xmlContent)

	if strings.Contains(content, "<CTe") || strings.Contains(content, "<cte") {
		return "CTE", nil
	} else if strings.Contains(content, "<MDFe") || strings.Contains(content, "<mdfe") {
		return "MDFE", nil
	} else if strings.Contains(content, "<evento") || strings.Contains(content, "<Evento") {
		return "EVENTO", nil
	}

	return "", errors.New("tipo de documento não identificado")
}

// processarCTe processa um CT-e
func processarCTe(db *gorm.DB, xmlContent []byte) (*DocumentoProcessado, error) {
	// Aqui iríamos usar o parser real de CT-e
	// Por enquanto, apenas simulamos o processamento

	// Simular um parser de CT-e
	cte, err := parsers.ParseCTe(xmlContent)
	if err != nil {
		return nil, err
	}

	// Buscar ou criar emitente
	var emitente models.Empresa
	result := db.FirstOrCreate(&emitente, models.Empresa{
		CNPJ:        &cte.Emitente.CNPJ,
		RazaoSocial: cte.Emitente.RazaoSocial,
	})
	if result.Error != nil {
		return nil, result.Error
	}

	// Buscar ou criar remetente e destinatário
	// (Código omitido para simplicidade)

	// Criar ou atualizar o CT-e no banco
	novoCtE := models.CTE{
		DocumentoFiscal: models.DocumentoFiscal{
			Chave:       cte.Chave,
			Tipo:        "CTE",
			Numero:      cte.Numero,
			Serie:       cte.Serie,
			DataEmissao: cte.DataEmissao,
			Status:      "PROCESSADO",
			ValorTotal:  cte.ValorTotal,
			EmitenteID:  emitente.ID.String(),
			// Outros campos conforme necessário
		},
		CFOP:            cte.CFOP,
		ModalidadeFrete: cte.ModalidadeFrete,
		// Outros campos específicos
	}

	// Verificar se já existe
	var existingCTe models.CTE
	result = db.Where("chave = ?", cte.Chave).First(&existingCTe)

	// Se não existir, criar novo
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			result = db.Create(&novoCtE)
			if result.Error != nil {
				return nil, result.Error
			}
		} else {
			return nil, result.Error
		}
	} else {
		// Se já existir, atualizar
		result = db.Model(&existingCTe).Updates(novoCtE)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &DocumentoProcessado{
		Chave:  cte.Chave,
		Tipo:   "CTE",
		Status: "PROCESSADO",
	}, nil
}

// processarMDFe processa um MDF-e
func processarMDFe(db *gorm.DB, xmlContent []byte) (*DocumentoProcessado, error) {
	// Implementação similar ao CT-e
	// Por brevidade, retornamos um resultado simulado
	return &DocumentoProcessado{
		Chave:  "MDFE12345",
		Tipo:   "MDFE",
		Status: "PROCESSADO",
	}, nil
}

// processarEvento processa um evento
func processarEvento(db *gorm.DB, xmlContent []byte) (*DocumentoProcessado, error) {
	// Implementação para eventos
	// Por brevidade, retornamos um resultado simulado
	return &DocumentoProcessado{
		Chave:  "EVENTO12345",
		Tipo:   "EVENTO",
		Status: "PROCESSADO",
	}, nil
}
