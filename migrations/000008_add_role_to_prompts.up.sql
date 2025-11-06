-- Migration to add role field to prompts table for MCP compliance
-- This adds the role column with check constraint to ensure only valid MCP roles are used

-- Add role column with default value 'assistant'
ALTER TABLE prompts 
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'assistant';

-- Add check constraint for valid MCP roles
ALTER TABLE prompts 
ADD CONSTRAINT check_prompt_role 
CHECK (role IN ('user', 'assistant'));

-- Create index for role column for potential future filtering
CREATE INDEX idx_prompts_role ON prompts(role);

-- Update existing prompts to have explicit role (they will default to 'assistant')
-- This is already handled by the DEFAULT value, but we can verify all records have a role
UPDATE prompts SET role = 'assistant' WHERE role IS NULL OR role = '';