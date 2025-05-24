#!/bin/bash

# Script para iniciar os containers Docker
# Funciona no Linux e Windows (Git Bash/WSL)

set -e

echo "🚀 Iniciando Destack Transport API com Docker..."
echo ""

# Verificar se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker Desktop."
    exit 1
fi

# Criar diretórios necessários
echo "📁 Criando diretórios..."
mkdir -p logs uploads docker/nginx/ssl docker/prometheus docker/grafana/provisioning

# Copiar arquivo de ambiente se não existir
if [ ! -f .env ]; then
    echo "📝 Criando arquivo .env..."
    cp .env.docker .env
fi

# Limpar containers antigos (opcional)
read -p "Deseja limpar containers antigos? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 Limpando containers antigos..."
    docker-compose down -v
fi

# Construir imagens
echo "🔨 Construindo imagens..."
docker-compose build --no-cache

# Iniciar serviços básicos
echo "🐳 Iniciando serviços básicos (postgres + api)..."
docker-compose up -d postgres

# Aguardar PostgreSQL ficar pronto
echo "⏳ Aguardando PostgreSQL..."
sleep 10

# Iniciar API
echo "🚀 Iniciando API..."
docker-compose up -d api

# Verificar status
echo ""
echo "✅ Serviços iniciados!"
echo ""
docker-compose ps

echo ""
echo "📊 Logs da API:"
docker-compose logs --tail=20 api

echo ""
echo "🌐 Acesse a API em: http://localhost:8080"
echo "📚 Documentação: http://localhost:8080/api"
echo ""
echo "💡 Comandos úteis:"
echo "  - Ver logs: docker-compose logs -f api"
echo "  - Parar tudo: docker-compose down"
echo "  - Reiniciar API: docker-compose restart api"
echo ""

# Iniciar serviços opcionais
read -p "Deseja iniciar PGAdmin? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose --profile tools up -d pgadmin
    echo "🗄️  PGAdmin disponível em: http://localhost:5050"
fi