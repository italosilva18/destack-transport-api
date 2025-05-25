package parsers

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CTeParsed é o resultado do parsing de CT-e
type CTeParsed struct {
	Chave           string
	Numero          int
	Serie           string
	DataEmissao     time.Time
	CFOP            string
	ModalidadeFrete string
	ValorTotal      float64
	ValorCarga      float64
	UFInicio        string
	UFDestino       string
	MunicipioInicio string
	MunicipioFim    string
	Status          string
	Protocolo       string

	// Entidades
	Emitente     EmpresaParsed
	Remetente    EmpresaParsed
	Destinatario EmpresaParsed
	Tomador      *EmpresaParsed // Pode ser null

	// Informações adicionais
	RNTRC             string
	PlacaVeiculo      string
	ObservacoesGerais string

	// Documentos vinculados
	ChavesNFe []string
}

// EmpresaParsed representa dados simplificados de uma empresa
type EmpresaParsed struct {
	CNPJ        string
	CPF         string
	RazaoSocial string
	IE          string
	UF          string
	Municipio   string
	CEP         string
}

// ParseCTe faz o parsing de um XML de CT-e
func ParseCTe(xmlContent []byte) (*CTeParsed, error) {
	var cteProc CTeProc

	// Tentar fazer unmarshal do XML
	err := xml.Unmarshal(xmlContent, &cteProc)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do XML: %w", err)
	}

	// Validar estrutura básica
	if cteProc.CTe.InfCte.Id == "" {
		return nil, errors.New("XML inválido: ID do CT-e não encontrado")
	}

	// Extrair a chave do ID (remover prefixo "CTe")
	chave := strings.TrimPrefix(cteProc.CTe.InfCte.Id, "CTe")
	if len(chave) != 44 {
		return nil, fmt.Errorf("chave de acesso inválida: %s", chave)
	}

	// Parsear data de emissão
	dataEmissao, err := ParseDate(cteProc.CTe.InfCte.Ide.DhEmi)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear data de emissão: %w", err)
	}

	// Parsear número do CT-e
	numero, err := strconv.Atoi(cteProc.CTe.InfCte.Ide.NCT)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear número do CT-e: %w", err)
	}

	// Parsear valores
	valorTotal, err := parseFloat(cteProc.CTe.InfCte.VPrest.VTPrest)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear valor total: %w", err)
	}

	valorCarga := 0.0
	if cteProc.CTe.InfCte.InfCTeNorm.InfCarga.VCarga != "" {
		valorCarga, err = parseFloat(cteProc.CTe.InfCte.InfCTeNorm.InfCarga.VCarga)
		if err != nil {
			return nil, fmt.Errorf("erro ao parsear valor da carga: %w", err)
		}
	}

	// Determinar modalidade de frete baseado no tomador
	modalidadeFrete := determinarModalidadeFrete(cteProc.CTe.InfCte.Ide.Toma3.Toma)

	// Criar resultado parseado
	result := &CTeParsed{
		Chave:           chave,
		Numero:          numero,
		Serie:           cteProc.CTe.InfCte.Ide.Serie,
		DataEmissao:     dataEmissao,
		CFOP:            cteProc.CTe.InfCte.Ide.CFOP,
		ModalidadeFrete: modalidadeFrete,
		ValorTotal:      valorTotal,
		ValorCarga:      valorCarga,
		UFInicio:        cteProc.CTe.InfCte.Ide.UFIni,
		UFDestino:       cteProc.CTe.InfCte.Ide.UFFim,
		MunicipioInicio: cteProc.CTe.InfCte.Ide.XMunIni,
		MunicipioFim:    cteProc.CTe.InfCte.Ide.XMunFim,
	}

	// Parsear emitente
	result.Emitente = EmpresaParsed{
		CNPJ:        cteProc.CTe.InfCte.Emit.CNPJ,
		RazaoSocial: cteProc.CTe.InfCte.Emit.XNome,
		IE:          cteProc.CTe.InfCte.Emit.IE,
		UF:          cteProc.CTe.InfCte.Emit.EnderEmit.UF,
		Municipio:   cteProc.CTe.InfCte.Emit.EnderEmit.XMun,
		CEP:         cteProc.CTe.InfCte.Emit.EnderEmit.CEP,
	}

	// Parsear remetente
	result.Remetente = EmpresaParsed{
		CNPJ:        cteProc.CTe.InfCte.Rem.CNPJ,
		CPF:         cteProc.CTe.InfCte.Rem.CPF,
		RazaoSocial: cteProc.CTe.InfCte.Rem.XNome,
		IE:          cteProc.CTe.InfCte.Rem.IE,
		UF:          cteProc.CTe.InfCte.Rem.EnderReme.UF,
		Municipio:   cteProc.CTe.InfCte.Rem.EnderReme.XMun,
		CEP:         cteProc.CTe.InfCte.Rem.EnderReme.CEP,
	}

	// Parsear destinatário
	result.Destinatario = EmpresaParsed{
		CNPJ:        cteProc.CTe.InfCte.Dest.CNPJ,
		CPF:         cteProc.CTe.InfCte.Dest.CPF,
		RazaoSocial: cteProc.CTe.InfCte.Dest.XNome,
		IE:          cteProc.CTe.InfCte.Dest.IE,
		UF:          cteProc.CTe.InfCte.Dest.EnderDest.UF,
		Municipio:   cteProc.CTe.InfCte.Dest.EnderDest.XMun,
		CEP:         cteProc.CTe.InfCte.Dest.EnderDest.CEP,
	}

	// Determinar tomador baseado no indicador
	switch cteProc.CTe.InfCte.Ide.Toma3.Toma {
	case "0": // Remetente
		tomador := result.Remetente
		result.Tomador = &tomador
	case "1": // Expedidor
		// Não temos expedidor no XML, usar remetente
		tomador := result.Remetente
		result.Tomador = &tomador
	case "2": // Recebedor
		// Não temos recebedor no XML, usar destinatário
		tomador := result.Destinatario
		result.Tomador = &tomador
	case "3": // Destinatário
		tomador := result.Destinatario
		result.Tomador = &tomador
	}

	// Informações do protocolo
	if cteProc.ProtCTe.InfProt.CStat != "" {
		result.Status = cteProc.ProtCTe.InfProt.CStat
		result.Protocolo = cteProc.ProtCTe.InfProt.NProt
	}

	// RNTRC
	if cteProc.CTe.InfCte.InfCTeNorm.InfModal.Rodo != nil {
		result.RNTRC = cteProc.CTe.InfCte.InfCTeNorm.InfModal.Rodo.RNTRC
	}

	// Placa do veículo (buscar nas observações)
	for _, obs := range cteProc.CTe.InfCte.Compl.ObsCont {
		if obs.XCampo == "PLACA" {
			result.PlacaVeiculo = obs.XTexto
			break
		}
	}

	// Observações gerais
	result.ObservacoesGerais = cteProc.CTe.InfCte.Compl.XObs

	// Chaves de NF-e vinculadas
	for _, nfe := range cteProc.CTe.InfCte.InfCTeNorm.InfDoc.InfNFe {
		if nfe.Chave != "" {
			result.ChavesNFe = append(result.ChavesNFe, nfe.Chave)
		}
	}

	return result, nil
}

// determinarModalidadeFrete determina CIF ou FOB baseado no tomador
func determinarModalidadeFrete(toma string) string {
	switch toma {
	case "0", "1": // Remetente ou Expedidor
		return "CIF"
	case "2", "3": // Recebedor ou Destinatário
		return "FOB"
	default:
		return "CIF" // Padrão
	}
}

// parseFloat converte string para float64
func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	// Remover espaços e substituir vírgula por ponto
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", -1)
	return strconv.ParseFloat(s, 64)
}

// ValidarChaveAcesso valida a chave de acesso
func ValidarChaveAcesso(chave string) error {
	if len(chave) != 44 {
		return fmt.Errorf("chave deve ter 44 dígitos, tem %d", len(chave))
	}

	// Validar se é numérica
	for _, c := range chave {
		if c < '0' || c > '9' {
			return errors.New("chave deve conter apenas números")
		}
	}

	// Validar dígito verificador
	dv := CalcularDigitoVerificador(chave[:43])
	if string(chave[43]) != strconv.Itoa(dv) {
		return errors.New("dígito verificador inválido")
	}

	return nil
}

// CalcularDigitoVerificador calcula o DV da chave
func CalcularDigitoVerificador(chave string) int {
	if len(chave) != 43 {
		return -1
	}

	soma := 0
	peso := 2

	for i := len(chave) - 1; i >= 0; i-- {
		digito, _ := strconv.Atoi(string(chave[i]))
		soma += digito * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}

	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}
