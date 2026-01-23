package system

import (
	"runtime"
	"testing"
)

func TestGetTotalRAMMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	detector := NewDetector()
	ram, err := detector.getTotalRAMMacOS()
	if err != nil {
		t.Fatalf("getTotalRAMMacOS failed: %v", err)
	}

	if ram == 0 {
		t.Error("expected RAM to be greater than 0")
	}

	// RAM should be at least 1GB on any modern Mac
	if ram < 1024*1024*1024 {
		t.Errorf("RAM seems too small: %d bytes", ram)
	}

	t.Logf("Detected RAM: %s", FormatBytes(ram))
}

func TestGetTotalDiskMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	detector := NewDetector()
	disk, err := detector.getTotalDiskMacOS()
	if err != nil {
		t.Fatalf("getTotalDiskMacOS failed: %v", err)
	}

	if disk == 0 {
		t.Error("expected disk to be greater than 0")
	}

	// Disk should be at least 10GB on any Mac
	if disk < 10*1024*1024*1024 {
		t.Errorf("Disk seems too small: %d bytes", disk)
	}

	t.Logf("Detected Disk: %s", FormatBytes(disk))
}

func TestSystemInfoMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	detector := NewDetector()
	info, err := detector.GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.OS != "darwin" {
		t.Errorf("expected OS 'darwin', got '%s'", info.OS)
	}

	if info.TotalRAM == 0 {
		t.Error("RAM should be detected on macOS")
	}

	if info.TotalDisk == 0 {
		t.Error("Disk should be detected on macOS")
	}

	t.Logf("System Info on macOS:")
	t.Logf("  OS: %s", info.OS)
	t.Logf("  Arch: %s", info.Arch)
	t.Logf("  CPU: %d cores", info.CPUCount)
	t.Logf("  RAM: %s", FormatBytes(info.TotalRAM))
	t.Logf("  Disk: %s", FormatBytes(info.TotalDisk))
}
