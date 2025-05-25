# Build stage
FROM golang:1.23-alpine AS builder

# Instalar dependências de build
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Criar usuário não-root
RUN adduser -D -g '' appuser

# Configurar diretório de trabalho
WORKDIR /build

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar aplicação
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o destack-api ./cmd/server

# Final stage
FROM alpine:3.18

# Instalar ca-certificates e bash
RUN apk --no-cache add ca-certificates bash curl

# Copiar certificados SSL e timezone
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar usuário não-root
COPY --from=builder /etc/passwd /etc/passwd

# Criar diretório da aplicação e dar permissões
RUN mkdir -p /app/logs /app/uploads && chown -R appuser:appuser /app

# Mudar para o diretório da aplicação
WORKDIR /app

# Copiar binário
COPY --from=builder /build/destack-api /app/destack-api

# Dar permissão ao usuário appuser
RUN chown appuser:appuser /app/destack-api

# Usar usuário não-root
USER appuser

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Comando para iniciar
ENTRYPOINT ["/app/destack-api"]