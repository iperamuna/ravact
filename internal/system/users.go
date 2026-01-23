package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// User represents a system user
type User struct {
	Username string
	UID      int
	GID      int
	HomeDir  string
	Shell    string
	HasSudo  bool
	Groups   []string
}

// Group represents a system group
type Group struct {
	Name    string
	GID     int
	Members []string
}

// UserManager handles user and group operations
type UserManager struct{}

// NewUserManager creates a new user manager
func NewUserManager() *UserManager {
	return &UserManager{}
}

// GetAllUsers returns all system users (UID >= 1000 for regular users)
func (um *UserManager) GetAllUsers() ([]User, error) {
	// Check if running on unsupported OS
	if runtime.GOOS == "darwin" {
		return um.getDarwinUsers()
	}

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/passwd: %w", err)
	}
	defer file.Close()

	var users []User
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}

		uid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}

		gid, err := strconv.Atoi(fields[3])
		if err != nil {
			continue
		}

		// Include system users and regular users (UID >= 1000 or root)
		if uid < 1000 && uid != 0 {
			continue
		}

		username := fields[0]
		user := User{
			Username: username,
			UID:      uid,
			GID:      gid,
			HomeDir:  fields[5],
			Shell:    fields[6],
			HasSudo:  um.userHasSudo(username),
			Groups:   um.getUserGroups(username),
		}

		users = append(users, user)
	}

	return users, scanner.Err()
}

// getDarwinUsers returns users on macOS with a warning
func (um *UserManager) getDarwinUsers() ([]User, error) {
	// On macOS, user management is different and not supported by this tool
	// Return a mock user with helpful message
	return []User{
		{
			Username: "⚠️  macOS Not Supported",
			UID:      0,
			GID:      0,
			HomeDir:  "This feature requires Linux (Ubuntu/Debian/RHEL)",
			Shell:    "Please run on a Linux VM or server",
			HasSudo:  false,
			Groups:   []string{"Deploy to Linux VM to use this feature"},
		},
	}, nil
}

// GetUser returns a specific user
func (um *UserManager) GetUser(username string) (*User, error) {
	users, err := um.GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", username)
}

// userHasSudo checks if user has sudo privileges
func (um *UserManager) userHasSudo(username string) bool {
	// Check if user is in sudo or wheel group
	groups := um.getUserGroups(username)
	for _, group := range groups {
		if group == "sudo" || group == "wheel" || group == "admin" {
			return true
		}
	}

	// Check if user is root
	if username == "root" {
		return true
	}

	return false
}

// getUserGroups returns all groups a user belongs to
func (um *UserManager) getUserGroups(username string) []string {
	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "groups", username)
	output, err := cmd.Output()
	if err != nil {
		// On macOS or if command fails, try reading from /etc/group
		if runtime.GOOS == "darwin" {
			return um.getUserGroupsFromFile(username)
		}
		return []string{}
	}

	// Output format: "username : group1 group2 group3"
	parts := strings.Split(string(output), ":")
	if len(parts) < 2 {
		return []string{}
	}

	groupsStr := strings.TrimSpace(parts[1])
	if groupsStr == "" {
		return []string{}
	}

	return strings.Fields(groupsStr)
}

// getUserGroupsFromFile reads groups from /etc/group file (fallback method)
func (um *UserManager) getUserGroupsFromFile(username string) []string {
	file, err := os.Open("/etc/group")
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var groups []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		groupName := fields[0]
		members := strings.Split(fields[3], ",")

		// Check if user is in this group's member list
		for _, member := range members {
			if member == username {
				groups = append(groups, groupName)
				break
			}
		}
	}

	return groups
}

// CreateUser creates a new user with the given username, password, and shell
func (um *UserManager) CreateUser(username, password, shell string) error {
	// Use useradd command to create user
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := []string{
		"-m",        // Create home directory
		"-s", shell, // Set shell
		username,
	}

	cmd := exec.CommandContext(ctx, "useradd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("useradd failed: %v - %s", err, string(output))
	}

	// Set password using chpasswd
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	passwdCmd := exec.CommandContext(ctx2, "chpasswd")
	passwdCmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s\n", username, password))
	output, err = passwdCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chpasswd failed: %v - %s", err, string(output))
	}

	return nil
}

// GrantSudo grants sudo privileges to a user
func (um *UserManager) GrantSudo(username string) error {
	// Add user to sudo group
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "usermod", "-aG", "sudo", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("usermod failed: %v - %s", err, string(output))
	}

	return nil
}

// RevokeSudo revokes sudo privileges from a user
func (um *UserManager) RevokeSudo(username string) error {
	// Remove user from sudo group
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "gpasswd", "-d", username, "sudo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpasswd failed: %v - %s", err, string(output))
	}

	return nil
}

// ToggleSudo toggles sudo privileges for a user
func (um *UserManager) ToggleSudo(username string) error {
	hasSudo := um.userHasSudo(username)
	
	if hasSudo {
		return um.RevokeSudo(username)
	}
	return um.GrantSudo(username)
}

// DeleteUser deletes a user and optionally their home directory
func (um *UserManager) DeleteUser(username string, removeHome bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args := []string{username}
	if removeHome {
		args = append([]string{"-r"}, args...) // -r removes home directory
	}

	cmd := exec.CommandContext(ctx, "userdel", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("userdel failed: %v - %s", err, string(output))
	}

	return nil
}

// GetAllGroups returns all system groups
func (um *UserManager) GetAllGroups() ([]Group, error) {
	file, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var groups []Group
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		gid, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}

		members := []string{}
		if fields[3] != "" {
			members = strings.Split(fields[3], ",")
		}

		group := Group{
			Name:    fields[0],
			GID:     gid,
			Members: members,
		}

		groups = append(groups, group)
	}

	return groups, scanner.Err()
}

// GetGroup returns a specific group
func (um *UserManager) GetGroup(groupname string) (*Group, error) {
	groups, err := um.GetAllGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.Name == groupname {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("group not found: %s", groupname)
}

// CreateGroup creates a new group
func (um *UserManager) CreateGroup(groupname string) error {
	cmd := exec.Command("groupadd", groupname)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// DeleteGroup deletes a group (only if empty)
func (um *UserManager) DeleteGroup(groupname string) error {
	// Check if group is empty
	group, err := um.GetGroup(groupname)
	if err != nil {
		return err
	}

	if len(group.Members) > 0 {
		return fmt.Errorf("group is not empty (has %d members)", len(group.Members))
	}

	cmd := exec.Command("groupdel", groupname)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// AddUserToGroup adds a user to a group
func (um *UserManager) AddUserToGroup(username, groupname string) error {
	cmd := exec.Command("usermod", "-a", "-G", groupname, username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (um *UserManager) RemoveUserFromGroup(username, groupname string) error {
	cmd := exec.Command("gpasswd", "-d", username, groupname)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	return nil
}

// ChangePassword changes a user's password
func (um *UserManager) ChangePassword(username, newPassword string) error {
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s\n", username, newPassword))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}
