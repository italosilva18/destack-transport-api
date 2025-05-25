package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
		resultado, err = processarCTe(db, xmlContent, uploadID)
	case "MDFE":
		resultado, err = processarMDFe(db, xmlContent, uploadID)
	case "EVENTO_CTE":
		resultado, err = processarEventoCTe(db, xmlContent)
	case "EVENTO_MDFE":
		resultado, err = processarEventoMDFe(db, xmlContent)
	default:
		err = errors.New("tipo de documento não suportado")
	}

	if err != nil {
		log.Error().Err(err).Str("tipo", tipoDoc).Msg("Erro ao processar documento")
		return nil, err
	}

	// Atualizar upload com a chave processada
	if uploadID != "" && resultado != nil {
		uploadUUID, _ := uuid.Parse(uploadID)
		db.Model(&models.Upload{}).Where("id = ?", uploadUUID).Updates(map[string]interface{}{
			"status":               "CONCLUIDO",
			"chave_doc_processado": &resultado.Chave,
		})
	}

	log.Info().Str("upload_id", uploadID).Str("tipo", tipoDoc).Str("chave", resultado.Chave).Msg("Processamento concluído com sucesso")

	return resultado, nil
}

// detectarTipoDocumento detecta o tipo de documento a partir do conteúdo XML
func detectarTipoDocumento(xmlContent []byte) (string, error) {
	content := string(xmlContent)

	// Checar tipos de documento
	if strings.Contains(content, "<cteProc") || strings.Contains(content, "<CTe") {
		return "CTE", nil
	} else if strings.Contains(content, "<mdfeProc") || strings.Contains(content, "<MDFe") {
		return "MDFE", nil
	} else if strings.Contains(content, "<procEventoCTe") || strings.Contains(content, "<eventoCTe") {
		return "EVENTO_CTE", nil
	} else if strings.Contains(content, "<procEventoMDFe") || strings.Contains(content, "<eventoMDFe") {
		return "EVENTO_MDFE", nil
	}

	return "", errors.New("tipo de documento não identificado")
}

// processarCTe processa um CT-e
func processarCTe(db *gorm.DB, xmlContent []byte, uploadID string) (*DocumentoProcessado, error) {
	// Parser do CT-e
	cteParsed, err := parsers.ParseCTe(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do CT-e: %w", err)
	}

	// Iniciar transação
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Buscar ou criar emitente
	emitente, err := buscarOuCriarEmpresa(tx, cteParsed.Emitente)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erro ao processar emitente: %w", err)
	}

	// Buscar ou criar remetente
	remetente, err := buscarOuCriarEmpresa(tx, cteParsed.Remetente)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erro ao processar remetente: %w", err)
	}

	// Buscar ou criar destinatário
	destinatario, err := buscarOuCriarEmpresa(tx, cteParsed.Destinatario)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erro ao processar destinatário: %w", err)
	}

	// Buscar ou criar tomador (se diferente)
	var tomadorID *uuid.UUID
	if cteParsed.Tomador != nil {
		tomador, err := buscarOuCriarEmpresa(tx, *cteParsed.Tomador)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("erro ao processar tomador: %w", err)
		}
		tomadorID = &tomador.ID
	}

	// Converter upload ID para UUID
	var uploadUUID *uuid.UUID
	if uploadID != "" {
		parsed, err := uuid.Parse(uploadID)
		if err == nil {
			uploadUUID = &parsed
		}
	}

	// Criar ou atualizar o CT-e
	novoCte := models.CTE{
		DocumentoFiscal: models.DocumentoFiscal{
			Chave:           cteParsed.Chave,
			Tipo:            "CTE",
			Numero:          cteParsed.Numero,
			Serie:           cteParsed.Serie,
			DataEmissao:     cteParsed.DataEmissao,
			Status:          cteParsed.Status,
			Protocolo:       cteParsed.Protocolo,
			ValorTotal:      cteParsed.ValorTotal,
			EmitenteID:      emitente.ID,
			UFInicio:        cteParsed.UFInicio,
			UFDestino:       cteParsed.UFDestino,
			MunicipioInicio: cteParsed.MunicipioInicio,
			MunicipioFim:    cteParsed.MunicipioFim,
			XMLOriginal:     string(xmlContent),
			UploadID:        uploadUUID,
		},
		RemetenteID:     remetente.ID,
		DestinatarioID:  destinatario.ID,
		TomadorID:       tomadorID,
		CFOP:            cteParsed.CFOP,
		ModalidadeFrete: cteParsed.ModalidadeFrete,
		ValorCarga:      cteParsed.ValorCarga,
		RNTRC:           cteParsed.RNTRC,
		PlacaVeiculo:    cteParsed.PlacaVeiculo,
		ObsGerais:       cteParsed.ObservacoesGerais,
	}

	// Verificar se já existe
	var existingCte models.CTE
	result := tx.Where("chave = ?", cteParsed.Chave).First(&existingCte)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Criar novo
			if err := tx.Create(&novoCte).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("erro ao criar CT-e: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("erro ao buscar CT-e existente: %w", result.Error)
		}
	} else {
		// Atualizar existente
		if err := tx.Model(&existingCte).Updates(novoCte).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("erro ao atualizar CT-e: %w", err)
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return &DocumentoProcessado{
		Chave:    cteParsed.Chave,
		Tipo:     "CTE",
		Status:   "PROCESSADO",
		Mensagem: fmt.Sprintf("CT-e %d processado com sucesso", cteParsed.Numero),
	}, nil
}

// processarMDFe processa um MDF-e
func processarMDFe(db *gorm.DB, xmlContent []byte, uploadID string) (*DocumentoProcessado, error) {
	// Parser do MDF-e
	mdfeParsed, err := parsers.ParseMDFe(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do MDF-e: %w", err)
	}

	// Iniciar transação
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Buscar ou criar emitente
	emitente, err := buscarOuCriarEmpresa(tx, mdfeParsed.Emitente)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erro ao processar emitente: %w", err)
	}

	// Buscar ou criar veículo
	veiculo, err := buscarOuCriarVeiculo(tx, mdfeParsed.PlacaVeiculo, mdfeParsed.UfVeiculo)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("erro ao processar veículo: %w", err)
	}

	// Converter upload ID para UUID
	var uploadUUID *uuid.UUID
	if uploadID != "" {
		parsed, err := uuid.Parse(uploadID)
		if err == nil {
			uploadUUID = &parsed
		}
	}

	// Criar ou atualizar o MDF-e
	novoMdfe := models.MDFE{
		DocumentoFiscal: models.DocumentoFiscal{
			Chave:           mdfeParsed.Chave,
			Tipo:            "MDFE",
			Numero:          mdfeParsed.Numero,
			Serie:           mdfeParsed.Serie,
			DataEmissao:     mdfeParsed.DataEmissao,
			Status:          mdfeParsed.Status,
			Protocolo:       mdfeParsed.Protocolo,
			ValorTotal:      mdfeParsed.ValorTotalCarga,
			EmitenteID:      emitente.ID,
			UFInicio:        mdfeParsed.UFInicio,
			UFDestino:       mdfeParsed.UFDestino,
			MunicipioInicio: mdfeParsed.MunicipioCarrega,
			XMLOriginal:     string(xmlContent),
			UploadID:        uploadUUID,
		},
		VeiculoTracaoID:     veiculo.ID,
		NomeMotorista:       mdfeParsed.NomeMotorista,
		CPFMotorista:        mdfeParsed.CPFMotorista,
		QtdCTe:              mdfeParsed.QtdCTe,
		QtdNFe:              mdfeParsed.QtdNFe,
		PesoBrutoTotal:      mdfeParsed.PesoBrutoTotal,
		ProdutoPredominante: mdfeParsed.ProdutoPredominante,
		TipoCarga:           mdfeParsed.TipoCarga,
		Encerrado:           mdfeParsed.Encerrado,
		DataEncerramento:    mdfeParsed.DataEncerramento,
	}

	// Adicionar informações de seguro se existirem
	if len(mdfeParsed.Seguradoras) > 0 {
		seg := mdfeParsed.Seguradoras[0] // Pegar primeira seguradora
		novoMdfe.SeguradoraNome = seg.Nome
		novoMdfe.SeguradoraCNPJ = seg.CNPJ
		novoMdfe.NumeroApolice = seg.Apolice
		novoMdfe.NumeroAverbacao = seg.Averbacao
	}

	// Verificar se já existe
	var existingMdfe models.MDFE
	result := tx.Where("chave = ?", mdfeParsed.Chave).First(&existingMdfe)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Criar novo
			if err := tx.Create(&novoMdfe).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("erro ao criar MDF-e: %w", err)
			}
			existingMdfe = novoMdfe
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("erro ao buscar MDF-e existente: %w", result.Error)
		}
	} else {
		// Atualizar existente
		if err := tx.Model(&existingMdfe).Updates(novoMdfe).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("erro ao atualizar MDF-e: %w", err)
		}
	}

	// Vincular CT-es ao MDF-e
	if len(mdfeParsed.ChavesCTe) > 0 {
		for _, chaveCte := range mdfeParsed.ChavesCTe {
			var cte models.CTE
			if err := tx.Where("chave = ?", chaveCte).First(&cte).Error; err == nil {
				// Adicionar relação many-to-many
				tx.Model(&existingMdfe).Association("CTes").Append(&cte)
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return &DocumentoProcessado{
		Chave:    mdfeParsed.Chave,
		Tipo:     "MDFE",
		Status:   "PROCESSADO",
		Mensagem: fmt.Sprintf("MDF-e %d processado com sucesso", mdfeParsed.Numero),
	}, nil
}

// processarEventoCTe processa um evento de CT-e
func processarEventoCTe(db *gorm.DB, xmlContent []byte) (*DocumentoProcessado, error) {
	// Parser do evento
	evento, err := parsers.ParseEventoCTe(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do evento: %w", err)
	}

	// Buscar CT-e relacionado
	var cte models.CTE
	if err := db.Where("chave = ?", evento.Chave).First(&cte).Error; err != nil {
		return nil, fmt.Errorf("CT-e não encontrado para o evento: %w", err)
	}

	// Processar tipo de evento
	switch evento.TipoEvento {
	case "110111": // Cancelamento
		cte.Cancelado = true
		cte.Status = "101"
		if err := db.Save(&cte).Error; err != nil {
			return nil, fmt.Errorf("erro ao cancelar CT-e: %w", err)
		}
	default:
		return nil, fmt.Errorf("tipo de evento não suportado: %s", evento.TipoEvento)
	}

	return &DocumentoProcessado{
		Chave:    evento.Chave,
		Tipo:     "EVENTO_CTE",
		Status:   "PROCESSADO",
		Mensagem: fmt.Sprintf("Evento %s processado para CT-e", evento.TipoEventoDesc),
	}, nil
}

// processarEventoMDFe processa um evento de MDF-e
func processarEventoMDFe(db *gorm.DB, xmlContent []byte) (*DocumentoProcessado, error) {
	// Parser do evento
	evento, err := parsers.ParseEventoMDFe(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do evento: %w", err)
	}

	// Buscar MDF-e relacionado
	var mdfe models.MDFE
	if err := db.Where("chave = ?", evento.Chave).First(&mdfe).Error; err != nil {
		return nil, fmt.Errorf("MDF-e não encontrado para o evento: %w", err)
	}

	// Processar tipo de evento
	switch evento.TipoEvento {
	case "110111": // Cancelamento
		mdfe.Cancelado = true
		mdfe.Status = "101"
		if err := db.Save(&mdfe).Error; err != nil {
			return nil, fmt.Errorf("erro ao cancelar MDF-e: %w", err)
		}
	case "110112": // Encerramento
		if err := mdfe.Encerrar(""); err != nil {
			return nil, fmt.Errorf("erro ao encerrar MDF-e: %w", err)
		}
		if err := db.Save(&mdfe).Error; err != nil {
			return nil, fmt.Errorf("erro ao salvar encerramento do MDF-e: %w", err)
		}
	default:
		return nil, fmt.Errorf("tipo de evento não suportado: %s", evento.TipoEvento)
	}

	return &DocumentoProcessado{
		Chave:    evento.Chave,
		Tipo:     "EVENTO_MDFE",
		Status:   "PROCESSADO",
		Mensagem: fmt.Sprintf("Evento %s processado para MDF-e", evento.TipoEventoDesc),
	}, nil
}

// buscarOuCriarEmpresa busca ou cria uma empresa
func buscarOuCriarEmpresa(tx *gorm.DB, empresaParsed parsers.EmpresaParsed) (*models.Empresa, error) {
	var empresa models.Empresa

	// Buscar por CNPJ ou CPF
	query := tx.Model(&models.Empresa{})
	if empresaParsed.CNPJ != "" {
		query = query.Where("cnpj = ?", empresaParsed.CNPJ)
	} else if empresaParsed.CPF != "" {
		query = query.Where("cpf = ?", empresaParsed.CPF)
	} else {
		return nil, errors.New("empresa sem CNPJ ou CPF")
	}

	result := query.First(&empresa)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Criar nova empresa
			empresa = models.Empresa{
				CNPJ:        nilIfEmpty(empresaParsed.CNPJ),
				CPF:         nilIfEmpty(empresaParsed.CPF),
				RazaoSocial: empresaParsed.RazaoSocial,
				UF:          empresaParsed.UF,
			}

			if empresaParsed.IE != "" && empresaParsed.IE != "ISENTO" {
				empresa.IE = &empresaParsed.IE
			}

			if err := tx.Create(&empresa).Error; err != nil {
				return nil, fmt.Errorf("erro ao criar empresa: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro ao buscar empresa: %w", result.Error)
		}
	}

	return &empresa, nil
}

// buscarOuCriarVeiculo busca ou cria um veículo
func buscarOuCriarVeiculo(tx *gorm.DB, placa, uf string) (*models.Veiculo, error) {
	if placa == "" {
		return nil, errors.New("placa do veículo não informada")
	}

	var veiculo models.Veiculo
	result := tx.Where("placa = ?", placa).First(&veiculo)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Criar novo veículo
			veiculo = models.Veiculo{
				Placa: placa,
				Tipo:  "PROPRIO", // Padrão
			}

			if err := tx.Create(&veiculo).Error; err != nil {
				return nil, fmt.Errorf("erro ao criar veículo: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro ao buscar veículo: %w", result.Error)
		}
	}

	return &veiculo, nil
}

// nilIfEmpty retorna um ponteiro para string ou nil se vazio
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
