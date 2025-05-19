package parsers

import (
	"encoding/xml"
	"errors"
	"strings"
	"time"
)

// Estruturas para mapeamento do XML de CT-e

// Empresa representa dados simplificados de uma empresa
type Empresa struct {
	CNPJ        string
	CPF         string
	RazaoSocial string
	UF          string
}

// CTeParsed é o resultado do parsing de CT-e
type CTeParsed struct {
	Chave           string
	Numero          int
	Serie           string
	DataEmissao     time.Time
	Emitente        Empresa
	Remetente       Empresa
	Destinatario    Empresa
	CFOP            string
	ModalidadeFrete string
	ValorTotal      float64
	ValorCarga      float64
	UFInicio        string
	UFDestino       string
}

// ParseCTe faz o parsing de um XML de CT-e
func ParseCTe(xmlContent []byte) (*CTeParsed, error) {
	// Aqui seria implementado o parser real de CT-e
	// Por simplicidade, retornamos uma estrutura simulada

	// Verificar se o XML é válido
	if len(xmlContent) < 50 {
		return nil, errors.New("XML inválido ou muito pequeno")
	}

	// Estrutura temporária para simular o parser
	// Em um parser real, usaríamos estruturas que correspondem exatamente ao XML
	type CTeParsedSimple struct {
		XMLName     xml.Name `xml:"CTe"`
		Id          string   `xml:"infCte>Id,attr"`
		ChaveAcesso string   `xml:",chardata"`
	}

	var cteSimple CTeParsedSimple
	err := xml.Unmarshal(xmlContent, &cteSimple)
	if err != nil {
		return nil, err
	}

	// Extrair a chave do documento (removendo prefixo CTe)
	chave := strings.TrimPrefix(cteSimple.Id, "CTe")
	if chave == "" {
		chave = "CT123456789" // Valor simulado
	}

	// Retornar uma estrutura simulada
	return &CTeParsed{
		Chave:       chave,
		Numero:      123456, // Valor simulado
		Serie:       "1",
		DataEmissao: time.Now(),
		Emitente: Empresa{
			CNPJ:        "12345678901234",
			RazaoSocial: "Empresa Emitente Teste",
			UF:          "SP",
		},
		Remetente: Empresa{
			CNPJ:        "98765432109876",
			RazaoSocial: "Empresa Remetente Teste",
			UF:          "RJ",
		},
		Destinatario: Empresa{
			CNPJ:        "11122233344455",
			RazaoSocial: "Empresa Destinatário Teste",
			UF:          "MG",
		},
		CFOP:            "5353",
		ModalidadeFrete: "CIF",
		ValorTotal:      1500.50,
		ValorCarga:      10000.00,
		UFInicio:        "SP",
		UFDestino:       "MG",
	}, nil
}
