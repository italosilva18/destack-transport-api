# Script PowerShell para corrigir problema de variáveis de ambiente

Write-Host "🔧 Corrigindo problema de variáveis de ambiente no Docker..." -ForegroundColor Cyan

# 1. Parar containers
Write-Host "Parando containers..." -ForegroundColor Yellow
docker-compose down

# 2. Criar arquivo .env correto
Write-Host "Criando arquivo .env..." -ForegroundColor Yellow
@"
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
"@ | Out-File -FilePath ".env" -Encoding UTF8

# 3. Criar arquivo app.env
Write-Host "Criando arquivo app.env..." -ForegroundColor Yellow
Copy-Item -Path ".env" -Destination "app.env" -Force

# 4. Criar docker-entrypoint.sh se não existir
if (!(Test-Path "docker-entrypoint.sh")) {
    Write-Host "Criando docker-entrypoint.sh..." -ForegroundColor Yellow
@'
#!/bin/sh

# Script de entrada para criar app.env a partir das variáveis de ambiente

echo "Creating app.env from environment variables..."

cat > /app/app.env << EOF
# Generated from environment variables
ENVIRONMENT=${ENVIRONMENT:-production}
SERVER_PORT=${SERVER_PORT:-8080}

# Database
DB_HOST=${DB_HOST:-postgres}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-destack_transport}
DB_SSLMODE=${DB_SSLMODE:-disable}

# JWT
JWT_SECRET=${JWT_SECRET:-default_jwt_secret_change_this}
JWT_EXPIRES_IN=${JWT_EXPIRES_IN:-24}
EOF

echo "app.env created with content:"
cat /app/app.env

# Execute the application
exec /app/destack-api
'@ | Out-File -FilePath "docker-entrypoint.sh" -Encoding UTF8 -NoNewline
}

# 5. Rebuild da imagem
Write-Host "Reconstruindo imagem Docker..." -ForegroundColor Yellow
docker-compose build --no-cache api

# 6. Iniciar novamente
Write-Host "Iniciando containers..." -ForegroundColor Yellow
docker-compose up -d

# 7. Aguardar inicialização
Write-Host "Aguardando inicialização..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# 8. Verificar logs
Write-Host "Verificando logs..." -ForegroundColor Yellow
docker-compose logs --tail=50 api

Write-Host ""
Write-Host "✅ Processo concluído!" -ForegroundColor Green
Write-Host "Se ainda houver erros, execute:" -ForegroundColor Yellow
Write-Host "  docker-compose logs -f api" -ForegroundColor White