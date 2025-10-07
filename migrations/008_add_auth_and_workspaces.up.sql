-- ============================================================================
-- Migration 008: Authentication and Workspace Support
-- Description: Add Firebase auth fields to users table and create workspace tables
-- Date: 2025-01-27
-- ============================================================================

-- ============================================================================
-- SECTION 1: Extend Users Table with Auth Fields
-- ============================================================================

ALTER TABLE users
ADD COLUMN firebase_uid VARCHAR(128) UNIQUE,
ADD COLUMN email_verified BOOLEAN DEFAULT FALSE,
ADD COLUMN phone_number VARCHAR(20),
ADD COLUMN photo_url TEXT,
ADD COLUMN provider VARCHAR(50),  -- 'email', 'google.com', 'facebook.com', 'apple.com'
ADD COLUMN last_login_at TIMESTAMPTZ,
ADD COLUMN deleted_at TIMESTAMPTZ,
ADD COLUMN preferred_language_id UUID,
ADD COLUMN country_id UUID;

-- Add indexes for auth fields
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid) WHERE firebase_uid IS NOT NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_last_login ON users(last_login_at DESC);
CREATE INDEX idx_users_email ON users(email);

-- ============================================================================
-- SECTION 2: Create Workspace Tables
-- ============================================================================

-- Workspaces table (multi-tenant isolation)
CREATE TABLE workspaces (
    workspace_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);
CREATE INDEX idx_workspaces_created_at ON workspaces(created_at DESC);

-- Workspace members table (role-based access control)
CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',  -- 'admin', 'member', 'viewer'
    joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_workspace_id ON workspace_members(workspace_id);

-- ============================================================================
-- SECTION 3: Update Gardens Table to Support Workspaces
-- ============================================================================

-- Add workspace_id to gardens table
ALTER TABLE gardens
ADD COLUMN workspace_id UUID REFERENCES workspaces(workspace_id) ON DELETE CASCADE;

-- Create index for workspace-based queries
CREATE INDEX idx_gardens_workspace_id ON gardens(workspace_id);

-- ============================================================================
-- SECTION 4: Auth Sessions Table
-- ============================================================================

-- Sessions table (for audit and revocation)
CREATE TABLE auth_sessions (
    session_id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_id VARCHAR(255),
    device_info JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_auth_sessions_user_id ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires_at ON auth_sessions(expires_at);
CREATE INDEX idx_auth_sessions_device_id ON auth_sessions(device_id) WHERE device_id IS NOT NULL;

-- ============================================================================
-- SECTION 5: Auth Audit Log Table (Partitioned)
-- ============================================================================

-- Audit log table (partitioned by month for performance)
CREATE TABLE auth_audit_log (
    id BIGSERIAL,
    user_id UUID REFERENCES users(user_id),
    event_type VARCHAR(50) NOT NULL,  -- 'login', 'logout', 'register', 'token_refresh', 'session_revoked'
    success BOOLEAN NOT NULL,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create first partition (January 2025)
CREATE TABLE auth_audit_log_2025_01 PARTITION OF auth_audit_log
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Create February 2025 partition
CREATE TABLE auth_audit_log_2025_02 PARTITION OF auth_audit_log
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Create indexes on partitioned table
CREATE INDEX idx_auth_audit_user_id ON auth_audit_log(user_id);
CREATE INDEX idx_auth_audit_created_at ON auth_audit_log(created_at DESC);
CREATE INDEX idx_auth_audit_event_type ON auth_audit_log(event_type);
CREATE INDEX idx_auth_audit_success ON auth_audit_log(success);

-- ============================================================================
-- SECTION 6: Linked Accounts Table
-- ============================================================================

-- Linked accounts (social login tracking)
CREATE TABLE linked_accounts (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,  -- 'google.com', 'facebook.com', 'apple.com'
    provider_user_id VARCHAR(255) NOT NULL,
    linked_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_linked_accounts_user_id ON linked_accounts(user_id);
CREATE INDEX idx_linked_accounts_provider ON linked_accounts(provider);

-- ============================================================================
-- SECTION 7: Helper Function for Workspace Member Management
-- ============================================================================

-- Function to automatically add workspace owner as admin
CREATE OR REPLACE FUNCTION add_workspace_owner_as_admin()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO workspace_members (workspace_id, user_id, role)
    VALUES (NEW.workspace_id, NEW.owner_id, 'admin');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to add owner as admin on workspace creation
CREATE TRIGGER trigger_add_workspace_owner
    AFTER INSERT ON workspaces
    FOR EACH ROW
    EXECUTE FUNCTION add_workspace_owner_as_admin();

-- ============================================================================
-- SECTION 8: Comments for Documentation
-- ============================================================================

COMMENT ON TABLE workspaces IS 'Multi-tenant workspaces for organizing gardens and plants';
COMMENT ON TABLE workspace_members IS 'User memberships in workspaces with role-based access control';
COMMENT ON TABLE auth_sessions IS 'Active and revoked authentication sessions for audit trail';
COMMENT ON TABLE auth_audit_log IS 'Comprehensive audit log for all authentication events (partitioned by month)';
COMMENT ON TABLE linked_accounts IS 'Social provider accounts linked to user accounts';

COMMENT ON COLUMN users.firebase_uid IS 'Firebase Authentication unique identifier';
COMMENT ON COLUMN users.provider IS 'Authentication provider: email, google.com, facebook.com, apple.com';
COMMENT ON COLUMN users.preferred_language_id IS 'User preferred language for localization';
COMMENT ON COLUMN users.country_id IS 'User country for regional localization';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp for GDPR compliance';

COMMENT ON COLUMN workspace_members.role IS 'User role in workspace: admin, member, viewer';
COMMENT ON COLUMN auth_sessions.revoked_at IS 'Timestamp when session was revoked (for logout/security)';
COMMENT ON COLUMN auth_audit_log.event_type IS 'Type of auth event: login, logout, register, token_refresh, session_revoked';
