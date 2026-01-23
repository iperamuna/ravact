package system

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Port            string
	RequirePass     string
	MaxMemory       string
	MaxMemoryPolicy string
	ConfigPath      string
}

// RedisManager handles Redis configuration operations
type RedisManager struct {
	configPath string
}

// NewRedisManager creates a new Redis manager
func NewRedisManager() *RedisManager {
	// Try common Redis config paths
	configPaths := []string{
		"/etc/redis/redis.conf",
		"/etc/redis.conf",
		"/usr/local/etc/redis.conf",
	}
	
	configPath := "/etc/redis/redis.conf" // Default
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}
	
	return &RedisManager{
		configPath: configPath,
	}
}

// GetConfig reads current Redis configuration
func (rm *RedisManager) GetConfig() (*RedisConfig, error) {
	data, err := os.ReadFile(rm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	config := &RedisConfig{
		ConfigPath: rm.configPath,
		Port:       "6379", // Default
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		switch parts[0] {
		case "port":
			config.Port = parts[1]
		case "requirepass":
			config.RequirePass = parts[1]
		case "maxmemory":
			config.MaxMemory = parts[1]
		case "maxmemory-policy":
			config.MaxMemoryPolicy = parts[1]
		}
	}
	
	return config, nil
}

// SetPassword sets Redis password (requirepass)
func (rm *RedisManager) SetPassword(password string) error {
	data, err := os.ReadFile(rm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	lines := strings.Split(string(data), "\n")
	found := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "requirepass") {
			// Update existing line
			lines[i] = fmt.Sprintf("requirepass %s", password)
			found = true
			break
		}
	}
	
	// If not found, add it
	if !found {
		// Find a good place to add it (after port or at end)
		insertIdx := len(lines)
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "port") {
				insertIdx = i + 1
				break
			}
		}
		
		newLine := fmt.Sprintf("requirepass %s", password)
		lines = append(lines[:insertIdx], append([]string{newLine}, lines[insertIdx:]...)...)
	}
	
	// Write back
	newConfig := strings.Join(lines, "\n")
	if err := os.WriteFile(rm.configPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// SetPort changes Redis port
func (rm *RedisManager) SetPort(port string) error {
	data, err := os.ReadFile(rm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	lines := strings.Split(string(data), "\n")
	found := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "port") && !strings.HasPrefix(trimmed, "#") {
			// Update existing line
			lines[i] = fmt.Sprintf("port %s", port)
			found = true
			break
		}
	}
	
	// If not found, add it
	if !found {
		lines = append([]string{fmt.Sprintf("port %s", port)}, lines...)
	}
	
	// Write back
	newConfig := strings.Join(lines, "\n")
	if err := os.WriteFile(rm.configPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// TestConnection tests Redis connection
func (rm *RedisManager) TestConnection() error {
	config, err := rm.GetConfig()
	if err != nil {
		return err
	}
	
	args := []string{"-p", config.Port, "ping"}
	if config.RequirePass != "" {
		args = []string{"-p", config.Port, "-a", config.RequirePass, "ping"}
	}
	
	cmd := exec.Command("redis-cli", args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("connection failed: %s", string(output))
	}
	
	if !strings.Contains(string(output), "PONG") {
		return fmt.Errorf("unexpected response: %s", string(output))
	}
	
	return nil
}

// RestartRedis restarts Redis service
func (rm *RedisManager) RestartRedis() error {
	cmd := exec.Command("systemctl", "restart", "redis-server")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		// Try alternative service name
		cmd = exec.Command("systemctl", "restart", "redis")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to restart: %s", string(output))
		}
	}
	
	return nil
}

// GetStatus gets Redis service status
func (rm *RedisManager) GetStatus() (string, error) {
	cmd := exec.Command("systemctl", "is-active", "redis-server")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		// Try alternative service name
		cmd = exec.Command("systemctl", "is-active", "redis")
		output, _ = cmd.CombinedOutput()
	}
	
	return strings.TrimSpace(string(output)), nil
}
