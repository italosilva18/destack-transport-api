package models

// CTE representa um Conhecimento de Transporte Eletrônico
type CTE struct {
	DocumentoFiscal // Embedar DocumentoFiscal para herdar seus campos

	// Campos específicos de CT-e
	RemetenteID     string  `json:"remetente_id" gorm:"index"`
	DestinatarioID  string  `json:"destinatario_id" gorm:"index"`
	TomadorID       *string `json:"tomador_id" gorm:"index"`
	ModalidadeFrete string  `json:"modalidade_frete" gorm:"size:3;index"` // CIF, FOB
	CFOP            string  `json:"cfop" gorm:"size:4"`

	// Valores específicos
	ValorICMS  float64 `json:"valor_icms"`
	ValorCarga float64 `json:"valor_carga"`
}

// TableName define o nome da tabela no banco de dados
func (CTE) TableName() string {
	return "ctes"
}
