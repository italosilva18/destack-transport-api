-- Script de inicialização do banco de dados Destack Transport API
-- Execute este script como superusuário PostgreSQL

-- Criar banco de dados se não existir
SELECT 'CREATE DATABASE destack_transport'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'destack_transport')\gexec

-- Conectar ao banco
\c destack_transport;

-- Criar extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- Para buscas com LIKE otimizadas

-- Criar schema se necessário
CREATE SCHEMA IF NOT EXISTS destack;

-- Configurar search_path
ALTER DATABASE destack_transport SET search_path TO public, destack;

-- Criar índices para otimização (serão criados após as migrações do GORM)
-- Os índices abaixo são exemplos que podem ser adicionados após a criação das tabelas

/*
-- Índices para a tabela users
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);

-- Índices para a tabela empresas
CREATE INDEX IF NOT EXISTS idx_empresas_cnpj ON empresas(cnpj);
CREATE INDEX IF NOT EXISTS idx_empresas_cpf ON empresas(cpf);
CREATE INDEX IF NOT EXISTS idx_empresas_uf ON empresas(uf);
CREATE INDEX IF NOT EXISTS idx_empresas_razao_social ON empresas USING gin(razao_social gin_trgm_ops);

-- Índices para a tabela ctes
CREATE INDEX IF NOT EXISTS idx_ctes_chave ON ctes(chave);
CREATE INDEX IF NOT EXISTS idx_ctes_data_emissao ON ctes(data_emissao);
CREATE INDEX IF NOT EXISTS idx_ctes_status ON ctes(status);
CREATE INDEX IF NOT EXISTS idx_ctes_modalidade_frete ON ctes(modalidade_frete);
CREATE INDEX IF NOT EXISTS idx_ctes_emitente_id ON ctes(emitente_id);
CREATE INDEX IF NOT EXISTS idx_ctes_destinatario_id ON ctes(destinatario_id);
CREATE INDEX IF NOT EXISTS idx_ctes_uf_inicio_destino ON ctes(uf_inicio, uf_destino);

-- Índices para a tabela mdfes
CREATE INDEX IF NOT EXISTS idx_mdfes_chave ON mdfes(chave);
CREATE INDEX IF NOT EXISTS idx_mdfes_data_emissao ON mdfes(data_emissao);
CREATE INDEX IF NOT EXISTS idx_mdfes_status ON mdfes(status);
CREATE INDEX IF NOT EXISTS idx_mdfes_encerrado ON mdfes(encerrado);
CREATE INDEX IF NOT EXISTS idx_mdfes_veiculo_tracao_id ON mdfes(veiculo_tracao_id);

-- Índices para a tabela veiculos
CREATE INDEX IF NOT EXISTS idx_veiculos_placa ON veiculos(placa);
CREATE INDEX IF NOT EXISTS idx_veiculos_tipo ON veiculos(tipo);

-- Índices para a tabela manutencoes
CREATE INDEX IF NOT EXISTS idx_manutencoes_veiculo_id ON manutencoes(veiculo_id);
CREATE INDEX IF NOT EXISTS idx_manutencoes_data_servico ON manutencoes(data_servico);
CREATE INDEX IF NOT EXISTS idx_manutencoes_status ON manutencoes(status);

-- Índices para a tabela uploads
CREATE INDEX IF NOT EXISTS idx_uploads_status ON uploads(status);
CREATE INDEX IF NOT EXISTS idx_uploads_data_upload ON uploads(data_upload);
CREATE INDEX IF NOT EXISTS idx_uploads_chave_doc_processado ON uploads(chave_doc_processado);
*/

-- Criar role para a aplicação (opcional)
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'destack_app') THEN

      CREATE ROLE destack_app LOGIN PASSWORD 'destack_password';
   END IF;
END
$do$;

-- Conceder permissões
GRANT ALL PRIVILEGES ON DATABASE destack_transport TO destack_app;
GRANT ALL ON SCHEMA public TO destack_app;
GRANT ALL ON SCHEMA destack TO destack_app;

-- Configurações de performance (ajuste conforme necessário)
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET min_wal_size = '1GB';
ALTER SYSTEM SET max_wal_size = '4GB';

-- Recarregar configurações
SELECT pg_reload_conf();

-- Mensagem de conclusão
DO $$
BEGIN
    RAISE NOTICE 'Banco de dados destack_transport criado e configurado com sucesso!';
    RAISE NOTICE 'Execute a aplicação para criar as tabelas automaticamente.';
END $$;