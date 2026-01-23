package models

import "time"

// ServiceType represents the type of service
type ServiceType string

const (
	ServiceTypeWeb      ServiceType = "web"
	ServiceTypeDatabase ServiceType = "database"
	ServiceTypeCache    ServiceType = "cache"
	ServiceTypeQueue    ServiceType = "queue"
	ServiceTypeMonitor  ServiceType = "monitor"
	ServiceTypeOther    ServiceType = "other"
)

// ServiceStatus represents the current status of a service
type ServiceStatus string

const (
	StatusNotInstalled ServiceStatus = "not_installed"
	StatusInstalled    ServiceStatus = "installed"
	StatusRunning      ServiceStatus = "running"
	StatusStopped      ServiceStatus = "stopped"
	StatusFailed       ServiceStatus = "failed"
	StatusUnknown      ServiceStatus = "unknown"
)

// Service represents an installable/manageable service
type Service struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Type        ServiceType   `json:"type"`
	Status      ServiceStatus `json:"status"`
	Version     string        `json:"version,omitempty"`
	Port        int           `json:"port,omitempty"`
	ConfigPath  string        `json:"config_path,omitempty"`
	ScriptPath  string        `json:"script_path,omitempty"`
}

// SetupScript represents an installation script
type SetupScript struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	ScriptPath   string            `json:"script_path"`
	ServiceID    string            `json:"service_id"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	ID          string                 `json:"id"`
	ServiceID   string                 `json:"service_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	FilePath    string                 `json:"file_path"`
	Fields      []ConfigField          `json:"fields"`
	Defaults    map[string]interface{} `json:"defaults,omitempty"`
}

// ConfigField represents a configurable field
type ConfigField struct {
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Type        string      `json:"type"` // string, int, bool, select
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"` // For select type
	Validation  string      `json:"validation,omitempty"`
	Description string      `json:"description,omitempty"`
}

// QuickCommand represents a quick action command
type QuickCommand struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	RequireRoot bool     `json:"require_root"`
	Confirm     bool     `json:"confirm"` // Show confirmation dialog
}

// SystemInfo represents system information
type SystemInfo struct {
	OS           string  `json:"os"`
	Distribution string  `json:"distribution"`
	Version      string  `json:"version"`
	Kernel       string  `json:"kernel"`
	Arch         string  `json:"arch"`
	Hostname     string  `json:"hostname"`
	CPUCount     int     `json:"cpu_count"`
	TotalRAM     uint64  `json:"total_ram"` // in bytes
	TotalDisk    uint64  `json:"total_disk"`
	IsRoot       bool    `json:"is_root"`
}

// ExecutionResult represents the result of a command/script execution
type ExecutionResult struct {
	Success   bool          `json:"success"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return v.Field + ": " + v.Message
}
