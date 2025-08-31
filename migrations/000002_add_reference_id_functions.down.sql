-- Rollback migration for reference ID functions

-- Drop composite indexes
DROP INDEX IF EXISTS idx_req_rel_target_type;
DROP INDEX IF EXISTS idx_req_rel_source_type;
DROP INDEX IF EXISTS idx_comments_author_created;
DROP INDEX IF EXISTS idx_comments_entity_resolved;
DROP INDEX IF EXISTS idx_requirements_type_status;
DROP INDEX IF EXISTS idx_requirements_user_story_status;
DROP INDEX IF EXISTS idx_acceptance_criteria_user_story_created;
DROP INDEX IF EXISTS idx_user_stories_epic_status;

-- Drop UUID indexes
DROP INDEX IF EXISTS idx_comments_uuid;
DROP INDEX IF EXISTS idx_requirement_relationships_uuid;
DROP INDEX IF EXISTS idx_relationship_types_uuid;
DROP INDEX IF EXISTS idx_requirement_types_uuid;
DROP INDEX IF EXISTS idx_users_uuid;
DROP INDEX IF EXISTS idx_requirements_uuid;
DROP INDEX IF EXISTS idx_acceptance_criteria_uuid;
DROP INDEX IF EXISTS idx_user_stories_uuid;
DROP INDEX IF EXISTS idx_epics_uuid;

-- Restore original default values
ALTER TABLE requirements ALTER COLUMN reference_id SET DEFAULT ('REQ-' || LPAD(nextval('requirement_ref_seq')::TEXT, 3, '0'));
ALTER TABLE acceptance_criteria ALTER COLUMN reference_id SET DEFAULT ('AC-' || LPAD(nextval('acceptance_criteria_ref_seq')::TEXT, 3, '0'));
ALTER TABLE user_stories ALTER COLUMN reference_id SET DEFAULT ('US-' || LPAD(nextval('user_story_ref_seq')::TEXT, 3, '0'));
ALTER TABLE epics ALTER COLUMN reference_id SET DEFAULT ('EP-' || LPAD(nextval('epic_ref_seq')::TEXT, 3, '0'));

-- Drop helper functions
DROP FUNCTION IF EXISTS get_next_requirement_ref_id();
DROP FUNCTION IF EXISTS get_next_acceptance_criteria_ref_id();
DROP FUNCTION IF EXISTS get_next_user_story_ref_id();
DROP FUNCTION IF EXISTS get_next_epic_ref_id();