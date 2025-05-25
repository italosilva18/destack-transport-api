# Build stage
FROM golang:1.23-alpine AS builder

# Set Timezone environment variable (good practice for consistency)
ENV TZ=America/Sao_Paulo

# Instalar dependências de build
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Criar usuário e grupo não-root no builder (para consistência de UID/GID se necessário, mas não estritamente para o binário estático)
RUN addgroup -S appgroup -g 1000 && \
    adduser -S -u 1000 -G appgroup -h /app appuser

# Configurar diretório de trabalho
WORKDIR /build

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download && go mod verify # Adicionado go mod verify

# Copiar código fonte
# Utilizar .dockerignore para evitar copiar arquivos desnecessários como .git, logs/, tmp/, etc.
COPY . .

# Compilar aplicação
# O binário já é estático, -a e -installsuffix cgo são redundantes com CGO_ENABLED=0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -o destack-api ./cmd/server

# Final stage
FROM alpine:3.18

# Set Timezone environment variable
ENV TZ=America/Sao_Paulo

# Instalar pacotes necessários para o runtime e healthcheck
RUN apk --no-cache add \
    ca-certificates \
    bash \
    curl \
    tzdata

# Criar grupo e usuário não-root no estágio final com IDs específicos
# É crucial que este usuário e grupo existam ANTES de tentar usar chown com eles.
RUN addgroup -S appgroup -g 1000 && \
    adduser -S -u 1000 -G appgroup -h /app appuser

# Criar diretório da aplicação
WORKDIR /app

# Copiar binário do builder stage
COPY --from=builder /build/destack-api /app/destack-api

# Criar diretórios de logs e uploads DEPOIS de criar /app e ANTES de mudar a propriedade
RUN mkdir -p /app/logs /app/uploads

# Dar permissões ao appuser para o diretório /app e seu conteúdo
# Isso inclui o binário, logs e uploads.
RUN chown -R appuser:appgroup /app

# Mudar para o usuário não-root
USER appuser

# Expor porta (a porta exposta aqui é a interna do container)
EXPOSE 8080

# Health check (verifique se a API realmente roda em / e não /api/health no healthcheck)
# O start-period é importante para dar tempo à API de iniciar antes que o healthcheck comece a falhar.
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Comando para iniciar
ENTRYPOINT ["/app/destack-api"]