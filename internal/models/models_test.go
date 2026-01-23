package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestServiceTypes(t *testing.T) {
	tests := []struct {
		name     string
		service  ServiceType
		expected string
	}{
		{"Web service", ServiceTypeWeb, "web"},
		{"Database service", ServiceTypeDatabase, "database"},
		{"Cache service", ServiceTypeCache, "cache"},
		{"Queue service", ServiceTypeQueue, "queue"},
		{"Monitor service", ServiceTypeMonitor, "monitor"},
		{"Other service", ServiceTypeOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.service) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.service))
			}
		})
	}
}

func TestServiceStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   ServiceStatus
		expected string
	}{
		{"Not installed", StatusNotInstalled, "not_installed"},
		{"Installed", StatusInstalled, "installed"},
		{"Running", StatusRunning, "running"},
		{"Stopped", StatusStopped, "stopped"},
		{"Failed", StatusFailed, "failed"},
		{"Unknown", StatusUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestServiceJSON(t *testing.T) {
	service := Service{
		ID:          "nginx",
		Name:        "Nginx",
		Description: "Web server",
		Type:        ServiceTypeWeb,
		Status:      StatusRunning,
		Version:     "1.24.0",
		Port:        80,
		ConfigPath:  "/etc/nginx/nginx.conf",
	}

	// Marshal to JSON
	data, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("failed to marshal service: %v", err)
	}

	// Unmarshal back
	var decoded Service
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal service: %v", err)
	}

	// Verify fields
	if decoded.ID != service.ID {
		t.Errorf("expected ID %s, got %s", service.ID, decoded.ID)
	}
	if decoded.Status != service.Status {
		t.Errorf("expected status %s, got %s", service.Status, decoded.Status)
	}
	if decoded.Port != service.Port {
		t.Errorf("expected port %d, got %d", service.Port, decoded.Port)
	}
}

func TestExecutionResult(t *testing.T) {
	result := ExecutionResult{
		Success:   true,
		Output:    "Command executed successfully",
		ExitCode:  0,
		Duration:  5 * time.Second,
		Timestamp: time.Now(),
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Duration != 5*time.Second {
		t.Errorf("expected duration 5s, got %v", result.Duration)
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "port",
		Message: "port must be between 1 and 65535",
	}

	expected := "port: port must be between 1 and 65535"
	if err.Error() != expected {
		t.Errorf("expected %s, got %s", expected, err.Error())
	}
}

func TestConfigTemplate(t *testing.T) {
	template := ConfigTemplate{
		ID:          "nginx-basic",
		ServiceID:   "nginx",
		Name:        "Basic Configuration",
		Description: "Basic Nginx configuration",
		FilePath:    "/etc/nginx/nginx.conf",
		Fields: []ConfigField{
			{
				Key:      "worker_processes",
				Label:    "Worker Processes",
				Type:     "int",
				Required: true,
				Default:  "auto",
			},
			{
				Key:      "worker_connections",
				Label:    "Worker Connections",
				Type:     "int",
				Required: true,
				Default:  1024,
			},
		},
	}

	if len(template.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(template.Fields))
	}

	// Test field types
	if template.Fields[0].Type != "int" {
		t.Errorf("expected type 'int', got %s", template.Fields[0].Type)
	}
}

func TestSystemInfo(t *testing.T) {
	info := SystemInfo{
		OS:           "linux",
		Distribution: "ubuntu",
		Version:      "24.04",
		Kernel:       "6.8.0",
		Arch:         "amd64",
		Hostname:     "test-server",
		CPUCount:     4,
		TotalRAM:     8589934592, // 8GB
		IsRoot:       true,
	}

	if info.OS != "linux" {
		t.Errorf("expected OS 'linux', got %s", info.OS)
	}
	if info.CPUCount != 4 {
		t.Errorf("expected 4 CPUs, got %d", info.CPUCount)
	}
	if !info.IsRoot {
		t.Error("expected IsRoot to be true")
	}
}
