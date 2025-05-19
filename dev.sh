#!/bin/bash

# Build e executar em desenvolvimento
echo "Construindo e iniciando em modo de desenvolvimento..."
go mod tidy
go build -o tmp/destack-api ./cmd/server
./tmp/destack-api