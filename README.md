# Destack Transport API

API REST para gerenciamento de documentos fiscais de transporte (CT-e e MDF-e), desenvolvida em Go com Gin Framework.

## 🚀 Tecnologias

- **Go 1.23+**
- **Gin Framework** - Framework web
- **GORM** - ORM para Go
- **PostgreSQL** - Banco de dados
- **JWT** - Autenticação
- **Docker** - Containerização
- **Zerolog** - Sistema de logs

## 📋 Pré-requisitos

- Go 1.23 ou superior
- PostgreSQL 15+
- Docker e Docker Compose (opcional)
- Make (opcional)

## 🔧 Instalação

### Método 1: Usando Docker Compose (Recomendado)

```bash
# Clone o repositório
git clone https://github.com/italosilva18/destack-transport-api.git
cd destack-transport-api

# Copie o arquivo de ambiente
cp .env.example .env

# Inicie os serviços
docker-compose up -d
```

### Método 2: Instalação Manual

```bash
# Clone o repositório
git clone https://github.com/italosilva18/destack-transport-api.git
cd destack-transport-api

# Instale as dependências
go mod download

# Configure o banco de dados PostgreSQL
# Crie um banco de dados chamado 'destack_transport'

# Copie e configure o arquivo de ambiente
cp .env.example .env
# Edite o arquivo .env com suas configurações

# Execute as migrações
go run cmd/server/main.go

# Ou use o Makefile
make run
```

## ⚙️ Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
# Ambiente e servidor
ENVIRONMENT=development
SERVER_PORT=8080

# Banco de dados PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=destack_transport
DB_SSLMODE=disable

# JWT
JWT_SECRET=sua_chave_secreta_aqui
JWT_EXPIRES_IN=24
```

## 📚 Estrutura do Projeto

```
destack-transport-api/
├── cmd/
│   └── server/
│       └── main.go         # Ponto de entrada da aplicação
├── configs/
│   └── config.go           # Configurações da aplicação
├── internal/
│   ├── api/
│   │   ├── handlers/       # Controllers/Handlers
│   │   ├── middlewares/    # Middlewares HTTP
│   │   └── routes/         # Definição de rotas
│   ├── models/             # Modelos de dados
│   ├── parsers/            # Parsers de XML
│   └── services/           # Lógica de negócio
├── pkg/
│   ├── database/           # Conexão e migrações
│   └── logger/             # Sistema de logs
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🚀 Executando o Projeto

### Desenvolvimento

```bash
# Usando Make
make run

# Ou diretamente
go run cmd/server/main.go

# Ou usando o script de desenvolvimento
./dev.sh
```

### Produção

```bash
# Build da aplicação
make build

# Executar
./tmp/destack-api
```

## 📡 API Endpoints

### Autenticação

```http
POST   /api/auth/login      # Login
POST   /api/auth/logout     # Logout
GET    /api/auth/profile    # Perfil do usuário (autenticado)
```

### CT-e (Conhecimento de Transporte Eletrônico)

```http
GET    /api/ctes                      # Listar CT-es
GET    /api/ctes/:chave               # Buscar CT-e por chave
GET    /api/ctes/:chave/download-xml  # Download do XML
GET    /api/ctes/:chave/dacte         # Gerar DACTE
POST   /api/ctes/:chave/reprocess     # Reprocessar CT-e
GET    /api/paineis/cte               # Painel de CT-e
```

### MDF-e (Manifesto Eletrônico de Documentos Fiscais)

```http
GET    /api/mdfes                      # Listar MDF-es
GET    /api/mdfes/:chave               # Buscar MDF-e por chave
GET    /api/mdfes/:chave/download-xml  # Download do XML
GET    /api/mdfes/:chave/damdfe        # Gerar DAMDFE
POST   /api/mdfes/:chave/reprocess     # Reprocessar MDF-e
POST   /api/mdfes/:chave/encerrar      # Encerrar MDF-e
GET    /api/mdfes/:chave/documentos    # Documentos vinculados
GET    /api/paineis/mdfe               # Painel de MDF-e
```

### Upload de Arquivos

```http
POST   /api/upload/single    # Upload de um arquivo XML
POST   /api/upload/batch     # Upload de múltiplos arquivos
GET    /api/uploads          # Listar uploads
GET    /api/uploads/:id      # Buscar upload por ID
DELETE /api/uploads/:id      # Excluir upload
```

### Dashboard

```http
GET    /api/dashboard/cards         # Cards do dashboard
GET    /api/dashboard/lancamentos   # Últimos lançamentos
GET    /api/dashboard/cif-fob       # Dados CIF/FOB
```

### Financeiro

```http
GET    /api/financeiro                     # Dados financeiros
GET    /api/financeiro/faturamento-mensal  # Faturamento mensal
GET    /api/financeiro/agrupado            # Dados agrupados
GET    /api/financeiro/detalhes/:tipo/:id  # Detalhes de item
```

### Geográfico

```http
GET    /api/geografico           # Dados geográficos
GET    /api/geografico/origens   # Top origens
GET    /api/geografico/destinos  # Top destinos
GET    /api/geografico/rotas     # Rotas frequentes
GET    /api/geografico/fluxo-ufs # Fluxo entre UFs
```

### Manutenções

```http
POST   /api/manutencoes              # Criar manutenção
GET    /api/manutencoes              # Listar manutenções
GET    /api/manutencoes/:id          # Buscar manutenção
PUT    /api/manutencoes/:id          # Atualizar manutenção
DELETE /api/manutencoes/:id          # Excluir manutenção
GET    /api/manutencoes/estatisticas # Estatísticas
```

## 🔐 Autenticação

A API utiliza JWT (JSON Web Tokens) para autenticação. Para acessar endpoints protegidos:

1. Faça login em `/api/auth/login`
2. Use o token retornado no header: `Authorization: Bearer {token}`

### Exemplo de Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

## 📊 Modelos de Dados

### User
- Gerenciamento de usuários do sistema
- Autenticação e autorização

### Empresa
- Cadastro de empresas (clientes, fornecedores, emitentes)
- Suporta CNPJ e CPF

### CTE
- Conhecimento de Transporte Eletrônico
- Modalidades: CIF, FOB

### MDFE
- Manifesto Eletrônico de Documentos Fiscais
- Controle de encerramento

### Veiculo
- Cadastro de veículos
- Tipos: PROPRIO, AGREGADO, TERCEIRO

### Manutencao
- Controle de manutenções de veículos
- Status: PENDENTE, AGENDADO, CONCLUIDO, PAGO, CANCELADO

### Upload
- Controle de uploads de XML
- Processamento assíncrono

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Ou
go test -v ./...
```

## 🛠️ Desenvolvimento

### Estrutura de um Handler

```go
type ExemploHandler struct {
    db     *gorm.DB
    logger logger.Logger
}

func NewExemploHandler(db *gorm.DB) *ExemploHandler {
    return &ExemploHandler{
        db:     db,
        logger: logger.GetLogger(),
    }
}
```

### Adicionando uma Nova Rota

1. Crie o handler em `internal/api/handlers/`
2. Crie o arquivo de rotas em `internal/api/routes/`
3. Registre as rotas em `internal/api/routes/routes.go`

## 📝 Logs

Os logs são gerenciados pelo Zerolog e salvos em:
- Desenvolvimento: Console
- Produção: `logs/api.log`

## 🐳 Docker

### Build da Imagem

```bash
docker build -t destack-api .
```

### Executar Container

```bash
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  destack-api
```

## 📦 Makefile

Comandos disponíveis:

```bash
make build         # Compila a aplicação
make run           # Executa a aplicação
make test          # Executa os testes
make clean         # Limpa arquivos temporários
make docker        # Build da imagem Docker
make docker-compose # Inicia com Docker Compose
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie sua Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a Branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 👥 Autores

- **Italo Silva** - *Trabalho Inicial* - [italosilva18](https://github.com/italosilva18)

## 🎯 Status do Projeto

O projeto está em desenvolvimento ativo. As seguintes funcionalidades estão planejadas:

- [ ] Sistema de alertas
- [ ] Geração de relatórios
- [ ] Configurações de empresa
- [ ] Integração com SEFAZ
- [ ] Dashboard em tempo real
- [ ] Notificações por email

## 📞 Suporte

Para suporte, envie um email para suporte@destack.com.br ou abra uma issue no GitHub.