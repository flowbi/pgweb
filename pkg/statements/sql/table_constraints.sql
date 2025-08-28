-- Flow.bi Custom Constraints Query with Safe Fallback
-- This query conditionally shows either Flow.bi custom constraints or real PostgreSQL constraints
-- based on environment variable and table availability

WITH flowbi_constraints AS (
  -- Step 1: Determine which mode to use based on safety checks
  SELECT 
    CASE 
      -- Check if Flow.bi mode is enabled via session variable
      WHEN current_setting('pgweb.custom_constraints', true) = 'true' 
        -- Verify the Flow.bi constraints table exists
        AND EXISTS (
          SELECT 1 FROM information_schema.tables 
          WHERE table_schema = 'intf_studio' 
            AND table_name = 'pgweb_constraints'
        )
        -- Verify all required columns exist (exactly 4 expected columns)
        AND (
          SELECT COUNT(*)
          FROM information_schema.columns 
          WHERE table_schema = 'intf_studio' 
            AND table_name = 'pgweb_constraints'
            AND column_name IN ('conname', 'definition', 'nspname', 'relname')
        ) = 4
      THEN 'use_flowbi'
      ELSE 'use_standard'
    END AS mode
),
flowbi_data AS (
  -- Step 2: Query Flow.bi custom constraints (only when safe)
  SELECT 
    fc.conname AS name,
    fc.definition
  FROM flowbi_constraints fc_mode
  CROSS JOIN intf_studio.pgweb_constraints fc
  WHERE fc_mode.mode = 'use_flowbi'
    AND fc.nspname = $1  -- Schema parameter
    AND fc.relname = $2  -- Table parameter
),
standard_data AS (
  -- Step 3: Query real PostgreSQL constraints (fallback mode)
  SELECT 
    c.conname AS name,
    pg_get_constraintdef(c.oid, true) AS definition
  FROM flowbi_constraints fc_mode
  CROSS JOIN pg_constraint c
  JOIN pg_namespace n ON n.oid = c.connamespace
  JOIN pg_class cl ON cl.oid = c.conrelid
  WHERE fc_mode.mode = 'use_standard'
    AND n.nspname = $1   -- Schema parameter
    AND cl.relname = $2  -- Table parameter
)
-- Step 4: Return results from either Flow.bi or standard mode (never both)
SELECT name, definition FROM flowbi_data
UNION ALL
SELECT name, definition FROM standard_data
ORDER BY name
