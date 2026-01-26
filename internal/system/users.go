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

// CreateUserPasswordless creates a new user without a password (SSH key-only authentication)
// This is the industry standard for server deployments where users authenticate via SSH keys
func (um *UserManager) CreateUserPasswordless(username, shell string) error {
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

	// Lock the password (disables password login but allows SSH key auth)
	// Using passwd -d removes the password, allowing passwordless su
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	passwdCmd := exec.CommandContext(ctx2, "passwd", "-d", username)
	output, err = passwdCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("passwd -d failed: %v - %s", err, string(output))
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

// GrantSudoNoPassword grants sudo privileges with NOPASSWD (no password required)
// This creates a sudoers.d file for the user with NOPASSWD:ALL
func (um *UserManager) GrantSudoNoPassword(username string) error {
	// First add to sudo group
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "usermod", "-aG", "sudo", username)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("usermod failed: %v - %s", err, string(output))
	}

	// Create sudoers.d file for NOPASSWD access
	sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s", username)
	sudoersContent := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)

	if err := os.WriteFile(sudoersFile, []byte(sudoersContent), 0440); err != nil {
		return fmt.Errorf("failed to create sudoers file: %v", err)
	}

	// Validate the sudoers file
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	checkCmd := exec.CommandContext(ctx2, "visudo", "-c", "-f", sudoersFile)
	output, err = checkCmd.CombinedOutput()
	if err != nil {
		// Remove invalid file
		os.Remove(sudoersFile)
		return fmt.Errorf("sudoers validation failed: %v - %s", err, string(output))
	}

	return nil
}

// EnablePasswordlessSu enables passwordless su for a user (without sudo)
// This is useful when you want to allow switching to a user without password
// but not grant full sudo privileges
func (um *UserManager) EnablePasswordlessSu(username string) error {
	// The user already has no password from CreateUserPasswordless
	// su without password works when the target user has no password set
	// No additional configuration needed - passwd -d already enables this
	return nil
}

// RevokeSudoNoPassword removes the NOPASSWD sudoers file for a user
func (um *UserManager) RevokeSudoNoPassword(username string) error {
	sudoersFile := fmt.Sprintf("/etc/sudoers.d/%s", username)
	if err := os.Remove(sudoersFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove sudoers file: %v", err)
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

// SSHKey represents an SSH key for a user
type SSHKey struct {
	Identifier     string // Key comment/identifier
	Type           string // rsa, ed25519, ecdsa
	Fingerprint    string // Key fingerprint
	IsLoginKey     bool   // Whether this key is authorized for login
	HasPassphrase  bool   // Whether the key has a passphrase
	IsInAgent      bool   // Whether this key is loaded in ssh-agent
	PublicKeyPath  string // Path to the public key file
	PrivateKeyPath string // Path to the private key file
}

// SSHKeyType represents the type of SSH key
type SSHKeyType string

const (
	SSHKeyTypeRSA     SSHKeyType = "rsa"
	SSHKeyTypeED25519 SSHKeyType = "ed25519"
	SSHKeyTypeECDSA   SSHKeyType = "ecdsa"
)

// GetUserSSHKeys returns all SSH keys for a user
func (um *UserManager) GetUserSSHKeys(username string) ([]SSHKey, error) {
	user, err := um.GetUser(username)
	if err != nil {
		return nil, err
	}

	sshDir := fmt.Sprintf("%s/.ssh", user.HomeDir)
	var keys []SSHKey

	// Check if .ssh directory exists
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return keys, nil
	}

	// Read authorized_keys to know which keys are for login
	authorizedKeys := make(map[string]bool)
	authKeysPath := fmt.Sprintf("%s/authorized_keys", sshDir)
	if authContent, err := os.ReadFile(authKeysPath); err == nil {
		lines := strings.Split(string(authContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				// Extract fingerprint from authorized key
				fp := um.getKeyFingerprint(line)
				if fp != "" {
					authorizedKeys[fp] = true
				}
			}
		}
	}

	// Find all key files in .ssh directory
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read .ssh directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// Skip non-key files
		if name == "authorized_keys" || name == "known_hosts" || name == "config" {
			continue
		}

		// Look for public key files
		if strings.HasSuffix(name, ".pub") {
			pubKeyPath := fmt.Sprintf("%s/%s", sshDir, name)
			privKeyPath := strings.TrimSuffix(pubKeyPath, ".pub")

			pubContent, err := os.ReadFile(pubKeyPath)
			if err != nil {
				continue
			}

			keyInfo := um.parseSSHPublicKey(string(pubContent))
			if keyInfo.Type == "" {
				continue
			}

			keyInfo.PublicKeyPath = pubKeyPath
			keyInfo.PrivateKeyPath = privKeyPath

			// Check if this key is authorized for login
			if keyInfo.Fingerprint != "" {
				keyInfo.IsLoginKey = authorizedKeys[keyInfo.Fingerprint]
			}

			// Check if private key has passphrase
			if _, err := os.Stat(privKeyPath); err == nil {
				keyInfo.HasPassphrase = um.checkKeyHasPassphrase(privKeyPath)
			}

			// Check if key is in ssh-agent
			keyInfo.IsInAgent = um.IsKeyInSSHAgent(pubKeyPath, username)

			keys = append(keys, keyInfo)
		}
	}

	return keys, nil
}

// parseSSHPublicKey parses an SSH public key and extracts information
func (um *UserManager) parseSSHPublicKey(pubKey string) SSHKey {
	parts := strings.Fields(strings.TrimSpace(pubKey))
	if len(parts) < 2 {
		return SSHKey{}
	}

	keyType := ""
	switch parts[0] {
	case "ssh-rsa":
		keyType = "rsa"
	case "ssh-ed25519":
		keyType = "ed25519"
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		keyType = "ecdsa"
	default:
		return SSHKey{}
	}

	identifier := ""
	if len(parts) >= 3 {
		identifier = strings.Join(parts[2:], " ")
	}

	// Get fingerprint
	fingerprint := um.getKeyFingerprint(pubKey)

	return SSHKey{
		Type:        keyType,
		Identifier:  identifier,
		Fingerprint: fingerprint,
	}
}

// getKeyFingerprint returns the fingerprint of an SSH public key
func (um *UserManager) getKeyFingerprint(pubKey string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ssh-keygen", "-lf", "-")
	cmd.Stdin = strings.NewReader(pubKey)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Output format: "256 SHA256:xxxx comment (TYPE)"
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// checkKeyHasPassphrase checks if a private key has a passphrase
func (um *UserManager) checkKeyHasPassphrase(privKeyPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to read the key with empty passphrase
	cmd := exec.CommandContext(ctx, "ssh-keygen", "-y", "-P", "", "-f", privKeyPath)
	err := cmd.Run()
	// If it fails, the key has a passphrase
	return err != nil
}

// GenerateSSHKey generates a new SSH key for a user
func (um *UserManager) GenerateSSHKey(username string, keyType SSHKeyType, identifier string, passphrase string, bits int, comment string) (string, error) {
	user, err := um.GetUser(username)
	if err != nil {
		return "", err
	}

	sshDir := fmt.Sprintf("%s/.ssh", user.HomeDir)

	// Create .ssh directory if it doesn't exist
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create .ssh directory: %w", err)
		}
		// Set ownership
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "chown", "-R", fmt.Sprintf("%s:%s", username, username), sshDir)
		cmd.Run()
	}

	// Generate key filename based on type and identifier
	safeIdentifier := strings.ReplaceAll(identifier, " ", "_")
	safeIdentifier = strings.ReplaceAll(safeIdentifier, "/", "_")
	keyName := fmt.Sprintf("id_%s_%s", keyType, safeIdentifier)
	keyPath := fmt.Sprintf("%s/%s", sshDir, keyName)

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		return "", fmt.Errorf("key with name %s already exists", keyName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use comment for the key comment, fallback to identifier if comment is empty
	keyComment := comment
	if keyComment == "" {
		keyComment = identifier
	}

	args := []string{
		"-t", string(keyType),
		"-f", keyPath,
		"-C", keyComment,
		"-N", passphrase,
	}

	// Add bits for RSA keys
	if keyType == SSHKeyTypeRSA {
		if bits == 0 {
			bits = 4096
		}
		args = append(args, "-b", fmt.Sprintf("%d", bits))
	}

	cmd := exec.CommandContext(ctx, "ssh-keygen", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %v - %s", err, string(output))
	}

	// Set proper ownership
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	chownCmd := exec.CommandContext(ctx2, "chown", fmt.Sprintf("%s:%s", username, username), keyPath, keyPath+".pub")
	chownCmd.Run()

	// Set proper permissions
	os.Chmod(keyPath, 0600)
	os.Chmod(keyPath+".pub", 0644)

	return keyPath, nil
}

// AddKeyToAuthorizedKeys adds a public key to the user's authorized_keys file
func (um *UserManager) AddKeyToAuthorizedKeys(username string, pubKeyPath string) error {
	user, err := um.GetUser(username)
	if err != nil {
		return err
	}

	sshDir := fmt.Sprintf("%s/.ssh", user.HomeDir)
	authKeysPath := fmt.Sprintf("%s/authorized_keys", sshDir)

	// Read the public key
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	// Create authorized_keys file if it doesn't exist
	file, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	// Add the key
	keyContent := strings.TrimSpace(string(pubKey))
	if _, err := file.WriteString(keyContent + "\n"); err != nil {
		return fmt.Errorf("failed to write to authorized_keys: %w", err)
	}

	// Set proper ownership
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "chown", fmt.Sprintf("%s:%s", username, username), authKeysPath)
	cmd.Run()

	return nil
}

// RemoveKeyFromAuthorizedKeys removes a public key from the user's authorized_keys file
func (um *UserManager) RemoveKeyFromAuthorizedKeys(username string, fingerprint string) error {
	user, err := um.GetUser(username)
	if err != nil {
		return err
	}

	authKeysPath := fmt.Sprintf("%s/.ssh/authorized_keys", user.HomeDir)

	content, err := os.ReadFile(authKeysPath)
	if err != nil {
		return fmt.Errorf("failed to read authorized_keys: %w", err)
	}

	var newLines []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			newLines = append(newLines, line)
			continue
		}

		// Get fingerprint of this key
		fp := um.getKeyFingerprint(line)
		if fp != fingerprint {
			newLines = append(newLines, line)
		}
	}

	// Write back
	err = os.WriteFile(authKeysPath, []byte(strings.Join(newLines, "\n")+"\n"), 0600)
	if err != nil {
		return fmt.Errorf("failed to write authorized_keys: %w", err)
	}

	return nil
}

// AddKeyToSSHAgent adds a private key to the SSH agent for a user
// It will start the ssh-agent if not running and add the key
func (um *UserManager) AddKeyToSSHAgent(privKeyPath string) error {
	// Extract username from the key path (e.g., /home/ubuntu/.ssh/id_ed25519 -> ubuntu)
	username := um.extractUsernameFromPath(privKeyPath)
	if username == "" {
		return fmt.Errorf("could not determine username from key path")
	}

	// Create a script that initializes ssh-agent and adds the key
	// This script will be run as the target user
	script := fmt.Sprintf(`
#!/bin/bash
set -e

# Check if SSH_AUTH_SOCK is set and agent is running
if [ -z "$SSH_AUTH_SOCK" ] || ! ssh-add -l &>/dev/null; then
    # Start ssh-agent and get its output
    eval $(ssh-agent -s) > /dev/null 2>&1
fi

# Add the key (use -t for a limited lifetime if desired)
ssh-add "%s" 2>&1

# Print the agent socket path so we can save it
echo "SSH_AUTH_SOCK=$SSH_AUTH_SOCK"
`, privKeyPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run as the target user
	cmd := exec.CommandContext(ctx, "su", "-", username, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add key to ssh-agent: %v - %s", err, string(output))
	}

	// Save the SSH_AUTH_SOCK to user's bashrc for persistence
	um.saveSSHAgentSocket(username, output)

	return nil
}

// extractUsernameFromPath extracts username from a path like /home/username/.ssh/...
func (um *UserManager) extractUsernameFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "home" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	// Check for root
	if strings.HasPrefix(path, "/root/") {
		return "root"
	}
	return ""
}

// saveSSHAgentSocket saves the SSH agent socket to user's profile for persistence
func (um *UserManager) saveSSHAgentSocket(username string, output []byte) {
	// Extract SSH_AUTH_SOCK from output
	lines := strings.Split(string(output), "\n")
	var socketPath string
	for _, line := range lines {
		if strings.HasPrefix(line, "SSH_AUTH_SOCK=") {
			socketPath = strings.TrimPrefix(line, "SSH_AUTH_SOCK=")
			break
		}
	}

	if socketPath == "" {
		return
	}

	user, err := um.GetUser(username)
	if err != nil {
		return
	}

	// Add to .bashrc if not already there
	bashrcPath := fmt.Sprintf("%s/.bashrc", user.HomeDir)
	content, err := os.ReadFile(bashrcPath)
	if err != nil {
		return
	}

	// Check if we already have SSH agent export
	agentExport := fmt.Sprintf("export SSH_AUTH_SOCK=%s", socketPath)
	if strings.Contains(string(content), "SSH_AUTH_SOCK=") {
		// Update existing
		lines := strings.Split(string(content), "\n")
		var newLines []string
		for _, line := range lines {
			if strings.Contains(line, "SSH_AUTH_SOCK=") {
				newLines = append(newLines, agentExport)
			} else {
				newLines = append(newLines, line)
			}
		}
		os.WriteFile(bashrcPath, []byte(strings.Join(newLines, "\n")), 0644)
	} else {
		// Add new
		f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf("\n# SSH Agent (added by Ravact)\n%s\n", agentExport))
	}
}

// DeleteSSHKey deletes an SSH key pair
func (um *UserManager) DeleteSSHKey(pubKeyPath string) error {
	privKeyPath := strings.TrimSuffix(pubKeyPath, ".pub")

	// Extract username from path
	username := um.extractUsernameFromPath(pubKeyPath)

	// Remove from authorized_keys first if present
	pubContent, err := os.ReadFile(pubKeyPath)
	if err == nil {
		keyInfo := um.parseSSHPublicKey(string(pubContent))
		if keyInfo.Fingerprint != "" && username != "" {
			um.RemoveKeyFromAuthorizedKeys(username, keyInfo.Fingerprint)
		}
	}

	// Remove from ssh-agent if present
	if username != "" {
		um.RemoveKeyFromSSHAgent(privKeyPath, username)
	}

	// Delete private key
	if err := os.Remove(privKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete private key: %w", err)
	}

	// Delete public key
	if err := os.Remove(pubKeyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete public key: %w", err)
	}

	return nil
}

// IsKeyInSSHAgent checks if a key is loaded in the SSH agent
func (um *UserManager) IsKeyInSSHAgent(pubKeyPath string, username string) bool {
	// Get the fingerprint of the key
	pubContent, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return false
	}
	
	keyFingerprint := um.getKeyFingerprint(string(pubContent))
	if keyFingerprint == "" {
		return false
	}

	user, err := um.GetUser(username)
	if err != nil {
		return false
	}

	// Try to find the SSH_AUTH_SOCK from user's environment
	// Check common locations and user's bashrc
	bashrcPath := fmt.Sprintf("%s/.bashrc", user.HomeDir)
	bashrcContent, err := os.ReadFile(bashrcPath)
	if err != nil {
		return false
	}

	// Extract SSH_AUTH_SOCK from bashrc
	var agentSocket string
	lines := strings.Split(string(bashrcContent), "\n")
	for _, line := range lines {
		if strings.Contains(line, "SSH_AUTH_SOCK=") {
			// Extract the path from "export SSH_AUTH_SOCK=/tmp/ssh-xxx/agent.xxx"
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				agentSocket = strings.TrimSpace(parts[1])
				agentSocket = strings.Trim(agentSocket, "\"'")
				break
			}
		}
	}

	if agentSocket == "" {
		return false
	}

	// Check if the socket file exists (agent is running)
	if _, err := os.Stat(agentSocket); os.IsNotExist(err) {
		return false
	}

	// Now check if the key is in the agent using the found socket
	script := fmt.Sprintf(`
export SSH_AUTH_SOCK="%s"
ssh-add -l 2>/dev/null || echo ""
`, agentSocket)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "su", "-", username, "-c", script)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if fingerprint is in the output
	return strings.Contains(string(output), keyFingerprint)
}

// RemoveKeyFromSSHAgent removes a private key from the SSH agent
func (um *UserManager) RemoveKeyFromSSHAgent(privKeyPath string, username string) error {
	// Create a script that removes the key from ssh-agent
	script := fmt.Sprintf(`
#!/bin/bash
# Try to remove from agent, ignore errors if agent not running or key not loaded
ssh-add -d "%s" 2>/dev/null || true
`, privKeyPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run as the target user
	cmd := exec.CommandContext(ctx, "su", "-", username, "-c", script)
	cmd.Run() // Ignore errors - key might not be in agent

	return nil
}

// DisablePasswordSSHLogin disables SSH login using password for a user or globally
func (um *UserManager) DisablePasswordSSHLogin(username string) error {
	sshdConfig := "/etc/ssh/sshd_config"

	content, err := os.ReadFile(sshdConfig)
	if err != nil {
		return fmt.Errorf("failed to read sshd_config: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	passwordAuthFound := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle PasswordAuthentication directive
		if strings.HasPrefix(trimmed, "PasswordAuthentication") || strings.HasPrefix(trimmed, "#PasswordAuthentication") {
			if !passwordAuthFound {
				newLines = append(newLines, "PasswordAuthentication no")
				passwordAuthFound = true
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// Add if not found
	if !passwordAuthFound {
		newLines = append(newLines, "PasswordAuthentication no")
	}

	// Write back
	if err := os.WriteFile(sshdConfig, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write sshd_config: %w", err)
	}

	// Reload SSH service
	return um.reloadSSHService()
}

// DisableSSHKeyLogin disables SSH key login for a specific user
func (um *UserManager) DisableSSHKeyLogin(username string) error {
	user, err := um.GetUser(username)
	if err != nil {
		return err
	}

	authKeysPath := fmt.Sprintf("%s/.ssh/authorized_keys", user.HomeDir)

	// Rename authorized_keys to authorized_keys.disabled
	disabledPath := authKeysPath + ".disabled"
	if _, err := os.Stat(authKeysPath); err == nil {
		if err := os.Rename(authKeysPath, disabledPath); err != nil {
			return fmt.Errorf("failed to disable authorized_keys: %w", err)
		}
	}

	return nil
}

// EnableSSHKeyLogin re-enables SSH key login for a specific user
func (um *UserManager) EnableSSHKeyLogin(username string) error {
	user, err := um.GetUser(username)
	if err != nil {
		return err
	}

	authKeysPath := fmt.Sprintf("%s/.ssh/authorized_keys", user.HomeDir)
	disabledPath := authKeysPath + ".disabled"

	// Rename authorized_keys.disabled back to authorized_keys
	if _, err := os.Stat(disabledPath); err == nil {
		if err := os.Rename(disabledPath, authKeysPath); err != nil {
			return fmt.Errorf("failed to enable authorized_keys: %w", err)
		}
	}

	return nil
}

// IsSSHKeyLoginDisabled checks if SSH key login is disabled for a user
func (um *UserManager) IsSSHKeyLoginDisabled(username string) bool {
	user, err := um.GetUser(username)
	if err != nil {
		return false
	}

	disabledPath := fmt.Sprintf("%s/.ssh/authorized_keys.disabled", user.HomeDir)
	_, err = os.Stat(disabledPath)
	return err == nil
}

// IsPasswordSSHLoginDisabled checks if password SSH login is disabled globally
func (um *UserManager) IsPasswordSSHLoginDisabled() bool {
	content, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "PasswordAuthentication") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 && strings.ToLower(parts[1]) == "no" {
				return true
			}
		}
	}

	return false
}

// EnablePasswordSSHLogin enables SSH login using password
func (um *UserManager) EnablePasswordSSHLogin() error {
	sshdConfig := "/etc/ssh/sshd_config"

	content, err := os.ReadFile(sshdConfig)
	if err != nil {
		return fmt.Errorf("failed to read sshd_config: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	passwordAuthFound := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle PasswordAuthentication directive
		if strings.HasPrefix(trimmed, "PasswordAuthentication") || strings.HasPrefix(trimmed, "#PasswordAuthentication") {
			if !passwordAuthFound {
				newLines = append(newLines, "PasswordAuthentication yes")
				passwordAuthFound = true
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// Add if not found
	if !passwordAuthFound {
		newLines = append(newLines, "PasswordAuthentication yes")
	}

	// Write back
	if err := os.WriteFile(sshdConfig, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write sshd_config: %w", err)
	}

	// Reload SSH service
	return um.reloadSSHService()
}

// reloadSSHService reloads the SSH service
func (um *UserManager) reloadSSHService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try systemctl first (modern systems)
	cmd := exec.CommandContext(ctx, "systemctl", "reload", "sshd")
	if err := cmd.Run(); err != nil {
		// Try ssh instead of sshd
		cmd = exec.CommandContext(ctx, "systemctl", "reload", "ssh")
		if err := cmd.Run(); err != nil {
			// Try service command (older systems)
			cmd = exec.CommandContext(ctx, "service", "ssh", "reload")
			if err := cmd.Run(); err != nil {
				cmd = exec.CommandContext(ctx, "service", "sshd", "reload")
				return cmd.Run()
			}
		}
	}
	return nil
}
