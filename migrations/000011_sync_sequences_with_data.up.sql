-- Synchronize sequences with existing data
-- This is necessary if reference IDs were created manually or by scripts
-- that bypassed the sequence mechanism

-- Sync epic sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    -- Get the maximum reference ID from the table
    SELECT reference_id INTO max_ref_id
    FROM epics
    ORDER BY reference_id DESC
    LIMIT 1;
    
    -- Extract the numeric part (remove 'EP-' prefix)
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 4) AS BIGINT);
        
        -- Set sequence to max + 1
        PERFORM setval('epic_ref_seq', max_number, true);
        
        RAISE NOTICE 'Epic sequence synced to %', max_number;
    END IF;
END $$;

-- Sync user story sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    SELECT reference_id INTO max_ref_id
    FROM user_stories
    ORDER BY reference_id DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 4) AS BIGINT);
        PERFORM setval('user_story_ref_seq', max_number, true);
        RAISE NOTICE 'User Story sequence synced to %', max_number;
    END IF;
END $$;

-- Sync acceptance criteria sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    SELECT reference_id INTO max_ref_id
    FROM acceptance_criteria
    ORDER BY reference_id DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 4) AS BIGINT);
        PERFORM setval('acceptance_criteria_ref_seq', max_number, true);
        RAISE NOTICE 'Acceptance Criteria sequence synced to %', max_number;
    END IF;
END $$;

-- Sync requirement sequence
DO $$
DECLARE
    max_ref_id TEXT;
    max_number BIGINT;
BEGIN
    SELECT reference_id INTO max_ref_id
    FROM requirements
    ORDER BY reference_id DESC
    LIMIT 1;
    
    IF max_ref_id IS NOT NULL THEN
        max_number := CAST(SUBSTRING(max_ref_id FROM 4) AS BIGINT);
        PERFORM setval('requirement_ref_seq', max_number, true);
        RAISE NOTICE 'Requirement sequence synced to %', max_number;
    END IF;
END $$;

-- Verify synchronization
SELECT 
    'epic' as entity,
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM epics) as max_in_table,
    currval('epic_ref_seq') as sequence_value
UNION ALL
SELECT 
    'user_story',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM user_stories),
    currval('user_story_ref_seq')
UNION ALL
SELECT 
    'acceptance_criteria',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM acceptance_criteria),
    currval('acceptance_criteria_ref_seq')
UNION ALL
SELECT 
    'requirement',
    (SELECT MAX(CAST(SUBSTRING(reference_id FROM 4) AS BIGINT)) FROM requirements),
    currval('requirement_ref_seq');
