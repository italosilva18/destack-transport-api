package parsers

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MDFeParsed é o resultado do parsing de MDF-e
type MDFeParsed struct {
	Chave            string
	Numero           int
	Serie            string
	DataEmissao      time.Time
	UFInicio         string
	UFDestino        string
	MunicipioCarrega string
	Status           string
	Protocolo        string
	Encerrado        bool
	DataEncerramento *time.Time

	// Emitente
	Emitente EmpresaParsed

	// Veículo
	PlacaVeiculo string
	UfVeiculo    string
	RNTRC        string
	TaraVeiculo  int
	CapacidadeKg int

	// Condutor
	NomeMotorista string
	CPFMotorista  string

	// Documentos transportados
	ChavesCTe []string
	ChavesNFe []string

	// Totalizadores
	QtdCTe          int
	QtdNFe          int
	ValorTotalCarga float64
	PesoBrutoTotal  float64

	// Produto predominante
	ProdutoPredominante string
	TipoCarga           string

	// Seguro
	Seguradoras []SeguradoraParsed
}

// SeguradoraParsed informações da seguradora
type SeguradoraParsed struct {
	Nome      string
	CNPJ      string
	Apolice   string
	Averbacao string
}

// ParseMDFe faz o parsing de um XML de MDF-e
func ParseMDFe(xmlContent []byte) (*MDFeParsed, error) {
	var mdfeProc MDFeProc

	// Tentar fazer unmarshal do XML
	err := xml.Unmarshal(xmlContent, &mdfeProc)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do XML: %w", err)
	}

	// Validar estrutura básica
	if mdfeProc.MDFe.InfMDFe.Id == "" {
		return nil, errors.New("XML inválido: ID do MDF-e não encontrado")
	}

	// Extrair a chave do ID (remover prefixo "MDFe")
	chave := strings.TrimPrefix(mdfeProc.MDFe.InfMDFe.Id, "MDFe")
	if len(chave) != 44 {
		return nil, fmt.Errorf("chave de acesso inválida: %s", chave)
	}

	// Parsear data de emissão
	dataEmissao, err := ParseDate(mdfeProc.MDFe.InfMDFe.Ide.DhEmi)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear data de emissão: %w", err)
	}

	// Parsear número do MDF-e
	numero, err := strconv.Atoi(mdfeProc.MDFe.InfMDFe.Ide.NMDF)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear número do MDF-e: %w", err)
	}

	// Criar resultado parseado
	result := &MDFeParsed{
		Chave:            chave,
		Numero:           numero,
		Serie:            mdfeProc.MDFe.InfMDFe.Ide.Serie,
		DataEmissao:      dataEmissao,
		UFInicio:         mdfeProc.MDFe.InfMDFe.Ide.UFIni,
		UFDestino:        mdfeProc.MDFe.InfMDFe.Ide.UFFim,
		MunicipioCarrega: mdfeProc.MDFe.InfMDFe.Ide.InfMunCarrega.XMunCarrega,
		Encerrado:        false,
	}

	// Parsear emitente
	result.Emitente = EmpresaParsed{
		CNPJ:        mdfeProc.MDFe.InfMDFe.Emit.CNPJ,
		RazaoSocial: mdfeProc.MDFe.InfMDFe.Emit.XNome,
		IE:          mdfeProc.MDFe.InfMDFe.Emit.IE,
		UF:          mdfeProc.MDFe.InfMDFe.Emit.EnderEmit.UF,
		Municipio:   mdfeProc.MDFe.InfMDFe.Emit.EnderEmit.XMun,
		CEP:         mdfeProc.MDFe.InfMDFe.Emit.EnderEmit.CEP,
	}

	// Informações do protocolo
	if mdfeProc.ProtMDFe.InfProt.CStat != "" {
		result.Status = mdfeProc.ProtMDFe.InfProt.CStat
		result.Protocolo = mdfeProc.ProtMDFe.InfProt.NProt
	}

	// Informações do veículo
	if mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.Placa != "" {
		result.PlacaVeiculo = mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.Placa
		result.UfVeiculo = mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.UF

		// Parsear tara e capacidade
		if tara, err := strconv.Atoi(mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.Tara); err == nil {
			result.TaraVeiculo = tara
		}
		if capKg, err := strconv.Atoi(mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.CapKG); err == nil {
			result.CapacidadeKg = capKg
		}
	}

	// RNTRC
	result.RNTRC = mdfeProc.MDFe.InfMDFe.InfModal.Rodo.InfANTT.RNTRC

	// Condutor
	if len(mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.Condutor) > 0 {
		condutor := mdfeProc.MDFe.InfMDFe.InfModal.Rodo.VeicTracao.Condutor[0]
		result.NomeMotorista = condutor.XNome
		result.CPFMotorista = condutor.CPF
	}

	// Documentos transportados
	for _, munDescarga := range mdfeProc.MDFe.InfMDFe.InfDoc.InfMunDescarga {
		// CT-es
		for _, cte := range munDescarga.InfCTe {
			if cte.ChCTe != "" {
				result.ChavesCTe = append(result.ChavesCTe, cte.ChCTe)
			}
		}
		// NF-es
		for _, nfe := range munDescarga.InfNFe {
			if nfe.Chave != "" {
				result.ChavesNFe = append(result.ChavesNFe, nfe.Chave)
			}
		}
	}

	// Totalizadores
	if qtdCTe, err := strconv.Atoi(mdfeProc.MDFe.InfMDFe.Tot.QCTe); err == nil {
		result.QtdCTe = qtdCTe
	}
	if qtdNFe, err := strconv.Atoi(mdfeProc.MDFe.InfMDFe.Tot.QNFe); err == nil {
		result.QtdNFe = qtdNFe
	}
	if valorCarga, err := parseFloat(mdfeProc.MDFe.InfMDFe.Tot.VCarga); err == nil {
		result.ValorTotalCarga = valorCarga
	}
	if pesoBruto, err := parseFloat(mdfeProc.MDFe.InfMDFe.Tot.QCarga); err == nil {
		result.PesoBrutoTotal = pesoBruto
	}

	// Produto predominante
	result.ProdutoPredominante = mdfeProc.MDFe.InfMDFe.ProdPred.XProd
	result.TipoCarga = mdfeProc.MDFe.InfMDFe.ProdPred.TpCarga

	// Seguradoras
	for _, seg := range mdfeProc.MDFe.InfMDFe.Seg {
		seguradora := SeguradoraParsed{
			Nome:      seg.InfSeg.XSeg,
			CNPJ:      seg.InfResp.CNPJ,
			Apolice:   seg.NApol,
			Averbacao: seg.NAver,
		}
		result.Seguradoras = append(result.Seguradoras, seguradora)
	}

	return result, nil
}

// ParseEventoMDFe faz o parsing de um evento de MDF-e
func ParseEventoMDFe(xmlContent []byte) (*EventoParsed, error) {
	var procEvento ProcEventoMDFe

	// Tentar fazer unmarshal do XML
	err := xml.Unmarshal(xmlContent, &procEvento)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do evento: %w", err)
	}

	// Extrair chave do MDF-e
	chave := procEvento.EventoMDFe.InfEvento.ChMDFe
	if len(chave) != 44 {
		return nil, fmt.Errorf("chave de acesso inválida: %s", chave)
	}

	// Parsear data do evento
	dataEvento, err := ParseDate(procEvento.EventoMDFe.InfEvento.DhEvento)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear data do evento: %w", err)
	}

	result := &EventoParsed{
		Chave:      chave,
		TipoEvento: procEvento.EventoMDFe.InfEvento.TpEvento,
		Sequencia:  procEvento.EventoMDFe.InfEvento.NSeqEvento,
		DataEvento: dataEvento,
		Protocolo:  procEvento.RetEventoMDFe.InfEvento.NProt,
		Status:     procEvento.RetEventoMDFe.InfEvento.CStat,
		Motivo:     procEvento.RetEventoMDFe.InfEvento.XMotivo,
	}

	// Detalhes específicos do evento
	if procEvento.EventoMDFe.InfEvento.DetEvento.EvCancMDFe != nil {
		result.TipoEventoDesc = "Cancelamento"
		result.Justificativa = procEvento.EventoMDFe.InfEvento.DetEvento.EvCancMDFe.XJust
		result.ProtocoloRef = procEvento.EventoMDFe.InfEvento.DetEvento.EvCancMDFe.NProt
	}

	return result, nil
}

// EventoParsed resultado do parsing de evento
type EventoParsed struct {
	Chave          string
	TipoEvento     string
	TipoEventoDesc string
	Sequencia      string
	DataEvento     time.Time
	Protocolo      string
	ProtocoloRef   string // Protocolo referenciado (cancelamento)
	Status         string
	Motivo         string
	Justificativa  string
}
