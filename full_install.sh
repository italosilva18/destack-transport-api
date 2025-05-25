#!/bin/bash

# Script de Instalação Completa para Destack Transport API
# Tenta configurar para Docker ou, como alternativa, para execução local.

set -e # Sai imediatamente se um comando falhar

echo "==========================================="
echo " Destack Transport API - Instalação Completa "
echo "==========================================="
echo ""

# --- Funções Auxiliares ---
check_command() {
    command -v "$1" &> /dev/null
}

# --- Verificação de Dependências Essenciais ---
echo "🔎 Verificando dependências..."

if ! check_command go; then
    echo "❌ Go não está instalado. Por favor, instale Go 1.23+ primeiro."
    echo "   Visite: https://golang.org/dl/"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go instalado: versão $GO_VERSION"

DOCKER_AVAILABLE=false
if check_command docker && check_command docker-compose; then
    echo "✅ Docker e Docker Compose encontrados."
    DOCKER_AVAILABLE=true
else
    echo "⚠️ Docker ou Docker Compose não encontrado. A instalação via Docker não será possível."
    echo "   Para Docker: Instale Docker Desktop (https://www.docker.com/products/docker-desktop)"
fi

echo ""

# --- Configuração de Diretórios e Arquivos Iniciais ---
echo "📁 Configurando diretórios Docker..."
if [ -f ./setup-docker-dirs.sh ]; then
    ./setup-docker-dirs.sh #
else
    echo "⚠️ Script setup-docker-dirs.sh não encontrado. Pulando esta etapa."
fi
echo ""

echo "🔧 Configurando arquivo de ambiente .env..."
if [ ! -f .env ]; then
    if [ -f .env.docker ]; then
        echo "📝 Copiando .env.docker para .env..."
        cp .env.docker .env
        echo "✅ .env criado a partir de .env.docker. Ajuste DB_HOST se for rodar localmente."
    elif [ -f .env.example ]; then
        echo "📝 Copiando .env.example para .env..."
        cp .env.example .env
        echo "✅ .env criado a partir de .env.example. Por favor, edite-o com suas configurações."
    else
        echo "❌ Nenhum arquivo .env.docker ou .env.example encontrado. Crie o .env manualmente."
        exit 1
    fi
else
    echo "ℹ️ Arquivo .env já existe. Verifique suas configurações."
fi
echo ""

# --- Instalação das Dependências Go ---
echo "📦 Instalando dependências do Go..."
go mod download #
go mod verify #
echo "✅ Dependências do Go instaladas."
echo ""

# --- Tentativa de Instalação com Docker (se disponível e desejado) ---
if [ "$DOCKER_AVAILABLE" = true ]; then
    read -p "🚀 Deseja tentar a instalação e execução com Docker agora? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🐳 Iniciando com Docker Compose..."
        if [ ! -f Dockerfile ]; then
            echo "❌ ERRO: Dockerfile não encontrado! Não é possível construir a imagem da API."
            echo "   Por favor, adicione o Dockerfile ao projeto e tente novamente."
        else
            echo "   (Isso pode levar alguns minutos na primeira vez, pois construirá a imagem)"
            if docker-compose up -d --build; then
                echo "✅ Serviços Docker iniciados com sucesso!"
                echo "   API deve estar acessível em http://localhost:8080 (verifique a porta no .env)"
                echo "   Para ver os logs: docker-compose logs -f api"
                echo "🎉 Instalação Docker concluída!"
                exit 0
            else
                echo "❌ Falha ao iniciar serviços Docker. Verifique os logs acima."
                echo "   Você pode tentar rodar localmente ou corrigir os problemas do Docker."
            fi
        fi
    else
        echo "ℹ️ Instalação com Docker pulada."
    fi
fi
echo ""

# --- Tentativa de Instalação Local (se Docker falhou ou não foi escolhido) ---
echo "🛠️ Prosseguindo com a tentativa de configuração para execução local..."

echo "🔨 Compilando aplicação Go..."
if go build -o tmp/destack-api ./cmd/server; then #
    echo "✅ Aplicação compilada com sucesso (./tmp/destack-api)."
else
    echo "❌ Falha na compilação da aplicação Go. Verifique os erros."
    exit 1
fi
echo ""

echo "🗄️ Configuração do banco de dados PostgreSQL local..."
if ! check_command psql; then
    echo "⚠️ psql (PostgreSQL client) não encontrado. Não será possível inicializar o banco de dados local automaticamente."
    echo "   Por favor, instale o PostgreSQL e certifique-se que 'psql' está no PATH."
    echo "   Você precisará criar o banco de dados e o usuário manualmente conforme o .env e o script 'scripts/init-db.sql'."
else
    echo "✅ psql encontrado."
    read -p "Deseja tentar inicializar o banco de dados local agora (requer psql e acesso)? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ -f scripts/init-db.sql ]; then
            echo "   Por favor, insira a senha do usuário superusuário do PostgreSQL (ex: 'postgres') se solicitado."
            # Você pode precisar ajustar -U <seu_usuario_postgres>
            if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f scripts/init-db.sql &>/dev/null || psql -U postgres -f scripts/init-db.sql ; then
                 echo "✅ Script de inicialização do banco de dados local executado (ou tentado)."
                 echo "   Verifique se não houve erros e se o banco 'destack_transport' e as extensões foram criados."
            else
                 echo "❌ Falha ao executar o script de inicialização do banco. Verifique os erros e as configurações no .env."
                 echo "   Comando tentado: psql -U postgres -f scripts/init-db.sql"
            fi
        else
            echo "❌ Script scripts/init-db.sql não encontrado."
        fi
    else
        echo "ℹ️ Inicialização do banco de dados local pulada. Configure-o manualmente."
    fi
fi
echo ""

echo "🎉 Configuração para execução local concluída (ou tentada)!"
echo "   Lembre-se de:"
echo "   1. Garantir que o PostgreSQL esteja rodando e acessível."
echo "   2. Que o banco de dados 'destack_transport' exista e o usuário no .env tenha permissões."
echo "   3. Ajustar o arquivo .env (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME) para seu ambiente local."
echo ""
echo "🚀 Para rodar a aplicação localmente (após configurar o BD):"
echo "   ./tmp/destack-api"
echo "   Ou use: make run"
echo ""
echo "✨ Fim do script de instalação completa."