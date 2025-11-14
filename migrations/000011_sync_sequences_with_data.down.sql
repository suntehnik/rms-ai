-- This migration cannot be safely reverted
-- Sequences have been synchronized with existing data
-- Rolling back would cause duplicate key errors

-- If you really need to revert, you would need to:
-- 1. Manually set sequences back to their previous values
-- 2. But this is dangerous and not recommended

SELECT 'Migration 000011 cannot be safely reverted' as warning;
