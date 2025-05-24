# Script PowerShell para criar estrutura de diretórios do Docker

Write-Host "📁 Criando estrutura de diretórios para Docker..." -ForegroundColor Green

# Criar diretórios
$directories = @(
    "docker\nginx\conf.d",
    "docker\nginx\ssl",
    "docker\prometheus",
    "docker\grafana\provisioning\dashboards",
    "docker\grafana\provisioning\datasources",
    "docker\scripts",
    "logs",
    "uploads"
)

foreach ($dir in $directories) {
    if (!(Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        Write-Host "✅ Criado: $dir" -ForegroundColor Green
    } else {
        Write-Host "ℹ️  Já existe: $dir" -ForegroundColor Yellow
    }
}

# Criar arquivo de configuração do Prometheus
$prometheusConfig = @"
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'destack-api'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
"@

$prometheusConfig | Out-File -FilePath "docker\prometheus\prometheus.yml" -Encoding UTF8

# Criar datasource do Grafana
$grafanaDataSource = @"
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
"@

$grafanaDataSource | Out-File -FilePath "docker\grafana\provisioning\datasources\prometheus.yml" -Encoding UTF8

# Criar arquivo nginx.conf principal
$nginxConfig = @"
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens off;

    gzip on;
    gzip_disable "msie6";
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/rss+xml application/atom+xml image/svg+xml;

    include /etc/nginx/conf.d/*.conf;
}
"@

$nginxConfig | Out-File -FilePath "docker\nginx\nginx.conf" -Encoding UTF8

Write-Host ""
Write-Host "✅ Estrutura de diretórios criada com sucesso!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Próximos passos:" -ForegroundColor Yellow
Write-Host "1. Copie .env.docker para .env" -ForegroundColor White
Write-Host "2. Execute: docker-compose up -d" -ForegroundColor White
Write-Host "3. Acesse: http://localhost:8080" -ForegroundColor White