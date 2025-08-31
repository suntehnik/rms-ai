-- Remove status model system tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_status_transitions_updated_at ON status_transitions;
DROP TRIGGER IF EXISTS update_statuses_updated_at ON statuses;
DROP TRIGGER IF EXISTS update_status_models_updated_at ON status_models;

-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS status_transitions;
DROP TABLE IF EXISTS statuses;
DROP TABLE IF EXISTS status_models;