FROM golang:1.20-alpine AS builder

# Instalar dependências
RUN apk update && apk add --no-cache git

# Configurar diretório de trabalho
WORKDIR /app

# Copiar arquivos go.mod e go.sum
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o destack-api ./cmd/server

# Imagem final
FROM alpine:3.17

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar o binário compilado
COPY --from=builder /app/destack-api .
COPY app.env .

# Expor porta
EXPOSE 8080

# Comando para iniciar a aplicação
CMD ["./destack-api"]