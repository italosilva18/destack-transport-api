#!/bin/bash

# Script para corrigir os imports nos handlers

echo "🔧 Corrigindo imports nos handlers..."

# Lista de arquivos para corrigir
handlers=(
    "internal/api/handlers/cte/cte_handler.go"
    "internal/api/handlers/mdfe/mdfe_handler.go"
    "internal/api/handlers/financeiro/financeiro_handler.go"
    "internal/api/handlers/geografico/geografico_handler.go"
    "internal/api/handlers/manutencao/manutencao_handler.go"
)

for file in "${handlers[@]}"; do
    if [ -f "$file" ]; then
        echo "📝 Corrigindo: $file"
        
        # Adicionar import do zerolog se não existir
        if ! grep -q '"github.com/rs/zerolog"' "$file"; then
            # Adicionar o import após o último import
            sed -i '/^import (/,/^)/ {
                /^)/ i\
\t"github.com/rs/zerolog"
            }' "$file" 2>/dev/null || \
            sed -i '' '/^import (/,/^)/ {
                /^)/ i\
\	"github.com/rs/zerolog"
            }' "$file"
        fi
        
        # Corrigir o tipo logger
        sed -i 's/logger logger\.Logger/logger zerolog.Logger/g' "$file" 2>/dev/null || \
        sed -i '' 's/logger logger\.Logger/logger zerolog.Logger/g' "$file"
        
        echo "✅ Corrigido: $file"
    else
        echo "⚠️  Arquivo não encontrado: $file"
    fi
done

echo ""
echo "✅ Correções aplicadas!"
echo ""
echo "🔍 Verificando compilação..."
go build ./...

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida!"
else
    echo "❌ Ainda há erros de compilação. Verifique manualmente."
fi