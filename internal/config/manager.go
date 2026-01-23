package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iperamuna/ravact/internal/models"
)

// Manager handles configuration templates and updates
type Manager struct {
	templatesDir string
}

// NewManager creates a new configuration manager
func NewManager(templatesDir string) *Manager {
	return &Manager{
		templatesDir: templatesDir,
	}
}

// GetAvailableTemplates returns all available configuration templates
func (m *Manager) GetAvailableTemplates() ([]models.ConfigTemplate, error) {
	entries, err := os.ReadDir(m.templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.ConfigTemplate{}, nil
		}
		return nil, err
	}

	var templates []models.ConfigTemplate
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		templatePath := filepath.Join(m.templatesDir, entry.Name())
		template, err := m.LoadTemplate(templatePath)
		if err != nil {
			continue // Skip invalid templates
		}

		templates = append(templates, *template)
	}

	return templates, nil
}

// LoadTemplate loads a configuration template from a JSON file
func (m *Manager) LoadTemplate(path string) (*models.ConfigTemplate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var template models.ConfigTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// SaveTemplate saves a configuration template to a JSON file
func (m *Manager) SaveTemplate(template models.ConfigTemplate, path string) error {
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ReadConfigFile reads a configuration file
func (m *Manager) ReadConfigFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteConfigFile writes content to a configuration file
func (m *Manager) WriteConfigFile(path string, content string) error {
	// Create backup
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".backup"
		if err := m.createBackup(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write new content
	return os.WriteFile(path, []byte(content), 0644)
}

// createBackup creates a backup of a file
func (m *Manager) createBackup(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// ValidateField validates a configuration field value
func (m *Manager) ValidateField(field models.ConfigField, value interface{}) error {
	// Check required
	if field.Required && value == nil {
		return models.ValidationError{
			Field:   field.Key,
			Message: "field is required",
		}
	}

	// Skip validation if value is nil and field is not required
	if !field.Required && value == nil {
		return nil
	}

	// Type validation
	switch field.Type {
	case "int":
		if _, ok := value.(int); !ok {
			if _, ok := value.(float64); !ok { // JSON numbers are float64
				return models.ValidationError{
					Field:   field.Key,
					Message: "must be an integer",
				}
			}
		}

	case "bool":
		if _, ok := value.(bool); !ok {
			return models.ValidationError{
				Field:   field.Key,
				Message: "must be a boolean",
			}
		}

	case "string":
		if _, ok := value.(string); !ok {
			return models.ValidationError{
				Field:   field.Key,
				Message: "must be a string",
			}
		}

	case "select":
		strValue, ok := value.(string)
		if !ok {
			return models.ValidationError{
				Field:   field.Key,
				Message: "must be a string",
			}
		}

		// Check if value is in options
		found := false
		for _, option := range field.Options {
			if option == strValue {
				found = true
				break
			}
		}
		if !found {
			return models.ValidationError{
				Field:   field.Key,
				Message: fmt.Sprintf("must be one of: %v", field.Options),
			}
		}
	}

	return nil
}

// ValidateTemplate validates all fields in a template with given values
func (m *Manager) ValidateTemplate(template models.ConfigTemplate, values map[string]interface{}) []error {
	var errors []error

	for _, field := range template.Fields {
		value, exists := values[field.Key]
		if !exists {
			value = field.Default
		}

		if err := m.ValidateField(field, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ApplyDefaults applies default values to a values map
func (m *Manager) ApplyDefaults(template models.ConfigTemplate) map[string]interface{} {
	values := make(map[string]interface{})

	for _, field := range template.Fields {
		if field.Default != nil {
			values[field.Key] = field.Default
		}
	}

	// Merge with template defaults
	for key, value := range template.Defaults {
		if _, exists := values[key]; !exists {
			values[key] = value
		}
	}

	return values
}

// GetTemplateByService returns a template for a specific service
func (m *Manager) GetTemplateByService(serviceID string) (*models.ConfigTemplate, error) {
	templates, err := m.GetAvailableTemplates()
	if err != nil {
		return nil, err
	}

	for _, template := range templates {
		if template.ServiceID == serviceID {
			return &template, nil
		}
	}

	return nil, fmt.Errorf("no template found for service: %s", serviceID)
}
