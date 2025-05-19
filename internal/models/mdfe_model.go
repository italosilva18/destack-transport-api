package models

import (
	"time"
)

// MDFE representa um Manifesto Eletrônico de Documentos Fiscais
type MDFE struct {
	DocumentoFiscal // Embedar DocumentoFiscal para herdar seus campos

	// Campos específicos de MDF-e
	VeiculoTracaoID string `json:"veiculo_tracao_id" gorm:"index"`

	// Totalizadores
	QtdCTe         int     `json:"qtd_cte"`
	QtdNFe         int     `json:"qtd_nfe"`
	PesoBrutoTotal float64 `json:"peso_bruto_total"`

	// Status específicos
	Encerrado        bool       `json:"encerrado" gorm:"index;default:false"`
	DataEncerramento *time.Time `json:"data_encerramento"`
}

// TableName define o nome da tabela no banco de dados
func (MDFE) TableName() string {
	return "mdfes"
}
