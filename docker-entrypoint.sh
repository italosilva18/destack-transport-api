#!/bin/sh

# Script de entrada para criar app.env a partir das variáveis de ambiente

echo "Creating app.env from environment variables..."

cat > /app/app.env << EOF
# Generated from environment variables
ENVIRONMENT=${ENVIRONMENT:-production}
SERVER_PORT=${SERVER_PORT:-8080}

# Database
DB_HOST=${DB_HOST:-postgres}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-destack_transport}
DB_SSLMODE=${DB_SSLMODE:-disable}

# JWT
JWT_SECRET=${JWT_SECRET:-default_jwt_secret_change_this}
JWT_EXPIRES_IN=${JWT_EXPIRES_IN:-24}
EOF

echo "app.env created with content:"
cat /app/app.env

# Execute the application
exec /app/destack-api