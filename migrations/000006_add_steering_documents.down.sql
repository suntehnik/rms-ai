-- Rollback migration for steering documents functionality
-- This removes all steering documents related tables, functions, and sequences

-- Drop trigger
DROP TRIGGER IF EXISTS update_steering_documents_updated_at ON steering_documents;

-- Drop junction table (this will cascade delete all relationships)
DROP TABLE IF EXISTS epic_steering_documents;

-- Drop main steering_documents table
DROP TABLE IF EXISTS steering_documents;

-- Drop the reference ID function
DROP FUNCTION IF EXISTS get_next_steering_document_ref_id();

-- Drop the sequence
DROP SEQUENCE IF EXISTS steering_document_ref_seq;