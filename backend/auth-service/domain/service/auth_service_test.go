package service

import (
	"context"
	"testing"
	"time"
	"twigger-backend/backend/auth-service/domain/entity"

	"github.com/google/uuid"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	users              map[string]*entity.User
	usersByFirebaseUID map[string]*entity.User
	usersByEmail       map[string]*entity.User
	workspaces         map[uuid.UUID][]*entity.Workspace
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:              make(map[string]*entity.User),
		usersByFirebaseUID: make(map[string]*entity.User),
		usersByEmail:       make(map[string]*entity.User),
		workspaces:         make(map[uuid.UUID][]*entity.Workspace),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.users[user.UserID.String()] = user
	if user.FirebaseUID != nil {
		m.usersByFirebaseUID[*user.FirebaseUID] = user
	}
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	if user, ok := m.users[userID.String()]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	if user, ok := m.usersByFirebaseUID[firebaseUID]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if user, ok := m.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.users[user.UserID.String()] = user
	return nil
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	if user, ok := m.users[userID.String()]; ok {
		now := time.Now()
		user.DeletedAt = &now
		return nil
	}
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	if user, ok := m.users[userID.String()]; ok {
		now := time.Now()
		user.LastLoginAt = &now
		return nil
	}
	return nil
}

func (m *MockUserRepository) LinkProvider(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	return nil
}

func (m *MockUserRepository) GetLinkedAccounts(ctx context.Context, userID uuid.UUID) ([]*entity.LinkedAccount, error) {
	return []*entity.LinkedAccount{}, nil
}

func (m *MockUserRepository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	if workspaces, ok := m.workspaces[userID]; ok {
		return workspaces, nil
	}
	return []*entity.Workspace{}, nil
}

// MockWorkspaceRepository is a mock implementation of WorkspaceRepository
type MockWorkspaceRepository struct {
	workspaces map[string]*entity.Workspace
	members    map[string][]*entity.WorkspaceMember
}

func NewMockWorkspaceRepository() *MockWorkspaceRepository {
	return &MockWorkspaceRepository{
		workspaces: make(map[string]*entity.Workspace),
		members:    make(map[string][]*entity.WorkspaceMember),
	}
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	m.workspaces[workspace.WorkspaceID.String()] = workspace
	return nil
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error) {
	return m.workspaces[workspaceID.String()], nil
}

func (m *MockWorkspaceRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entity.Workspace, error) {
	return []*entity.Workspace{}, nil
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	return nil
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	return nil
}

func (m *MockWorkspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	key := member.WorkspaceID.String()
	m.members[key] = append(m.members[key], member)
	return nil
}

func (m *MockWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return nil
}

func (m *MockWorkspaceRepository) GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error) {
	return m.members[workspaceID.String()], nil
}

func (m *MockWorkspaceRepository) GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	return "admin", nil
}

func (m *MockWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	return nil
}

func (m *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error) {
	return []*entity.Workspace{}, nil
}

func (m *MockWorkspaceRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	return true, nil
}

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	sessions map[string]*entity.Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*entity.Session),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	m.sessions[session.SessionID.String()] = session
	return nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*entity.Session, error) {
	return m.sessions[sessionID.String()], nil
}

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	return []*entity.Session{}, nil
}

func (m *MockSessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Session, error) {
	return []*entity.Session{}, nil
}

func (m *MockSessionRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// MockAuditRepository is a mock implementation of AuditRepository
type MockAuditRepository struct {
	events []*entity.AuditEvent
}

func NewMockAuditRepository() *MockAuditRepository {
	return &MockAuditRepository{
		events: []*entity.AuditEvent{},
	}
}

func (m *MockAuditRepository) LogEvent(ctx context.Context, event *entity.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockAuditRepository) GetUserEvents(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*entity.AuditEvent, error) {
	return m.events, nil
}

func (m *MockAuditRepository) GetUserEventsByType(ctx context.Context, userID uuid.UUID, eventType entity.AuditEventType, limit int) ([]*entity.AuditEvent, error) {
	return m.events, nil
}

func (m *MockAuditRepository) GetEventsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.AuditEvent, error) {
	return m.events, nil
}

func (m *MockAuditRepository) CountEventsByType(ctx context.Context, eventType entity.AuditEventType, startDate, endDate time.Time) (int64, error) {
	return int64(len(m.events)), nil
}

func (m *MockAuditRepository) GetFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	return 0, nil
}

// TestCompleteAuthentication_NewUser tests new user registration
func TestCompleteAuthentication_NewUser(t *testing.T) {
	// Setup mocks
	userRepo := NewMockUserRepository()
	workspaceRepo := NewMockWorkspaceRepository()
	sessionRepo := NewMockSessionRepository()
	auditRepo := NewMockAuditRepository()

	// Note: In real tests, we'd use a test database or mock DB
	// For now, passing nil for db as we're using mocks
	service := NewAuthService(userRepo, workspaceRepo, sessionRepo, auditRepo, nil)

	// Test data
	firebaseUID := "test-firebase-uid-123"
	email := "test@example.com"
	provider := "google.com"
	emailVerified := true

	// Execute
	ctx := context.Background()
	response, err := service.CompleteAuthentication(
		ctx,
		firebaseUID,
		email,
		provider,
		emailVerified,
		nil,
		nil,
		nil,
		nil,
	)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.IsNewUser {
		t.Error("Expected IsNewUser to be true for new user")
	}

	if response.User == nil {
		t.Fatal("Expected user, got nil")
	}

	if response.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, response.User.Email)
	}

	if *response.User.FirebaseUID != firebaseUID {
		t.Errorf("Expected firebase_uid %s, got %s", firebaseUID, *response.User.FirebaseUID)
	}

	if response.User.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, response.User.Provider)
	}

	if len(response.Workspaces) == 0 {
		t.Error("Expected at least one workspace for new user")
	}

	if response.SessionID == uuid.Nil {
		t.Error("Expected valid session ID")
	}
}

// TestCompleteAuthentication_ExistingUser tests existing user login
func TestCompleteAuthentication_ExistingUser(t *testing.T) {
	// Setup mocks
	userRepo := NewMockUserRepository()
	workspaceRepo := NewMockWorkspaceRepository()
	sessionRepo := NewMockSessionRepository()
	auditRepo := NewMockAuditRepository()

	// Create existing user
	firebaseUID := "existing-firebase-uid"
	existingUser := &entity.User{
		UserID:        uuid.New(),
		FirebaseUID:   &firebaseUID,
		Email:         "existing@example.com",
		Username:      "existing_user",
		Provider:      "google.com",
		EmailVerified: true,
		CreatedAt:     time.Now().Add(-24 * time.Hour),
	}
	userRepo.Create(context.Background(), existingUser)

	// Add workspace for existing user
	workspace := &entity.Workspace{
		WorkspaceID: uuid.New(),
		OwnerID:     existingUser.UserID,
		Name:        "Existing User's Garden",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}
	userRepo.workspaces[existingUser.UserID] = []*entity.Workspace{workspace}

	service := NewAuthService(userRepo, workspaceRepo, sessionRepo, auditRepo, nil)

	// Execute
	ctx := context.Background()
	response, err := service.CompleteAuthentication(
		ctx,
		firebaseUID,
		existingUser.Email,
		existingUser.Provider,
		true,
		nil,
		nil,
		nil,
		nil,
	)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.IsNewUser {
		t.Error("Expected IsNewUser to be false for existing user")
	}

	if response.User.UserID != existingUser.UserID {
		t.Errorf("Expected user ID %s, got %s", existingUser.UserID, response.User.UserID)
	}

	if len(response.Workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(response.Workspaces))
	}

	if response.User.LastLoginAt == nil {
		t.Error("Expected last_login_at to be updated")
	}
}

// TestGenerateUsername tests username generation
func TestGenerateUsername(t *testing.T) {
	tests := []struct {
		email    string
		contains string
	}{
		{"test@example.com", "test_"},
		{"john.doe@example.com", "john_doe_"},
		{"user+tag@example.com", "user_tag_"},
	}

	for _, tt := range tests {
		username := generateUsername(tt.email)
		if len(username) == 0 {
			t.Errorf("Expected non-empty username for email %s", tt.email)
		}
		// Username should contain sanitized email prefix
		if !contains(username, tt.contains[:len(tt.contains)-1]) {
			t.Errorf("Expected username to contain %s, got %s", tt.contains, username)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
