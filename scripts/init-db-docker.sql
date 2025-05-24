-- Script de inicialização do banco de dados para Docker
-- Este script é executado automaticamente quando o container PostgreSQL é criado

-- Criar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- Criar schema
CREATE SCHEMA IF NOT EXISTS destack;

-- Configurar search_path
ALTER DATABASE destack_transport SET search_path TO public, destack;

-- Criar função para geração de slugs
CREATE OR REPLACE FUNCTION generate_slug(input_text TEXT)
RETURNS TEXT AS $$
BEGIN
    RETURN LOWER(
        REGEXP_REPLACE(
            REGEXP_REPLACE(
                UNACCENT(input_text),
                '[^a-zA-Z0-9\s-]', '', 'g'
            ),
            '\s+', '-', 'g'
        )
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Criar função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Configurações de performance para Docker
ALTER SYSTEM SET shared_buffers = '128MB';
ALTER SYSTEM SET effective_cache_size = '512MB';
ALTER SYSTEM SET maintenance_work_mem = '32MB';
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET max_connections = 100;
ALTER SYSTEM SET random_page_cost = 1.1;

-- Aplicar configurações
SELECT pg_reload_conf();

-- Criar usuário padrão se não existir (apenas para desenvolvimento)
DO $$
BEGIN
    -- Este bloco será executado apenas se a tabela users existir
    -- A aplicação Go criará as tabelas através das migrações
    RAISE NOTICE 'Banco de dados inicializado com sucesso!';
    RAISE NOTICE 'As tabelas serão criadas automaticamente pela aplicação.';
END $$;