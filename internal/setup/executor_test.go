package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iperamuna/ravact/internal/models"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor("/tmp/scripts")
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
	if executor.scriptsDir != "/tmp/scripts" {
		t.Errorf("expected scriptsDir /tmp/scripts, got %s", executor.scriptsDir)
	}
}

func TestExecuteScript_Success(t *testing.T) {
	// Create temporary directory for test scripts
	tmpDir := t.TempDir()

	// Create a simple test script
	scriptPath := filepath.Join(tmpDir, "test_success.sh")
	scriptContent := `#!/bin/bash
echo "Test script running"
echo "Installation successful"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "test_success",
		Name:       "Test Success",
		ScriptPath: "test_success.sh",
		Timeout:    5 * time.Second,
	}

	result, err := executor.ExecuteScript(script)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
	if !strings.Contains(result.Output, "Installation successful") {
		t.Errorf("expected output to contain 'Installation successful', got: %s", result.Output)
	}
	if result.Duration <= 0 {
		t.Error("expected duration to be positive")
	}
}

func TestExecuteScript_Failure(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "test_failure.sh")
	scriptContent := `#!/bin/bash
echo "Test script running"
echo "Error occurred" >&2
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "test_failure",
		Name:       "Test Failure",
		ScriptPath: "test_failure.sh",
		Timeout:    5 * time.Second,
	}

	result, err := executor.ExecuteScript(script)
	if err == nil {
		t.Error("expected error for failed script")
	}

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteScript_Timeout(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "test_timeout.sh")
	scriptContent := `#!/bin/bash
echo "Starting long operation"
sleep 10
echo "This should not be reached"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "test_timeout",
		Name:       "Test Timeout",
		ScriptPath: "test_timeout.sh",
		Timeout:    1 * time.Second, // Short timeout
	}

	result, err := executor.ExecuteScript(script)
	if err == nil {
		t.Error("expected timeout error")
	}

	if result.Success {
		t.Error("expected success to be false for timeout")
	}
	if !strings.Contains(result.Error, "timed out") {
		t.Errorf("expected timeout error, got: %s", result.Error)
	}
}

func TestExecuteScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "nonexistent",
		Name:       "Non-existent Script",
		ScriptPath: "nonexistent.sh",
	}

	result, err := executor.ExecuteScript(script)
	if err == nil {
		t.Error("expected error for non-existent script")
	}

	if result.Success {
		t.Error("expected success to be false")
	}
	if !strings.Contains(result.Error, "not found") {
		t.Errorf("expected 'not found' error, got: %s", result.Error)
	}
}

func TestExecuteScript_WithEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	scriptPath := filepath.Join(tmpDir, "test_env.sh")
	scriptContent := `#!/bin/bash
echo "APP_NAME=$APP_NAME"
echo "APP_VERSION=$APP_VERSION"
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "test_env",
		Name:       "Test Environment",
		ScriptPath: "test_env.sh",
		Environment: map[string]string{
			"APP_NAME":    "Ravact",
			"APP_VERSION": "0.1.0",
		},
		Timeout: 5 * time.Second,
	}

	result, err := executor.ExecuteScript(script)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if !strings.Contains(result.Output, "APP_NAME=Ravact") {
		t.Errorf("expected output to contain APP_NAME=Ravact, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "APP_VERSION=0.1.0") {
		t.Errorf("expected output to contain APP_VERSION=0.1.0, got: %s", result.Output)
	}
}

func TestValidateScript(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "Valid bash script",
			content: `#!/bin/bash
echo "Valid script"`,
			expectError: false,
		},
		{
			name: "Valid sh script",
			content: `#!/bin/sh
echo "Valid script"`,
			expectError: false,
		},
		{
			name: "Invalid shebang",
			content: `#!/usr/bin/python
print("Invalid")`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(tmpDir, "test_validate.sh")
			if err := os.WriteFile(scriptPath, []byte(tt.content), 0755); err != nil {
				t.Fatalf("failed to create test script: %v", err)
			}

			executor := NewExecutor(tmpDir)
			script := models.SetupScript{
				ScriptPath: "test_validate.sh",
			}

			err := executor.ValidateScript(script)
			if tt.expectError && err == nil {
				t.Error("expected validation error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}

			// Clean up
			os.Remove(scriptPath)
		})
	}
}

func TestGetAvailableScripts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test scripts
	scripts := []string{"nginx.sh", "mysql.sh", "php.sh"}
	for _, script := range scripts {
		scriptPath := filepath.Join(tmpDir, script)
		content := `#!/bin/bash
echo "Installing ` + script + `"`
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			t.Fatalf("failed to create test script: %v", err)
		}
	}

	// Create a non-script file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("Not a script"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	executor := NewExecutor(tmpDir)
	availableScripts, err := executor.GetAvailableScripts()
	if err != nil {
		t.Fatalf("GetAvailableScripts failed: %v", err)
	}

	if len(availableScripts) != 3 {
		t.Errorf("expected 3 scripts, got %d", len(availableScripts))
	}

	// Verify script IDs
	foundScripts := make(map[string]bool)
	for _, script := range availableScripts {
		foundScripts[script.ID] = true
	}

	if !foundScripts["nginx"] || !foundScripts["mysql"] || !foundScripts["php"] {
		t.Error("expected to find nginx, mysql, and php scripts")
	}
}

func TestGetAvailableScripts_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	executor := NewExecutor(tmpDir)
	scripts, err := executor.GetAvailableScripts()
	if err != nil {
		t.Fatalf("GetAvailableScripts failed: %v", err)
	}

	if len(scripts) != 0 {
		t.Errorf("expected 0 scripts in empty dir, got %d", len(scripts))
	}
}
