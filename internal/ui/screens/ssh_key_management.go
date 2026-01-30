package screens

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SSHKeyManagementState represents the current state of the SSH key management screen
type SSHKeyManagementState int

const (
	SSHKeyStateList SSHKeyManagementState = iota
	SSHKeyStateGenerateForm
	SSHKeyStateKeyDetails
	SSHKeyStateConfirmDelete
	SSHKeyStateCopyKey
	SSHKeyStateExportOptions
)

// SSHKeyManagementModel represents the SSH key management screen
type SSHKeyManagementModel struct {
	theme       *theme.Theme
	width       int
	height      int
	username    string
	userManager *system.UserManager

	// State
	state        SSHKeyManagementState
	keys         []system.SSHKey
	cursor       int
	actionCursor int
	exportCursor int // cursor for export options
	err          error
	message      string
	copyableKey  string // Public key content for mouse copying

	// Key generation form
	form           *huh.Form
	keyType        string
	keyIdentifier  string
	keyEmail       string
	keyPassphrase  string
	addToAgent     bool
	useForLogin    bool

	// Currently selected key for details
	selectedKey *system.SSHKey
}

// NewSSHKeyManagementModel creates a new SSH key management model
func NewSSHKeyManagementModel(username string) SSHKeyManagementModel {
	t := theme.DefaultTheme()
	um := system.NewUserManager()

	m := SSHKeyManagementModel{
		theme:       t,
		username:    username,
		userManager: um,
		state:       SSHKeyStateList,
		cursor:      0,
		actionCursor: 0,
		keyType:     "ed25519",
	}

	// Load keys
	m.loadKeys()

	return m
}

// loadKeys loads SSH keys for the user
func (m *SSHKeyManagementModel) loadKeys() {
	keys, err := m.userManager.GetUserSSHKeys(m.username)
	if err != nil {
		m.err = err
		return
	}
	m.keys = keys
}

// buildGenerateFormWithAccessors creates the key generation form with accessor functions
func (m *SSHKeyManagementModel) buildGenerateFormWithAccessors() *huh.Form {
	// Use Key() to set accessor functions that read/write to model fields
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("keyType").
				Title("Key Type").
				Description("Select the SSH key algorithm").
				Options(
					huh.NewOption("ED25519 (Recommended)", "ed25519"),
					huh.NewOption("RSA 4096-bit", "rsa"),
					huh.NewOption("ECDSA", "ecdsa"),
				).
				Value(&m.keyType),

			huh.NewInput().
				Key("keyIdentifier").
				Title("Key Name").
				Description("A name for this key file (e.g., 'server-deploy', 'github')").
				Placeholder("Enter key name...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("key name cannot be empty")
					}
					if len(s) < 2 {
						return fmt.Errorf("key name must be at least 2 characters")
					}
					return nil
				}).
				Value(&m.keyIdentifier),

			huh.NewInput().
				Key("keyEmail").
				Title("Email").
				Description("Email address for key comment (e.g., 'user@example.com')").
				Placeholder("Enter email...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("email cannot be empty")
					}
					if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
						return fmt.Errorf("please enter a valid email address")
					}
					return nil
				}).
				Value(&m.keyEmail),

			huh.NewInput().
				Key("keyPassphrase").
				Title("Passphrase (Optional)").
				Description("Leave empty for no passphrase").
				Placeholder("Enter passphrase...").
				EchoMode(huh.EchoModePassword).
				Value(&m.keyPassphrase),

			huh.NewConfirm().
				Key("addToAgent").
				Title("Add to SSH Agent?").
				Description("Automatically add the key to ssh-agent after generation").
				Affirmative("Yes").
				Negative("No").
				Value(&m.addToAgent),

			huh.NewConfirm().
				Key("useForLogin").
				Title("Use for Login?").
				Description("Add to authorized_keys to allow SSH login with this key").
				Affirmative("Yes").
				Negative("No").
				Value(&m.useForLogin),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// Init initializes the SSH key management screen
func (m SSHKeyManagementModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for SSH key management
func (m SSHKeyManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update form first if in generate state (before handling key messages)
	if m.state == SSHKeyStateGenerateForm && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		// Check if form is completed
		if m.form.State == huh.StateCompleted {
			return m.generateKey()
		}

		// Handle special keys for form state
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = SSHKeyStateList
				m.form = nil
				return m, nil
			}
		}

		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle message clearing
		if m.message != "" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.message = ""
				return m, nil
			default:
				m.message = ""
				return m, nil
			}
		}

		// Handle error clearing
		if m.err != nil {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.err = nil
				return m, nil
			default:
				m.err = nil
				return m, nil
			}
		}

		// Handle different states
		switch m.state {
		case SSHKeyStateList:
			return m.updateList(msg)
		case SSHKeyStateKeyDetails:
			return m.updateKeyDetails(msg)
		case SSHKeyStateConfirmDelete:
			return m.updateConfirmDelete(msg)
		case SSHKeyStateCopyKey:
			return m.updateCopyKey(msg)
		case SSHKeyStateExportOptions:
			return m.updateExportOptions(msg)
		}
	}

	return m, nil
}

// updateList handles key presses in the list view
func (m SSHKeyManagementModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalItems := len(m.keys) + 1 // +1 for "Generate New Key" option

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "backspace":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: UserDetailsScreen, Data: m.username}
		}

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < totalItems-1 {
			m.cursor++
		}

	case "enter", " ":
		if m.cursor == 0 {
			// Generate new key - reset form values
			m.state = SSHKeyStateGenerateForm
			m.keyType = "ed25519"
			m.keyIdentifier = ""
			m.keyPassphrase = ""
			m.addToAgent = false
			m.useForLogin = true
			// Build form with fresh local variables that will be read on completion
			m.form = m.buildGenerateFormWithAccessors()
			return m, m.form.Init()
		} else if m.cursor > 0 && m.cursor <= len(m.keys) {
			// View key details
			key := m.keys[m.cursor-1]
			m.selectedKey = &key
			m.state = SSHKeyStateKeyDetails
			m.actionCursor = 0
		}
	}

	return m, nil
}

// updateKeyDetails handles key presses in the key details view
func (m SSHKeyManagementModel) updateKeyDetails(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := m.getKeyActions()

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "backspace":
		m.state = SSHKeyStateList
		m.selectedKey = nil
		return m, nil

	case "up", "k":
		if m.actionCursor > 0 {
			m.actionCursor--
		}

	case "down", "j":
		if m.actionCursor < len(actions)-1 {
			m.actionCursor++
		}

	case "c", "C":
		// Shortcut for copy/view public key
		return m.openKeyInEditor()

	case "e", "E":
		// Shortcut for export private key (only if login is enabled)
		if m.selectedKey != nil && m.selectedKey.IsLoginKey {
			return m.showExportOptions()
		}

	case "enter", " ":
		return m.executeKeyAction(actions[m.actionCursor])
	}

	return m, nil
}

// updateCopyKey handles key presses in the copy key view
func (m SSHKeyManagementModel) updateCopyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace", "enter", " ":
		m.state = SSHKeyStateKeyDetails
		m.copyableKey = ""
		return m, nil
	}
	return m, nil
}

// updateExportOptions handles key presses in the export options view
func (m SSHKeyManagementModel) updateExportOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	exportOptions := []string{"Linux/macOS (PEM format)", "Windows PuTTY (PPK format)"}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "backspace":
		m.state = SSHKeyStateKeyDetails
		m.exportCursor = 0
		return m, nil

	case "up", "k":
		if m.exportCursor > 0 {
			m.exportCursor--
		}

	case "down", "j":
		if m.exportCursor < len(exportOptions)-1 {
			m.exportCursor++
		}

	case "enter", " ":
		if m.exportCursor == 0 {
			// Linux/macOS PEM format
			return m.exportPrivateKeyPEM()
		} else {
			// Windows PuTTY PPK format
			return m.exportPrivateKeyPPK()
		}
	}

	return m, nil
}

// showExportOptions shows the export format selection
func (m SSHKeyManagementModel) showExportOptions() (tea.Model, tea.Cmd) {
	m.state = SSHKeyStateExportOptions
	m.exportCursor = 0
	return m, nil
}

// exportPrivateKeyPEM exports the private key in PEM format (Linux/macOS)
func (m SSHKeyManagementModel) exportPrivateKeyPEM() (tea.Model, tea.Cmd) {
	if m.selectedKey == nil {
		return m, nil
	}

	// Show the private key in less for viewing/copying
	copyScript := fmt.Sprintf(`
#!/bin/bash
KEY_CONTENT=$(cat "%s")
COPIED=0
STATUS_MSG=""

# Try different clipboard tools (suppress errors for headless environments)
if command -v xclip &> /dev/null; then
    if echo -n "$KEY_CONTENT" | xclip -selection clipboard 2>/dev/null; then
        STATUS_MSG="✓ Private key (PEM) copied to clipboard"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ] && command -v xsel &> /dev/null; then
    if echo -n "$KEY_CONTENT" | xsel --clipboard --input 2>/dev/null; then
        STATUS_MSG="✓ Private key (PEM) copied to clipboard"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ] && command -v pbcopy &> /dev/null; then
    if echo -n "$KEY_CONTENT" | pbcopy 2>/dev/null; then
        STATUS_MSG="✓ Private key (PEM) copied to clipboard"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ]; then
    STATUS_MSG="ℹ  No display for clipboard. Hold SHIFT + select with mouse to copy."
fi

# Create temp file with content to show in less
TMPFILE=$(mktemp)
echo "$STATUS_MSG" > "$TMPFILE"
echo "" >> "$TMPFILE"
echo "Private Key - PEM Format (for Linux/macOS)" >> "$TMPFILE"
echo "Press 'q' to go back" >> "$TMPFILE"
echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"
echo "$KEY_CONTENT" >> "$TMPFILE"
echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"
echo "" >> "$TMPFILE"
echo "⚠  SECURITY WARNING: Keep this private key secure!" >> "$TMPFILE"
echo "   Save as: ~/.ssh/id_key (with permissions 600)" >> "$TMPFILE"

# Show in less - user presses 'q' to quit
less -R "$TMPFILE"

# Cleanup
rm -f "$TMPFILE"
`, m.selectedKey.PrivateKeyPath)

	c := exec.Command("bash", "-c", copyScript)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return nil
	})
}

// exportPrivateKeyPPK exports the private key in PPK format (Windows PuTTY)
func (m SSHKeyManagementModel) exportPrivateKeyPPK() (tea.Model, tea.Cmd) {
	if m.selectedKey == nil {
		return m, nil
	}

	// Convert to PPK format using puttygen if available, otherwise show instructions
	copyScript := fmt.Sprintf(`
#!/bin/bash
PRIV_KEY="%s"
TMPFILE=$(mktemp)

# Check if puttygen is available
if command -v puttygen &> /dev/null; then
    # Convert to PPK format
    PPK_CONTENT=$(puttygen "$PRIV_KEY" -O private -o /dev/stdout 2>/dev/null)
    
    if [ -n "$PPK_CONTENT" ]; then
        COPIED=0
        STATUS_MSG=""
        
        # Try to copy to clipboard
        if command -v xclip &> /dev/null; then
            if echo -n "$PPK_CONTENT" | xclip -selection clipboard 2>/dev/null; then
                STATUS_MSG="✓ Private key (PPK) copied to clipboard"
                COPIED=1
            fi
        fi
        
        if [ $COPIED -eq 0 ] && command -v xsel &> /dev/null; then
            if echo -n "$PPK_CONTENT" | xsel --clipboard --input 2>/dev/null; then
                STATUS_MSG="✓ Private key (PPK) copied to clipboard"
                COPIED=1
            fi
        fi
        
        if [ $COPIED -eq 0 ]; then
            STATUS_MSG="ℹ  No display for clipboard. Hold SHIFT + select with mouse to copy."
        fi
        
        echo "$STATUS_MSG" > "$TMPFILE"
        echo "" >> "$TMPFILE"
        echo "Private Key - PPK Format (for Windows PuTTY)" >> "$TMPFILE"
        echo "Press 'q' to go back" >> "$TMPFILE"
        echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"
        echo "$PPK_CONTENT" >> "$TMPFILE"
        echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"
        echo "" >> "$TMPFILE"
        echo "⚠  SECURITY WARNING: Keep this private key secure!" >> "$TMPFILE"
        echo "   Save as: key.ppk on your Windows machine" >> "$TMPFILE"
    else
        echo "Error: Failed to convert key to PPK format" > "$TMPFILE"
        echo "" >> "$TMPFILE"
        echo "The key may be passphrase protected." >> "$TMPFILE"
        echo "Try: puttygen $PRIV_KEY -o output.ppk" >> "$TMPFILE"
    fi
else
    echo "⚠  puttygen is not installed" > "$TMPFILE"
    echo "" >> "$TMPFILE"
    echo "To convert keys to PPK format, install putty-tools:" >> "$TMPFILE"
    echo "" >> "$TMPFILE"
    echo "  Ubuntu/Debian: sudo apt install putty-tools" >> "$TMPFILE"
    echo "  CentOS/RHEL:   sudo yum install putty" >> "$TMPFILE"
    echo "  Fedora:        sudo dnf install putty" >> "$TMPFILE"
    echo "" >> "$TMPFILE"
    echo "Or convert on Windows using PuTTYgen:" >> "$TMPFILE"
    echo "  1. Open PuTTYgen" >> "$TMPFILE"
    echo "  2. Click 'Load' and select the private key file" >> "$TMPFILE"
    echo "  3. Click 'Save private key' to save as .ppk" >> "$TMPFILE"
    echo "" >> "$TMPFILE"
    echo "Private key location: $PRIV_KEY" >> "$TMPFILE"
fi

# Show in less - user presses 'q' to quit
less -R "$TMPFILE"

# Cleanup
rm -f "$TMPFILE"
`, m.selectedKey.PrivateKeyPath)

	c := exec.Command("bash", "-c", copyScript)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return nil
	})
}

// updateConfirmDelete handles key presses in the confirm delete view
func (m SSHKeyManagementModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "backspace", "n", "N":
		m.state = SSHKeyStateKeyDetails
		return m, nil

	case "y", "Y":
		if m.selectedKey != nil {
			err := m.userManager.DeleteSSHKey(m.selectedKey.PublicKeyPath)
			if err != nil {
				m.err = fmt.Errorf("failed to delete key: %v", err)
			} else {
				m.message = fmt.Sprintf("%s Key '%s' deleted successfully!", m.theme.Symbols.CheckMark, m.selectedKey.Identifier)
				m.loadKeys()
				m.state = SSHKeyStateList
				m.selectedKey = nil
			}
		}
		return m, nil
	}

	return m, nil
}

// getKeyActions returns available actions for the selected key
func (m SSHKeyManagementModel) getKeyActions() []string {
	if m.selectedKey == nil {
		return []string{}
	}

	actions := []string{}

	if m.selectedKey.IsLoginKey {
		actions = append(actions, "Remove from Login Keys")
	} else {
		actions = append(actions, "Add to Login Keys")
	}

	if m.selectedKey.IsInAgent {
		actions = append(actions, "✓ Already in SSH Agent")
	} else {
		actions = append(actions, "Add to SSH Agent")
	}

	actions = append(actions, "View/Copy Public Key (c)")
	
	// Only allow export when key is enabled for login
	if m.selectedKey.IsLoginKey {
		actions = append(actions, "Export Private Key (e)")
	}
	
	actions = append(actions, "Delete Key")
	actions = append(actions, "Back to List")

	return actions
}

// executeKeyAction executes the selected action on the key
func (m SSHKeyManagementModel) executeKeyAction(action string) (tea.Model, tea.Cmd) {
	if m.selectedKey == nil {
		return m, nil
	}

	switch action {
	case "Add to Login Keys":
		err := m.userManager.AddKeyToAuthorizedKeys(m.username, m.selectedKey.PublicKeyPath)
		if err != nil {
			m.err = fmt.Errorf("failed to add to authorized_keys: %v", err)
		} else {
			m.message = fmt.Sprintf("%s Key added to authorized_keys!", m.theme.Symbols.CheckMark)
			m.loadKeys()
			// Update selected key
			for _, k := range m.keys {
				if k.PublicKeyPath == m.selectedKey.PublicKeyPath {
					m.selectedKey = &k
					break
				}
			}
		}

	case "Remove from Login Keys":
		err := m.userManager.RemoveKeyFromAuthorizedKeys(m.username, m.selectedKey.Fingerprint)
		if err != nil {
			m.err = fmt.Errorf("failed to remove from authorized_keys: %v", err)
		} else {
			m.message = fmt.Sprintf("%s Key removed from authorized_keys!", m.theme.Symbols.CheckMark)
			m.loadKeys()
			// Update selected key
			for _, k := range m.keys {
				if k.PublicKeyPath == m.selectedKey.PublicKeyPath {
					m.selectedKey = &k
					break
				}
			}
		}

	case "Add to SSH Agent":
		err := m.userManager.AddKeyToSSHAgent(m.selectedKey.PrivateKeyPath)
		if err != nil {
			m.err = fmt.Errorf("failed to add to ssh-agent: %v", err)
		} else {
			m.message = fmt.Sprintf("%s Key added to ssh-agent!", m.theme.Symbols.CheckMark)
			// Reload keys to update agent status
			m.loadKeys()
			// Update selected key
			for _, k := range m.keys {
				if k.PublicKeyPath == m.selectedKey.PublicKeyPath {
					m.selectedKey = &k
					break
				}
			}
		}

	case "✓ Already in SSH Agent":
		// Do nothing - key is already in agent
		m.message = "Key is already loaded in SSH agent"

	case "View/Copy Public Key (c)":
		return m.openKeyInEditor()

	case "Export Private Key (e)":
		return m.showExportOptions()

	case "Delete Key":
		m.state = SSHKeyStateConfirmDelete

	case "Back to List":
		m.state = SSHKeyStateList
		m.selectedKey = nil
	}

	return m, nil
}

// openKeyInEditor opens the public key in a readonly editor for copying
func (m SSHKeyManagementModel) openKeyInEditor() (tea.Model, tea.Cmd) {
	if m.selectedKey == nil {
		return m, nil
	}

	// Try to copy to system clipboard first, then show the key in less for viewing
	// This way user gets the key in clipboard automatically if possible
	copyScript := fmt.Sprintf(`
#!/bin/bash
KEY_CONTENT=$(cat "%s")
COPIED=0
STATUS_MSG=""

# Try different clipboard tools (suppress errors for headless environments)
if command -v xclip &> /dev/null; then
    if echo -n "$KEY_CONTENT" | xclip -selection clipboard 2>/dev/null; then
        STATUS_MSG="✓ Key copied to clipboard using xclip"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ] && command -v xsel &> /dev/null; then
    if echo -n "$KEY_CONTENT" | xsel --clipboard --input 2>/dev/null; then
        STATUS_MSG="✓ Key copied to clipboard using xsel"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ] && command -v pbcopy &> /dev/null; then
    if echo -n "$KEY_CONTENT" | pbcopy 2>/dev/null; then
        STATUS_MSG="✓ Key copied to clipboard using pbcopy"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ] && command -v wl-copy &> /dev/null; then
    if echo -n "$KEY_CONTENT" | wl-copy 2>/dev/null; then
        STATUS_MSG="✓ Key copied to clipboard using wl-copy"
        COPIED=1
    fi
fi

if [ $COPIED -eq 0 ]; then
    STATUS_MSG="ℹ  No display for clipboard. Hold SHIFT + select with mouse to copy."
fi

# Create temp file with content to show in less
TMPFILE=$(mktemp)
echo "$STATUS_MSG" > "$TMPFILE"
echo "" >> "$TMPFILE"
echo "Public Key (press 'q' to go back):" >> "$TMPFILE"
echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"
echo "$KEY_CONTENT" >> "$TMPFILE"
echo "─────────────────────────────────────────────────────────────────────" >> "$TMPFILE"

# Show in less - user presses 'q' to quit
less -R "$TMPFILE"

# Cleanup
rm -f "$TMPFILE"
`, m.selectedKey.PublicKeyPath)

	c := exec.Command("bash", "-c", copyScript)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		// Return nil to just refresh the screen without triggering any action
		return nil
	})
}

// generateKey generates a new SSH key
func (m SSHKeyManagementModel) generateKey() (tea.Model, tea.Cmd) {
	// Read values from form using GetString/GetBool
	keyTypeStr := m.form.GetString("keyType")
	keyIdentifier := m.form.GetString("keyIdentifier")
	keyEmail := m.form.GetString("keyEmail")
	keyPassphrase := m.form.GetString("keyPassphrase")
	addToAgent := m.form.GetBool("addToAgent")
	useForLogin := m.form.GetBool("useForLogin")

	keyType := system.SSHKeyTypeED25519
	switch keyTypeStr {
	case "rsa":
		keyType = system.SSHKeyTypeRSA
	case "ecdsa":
		keyType = system.SSHKeyTypeECDSA
	}

	// Use email as the key comment, with identifier for file name
	keyComment := keyEmail
	keyPath, err := m.userManager.GenerateSSHKey(m.username, keyType, keyIdentifier, keyPassphrase, 0, keyComment)
	if err != nil {
		m.err = fmt.Errorf("failed to generate key: %v", err)
		m.state = SSHKeyStateList
		m.form = nil
		return m, nil
	}

	// Add to authorized_keys if requested
	if useForLogin {
		err = m.userManager.AddKeyToAuthorizedKeys(m.username, keyPath+".pub")
		if err != nil {
			m.message = fmt.Sprintf("%s Key generated but failed to add to authorized_keys: %v", m.theme.Symbols.Warning, err)
			m.loadKeys()
			m.state = SSHKeyStateList
			m.form = nil
			return m, nil
		}
	}

	// Add to SSH agent if requested
	agentWarning := ""
	if addToAgent {
		err = m.userManager.AddKeyToSSHAgent(keyPath)
		if err != nil {
			// SSH agent failure is non-fatal - just show a warning
			agentWarning = fmt.Sprintf("\n\n%s Note: Could not add to ssh-agent (agent may not be running).\nYou can manually add it later with: ssh-add %s", m.theme.Symbols.Warning, keyPath)
		}
	}

	loginInfo := ""
	if useForLogin {
		loginInfo = "\n\nKey has been added to authorized_keys for SSH login."
	}

	m.message = fmt.Sprintf("%s SSH key '%s' generated successfully!\n\nKey path: %s%s%s", m.theme.Symbols.CheckMark, keyIdentifier, keyPath, loginInfo, agentWarning)
	m.loadKeys()
	m.state = SSHKeyStateList
	m.form = nil

	return m, nil
}

// View renders the SSH key management screen
func (m SSHKeyManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show error if there's one
	if m.err != nil {
		return m.renderError()
	}

	// Show message if there's one
	if m.message != "" {
		return m.renderMessage()
	}

	// Render based on state
	switch m.state {
	case SSHKeyStateList:
		return m.renderList()
	case SSHKeyStateGenerateForm:
		return m.renderGenerateForm()
	case SSHKeyStateKeyDetails:
		return m.renderKeyDetails()
	case SSHKeyStateConfirmDelete:
		return m.renderConfirmDelete()
	case SSHKeyStateCopyKey:
		return m.renderCopyKey()
	case SSHKeyStateExportOptions:
		return m.renderExportOptions()
	}

	return m.renderList()
}

// renderError renders the error view
func (m SSHKeyManagementModel) renderError() string {
	errorMsg := m.theme.Title.Render("SSH Key Management") + "\n\n" +
		m.theme.ErrorStyle.Render("Error:") + "\n" +
		m.theme.DescriptionStyle.Render(m.err.Error()) + "\n\n" +
		m.theme.Help.Render("Press any key to continue")

	bordered := m.theme.RenderBox(errorMsg)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderMessage renders the message view
func (m SSHKeyManagementModel) renderMessage() string {
	msgStyle := m.theme.InfoStyle
	if strings.Contains(m.message, m.theme.Symbols.CheckMark) {
		msgStyle = m.theme.SuccessStyle
	} else if strings.Contains(m.message, m.theme.Symbols.Warning) {
		msgStyle = m.theme.WarningStyle
	}

	messageDisplay := m.theme.Title.Render("SSH Key Management") + "\n\n" +
		msgStyle.Render(m.message) + "\n\n" +
		m.theme.Help.Render("Press any key to continue")

	bordered := m.theme.RenderBox(messageDisplay)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderList renders the key list view
func (m SSHKeyManagementModel) renderList() string {
	header := m.theme.Title.Render(fmt.Sprintf("SSH Keys for %s", m.username))

	var items []string

	// Generate new key option
	cursor := "  "
	if m.cursor == 0 {
		cursor = m.theme.KeyStyle.Render("▶ ")
	}
	generateItem := fmt.Sprintf("%s+ Generate New SSH Key", cursor)
	if m.cursor == 0 {
		generateItem = m.theme.SelectedItem.Render(generateItem)
	} else {
		generateItem = m.theme.MenuItem.Render(generateItem)
	}
	items = append(items, generateItem)
	items = append(items, "")

	// Key list header
	if len(m.keys) > 0 {
		items = append(items, m.theme.Label.Render("Existing Keys:"))
		items = append(items, "")

		for i, key := range m.keys {
			cursor := "  "
			if m.cursor == i+1 {
				cursor = m.theme.KeyStyle.Render("▶ ")
			}

			// Build key display
			identifier := key.Identifier
			if identifier == "" {
				identifier = filepath.Base(key.PublicKeyPath)
			}

			// Status indicators
			loginStatus := m.theme.ErrorStyle.Render("✗")
			if key.IsLoginKey {
				loginStatus = m.theme.SuccessStyle.Render("✓")
			}

			passphraseStatus := "No"
			if key.HasPassphrase {
				passphraseStatus = "Yes"
			}

			keyLine := fmt.Sprintf("%s[%s] %s (%s) | Login: %s | Passphrase: %s",
				cursor,
				strings.ToUpper(key.Type),
				identifier,
				key.Fingerprint[:min(20, len(key.Fingerprint))]+"...",
				loginStatus,
				passphraseStatus,
			)

			if m.cursor == i+1 {
				keyLine = m.theme.SelectedItem.Render(keyLine)
			} else {
				keyLine = m.theme.MenuItem.Render(keyLine)
			}

			items = append(items, keyLine)
		}
	} else {
		items = append(items, m.theme.DescriptionStyle.Render("No SSH keys found for this user."))
	}

	// Legend
	items = append(items, "")
	items = append(items, m.theme.DescriptionStyle.Render("Legend: Login "+m.theme.Symbols.CheckMark+"=authorized_keys | Pass=passphrase required"))

	list := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		list,
		"",
		help,
	)

	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderGenerateForm renders the key generation form
func (m SSHKeyManagementModel) renderGenerateForm() string {
	header := m.theme.Title.Render("Generate New SSH Key")

	formView := ""
	if m.form != nil {
		formView = m.form.View()
	}

	help := m.theme.Help.Render("Tab: Next Field • Enter: Submit • Esc: Cancel")

	// Apply padding similar to file browser
	paddingH := 10
	paddingV := 2

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		formView,
		"",
		help,
	)

	// Create padded container
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

// renderKeyDetails renders the key details view
func (m SSHKeyManagementModel) renderKeyDetails() string {
	if m.selectedKey == nil {
		return m.renderList()
	}

	header := m.theme.Title.Render("SSH Key Details")

	// Key info
	identifier := m.selectedKey.Identifier
	if identifier == "" {
		identifier = filepath.Base(m.selectedKey.PublicKeyPath)
	}

	infoLines := []string{
		m.theme.Label.Render("Identifier:  ") + m.theme.MenuItem.Render(identifier),
		m.theme.Label.Render("Type:        ") + m.theme.MenuItem.Render(strings.ToUpper(m.selectedKey.Type)),
		m.theme.Label.Render("Fingerprint: ") + m.theme.MenuItem.Render(m.selectedKey.Fingerprint),
		m.theme.Label.Render("Public Key:  ") + m.theme.MenuItem.Render(m.selectedKey.PublicKeyPath),
		m.theme.Label.Render("Private Key: ") + m.theme.MenuItem.Render(m.selectedKey.PrivateKeyPath),
	}

	// Login status
	loginStatus := m.theme.ErrorStyle.Render("✗ Not in authorized_keys")
	if m.selectedKey.IsLoginKey {
		loginStatus = m.theme.SuccessStyle.Render("✓ In authorized_keys (login enabled)")
	}
	infoLines = append(infoLines, m.theme.Label.Render("Login:       ")+loginStatus)

	// Passphrase status
	passphraseStatus := m.theme.InfoStyle.Render("No passphrase")
	if m.selectedKey.HasPassphrase {
		passphraseStatus = m.theme.SuccessStyle.Render("✓ Passphrase protected")
	}
	infoLines = append(infoLines, m.theme.Label.Render("Passphrase:  ")+passphraseStatus)

	// SSH Agent status
	agentStatus := m.theme.InfoStyle.Render("✗ Not loaded")
	if m.selectedKey.IsInAgent {
		agentStatus = m.theme.SuccessStyle.Render("✓ Loaded in ssh-agent")
	}
	infoLines = append(infoLines, m.theme.Label.Render("SSH Agent:   ")+agentStatus)

	info := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Actions
	actions := m.getKeyActions()
	var actionItems []string
	actionItems = append(actionItems, "")
	actionItems = append(actionItems, m.theme.Label.Render("Actions:"))

	for i, action := range actions {
		cursor := "  "
		if i == m.actionCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		actionLine := fmt.Sprintf("%s%s", cursor, action)
		if i == m.actionCursor {
			actionLine = m.theme.SelectedItem.Render(actionLine)
		} else {
			actionLine = m.theme.MenuItem.Render(actionLine)
		}

		actionItems = append(actionItems, actionLine)
	}

	actionList := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • c: Copy Public • e: Export Private • Esc: Back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		info,
		actionList,
		"",
		help,
	)

	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderCopyKey renders the copy key view with selectable text
func (m SSHKeyManagementModel) renderCopyKey() string {
	header := m.theme.Title.Render("Copy Public Key")

	instruction := m.theme.DescriptionStyle.Render("Select the key below with your mouse and copy (Cmd+C / Ctrl+C):")

	// Display the key in a plain text box that's selectable
	// Break the key into multiple lines for better display
	keyContent := m.copyableKey
	
	// Wrap the key for display (SSH keys are typically one long line)
	maxLineLen := 70
	var wrappedLines []string
	for len(keyContent) > 0 {
		if len(keyContent) <= maxLineLen {
			wrappedLines = append(wrappedLines, keyContent)
			break
		}
		wrappedLines = append(wrappedLines, keyContent[:maxLineLen])
		keyContent = keyContent[maxLineLen:]
	}
	wrappedKey := strings.Join(wrappedLines, "\n")

	// Use a simple style without background that allows text selection
	keyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Render(wrappedKey)

	pathInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Path: %s", m.selectedKey.PublicKeyPath))

	help := m.theme.Help.Render("Press Esc or Enter to go back")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		instruction,
		"",
		keyBox,
		"",
		pathInfo,
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

// renderExportOptions renders the export format selection view
func (m SSHKeyManagementModel) renderExportOptions() string {
	if m.selectedKey == nil {
		return m.renderKeyDetails()
	}

	header := m.theme.Title.Render("Export Private Key")

	identifier := m.selectedKey.Identifier
	if identifier == "" {
		identifier = filepath.Base(m.selectedKey.PublicKeyPath)
	}

	keyInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Key: %s (%s)", identifier, strings.ToUpper(m.selectedKey.Type)))

	question := m.theme.Label.Render("Select export format:")

	exportOptions := []string{"Linux/macOS (PEM format)", "Windows PuTTY (PPK format)"}
	var optionItems []string

	for i, option := range exportOptions {
		cursor := "  "
		if i == m.exportCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		optionLine := fmt.Sprintf("%s%s", cursor, option)
		if i == m.exportCursor {
			optionLine = m.theme.SelectedItem.Render(optionLine)
		} else {
			optionLine = m.theme.MenuItem.Render(optionLine)
		}

		optionItems = append(optionItems, optionLine)
	}

	optionList := lipgloss.JoinVertical(lipgloss.Left, optionItems...)

	warning := m.theme.WarningStyle.Render("⚠  Private keys are sensitive! Only export when necessary.")

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

	// Apply padding
	paddingH := 4
	paddingV := 1

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		keyInfo,
		"",
		question,
		"",
		optionList,
		"",
		warning,
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

// renderConfirmDelete renders the delete confirmation view
func (m SSHKeyManagementModel) renderConfirmDelete() string {
	if m.selectedKey == nil {
		return m.renderList()
	}

	identifier := m.selectedKey.Identifier
	if identifier == "" {
		identifier = filepath.Base(m.selectedKey.PublicKeyPath)
	}

	header := m.theme.Title.Render("Confirm Delete")

	warning := m.theme.WarningStyle.Render(fmt.Sprintf(
		"%s Are you sure you want to delete the SSH key '%s'?\n\n"+
			"This will:\n"+
			"  • Delete the private key: %s\n"+
			"  • Delete the public key: %s\n"+
			"  • Remove from authorized_keys if present\n\n"+
			"This action cannot be undone!",
		m.theme.Symbols.Warning,
		identifier,
		m.selectedKey.PrivateKeyPath,
		m.selectedKey.PublicKeyPath,
	))

	help := m.theme.Help.Render("y: Yes, delete • n/Esc: Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		warning,
		"",
		help,
	)

	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

