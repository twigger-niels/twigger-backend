-- ============================================================================
-- Migration 008 Rollback: Remove Authentication and Workspace Support
-- Description: Rollback auth fields and workspace tables
-- Date: 2025-01-27
-- ============================================================================

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_add_workspace_owner ON workspaces;
DROP FUNCTION IF EXISTS add_workspace_owner_as_admin();

-- Drop linked accounts table
DROP TABLE IF EXISTS linked_accounts;

-- Drop audit log partitions and table
DROP TABLE IF EXISTS auth_audit_log_2025_02;
DROP TABLE IF EXISTS auth_audit_log_2025_01;
DROP TABLE IF EXISTS auth_audit_log;

-- Drop sessions table
DROP TABLE IF EXISTS auth_sessions;

-- Remove workspace_id from gardens
ALTER TABLE gardens DROP COLUMN IF EXISTS workspace_id;

-- Drop workspace tables
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;

-- Remove auth fields from users table
ALTER TABLE users
DROP COLUMN IF EXISTS country_id,
DROP COLUMN IF EXISTS preferred_language_id,
DROP COLUMN IF EXISTS deleted_at,
DROP COLUMN IF EXISTS last_login_at,
DROP COLUMN IF EXISTS provider,
DROP COLUMN IF EXISTS photo_url,
DROP COLUMN IF EXISTS phone_number,
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS firebase_uid;
