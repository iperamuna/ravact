package screens

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// GitState represents the current state of the git management screen
type GitState int

const (
	GitStateMenu GitState = iota
	GitStateTestConnectionForm
	GitStateCloneForm
	GitStateConfirmClone
	GitStateAddRemoteForm
	GitStateConfirmRemote
	GitStateGitOpForm
	GitStateSetSystemUserForm
)

// GitInfo holds information about the current git repository
type GitInfo struct {
	IsRepo           bool
	DubiousOwnership bool
	ErrorMsg         string
	RemoteURL        string
	RemoteName       string
	Branch           string
	LastCommit       string
	CommitMsg        string
	HasChanges       bool
	Ahead            int
	Behind           int
	SystemUser       string // meta.systemuser config value
}

// GitAction represents a git action menu item
type GitAction struct {
	ID          string
	Name        string
	Description string
}

// GitManagementModel represents the git management screen
type GitManagementModel struct {
	theme   *theme.Theme
	width   int
	height  int
	cursor  int
	actions []GitAction
	gitInfo GitInfo
	err     error
	success string

	// State management
	state      GitState
	currentDir string

	// Form for test connection
	testForm     *huh.Form
	selectedUser string
	selectedKey  string

	// Form for add remote
	remoteForm *huh.Form
	remoteUser string
	remoteURL  string

	// Form for clone
	cloneForm *huh.Form
	cloneUser string
	cloneURL  string

	// Form for git operations (pull, fetch, status, etc.)
	gitOpForm   *huh.Form
	gitOpUser   string
	gitOpAction string

	// Form for setting system user
	systemUserForm *huh.Form
	systemUser     string

	// User manager
	userManager    *system.UserManager
	availableUsers []string
}

// NewGitManagementModel creates a new git management model
func NewGitManagementModel() GitManagementModel {
	gitInfo := getGitInfo()

	// Get current directory
	currentDir, _ := os.Getwd()

	actions := []GitAction{
		{ID: "refresh", Name: "Refresh Git Info", Description: "Refresh repository information"},
	}

	if gitInfo.DubiousOwnership {
		actions = append(actions, GitAction{ID: "fix_ownership", Name: "Fix Git Ownership Detection", Description: "Add this directory to safe.directory config"})
	}

	actions = append(actions, []GitAction{
		{ID: "test_connection", Name: "Test Git Connection", Description: "Test SSH connection to GitHub/GitLab"},
		{ID: "clone_repo", Name: "Clone Git Repo", Description: "Clone a repository into this directory"},
		{ID: "add_remote", Name: "Add/Setup Git Remote", Description: "Add a new git remote URL"},
		{ID: "change_remote", Name: "Change Remote URL", Description: "Update the remote URL"},
		{ID: "remove_remote", Name: "Remove Remote", Description: "Remove the git remote"},
		{ID: "git_pull", Name: "Git Pull", Description: "Pull latest changes from remote"},
		{ID: "git_fetch", Name: "Git Fetch", Description: "Fetch changes from remote without merging"},
		{ID: "git_status", Name: "Git Status", Description: "Show detailed git status"},
		{ID: "set_system_user", Name: "Set System User", Description: "Set the user for git operations in this repo"},
		{ID: "back", Name: "← Back to Site Commands", Description: "Return to site commands menu"},
	}...)

	// Get user manager and available users
	um := system.NewUserManager()
	var availableUsers []string
	if users, err := um.GetAllUsers(); err == nil {
		for _, u := range users {
			// Only include users with home directories (real users)
			if strings.HasPrefix(u.HomeDir, "/home/") || u.Username == "root" {
				availableUsers = append(availableUsers, u.Username)
			}
		}
	}

	return GitManagementModel{
		theme:          theme.DefaultTheme(),
		cursor:         0,
		actions:        actions,
		gitInfo:        gitInfo,
		state:          GitStateMenu,
		currentDir:     currentDir,
		userManager:    um,
		availableUsers: availableUsers,
	}
}

// getGitInfo retrieves git repository information
func getGitInfo() GitInfo {
	info := GitInfo{}

	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if strings.Contains(errStr, "dubious ownership") {
			info.IsRepo = false
			info.DubiousOwnership = true
			return info
		}
		info.IsRepo = false
		return info
	}
	info.IsRepo = true

	// Get current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	// Get remote name and URL
	cmd = exec.Command("git", "remote")
	if output, err := cmd.Output(); err == nil {
		remotes := strings.Fields(string(output))
		if len(remotes) > 0 {
			info.RemoteName = remotes[0]

			// Get remote URL
			cmd = exec.Command("git", "remote", "get-url", info.RemoteName)
			if urlOutput, err := cmd.Output(); err == nil {
				info.RemoteURL = strings.TrimSpace(string(urlOutput))
			}
		}
	}

	// Get last commit hash (short)
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.LastCommit = strings.TrimSpace(string(output))
	}

	// Get last commit message
	cmd = exec.Command("git", "log", "-1", "--pretty=%s")
	if output, err := cmd.Output(); err == nil {
		info.CommitMsg = strings.TrimSpace(string(output))
		// Truncate if too long
		if len(info.CommitMsg) > 60 {
			info.CommitMsg = info.CommitMsg[:57] + "..."
		}
	}

	// Check for uncommitted changes
	cmd = exec.Command("git", "status", "--porcelain")
	if output, err := cmd.Output(); err == nil {
		info.HasChanges = len(strings.TrimSpace(string(output))) > 0
	}

	// Get ahead/behind info
	if info.RemoteName != "" && info.Branch != "" {
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", fmt.Sprintf("%s/%s...HEAD", info.RemoteName, info.Branch))
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &info.Behind)
				fmt.Sscanf(parts[1], "%d", &info.Ahead)
			}
		}
	}

	// Get meta.systemuser config
	cmd = exec.Command("git", "config", "--get", "meta.systemuser")
	if output, err := cmd.Output(); err == nil {
		info.SystemUser = strings.TrimSpace(string(output))
	}

	return info
}

// Init initializes the git management screen
func (m GitManagementModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for git management
func (m GitManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	// Handle different states
	switch m.state {
	case GitStateMenu:
		return m.updateMenu(msg)
	case GitStateTestConnectionForm:
		return m.updateTestConnectionForm(msg)
	case GitStateCloneForm:
		return m.updateCloneForm(msg)
	case GitStateConfirmClone:
		return m.updateConfirmClone(msg)
	case GitStateAddRemoteForm:
		return m.updateAddRemoteForm(msg)
	case GitStateConfirmRemote:
		return m.updateConfirmRemote(msg)
	case GitStateGitOpForm:
		return m.updateGitOpForm(msg)
	case GitStateSetSystemUserForm:
		return m.updateSetSystemUserForm(msg)
	}

	return m, nil
}

// updateMenu handles the main menu state
func (m GitManagementModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SiteCommandsScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.executeAction()
		}
	}

	return m, nil
}

// updateTestConnectionForm handles the test connection form state
func (m GitManagementModel) updateTestConnectionForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.testForm != nil {
		form, cmd := m.testForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.testForm = f
		}

		// Check if form is completed
		if m.testForm.State == huh.StateCompleted {
			return m.runTestConnection()
		}

		// Handle escape to cancel
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = GitStateMenu
				m.testForm = nil
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateCloneForm handles the clone form state
func (m GitManagementModel) updateCloneForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.cloneForm != nil {
		form, cmd := m.cloneForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.cloneForm = f
		}

		// Check if form is completed
		if m.cloneForm.State == huh.StateCompleted {
			// Read form values
			m.cloneUser = m.cloneForm.GetString("cloneUser")
			m.cloneURL = m.cloneForm.GetString("cloneURL")

			// Move to confirmation state
			m.state = GitStateConfirmClone
			return m, nil
		}

		// Handle escape to cancel
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = GitStateMenu
				m.cloneForm = nil
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateConfirmClone handles the clone confirmation state
func (m GitManagementModel) updateConfirmClone(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "backspace", "n", "N":
			m.state = GitStateMenu
			m.cloneForm = nil
			return m, nil
		case "y", "Y", "enter":
			// Check folder permissions and change if needed, then clone
			return m.prepareAndClone()
		}
	}
	return m, nil
}

// updateGitOpForm handles the git operation form state
func (m GitManagementModel) updateGitOpForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.gitOpForm != nil {
		form, cmd := m.gitOpForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.gitOpForm = f
		}

		// Check if form is completed
		if m.gitOpForm.State == huh.StateCompleted {
			// Read form values
			m.gitOpUser = m.gitOpForm.GetString("gitOpUser")

			// Execute the git operation
			return m.executeGitOp()
		}

		// Handle escape to cancel
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = GitStateMenu
				m.gitOpForm = nil
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateSetSystemUserForm handles the set system user form state
func (m GitManagementModel) updateSetSystemUserForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.systemUserForm != nil {
		form, cmd := m.systemUserForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.systemUserForm = f
		}

		// Check if form is completed
		if m.systemUserForm.State == huh.StateCompleted {
			// Read form values
			m.systemUser = m.systemUserForm.GetString("systemUser")

			// Set the git config
			return m.setSystemUser()
		}

		// Handle escape to cancel
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = GitStateMenu
				m.systemUserForm = nil
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// buildSetSystemUserForm creates the set system user form
func (m *GitManagementModel) buildSetSystemUserForm() *huh.Form {
	// Build user options
	var userOptions []huh.Option[string]
	for _, user := range m.availableUsers {
		userOptions = append(userOptions, huh.NewOption(user, user))
	}

	// Set default user from current config or first available
	if m.systemUser == "" {
		if m.gitInfo.SystemUser != "" {
			m.systemUser = m.gitInfo.SystemUser
		} else if len(m.availableUsers) > 0 {
			m.systemUser = m.availableUsers[0]
		}
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("systemUser").
				Title("Select System User").
				Description("This user will be used for all git operations in this repository").
				Options(userOptions...).
				Value(&m.systemUser),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// setSystemUser saves the system user to git config
func (m GitManagementModel) setSystemUser() (tea.Model, tea.Cmd) {
	if m.systemUser == "" {
		m.state = GitStateMenu
		m.err = fmt.Errorf("no user selected")
		m.systemUserForm = nil
		return m, nil
	}

	// Set the git config
	cmd := exec.Command("git", "config", "meta.systemuser", m.systemUser)
	if output, err := cmd.CombinedOutput(); err != nil {
		m.err = fmt.Errorf("failed to set system user: %s", strings.TrimSpace(string(output)))
	} else {
		m.success = fmt.Sprintf("✓ System user set to '%s' for this repository", m.systemUser)
		m.gitInfo = getGitInfo()
	}

	m.state = GitStateMenu
	m.systemUserForm = nil
	return m, nil
}

// buildGitOpForm creates the git operation form with user selection
func (m *GitManagementModel) buildGitOpForm(action string) *huh.Form {
	// Build user options
	var userOptions []huh.Option[string]
	for _, user := range m.availableUsers {
		userOptions = append(userOptions, huh.NewOption(user, user))
	}

	// Set default user if not set
	if m.gitOpUser == "" && len(m.availableUsers) > 0 {
		m.gitOpUser = m.availableUsers[0]
	}

	// Store the action
	m.gitOpAction = action

	// Get action description
	actionDesc := ""
	switch action {
	case "git_pull":
		actionDesc = "Pull latest changes from remote"
	case "git_fetch":
		actionDesc = "Fetch changes from remote without merging"
	case "git_status":
		actionDesc = "Show detailed git status"
	case "change_remote":
		actionDesc = "Change the remote URL"
	case "remove_remote":
		actionDesc = "Remove the git remote"
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("gitOpUser").
				Title("Select User").
				Description(fmt.Sprintf("Run as this user: %s", actionDesc)).
				Options(userOptions...).
				Value(&m.gitOpUser),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// executeGitOp executes the selected git operation
func (m GitManagementModel) executeGitOp() (tea.Model, tea.Cmd) {
	if m.gitOpUser == "" {
		m.state = GitStateMenu
		m.err = fmt.Errorf("no user selected")
		m.gitOpForm = nil
		return m, nil
	}

	var gitCmd string
	var description string

	switch m.gitOpAction {
	case "git_pull":
		gitCmd = "git pull"
		description = "Pulling latest changes"
	case "git_fetch":
		gitCmd = "git fetch --all"
		description = "Fetching from all remotes"
	case "git_status":
		gitCmd = "git status"
		description = "Git Status"
	case "change_remote":
		// For change_remote, we need to go to the add remote form
		m.state = GitStateAddRemoteForm
		m.remoteUser = m.gitOpUser
		m.remoteForm = m.buildAddRemoteForm()
		m.gitOpForm = nil
		return m, m.remoteForm.Init()
	case "remove_remote":
		// Execute remove remote directly
		if m.gitInfo.RemoteName == "" {
			m.err = fmt.Errorf("no remote configured")
			m.state = GitStateMenu
			m.gitOpForm = nil
			return m, nil
		}

		script := fmt.Sprintf(`
echo "Removing remote as user %s..."

sudo -i -u %s bash << 'EOF'
cd "%s"
git remote remove %s 2>&1
EOF
`, m.gitOpUser, m.gitOpUser, m.currentDir, m.gitInfo.RemoteName)

		cmd := exec.Command("bash", "-c", script)
		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))

		if err != nil {
			m.err = fmt.Errorf("failed to remove remote: %s", outputStr)
		} else {
			m.success = fmt.Sprintf("✓ Remote '%s' removed successfully", m.gitInfo.RemoteName)
			m.gitInfo = getGitInfo()
		}

		m.state = GitStateMenu
		m.gitOpForm = nil
		return m, nil
	}

	// For pull, fetch, status - build script with ssh-agent
	script := fmt.Sprintf(`
echo "Running git %s as user %s..."
echo ""

sudo -i -u %s bash << 'EOF'
cd "%s"

# Start ssh-agent
eval $(ssh-agent -s) > /dev/null 2>&1

# Function to add keys safely
add_ssh_keys() {
    local added=0
    # Search for private keys in ~/.ssh
    for key in ~/.ssh/id_ed25519 ~/.ssh/id_rsa ~/.ssh/id_ecdsa ~/.ssh/id_dsa ~/.ssh/id_*; do
        if [[ -f "$key" && ! "$key" == *.pub && ! "$key" == *known_hosts* && ! "$key" == *config* && ! "$key" == *authorized_keys* ]]; then
            # Only add if it seems like a private key
            if head -n 1 "$key" | grep -q "PRIVATE KEY" 2>/dev/null; then
                ssh-add "$key" 2>/dev/null && ((added++))
            fi
        fi
    done
    
    if [[ $added -gt 0 ]]; then
        echo "✓ Loaded $added SSH key(s)"
    fi
}

add_ssh_keys

# Run git command
echo ""
echo "Executing: %s"
%s 2>&1
EXIT_CODE=$?

# Cleanup
ssh-agent -k > /dev/null 2>&1 || true

exit $EXIT_CODE
EOF
`, m.gitOpAction, m.gitOpUser, m.gitOpUser, m.currentDir, gitCmd, gitCmd)

	m.state = GitStateMenu
	m.gitOpForm = nil

	// Use ExecutionStartMsg to show output in execution screen
	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     script,
			Description: description,
		}
	}
}

// updateAddRemoteForm handles the add remote form state
func (m GitManagementModel) updateAddRemoteForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.remoteForm != nil {
		form, cmd := m.remoteForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.remoteForm = f
		}

		// Check if form is completed
		if m.remoteForm.State == huh.StateCompleted {
			// Read form values
			m.remoteUser = m.remoteForm.GetString("remoteUser")
			m.remoteURL = m.remoteForm.GetString("remoteURL")

			// Move to confirmation state
			m.state = GitStateConfirmRemote
			return m, nil
		}

		// Handle escape to cancel
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = GitStateMenu
				m.remoteForm = nil
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

// updateConfirmRemote handles the confirmation state
func (m GitManagementModel) updateConfirmRemote(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "backspace", "n", "N":
			m.state = GitStateMenu
			m.remoteForm = nil
			return m, nil
		case "y", "Y", "enter":
			return m.setupGitRemote()
		}
	}
	return m, nil
}

// buildTestConnectionForm creates the test connection form
func (m *GitManagementModel) buildTestConnectionForm() *huh.Form {
	// Build user options
	var userOptions []huh.Option[string]
	for _, user := range m.availableUsers {
		userOptions = append(userOptions, huh.NewOption(user, user))
	}

	// Set default user if not set
	if m.selectedUser == "" && len(m.availableUsers) > 0 {
		m.selectedUser = m.availableUsers[0]
	}

	// Build key options based on selected user
	keyOptions := m.getKeyOptionsForUser(m.selectedUser)

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("selectedUser").
				Title("Select User").
				Description("SSH connection will be tested using this user's SSH keys").
				Options(userOptions...).
				Value(&m.selectedUser),

			huh.NewSelect[string]().
				Key("selectedKey").
				Title("Select SSH Key").
				Description("Choose a specific key or let SSH auto-detect").
				Options(keyOptions...).
				Value(&m.selectedKey),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// getKeyOptionsForUser returns SSH key options for a given user
func (m *GitManagementModel) getKeyOptionsForUser(username string) []huh.Option[string] {
	keyOptions := []huh.Option[string]{
		huh.NewOption("Auto-detect (try all keys)", "auto"),
	}

	if username == "" {
		return keyOptions
	}

	// Get SSH keys for the user
	keys, err := m.userManager.GetUserSSHKeys(username)
	if err != nil {
		return keyOptions
	}

	for _, key := range keys {
		// Create a display name for the key
		identifier := key.Identifier
		if identifier == "" {
			identifier = strings.TrimSuffix(key.PublicKeyPath, ".pub")
			parts := strings.Split(identifier, "/")
			if len(parts) > 0 {
				identifier = parts[len(parts)-1]
			}
		}

		displayName := fmt.Sprintf("%s (%s)", identifier, strings.ToUpper(key.Type))
		if key.IsLoginKey {
			displayName += " ✓"
		}

		keyOptions = append(keyOptions, huh.NewOption(displayName, key.PrivateKeyPath))
	}

	return keyOptions
}

// buildAddRemoteForm creates the add remote form
func (m *GitManagementModel) buildAddRemoteForm() *huh.Form {
	// Build user options
	var userOptions []huh.Option[string]
	for _, user := range m.availableUsers {
		userOptions = append(userOptions, huh.NewOption(user, user))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("remoteUser").
				Title("Select User").
				Description("Git operations will use this user's SSH keys").
				Options(userOptions...).
				Value(&m.remoteUser),

			huh.NewInput().
				Key("remoteURL").
				Title("Git Remote URL").
				Description("SSH URL (e.g., git@github.com:user/repo.git)").
				Placeholder("Paste or type git remote URL...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("remote URL cannot be empty")
					}
					if !strings.Contains(s, "git@") && !strings.Contains(s, "https://") {
						return fmt.Errorf("invalid URL format. Use SSH (git@...) or HTTPS (https://...)")
					}
					return nil
				}).
				Value(&m.remoteURL),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// buildCloneForm creates the clone repository form
func (m *GitManagementModel) buildCloneForm() *huh.Form {
	// Build user options
	var userOptions []huh.Option[string]
	for _, user := range m.availableUsers {
		userOptions = append(userOptions, huh.NewOption(user, user))
	}

	// Set default user if not set
	if m.cloneUser == "" && len(m.availableUsers) > 0 {
		m.cloneUser = m.availableUsers[0]
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("cloneUser").
				Title("Select User").
				Description("Git clone will use this user's SSH keys").
				Options(userOptions...).
				Value(&m.cloneUser),

			huh.NewInput().
				Key("cloneURL").
				Title("Git Repository URL").
				Description("SSH URL (e.g., git@github.com:user/repo.git)").
				Placeholder("Paste or type git repository URL...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("repository URL cannot be empty")
					}
					if !strings.Contains(s, "git@") && !strings.Contains(s, "https://") {
						return fmt.Errorf("invalid URL format. Use SSH (git@...) or HTTPS (https://...)")
					}
					return nil
				}).
				Value(&m.cloneURL),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// prepareAndClone checks folder permissions and changes ownership if needed, then clones
func (m GitManagementModel) prepareAndClone() (tea.Model, tea.Cmd) {
	if m.cloneUser == "" || m.cloneURL == "" {
		m.state = GitStateMenu
		m.err = fmt.Errorf("user and repository URL are required")
		m.cloneForm = nil
		return m, nil
	}

	// Check if directory exists
	info, err := os.Stat(m.currentDir)
	if err != nil {
		m.state = GitStateMenu
		m.err = fmt.Errorf("directory does not exist: %s", m.currentDir)
		m.cloneForm = nil
		return m, nil
	}

	if !info.IsDir() {
		m.state = GitStateMenu
		m.err = fmt.Errorf("path is not a directory: %s", m.currentDir)
		m.cloneForm = nil
		return m, nil
	}

	// Now execute the clone using the execution screen for animation
	return m.executeClone()
}

// executeClone performs the git clone operation with progress display
func (m GitManagementModel) executeClone() (tea.Model, tea.Cmd) {
	if m.cloneUser == "" || m.cloneURL == "" {
		m.state = GitStateMenu
		m.err = fmt.Errorf("user and repository URL are required")
		m.cloneForm = nil
		return m, nil
	}

	// Build a script that starts ssh-agent, adds the keys, and clones the repo
	// After cloning, set proper permissions for web server access
	// Run the clone as the specified user
	script := fmt.Sprintf(`
echo "══════════════════════════════════════════════════════════"
echo "  Git Clone"
echo "══════════════════════════════════════════════════════════"
echo ""
echo "  Repository:  %s"
echo "  Directory:   %s"
echo "  User:        %s"
echo ""
echo "══════════════════════════════════════════════════════════"

TARGET_DIR="%s"
CLONE_USER="%s"
CLONE_URL="%s"

# Ensure target directory exists
if [ ! -d "$TARGET_DIR" ]; then
    echo ""
    echo "  ✗ Error: Directory does not exist"
    exit 1
fi

echo ""
echo "  [1/4] Setting up SSH authentication..."

sudo -i -u "$CLONE_USER" bash << 'EOF'
cd "$TARGET_DIR"

# Start ssh-agent
eval $(ssh-agent -s) > /dev/null 2>&1

# Function to add keys safely
add_ssh_keys() {
    local added=0
    for key in ~/.ssh/id_ed25519 ~/.ssh/id_rsa ~/.ssh/id_ecdsa ~/.ssh/id_dsa ~/.ssh/id_*; do
        if [[ -f "$key" && ! "$key" == *.pub && ! "$key" == *known_hosts* && ! "$key" == *config* && ! "$key" == *authorized_keys* ]]; then
            if head -n 1 "$key" | grep -q "PRIVATE KEY" 2>/dev/null; then
                ssh-add "$key" 2>/dev/null && ((added++))
            fi
        fi
    done
    if [[ $added -gt 0 ]]; then
        echo "        ✓ Loaded $added SSH key(s)"
    fi
}

add_ssh_keys

echo ""
echo "  [2/4] Cloning repository..."
echo ""

git clone --progress "$CLONE_URL" . 2>&1
CLONE_EXIT=$?

ssh-agent -k > /dev/null 2>&1 || true
exit $CLONE_EXIT
EOF

CLONE_EXIT=$?

if [ $CLONE_EXIT -eq 0 ]; then
    echo ""
    echo "        ✓ Repository cloned successfully"
    echo ""
    echo "  [3/4] Setting ownership..."
    
    WEB_GROUP="www-data"
    if getent group "$WEB_GROUP" > /dev/null 2>&1; then
        chown -R "$CLONE_USER:$WEB_GROUP" "$TARGET_DIR"
        echo "        ✓ Ownership set to $CLONE_USER:$WEB_GROUP"
    else
        chown -R "$CLONE_USER:$CLONE_USER" "$TARGET_DIR"
        echo "        ✓ Ownership set to $CLONE_USER:$CLONE_USER"
    fi
    
    echo ""
    echo "  [4/4] Setting permissions..."
    
    find "$TARGET_DIR" -type d -exec chmod 755 {} \;
    find "$TARGET_DIR" -type f -exec chmod 644 {} \;
    find "$TARGET_DIR" -type f -name "*.sh" -exec chmod 755 {} \; 2>/dev/null || true
    echo "        ✓ Base permissions set (755/644)"
    
    # Laravel specific
    if [ -d "$TARGET_DIR/storage" ]; then
        chmod -R 775 "$TARGET_DIR/storage"
        chmod -R 775 "$TARGET_DIR/bootstrap/cache" 2>/dev/null || true
        echo "        ✓ Laravel writable directories configured"
    fi
    
    # WordPress specific
    if [ -d "$TARGET_DIR/wp-content" ]; then
        chmod -R 775 "$TARGET_DIR/wp-content"
        echo "        ✓ WordPress wp-content configured"
    fi
    
    echo ""
    echo "  [5/5] Setting system user for git operations..."
    
    # Set meta.systemuser config so future git operations use this user
    cd "$TARGET_DIR"
    git config meta.systemuser "$CLONE_USER"
    echo "        ✓ System user set to '$CLONE_USER'"
    
    echo ""
    echo "══════════════════════════════════════════════════════════"
    echo "  ✓ Clone completed successfully!"
    echo "══════════════════════════════════════════════════════════"
    echo ""
else
    echo ""
    echo "══════════════════════════════════════════════════════════"
    echo "  ✗ Clone failed (exit code: $CLONE_EXIT)"
    echo "══════════════════════════════════════════════════════════"
    echo ""
    exit $CLONE_EXIT
fi
`, m.cloneURL, m.currentDir, m.cloneUser, m.currentDir, m.cloneUser, m.cloneURL)

	m.state = GitStateMenu
	m.cloneForm = nil

	// Use ExecutionStartMsg to show the clone progress in the execution screen
	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     script,
			Description: fmt.Sprintf("Cloning %s", m.cloneURL),
		}
	}
}

// runTestConnection runs the SSH test connection
func (m GitManagementModel) runTestConnection() (tea.Model, tea.Cmd) {
	selectedUser := m.testForm.GetString("selectedUser")
	selectedKey := m.testForm.GetString("selectedKey")
	m.testForm = nil

	if selectedUser == "" {
		m.state = GitStateMenu
		m.err = fmt.Errorf("no user selected")
		return m, nil
	}

	// Build a script that starts ssh-agent, adds the key, and tests the connection
	script := fmt.Sprintf(`
echo "Testing connection to GitHub for user %s..."

sudo -i -u %s bash << 'EOF'
# Start ssh-agent
eval $(ssh-agent -s) > /dev/null 2>&1

# Function to add keys safely
add_ssh_keys() {
    local added=0
    # Search for private keys in ~/.ssh
    # We look for files that don't end in .pub and are actually files
    for key in ~/.ssh/id_ed25519 ~/.ssh/id_rsa ~/.ssh/id_ecdsa ~/.ssh/id_dsa ~/.ssh/id_*; do
        if [[ -f "$key" && ! "$key" == *.pub && ! "$key" == *known_hosts* && ! "$key" == *config* && ! "$key" == *authorized_keys* ]]; then
            # Check if it looks like a private key (has a header)
            if head -n 1 "$key" | grep -q "PRIVATE KEY" 2>/dev/null; then
                ssh-add "$key" 2>/dev/null && ((added++))
            fi
        fi
    done
    
    if [[ $added -gt 0 ]]; then
        echo "✓ Loaded $added SSH key(s)"
    fi
}

add_ssh_keys

# Test connection
echo ""
echo "Executing: ssh -T git@github.com"
ssh -o StrictHostKeyChecking=accept-new -o BatchMode=yes -T git@github.com 2>&1
EXIT_CODE=$?

# Cleanup
ssh-agent -k > /dev/null 2>&1 || true

exit $EXIT_CODE
EOF
`, selectedUser, selectedUser)

	cmd := exec.Command("bash", "-c", script)
	output, err := cmd.CombinedOutput()

	outputStr := strings.TrimSpace(string(output))

	// Build key info for display
	keyInfo := "Auto-detect"
	if selectedKey != "auto" && selectedKey != "" {
		parts := strings.Split(selectedKey, "/")
		if len(parts) > 0 {
			keyInfo = parts[len(parts)-1]
		}
	}

	// GitHub returns exit code 1 even on success (it says "Hi username!")
	// So we check the output content instead of error
	if strings.Contains(outputStr, "successfully authenticated") ||
		strings.Contains(outputStr, "Hi ") ||
		strings.Contains(outputStr, "You've successfully authenticated") {
		m.success = fmt.Sprintf("✓ SSH Connection Successful!\n\nUser: %s\nKey: %s\n\nResponse: %s", selectedUser, keyInfo, outputStr)
	} else if strings.Contains(outputStr, "Permission denied") ||
		strings.Contains(outputStr, "publickey") {
		m.err = fmt.Errorf("SSH Connection Failed\n\nUser: %s\nKey: %s\n\n%s\n\nTroubleshooting:\n• Check if SSH key exists for this user\n• Verify key is added to GitHub/GitLab\n• Make sure the key has login enabled", selectedUser, keyInfo, outputStr)
	} else if strings.Contains(outputStr, "Could not resolve") ||
		strings.Contains(outputStr, "Network is unreachable") {
		m.err = fmt.Errorf("Network Error\n\n%s\n\nCheck your internet connection", outputStr)
	} else if err != nil {
		m.err = fmt.Errorf("Connection test failed\n\nUser: %s\nKey: %s\n\n%s", selectedUser, keyInfo, outputStr)
	} else {
		m.success = fmt.Sprintf("Connection test completed\n\nUser: %s\nKey: %s\n\nResponse:\n%s", selectedUser, keyInfo, outputStr)
	}

	m.state = GitStateMenu
	return m, nil
}

// setupGitRemote sets up the git remote
func (m GitManagementModel) setupGitRemote() (tea.Model, tea.Cmd) {
	remoteName := "origin"
	if m.gitInfo.RemoteName != "" {
		remoteName = m.gitInfo.RemoteName
	}

	var cmd *exec.Cmd
	if m.gitInfo.RemoteURL == "" {
		// Add new remote
		cmd = exec.Command("git", "remote", "add", remoteName, m.remoteURL)
	} else {
		// Change existing remote
		cmd = exec.Command("git", "remote", "set-url", remoteName, m.remoteURL)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		m.err = fmt.Errorf("%s", stderr.String())
	} else {
		if m.gitInfo.RemoteURL == "" {
			m.success = fmt.Sprintf("✓ Remote '%s' added successfully with URL: %s", remoteName, m.remoteURL)
		} else {
			m.success = fmt.Sprintf("✓ Remote '%s' URL updated to: %s", remoteName, m.remoteURL)
		}
		m.gitInfo = getGitInfo()
	}

	m.state = GitStateMenu
	m.remoteForm = nil
	return m, nil
}

// executeAction executes the selected git action
func (m GitManagementModel) executeAction() (tea.Model, tea.Cmd) {
	m.err = nil
	m.success = ""

	action := m.actions[m.cursor]

	switch action.ID {
	case "refresh":
		m.gitInfo = getGitInfo()
		m.currentDir, _ = os.Getwd()
		m.success = "✓ Git info refreshed"

	case "fix_ownership":
		cmd := exec.Command("git", "config", "--global", "--add", "safe.directory", m.currentDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			m.err = fmt.Errorf("failed to fix ownership: %s", strings.TrimSpace(string(output)))
		} else {
			m.success = "✓ Directory added to safe.directory. Git info refreshed."
			// Rebuild the whole model state to refresh actions and info
			newModel := NewGitManagementModel()
			m.gitInfo = newModel.gitInfo
			m.actions = newModel.actions
			m.cursor = 0
		}

	case "test_connection":
		// Show user selection form
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		m.state = GitStateTestConnectionForm
		m.testForm = m.buildTestConnectionForm()
		return m, m.testForm.Init()

	case "clone_repo":
		// Check if already a git repo
		if m.gitInfo.IsRepo {
			m.err = fmt.Errorf("this directory is already a Git repository")
			return m, nil
		}
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		m.state = GitStateCloneForm
		m.cloneForm = m.buildCloneForm()
		return m, m.cloneForm.Init()

	case "add_remote":
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		m.state = GitStateAddRemoteForm
		m.remoteForm = m.buildAddRemoteForm()
		return m, m.remoteForm.Init()

	case "change_remote":
		if m.gitInfo.RemoteURL == "" {
			m.err = fmt.Errorf("no remote to change. Use 'Add/Setup Git Remote' first")
			return m, nil
		}
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		// Use system user if configured, otherwise show system user setting form
		if m.gitInfo.SystemUser != "" {
			m.gitOpUser = m.gitInfo.SystemUser
			m.gitOpAction = "change_remote"
			return m.executeGitOp()
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "remove_remote":
		if m.gitInfo.RemoteName == "" {
			m.err = fmt.Errorf("no remote configured")
			return m, nil
		}
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		// Use system user if configured, otherwise show system user setting form
		if m.gitInfo.SystemUser != "" {
			m.gitOpUser = m.gitInfo.SystemUser
			m.gitOpAction = "remove_remote"
			return m.executeGitOp()
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "git_pull":
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		// Use system user if configured, otherwise show system user setting form
		if m.gitInfo.SystemUser != "" {
			m.gitOpUser = m.gitInfo.SystemUser
			m.gitOpAction = "git_pull"
			return m.executeGitOp()
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "git_fetch":
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		// Use system user if configured, otherwise show system user setting form
		if m.gitInfo.SystemUser != "" {
			m.gitOpUser = m.gitInfo.SystemUser
			m.gitOpAction = "git_fetch"
			return m.executeGitOp()
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "git_status":
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		// Use system user if configured, otherwise show system user setting form
		if m.gitInfo.SystemUser != "" {
			m.gitOpUser = m.gitInfo.SystemUser
			m.gitOpAction = "git_status"
			return m.executeGitOp()
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "set_system_user":
		if !m.gitInfo.IsRepo {
			m.err = fmt.Errorf("not a git repository")
			return m, nil
		}
		if len(m.availableUsers) == 0 {
			m.err = fmt.Errorf("no users available")
			return m, nil
		}
		m.state = GitStateSetSystemUserForm
		m.systemUserForm = m.buildSetSystemUserForm()
		return m, m.systemUserForm.Init()

	case "back":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: SiteCommandsScreen}
		}
	}

	return m, nil
}

// View renders the git management screen
func (m GitManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render based on state
	switch m.state {
	case GitStateTestConnectionForm:
		return m.renderTestConnectionForm()
	case GitStateCloneForm:
		return m.renderCloneForm()
	case GitStateConfirmClone:
		return m.renderConfirmClone()
	case GitStateAddRemoteForm:
		return m.renderAddRemoteForm()
	case GitStateConfirmRemote:
		return m.renderConfirmRemote()
	case GitStateGitOpForm:
		return m.renderGitOpForm()
	case GitStateSetSystemUserForm:
		return m.renderSetSystemUserForm()
	default:
		return m.renderMenu()
	}
}

// renderMenu renders the main menu view
func (m GitManagementModel) renderMenu() string {
	// Header
	header := m.theme.Title.Render("Git Operations")

	// Current directory info
	dirInfo := m.theme.Label.Render("Directory: ") + m.theme.InfoStyle.Render(m.currentDir)

	// Git repository info
	var infoLines []string
	infoLines = append(infoLines, dirInfo)
	infoLines = append(infoLines, "")

	if !m.gitInfo.IsRepo {
		if m.gitInfo.DubiousOwnership {
			infoLines = append(infoLines, m.theme.ErrorStyle.Render("⚠ Git repository detected with dubious ownership"))
			infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Git refuses to work in this folder because it's owned by another user."))
			infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  To fix, run: git config --global --add safe.directory "+m.currentDir))
			infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Or run ravact with 'sudo' to manage this repository."))
		} else {
			infoLines = append(infoLines, m.theme.WarningStyle.Render("⚠ Not a Git repository"))
			infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Navigate to a directory with a Git repository"))
		}
	} else {
		// Branch
		branchLabel := m.theme.Label.Render("Branch: ")
		branchValue := m.theme.SuccessStyle.Render(m.gitInfo.Branch)
		infoLines = append(infoLines, branchLabel+branchValue)

		// Remote
		remoteLabel := m.theme.Label.Render("Remote: ")
		if m.gitInfo.RemoteURL != "" {
			remoteValue := m.theme.InfoStyle.Render(fmt.Sprintf("%s (%s)", m.gitInfo.RemoteURL, m.gitInfo.RemoteName))
			infoLines = append(infoLines, remoteLabel+remoteValue)
		} else {
			infoLines = append(infoLines, remoteLabel+m.theme.WarningStyle.Render("No remote configured"))
		}

		// Last commit
		if m.gitInfo.LastCommit != "" {
			commitLabel := m.theme.Label.Render("Last Commit: ")
			commitValue := m.theme.KeyStyle.Render(m.gitInfo.LastCommit) + " " + m.theme.DescriptionStyle.Render(m.gitInfo.CommitMsg)
			infoLines = append(infoLines, commitLabel+commitValue)
		}

		// Status indicators
		var statusParts []string
		if m.gitInfo.HasChanges {
			statusParts = append(statusParts, m.theme.WarningStyle.Render("● Uncommitted changes"))
		}
		if m.gitInfo.Ahead > 0 {
			statusParts = append(statusParts, m.theme.SuccessStyle.Render(fmt.Sprintf("↑ %d ahead", m.gitInfo.Ahead)))
		}
		if m.gitInfo.Behind > 0 {
			statusParts = append(statusParts, m.theme.ErrorStyle.Render(fmt.Sprintf("↓ %d behind", m.gitInfo.Behind)))
		}
		if len(statusParts) > 0 {
			infoLines = append(infoLines, m.theme.Label.Render("Status: ")+strings.Join(statusParts, " • "))
		} else if !m.gitInfo.HasChanges {
			infoLines = append(infoLines, m.theme.Label.Render("Status: ")+m.theme.SuccessStyle.Render("✓ Clean working tree"))
		}

		// System user
		sysUserLabel := m.theme.Label.Render("System User: ")
		if m.gitInfo.SystemUser != "" {
			infoLines = append(infoLines, sysUserLabel+m.theme.SuccessStyle.Render(m.gitInfo.SystemUser))
		} else {
			infoLines = append(infoLines, sysUserLabel+m.theme.WarningStyle.Render("Not set (will prompt on git operations)"))
		}
	}

	infoSection := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Actions menu
	var actionItems []string
	actionItems = append(actionItems, m.theme.Subtitle.Render("Actions:"))
	actionItems = append(actionItems, "")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		}

		actionItems = append(actionItems, renderedItem)
	}

	actionsMenu := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	// Messages
	var messages []string
	if m.success != "" {
		messages = append(messages, m.theme.SuccessStyle.Render(m.success))
	}
	if m.err != nil {
		messages = append(messages, m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	// Combine all sections
	sections := []string{
		header,
		"",
		infoSection,
		"",
		actionsMenu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add border and center
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderTestConnectionForm renders the test connection form
func (m GitManagementModel) renderTestConnectionForm() string {
	header := m.theme.Title.Render("Test Git Connection")

	description := m.theme.DescriptionStyle.Render("Select a user to test SSH connection to GitHub.\nThis will run: ssh -T git@github.com")

	formView := ""
	if m.testForm != nil {
		formView = m.testForm.View()
	}

	help := m.theme.Help.Render("Tab: Next • Enter: Submit • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		description,
		"",
		formView,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderCloneForm renders the clone repository form
func (m GitManagementModel) renderCloneForm() string {
	header := m.theme.Title.Render("Clone Git Repository")

	// Show current directory
	dirInfo := m.theme.Label.Render("Directory: ") + m.theme.InfoStyle.Render(m.currentDir)

	description := m.theme.DescriptionStyle.Render("Clone a repository into the current directory.\nThe directory should be empty.")

	formView := ""
	if m.cloneForm != nil {
		formView = m.cloneForm.View()
	}

	help := m.theme.Help.Render("Tab: Next • Enter: Submit • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		dirInfo,
		"",
		description,
		"",
		formView,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderConfirmClone renders the clone confirmation screen
func (m GitManagementModel) renderConfirmClone() string {
	header := m.theme.Title.Render("Confirm Git Clone")

	// Check current directory ownership
	statCmd := exec.Command("stat", "-c", "%U:%G", m.currentDir)
	statOutput, _ := statCmd.Output()
	currentOwner := strings.TrimSpace(string(statOutput))
	expectedOwner := fmt.Sprintf("%s:www-data", m.cloneUser)
	needsOwnershipChange := currentOwner != expectedOwner && currentOwner != ""

	// Summary
	var summaryLines []string
	summaryLines = append(summaryLines, m.theme.Label.Render("Directory:   ")+m.theme.InfoStyle.Render(m.currentDir))
	summaryLines = append(summaryLines, m.theme.Label.Render("User:        ")+m.theme.InfoStyle.Render(m.cloneUser))
	summaryLines = append(summaryLines, m.theme.Label.Render("Repository:  ")+m.theme.SuccessStyle.Render(m.cloneURL))

	if needsOwnershipChange {
		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, m.theme.Label.Render("Current Owner: ")+m.theme.WarningStyle.Render(currentOwner))
	}

	summary := lipgloss.JoinVertical(lipgloss.Left, summaryLines...)

	var warning string
	if needsOwnershipChange {
		warning = m.theme.WarningStyle.Render("\n⚠ Directory ownership will be changed to " + m.cloneUser + ":www-data\n  for web server compatibility.")
	} else {
		warning = m.theme.WarningStyle.Render("\n⚠ This will clone the repository contents into the current directory.")
	}

	question := m.theme.Label.Render("\nProceed with clone?")

	help := m.theme.Help.Render("y/Enter: Confirm • n/Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		warning,
		question,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderGitOpForm renders the git operation form
func (m GitManagementModel) renderGitOpForm() string {
	// Get title based on action
	title := "Git Operation"
	switch m.gitOpAction {
	case "git_pull":
		title = "Git Pull"
	case "git_fetch":
		title = "Git Fetch"
	case "git_status":
		title = "Git Status"
	case "change_remote":
		title = "Change Remote URL"
	case "remove_remote":
		title = "Remove Remote"
	}

	header := m.theme.Title.Render(title)

	// Show current directory
	dirInfo := m.theme.Label.Render("Directory: ") + m.theme.InfoStyle.Render(m.currentDir)

	description := m.theme.DescriptionStyle.Render("Select the user to run this operation as.")

	formView := ""
	if m.gitOpForm != nil {
		formView = m.gitOpForm.View()
	}

	help := m.theme.Help.Render("Enter: Submit • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		dirInfo,
		"",
		description,
		"",
		formView,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderSetSystemUserForm renders the set system user form
func (m GitManagementModel) renderSetSystemUserForm() string {
	header := m.theme.Title.Render("Set System User")

	// Show current directory
	dirInfo := m.theme.Label.Render("Directory: ") + m.theme.InfoStyle.Render(m.currentDir)

	// Current system user if set
	currentUser := ""
	if m.gitInfo.SystemUser != "" {
		currentUser = m.theme.Label.Render("Current: ") + m.theme.InfoStyle.Render(m.gitInfo.SystemUser)
	}

	description := m.theme.DescriptionStyle.Render("Select the user for git operations in this repository.\nThis user will be used for pull, fetch, status, etc.")

	formView := ""
	if m.systemUserForm != nil {
		formView = m.systemUserForm.View()
	}

	help := m.theme.Help.Render("Enter: Submit • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	sections := []string{header, "", dirInfo}
	if currentUser != "" {
		sections = append(sections, currentUser)
	}
	sections = append(sections, "", description, "", formView, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderAddRemoteForm renders the add remote form
func (m GitManagementModel) renderAddRemoteForm() string {
	header := m.theme.Title.Render("Add/Setup Git Remote")

	// Show current directory
	dirInfo := m.theme.Label.Render("Directory: ") + m.theme.InfoStyle.Render(m.currentDir)

	description := m.theme.DescriptionStyle.Render("Configure git remote URL for this repository.\nYou can paste the URL directly from GitHub/GitLab.")

	formView := ""
	if m.remoteForm != nil {
		formView = m.remoteForm.View()
	}

	help := m.theme.Help.Render("Tab: Next • Enter: Submit • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		dirInfo,
		"",
		description,
		"",
		formView,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderConfirmRemote renders the confirmation screen
func (m GitManagementModel) renderConfirmRemote() string {
	header := m.theme.Title.Render("Confirm Git Remote Setup")

	// Summary
	var summaryLines []string
	summaryLines = append(summaryLines, m.theme.Label.Render("Directory:   ")+m.theme.InfoStyle.Render(m.currentDir))
	summaryLines = append(summaryLines, m.theme.Label.Render("User:        ")+m.theme.InfoStyle.Render(m.remoteUser))
	summaryLines = append(summaryLines, m.theme.Label.Render("Remote URL:  ")+m.theme.SuccessStyle.Render(m.remoteURL))

	if m.gitInfo.RemoteURL != "" {
		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, m.theme.WarningStyle.Render("⚠ This will replace the existing remote URL:"))
		summaryLines = append(summaryLines, m.theme.DescriptionStyle.Render("  "+m.gitInfo.RemoteURL))
	}

	summary := lipgloss.JoinVertical(lipgloss.Left, summaryLines...)

	question := m.theme.Label.Render("\nProceed with this configuration?")

	help := m.theme.Help.Render("y/Enter: Confirm • n/Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		question,
		"",
		help,
	)

	paddedContent := lipgloss.NewStyle().
		Padding(paddingV, paddingH).
		Render(content)

	bordered := m.theme.RenderBox(paddedContent)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
