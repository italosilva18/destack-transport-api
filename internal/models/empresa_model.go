package models

// Empresa representa uma empresa (cliente, fornecedor, etc.)
type Empresa struct {
	BaseModel
	CNPJ         *string `json:"cnpj" gorm:"uniqueIndex"`
	CPF          *string `json:"cpf" gorm:"uniqueIndex"`
	RazaoSocial  string  `json:"razao_social" gorm:"not null"`
	NomeFantasia *string `json:"nome_fantasia"`
	IE           *string `json:"ie" gorm:"index"` // Inscrição Estadual
	IM           *string `json:"im"`              // Inscrição Municipal
	UF           string  `json:"uf" gorm:"size:2;index"`
	Municipio    string  `json:"municipio" gorm:"size:100"`
	CEP          string  `json:"cep" gorm:"size:8"`
	Logradouro   string  `json:"logradouro"`
	Numero       string  `json:"numero" gorm:"size:10"`
	Complemento  *string `json:"complemento"`
	Bairro       string  `json:"bairro"`
	Telefone     *string `json:"telefone" gorm:"size:20"`
	Email        *string `json:"email" gorm:"size:100"`
	Ativo        bool    `json:"ativo" gorm:"default:true"`

	// Campos adicionais podem ser incluídos
}

// TableName define o nome da tabela no banco de dados
func (Empresa) TableName() string {
	return "empresas"
}

// ValidarCNPJ valida o CNPJ
func ValidarCNPJ(cnpj string) bool {
	// Remove caracteres não numéricos
	cnpj = removeNonDigits(cnpj)

	// CNPJ deve ter 14 dígitos
	if len(cnpj) != 14 {
		return false
	}

	// Verifica se todos os dígitos são iguais
	if allDigitsSame(cnpj) {
		return false
	}

	// Calcula primeiro dígito verificador
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(cnpj[i] - '0')
		if i < 4 {
			sum += digit * (5 - i)
		} else {
			sum += digit * (13 - i)
		}
	}
	remainder := sum % 11
	d1 := 0
	if remainder >= 2 {
		d1 = 11 - remainder
	}

	// Calcula segundo dígito verificador
	sum = 0
	for i := 0; i < 13; i++ {
		digit := int(cnpj[i] - '0')
		if i < 5 {
			sum += digit * (6 - i)
		} else {
			sum += digit * (14 - i)
		}
	}
	remainder = sum % 11
	d2 := 0
	if remainder >= 2 {
		d2 = 11 - remainder
	}

	// Verifica se os dígitos calculados conferem
	return int(cnpj[12]-'0') == d1 && int(cnpj[13]-'0') == d2
}

// ValidarCPF valida o CPF
func ValidarCPF(cpf string) bool {
	// Remove caracteres não numéricos
	cpf = removeNonDigits(cpf)

	// CPF deve ter 11 dígitos
	if len(cpf) != 11 {
		return false
	}

	// Verifica se todos os dígitos são iguais
	if allDigitsSame(cpf) {
		return false
	}

	// Calcula primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	remainder := sum % 11
	d1 := 0
	if remainder >= 2 {
		d1 = 11 - remainder
	}

	// Calcula segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	remainder = sum % 11
	d2 := 0
	if remainder >= 2 {
		d2 = 11 - remainder
	}

	// Verifica se os dígitos calculados conferem
	return int(cpf[9]-'0') == d1 && int(cpf[10]-'0') == d2
}

// removeNonDigits remove caracteres não numéricos
func removeNonDigits(s string) string {
	result := ""
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}

// allDigitsSame verifica se todos os dígitos são iguais
func allDigitsSame(s string) bool {
	if len(s) == 0 {
		return true
	}
	first := s[0]
	for _, c := range s {
		if byte(c) != first {
			return false
		}
	}
	return true
}
