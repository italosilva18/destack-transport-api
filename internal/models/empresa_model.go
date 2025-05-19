package models

// Empresa representa uma empresa (cliente, fornecedor, etc.)
type Empresa struct {
	BaseModel
	CNPJ         *string `json:"cnpj" gorm:"uniqueIndex"`
	CPF          *string `json:"cpf" gorm:"uniqueIndex"`
	RazaoSocial  string  `json:"razao_social" gorm:"not null"`
	NomeFantasia *string `json:"nome_fantasia"`
	UF           string  `json:"uf" gorm:"size:2;index"`

	// Campos adicionais podem ser incluídos
}

// TableName define o nome da tabela no banco de dados
func (Empresa) TableName() string {
	return "empresas"
}
