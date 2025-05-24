#!/bin/bash

# Script de instalação para Destack Transport API

set -e

echo "==================================="
echo "Destack Transport API - Instalação"
echo "==================================="
echo ""

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo "❌ Go não está instalado. Por favor, instale Go 1.23+ primeiro."
    echo "   Visite: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go instalado: versão $GO_VERSION"

# Verificar se PostgreSQL está instalado
if ! command -v psql &> /dev/null; then
    echo "⚠️  PostgreSQL não encontrado. Certifique-se de ter o PostgreSQL instalado."
else
    echo "✅ PostgreSQL encontrado"
fi

# Verificar se Docker está instalado (opcional)
if command -v docker &> /dev/null; then
    echo "✅ Docker encontrado (opcional)"
else
    echo "ℹ️  Docker não encontrado (opcional)"
fi

echo ""
echo "📦 Instalando dependências do Go..."
go mod download
go mod verify

echo ""
echo "🔧 Criando arquivo de configuração..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✅ Arquivo .env criado. Por favor, edite com suas configurações."
else
    echo "ℹ️  Arquivo .env já existe."
fi

echo ""
echo "📁 Criando diretórios necessários..."
mkdir -p tmp logs

echo ""
echo "🔨 Compilando aplicação..."
go build -o tmp/destack-api ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida!"
else
    echo "❌ Erro na compilação"
    exit 1
fi

echo ""
echo "🗄️  Configuração do banco de dados"
echo "================================="
read -p "Deseja inicializar o banco de dados agora? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f scripts/init-db.sql ]; then
        echo "Executando script de inicialização..."
        psql -U postgres -f scripts/init-db.sql
        echo "✅ Banco de dados inicializado"
    else
        echo "❌ Script de inicialização não encontrado"
    fi
else
    echo "ℹ️  Você pode inicializar o banco mais tarde com: make db-init"
fi

echo ""
echo "🚀 Instalação concluída!"
echo ""
echo "Próximos passos:"
echo "1. Edite o arquivo .env com suas configurações"
echo "2. Execute 'make run' para iniciar a aplicação"
echo "3. Ou use 'docker-compose up' para executar com Docker"
echo ""
echo "Para ver todos os comandos disponíveis, use: make help"
echo ""
echo "Documentação da API estará disponível em: http://localhost:8080/swagger"
echo "(após implementar Swagger)"
echo ""
echo "Bom desenvolvimento! 🎉"