package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PostgreSQLConfig represents PostgreSQL configuration
type PostgreSQLConfig struct {
	Port        int
	MaxConn     int
	SharedBuf   string
	ConfigPath  string
	DataDir     string
	HBAPath     string
	LogDir      string
}

// PostgreSQLManager handles PostgreSQL operations
type PostgreSQLManager struct {
	configPath string
	hbaPath    string
	version    string
}

// NewPostgreSQLManager creates a new PostgreSQL manager
func NewPostgreSQLManager() *PostgreSQLManager {
	return &PostgreSQLManager{
		configPath: "/etc/postgresql/*/main/postgresql.conf",
		hbaPath:    "/etc/postgresql/*/main/pg_hba.conf",
	}
}

// detectConfigPath finds the actual PostgreSQL config path
func (p *PostgreSQLManager) detectConfigPath() error {
	// Try to find the actual config path
	cmd := exec.Command("bash", "-c", "ls /etc/postgresql/*/main/postgresql.conf 2>/dev/null | head -1")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("PostgreSQL config file not found")
	}
	
	p.configPath = strings.TrimSpace(string(output))
	
	// Also set HBA path
	dir := filepath.Dir(p.configPath)
	p.hbaPath = filepath.Join(dir, "pg_hba.conf")
	
	// Extract version from path
	parts := strings.Split(p.configPath, "/")
	if len(parts) > 3 {
		p.version = parts[3]
	}
	
	return nil
}

// GetConfig reads the current PostgreSQL configuration
func (p *PostgreSQLManager) GetConfig() (*PostgreSQLConfig, error) {
	if err := p.detectConfigPath(); err != nil {
		return nil, err
	}

	config := &PostgreSQLConfig{
		Port:       5432,
		MaxConn:    100,
		SharedBuf:  "128MB",
		ConfigPath: p.configPath,
		HBAPath:    p.hbaPath,
		DataDir:    filepath.Dir(p.configPath),
		LogDir:     "/var/log/postgresql",
	}

	// Read and parse config file
	data, err := os.ReadFile(p.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PostgreSQL config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove comments from value
		if idx := strings.Index(value, "#"); idx >= 0 {
			value = strings.TrimSpace(value[:idx])
		}
		
		// Remove quotes
		value = strings.Trim(value, "'\"")

		switch key {
		case "port":
			fmt.Sscanf(value, "%d", &config.Port)
		case "max_connections":
			fmt.Sscanf(value, "%d", &config.MaxConn)
		case "shared_buffers":
			config.SharedBuf = value
		case "data_directory":
			config.DataDir = value
		case "log_directory":
			config.LogDir = value
		}
	}

	return config, nil
}

// ChangePort changes the PostgreSQL port
func (p *PostgreSQLManager) ChangePort(newPort int) error {
	if newPort < 1024 || newPort > 65535 {
		return fmt.Errorf("invalid port number: must be between 1024 and 65535")
	}

	if err := p.detectConfigPath(); err != nil {
		return err
	}

	// Read current config
	data, err := os.ReadFile(p.configPath)
	if err != nil {
		return fmt.Errorf("failed to read PostgreSQL config: %w", err)
	}

	// Backup original config
	backupPath := p.configPath + ".bak"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Modify config
	lines := strings.Split(string(data), "\n")
	portFound := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			// Check if it's a commented port line
			if strings.Contains(trimmed, "port =") || strings.Contains(trimmed, "port=") {
				lines[i] = fmt.Sprintf("port = %d", newPort)
				portFound = true
			}
			continue
		}

		// Update port line
		if strings.HasPrefix(trimmed, "port") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				lines[i] = fmt.Sprintf("port = %d", newPort)
				portFound = true
			}
		}
	}

	// If port not found, add it
	if !portFound {
		lines = append([]string{fmt.Sprintf("port = %d", newPort)}, lines...)
	}

	// Write updated config
	newData := strings.Join(lines, "\n")
	if err := os.WriteFile(p.configPath, []byte(newData), 0644); err != nil {
		// Restore backup on failure
		os.WriteFile(p.configPath, data, 0644)
		return fmt.Errorf("failed to write PostgreSQL config: %w", err)
	}

	return nil
}

// ChangeRootPassword changes the PostgreSQL postgres user password
func (p *PostgreSQLManager) ChangeRootPassword(newPassword string) error {
	if newPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Escape single quotes in password
	escapedPassword := strings.ReplaceAll(newPassword, "'", "''")

	// Change password using psql as postgres user
	sqlCmd := fmt.Sprintf("ALTER USER postgres WITH PASSWORD '%s';", escapedPassword)
	
	cmd := exec.Command("sudo", "-u", "postgres", "psql", "-c", sqlCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to change postgres password: %s", string(output))
	}

	return nil
}

// RestartService restarts the PostgreSQL service
func (p *PostgreSQLManager) RestartService() error {
	cmd := exec.Command("systemctl", "restart", "postgresql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart PostgreSQL: %s", string(output))
	}
	return nil
}

// GetStatus returns the PostgreSQL service status
func (p *PostgreSQLManager) GetStatus() (string, error) {
	cmd := exec.Command("systemctl", "status", "postgresql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Status command returns non-zero if service is not running
		return string(output), nil
	}
	return string(output), nil
}

// IsInstalled checks if PostgreSQL is installed
func (p *PostgreSQLManager) IsInstalled() bool {
	cmd := exec.Command("which", "psql")
	return cmd.Run() == nil
}

// GetVersion returns the PostgreSQL version
func (p *PostgreSQLManager) GetVersion() (string, error) {
	cmd := exec.Command("psql", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateDatabase creates a new PostgreSQL database
func (p *PostgreSQLManager) CreateDatabase(dbName, username, password string) error {
	// Create database
	createDBCmd := fmt.Sprintf("CREATE DATABASE \"%s\";", dbName)
	cmd := exec.Command("sudo", "-u", "postgres", "psql", "-c", createDBCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if database already exists
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("failed to create database: %s", string(output))
		}
	}

	// Create user and grant privileges if username is provided
	if username != "" {
		escapedPassword := strings.ReplaceAll(password, "'", "''")
		
		// Create user
		createUserCmd := fmt.Sprintf(
			"CREATE USER \"%s\" WITH PASSWORD '%s';",
			username, escapedPassword,
		)
		cmd = exec.Command("sudo", "-u", "postgres", "psql", "-c", createUserCmd)
		output, err = cmd.CombinedOutput()
		if err != nil {
			if !strings.Contains(string(output), "already exists") {
				return fmt.Errorf("failed to create user: %s", string(output))
			}
		}

		// Grant privileges
		grantCmd := fmt.Sprintf(
			"GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\";",
			dbName, username,
		)
		cmd = exec.Command("sudo", "-u", "postgres", "psql", "-c", grantCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to grant privileges: %w", err)
		}
	}

	return nil
}

// ListDatabases returns a list of all databases
func (p *PostgreSQLManager) ListDatabases() ([]string, error) {
	cmd := exec.Command("sudo", "-u", "postgres", "psql", "-t", "-c", "SELECT datname FROM pg_database WHERE datistemplate = false;")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	databases := make([]string, 0)
	
	for _, line := range lines {
		dbName := strings.TrimSpace(line)
		if dbName != "" && dbName != "postgres" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// ExportDatabase exports a database to SQL file
func (p *PostgreSQLManager) ExportDatabase(dbName, outputPath string) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Run pg_dump
	cmd := exec.Command("sudo", "-u", "postgres", "pg_dump", dbName)
	cmd.Stdout = outFile
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}

	return nil
}

// UpdateMaxConnections updates the max_connections setting
func (p *PostgreSQLManager) UpdateMaxConnections(maxConn int) error {
	if maxConn < 10 || maxConn > 10000 {
		return fmt.Errorf("invalid max_connections: must be between 10 and 10000")
	}

	if err := p.detectConfigPath(); err != nil {
		return err
	}

	return p.updateConfigValue("max_connections", fmt.Sprintf("%d", maxConn))
}

// UpdateSharedBuffers updates the shared_buffers setting
func (p *PostgreSQLManager) UpdateSharedBuffers(sharedBuf string) error {
	if err := p.detectConfigPath(); err != nil {
		return err
	}

	return p.updateConfigValue("shared_buffers", sharedBuf)
}

// updateConfigValue is a helper to update a config value
func (p *PostgreSQLManager) updateConfigValue(key, value string) error {
	// Read current config
	data, err := os.ReadFile(p.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Backup
	backupPath := p.configPath + ".bak"
	os.WriteFile(backupPath, data, 0644)

	// Modify config
	lines := strings.Split(string(data), "\n")
	found := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Handle commented lines
		if strings.HasPrefix(trimmed, "#") {
			trimmed = strings.TrimPrefix(trimmed, "#")
			trimmed = strings.TrimSpace(trimmed)
		}

		if strings.HasPrefix(trimmed, key) {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				lines[i] = fmt.Sprintf("%s = %s", key, value)
				found = true
				break
			}
		}
	}

	// If not found, add it
	if !found {
		lines = append([]string{fmt.Sprintf("%s = %s", key, value)}, lines...)
	}

	// Write updated config
	newData := strings.Join(lines, "\n")
	if err := os.WriteFile(p.configPath, []byte(newData), 0644); err != nil {
		os.WriteFile(p.configPath, data, 0644)
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
