package models

import (
	"time"
)

// Manutencao representa uma manutenção de veículo
type Manutencao struct {
	BaseModel
	VeiculoID        string    `json:"veiculo_id" gorm:"index;not null"`
	DataServico      time.Time `json:"data_servico" gorm:"not null"`
	ServicoRealizado string    `json:"servico_realizado" gorm:"not null"`
	Oficina          string    `json:"oficina"`
	Quilometragem    *int      `json:"quilometragem"`
	PecaUtilizada    string    `json:"peca_utilizada"`
	NotaFiscal       string    `json:"nota_fiscal"`
	ValorPeca        float64   `json:"valor_peca" gorm:"default:0"`
	ValorMaoObra     float64   `json:"valor_mao_obra" gorm:"default:0"`
	Status           string    `json:"status" gorm:"not null;default:'PENDENTE'"`
	Observacoes      string    `json:"observacoes"`
}

// TableName define o nome da tabela no banco de dados
func (Manutencao) TableName() string {
	return "manutencoes"
}

// ValorTotal retorna o valor total da manutenção
func (m *Manutencao) ValorTotal() float64 {
	return m.ValorPeca + m.ValorMaoObra
}
