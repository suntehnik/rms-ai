-- Add Personal Access Tokens table

CREATE TABLE personal_access_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    prefix VARCHAR(20) NOT NULL DEFAULT 'mcp_pat_',
    scopes JSONB DEFAULT '["full_access"]',
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT unique_user_token_name UNIQUE(user_id, name)
);

-- Create indexes for personal_access_tokens
CREATE INDEX idx_pat_user_id ON personal_access_tokens(user_id);
CREATE INDEX idx_pat_prefix ON personal_access_tokens(prefix);
CREATE INDEX idx_pat_expires_at ON personal_access_tokens(expires_at);
CREATE INDEX idx_pat_last_used_at ON personal_access_tokens(last_used_at);
CREATE INDEX idx_pat_created_at ON personal_access_tokens(created_at);

-- Add updated_at trigger for personal_access_tokens
CREATE TRIGGER update_personal_access_tokens_updated_at 
    BEFORE UPDATE ON personal_access_tokens 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment to table for documentation
COMMENT ON TABLE personal_access_tokens IS 'Personal Access Tokens for API authentication';
COMMENT ON COLUMN personal_access_tokens.user_id IS 'Foreign key to users table with CASCADE delete';
COMMENT ON COLUMN personal_access_tokens.name IS 'User-defined name for the token';
COMMENT ON COLUMN personal_access_tokens.token_hash IS 'bcrypt hash of the token secret (never store plaintext)';
COMMENT ON COLUMN personal_access_tokens.prefix IS 'Token prefix for identification (default: mcp_pat_)';
COMMENT ON COLUMN personal_access_tokens.scopes IS 'JSONB array of permission scopes (default: ["full_access"])';
COMMENT ON COLUMN personal_access_tokens.expires_at IS 'Optional expiration timestamp';
COMMENT ON COLUMN personal_access_tokens.last_used_at IS 'Timestamp of last successful authentication';