package system

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/iperamuna/ravact/internal/models"
)

// Detector provides system detection capabilities
type Detector struct{}

// NewDetector creates a new system detector
func NewDetector() *Detector {
	return &Detector{}
}

// GetSystemInfo retrieves comprehensive system information
func (d *Detector) GetSystemInfo() (*models.SystemInfo, error) {
	info := &models.SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
		IsRoot:   d.IsRoot(),
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		info.Hostname = hostname
	}

	// Get distribution info (Linux only)
	if runtime.GOOS == "linux" {
		if err := d.detectLinuxDistribution(info); err != nil {
			// Non-fatal, continue
		}

		// Get kernel version
		if kernel, err := d.getKernelVersion(); err == nil {
			info.Kernel = kernel
		}
	}

	// Get RAM info (works on both Linux and macOS)
	if ram, err := d.getTotalRAM(); err == nil {
		info.TotalRAM = ram
	}

	// Get disk info (works on both Linux and macOS)
	if disk, err := d.getTotalDisk(); err == nil {
		info.TotalDisk = disk
	}

	return info, nil
}

// IsRoot checks if the current process is running as root
func (d *Detector) IsRoot() bool {
	return os.Geteuid() == 0
}

// detectLinuxDistribution detects the Linux distribution
func (d *Detector) detectLinuxDistribution(info *models.SystemInfo) error {
	// Try /etc/os-release first (modern standard)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		return d.parseOSRelease(data, info)
	}

	// Try /etc/lsb-release
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		return d.parseLSBRelease(data, info)
	}

	// Fallback to checking specific files
	distros := map[string]string{
		"/etc/debian_version": "debian",
		"/etc/redhat-release": "redhat",
		"/etc/centos-release": "centos",
		"/etc/fedora-release": "fedora",
	}

	for file, distro := range distros {
		if _, err := os.Stat(file); err == nil {
			info.Distribution = distro
			if data, err := os.ReadFile(file); err == nil {
				info.Version = strings.TrimSpace(string(data))
			}
			return nil
		}
	}

	return fmt.Errorf("unable to detect distribution")
}

// parseOSRelease parses /etc/os-release file
func (d *Detector) parseOSRelease(data []byte, info *models.SystemInfo) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			info.Distribution = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			info.Version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}
	return scanner.Err()
}

// parseLSBRelease parses /etc/lsb-release file
func (d *Detector) parseLSBRelease(data []byte, info *models.SystemInfo) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DISTRIB_ID=") {
			info.Distribution = strings.ToLower(strings.TrimPrefix(line, "DISTRIB_ID="))
		} else if strings.HasPrefix(line, "DISTRIB_RELEASE=") {
			info.Version = strings.TrimPrefix(line, "DISTRIB_RELEASE=")
		}
	}
	return scanner.Err()
}

// getKernelVersion gets the kernel version
func (d *Detector) getKernelVersion() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getTotalRAM gets total system RAM in bytes
func (d *Detector) getTotalRAM() (uint64, error) {
	if runtime.GOOS == "linux" {
		return d.getTotalRAMLinux()
	} else if runtime.GOOS == "darwin" {
		return d.getTotalRAMMacOS()
	}
	return 0, fmt.Errorf("unsupported platform for RAM detection")
}

// getTotalRAMLinux gets RAM on Linux
func (d *Detector) getTotalRAMLinux() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err != nil {
					return 0, err
				}
				return kb * 1024, nil // Convert KB to bytes
			}
		}
	}
	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

// getTotalRAMMacOS gets RAM on macOS
func (d *Detector) getTotalRAMMacOS() (uint64, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	ramBytes, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, err
	}

	return ramBytes, nil
}

// getTotalDisk gets total disk space in bytes
func (d *Detector) getTotalDisk() (uint64, error) {
	if runtime.GOOS == "linux" {
		return d.getTotalDiskLinux()
	} else if runtime.GOOS == "darwin" {
		return d.getTotalDiskMacOS()
	}
	return 0, fmt.Errorf("unsupported platform for disk detection")
}

// getTotalDiskLinux gets disk space on Linux
func (d *Detector) getTotalDiskLinux() (uint64, error) {
	cmd := exec.Command("df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// getTotalDiskMacOS gets disk space on macOS
func (d *Detector) getTotalDiskMacOS() (uint64, error) {
	cmd := exec.Command("df", "-k", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	// df -k returns 512-byte blocks on macOS, so multiply by 512
	// Actually, -k means 1K blocks
	kb, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return kb * 1024, nil // Convert KB to bytes
}

// IsServiceInstalled checks if a service/package is installed
func (d *Detector) IsServiceInstalled(serviceName string) (bool, error) {
	// For tools that don't run as services, check binary directly
	binaryOnlyTools := map[string][]string{
		"certbot": {"certbot"},
		"git":     {"git"},
		"node":    {"node", "nodejs"},
		"ufw":     {"ufw"},
	}

	if binaries, isBinaryOnly := binaryOnlyTools[serviceName]; isBinaryOnly {
		for _, binary := range binaries {
			cmd := exec.Command("which", binary)
			if err := cmd.Run(); err == nil {
				return true, nil
			}
		}
		return false, nil
	}

	// Try systemctl first for services
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), serviceName) {
		return true, nil
	}

	// Try which command as fallback
	cmd = exec.Command("which", serviceName)
	err = cmd.Run()
	return err == nil, nil
}

// GetServiceStatus gets the status of a service
func (d *Detector) GetServiceStatus(serviceName string) (models.ServiceStatus, error) {
	// Check if installed first
	installed, err := d.IsServiceInstalled(serviceName)
	if err != nil {
		return models.StatusUnknown, err
	}
	if !installed {
		return models.StatusNotInstalled, nil
	}

	// For tools that don't run as services, just return Installed
	binaryOnlyTools := map[string]bool{
		"certbot": true,
		"git":     true,
		"node":    true,
		"ufw":     true,
	}

	if binaryOnlyTools[serviceName] {
		return models.StatusInstalled, nil
	}

	// Check if running via systemctl
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	status := strings.TrimSpace(string(output))

	switch status {
	case "active":
		return models.StatusRunning, nil
	case "inactive":
		return models.StatusStopped, nil
	case "failed":
		return models.StatusFailed, nil
	default:
		return models.StatusInstalled, nil
	}
}

// GetRecommendedWorkerProcesses returns recommended nginx worker processes
func (d *Detector) GetRecommendedWorkerProcesses() int {
	return runtime.NumCPU()
}

// GetRecommendedWorkerConnections returns recommended nginx worker connections based on RAM
func (d *Detector) GetRecommendedWorkerConnections(totalRAM uint64) int {
	// Basic formula: 1024 connections per GB of RAM, max 4096
	gb := totalRAM / (1024 * 1024 * 1024)
	connections := int(gb) * 1024
	if connections < 1024 {
		connections = 1024
	}
	if connections > 4096 {
		connections = 4096
	}
	return connections
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetHostInfo returns hostname with IP in format "hostname (ip)" or just "hostname"
func GetHostInfo() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	ipAddr := GetPrimaryIP()
	if ipAddr != "" && ipAddr != "N/A" {
		return fmt.Sprintf("%s (%s)", hostname, ipAddr)
	}
	return hostname
}

// GetPrimaryIP returns the primary IP address of the system
func GetPrimaryIP() string {
	// Try to get IP from hostname command first (most reliable for primary IP)
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err == nil {
		ips := strings.Fields(strings.TrimSpace(string(output)))
		if len(ips) > 0 {
			return ips[0] // Return the first (primary) IP
		}
	}

	// Fallback: try ip command on Linux
	cmd = exec.Command("ip", "-4", "addr", "show", "scope", "global")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "inet ") {
				fields := strings.Fields(line)
				for i, field := range fields {
					if field == "inet" && i+1 < len(fields) {
						ip := strings.Split(fields[i+1], "/")[0]
						return ip
					}
				}
			}
		}
	}

	// Fallback for macOS: use ifconfig
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ipconfig", "getifaddr", "en0")
		output, err = cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	return "N/A"
}

// IsPortInUse checks if a TCP port is currently in use
func (d *Detector) IsPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}
