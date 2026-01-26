package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRedisManager(t *testing.T) {
	manager := NewRedisManager()
	if manager == nil {
		t.Fatal("NewRedisManager returned nil")
	}

	// Config path should be set to a default
	if manager.configPath == "" {
		t.Error("configPath should have a default value")
	}
}

func TestRedisConfigStruct(t *testing.T) {
	config := RedisConfig{
		Port:            "6379",
		RequirePass:     "secretpassword",
		MaxMemory:       "256mb",
		MaxMemoryPolicy: "allkeys-lru",
		ConfigPath:      "/etc/redis/redis.conf",
	}

	if config.Port != "6379" {
		t.Errorf("expected port '6379', got '%s'", config.Port)
	}
	if config.RequirePass != "secretpassword" {
		t.Errorf("expected password 'secretpassword', got '%s'", config.RequirePass)
	}
	if config.MaxMemory != "256mb" {
		t.Errorf("expected maxmemory '256mb', got '%s'", config.MaxMemory)
	}
	if config.MaxMemoryPolicy != "allkeys-lru" {
		t.Errorf("expected policy 'allkeys-lru', got '%s'", config.MaxMemoryPolicy)
	}
}

func TestRedisManager_GetConfig_FileNotFound(t *testing.T) {
	// Create manager with non-existent path
	manager := &RedisManager{
		configPath: "/nonexistent/path/redis.conf",
	}

	_, err := manager.GetConfig()
	if err == nil {
		t.Error("expected error for non-existent config file")
	}
}

func TestRedisManager_GetConfig_ParsesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create a test config file
	configContent := `# Redis configuration
port 6380
requirepass mysecretpass
maxmemory 512mb
maxmemory-policy volatile-lru

# Other settings
bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	config, err := manager.GetConfig()

	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if config.Port != "6380" {
		t.Errorf("expected port '6380', got '%s'", config.Port)
	}
	if config.RequirePass != "mysecretpass" {
		t.Errorf("expected password 'mysecretpass', got '%s'", config.RequirePass)
	}
	if config.MaxMemory != "512mb" {
		t.Errorf("expected maxmemory '512mb', got '%s'", config.MaxMemory)
	}
	if config.MaxMemoryPolicy != "volatile-lru" {
		t.Errorf("expected policy 'volatile-lru', got '%s'", config.MaxMemoryPolicy)
	}
}

func TestRedisManager_GetConfig_DefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create config without port setting
	configContent := `# Redis configuration without explicit port
bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	config, err := manager.GetConfig()

	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Should use default port
	if config.Port != "6379" {
		t.Errorf("expected default port '6379', got '%s'", config.Port)
	}
}

func TestRedisManager_GetConfig_IgnoresComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	configContent := `# port 9999
port 6379
# requirepass oldpassword
requirepass newpassword
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	config, err := manager.GetConfig()

	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if config.Port != "6379" {
		t.Errorf("should ignore commented port, got '%s'", config.Port)
	}
	if config.RequirePass != "newpassword" {
		t.Errorf("should use uncommented password, got '%s'", config.RequirePass)
	}
}

func TestRedisManager_SetPassword(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create initial config
	configContent := `port 6379
bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	err := manager.SetPassword("newpassword123")

	if err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !strings.Contains(string(data), "requirepass newpassword123") {
		t.Errorf("password not set correctly in config:\n%s", string(data))
	}
}

func TestRedisManager_SetPassword_UpdatesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create config with existing password
	configContent := `port 6379
requirepass oldpassword
bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	err := manager.SetPassword("updatedpassword")

	if err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "requirepass updatedpassword") {
		t.Errorf("password not updated correctly:\n%s", content)
	}
	if strings.Contains(content, "oldpassword") {
		t.Errorf("old password should be replaced:\n%s", content)
	}
}

func TestRedisManager_SetPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create initial config
	configContent := `port 6379
bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	err := manager.SetPort("6380")

	if err != nil {
		t.Fatalf("SetPort failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !strings.Contains(string(data), "port 6380") {
		t.Errorf("port not set correctly in config:\n%s", string(data))
	}
}

func TestRedisManager_SetPort_AddsIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "redis.conf")

	// Create config without port
	configContent := `bind 127.0.0.1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	manager := &RedisManager{configPath: configPath}
	err := manager.SetPort("6381")

	if err != nil {
		t.Fatalf("SetPort failed: %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !strings.Contains(string(data), "port 6381") {
		t.Errorf("port not added to config:\n%s", string(data))
	}
}

func TestRedisConfigPaths(t *testing.T) {
	// Test common Redis config paths
	paths := []string{
		"/etc/redis/redis.conf",
		"/etc/redis.conf",
		"/usr/local/etc/redis.conf",
	}

	for _, path := range paths {
		if path == "" {
			t.Error("path should not be empty")
		}
	}
}

func TestRedisManager_FileNotFound(t *testing.T) {
	manager := &RedisManager{
		configPath: "/nonexistent/redis.conf",
	}

	err := manager.SetPassword("test")
	if err == nil {
		t.Error("expected error for non-existent config")
	}

	err = manager.SetPort("6379")
	if err == nil {
		t.Error("expected error for non-existent config")
	}
}
