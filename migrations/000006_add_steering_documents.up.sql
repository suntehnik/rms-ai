-- Migration to add steering documents functionality
-- This adds steering documents as a full entity with many-to-many relationship to epics

-- Create sequence for steering document reference IDs
CREATE SEQUENCE steering_document_ref_seq START 1;

-- Function to get next steering document reference ID
CREATE OR REPLACE FUNCTION get_next_steering_document_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'STD-' || LPAD(nextval('steering_document_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Create steering_documents table
CREATE TABLE steering_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(50) UNIQUE NOT NULL DEFAULT get_next_steering_document_ref_id(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for steering_documents
CREATE INDEX idx_steering_documents_creator_id ON steering_documents(creator_id);
CREATE INDEX idx_steering_documents_reference_id ON steering_documents(reference_id);
CREATE INDEX idx_steering_documents_created_at ON steering_documents(created_at);
CREATE INDEX idx_steering_documents_updated_at ON steering_documents(updated_at);

-- Full-text search indexes for steering documents
CREATE INDEX idx_steering_documents_title ON steering_documents USING gin(to_tsvector('english', title));
CREATE INDEX idx_steering_documents_description ON steering_documents USING gin(to_tsvector('english', description));
CREATE INDEX idx_steering_documents_search ON steering_documents USING gin(to_tsvector('english', reference_id || ' ' || title || ' ' || COALESCE(description, '')));

-- Create epic_steering_documents junction table for many-to-many relationship
CREATE TABLE epic_steering_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    steering_document_id UUID NOT NULL REFERENCES steering_documents(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(epic_id, steering_document_id)
);

-- Create indexes for epic_steering_documents junction table
CREATE INDEX idx_epic_steering_documents_epic_id ON epic_steering_documents(epic_id);
CREATE INDEX idx_epic_steering_documents_steering_document_id ON epic_steering_documents(steering_document_id);
CREATE INDEX idx_epic_steering_documents_created_at ON epic_steering_documents(created_at);

-- Add updated_at trigger for steering_documents table
CREATE TRIGGER update_steering_documents_updated_at 
    BEFORE UPDATE ON steering_documents 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add UUID index for steering_documents (following existing pattern)
CREATE INDEX idx_steering_documents_uuid ON steering_documents(id);