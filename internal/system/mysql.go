package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MySQLConfig represents MySQL configuration
type MySQLConfig struct {
	Port         int
	RootPassword string
	BindAddress  string
	ConfigPath   string
	DataDir      string
	Socket       string
}

// MySQLManager handles MySQL operations
type MySQLManager struct {
	configPath string
}

// NewMySQLManager creates a new MySQL manager
func NewMySQLManager() *MySQLManager {
	return &MySQLManager{
		configPath: "/etc/mysql/mysql.conf.d/mysqld.cnf",
	}
}

// GetConfig reads the current MySQL configuration
func (m *MySQLManager) GetConfig() (*MySQLConfig, error) {
	config := &MySQLConfig{
		Port:        3306,
		BindAddress: "127.0.0.1",
		ConfigPath:  m.configPath,
		DataDir:     "/var/lib/mysql",
		Socket:      "/var/run/mysqld/mysqld.sock",
	}

	// Check if config file exists
	if _, err := os.Stat(m.configPath); err != nil {
		// Try alternative paths
		altPaths := []string{
			"/etc/mysql/my.cnf",
			"/etc/my.cnf",
			"/usr/etc/my.cnf",
		}
		
		found := false
		for _, path := range altPaths {
			if _, err := os.Stat(path); err == nil {
				m.configPath = path
				config.ConfigPath = path
				found = true
				break
			}
		}
		
		if !found {
			return nil, fmt.Errorf("MySQL config file not found")
		}
	}

	// Read and parse config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MySQL config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "port") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%d", &config.Port)
			}
		} else if strings.HasPrefix(line, "bind-address") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				config.BindAddress = parts[2]
			}
		} else if strings.HasPrefix(line, "datadir") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				config.DataDir = parts[2]
			}
		} else if strings.HasPrefix(line, "socket") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				config.Socket = parts[2]
			}
		}
	}

	return config, nil
}

// ChangePort changes the MySQL port
func (m *MySQLManager) ChangePort(newPort int) error {
	if newPort < 1024 || newPort > 65535 {
		return fmt.Errorf("invalid port number: must be between 1024 and 65535")
	}

	// Read current config
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read MySQL config: %w", err)
	}

	// Backup original config
	backupPath := m.configPath + ".bak"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Modify config
	lines := strings.Split(string(data), "\n")
	portFound := false
	inMysqldSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check if we're in [mysqld] section
		if trimmed == "[mysqld]" {
			inMysqldSection = true
			continue
		} else if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			inMysqldSection = false
		}

		// Update port line if in mysqld section
		if inMysqldSection && strings.HasPrefix(trimmed, "port") {
			lines[i] = fmt.Sprintf("port = %d", newPort)
			portFound = true
		}
	}

	// If port not found, add it to mysqld section
	if !portFound {
		for i, line := range lines {
			if strings.TrimSpace(line) == "[mysqld]" {
				// Insert port after [mysqld] line
				newLines := make([]string, 0, len(lines)+1)
				newLines = append(newLines, lines[:i+1]...)
				newLines = append(newLines, fmt.Sprintf("port = %d", newPort))
				newLines = append(newLines, lines[i+1:]...)
				lines = newLines
				break
			}
		}
	}

	// Write updated config
	newData := strings.Join(lines, "\n")
	if err := os.WriteFile(m.configPath, []byte(newData), 0644); err != nil {
		// Restore backup on failure
		os.WriteFile(m.configPath, data, 0644)
		return fmt.Errorf("failed to write MySQL config: %w", err)
	}

	return nil
}

// ChangeRootPassword changes the MySQL root password
func (m *MySQLManager) ChangeRootPassword(newPassword string) error {
	if newPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check if MySQL is running
	cmd := exec.Command("systemctl", "is-active", "mysql")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("MySQL service is not running")
	}

	// Change password using mysql command
	sqlCmd := fmt.Sprintf("ALTER USER 'root'@'localhost' IDENTIFIED BY '%s';", 
		strings.ReplaceAll(newPassword, "'", "\\'"))
	
	cmd = exec.Command("mysql", "-u", "root", "-e", sqlCmd)
	
	// Try with existing password from debian-sys-maint
	debianCnfPath := "/etc/mysql/debian.cnf"
	if _, err := os.Stat(debianCnfPath); err == nil {
		cmd = exec.Command("mysql", "--defaults-file="+debianCnfPath, "-e", sqlCmd)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to change root password: %s", string(output))
	}

	// Flush privileges
	cmd = exec.Command("mysql", "-u", "root", "-p"+newPassword, "-e", "FLUSH PRIVILEGES;")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to flush privileges: %w", err)
	}

	return nil
}

// RestartService restarts the MySQL service
func (m *MySQLManager) RestartService() error {
	cmd := exec.Command("systemctl", "restart", "mysql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart MySQL: %s", string(output))
	}
	return nil
}

// GetStatus returns the MySQL service status
func (m *MySQLManager) GetStatus() (string, error) {
	cmd := exec.Command("systemctl", "status", "mysql")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Status command returns non-zero if service is not running
		return string(output), nil
	}
	return string(output), nil
}

// IsInstalled checks if MySQL is installed
func (m *MySQLManager) IsInstalled() bool {
	cmd := exec.Command("which", "mysql")
	return cmd.Run() == nil
}

// GetVersion returns the MySQL version
func (m *MySQLManager) GetVersion() (string, error) {
	cmd := exec.Command("mysql", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateDatabase creates a new database
func (m *MySQLManager) CreateDatabase(dbName, username, password string) error {
	// Create database
	createDBCmd := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;", dbName)
	cmd := exec.Command("mysql", "-u", "root", "-e", createDBCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Create user and grant privileges if username is provided
	if username != "" {
		createUserCmd := fmt.Sprintf(
			"CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';",
			username, strings.ReplaceAll(password, "'", "\\'"),
		)
		cmd = exec.Command("mysql", "-u", "root", "-e", createUserCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		grantCmd := fmt.Sprintf(
			"GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'localhost'; FLUSH PRIVILEGES;",
			dbName, username,
		)
		cmd = exec.Command("mysql", "-u", "root", "-e", grantCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to grant privileges: %w", err)
		}
	}

	return nil
}

// ListDatabases returns a list of all databases
func (m *MySQLManager) ListDatabases() ([]string, error) {
	cmd := exec.Command("mysql", "-u", "root", "-e", "SHOW DATABASES;")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	databases := make([]string, 0)
	
	for i, line := range lines {
		// Skip header and system databases
		if i == 0 || line == "" {
			continue
		}
		dbName := strings.TrimSpace(line)
		if dbName != "information_schema" && dbName != "performance_schema" && 
		   dbName != "mysql" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// ExportDatabase exports a database to SQL file
func (m *MySQLManager) ExportDatabase(dbName, outputPath string) error {
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

	// Run mysqldump
	cmd := exec.Command("mysqldump", "-u", "root", dbName)
	cmd.Stdout = outFile
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}

	return nil
}
