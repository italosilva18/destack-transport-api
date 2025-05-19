package models

import (
	"time"
)

// Upload representa um registro de upload de arquivo
type Upload struct {
	BaseModel
	NomeArquivo           string    `json:"nome_arquivo" gorm:"not null"`
	Status                string    `json:"status" gorm:"index;not null"`
	DataUpload            time.Time `json:"data_upload" gorm:"index;not null"`
	ChaveDocProcessado    *string   `json:"chave_doc_processado" gorm:"index"`
	DetalhesProcessamento string    `json:"detalhes_processamento"`
}

// TableName define o nome da tabela no banco de dados
func (Upload) TableName() string {
	return "uploads"
}
