# Destack Transport API - Docker

Este guia explica como executar a aplicação usando Docker.

## 🚀 Quick Start

### Windows (PowerShell)
```powershell
# Clone o repositório
git clone https://github.com/italosilva18/destack-transport-api.git
cd destack-transport-api

# Execute o script de inicialização
.\docker\scripts\docker-up.ps1
```

### Linux/Mac/WSL
```bash
# Clone o repositório
git clone https://github.com/italosilva18/destack-transport-api.git
cd destack-transport-api

# Torne o script executável
chmod +x docker/scripts/docker-up.sh

# Execute o script de inicialização
./docker/scripts/docker-up.sh
```

### Manualmente
```bash
# Copie o arquivo de ambiente
cp .env.docker .env

# Inicie os serviços
docker-compose up -d

# Verifique os logs
docker-compose logs -f api
```

## 📦 Serviços Disponíveis

### Serviços Principais (sempre ativos)
- **API**: http://localhost:8080
- **PostgreSQL**: localhost:5432

### Serviços Opcionais
Para ativar serviços opcionais, use profiles:

```bash
# PGAdmin (gerenciador de banco)
docker-compose --profile tools up -d pgadmin
# Acesse em: http://localhost:5050

# Redis (cache)
docker-compose --profile cache up -d redis

# Nginx (proxy reverso)
docker-compose --profile production up -d nginx

# Monitoramento (Prometheus + Grafana)
docker-compose --profile monitoring up -d prometheus grafana
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000
```

## 🛠️ Comandos Úteis

### Usando Make
```bash
make docker-up         # Iniciar todos os serviços
make docker-down       # Parar todos os serviços
make docker-logs       # Ver logs
make docker-restart    # Reiniciar serviços
make docker-shell      # Acessar shell do container API
make docker-db-shell   # Acessar PostgreSQL
make docker-clean      # Limpar volumes e containers
```

### Usando Docker Compose
```bash
# Logs
docker-compose logs -f api           # Logs da API
docker-compose logs -f postgres      # Logs do PostgreSQL

# Gerenciar serviços
docker-compose restart api           # Reiniciar API
docker-compose stop                  # Parar sem remover
docker-compose down                  # Parar e remover
docker-compose down -v               # Parar e remover com volumes

# Executar comandos
docker-compose exec api sh           # Shell no container
docker-compose exec postgres psql -U postgres -d destack_transport
```

## 🔧 Desenvolvimento

### Ambiente de Desenvolvimento
Use o arquivo `docker-compose.dev.yml`:

```bash
# Iniciar ambiente de desenvolvimento
docker-compose -f docker-compose.dev.yml up -d

# Inclui:
# - PostgreSQL (sem limites rígidos)
# - PGAdmin (interface web)
# - Redis (sem senha)
# - Mailhog (teste de emails)
```

### Hot Reload (desenvolvimento local)
Para desenvolvimento com hot reload, execute apenas os serviços auxiliares:

```bash
# Iniciar apenas PostgreSQL e Redis
docker-compose -f docker-compose.dev.yml up -d postgres-dev redis-dev

# Execute a API localmente
go run cmd/server/main.go
```

## 📊 Monitoramento

### Logs Estruturados
Os logs são salvos em `./logs/` e podem ser visualizados:

```bash
# Logs em tempo real
docker-compose logs -f api

# Últimas 100 linhas
docker-compose logs --tail=100 api
```

### Métricas (com Prometheus)
1. Inicie o Prometheus:
   ```bash
   docker-compose --profile monitoring up -d
   ```

2. Acesse:
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## 🐛 Troubleshooting

### Container não inicia
```bash
# Verificar logs
docker-compose logs api

# Reconstruir imagem
docker-compose build --no-cache api
docker-compose up -d
```

### Erro de conexão com banco
```bash
# Verificar se PostgreSQL está rodando
docker-compose ps postgres

# Verificar logs do PostgreSQL
docker-compose logs postgres

# Reiniciar PostgreSQL
docker-compose restart postgres
```

### Limpar tudo e começar do zero
```bash
# Para e remove tudo
docker-compose down -v

# Remove imagens
docker rmi destack-transport-api_api

# Inicia novamente
docker-compose up -d --build
```

### Problemas de permissão (Linux)
```bash
# Adicionar usuário ao grupo docker
sudo usermod -aG docker $USER

# Fazer logout e login novamente
```

## 🔒 Segurança

### Produção
Para produção, sempre:

1. Use senhas fortes no `.env`
2. Configure SSL/TLS no Nginx
3. Limite acesso aos serviços auxiliares
4. Use secrets do Docker Swarm ou Kubernetes

### Exemplo de configuração segura
```env
# .env.production
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
REDIS_PASSWORD=$(openssl rand -base64 32)
```

## 📈 Performance

### Otimizações do Docker
1. Use multi-stage builds (já configurado)
2. Minimize layers
3. Use cache eficientemente

### Limites de recursos
Adicione ao `docker-compose.yml`:

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 256M
```

## 🔄 Backup e Restore

### Backup do banco
```bash
# Backup
docker-compose exec postgres pg_dump -U postgres destack_transport > backup.sql

# Restore
docker-compose exec -T postgres psql -U postgres destack_transport < backup.sql
```

### Backup de volumes
```bash
# Backup completo
docker run --rm -v destack-transport-api_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .

# Restore
docker run --rm -v destack-transport-api_postgres_data:/data -v $(pwd):/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

## 📝 Notas

- Os dados do PostgreSQL são persistidos em volumes Docker
- Logs são salvos no host em `./logs/`
- Uploads são salvos em `./uploads/`
- Configurações sensíveis devem estar no `.env` (não commitar!)

## 🆘 Suporte

Para problemas específicos do Docker:
1. Verifique os logs: `docker-compose logs`
2. Consulte a documentação do Docker
3. Abra uma issue no GitHub