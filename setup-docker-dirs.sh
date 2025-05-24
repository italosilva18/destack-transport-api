#!/bin/bash

# Script para criar estrutura de diretórios do Docker

echo "📁 Criando estrutura de diretórios para Docker..."

# Criar diretórios
directories=(
    "docker/nginx/conf.d"
    "docker/nginx/ssl"
    "docker/prometheus"
    "docker/grafana/provisioning/dashboards"
    "docker/grafana/provisioning/datasources"
    "docker/scripts"
    "logs"
    "uploads"
)

for dir in "${directories[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        echo "✅ Criado: $dir"
    else
        echo "ℹ️  Já existe: $dir"
    fi
done

# Criar arquivo de configuração do Prometheus
cat > docker/prometheus/prometheus.yml << 'EOF'
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
EOF

# Criar datasource do Grafana
cat > docker/grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

# Criar arquivo nginx.conf principal
cat > docker/nginx/nginx.conf << 'EOF'
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
EOF

echo ""
echo "✅ Estrutura de diretórios criada com sucesso!"
echo ""
echo "📝 Próximos passos:"
echo "1. Copie .env.docker para .env"
echo "2. Execute: docker-compose up -d"
echo "3. Acesse: http://localhost:8080"