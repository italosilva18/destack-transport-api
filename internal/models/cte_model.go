package models

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CTE representa um Conhecimento de Transporte Eletrônico
type CTE struct {
	DocumentoFiscal // Embedar DocumentoFiscal para herdar seus campos

	// Relacionamentos
	Emitente     *Empresa `gorm:"foreignKey:EmitenteID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"emitente,omitempty"`
	Remetente    *Empresa `gorm:"foreignKey:RemetenteID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"remetente,omitempty"`
	Destinatario *Empresa `gorm:"foreignKey:DestinatarioID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"destinatario,omitempty"`
	Tomador      *Empresa `gorm:"foreignKey:TomadorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"tomador,omitempty"`

	// Campos específicos de CT-e
	RemetenteID     uuid.UUID  `json:"remetente_id" gorm:"type:uuid;index"`
	DestinatarioID  uuid.UUID  `json:"destinatario_id" gorm:"type:uuid;index"`
	TomadorID       *uuid.UUID `json:"tomador_id" gorm:"type:uuid;index"`
	ModalidadeFrete string     `json:"modalidade_frete" gorm:"size:3;index"` // CIF, FOB
	CFOP            string     `json:"cfop" gorm:"size:4;index"`

	// Valores específicos
	ValorICMS     float64 `json:"valor_icms"`
	ValorCarga    float64 `json:"valor_carga"`
	ValorPedagio  float64 `json:"valor_pedagio"`
	OutrosValores float64 `json:"outros_valores"`

	// Informações adicionais
	PlacaVeiculo string `json:"placa_veiculo" gorm:"size:10;index"`
	RNTRC        string `json:"rntrc" gorm:"size:20"`
	ObsGerais    string `json:"obs_gerais" gorm:"type:text"`

	// Documentos vinculados ao MDF-e
	MDFes []MDFE `gorm:"many2many:mdfe_ctes;" json:"mdfes,omitempty"`
}

// TableName define o nome da tabela no banco de dados
func (CTE) TableName() string {
	return "ctes"
}

// BeforeCreate hook do GORM
func (c *CTE) BeforeCreate(tx *gorm.DB) error {
	// Chamar o BeforeCreate do DocumentoFiscal
	if err := c.DocumentoFiscal.BeforeCreate(tx); err != nil {
		return err
	}

	// Validações específicas do CT-e
	if c.RemetenteID == uuid.Nil {
		return errors.New("remetente é obrigatório")
	}
	if c.DestinatarioID == uuid.Nil {
		return errors.New("destinatário é obrigatório")
	}
	if c.CFOP == "" {
		return errors.New("CFOP é obrigatório")
	}
	if c.ModalidadeFrete == "" {
		c.ModalidadeFrete = "CIF" // Padrão
	}

	return nil
}

// GetValorTotal retorna o valor total do CT-e
func (c *CTE) GetValorTotal() float64 {
	return c.ValorTotal
}

// IsValid verifica se o CT-e está válido
func (c *CTE) IsValid() bool {
	return c.Status == "100" && !c.Cancelado
}
