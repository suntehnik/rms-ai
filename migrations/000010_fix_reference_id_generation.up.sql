-- Fix reference ID generation to support unlimited growth and ensure atomicity
-- This migration updates the functions to:
-- 1. Support more than 999 records (EP-001 to EP-999, then EP-1000, EP-1001, etc.)
-- 2. Mark functions as VOLATILE to prevent caching
-- 3. Remove advisory locks (they don't work with connection pooling)

-- Function to get next epic reference ID
CREATE OR REPLACE FUNCTION get_next_epic_ref_id() RETURNS VARCHAR(20) AS $$
DECLARE
    next_id BIGINT;
BEGIN
    next_id := nextval('epic_ref_seq');
    IF next_id < 1000 THEN
        RETURN 'EP-' || LPAD(next_id::TEXT, 3, '0');
    ELSE
        RETURN 'EP-' || next_id::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql VOLATILE;

-- Function to get next user story reference ID
CREATE OR REPLACE FUNCTION get_next_user_story_ref_id() RETURNS VARCHAR(20) AS $$
DECLARE
    next_id BIGINT;
BEGIN
    next_id := nextval('user_story_ref_seq');
    IF next_id < 1000 THEN
        RETURN 'US-' || LPAD(next_id::TEXT, 3, '0');
    ELSE
        RETURN 'US-' || next_id::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql VOLATILE;

-- Function to get next acceptance criteria reference ID
CREATE OR REPLACE FUNCTION get_next_acceptance_criteria_ref_id() RETURNS VARCHAR(20) AS $$
DECLARE
    next_id BIGINT;
BEGIN
    next_id := nextval('acceptance_criteria_ref_seq');
    IF next_id < 1000 THEN
        RETURN 'AC-' || LPAD(next_id::TEXT, 3, '0');
    ELSE
        RETURN 'AC-' || next_id::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql VOLATILE;

-- Function to get next requirement reference ID
CREATE OR REPLACE FUNCTION get_next_requirement_ref_id() RETURNS VARCHAR(20) AS $$
DECLARE
    next_id BIGINT;
BEGIN
    next_id := nextval('requirement_ref_seq');
    IF next_id < 1000 THEN
        RETURN 'REQ-' || LPAD(next_id::TEXT, 3, '0');
    ELSE
        RETURN 'REQ-' || next_id::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql VOLATILE;
