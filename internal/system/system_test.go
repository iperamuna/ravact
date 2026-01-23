package system

import (
	"runtime"
	"testing"

	"github.com/iperamuna/ravact/internal/models"
)

func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	if detector == nil {
		t.Fatal("expected non-nil detector")
	}
}

func TestGetSystemInfo(t *testing.T) {
	detector := NewDetector()
	info, err := detector.GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	// Verify basic fields
	if info.OS == "" {
		t.Error("expected OS to be set")
	}
	if info.Arch == "" {
		t.Error("expected Arch to be set")
	}
	if info.CPUCount <= 0 {
		t.Error("expected CPUCount to be positive")
	}
	if info.Hostname == "" {
		t.Error("expected Hostname to be set")
	}

	// Verify OS matches runtime
	if info.OS != runtime.GOOS {
		t.Errorf("expected OS %s, got %s", runtime.GOOS, info.OS)
	}
	if info.Arch != runtime.GOARCH {
		t.Errorf("expected Arch %s, got %s", runtime.GOARCH, info.Arch)
	}
	if info.CPUCount != runtime.NumCPU() {
		t.Errorf("expected CPUCount %d, got %d", runtime.NumCPU(), info.CPUCount)
	}

	// Linux-specific checks
	if runtime.GOOS == "linux" {
		if info.Distribution == "" {
			t.Log("Warning: Distribution not detected (may be expected on non-Linux)")
		}
		if info.Kernel == "" {
			t.Log("Warning: Kernel version not detected")
		}
		if info.TotalRAM == 0 {
			t.Log("Warning: Total RAM not detected")
		}
	}
}

func TestIsRoot(t *testing.T) {
	detector := NewDetector()
	isRoot := detector.IsRoot()

	// Just verify it returns a boolean without error
	t.Logf("Running as root: %v", isRoot)

	// We can't reliably test this as it depends on execution context
	// But we can verify the function works
	if isRoot {
		t.Log("Running with root privileges")
	} else {
		t.Log("Running without root privileges")
	}
}

func TestParseOSRelease(t *testing.T) {
	detector := NewDetector()
	info := &models.SystemInfo{}

	testData := []byte(`NAME="Ubuntu"
VERSION="24.04 LTS (Noble Numbat)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 24.04 LTS"
VERSION_ID="24.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"`)

	err := detector.parseOSRelease(testData, info)
	if err != nil {
		t.Fatalf("parseOSRelease failed: %v", err)
	}

	if info.Distribution != "ubuntu" {
		t.Errorf("expected distribution 'ubuntu', got '%s'", info.Distribution)
	}
	if info.Version != "24.04" {
		t.Errorf("expected version '24.04', got '%s'", info.Version)
	}
}

func TestParseLSBRelease(t *testing.T) {
	detector := NewDetector()
	info := &models.SystemInfo{}

	testData := []byte(`DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=24.04
DISTRIB_CODENAME=noble
DISTRIB_DESCRIPTION="Ubuntu 24.04 LTS"`)

	err := detector.parseLSBRelease(testData, info)
	if err != nil {
		t.Fatalf("parseLSBRelease failed: %v", err)
	}

	if info.Distribution != "ubuntu" {
		t.Errorf("expected distribution 'ubuntu', got '%s'", info.Distribution)
	}
	if info.Version != "24.04" {
		t.Errorf("expected version '24.04', got '%s'", info.Version)
	}
}

func TestGetRecommendedWorkerProcesses(t *testing.T) {
	detector := NewDetector()
	workers := detector.GetRecommendedWorkerProcesses()

	if workers <= 0 {
		t.Error("expected positive worker processes")
	}
	if workers != runtime.NumCPU() {
		t.Errorf("expected %d workers, got %d", runtime.NumCPU(), workers)
	}
}

func TestGetRecommendedWorkerConnections(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		ram      uint64
		expected int
	}{
		{"512MB RAM", 512 * 1024 * 1024, 1024},
		{"1GB RAM", 1024 * 1024 * 1024, 1024},
		{"2GB RAM", 2 * 1024 * 1024 * 1024, 2048},
		{"4GB RAM", 4 * 1024 * 1024 * 1024, 4096},
		{"8GB RAM", 8 * 1024 * 1024 * 1024, 4096}, // Capped at 4096
		{"16GB RAM", 16 * 1024 * 1024 * 1024, 4096}, // Capped at 4096
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connections := detector.GetRecommendedWorkerConnections(tt.ram)
			if connections != tt.expected {
				t.Errorf("expected %d connections, got %d", tt.expected, connections)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{"100 bytes", 100, "100 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1.5 GB", uint64(1.5 * 1024 * 1024 * 1024), "1.5 GB"},
		{"10 GB", 10 * 1024 * 1024 * 1024, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsServiceInstalled(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	detector := NewDetector()

	// Test with a service that likely exists
	installed, err := detector.IsServiceInstalled("bash")
	if err != nil {
		t.Logf("IsServiceInstalled error (non-fatal): %v", err)
	}
	if !installed {
		t.Log("bash not detected as installed (may be expected)")
	}

	// Test with a service that likely doesn't exist
	installed, err = detector.IsServiceInstalled("nonexistent-service-xyz123")
	if err != nil {
		t.Logf("IsServiceInstalled error for nonexistent service (expected): %v", err)
	}
	if installed {
		t.Error("nonexistent service should not be installed")
	}
}

func TestGetServiceStatus(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	detector := NewDetector()

	// Test with a non-existent service
	status, err := detector.GetServiceStatus("nonexistent-service-xyz123")
	if err != nil {
		t.Logf("GetServiceStatus error (may be expected): %v", err)
	}
	if status != models.StatusNotInstalled && status != models.StatusUnknown {
		t.Logf("Expected not_installed or unknown for nonexistent service, got %s", status)
	}
}
