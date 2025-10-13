-- Database initialization script for production
-- This script runs when the PostgreSQL container starts for the first time

-- Create database if it doesn't exist (handled by POSTGRES_DB environment variable)
-- CREATE DATABASE IF NOT EXISTS requirements_db;

-- Set up database configuration for optimal performance
ALTER DATABASE requirements_db SET timezone TO 'UTC';
ALTER DATABASE requirements_db SET log_statement TO 'none';
ALTER DATABASE requirements_db SET log_min_duration_statement TO 1000;

-- Create extensions if they don't exist
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Set up connection limits and performance settings
ALTER DATABASE requirements_db SET max_connections TO 100;
ALTER DATABASE requirements_db SET shared_buffers TO '256MB';
ALTER DATABASE requirements_db SET effective_cache_size TO '1GB';
ALTER DATABASE requirements_db SET maintenance_work_mem TO '64MB';
ALTER DATABASE requirements_db SET checkpoint_completion_target TO 0.9;
ALTER DATABASE requirements_db SET wal_buffers TO '16MB';
ALTER DATABASE requirements_db SET default_statistics_target TO 100;

-- Create application user (optional, for better security)
-- DO $$ 
-- BEGIN
--     IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user') THEN
--         CREATE ROLE app_user WITH LOGIN PASSWORD 'your_app_password_here';
--         GRANT CONNECT ON DATABASE requirements_db TO app_user;
--         GRANT USAGE ON SCHEMA public TO app_user;
--         GRANT CREATE ON SCHEMA public TO app_user;
--     END IF;
-- END
-- $$;