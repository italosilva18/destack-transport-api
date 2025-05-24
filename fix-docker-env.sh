#!/bin/bash

echo "🔧 Corrigindo problema de variáveis de ambiente no Docker..."

# 1. Parar containers
echo "Parando containers..."
docker-compose down

# 2. Criar arquivo .env correto
echo "Criando arquivo .env..."
cat > .env << 'EOF'
# Ambiente e servidor
ENVIRONMENT=production
SERVER_PORT=8080

# Banco de dados PostgreSQL
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=destack_transport
DB_SSLMODE=disable

# JWT
JWT_SECRET=your_jwt_secret_key_change_this_in_production
JWT_EXPIRES_IN=24
EOF

# 3. Criar arquivo app.env
echo "Criando arquivo app.env..."
cp .env app.env

# 4. Tornar o script de entrada executável
chmod +x docker-entrypoint.sh

# 5. Rebuild da imagem
echo "Reconstruindo imagem Docker..."
docker-compose build --no-cache api

# 6. Iniciar novamente
echo "Iniciando containers..."
docker-compose up -d

# 7. Aguardar inicialização
echo "Aguardando inicialização..."
sleep 10

# 8. Verificar logs
echo "Verificando logs..."
docker-compose logs --tail=50 api

echo ""
echo "✅ Processo concluído!"
echo "Se ainda houver erros, execute:"
echo "  docker-compose logs -f api"