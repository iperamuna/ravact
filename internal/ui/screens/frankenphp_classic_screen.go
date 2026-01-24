package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FrankenPHPClassicModel represents the FrankenPHP Classic Mode screen
type FrankenPHPClassicModel struct {
	theme           *theme.Theme
	width           int
	height          int
	cursor          int
	mode            string // "check", "install_options", "site_setup", "confirm", "custom_url_input"
	binaryPath      string
	binaryVersion   string
	binaryFound     bool
	installOptions  []FrankenPHPInstallOption
	siteSetupFields []FrankenPHPField
	currentField    int
	err             error
	success         string
	inputMode       bool
	customURL       string // For custom URL input
}

// FrankenPHPInstallOption represents an installation option
type FrankenPHPInstallOption struct {
	ID          string
	Name        string
	Description string
}

// FrankenPHPField represents an input field for site setup
type FrankenPHPField struct {
	ID          string
	Label       string
	Value       string
	Placeholder string
	Description string
}

// NewFrankenPHPClassicModel creates a new FrankenPHP Classic Mode model
func NewFrankenPHPClassicModel() FrankenPHPClassicModel {
	binaryPath, version, found := detectFrankenPHPBinary()

	installOptions := []FrankenPHPInstallOption{
		{
			ID:          "download_official",
			Name:        "Download from Official Source",
			Description: "Download latest FrankenPHP binary from github.com/dunglas/frankenphp",
		},
		{
			ID:          "custom_url",
			Name:        "Download from Custom URL",
			Description: "Provide a URL to download the FrankenPHP binary",
		},
		{
			ID:          "manual",
			Name:        "Manual Installation",
			Description: "View instructions to manually install FrankenPHP",
		},
		{
			ID:          "back",
			Name:        "← Back to Site Commands",
			Description: "",
		},
	}

	siteSetupFields := []FrankenPHPField{
		{ID: "binary_path", Label: "FrankenPHP Binary", Value: binaryPath, Placeholder: "/usr/local/bin/frankenphp", Description: "Path to FrankenPHP binary (auto-detected)"},
		{ID: "site_root", Label: "Site Root", Value: "", Placeholder: "/var/www/mysite/current", Description: "Full path to your application root"},
		{ID: "site_key", Label: "Site Key", Value: "", Placeholder: "mysite", Description: "Unique identifier for service/socket names"},
		{ID: "docroot", Label: "Document Root", Value: "", Placeholder: "/var/www/mysite/current/public", Description: "Web-accessible directory (usually /public for Laravel)"},
		{ID: "domains", Label: "Domain Names", Value: "", Placeholder: "mysite.com www.mysite.com", Description: "Space-separated domain names for Nginx"},
		{ID: "user", Label: "Run as User", Value: "www-data", Placeholder: "www-data", Description: "System user to run the service"},
		{ID: "group", Label: "Run as Group", Value: "www-data", Placeholder: "www-data", Description: "System group to run the service"},
	}

	mode := "install_options"
	if found {
		mode = "site_setup"
	}

	return FrankenPHPClassicModel{
		theme:           theme.DefaultTheme(),
		cursor:          0,
		mode:            mode,
		binaryPath:      binaryPath,
		binaryVersion:   version,
		binaryFound:     found,
		installOptions:  installOptions,
		siteSetupFields: siteSetupFields,
		currentField:    0,
	}
}

// detectFrankenPHPBinary checks if FrankenPHP is installed
func detectFrankenPHPBinary() (path string, version string, found bool) {
	// Check common paths
	paths := []string{
		"/usr/local/bin/frankenphp",
		"/usr/bin/frankenphp",
	}

	// Also check PATH
	if p, err := exec.LookPath("frankenphp"); err == nil {
		paths = append([]string{p}, paths...)
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			// Get version
			cmd := exec.Command(p, "version")
			output, err := cmd.Output()
			if err == nil {
				version = strings.TrimSpace(string(output))
				// Extract just first line
				if lines := strings.Split(version, "\n"); len(lines) > 0 {
					version = lines[0]
				}
			} else {
				version = "installed"
			}
			return p, version, true
		}
	}

	return "", "", false
}

// Init initializes the FrankenPHP Classic screen
func (m FrankenPHPClassicModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for FrankenPHP Classic Mode
func (m FrankenPHPClassicModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle custom URL input mode
		if m.mode == "custom_url_input" {
			switch msg.String() {
			case "enter":
				if m.customURL != "" {
					return m.executeCustomURLDownload()
				}
				return m, nil
			case "esc":
				m.mode = "install_options"
				m.customURL = ""
				return m, nil
			case "backspace":
				if len(m.customURL) > 0 {
					m.customURL = m.customURL[:len(m.customURL)-1]
				}
			case "ctrl+v":
				// Ctrl+V is handled by terminal paste, which sends the text directly
				return m, nil
			default:
				input := msg.String()
				// Accept any printable input (single chars or pasted text)
				if len(input) > 0 && input != "ctrl+c" && input != "ctrl+z" {
					m.customURL += input
				}
			}
			return m, nil
		}

		// Handle input mode for site setup
		if m.inputMode && m.mode == "site_setup" {
			switch msg.String() {
			case "enter":
				m.inputMode = false
				// Auto-fill dependent fields
				m.autoFillFields()
				return m, nil
			case "esc":
				m.inputMode = false
				return m, nil
			case "backspace":
				field := &m.siteSetupFields[m.currentField]
				if len(field.Value) > 0 {
					field.Value = field.Value[:len(field.Value)-1]
				}
			case "tab":
				m.inputMode = false
				if m.currentField < len(m.siteSetupFields)-1 {
					m.currentField++
					m.cursor = m.currentField
				}
				m.inputMode = true
			case "ctrl+v":
				// Ctrl+V is handled by terminal paste, which sends the text directly
				return m, nil
			default:
				input := msg.String()
				// Accept any printable input (single chars or pasted text)
				// Filter out control sequences
				if len(input) > 0 && !strings.HasPrefix(input, "ctrl+") && !strings.HasPrefix(input, "alt+") {
					m.siteSetupFields[m.currentField].Value += input
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			if m.mode == "confirm" {
				m.mode = "site_setup"
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SiteCommandsScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.mode == "site_setup" {
					m.currentField = m.cursor
				}
			}

		case "down", "j":
			maxIdx := m.getMaxIndex()
			if m.cursor < maxIdx {
				m.cursor++
				if m.mode == "site_setup" {
					m.currentField = m.cursor
				}
			}

		case "enter", " ":
			return m.executeAction()

		case "e":
			// Edit current field in site setup mode
			if m.mode == "site_setup" && m.cursor < len(m.siteSetupFields) {
				m.currentField = m.cursor
				m.inputMode = true
			}
		}
	}

	return m, nil
}

// autoFillFields auto-fills dependent fields based on site_root
func (m *FrankenPHPClassicModel) autoFillFields() {
	siteRoot := ""
	for _, f := range m.siteSetupFields {
		if f.ID == "site_root" {
			siteRoot = f.Value
			break
		}
	}

	if siteRoot == "" {
		return
	}

	for i := range m.siteSetupFields {
		field := &m.siteSetupFields[i]
		switch field.ID {
		case "site_key":
			if field.Value == "" {
				// Derive from site root
				field.Value = suggestSiteKey(siteRoot)
			}
		case "docroot":
			if field.Value == "" {
				field.Value = filepath.Join(siteRoot, "public")
			}
		case "domains":
			if field.Value == "" {
				key := ""
				for _, f := range m.siteSetupFields {
					if f.ID == "site_key" {
						key = f.Value
						break
					}
				}
				if key != "" {
					field.Value = key + ".test"
				}
			}
		}
	}
}

// suggestSiteKey derives a site key from the site root path
func suggestSiteKey(siteRoot string) string {
	base := filepath.Base(filepath.Dir(siteRoot))
	if base == "" || base == "/" || base == "www" || base == "var" {
		base = filepath.Base(siteRoot)
	}
	// Sanitize
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, " ", "-")
	// Keep only safe chars
	var safe strings.Builder
	for _, c := range base {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '.' || c == '_' || c == '-' {
			safe.WriteRune(c)
		}
	}
	result := safe.String()
	if result == "" {
		return "site"
	}
	return result
}

// getMaxIndex returns the max cursor index for current mode
func (m FrankenPHPClassicModel) getMaxIndex() int {
	switch m.mode {
	case "install_options":
		return len(m.installOptions) - 1
	case "site_setup":
		return len(m.siteSetupFields) + 1 // fields + "Create Site" + "Back"
	case "confirm":
		return 1 // Yes/No
	}
	return 0
}

// executeAction handles the selected action
func (m FrankenPHPClassicModel) executeAction() (FrankenPHPClassicModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	switch m.mode {
	case "install_options":
		option := m.installOptions[m.cursor]
		switch option.ID {
		case "download_official":
			// Download from official source
			// FrankenPHP release names: frankenphp-linux-x86_64 or frankenphp-linux-aarch64
			downloadCmd := `#!/bin/bash
set -e
echo "=== FrankenPHP Download ==="
echo ""

# Detect architecture
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

case "$ARCH" in
    x86_64) 
        FRANKEN_ARCH="x86_64"
        ;;
    aarch64|arm64) 
        FRANKEN_ARCH="aarch64"
        ;;
    *) 
        echo "Error: Unsupported architecture: $ARCH"
        echo "Supported: x86_64, aarch64/arm64"
        exit 1 
        ;;
esac

# Build download URL
URL="https://github.com/dunglas/frankenphp/releases/latest/download/frankenphp-linux-${FRANKEN_ARCH}"
echo "Download URL: $URL"
echo ""

# Download with progress
echo "Downloading FrankenPHP binary..."
curl --fail --location --progress-bar --output /tmp/frankenphp "$URL"

# Check if download was successful
if [ ! -f /tmp/frankenphp ] || [ ! -s /tmp/frankenphp ]; then
    echo "Error: Download failed or file is empty"
    exit 1
fi

# Check if it's an actual binary (not HTML error page)
FILE_TYPE=$(file /tmp/frankenphp 2>/dev/null || echo "unknown")
if echo "$FILE_TYPE" | grep -q "HTML\|text"; then
    echo "Error: Downloaded file appears to be HTML, not a binary"
    echo "The download URL may have changed. Try manual download."
    rm -f /tmp/frankenphp
    exit 1
fi

echo ""
echo "Making binary executable..."
chmod +x /tmp/frankenphp

echo "Moving to /usr/local/bin/frankenphp..."
mv /tmp/frankenphp /usr/local/bin/frankenphp

echo ""
echo "========================================="
echo "✓ FrankenPHP installed successfully!"
echo "========================================="
echo ""
echo "Location: /usr/local/bin/frankenphp"
echo ""
frankenphp version || echo "Note: Run 'frankenphp version' to verify"
`
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     downloadCmd,
					Description: "Downloading FrankenPHP from official source",
				}
			}

		case "custom_url":
			// Switch to custom URL input mode
			m.mode = "custom_url_input"
			m.cursor = 0
			return m, nil

		case "manual":
			// Show manual instructions
			m.err = fmt.Errorf(`Manual Installation Instructions:

1. Download FrankenPHP binary from:
   https://github.com/dunglas/frankenphp/releases

2. Make it executable:
   chmod +x frankenphp

3. Move to system path:
   sudo mv frankenphp /usr/local/bin/frankenphp

4. Verify installation:
   frankenphp version

5. Return here to set up sites!`)
			return m, nil

		case "back":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SiteCommandsScreen}
			}
		}

	case "site_setup":
		// Check if cursor is on a field or action
		if m.cursor < len(m.siteSetupFields) {
			// Edit field
			m.currentField = m.cursor
			m.inputMode = true
			return m, nil
		}

		actionIdx := m.cursor - len(m.siteSetupFields)
		if actionIdx == 0 {
			// Create Site
			// Validate required fields
			var missing []string
			for _, f := range m.siteSetupFields {
				if f.ID == "site_root" && f.Value == "" {
					missing = append(missing, "Site Root")
				}
				if f.ID == "site_key" && f.Value == "" {
					missing = append(missing, "Site Key")
				}
			}
			if len(missing) > 0 {
				m.err = fmt.Errorf("required fields missing: %s", strings.Join(missing, ", "))
				return m, nil
			}
			m.mode = "confirm"
			m.cursor = 0
			return m, nil
		} else {
			// Back
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SiteCommandsScreen}
			}
		}

	case "confirm":
		if m.cursor == 0 {
			// Yes - create the site
			cmd := m.buildCreateSiteCommand()
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     cmd,
					Description: "Creating FrankenPHP site configuration",
				}
			}
		} else {
			// No - go back
			m.mode = "site_setup"
			m.cursor = 0
			return m, nil
		}
	}

	return m, nil
}

// executeCustomURLDownload downloads FrankenPHP from a custom URL
func (m FrankenPHPClassicModel) executeCustomURLDownload() (FrankenPHPClassicModel, tea.Cmd) {
	url := strings.TrimSpace(m.customURL)
	
	downloadCmd := fmt.Sprintf(`#!/bin/bash
set -e
echo "=== FrankenPHP Download from Custom URL ==="
echo ""
echo "URL: %s"
echo ""

# Download with progress
echo "Downloading FrankenPHP binary..."
curl --fail --location --progress-bar --output /tmp/frankenphp "%s"

# Check if download was successful
if [ ! -f /tmp/frankenphp ] || [ ! -s /tmp/frankenphp ]; then
    echo "Error: Download failed or file is empty"
    exit 1
fi

# Check file size (should be > 1MB for a real binary)
FILE_SIZE=$(stat -f%%z /tmp/frankenphp 2>/dev/null || stat -c%%s /tmp/frankenphp 2>/dev/null || echo "0")
if [ "$FILE_SIZE" -lt 1000000 ]; then
    echo "Warning: Downloaded file is smaller than expected ($FILE_SIZE bytes)"
    echo "This might not be the correct binary."
fi

# Check if it's an actual binary (not HTML error page)
FILE_TYPE=$(file /tmp/frankenphp 2>/dev/null || echo "unknown")
echo "File type: $FILE_TYPE"
if echo "$FILE_TYPE" | grep -q "HTML\|text"; then
    echo "Error: Downloaded file appears to be HTML/text, not a binary"
    rm -f /tmp/frankenphp
    exit 1
fi

echo ""
echo "Making binary executable..."
chmod +x /tmp/frankenphp

echo "Moving to /usr/local/bin/frankenphp..."
mv /tmp/frankenphp /usr/local/bin/frankenphp

echo ""
echo "========================================="
echo "✓ FrankenPHP installed successfully!"
echo "========================================="
echo ""
echo "Location: /usr/local/bin/frankenphp"
echo ""
frankenphp version || echo "Note: Run 'frankenphp version' to verify"
`, url, url)

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     downloadCmd,
			Description: "Downloading FrankenPHP from custom URL",
		}
	}
}

// buildCreateSiteCommand generates the bash script to create the site
func (m FrankenPHPClassicModel) buildCreateSiteCommand() string {
	var binaryPath, siteRoot, siteKey, docroot, domains, user, group string

	for _, f := range m.siteSetupFields {
		switch f.ID {
		case "binary_path":
			binaryPath = f.Value
		case "site_root":
			siteRoot = f.Value
		case "site_key":
			siteKey = f.Value
		case "docroot":
			docroot = f.Value
		case "domains":
			domains = f.Value
		case "user":
			user = f.Value
		case "group":
			group = f.Value
		}
	}

	// Use detected binary path if not specified
	if binaryPath == "" {
		binaryPath = m.binaryPath
	}
	if binaryPath == "" {
		binaryPath = "/usr/local/bin/frankenphp"
	}

	// Use defaults if empty
	if docroot == "" {
		docroot = siteRoot + "/public"
	}
	if domains == "" {
		domains = siteKey + ".test"
	}
	if user == "" {
		user = "www-data"
	}
	if group == "" {
		group = "www-data"
	}

	return fmt.Sprintf(`#!/bin/bash
set -e

SITE_ROOT="%s"
SITE_KEY="%s"
DOCROOT="%s"
DOMAINS="%s"
RUN_USER="%s"
RUN_GROUP="%s"
FRANKENPHP_BIN="%s"

# Verify FrankenPHP binary exists
if [ ! -x "$FRANKENPHP_BIN" ]; then
    echo "Error: FrankenPHP binary not found or not executable at: $FRANKENPHP_BIN"
    exit 1
fi

SERVICE_NAME="frankenphp-${SITE_KEY}"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
SOCK="/run/frankenphp/${SITE_KEY}.sock"

echo "Creating FrankenPHP Classic Mode site: ${SITE_KEY}"
echo ""

# Create systemd service file
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=FrankenPHP classic mode (${SITE_KEY})
After=network.target
Wants=network.target

[Service]
Type=simple
User=${RUN_USER}
Group=${RUN_GROUP}
WorkingDirectory=${SITE_ROOT}

Environment=APP_ENV=production
Environment=APP_BASE_PATH=${SITE_ROOT}

RuntimeDirectory=frankenphp
RuntimeDirectoryMode=0755

ExecStart=${FRANKENPHP_BIN} php-server \\
    --listen unix:${SOCK} \\
    --root ${DOCROOT}

Restart=always
RestartSec=2
TimeoutStopSec=10

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

echo "✓ Created systemd service: ${SERVICE_FILE}"

# Reload systemd and start service
systemctl daemon-reload
systemctl enable --now "${SERVICE_NAME}"
echo "✓ Service enabled and started"

# Check status
systemctl status "${SERVICE_NAME}" --no-pager || true

# Generate Nginx config
NGINX_CONF="/etc/nginx/sites-available/${SITE_KEY}.conf"
cat > "$NGINX_CONF" <<EOF
server {
    listen 80;
    server_name ${DOMAINS};

    location / {
        proxy_pass http://unix:${SOCK};
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }
}
EOF

echo "✓ Created Nginx config: ${NGINX_CONF}"

# Enable site
if [ -d /etc/nginx/sites-enabled ]; then
    ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/${SITE_KEY}.conf"
    echo "✓ Enabled Nginx site"
    
    # Test and reload nginx
    nginx -t && systemctl reload nginx
    echo "✓ Nginx reloaded"
fi

echo ""
echo "========================================="
echo "FrankenPHP Classic Mode site created!"
echo "========================================="
echo ""
echo "Service: ${SERVICE_NAME}"
echo "Socket: ${SOCK}"
echo "Nginx: ${NGINX_CONF}"
echo ""
echo "Commands:"
echo "  systemctl status ${SERVICE_NAME}"
echo "  journalctl -u ${SERVICE_NAME} -f"
echo ""
`, siteRoot, siteKey, docroot, domains, user, group, binaryPath)
}

// View renders the FrankenPHP Classic Mode screen
func (m FrankenPHPClassicModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.mode {
	case "install_options":
		return m.viewInstallOptions()
	case "custom_url_input":
		return m.viewCustomURLInput()
	case "site_setup":
		return m.viewSiteSetup()
	case "confirm":
		return m.viewConfirm()
	}

	return "Unknown mode"
}

// viewCustomURLInput renders the custom URL input view
func (m FrankenPHPClassicModel) viewCustomURLInput() string {
	header := m.theme.Title.Render("FrankenPHP - Download from Custom URL")

	instructions := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.DescriptionStyle.Render("Enter the direct download URL for the FrankenPHP binary."),
		"",
		m.theme.DescriptionStyle.Render("Examples:"),
		m.theme.InfoStyle.Render("  • https://github.com/dunglas/frankenphp/releases/download/v1.0.0/frankenphp-linux-x86_64"),
		m.theme.InfoStyle.Render("  • https://your-server.com/frankenphp"),
		"",
		m.theme.WarningStyle.Render("Note: URL must point directly to the binary file."),
	)

	// Input field
	inputLabel := m.theme.Label.Render("URL: ")
	inputValue := m.theme.SelectedItem.Render(m.customURL + "_")
	inputField := inputLabel + inputValue

	help := m.theme.Help.Render("Enter: Download • Esc: Cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", instructions, "", inputField, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewInstallOptions renders the installation options view
func (m FrankenPHPClassicModel) viewInstallOptions() string {
	header := m.theme.Title.Render("FrankenPHP Classic Mode")

	// Warning that binary not found
	warning := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.WarningStyle.Render("⚠ FrankenPHP binary not found"),
		"",
		m.theme.DescriptionStyle.Render("FrankenPHP needs to be installed before setting up sites."),
		m.theme.DescriptionStyle.Render("The binary should be at /usr/local/bin/frankenphp or /usr/bin/frankenphp"),
	)

	// Options
	var items []string
	items = append(items, "")
	items = append(items, m.theme.Subtitle.Render("Installation Options:"))
	items = append(items, "")

	for i, opt := range m.installOptions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, opt.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, opt.Name))
		}
		items = append(items, renderedItem)

		if i == m.cursor && opt.Description != "" {
			items = append(items, "    "+m.theme.DescriptionStyle.Render(opt.Description))
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Error message (used for showing instructions)
	messageSection := ""
	if m.err != nil {
		messageSection = m.theme.InfoStyle.Render(m.err.Error())
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit")

	sections := []string{header, "", warning, menu}
	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewSiteSetup renders the site setup view
func (m FrankenPHPClassicModel) viewSiteSetup() string {
	header := m.theme.Title.Render("FrankenPHP Classic Mode - Site Setup")

	// Binary info
	binaryInfo := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.SuccessStyle.Render("✓ FrankenPHP detected"),
		m.theme.Label.Render("Binary: ")+m.theme.InfoStyle.Render(m.binaryPath),
		m.theme.Label.Render("Version: ")+m.theme.DescriptionStyle.Render(m.binaryVersion),
	)

	// Architecture description
	archInfo := m.theme.DescriptionStyle.Render("Creates: systemd service + Nginx vhost + Unix socket for each site")

	// Form fields
	var fieldItems []string
	fieldItems = append(fieldItems, "")
	fieldItems = append(fieldItems, m.theme.Subtitle.Render("Site Configuration:"))
	fieldItems = append(fieldItems, "")

	for i, field := range m.siteSetupFields {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		label := field.Label + ": "
		value := field.Value
		if value == "" {
			value = field.Placeholder
		}

		var renderedField string
		if i == m.cursor {
			if m.inputMode && i == m.currentField {
				// Editing mode
				renderedField = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, label)) +
					m.theme.SelectedItem.Render(field.Value+"_")
			} else {
				renderedField = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, label)) +
					m.theme.InfoStyle.Render(value)
			}
			fieldItems = append(fieldItems, renderedField)
			fieldItems = append(fieldItems, "    "+m.theme.DescriptionStyle.Render(field.Description))
			fieldItems = append(fieldItems, "    "+m.theme.KeyStyle.Render("Press Enter or 'e' to edit"))
		} else {
			valueStyle := m.theme.DescriptionStyle
			if field.Value != "" {
				valueStyle = m.theme.InfoStyle
			}
			renderedField = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, label)) +
				valueStyle.Render(value)
			fieldItems = append(fieldItems, renderedField)
		}
	}

	// Actions
	fieldItems = append(fieldItems, "")
	actions := []string{"Create Site", "← Back"}
	for i, action := range actions {
		actualIdx := len(m.siteSetupFields) + i
		cursor := "  "
		if actualIdx == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if actualIdx == m.cursor {
			if i == 0 {
				renderedItem = m.theme.SuccessStyle.Render(fmt.Sprintf("%s%s", cursor, action))
			} else {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action))
			}
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action))
		}
		fieldItems = append(fieldItems, renderedItem)
	}

	form := lipgloss.JoinVertical(lipgloss.Left, fieldItems...)

	// Error message
	messageSection := ""
	if m.err != nil {
		messageSection = m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.success != "" {
		messageSection = m.theme.SuccessStyle.Render(m.success)
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter/e: Edit • Tab: Next field • Esc: Back")

	sections := []string{header, "", binaryInfo, "", archInfo, form}
	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewConfirm renders the confirmation view
func (m FrankenPHPClassicModel) viewConfirm() string {
	header := m.theme.Title.Render("Confirm Site Creation")

	// Show summary
	var summary []string
	summary = append(summary, m.theme.Subtitle.Render("Site Configuration Summary:"))
	summary = append(summary, "")

	for _, f := range m.siteSetupFields {
		if f.Value != "" {
			summary = append(summary, m.theme.Label.Render(f.Label+": ")+m.theme.InfoStyle.Render(f.Value))
		}
	}

	// What will be created
	siteKey := ""
	for _, f := range m.siteSetupFields {
		if f.ID == "site_key" {
			siteKey = f.Value
			break
		}
	}

	summary = append(summary, "")
	summary = append(summary, m.theme.Subtitle.Render("Will create:"))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /etc/systemd/system/frankenphp-%s.service", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /etc/nginx/sites-available/%s.conf", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /run/frankenphp/%s.sock (runtime)", siteKey)))

	summarySection := lipgloss.JoinVertical(lipgloss.Left, summary...)

	// Yes/No options
	var options []string
	options = append(options, "")
	choices := []string{"Yes, create the site", "No, go back"}
	for i, choice := range choices {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			if i == 0 {
				renderedItem = m.theme.SuccessStyle.Render(fmt.Sprintf("%s%s", cursor, choice))
			} else {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, choice))
			}
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, choice))
		}
		options = append(options, renderedItem)
	}

	optionsSection := lipgloss.JoinVertical(lipgloss.Left, options...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Confirm • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", summarySection, optionsSection, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
