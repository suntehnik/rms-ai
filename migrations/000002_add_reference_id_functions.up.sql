-- Migration to add helper functions for reference ID generation
-- This ensures GORM models work correctly with the dual ID system

-- Function to get next epic reference ID
CREATE OR REPLACE FUNCTION get_next_epic_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'EP-' || LPAD(nextval('epic_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Function to get next user story reference ID
CREATE OR REPLACE FUNCTION get_next_user_story_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'US-' || LPAD(nextval('user_story_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Function to get next acceptance criteria reference ID
CREATE OR REPLACE FUNCTION get_next_acceptance_criteria_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'AC-' || LPAD(nextval('acceptance_criteria_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Function to get next requirement reference ID
CREATE OR REPLACE FUNCTION get_next_requirement_ref_id() RETURNS VARCHAR(20) AS $$
BEGIN
    RETURN 'REQ-' || LPAD(nextval('requirement_ref_seq')::TEXT, 3, '0');
END;
$$ LANGUAGE plpgsql;

-- Update default values to use functions (for better GORM compatibility)
ALTER TABLE epics ALTER COLUMN reference_id SET DEFAULT get_next_epic_ref_id();
ALTER TABLE user_stories ALTER COLUMN reference_id SET DEFAULT get_next_user_story_ref_id();
ALTER TABLE acceptance_criteria ALTER COLUMN reference_id SET DEFAULT get_next_acceptance_criteria_ref_id();
ALTER TABLE requirements ALTER COLUMN reference_id SET DEFAULT get_next_requirement_ref_id();

-- Add indexes for UUID lookups (in addition to reference ID indexes)
CREATE INDEX IF NOT EXISTS idx_epics_uuid ON epics(id);
CREATE INDEX IF NOT EXISTS idx_user_stories_uuid ON user_stories(id);
CREATE INDEX IF NOT EXISTS idx_acceptance_criteria_uuid ON acceptance_criteria(id);
CREATE INDEX IF NOT EXISTS idx_requirements_uuid ON requirements(id);
CREATE INDEX IF NOT EXISTS idx_users_uuid ON users(id);
CREATE INDEX IF NOT EXISTS idx_requirement_types_uuid ON requirement_types(id);
CREATE INDEX IF NOT EXISTS idx_relationship_types_uuid ON relationship_types(id);
CREATE INDEX IF NOT EXISTS idx_requirement_relationships_uuid ON requirement_relationships(id);
CREATE INDEX IF NOT EXISTS idx_comments_uuid ON comments(id);

-- Add composite indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_stories_epic_status ON user_stories(epic_id, status);
CREATE INDEX IF NOT EXISTS idx_acceptance_criteria_user_story_created ON acceptance_criteria(user_story_id, created_at);
CREATE INDEX IF NOT EXISTS idx_requirements_user_story_status ON requirements(user_story_id, status);
CREATE INDEX IF NOT EXISTS idx_requirements_type_status ON requirements(type_id, status);
CREATE INDEX IF NOT EXISTS idx_comments_entity_resolved ON comments(entity_type, entity_id, is_resolved);
CREATE INDEX IF NOT EXISTS idx_comments_author_created ON comments(author_id, created_at);

-- Add index for requirement relationships lookup
CREATE INDEX IF NOT EXISTS idx_req_rel_source_type ON requirement_relationships(source_requirement_id, relationship_type_id);
CREATE INDEX IF NOT EXISTS idx_req_rel_target_type ON requirement_relationships(target_requirement_id, relationship_type_id);