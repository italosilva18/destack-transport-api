package models

// Veiculo representa um veículo no sistema
type Veiculo struct {
	BaseModel
	Placa   string  `json:"placa" gorm:"uniqueIndex;size:7"`
	RENAVAM *string `json:"renavam"`
	Tipo    string  `json:"tipo" gorm:"size:10;index"` // PROPRIO, AGREGADO, TERCEIRO

	// Campos adicionais podem ser incluídos
}

// TableName define o nome da tabela no banco de dados
func (Veiculo) TableName() string {
	return "veiculos"
}
