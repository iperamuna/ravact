package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/iperamuna/ravact/internal/models"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("/tmp/templates")
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	if manager.templatesDir != "/tmp/templates" {
		t.Errorf("expected templatesDir /tmp/templates, got %s", manager.templatesDir)
	}
}

func TestLoadTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test template
	template := models.ConfigTemplate{
		ID:          "nginx-basic",
		ServiceID:   "nginx",
		Name:        "Basic Nginx Configuration",
		Description: "Basic settings for Nginx",
		FilePath:    "/etc/nginx/nginx.conf",
		Fields: []models.ConfigField{
			{
				Key:      "worker_processes",
				Label:    "Worker Processes",
				Type:     "int",
				Required: true,
				Default:  4,
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

	templatePath := filepath.Join(tmpDir, "nginx.json")
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal template: %v", err)
	}

	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Load template
	manager := NewManager(tmpDir)
	loaded, err := manager.LoadTemplate(templatePath)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}

	if loaded.ID != template.ID {
		t.Errorf("expected ID %s, got %s", template.ID, loaded.ID)
	}
	if loaded.ServiceID != template.ServiceID {
		t.Errorf("expected ServiceID %s, got %s", template.ServiceID, loaded.ServiceID)
	}
	if len(loaded.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(loaded.Fields))
	}
}

func TestSaveTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	template := models.ConfigTemplate{
		ID:        "test-template",
		ServiceID: "test",
		Name:      "Test Template",
		Fields: []models.ConfigField{
			{
				Key:      "test_field",
				Label:    "Test Field",
				Type:     "string",
				Required: false,
				Default:  "test_value",
			},
		},
	}

	templatePath := filepath.Join(tmpDir, "test.json")
	manager := NewManager(tmpDir)

	if err := manager.SaveTemplate(template, templatePath); err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Fatal("template file was not created")
	}

	// Load and verify
	loaded, err := manager.LoadTemplate(templatePath)
	if err != nil {
		t.Fatalf("failed to load saved template: %v", err)
	}

	if loaded.ID != template.ID {
		t.Errorf("expected ID %s, got %s", template.ID, loaded.ID)
	}
}

func TestValidateField(t *testing.T) {
	manager := NewManager("/tmp")

	tests := []struct {
		name        string
		field       models.ConfigField
		value       interface{}
		expectError bool
	}{
		{
			name: "Valid integer",
			field: models.ConfigField{
				Key:      "port",
				Type:     "int",
				Required: true,
			},
			value:       8080,
			expectError: false,
		},
		{
			name: "Valid string",
			field: models.ConfigField{
				Key:      "hostname",
				Type:     "string",
				Required: true,
			},
			value:       "localhost",
			expectError: false,
		},
		{
			name: "Valid boolean",
			field: models.ConfigField{
				Key:      "enabled",
				Type:     "bool",
				Required: true,
			},
			value:       true,
			expectError: false,
		},
		{
			name: "Valid select",
			field: models.ConfigField{
				Key:      "log_level",
				Type:     "select",
				Options:  []string{"debug", "info", "warn", "error"},
				Required: true,
			},
			value:       "info",
			expectError: false,
		},
		{
			name: "Invalid select option",
			field: models.ConfigField{
				Key:      "log_level",
				Type:     "select",
				Options:  []string{"debug", "info", "warn", "error"},
				Required: true,
			},
			value:       "invalid",
			expectError: true,
		},
		{
			name: "Required field missing",
			field: models.ConfigField{
				Key:      "port",
				Type:     "int",
				Required: true,
			},
			value:       nil,
			expectError: true,
		},
		{
			name: "Optional field missing",
			field: models.ConfigField{
				Key:      "port",
				Type:     "int",
				Required: false,
			},
			value:       nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateField(tt.field, tt.value)
			if tt.expectError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	manager := NewManager("/tmp")

	template := models.ConfigTemplate{
		ID:        "test",
		ServiceID: "test",
		Fields: []models.ConfigField{
			{
				Key:      "port",
				Type:     "int",
				Required: true,
			},
			{
				Key:      "hostname",
				Type:     "string",
				Required: true,
			},
			{
				Key:      "debug",
				Type:     "bool",
				Required: false,
				Default:  false,
			},
		},
	}

	tests := []struct {
		name        string
		values      map[string]interface{}
		expectError bool
	}{
		{
			name: "All valid",
			values: map[string]interface{}{
				"port":     8080,
				"hostname": "localhost",
				"debug":    true,
			},
			expectError: false,
		},
		{
			name: "Missing required field",
			values: map[string]interface{}{
				"hostname": "localhost",
			},
			expectError: true,
		},
		{
			name: "Invalid type",
			values: map[string]interface{}{
				"port":     "not-a-number",
				"hostname": "localhost",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := manager.ValidateTemplate(template, tt.values)
			if tt.expectError && len(errors) == 0 {
				t.Error("expected validation errors")
			}
			if !tt.expectError && len(errors) > 0 {
				t.Errorf("unexpected validation errors: %v", errors)
			}
		})
	}
}

func TestApplyDefaults(t *testing.T) {
	manager := NewManager("/tmp")

	template := models.ConfigTemplate{
		ID:        "test",
		ServiceID: "test",
		Fields: []models.ConfigField{
			{
				Key:     "port",
				Type:    "int",
				Default: 8080,
			},
			{
				Key:     "hostname",
				Type:    "string",
				Default: "localhost",
			},
		},
		Defaults: map[string]interface{}{
			"extra_field": "extra_value",
		},
	}

	values := manager.ApplyDefaults(template)

	if values["port"] != 8080 {
		t.Errorf("expected port 8080, got %v", values["port"])
	}
	if values["hostname"] != "localhost" {
		t.Errorf("expected hostname localhost, got %v", values["hostname"])
	}
	if values["extra_field"] != "extra_value" {
		t.Errorf("expected extra_field extra_value, got %v", values["extra_field"])
	}
}

func TestReadWriteConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")

	manager := NewManager(tmpDir)

	// Test write
	content := "# Test configuration\nport=8080\nhostname=localhost\n"
	if err := manager.WriteConfigFile(configPath, content); err != nil {
		t.Fatalf("WriteConfigFile failed: %v", err)
	}

	// Test read
	readContent, err := manager.ReadConfigFile(configPath)
	if err != nil {
		t.Fatalf("ReadConfigFile failed: %v", err)
	}

	if readContent != content {
		t.Errorf("expected content %s, got %s", content, readContent)
	}

	// Test backup creation on overwrite
	newContent := "# Updated configuration\nport=9090\n"
	if err := manager.WriteConfigFile(configPath, newContent); err != nil {
		t.Fatalf("WriteConfigFile (overwrite) failed: %v", err)
	}

	// Verify backup exists
	backupPath := configPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file was not created")
	}

	// Verify backup contains old content
	backupContent, err := manager.ReadConfigFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}
	if backupContent != content {
		t.Errorf("backup content mismatch, expected %s, got %s", content, backupContent)
	}
}

func TestGetAvailableTemplates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test templates
	templates := []models.ConfigTemplate{
		{
			ID:        "nginx",
			ServiceID: "nginx",
			Name:      "Nginx Config",
			Fields:    []models.ConfigField{},
		},
		{
			ID:        "mysql",
			ServiceID: "mysql",
			Name:      "MySQL Config",
			Fields:    []models.ConfigField{},
		},
	}

	for _, tmpl := range templates {
		data, _ := json.Marshal(tmpl)
		filepath := filepath.Join(tmpDir, tmpl.ID+".json")
		os.WriteFile(filepath, data, 0644)
	}

	// Create a non-JSON file
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("Not a template"), 0644)

	manager := NewManager(tmpDir)
	loaded, err := manager.GetAvailableTemplates()
	if err != nil {
		t.Fatalf("GetAvailableTemplates failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 templates, got %d", len(loaded))
	}
}

func TestGetTemplateByService(t *testing.T) {
	tmpDir := t.TempDir()

	template := models.ConfigTemplate{
		ID:        "nginx-basic",
		ServiceID: "nginx",
		Name:      "Nginx Config",
		Fields:    []models.ConfigField{},
	}

	data, _ := json.Marshal(template)
	os.WriteFile(filepath.Join(tmpDir, "nginx.json"), data, 0644)

	manager := NewManager(tmpDir)

	// Test found
	found, err := manager.GetTemplateByService("nginx")
	if err != nil {
		t.Fatalf("GetTemplateByService failed: %v", err)
	}
	if found.ID != template.ID {
		t.Errorf("expected template ID %s, got %s", template.ID, found.ID)
	}

	// Test not found
	_, err = manager.GetTemplateByService("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}
}
