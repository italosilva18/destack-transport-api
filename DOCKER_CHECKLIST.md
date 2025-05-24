# ✅ Checklist - Docker Setup Completo

## 📋 Arquivos Criados/Atualizados

### 1. **Docker Core**
- [x] `Dockerfile` - Otimizado com multi-stage build
- [x] `docker-compose.yml` - Configuração completa de produção
- [x] `docker-compose.dev.yml` - Configuração para desenvolvimento
- [x] `.dockerignore` - Otimização do build

### 2. **Configurações**
- [x] `.env.docker` - Variáveis de ambiente para Docker
- [x] `docker/nginx/conf.d/default.conf` - Configuração Nginx
- [x] `docker/nginx/nginx.conf` - Configuração principal Nginx
- [x] `docker/prometheus/prometheus.yml` - Configuração Prometheus
- [x] `docker/grafana/provisioning/` - Configurações Grafana

### 3. **Scripts**
- [x] `docker/scripts/docker-up.sh` - Script Linux/Mac
- [x] `docker/scripts/docker-up.ps1` - Script Windows PowerShell
- [x] `setup-docker-dirs.sh` - Criar estrutura de pastas (Linux)
- [x] `setup-docker-dirs.ps1` - Criar estrutura de pastas (Windows)

### 4. **Banco de Dados**
- [x] `scripts/init-db-docker.sql` - Script otimizado para Docker
- [x] Health check e retry no `main.go`

### 5. **Documentação**
- [x] `docker/README.md` - Guia completo do Docker
- [x] `Makefile` atualizado com comandos Docker

## 🚀 Como Usar

### Windows

1. **Criar estrutura de diretórios:**
   ```powershell
   .\setup-docker-dirs.ps1
   ```

2. **Copiar arquivo de ambiente:**
   ```powershell
   Copy-Item .env.docker .env
   ```

3. **Iniciar com script:**
   ```powershell
   .\docker\scripts\docker-up.ps1
   ```

   **OU manualmente:**
   ```powershell
   docker-compose up -d
   ```

### Linux/Mac

1. **Criar estrutura de diretórios:**
   ```bash
   chmod +x setup-docker-dirs.sh
   ./setup-docker-dirs.sh
   ```

2. **Copiar arquivo de ambiente:**
   ```bash
   cp .env.docker .env
   ```

3. **Iniciar com script:**
   ```bash
   chmod +x docker/scripts/docker-up.sh
   ./docker/scripts/docker-up.sh
   ```

   **OU manualmente:**
   ```bash
   docker-compose up -d
   ```

## 🔍 Verificação

### 1. **Verificar serviços rodando:**
```bash
docker-compose ps
```

Deve mostrar:
- `destack-postgres` - Running
- `destack-api` - Running

### 2. **Verificar logs:**
```bash
docker-compose logs -f api
```

### 3. **Testar endpoints:**
```bash
# Health check
curl http://localhost:8080/health

# Login (Windows PowerShell)
Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body '{"username":"admin","password":"admin123"}' -ContentType "application/json"

# Login (Linux/Mac)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 📊 Serviços Disponíveis

| Serviço | URL | Credenciais | Profile |
|---------|-----|-------------|---------|
| API | http://localhost:8080 | admin/admin123 | default |
| PostgreSQL | localhost:5432 | postgres/postgres | default |
| PGAdmin | http://localhost:5050 | admin@destack.com/admin | tools |
| Redis | localhost:6379 | redis_password | cache |
| Prometheus | http://localhost:9090 | - | monitoring |
| Grafana | http://localhost:3000 | admin/admin | monitoring |
| Mailhog | http://localhost:8025 | - | dev only |

## 🛑 Troubleshooting

### Erro: "Cannot connect to database"
```bash
# Verificar se PostgreSQL está rodando
docker-compose logs postgres

# Reiniciar PostgreSQL
docker-compose restart postgres
```

### Erro: "Port already in use"
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :8080
kill -9 <PID>
```

### Limpar tudo e recomeçar
```bash
docker-compose down -v
docker system prune -f
docker-compose up -d --build
```

## 📈 Performance

### Recursos recomendados:
- **CPU**: 2+ cores
- **RAM**: 4GB+ (2GB mínimo)
- **Disco**: 10GB+ livres

### Docker Desktop (Windows/Mac):
- Alocar pelo menos 4GB RAM
- Alocar 2+ CPUs

## 🔒 Segurança

⚠️ **IMPORTANTE para produção:**

1. **Altere TODAS as senhas no `.env`**
2. **Use certificados SSL válidos**
3. **Configure firewall adequadamente**
4. **Não exponha serviços desnecessários**

## ✨ Comandos Úteis

```bash
# Desenvolvimento
make docker-dev         # Ambiente de desenvolvimento
make docker-shell       # Shell no container
make docker-logs        # Ver logs

# Produção
make docker-up          # Iniciar tudo
make docker-down        # Parar tudo
make docker-restart     # Reiniciar

# Manutenção
make docker-clean       # Limpar volumes
docker-compose exec postgres pg_dump -U postgres destack_transport > backup.sql
```

## 🎉 Pronto!

Se tudo correu bem, você deve ter:
- ✅ API rodando em http://localhost:8080
- ✅ Banco de dados PostgreSQL configurado
- ✅ Usuário admin criado (admin/admin123)
- ✅ Logs sendo salvos em ./logs/
- ✅ Sistema pronto para desenvolvimento ou produção

**Próximos passos:**
1. Testar os endpoints da API
2. Configurar seu frontend
3. Personalizar as configurações
4. Implementar features adicionais