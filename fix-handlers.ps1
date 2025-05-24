# Script PowerShell para corrigir os imports nos handlers

Write-Host "🔧 Corrigindo imports nos handlers..." -ForegroundColor Green

# Lista de arquivos para corrigir
$handlers = @(
    "internal\api\handlers\cte\cte_handler.go",
    "internal\api\handlers\mdfe\mdfe_handler.go",
    "internal\api\handlers\financeiro\financeiro_handler.go",
    "internal\api\handlers\geografico\geografico_handler.go",
    "internal\api\handlers\manutencao\manutencao_handler.go"
)

foreach ($file in $handlers) {
    if (Test-Path $file) {
        Write-Host "📝 Corrigindo: $file" -ForegroundColor Yellow
        
        # Ler o conteúdo do arquivo
        $content = Get-Content $file -Raw
        
        # Verificar se o import do zerolog existe
        if ($content -notmatch '"github.com/rs/zerolog"') {
            # Adicionar o import
            $content = $content -replace '(import \([^)]+)', '$1`n`t"github.com/rs/zerolog"'
        }
        
        # Corrigir o tipo logger
        $content = $content -replace 'logger logger\.Logger', 'logger zerolog.Logger'
        
        # Salvar o arquivo
        Set-Content -Path $file -Value $content -Encoding UTF8
        
        Write-Host "✅ Corrigido: $file" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Arquivo não encontrado: $file" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "✅ Correções aplicadas!" -ForegroundColor Green
Write-Host ""
Write-Host "🔍 Verificando compilação..." -ForegroundColor Yellow

# Verificar compilação
$process = Start-Process -FilePath "go" -ArgumentList "build", "./..." -NoNewWindow -Wait -PassThru

if ($process.ExitCode -eq 0) {
    Write-Host "✅ Compilação bem-sucedida!" -ForegroundColor Green
} else {
    Write-Host "❌ Ainda há erros de compilação. Verifique manualmente." -ForegroundColor Red
}

Write-Host ""
Write-Host "Pressione qualquer tecla para continuar..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")