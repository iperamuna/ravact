package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SupervisorProgram represents a Supervisor program configuration
type SupervisorProgram struct {
	Name       string
	ConfigPath string
	IsEnabled  bool
	State      string // RUNNING, STOPPED, etc.
	Command    string
	Directory  string
	User       string
	AutoStart  bool
}

// SupervisorXMLRPCConfig represents XML-RPC server configuration
type SupervisorXMLRPCConfig struct {
	Enabled  bool
	IP       string
	Port     string
	Username string
	Password string
}

// SupervisorManager handles Supervisor configuration operations
type SupervisorManager struct {
	programsDir string
	configPath  string
}

// NewSupervisorManager creates a new Supervisor manager
func NewSupervisorManager() *SupervisorManager {
	// Try common Supervisor config paths
	configPaths := []string{
		"/etc/supervisor/supervisord.conf",
		"/etc/supervisord.conf",
	}
	
	configPath := "/etc/supervisor/supervisord.conf" // Default
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}
	
	return &SupervisorManager{
		programsDir: "/etc/supervisor/conf.d",
		configPath:  configPath,
	}
}

// GetAllPrograms returns all Supervisor programs
func (sm *SupervisorManager) GetAllPrograms() ([]SupervisorProgram, error) {
	entries, err := os.ReadDir(sm.programsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SupervisorProgram{}, nil
		}
		return nil, err
	}

	var programs []SupervisorProgram
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only process .conf files
		if !strings.HasSuffix(name, ".conf") {
			continue
		}

		programName := strings.TrimSuffix(name, ".conf")
		configPath := filepath.Join(sm.programsDir, name)
		
		// Parse config to get details
		command, directory, user, autostart := sm.parseConfig(configPath)
		
		// Get state from supervisorctl
		state := sm.getProgramState(programName)

		program := SupervisorProgram{
			Name:       programName,
			ConfigPath: configPath,
			IsEnabled:  true, // If file exists, it's enabled
			State:      state,
			Command:    command,
			Directory:  directory,
			User:       user,
			AutoStart:  autostart,
		}

		programs = append(programs, program)
	}

	return programs, nil
}

// parseConfig extracts basic info from supervisor config
func (sm *SupervisorManager) parseConfig(configPath string) (command, directory, user string, autostart bool) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", "", false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "command=") {
			command = strings.TrimPrefix(line, "command=")
		} else if strings.HasPrefix(line, "directory=") {
			directory = strings.TrimPrefix(line, "directory=")
		} else if strings.HasPrefix(line, "user=") {
			user = strings.TrimPrefix(line, "user=")
		} else if strings.HasPrefix(line, "autostart=") {
			autostart = strings.TrimPrefix(line, "autostart=") == "true"
		}
	}

	return command, directory, user, autostart
}

// getProgramState gets the state of a program from supervisorctl
func (sm *SupervisorManager) getProgramState(programName string) string {
	cmd := exec.Command("supervisorctl", "status", programName)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return "UNKNOWN"
	}
	
	// Parse output like: "program_name RUNNING pid 12345, uptime 0:01:23"
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return parts[1]
	}
	
	return "UNKNOWN"
}

// StartProgram starts a supervisor program
func (sm *SupervisorManager) StartProgram(programName string) error {
	cmd := exec.Command("supervisorctl", "start", programName)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to start: %s", string(output))
	}
	
	return nil
}

// StopProgram stops a supervisor program
func (sm *SupervisorManager) StopProgram(programName string) error {
	cmd := exec.Command("supervisorctl", "stop", programName)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to stop: %s", string(output))
	}
	
	return nil
}

// RestartProgram restarts a supervisor program
func (sm *SupervisorManager) RestartProgram(programName string) error {
	cmd := exec.Command("supervisorctl", "restart", programName)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to restart: %s", string(output))
	}
	
	return nil
}

// DeleteProgram deletes a program configuration
func (sm *SupervisorManager) DeleteProgram(programName string) error {
	// Stop first if running
	_ = sm.StopProgram(programName)
	
	// Delete config file
	configPath := filepath.Join(sm.programsDir, programName+".conf")
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	
	// Reload supervisor
	return sm.Reread()
}

// CreateProgram creates a new program configuration
func (sm *SupervisorManager) CreateProgram(name, command, directory, user string, autostart bool) error {
	configPath := filepath.Join(sm.programsDir, name+".conf")

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("program already exists: %s", name)
	}

	// Generate config
	config := fmt.Sprintf(`[program:%s]
command=%s
directory=%s
user=%s
autostart=%t
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/supervisor/%s.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=10
`, name, command, directory, user, autostart, name)

	// Write config file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Reload supervisor
	return sm.Reread()
}

// Reread tells supervisor to reload configuration
func (sm *SupervisorManager) Reread() error {
	cmd := exec.Command("supervisorctl", "reread")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to reread: %s", string(output))
	}
	
	// Update
	cmd = exec.Command("supervisorctl", "update")
	output, err = cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	
	return nil
}

// GetXMLRPCConfig gets the XML-RPC server configuration
func (sm *SupervisorManager) GetXMLRPCConfig() (*SupervisorXMLRPCConfig, error) {
	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := &SupervisorXMLRPCConfig{
		Enabled: false,
		IP:      "127.0.0.1",
		Port:    "9001",
	}

	lines := strings.Split(string(data), "\n")
	inInetSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "[inet_http_server]") {
			inInetSection = true
			config.Enabled = true
			continue
		}
		
		if strings.HasPrefix(line, "[") {
			inInetSection = false
		}
		
		if !inInetSection {
			continue
		}
		
		if strings.HasPrefix(line, "port=") {
			portStr := strings.TrimPrefix(line, "port=")
			parts := strings.Split(portStr, ":")
			if len(parts) == 2 {
				config.IP = parts[0]
				config.Port = parts[1]
			}
		} else if strings.HasPrefix(line, "username=") {
			config.Username = strings.TrimPrefix(line, "username=")
		} else if strings.HasPrefix(line, "password=") {
			config.Password = strings.TrimPrefix(line, "password=")
		}
	}

	return config, nil
}

// SetXMLRPCConfig configures the XML-RPC server
func (sm *SupervisorManager) SetXMLRPCConfig(ip, port, username, password string) error {
	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inInetSection := false
	sectionFound := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "[inet_http_server]") {
			inInetSection = true
			sectionFound = true
			newLines = append(newLines, line)
			continue
		}
		
		if strings.HasPrefix(trimmed, "[") && inInetSection {
			inInetSection = false
		}
		
		// Skip old inet_http_server config lines
		if inInetSection {
			if strings.HasPrefix(trimmed, "port=") || 
			   strings.HasPrefix(trimmed, "username=") || 
			   strings.HasPrefix(trimmed, "password=") {
				continue
			}
		}
		
		newLines = append(newLines, line)
	}

	// Add inet_http_server section if not found
	if !sectionFound {
		newLines = append(newLines, "", "[inet_http_server]")
	}

	// Find and update/add inet section
	for i, line := range newLines {
		if strings.TrimSpace(line) == "[inet_http_server]" {
			// Insert config after section header
			config := []string{
				fmt.Sprintf("port=%s:%s", ip, port),
				fmt.Sprintf("username=%s", username),
				fmt.Sprintf("password=%s", password),
			}
			newLines = append(newLines[:i+1], append(config, newLines[i+1:]...)...)
			break
		}
	}

	// Write back
	newConfig := strings.Join(newLines, "\n")
	if err := os.WriteFile(sm.configPath, []byte(newConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Restart supervisor
	return sm.RestartSupervisor()
}

// RestartSupervisor restarts the supervisor service
func (sm *SupervisorManager) RestartSupervisor() error {
	cmd := exec.Command("systemctl", "restart", "supervisor")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to restart: %s", string(output))
	}
	
	return nil
}
