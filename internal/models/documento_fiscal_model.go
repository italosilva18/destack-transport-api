package models

import (
	"time"
)

// DocumentoFiscal representa a base para CT-e e MDF-e
type DocumentoFiscal struct {
	BaseModel
	Chave       string    `json:"chave" gorm:"uniqueIndex;not null"`
	Tipo        string    `json:"tipo" gorm:"index;not null"` // CTE, MDFE
	Numero      int       `json:"numero" gorm:"not null"`
	Serie       string    `json:"serie" gorm:"not null"`
	DataEmissao time.Time `json:"data_emissao" gorm:"index;not null"`
	Status      string    `json:"status" gorm:"index;not null"`
	Cancelado   bool      `json:"cancelado" gorm:"index;default:false"`
	ValorTotal  float64   `json:"valor_total"`

	// Entidades principais (simplificadas)
	EmitenteID string `json:"emitente_id" gorm:"index"`

	// Dados de localização
	UFInicio        string `json:"uf_inicio" gorm:"size:2;index"`
	UFDestino       string `json:"uf_destino" gorm:"size:2;index"`
	MunicipioInicio string `json:"municipio_inicio"`
	MunicipioFim    string `json:"municipio_fim"`

	// Metadados de processamento
	DataProcessamento time.Time `json:"data_processamento"`
}
