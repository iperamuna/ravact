//go:build integration
// +build integration

package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iperamuna/ravact/internal/config"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/setup"
	"github.com/iperamuna/ravact/internal/system"
)

func TestSystemDetection_Integration(t *testing.T) {
	detector := system.NewDetector()
	info, err := detector.GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	t.Logf("OS: %s", info.OS)
	t.Logf("Distribution: %s", info.Distribution)
	t.Logf("Version: %s", info.Version)
	t.Logf("Kernel: %s", info.Kernel)
	t.Logf("Arch: %s", info.Arch)
	t.Logf("CPU Count: %d", info.CPUCount)
	t.Logf("Total RAM: %s", system.FormatBytes(info.TotalRAM))
	t.Logf("Is Root: %v", info.IsRoot)

	// Basic validation
	if info.OS == "" {
		t.Error("OS should not be empty")
	}
	if info.CPUCount <= 0 {
		t.Error("CPU count should be positive")
	}
}

func TestScriptExecution_Integration(t *testing.T) {
	// Create test script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.sh")
	scriptContent := `#!/bin/bash
echo "Integration test script"
echo "System: $(uname -s)"
echo "User: $(whoami)"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := setup.NewExecutor(tmpDir)
	script := models.SetupScript{
		ID:         "test",
		Name:       "Integration Test Script",
		ScriptPath: "test.sh",
	}

	result, err := executor.ExecuteScript(script)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}

	t.Logf("Script output:\n%s", result.Output)
	t.Logf("Duration: %v", result.Duration)

	if !result.Success {
		t.Error("expected script to succeed")
	}
}

func TestConfigManagement_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config file
	configPath := filepath.Join(tmpDir, "test.conf")
	originalContent := `# Test Configuration
port=8080
hostname=localhost
debug=false
`
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	manager := config.NewManager(tmpDir)

	// Read config
	content, err := manager.ReadConfigFile(configPath)
	if err != nil {
		t.Fatalf("ReadConfigFile failed: %v", err)
	}

	if content != originalContent {
		t.Errorf("content mismatch")
	}

	// Write new config (should create backup)
	newContent := `# Test Configuration (Updated)
port=9090
hostname=example.com
debug=true
`
	if err := manager.WriteConfigFile(configPath, newContent); err != nil {
		t.Fatalf("WriteConfigFile failed: %v", err)
	}

	// Verify backup was created
	backupPath := configPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file was not created")
	}

	// Verify backup contains original content
	backupContent, err := manager.ReadConfigFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}

	if backupContent != originalContent {
		t.Error("backup content does not match original")
	}

	// Verify new content was written
	updatedContent, err := manager.ReadConfigFile(configPath)
	if err != nil {
		t.Fatalf("failed to read updated config: %v", err)
	}

	if updatedContent != newContent {
		t.Error("updated content does not match expected")
	}
}

func TestServiceDetection_Integration(t *testing.T) {
	// Skip if not on Linux
	detector := system.NewDetector()
	info, _ := detector.GetSystemInfo()
	
	if info.OS != "linux" {
		t.Skip("Skipping service detection test on non-Linux system")
	}

	// Test common commands that should exist
	commonCommands := []string{"bash", "sh", "ls", "cat"}

	for _, cmd := range commonCommands {
		installed, err := detector.IsServiceInstalled(cmd)
		if err != nil {
			t.Logf("Error checking %s: %v", cmd, err)
		}
		t.Logf("Command '%s' installed: %v", cmd, installed)
	}
}

func TestRecommendations_Integration(t *testing.T) {
	detector := system.NewDetector()
	info, err := detector.GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	workerProcesses := detector.GetRecommendedWorkerProcesses()
	t.Logf("Recommended worker processes: %d", workerProcesses)

	if workerProcesses != info.CPUCount {
		t.Errorf("expected worker processes to equal CPU count (%d), got %d", info.CPUCount, workerProcesses)
	}

	if info.TotalRAM > 0 {
		workerConnections := detector.GetRecommendedWorkerConnections(info.TotalRAM)
		t.Logf("Recommended worker connections: %d (for %s RAM)", workerConnections, system.FormatBytes(info.TotalRAM))

		if workerConnections < 1024 || workerConnections > 4096 {
			t.Errorf("worker connections should be between 1024 and 4096, got %d", workerConnections)
		}
	}
}

func TestEndToEnd_Integration(t *testing.T) {
	t.Log("Running end-to-end integration test")

	// 1. Detect system
	detector := system.NewDetector()
	sysInfo, err := detector.GetSystemInfo()
	if err != nil {
		t.Fatalf("System detection failed: %v", err)
	}
	t.Logf("✓ System detected: %s %s", sysInfo.Distribution, sysInfo.Version)

	// 2. Create test environment
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	configsDir := filepath.Join(tmpDir, "configs")
	
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}
	if err := os.MkdirAll(configsDir, 0755); err != nil {
		t.Fatalf("failed to create configs dir: %v", err)
	}

	// 3. Test script execution
	scriptPath := filepath.Join(scriptsDir, "test_install.sh")
	scriptContent := `#!/bin/bash
echo "Starting installation..."
sleep 1
echo "Installation complete!"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	executor := setup.NewExecutor(scriptsDir)
	availableScripts, err := executor.GetAvailableScripts()
	if err != nil {
		t.Fatalf("failed to get available scripts: %v", err)
	}
	t.Logf("✓ Found %d available scripts", len(availableScripts))

	if len(availableScripts) > 0 {
		result, err := executor.ExecuteScript(availableScripts[0])
		if err != nil {
			t.Fatalf("script execution failed: %v", err)
		}
		t.Logf("✓ Script executed successfully in %v", result.Duration)
	}

	// 4. Test configuration management
	configManager := config.NewManager(configsDir)
	
	template := models.ConfigTemplate{
		ID:        "test-config",
		ServiceID: "test",
		Name:      "Test Configuration",
		Fields: []models.ConfigField{
			{
				Key:      "port",
				Type:     "int",
				Required: true,
				Default:  8080,
			},
		},
	}

	templatePath := filepath.Join(configsDir, "test.json")
	if err := configManager.SaveTemplate(template, templatePath); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}
	t.Logf("✓ Configuration template saved")

	loadedTemplate, err := configManager.LoadTemplate(templatePath)
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}
	t.Logf("✓ Configuration template loaded: %s", loadedTemplate.Name)

	// 5. Test validation
	values := map[string]interface{}{
		"port": 9090,
	}
	errors := configManager.ValidateTemplate(*loadedTemplate, values)
	if len(errors) > 0 {
		t.Fatalf("validation failed: %v", errors)
	}
	t.Logf("✓ Configuration validation passed")

	t.Log("✓ End-to-end integration test completed successfully")
}
