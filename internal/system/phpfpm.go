package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PHPFPMPool represents a PHP-FPM pool configuration
type PHPFPMPool struct {
	Name                string
	User                string
	Group               string
	Listen              string
	ListenOwner         string
	ListenGroup         string
	ListenMode          string
	PM                  string // static, dynamic, ondemand
	PMMaxChildren       int
	PMStartServers      int
	PMMinSpareServers   int
	PMMaxSpareServers   int
	PMMaxRequests       int
	ConfigPath          string
	PHPVersion          string
}

// PHPFPMManager handles PHP-FPM pool operations
type PHPFPMManager struct {
	phpVersion  string
	poolDir     string
}

// NewPHPFPMManager creates a new PHP-FPM manager
func NewPHPFPMManager(phpVersion string) *PHPFPMManager {
	if phpVersion == "" {
		phpVersion = "8.3" // Default version
	}
	
	return &PHPFPMManager{
		phpVersion: phpVersion,
		poolDir:    fmt.Sprintf("/etc/php/%s/fpm/pool.d", phpVersion),
	}
}

// DetectPHPVersion attempts to detect installed PHP version
func (p *PHPFPMManager) DetectPHPVersion() (string, error) {
	// Check common PHP versions
	versions := []string{"8.3", "8.2", "8.1", "8.0", "7.4"}
	
	for _, ver := range versions {
		poolDir := fmt.Sprintf("/etc/php/%s/fpm/pool.d", ver)
		if _, err := os.Stat(poolDir); err == nil {
			p.phpVersion = ver
			p.poolDir = poolDir
			return ver, nil
		}
	}
	
	return "", fmt.Errorf("no PHP-FPM installation found")
}

// ListPools returns all configured PHP-FPM pools
func (p *PHPFPMManager) ListPools() ([]PHPFPMPool, error) {
	if _, err := os.Stat(p.poolDir); err != nil {
		return nil, fmt.Errorf("pool directory not found: %s", p.poolDir)
	}

	files, err := filepath.Glob(filepath.Join(p.poolDir, "*.conf"))
	if err != nil {
		return nil, fmt.Errorf("failed to list pool files: %w", err)
	}

	pools := make([]PHPFPMPool, 0)
	for _, file := range files {
		pool, err := p.ReadPool(filepath.Base(file))
		if err != nil {
			continue // Skip invalid pools
		}
		pools = append(pools, *pool)
	}

	return pools, nil
}

// ReadPool reads a specific pool configuration
func (p *PHPFPMManager) ReadPool(poolName string) (*PHPFPMPool, error) {
	// Ensure .conf extension
	if !strings.HasSuffix(poolName, ".conf") {
		poolName = poolName + ".conf"
	}

	configPath := filepath.Join(p.poolDir, poolName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool config: %w", err)
	}

	pool := &PHPFPMPool{
		Name:              strings.TrimSuffix(poolName, ".conf"),
		ConfigPath:        configPath,
		PHPVersion:        p.phpVersion,
		User:              "www-data",
		Group:             "www-data",
		Listen:            "/run/php/php-fpm.sock",
		ListenOwner:       "www-data",
		ListenGroup:       "www-data",
		ListenMode:        "0660",
		PM:                "dynamic",
		PMMaxChildren:     5,
		PMStartServers:    2,
		PMMinSpareServers: 1,
		PMMaxSpareServers: 3,
		PMMaxRequests:     500,
	}

	// Parse configuration
	lines := strings.Split(string(data), "\n")
	inPoolSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, ";") || line == "" {
			continue
		}

		// Check for pool name
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			poolNameFromFile := strings.Trim(line, "[]")
			pool.Name = poolNameFromFile
			inPoolSection = true
			continue
		}

		if !inPoolSection {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user":
			pool.User = value
		case "group":
			pool.Group = value
		case "listen":
			pool.Listen = value
		case "listen.owner":
			pool.ListenOwner = value
		case "listen.group":
			pool.ListenGroup = value
		case "listen.mode":
			pool.ListenMode = value
		case "pm":
			pool.PM = value
		case "pm.max_children":
			fmt.Sscanf(value, "%d", &pool.PMMaxChildren)
		case "pm.start_servers":
			fmt.Sscanf(value, "%d", &pool.PMStartServers)
		case "pm.min_spare_servers":
			fmt.Sscanf(value, "%d", &pool.PMMinSpareServers)
		case "pm.max_spare_servers":
			fmt.Sscanf(value, "%d", &pool.PMMaxSpareServers)
		case "pm.max_requests":
			fmt.Sscanf(value, "%d", &pool.PMMaxRequests)
		}
	}

	return pool, nil
}

// CreatePool creates a new PHP-FPM pool
func (p *PHPFPMManager) CreatePool(pool *PHPFPMPool) error {
	if pool.Name == "" {
		return fmt.Errorf("pool name is required")
	}

	// Set defaults if not provided
	if pool.User == "" {
		pool.User = "www-data"
	}
	if pool.Group == "" {
		pool.Group = "www-data"
	}
	if pool.Listen == "" {
		pool.Listen = fmt.Sprintf("/run/php/php%s-%s-fpm.sock", p.phpVersion, pool.Name)
	}
	if pool.ListenOwner == "" {
		pool.ListenOwner = "www-data"
	}
	if pool.ListenGroup == "" {
		pool.ListenGroup = "www-data"
	}
	if pool.ListenMode == "" {
		pool.ListenMode = "0660"
	}
	if pool.PM == "" {
		pool.PM = "dynamic"
	}
	if pool.PMMaxChildren == 0 {
		pool.PMMaxChildren = 5
	}
	if pool.PMStartServers == 0 {
		pool.PMStartServers = 2
	}
	if pool.PMMinSpareServers == 0 {
		pool.PMMinSpareServers = 1
	}
	if pool.PMMaxSpareServers == 0 {
		pool.PMMaxSpareServers = 3
	}
	if pool.PMMaxRequests == 0 {
		pool.PMMaxRequests = 500
	}

	configPath := filepath.Join(p.poolDir, pool.Name+".conf")
	pool.ConfigPath = configPath

	// Check if pool already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("pool '%s' already exists", pool.Name)
	}

	// Generate pool configuration
	config := p.generatePoolConfig(pool)

	// Write config file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	return nil
}

// UpdatePool updates an existing PHP-FPM pool
func (p *PHPFPMManager) UpdatePool(pool *PHPFPMPool) error {
	if pool.Name == "" {
		return fmt.Errorf("pool name is required")
	}

	configPath := filepath.Join(p.poolDir, pool.Name+".conf")
	pool.ConfigPath = configPath

	// Check if pool exists
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("pool '%s' not found", pool.Name)
	}

	// Backup existing config
	backupPath := configPath + ".bak"
	data, _ := os.ReadFile(configPath)
	os.WriteFile(backupPath, data, 0644)

	// Generate new pool configuration
	config := p.generatePoolConfig(pool)

	// Write config file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		// Restore backup on failure
		if data != nil {
			os.WriteFile(configPath, data, 0644)
		}
		return fmt.Errorf("failed to write pool config: %w", err)
	}

	return nil
}

// DeletePool deletes a PHP-FPM pool
func (p *PHPFPMManager) DeletePool(poolName string) error {
	if poolName == "" {
		return fmt.Errorf("pool name is required")
	}

	// Don't allow deleting the www pool
	if poolName == "www" {
		return fmt.Errorf("cannot delete the default 'www' pool")
	}

	configPath := filepath.Join(p.poolDir, poolName+".conf")

	// Check if pool exists
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("pool '%s' not found", poolName)
	}

	// Delete the config file
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to delete pool: %w", err)
	}

	return nil
}

// generatePoolConfig generates the pool configuration content
func (p *PHPFPMManager) generatePoolConfig(pool *PHPFPMPool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("; Pool: %s\n", pool.Name))
	sb.WriteString(fmt.Sprintf("; Generated by Ravact\n\n"))
	sb.WriteString(fmt.Sprintf("[%s]\n\n", pool.Name))
	
	sb.WriteString("; Unix user/group of processes\n")
	sb.WriteString(fmt.Sprintf("user = %s\n", pool.User))
	sb.WriteString(fmt.Sprintf("group = %s\n\n", pool.Group))
	
	sb.WriteString("; The address on which to accept FastCGI requests\n")
	sb.WriteString(fmt.Sprintf("listen = %s\n\n", pool.Listen))
	
	sb.WriteString("; Set permissions for unix socket\n")
	sb.WriteString(fmt.Sprintf("listen.owner = %s\n", pool.ListenOwner))
	sb.WriteString(fmt.Sprintf("listen.group = %s\n", pool.ListenGroup))
	sb.WriteString(fmt.Sprintf("listen.mode = %s\n\n", pool.ListenMode))
	
	sb.WriteString("; Process manager settings\n")
	sb.WriteString(fmt.Sprintf("pm = %s\n", pool.PM))
	sb.WriteString(fmt.Sprintf("pm.max_children = %d\n", pool.PMMaxChildren))
	
	if pool.PM == "dynamic" {
		sb.WriteString(fmt.Sprintf("pm.start_servers = %d\n", pool.PMStartServers))
		sb.WriteString(fmt.Sprintf("pm.min_spare_servers = %d\n", pool.PMMinSpareServers))
		sb.WriteString(fmt.Sprintf("pm.max_spare_servers = %d\n", pool.PMMaxSpareServers))
	}
	
	sb.WriteString(fmt.Sprintf("pm.max_requests = %d\n\n", pool.PMMaxRequests))
	
	sb.WriteString("; Additional settings\n")
	sb.WriteString("pm.status_path = /status\n")
	sb.WriteString("ping.path = /ping\n")
	sb.WriteString("ping.response = pong\n")

	return sb.String()
}

// RestartService restarts the PHP-FPM service
func (p *PHPFPMManager) RestartService() error {
	serviceName := fmt.Sprintf("php%s-fpm", p.phpVersion)
	cmd := exec.Command("systemctl", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart PHP-FPM: %s", string(output))
	}
	return nil
}

// ReloadService reloads the PHP-FPM service (graceful reload)
func (p *PHPFPMManager) ReloadService() error {
	serviceName := fmt.Sprintf("php%s-fpm", p.phpVersion)
	cmd := exec.Command("systemctl", "reload", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload PHP-FPM: %s", string(output))
	}
	return nil
}

// GetStatus returns the PHP-FPM service status
func (p *PHPFPMManager) GetStatus() (string, error) {
	serviceName := fmt.Sprintf("php%s-fpm", p.phpVersion)
	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil
	}
	return string(output), nil
}

// IsInstalled checks if PHP-FPM is installed
func (p *PHPFPMManager) IsInstalled() bool {
	serviceName := fmt.Sprintf("php%s-fpm", p.phpVersion)
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.Output()
	return err == nil && strings.Contains(string(output), serviceName)
}

// GetVersion returns the PHP version
func (p *PHPFPMManager) GetVersion() (string, error) {
	cmd := exec.Command("php", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
