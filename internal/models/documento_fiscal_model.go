package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentoFiscal representa a base para CT-e e MDF-e
type DocumentoFiscal struct {
	BaseModel
	Chave       string    `json:"chave" gorm:"uniqueIndex;not null;size:44"`
	Tipo        string    `json:"tipo" gorm:"index;not null;size:10"` // CTE, MDFE
	Numero      int       `json:"numero" gorm:"not null"`
	Serie       string    `json:"serie" gorm:"not null;size:3"`
	DataEmissao time.Time `json:"data_emissao" gorm:"index;not null"`
	Status      string    `json:"status" gorm:"index;not null;size:3;default:'000'"`
	Cancelado   bool      `json:"cancelado" gorm:"index;default:false"`
	ValorTotal  float64   `json:"valor_total"`

	// Protocolo de autorização
	Protocolo       string     `json:"protocolo" gorm:"size:20"`
	DataAutorizacao *time.Time `json:"data_autorizacao"`

	// Entidades principais
	EmitenteID uuid.UUID `json:"emitente_id" gorm:"type:uuid;index;not null"`

	// Dados de localização
	UFInicio        string `json:"uf_inicio" gorm:"size:2;index"`
	UFDestino       string `json:"uf_destino" gorm:"size:2;index"`
	MunicipioInicio string `json:"municipio_inicio" gorm:"size:100"`
	MunicipioFim    string `json:"municipio_fim" gorm:"size:100"`

	// Metadados de processamento
	DataProcessamento *time.Time `json:"data_processamento"`
	XMLOriginal       string     `json:"-" gorm:"type:text"` // Armazenar XML completo

	// Dados de upload
	UploadID *uuid.UUID `json:"upload_id" gorm:"type:uuid;index"`
	Upload   *Upload    `gorm:"foreignKey:UploadID" json:"upload,omitempty"`
}

// BeforeCreate hook do GORM para DocumentoFiscal
func (d *DocumentoFiscal) BeforeCreate(tx *gorm.DB) error {
	// Validar chave de acesso
	if len(d.Chave) != 44 {
		return errors.New("chave de acesso deve ter 44 caracteres")
	}

	// Validar tipo
	if d.Tipo != "CTE" && d.Tipo != "MDFE" {
		return errors.New("tipo deve ser CTE ou MDFE")
	}

	// Validar UFs
	if len(d.UFInicio) != 2 || len(d.UFDestino) != 2 {
		return errors.New("UF deve ter 2 caracteres")
	}

	// Definir data de processamento
	agora := time.Now()
	d.DataProcessamento = &agora

	return nil
}

// IsCancelado verifica se o documento está cancelado
func (d *DocumentoFiscal) IsCancelado() bool {
	return d.Cancelado
}

// IsAutorizado verifica se o documento está autorizado
func (d *DocumentoFiscal) IsAutorizado() bool {
	return d.Status == "100" && d.Protocolo != ""
}

// GetDescricaoStatus retorna a descrição do status
func (d *DocumentoFiscal) GetDescricaoStatus() string {
	statusMap := map[string]string{
		"100": "Autorizado",
		"101": "Cancelado",
		"102": "Inutilizado",
		"103": "Denegado",
		"104": "Recusado",
		"000": "Não processado",
	}

	if desc, ok := statusMap[d.Status]; ok {
		return desc
	}
	return "Status desconhecido"
}
