# Script PowerShell para iniciar os containers Docker no Windows
# Execute com: .\docker\scripts\docker-up.ps1

Write-Host "🚀 Iniciando Destack Transport API com Docker..." -ForegroundColor Green
Write-Host ""

# Verificar se Docker está rodando
try {
    docker info | Out-Null
} catch {
    Write-Host "❌ Docker não está rodando. Por favor, inicie o Docker Desktop." -ForegroundColor Red
    exit 1
}

# Criar diretórios necessários
Write-Host "📁 Criando diretórios..." -ForegroundColor Yellow
$directories = @("logs", "uploads", "docker\nginx\ssl", "docker\prometheus", "docker\grafana\provisioning")
foreach ($dir in $directories) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
}

# Copiar arquivo de ambiente se não existir
if (!(Test-Path ".env")) {
    Write-Host "📝 Criando arquivo .env..." -ForegroundColor Yellow
    Copy-Item ".env.docker" -Destination ".env"
}

# Perguntar se deve limpar containers antigos
$cleanContainers = Read-Host "Deseja limpar containers antigos? (s/N)"
if ($cleanContainers -eq 's' -or $cleanContainers -eq 'S') {
    Write-Host "🧹 Limpando containers antigos..." -ForegroundColor Yellow
    docker-compose down -v
}

# Construir imagens
Write-Host "🔨 Construindo imagens..." -ForegroundColor Yellow
docker-compose build --no-cache

# Iniciar serviços básicos
Write-Host "🐳 Iniciando serviços básicos (postgres + api)..." -ForegroundColor Yellow
docker-compose up -d postgres

# Aguardar PostgreSQL ficar pronto
Write-Host "⏳ Aguardando PostgreSQL..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Iniciar API
Write-Host "🚀 Iniciando API..." -ForegroundColor Yellow
docker-compose up -d api

# Verificar status
Write-Host ""
Write-Host "✅ Serviços iniciados!" -ForegroundColor Green
Write-Host ""
docker-compose ps

Write-Host ""
Write-Host "📊 Logs da API:" -ForegroundColor Cyan
docker-compose logs --tail=20 api

Write-Host ""
Write-Host "🌐 Acesse a API em: http://localhost:8080" -ForegroundColor Green
Write-Host "📚 Documentação: http://localhost:8080/api" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Comandos úteis:" -ForegroundColor Yellow
Write-Host "  - Ver logs: docker-compose logs -f api"
Write-Host "  - Parar tudo: docker-compose down"
Write-Host "  - Reiniciar API: docker-compose restart api"
Write-Host ""

# Iniciar serviços opcionais
$startPgAdmin = Read-Host "Deseja iniciar PGAdmin? (s/N)"
if ($startPgAdmin -eq 's' -or $startPgAdmin -eq 'S') {
    docker-compose --profile tools up -d pgadmin
    Write-Host "🗄️  PGAdmin disponível em: http://localhost:5050" -ForegroundColor Green
}

Write-Host ""
Write-Host "✨ Tudo pronto! Pressione qualquer tecla para sair..." -ForegroundColor Green
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")