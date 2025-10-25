-- Migration to add prompts table for system prompt management
-- This adds prompts as a full entity with reference IDs and active prompt management

-- Create sequence for prompt reference IDs
CREATE SEQUENCE prompt_ref_seq START 1;

-- Function to get next prompt reference ID
CREATE OR REPLACE FUNCTION get_next_prompt_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'PROMPT-' || LPAD(nextval('prompt_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Create prompts table
CREATE TABLE prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(50) UNIQUE NOT NULL DEFAULT get_next_prompt_ref_id(),
    name VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for prompts
CREATE INDEX idx_prompts_reference_id ON prompts(reference_id);
CREATE INDEX idx_prompts_name ON prompts(name);
CREATE INDEX idx_prompts_is_active ON prompts(is_active);
CREATE INDEX idx_prompts_creator_id ON prompts(creator_id);
CREATE INDEX idx_prompts_created_at ON prompts(created_at);
CREATE INDEX idx_prompts_updated_at ON prompts(updated_at);

-- Ensure only one prompt can be active at a time
CREATE UNIQUE INDEX idx_prompts_single_active ON prompts(is_active) WHERE is_active = true;

-- Full-text search indexes for prompts
CREATE INDEX idx_prompts_title ON prompts USING gin(to_tsvector('english', title));
CREATE INDEX idx_prompts_description ON prompts USING gin(to_tsvector('english', description));
CREATE INDEX idx_prompts_content ON prompts USING gin(to_tsvector('english', content));
CREATE INDEX idx_prompts_search ON prompts USING gin(to_tsvector('english', reference_id || ' ' || name || ' ' || title || ' ' || COALESCE(description, '') || ' ' || content));

-- Add updated_at trigger for prompts table
CREATE TRIGGER update_prompts_updated_at 
    BEFORE UPDATE ON prompts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add UUID index for prompts (following existing pattern)
CREATE INDEX idx_prompts_uuid ON prompts(id);

-- Insert default system prompt
INSERT INTO prompts (name, title, description, content, is_active, creator_id) 
SELECT 
    'requirements-analyst',
    'Requirements Analyst Assistant',
    'AI assistant specialized in requirements analysis and management',
    'You are an expert requirements analyst working with a Product Requirements Management System. Your role is to help users create, analyze, and manage requirements through a hierarchical structure: Epics (high-level features), User Stories (specific user needs), Acceptance Criteria (testable conditions), and Requirements (detailed specifications). You have access to tools for CRUD operations and can analyze requirement quality, suggest improvements, and identify dependencies. Always focus on clarity, testability, and traceability.',
    true,
    (SELECT id FROM users WHERE role = 'Administrator' LIMIT 1)
WHERE EXISTS (SELECT 1 FROM users WHERE role = 'Administrator');