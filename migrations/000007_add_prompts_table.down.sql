-- Remove prompts table and related objects
DROP TABLE IF EXISTS prompts;
DROP FUNCTION IF EXISTS get_next_prompt_ref_id();
DROP SEQUENCE IF EXISTS prompt_ref_seq;