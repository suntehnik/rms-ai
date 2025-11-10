-- Create refresh_tokens table for JWT refresh token management
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Foreign key constraint with cascade delete
    CONSTRAINT fk_refresh_tokens_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

-- Create index on user_id for efficient user session queries
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id 
    ON refresh_tokens(user_id);

-- Create index on expires_at for efficient cleanup of expired tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at 
    ON refresh_tokens(expires_at);

-- Create index on token_hash for efficient token lookup
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash 
    ON refresh_tokens(token_hash);
