package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MDFE representa um Manifesto Eletrônico de Documentos Fiscais
type MDFE struct {
	DocumentoFiscal // Embedar DocumentoFiscal para herdar seus campos

	// Relacionamentos
	VeiculoTracao *Veiculo `gorm:"foreignKey:VeiculoTracaoID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"veiculo_tracao,omitempty"`
	CTes          []CTE    `gorm:"many2many:mdfe_ctes;" json:"ctes,omitempty"`

	// Campos específicos de MDF-e
	VeiculoTracaoID uuid.UUID `json:"veiculo_tracao_id" gorm:"type:uuid;index"`

	// Condutor
	NomeMotorista string `json:"nome_motorista" gorm:"size:100"`
	CPFMotorista  string `json:"cpf_motorista" gorm:"size:11;index"`

	// Totalizadores
	QtdCTe         int     `json:"qtd_cte"`
	QtdNFe         int     `json:"qtd_nfe"`
	PesoBrutoTotal float64 `json:"peso_bruto_total"`

	// Informações da carga
	ProdutoPredominante string `json:"produto_predominante" gorm:"size:100"`
	TipoCarga           string `json:"tipo_carga" gorm:"size:2"`

	// Status específicos
	Encerrado         bool       `json:"encerrado" gorm:"index;default:false"`
	DataEncerramento  *time.Time `json:"data_encerramento"`
	LocalEncerramento string     `json:"local_encerramento" gorm:"size:100"`

	// Seguro
	SeguradoraNome  string `json:"seguradora_nome" gorm:"size:100"`
	SeguradoraCNPJ  string `json:"seguradora_cnpj" gorm:"size:14"`
	NumeroApolice   string `json:"numero_apolice" gorm:"size:50"`
	NumeroAverbacao string `json:"numero_averbacao" gorm:"size:50"`
}

// TableName define o nome da tabela no banco de dados
func (MDFE) TableName() string {
	return "mdfes"
}

// BeforeCreate hook do GORM
func (m *MDFE) BeforeCreate(tx *gorm.DB) error {
	// Chamar o BeforeCreate do DocumentoFiscal
	if err := m.DocumentoFiscal.BeforeCreate(tx); err != nil {
		return err
	}

	// Validações específicas do MDF-e
	if m.VeiculoTracaoID == uuid.Nil {
		return errors.New("veículo de tração é obrigatório")
	}
	if m.CPFMotorista == "" {
		return errors.New("CPF do motorista é obrigatório")
	}
	if m.NomeMotorista == "" {
		return errors.New("nome do motorista é obrigatório")
	}

	// Tipo é sempre MDFE
	m.Tipo = "MDFE"

	return nil
}

// Encerrar encerra o MDF-e
func (m *MDFE) Encerrar(local string) error {
	if m.Encerrado {
		return errors.New("MDF-e já está encerrado")
	}

	agora := time.Now()
	m.Encerrado = true
	m.DataEncerramento = &agora
	m.LocalEncerramento = local

	return nil
}

// PodeEncerrar verifica se o MDF-e pode ser encerrado
func (m *MDFE) PodeEncerrar() bool {
	return m.Status == "100" && !m.Cancelado && !m.Encerrado
}

// AddCTe adiciona um CT-e ao MDF-e
func (m *MDFE) AddCTe(cte *CTE) error {
	// Verificar se o CT-e já não está vinculado
	for _, c := range m.CTes {
		if c.ID == cte.ID {
			return errors.New("CT-e já está vinculado a este MDF-e")
		}
	}

	// Verificar se as UFs são compatíveis
	if cte.UFInicio != m.UFInicio && cte.UFDestino != m.UFDestino {
		return errors.New("CT-e com rota incompatível com o MDF-e")
	}

	m.CTes = append(m.CTes, *cte)
	m.QtdCTe++
	m.ValorTotal += cte.ValorTotal

	return nil
}
