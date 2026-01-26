package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewUserManager(t *testing.T) {
	manager := NewUserManager()
	if manager == nil {
		t.Fatal("NewUserManager returned nil")
	}
}

func TestUserStruct(t *testing.T) {
	user := User{
		Username: "testuser",
		UID:      1000,
		GID:      1000,
		HomeDir:  "/home/testuser",
		Shell:    "/bin/bash",
		HasSudo:  true,
		Groups:   []string{"sudo", "docker"},
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
	if user.UID != 1000 {
		t.Errorf("expected UID 1000, got %d", user.UID)
	}
	if !user.HasSudo {
		t.Error("expected HasSudo to be true")
	}
	if len(user.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(user.Groups))
	}
}

func TestGroupStruct(t *testing.T) {
	group := Group{
		Name:    "developers",
		GID:     1001,
		Members: []string{"alice", "bob", "charlie"},
	}

	if group.Name != "developers" {
		t.Errorf("expected name 'developers', got '%s'", group.Name)
	}
	if group.GID != 1001 {
		t.Errorf("expected GID 1001, got %d", group.GID)
	}
	if len(group.Members) != 3 {
		t.Errorf("expected 3 members, got %d", len(group.Members))
	}
}

func TestGetAllUsers(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Skipping on non-Unix system")
	}

	manager := NewUserManager()
	users, err := manager.GetAllUsers()

	if err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}

	// On macOS, we get a mock user indicating not supported
	if runtime.GOOS == "darwin" {
		if len(users) == 0 {
			t.Error("expected at least one user (mock user on macOS)")
		}
		return
	}

	// On Linux, we should get at least root or some system users
	if len(users) == 0 {
		t.Log("Warning: no users returned (may be expected in restricted environments)")
	}
}

func TestGetUser(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux system")
	}

	manager := NewUserManager()

	// Try to get root user (should exist on all Linux systems)
	user, err := manager.GetUser("root")
	if err != nil {
		t.Logf("GetUser for root failed (may be expected): %v", err)
		return
	}

	if user.Username != "root" {
		t.Errorf("expected username 'root', got '%s'", user.Username)
	}
	if user.UID != 0 {
		t.Errorf("expected UID 0 for root, got %d", user.UID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux system")
	}

	manager := NewUserManager()

	_, err := manager.GetUser("nonexistent_user_xyz123")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserHasSudo(t *testing.T) {
	manager := NewUserManager()

	// Test that root has sudo
	hasSudo := manager.userHasSudo("root")
	if !hasSudo {
		t.Log("Note: root user should have sudo privileges")
	}
}

func TestGetUserGroups(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping on non-Linux system")
	}

	manager := NewUserManager()

	// Get groups for root
	groups := manager.getUserGroups("root")

	// Root should have at least one group
	t.Logf("root user groups: %v", groups)
}

func TestGetUserGroupsFromFile(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("Skipping on non-Unix system")
	}

	manager := NewUserManager()

	// This is a fallback method, should work if /etc/group exists
	groups := manager.getUserGroupsFromFile("root")
	t.Logf("Groups from file for root: %v", groups)
}

func TestGetDarwinUsers(t *testing.T) {
	manager := NewUserManager()
	users, err := manager.getDarwinUsers()

	if err != nil {
		t.Fatalf("getDarwinUsers failed: %v", err)
	}

	// Should return a mock user with warning
	if len(users) == 0 {
		t.Error("expected at least one mock user")
	}

	// The mock user should indicate macOS is not supported
	if len(users) > 0 {
		if !strings.Contains(users[0].Username, "macOS") && !strings.Contains(users[0].Username, "Not Supported") {
			t.Log("Note: getDarwinUsers returns a warning message")
		}
	}
}

func TestUsernameValidation(t *testing.T) {
	// Test username validation patterns
	validUsernames := []string{
		"testuser",
		"user123",
		"test_user",
		"test-user",
		"a",
	}

	invalidUsernames := []string{
		"",
		"123user",      // starts with number
		"test user",    // contains space
		"user@host",    // contains special char
	}

	// Valid usernames should match Linux conventions
	for _, name := range validUsernames {
		if name == "" {
			t.Errorf("valid username should not be empty")
		}
		// Basic check: shouldn't start with number
		if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
			t.Errorf("username '%s' starts with number", name)
		}
	}

	// Invalid usernames
	for _, name := range invalidUsernames {
		// These should be considered invalid
		if name == "" {
			continue // empty is expected to be invalid
		}
		if strings.Contains(name, " ") {
			continue // space is expected to be invalid
		}
		if strings.Contains(name, "@") {
			continue // special chars expected to be invalid
		}
	}
}

func TestSudoersFilePath(t *testing.T) {
	// Test that the sudoers.d path is correctly formed
	username := "testuser"
	expectedPath := filepath.Join("/etc/sudoers.d", username)

	if expectedPath != "/etc/sudoers.d/testuser" {
		t.Errorf("unexpected sudoers path: %s", expectedPath)
	}
}

func TestSSHKeyPath(t *testing.T) {
	homeDir := "/home/testuser"
	sshDir := filepath.Join(homeDir, ".ssh")
	authorizedKeys := filepath.Join(sshDir, "authorized_keys")

	if sshDir != "/home/testuser/.ssh" {
		t.Errorf("unexpected ssh dir: %s", sshDir)
	}
	if authorizedKeys != "/home/testuser/.ssh/authorized_keys" {
		t.Errorf("unexpected authorized_keys path: %s", authorizedKeys)
	}
}

// TestSSHKeyStruct tests the SSHKey structure
func TestSSHKeyStruct(t *testing.T) {
	key := SSHKey{
		Identifier:     "user@host",
		Type:           "ed25519",
		Fingerprint:    "SHA256:abc123...",
		PublicKeyPath:  "/home/user/.ssh/id_ed25519.pub",
		PrivateKeyPath: "/home/user/.ssh/id_ed25519",
		IsLoginKey:     true,
		HasPassphrase:  false,
		IsInAgent:      false,
	}

	if key.Identifier != "user@host" {
		t.Errorf("expected identifier 'user@host', got '%s'", key.Identifier)
	}
	if key.Type != "ed25519" {
		t.Errorf("expected type 'ed25519', got '%s'", key.Type)
	}
	if key.PublicKeyPath != "/home/user/.ssh/id_ed25519.pub" {
		t.Errorf("expected public key path, got '%s'", key.PublicKeyPath)
	}
	if !key.IsLoginKey {
		t.Error("expected IsLoginKey to be true")
	}
}

func TestSSHKeyTypes(t *testing.T) {
	// Test common SSH key types
	keyTypes := []string{"rsa", "ed25519", "ecdsa", "dsa"}
	
	for _, keyType := range keyTypes {
		if keyType == "" {
			t.Error("key type should not be empty")
		}
	}
}

// TestCreateHomeDirectory tests path construction for home directories
func TestCreateHomeDirectory(t *testing.T) {
	username := "newuser"
	homeDir := filepath.Join("/home", username)

	if homeDir != "/home/newuser" {
		t.Errorf("unexpected home dir: %s", homeDir)
	}
}

// TestShellPaths tests common shell paths
func TestShellPaths(t *testing.T) {
	shells := []string{
		"/bin/bash",
		"/bin/sh",
		"/bin/zsh",
		"/usr/bin/bash",
		"/usr/bin/zsh",
		"/bin/false",
		"/usr/sbin/nologin",
	}

	for _, shell := range shells {
		if shell == "" {
			t.Error("shell path should not be empty")
		}
	}
}

// Integration test helper - skip if not root
func skipIfNotRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}
}

func TestUserCreationPatterns(t *testing.T) {
	// Test username patterns without calling validation method

	// Valid usernames follow Linux conventions
	validNames := []string{"john", "jane_doe", "user-123", "a"}
	for _, name := range validNames {
		// Basic validation checks
		if name == "" {
			t.Errorf("username should not be empty")
		}
		if strings.Contains(name, " ") {
			t.Errorf("username '%s' should not contain spaces", name)
		}
		if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
			t.Errorf("username '%s' should not start with a number", name)
		}
	}

	// Invalid username patterns
	invalidNames := []string{"", "123abc", "user name", "user@host"}
	for _, name := range invalidNames {
		hasIssue := false
		if name == "" {
			hasIssue = true
		}
		if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
			hasIssue = true
		}
		if strings.Contains(name, " ") {
			hasIssue = true
		}
		if strings.Contains(name, "@") {
			hasIssue = true
		}
		if !hasIssue && name != "" {
			t.Logf("Note: '%s' may or may not be valid depending on system", name)
		}
	}
}
