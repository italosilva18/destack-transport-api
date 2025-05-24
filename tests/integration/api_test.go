package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/italosilva18/destack-transport-api/internal/api/routes"
	"github.com/italosilva18/destack-transport-api/internal/models"
	"github.com/italosilva18/destack-transport-api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// APITestSuite é a suite de testes de integração
type APITestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
	token  string
}

// SetupSuite configura o ambiente de teste
func (suite *APITestSuite) SetupSuite() {
	// Configurar logger
	logger.InitLogger()

	// Criar banco de dados em memória
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.NoError(err)

	// Migrar modelos
	err = db.AutoMigrate(
		&models.User{},
		&models.Empresa{},
		&models.Veiculo{},
		&models.CTE{},
		&models.MDFE{},
		&models.Upload{},
		&models.Manutencao{},
	)
	suite.NoError(err)

	// Criar usuário de teste
	user := models.User{
		Name:     "Admin Test",
		Username: "admin",
		Email:    "admin@test.com",
		Password: "admin123",
		Role:     "admin",
		Active:   true,
	}
	db.Create(&user)

	// Configurar router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	routes.SetupRoutes(router, db)

	suite.db = db
	suite.router = router

	// Fazer login para obter token
	suite.login()
}

// TearDownSuite limpa o ambiente de teste
func (suite *APITestSuite) TearDownSuite() {
	// Limpar recursos se necessário
}

// login faz login e armazena o token JWT
func (suite *APITestSuite) login() {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	suite.token = response["token"].(string)
}

// makeAuthRequest cria uma requisição autenticada
func (suite *APITestSuite) makeAuthRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonData, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}

	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// TestHealthCheck testa o endpoint de saúde
func (suite *APITestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "API Destack Transport")
}

// TestAuthFlow testa o fluxo de autenticação
func (suite *APITestSuite) TestAuthFlow() {
	// Teste 1: Login com credenciais válidas
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NotEmpty(suite.T(), loginResponse["token"])

	// Teste 2: Acessar perfil com token
	token := loginResponse["token"].(string)
	req, _ = http.NewRequest("GET", "/api/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestCTECRUD testa as operações CRUD de CT-e
func (suite *APITestSuite) TestCTECRUD() {
	// Criar empresas necessárias
	emitente := models.Empresa{
		CNPJ:        &[]string{"12345678901234"}[0],
		RazaoSocial: "Emitente Test",
		UF:          "SP",
	}
	suite.db.Create(&emitente)

	destinatario := models.Empresa{
		CNPJ:        &[]string{"98765432109876"}[0],
		RazaoSocial: "Destinatário Test",
		UF:          "RJ",
	}
	suite.db.Create(&destinatario)

	// Criar CT-e
	cte := models.CTE{
		DocumentoFiscal: models.DocumentoFiscal{
			Chave:       "31234567890123456789012345678901234567890123",
			Tipo:        "CTE",
			Numero:      123456,
			Serie:       "1",
			DataEmissao: time.Now(),
			Status:      "100",
			ValorTotal:  1500.50,
			EmitenteID:  emitente.ID.String(),
			UFInicio:    "SP",
			UFDestino:   "RJ",
		},
		DestinatarioID:  destinatario.ID.String(),
		CFOP:            "5353",
		ModalidadeFrete: "CIF",
	}
	suite.db.Create(&cte)

	// Teste 1: Listar CT-es
	w := suite.makeAuthRequest("GET", "/api/ctes", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var listResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &listResponse)
	assert.NotNil(suite.T(), listResponse["data"])

	// Teste 2: Buscar CT-e específico
	w = suite.makeAuthRequest("GET", "/api/ctes/"+cte.Chave, nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var cteResponse models.CTE
	json.Unmarshal(w.Body.Bytes(), &cteResponse)
	assert.Equal(suite.T(), cte.Chave, cteResponse.Chave)
}

// TestDashboard testa os endpoints do dashboard
func (suite *APITestSuite) TestDashboard() {
	// Teste 1: Cards do dashboard
	w := suite.makeAuthRequest("GET", "/api/dashboard/cards", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var cardsResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &cardsResponse)
	assert.NotNil(suite.T(), cardsResponse["total_ctes"])
	assert.NotNil(suite.T(), cardsResponse["valor_total_cte"])

	// Teste 2: Últimos lançamentos
	w = suite.makeAuthRequest("GET", "/api/dashboard/lancamentos", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestUpload testa o upload de arquivos
func (suite *APITestSuite) TestUpload() {
	// Criar um arquivo XML simulado
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<CTe xmlns="http://www.portalfiscal.inf.br/cte">
  <infCte Id="CTe31234567890123456789012345678901234567890123">
    <ide>
      <cUF>31</cUF>
      <cCT>12345678</cCT>
      <CFOP>5353</CFOP>
    </ide>
  </infCte>
</CTe>`

	// Criar multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("arquivo_xml", "test.xml")
	part.Write([]byte(xmlContent))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/upload/single", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusAccepted, w.Code)

	var uploadResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	assert.NotEmpty(suite.T(), uploadResponse["id"])
	assert.Contains(suite.T(), uploadResponse["message"], "processamento iniciado")
}

// TestManutencao testa o CRUD de manutenções
func (suite *APITestSuite) TestManutencao() {
	// Criar veículo
	veiculo := models.Veiculo{
		Placa: "ABC1234",
		Tipo:  "PROPRIO",
	}
	suite.db.Create(&veiculo)

	// Criar manutenção
	manutencaoData := map[string]interface{}{
		"veiculo_id":        veiculo.ID.String(),
		"data_servico":      time.Now().Format("2006-01-02"),
		"servico_realizado": "Troca de óleo",
		"oficina":           "Oficina Test",
		"valor_peca":        150.00,
		"valor_mao_obra":    100.00,
		"status":            "CONCLUIDO",
	}

	// Teste 1: Criar manutenção
	w := suite.makeAuthRequest("POST", "/api/manutencoes", manutencaoData)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse models.Manutencao
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NotEmpty(suite.T(), createResponse.ID)

	// Teste 2: Listar manutenções
	w = suite.makeAuthRequest("GET", "/api/manutencoes", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Teste 3: Estatísticas de manutenção
	w = suite.makeAuthRequest("GET", "/api/manutencoes/estatisticas", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var statsResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &statsResponse)
	assert.NotNil(suite.T(), statsResponse["total_manutencoes"])
}

// TestFinanceiro testa os endpoints financeiros
func (suite *APITestSuite) TestFinanceiro() {
	// Teste 1: Dados financeiros gerais
	w := suite.makeAuthRequest("GET", "/api/financeiro", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var financeiroResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &financeiroResponse)
	assert.NotNil(suite.T(), financeiroResponse["faturamento_total"])

	// Teste 2: Faturamento mensal
	w = suite.makeAuthRequest("GET", "/api/financeiro/faturamento-mensal", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestGeografico testa os endpoints geográficos
func (suite *APITestSuite) TestGeografico() {
	// Teste 1: Dados geográficos gerais
	w := suite.makeAuthRequest("GET", "/api/geografico", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var geoResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &geoResponse)
	assert.NotNil(suite.T(), geoResponse["total_origens"])

	// Teste 2: Top origens
	w = suite.makeAuthRequest("GET", "/api/geografico/origens", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Teste 3: Fluxo entre UFs
	w = suite.makeAuthRequest("GET", "/api/geografico/fluxo-ufs", nil)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestMain é o ponto de entrada dos testes
func TestMain(m *testing.M) {
	// Criar arquivo .env temporário para testes
	envContent := `
ENVIRONMENT=test
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=test
DB_PASSWORD=test
DB_NAME=test
DB_SSLMODE=disable
JWT_SECRET=test_secret_key_for_testing_only
JWT_EXPIRES_IN=24
`
	os.WriteFile(".env", []byte(envContent), 0644)

	// Executar testes
	code := m.Run()

	// Limpar
	os.Remove(".env")
	os.Exit(code)
}

// TestAPISuite executa a suite de testes
func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
