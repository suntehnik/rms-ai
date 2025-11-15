-- Fix reference ID generation to support unlimited growth and ensure atomicity
-- This migration:
-- 1. Updates functions to support more than 999 records (EP-001 to EP-999, then EP-1000, EP-1001, etc.)
-- 2. Marks functions as VOLATILE to prevent caching
-- 3. Removes advisory locks (they don't work with connection pooling)
-- 4. Synchronizes sequences with existing data

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

-- Synchronize sequences with existing data
-- This is necessary if reference IDs were created manually or by scripts
-- that bypassed the sequence mechanism

-- Sync epic sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    -- Get the maximum valid numeric reference ID (EP-001, EP-002, etc.)
    -- Filter out any invalid reference IDs that don't match the expected pattern
    SELECT reference_id INTO max_ref_id
    FROM epics
    WHERE reference_id ~ '^EP-[0-9]+$'
    ORDER BY CAST(SUBSTRING(reference_id FROM 4) AS BIGINT) DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 4) AS BIGINT);
        PERFORM setval('epic_ref_seq', GREATEST(max_number, 1), true);
        RAISE NOTICE 'Epic sequence synced to %', max_number;
    ELSE
        -- No valid reference IDs found, reset to 1
        PERFORM setval('epic_ref_seq', 1, false);
        RAISE NOTICE 'Epic sequence reset to 1 (no valid reference IDs found)';
    END IF;
END $$;

-- Sync user story sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    -- Get the maximum valid numeric reference ID (US-001, US-002, etc.)
    -- Filter out any invalid reference IDs that don't match the expected pattern
    SELECT reference_id INTO max_ref_id
    FROM user_stories
    WHERE reference_id ~ '^US-[0-9]+$'
    ORDER BY CAST(SUBSTRING(reference_id FROM 4) AS BIGINT) DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(reference_id FROM 4) AS BIGINT);
        PERFORM setval('user_story_ref_seq', GREATEST(max_number, 1), true);
        RAISE NOTICE 'User Story sequence synced to %', max_number;
    ELSE
        -- No valid reference IDs found, reset to 1
        PERFORM setval('user_story_ref_seq', 1, false);
        RAISE NOTICE 'User Story sequence reset to 1 (no valid reference IDs found)';
    END IF;
END $$;

-- Sync acceptance criteria sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    -- Get the maximum valid numeric reference ID (AC-001, AC-002, etc.)
    -- Filter out any invalid reference IDs that don't match the expected pattern
    SELECT reference_id INTO max_ref_id
    FROM acceptance_criteria
    WHERE reference_id ~ '^AC-[0-9]+$'
    ORDER BY CAST(SUBSTRING(reference_id FROM 4) AS BIGINT) DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(reference_id FROM 4) AS BIGINT);
        PERFORM setval('acceptance_criteria_ref_seq', GREATEST(max_number, 1), true);
        RAISE NOTICE 'Acceptance Criteria sequence synced to %', max_number;
    ELSE
        -- No valid reference IDs found, reset to 1
        PERFORM setval('acceptance_criteria_ref_seq', 1, false);
        RAISE NOTICE 'Acceptance Criteria sequence reset to 1 (no valid reference IDs found)';
    END IF;
END $$;

-- Sync requirement sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    -- Get the maximum valid numeric reference ID (REQ-001, REQ-002, etc.)
    -- Filter out any invalid reference IDs that don't match the expected pattern
    SELECT reference_id INTO max_ref_id
    FROM requirements
    WHERE reference_id ~ '^REQ-[0-9]+$'
    ORDER BY CAST(SUBSTRING(reference_id FROM 5) AS BIGINT) DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 5) AS BIGINT);
        PERFORM setval('requirement_ref_seq', GREATEST(max_number, 1), true);
        RAISE NOTICE 'Requirement sequence synced to %', max_number;
    ELSE
        -- No valid reference IDs found, reset to 1
        PERFORM setval('requirement_ref_seq', 1, false);
        RAISE NOTICE 'Requirement sequence reset to 1 (no valid reference IDs found)';
    END IF;
END $$;

-- Verify synchronization (only valid reference IDs)
SELECT 
    'epic' as entity,
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM epics WHERE reference_id ~ '^EP-[0-9]+$') as max_in_table,
    (SELECT last_value FROM epic_ref_seq) as sequence_value
UNION ALL
SELECT 
    'user_story',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM user_stories WHERE reference_id ~ '^US-[0-9]+$'),
    (SELECT last_value FROM user_story_ref_seq)
UNION ALL
SELECT 
    'acceptance_criteria',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM acceptance_criteria WHERE reference_id ~ '^AC-[0-9]+$'),
    (SELECT last_value FROM acceptance_criteria_ref_seq)
UNION ALL
SELECT 
    'requirement',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 5) AS BIGINT)) FROM requirements WHERE reference_id ~ '^REQ-[0-9]+$'),
    (SELECT last_value FROM requirement_ref_seq);
