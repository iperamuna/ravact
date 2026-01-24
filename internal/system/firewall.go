package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// FirewallType represents the type of firewall installed
type FirewallType string

const (
	FirewallUFW      FirewallType = "ufw"
	FirewallFirewalld FirewallType = "firewalld"
	FirewallNone     FirewallType = "none"
)

// FirewallRule represents a firewall rule
type FirewallRule struct {
	Port     string
	Protocol string
	Action   string // allow, deny
	From     string // IP or "Anywhere"
	Comment  string
}

// FirewallManager handles firewall operations
type FirewallManager struct {
	firewallType FirewallType
}

// NewFirewallManager creates a new firewall manager
func NewFirewallManager() *FirewallManager {
	return &FirewallManager{
		firewallType: detectFirewallType(),
	}
}

// detectFirewallType detects which firewall is installed
func detectFirewallType() FirewallType {
	if cmd := exec.Command("which", "ufw"); cmd.Run() == nil {
		return FirewallUFW
	}
	if cmd := exec.Command("which", "firewall-cmd"); cmd.Run() == nil {
		return FirewallFirewalld
	}
	return FirewallNone
}

// GetFirewallType returns the detected firewall type
func (m *FirewallManager) GetFirewallType() FirewallType {
	return m.firewallType
}

// GetStatus returns the firewall status
func (m *FirewallManager) GetStatus() (string, error) {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "unknown", err
		}
		if strings.Contains(string(output), "Status: active") {
			return "active", nil
		}
		return "inactive", nil

	case FirewallFirewalld:
		cmd := exec.Command("systemctl", "is-active", "firewalld")
		output, _ := cmd.Output()
		return strings.TrimSpace(string(output)), nil

	default:
		return "not installed", nil
	}
}

// GetRules returns the current firewall rules
func (m *FirewallManager) GetRules() ([]FirewallRule, error) {
	var rules []FirewallRule

	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "status", "numbered")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			// Parse UFW rules like: [ 1] 22/tcp                     ALLOW IN    Anywhere
			if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					portProto := parts[1]
					action := parts[2]
					from := "Anywhere"
					if len(parts) > 4 {
						from = parts[4]
					}

					portParts := strings.Split(portProto, "/")
					port := portParts[0]
					protocol := "tcp"
					if len(portParts) > 1 {
						protocol = portParts[1]
					}

					rules = append(rules, FirewallRule{
						Port:     port,
						Protocol: protocol,
						Action:   strings.ToLower(action),
						From:     from,
					})
				}
			}
		}

	case FirewallFirewalld:
		// Get open ports
		cmd := exec.Command("firewall-cmd", "--list-ports")
		output, err := cmd.Output()
		if err == nil {
			ports := strings.Fields(string(output))
			for _, portProto := range ports {
				parts := strings.Split(portProto, "/")
				if len(parts) == 2 {
					rules = append(rules, FirewallRule{
						Port:     parts[0],
						Protocol: parts[1],
						Action:   "allow",
						From:     "Anywhere",
					})
				}
			}
		}

		// Get open services
		cmd = exec.Command("firewall-cmd", "--list-services")
		output, err = cmd.Output()
		if err == nil {
			services := strings.Fields(string(output))
			for _, service := range services {
				rules = append(rules, FirewallRule{
					Port:     service,
					Protocol: "service",
					Action:   "allow",
					From:     "Anywhere",
				})
			}
		}
	}

	return rules, nil
}

// AllowPort allows a port through the firewall
func (m *FirewallManager) AllowPort(port, protocol string) error {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "allow", fmt.Sprintf("%s/%s", port, protocol))
		return cmd.Run()

	case FirewallFirewalld:
		cmd := exec.Command("firewall-cmd", "--permanent", fmt.Sprintf("--add-port=%s/%s", port, protocol))
		if err := cmd.Run(); err != nil {
			return err
		}
		// Reload to apply
		return exec.Command("firewall-cmd", "--reload").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// DenyPort denies a port through the firewall
func (m *FirewallManager) DenyPort(port, protocol string) error {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "deny", fmt.Sprintf("%s/%s", port, protocol))
		return cmd.Run()

	case FirewallFirewalld:
		cmd := exec.Command("firewall-cmd", "--permanent", fmt.Sprintf("--remove-port=%s/%s", port, protocol))
		if err := cmd.Run(); err != nil {
			return err
		}
		return exec.Command("firewall-cmd", "--reload").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// DeleteRule deletes a firewall rule by port
func (m *FirewallManager) DeleteRule(port, protocol string) error {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "delete", "allow", fmt.Sprintf("%s/%s", port, protocol))
		return cmd.Run()

	case FirewallFirewalld:
		cmd := exec.Command("firewall-cmd", "--permanent", fmt.Sprintf("--remove-port=%s/%s", port, protocol))
		if err := cmd.Run(); err != nil {
			return err
		}
		return exec.Command("firewall-cmd", "--reload").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// EnableFirewall enables the firewall
func (m *FirewallManager) EnableFirewall() error {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "--force", "enable")
		return cmd.Run()

	case FirewallFirewalld:
		if err := exec.Command("systemctl", "enable", "firewalld").Run(); err != nil {
			return err
		}
		return exec.Command("systemctl", "start", "firewalld").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// DisableFirewall disables the firewall
func (m *FirewallManager) DisableFirewall() error {
	switch m.firewallType {
	case FirewallUFW:
		cmd := exec.Command("ufw", "disable")
		return cmd.Run()

	case FirewallFirewalld:
		return exec.Command("systemctl", "stop", "firewalld").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// ReloadFirewall reloads firewall rules
func (m *FirewallManager) ReloadFirewall() error {
	switch m.firewallType {
	case FirewallUFW:
		if err := exec.Command("ufw", "disable").Run(); err != nil {
			return err
		}
		return exec.Command("ufw", "--force", "enable").Run()

	case FirewallFirewalld:
		return exec.Command("firewall-cmd", "--reload").Run()

	default:
		return fmt.Errorf("no firewall installed")
	}
}

// AllowService allows a service through firewalld
func (m *FirewallManager) AllowService(service string) error {
	if m.firewallType != FirewallFirewalld {
		return fmt.Errorf("service-based rules only supported on firewalld")
	}

	cmd := exec.Command("firewall-cmd", "--permanent", fmt.Sprintf("--add-service=%s", service))
	if err := cmd.Run(); err != nil {
		return err
	}
	return exec.Command("firewall-cmd", "--reload").Run()
}
