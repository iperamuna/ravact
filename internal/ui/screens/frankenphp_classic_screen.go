package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FrankenPHPClassicModel represents the FrankenPHP Classic Mode screen
type FrankenPHPClassicModel struct {
	theme         *theme.Theme
	width         int
	height        int
	cursor        int
	mode          string // "install_options", "site_setup", "confirm", "custom_url_input", "composer_setup"
	binaryPath    string
	binaryVersion string
	binaryFound   bool

	// Install options (used when binary not found)
	installOptions []FrankenPHPInstallOption
	customURL      string

	// Current directory (for auto-detection)
	currentDir string

	// Form fields for site setup (huh form)
	form         *huh.Form
	formSiteRoot string
	formSiteKey  string
	formDocroot  string
	formDomains  string
	formPort     string
	formUser     string
	formGroup    string

	// Composer setup options
	composerOptions []ComposerSetupOption
	composerCursor  int

	// UI state
	err     error
	message string
}

// ComposerSetupOption represents a composer setup option
type ComposerSetupOption struct {
	ID          string
	Name        string
	Description string
}

// FrankenPHPInstallOption represents an installation option
type FrankenPHPInstallOption struct {
	ID          string
	Name        string
	Description string
}

// NewFrankenPHPClassicModel creates a new FrankenPHP Classic Mode model
func NewFrankenPHPClassicModel() FrankenPHPClassicModel {
	return NewFrankenPHPClassicModelWithDir("")
}

// NewFrankenPHPClassicModelWithDir creates a new FrankenPHP Classic Mode model with a specific directory
func NewFrankenPHPClassicModelWithDir(currentDir string) FrankenPHPClassicModel {
	t := theme.DefaultTheme()
	binaryPath, version, found := detectFrankenPHPBinary()

	// Auto-detect current directory if not provided
	if currentDir == "" {
		currentDir, _ = os.Getwd()
	}

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

	composerOptions := []ComposerSetupOption{
		{
			ID:          "option_a",
			Name:        "Option A: Replace system PHP (Recommended)",
			Description: "Create symlink so 'php' command uses FrankenPHP. Required for Laravel/Symfony scripts.",
		},
		{
			ID:          "option_both",
			Name:        "Option A+C: PHP symlink + Composer wrapper (Best)",
			Description: "Both PHP symlink AND composer wrapper. Full compatibility with all projects.",
		},
		{
			ID:          "option_c",
			Name:        "Option C: Composer wrapper only",
			Description: "Only wrap composer. Note: Laravel @php scripts won't work without Option A.",
		},
		{
			ID:          "skip",
			Name:        "Skip Composer Setup",
			Description: "Don't configure Composer integration now. You can do it manually later.",
		},
	}

	mode := "install_options"
	if found {
		mode = "site_setup"
	}

	// Pre-fill site root with current directory
	siteRoot := currentDir
	siteKey := suggestSiteKey(siteRoot)

	m := FrankenPHPClassicModel{
		theme:           t,
		cursor:          0,
		mode:            mode,
		binaryPath:      binaryPath,
		binaryVersion:   version,
		binaryFound:     found,
		installOptions:  installOptions,
		composerOptions: composerOptions,
		currentDir:      currentDir,
		formSiteRoot:    siteRoot,
		formSiteKey:     siteKey,
		formUser:        "www-data",
		formGroup:       "www-data",
		formPort:        "8000",
	}

	// Build the huh form for site setup
	m.form = m.buildSiteSetupForm()

	return m
}

// buildSiteSetupForm creates the huh form for site configuration
func (m *FrankenPHPClassicModel) buildSiteSetupForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("siteRoot").
				Title("Site Root").
				Description("Full path to your application root").
				Placeholder("/var/www/mysite").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("site root is required")
					}
					if !strings.HasPrefix(s, "/") {
						return fmt.Errorf("must be an absolute path starting with /")
					}
					return nil
				}).
				Value(&m.formSiteRoot),

			huh.NewInput().
				Key("siteKey").
				Title("Site Key").
				Description("Unique identifier for service/socket names").
				Placeholder("mysite").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("site key is required")
					}
					if strings.Contains(s, " ") {
						return fmt.Errorf("site key cannot contain spaces")
					}
					return nil
				}).
				Value(&m.formSiteKey),

			huh.NewInput().
				Key("docroot").
				Title("Document Root").
				Description("Web-accessible directory (usually /public for Laravel)").
				Placeholder("/var/www/mysite/public").
				Value(&m.formDocroot),

			huh.NewInput().
				Key("domains").
				Title("Domain Names").
				Description("Space-separated domain names for Nginx proxy").
				Placeholder("mysite.com www.mysite.com").
				Value(&m.formDomains),

			huh.NewInput().
				Key("port").
				Title("Port").
				Description("Port for FrankenPHP (fallback if not using Unix socket)").
				Placeholder("8000").
				Validate(func(s string) error {
					if s == "" {
						return nil // Optional, will use default
					}
					port, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("port must be a number")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("port must be between 1 and 65535")
					}
					return nil
				}).
				Value(&m.formPort),

			huh.NewInput().
				Key("user").
				Title("Run as User").
				Description("System user to run the FrankenPHP service").
				Placeholder("www-data").
				Value(&m.formUser),

			huh.NewInput().
				Key("group").
				Title("Run as Group").
				Description("System group to run the FrankenPHP service").
				Placeholder("www-data").
				Value(&m.formGroup),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
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
	// Always initialize the form when in site_setup mode
	if m.mode == "site_setup" && m.form != nil {
		return m.form.Init()
	}
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
		// Handle message display state (success/error)
		if m.message != "" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace", "enter", " ":
				if strings.Contains(m.message, "✓") {
					return m, func() tea.Msg {
						return NavigateMsg{Screen: SiteCommandsScreen}
					}
				}
				m.message = ""
				return m, nil
			default:
				m.message = ""
				return m, nil
			}
		}

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
			default:
				input := msg.String()
				if len(input) > 0 && input != "ctrl+c" && input != "ctrl+z" {
					m.customURL += input
				}
			}
			return m, nil
		}

		// Handle confirm mode
		if m.mode == "confirm" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.mode = "site_setup"
				m.form = m.buildSiteSetupForm()
				return m, m.form.Init()
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < 1 {
					m.cursor++
				}
			case "enter", " ":
				if m.cursor == 0 {
					// Yes - create the site, then show composer setup
					m.mode = "composer_setup"
					m.composerCursor = 0
					return m, nil
				} else {
					// No - go back to form
					m.mode = "site_setup"
					m.form = m.buildSiteSetupForm()
					return m, m.form.Init()
				}
			}
			return m, nil
		}

		// Handle composer_setup mode
		if m.mode == "composer_setup" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.mode = "confirm"
				m.cursor = 0
				return m, nil
			case "up", "k":
				if m.composerCursor > 0 {
					m.composerCursor--
				}
			case "down", "j":
				if m.composerCursor < len(m.composerOptions)-1 {
					m.composerCursor++
				}
			case "enter", " ":
				return m.executeWithComposerSetup()
			}
			return m, nil
		}

		// Handle install options mode
		if m.mode == "install_options" {
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
				if m.cursor < len(m.installOptions)-1 {
					m.cursor++
				}
			case "enter", " ":
				return m.executeInstallOption()
			}
			return m, nil
		}

		// Handle site_setup mode with huh form
		if m.mode == "site_setup" {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				if m.form.State == huh.StateNormal {
					return m, func() tea.Msg {
						return NavigateMsg{Screen: SiteCommandsScreen}
					}
				}
			}
		}
	}

	// Update huh form in site_setup mode
	if m.mode == "site_setup" && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		// Check if form is completed
		if m.form.State == huh.StateCompleted {
			// Read form values explicitly (in case pointer binding didn't work)
			if v := m.form.GetString("siteRoot"); v != "" {
				m.formSiteRoot = v
			}
			if v := m.form.GetString("siteKey"); v != "" {
				m.formSiteKey = v
			}
			if v := m.form.GetString("docroot"); v != "" {
				m.formDocroot = v
			}
			if v := m.form.GetString("domains"); v != "" {
				m.formDomains = v
			}
			if v := m.form.GetString("port"); v != "" {
				m.formPort = v
			}
			if v := m.form.GetString("user"); v != "" {
				m.formUser = v
			}
			if v := m.form.GetString("group"); v != "" {
				m.formGroup = v
			}
			// Auto-fill empty fields
			m.autoFillFields()
			// Go to confirmation
			m.mode = "confirm"
			m.cursor = 0
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

// autoFillFields auto-fills dependent fields based on site_root
func (m *FrankenPHPClassicModel) autoFillFields() {
	if m.formSiteRoot == "" {
		return
	}

	// Auto-fill site key from site root
	if m.formSiteKey == "" {
		m.formSiteKey = suggestSiteKey(m.formSiteRoot)
	}

	// Auto-fill document root
	if m.formDocroot == "" {
		m.formDocroot = filepath.Join(m.formSiteRoot, "public")
	}

	// Auto-fill domains
	if m.formDomains == "" && m.formSiteKey != "" {
		m.formDomains = m.formSiteKey + ".test"
	}

	// Default port
	if m.formPort == "" {
		m.formPort = "8000"
	}

	// Default user/group
	if m.formUser == "" {
		m.formUser = "www-data"
	}
	if m.formGroup == "" {
		m.formGroup = "www-data"
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

// executeInstallOption handles the selected installation option
func (m FrankenPHPClassicModel) executeInstallOption() (tea.Model, tea.Cmd) {
	m.err = nil

	option := m.installOptions[m.cursor]
	switch option.ID {
	case "download_official":
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
		m.mode = "custom_url_input"
		m.cursor = 0
		return m, nil

	case "manual":
		m.message = `Manual Installation Instructions:

1. Download FrankenPHP binary from:
   https://github.com/dunglas/frankenphp/releases

2. Make it executable:
   chmod +x frankenphp

3. Move to system path:
   sudo mv frankenphp /usr/local/bin/frankenphp

4. Verify installation:
   frankenphp version

5. Return here to set up sites!

Press any key to continue...`
		return m, nil

	case "back":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: SiteCommandsScreen}
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
	// Get values from form fields
	siteRoot := m.formSiteRoot
	siteKey := m.formSiteKey
	docroot := m.formDocroot
	domains := m.formDomains
	port := m.formPort
	user := m.formUser
	group := m.formGroup

	// Use detected binary path
	binaryPath := m.binaryPath
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
	if port == "" {
		port = "8000"
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
PORT="%s"
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

# Create runtime directory
mkdir -p /run/frankenphp
chown ${RUN_USER}:${RUN_GROUP} /run/frankenphp

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

# Listen on Unix socket (preferred) with fallback port
ExecStart=${FRANKENPHP_BIN} php-server \\
    --listen unix:${SOCK} \\
    --listen 127.0.0.1:${PORT} \\
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

if [ -f "$NGINX_CONF" ]; then
    echo ""
    echo "========================================="
    echo "⚠ Nginx config already exists!"
    echo "========================================="
    echo ""
    echo "File: ${NGINX_CONF}"
    echo ""
    echo "Existing configuration:"
    echo "----------------------------------------"
    cat "$NGINX_CONF"
    echo "----------------------------------------"
    echo ""
    echo "Skipping Nginx config creation to preserve existing settings."
    echo "If you want to update it, please edit manually or delete the file first."
else
    cat > "$NGINX_CONF" <<EOF
upstream frankenphp_${SITE_KEY} {
    # Prefer Unix socket for better performance
    server unix:${SOCK} fail_timeout=0;
    # Fallback to TCP port
    server 127.0.0.1:${PORT} backup;
}

server {
    listen 80;
    server_name ${DOMAINS};

    location / {
        proxy_pass http://frankenphp_${SITE_KEY};
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
fi

echo ""
echo "========================================="
echo "Creating fpcli CLI wrapper..."
echo "========================================="

# Create fpcli CLI wrapper script (FrankenPHP CLI)
cat > /usr/local/bin/fpcli <<'FPCLI'
#!/usr/bin/env bash
set -euo pipefail

# Defaults that work under sudo/systemd too
DEFAULT_HOME="/var/www"
if [ "$(id -u)" -eq 0 ]; then
  DEFAULT_HOME="/root"
fi

export HOME="${HOME:-$DEFAULT_HOME}"
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"

FRANKENPHP="/usr/local/bin/frankenphp"
php_args=()

# Parse php CLI flags that may appear before the script
# We intentionally ignore -d flags like allow_url_fopen=1, memory_limit=..., etc.
while [ $# -gt 0 ]; do
  case "$1" in
    -d)
      # ignore -d and its value
      shift 2
      ;;
    -d*)
      # ignore combined form like -dallow_url_fopen=1
      shift
      ;;
    -c|-f)
      # ignore and its value
      shift 2
      ;;
    -n|-q)
      # ignore
      shift
      ;;
    -v|--version)
      exec "$FRANKENPHP" php-cli -r 'echo PHP_VERSION, PHP_EOL;'
      ;;
    -m)
      exec "$FRANKENPHP" php-cli -r 'foreach (get_loaded_extensions() as $e) echo $e, PHP_EOL;'
      ;;
    -i)
      exec "$FRANKENPHP" php-cli -r 'phpinfo();'
      ;;
    --ini)
      exec "$FRANKENPHP" php-cli -r 'echo "Loaded Configuration File: ", (php_ini_loaded_file() ?: "(none)"), PHP_EOL; echo "Scan this dir for additional .ini files: ", (php_ini_scanned_files() ? dirname(explode(",", php_ini_scanned_files())[0]) : "(none)"), PHP_EOL;'
      ;;
    -r)
      shift
      exec "$FRANKENPHP" php-cli "${php_args[@]}" -r "${1-}"
      ;;
    --)
      shift
      break
      ;;
    -*)
      # pass unknown flags through (best effort)
      php_args+=("$1")
      shift
      ;;
    *)
      # first non-flag is script (artisan, composer.phar, file.php)
      break
      ;;
  esac
done

exec "$FRANKENPHP" php-cli "${php_args[@]}" "$@"
FPCLI

chmod +x /usr/local/bin/fpcli
echo "✓ Created /usr/local/bin/fpcli (FrankenPHP CLI wrapper)"

echo ""
echo "========================================="
echo "FrankenPHP Classic Mode site created!"
echo "========================================="
echo ""
echo "Service: ${SERVICE_NAME}"
echo "Socket: ${SOCK}"
echo "Port: 127.0.0.1:${PORT} (fallback)"
echo "Nginx: ${NGINX_CONF}"
echo "CLI: /usr/local/bin/fpcli"
echo ""
echo "Commands:"
echo "  systemctl status ${SERVICE_NAME}"
echo "  journalctl -u ${SERVICE_NAME} -f"
echo "  fpcli -v  (PHP version via FrankenPHP)"
echo ""
`, siteRoot, siteKey, docroot, domains, port, user, group, binaryPath)
}

// executeWithComposerSetup runs the site creation with the selected composer option
func (m FrankenPHPClassicModel) executeWithComposerSetup() (tea.Model, tea.Cmd) {
	// Get the base site creation command
	siteCmd := m.buildCreateSiteCommand()

	// Get the selected composer option
	option := m.composerOptions[m.composerCursor]

	var composerCmd string
	switch option.ID {
	case "option_a":
		// Option A: Replace system PHP with fpcli symlink
		composerCmd = `
echo ""
echo "========================================="
echo "Setting up Composer with FrankenPHP..."
echo "========================================="
echo ""
echo "Option A: Creating PHP symlink to fpcli"
echo ""

# Backup existing php if it exists and is not a symlink
if [ -f /usr/local/bin/php ] && [ ! -L /usr/local/bin/php ]; then
    echo "Backing up existing /usr/local/bin/php to /usr/local/bin/php.bak"
    mv /usr/local/bin/php /usr/local/bin/php.bak
fi

# Create symlink
ln -sf /usr/local/bin/fpcli /usr/local/bin/php
hash -r 2>/dev/null || true

echo "✓ Created symlink: /usr/local/bin/php -> /usr/local/bin/fpcli"
echo ""
echo "Verification:"
which php
php -v
echo ""
echo "✓ Composer will now use FrankenPHP automatically!"
echo ""
echo "Note: System PHP (if installed) is still available at /usr/bin/php"
`
	case "option_both":
		// Option A+C: Both PHP symlink and Composer wrapper
		composerCmd = `
echo ""
echo "========================================="
echo "Setting up Composer with FrankenPHP..."
echo "========================================="
echo ""
echo "Option A+C: Creating PHP symlink AND Composer wrapper"
echo ""

# === Part 1: Create PHP symlink (Option A) ===
echo "[1/2] Creating PHP symlink..."

# Backup existing php if it exists and is not a symlink
if [ -f /usr/local/bin/php ] && [ ! -L /usr/local/bin/php ]; then
    echo "  Backing up existing /usr/local/bin/php to /usr/local/bin/php.bak"
    mv /usr/local/bin/php /usr/local/bin/php.bak
fi

# Create symlink
ln -sf /usr/local/bin/fpcli /usr/local/bin/php
echo "  ✓ Created symlink: /usr/local/bin/php -> /usr/local/bin/fpcli"

# === Part 2: Create Composer wrapper (Option C) ===
echo ""
echo "[2/2] Setting up Composer wrapper..."

# Check if composer.phar already exists at expected location
if [ -f /usr/local/bin/composer.phar ]; then
    echo "  ✓ composer.phar already exists at /usr/local/bin/composer.phar"
else
    echo "  Downloading Composer..."
    
    # Download composer installer
    cd /tmp
    curl -sS https://getcomposer.org/installer -o composer-setup.php
    
    if [ ! -f composer-setup.php ]; then
        echo "  Error: Failed to download composer installer"
        exit 1
    fi
    
    # Run installer with fpcli
    /usr/local/bin/fpcli composer-setup.php --install-dir=/usr/local/bin --filename=composer.phar
    
    # Clean up installer
    rm -f composer-setup.php
    
    # Verify download
    if [ -f /usr/local/bin/composer.phar ]; then
        echo "  ✓ Composer downloaded to /usr/local/bin/composer.phar"
    else
        echo "  Trying alternative download method..."
        curl -sS https://getcomposer.org/download/latest-stable/composer.phar -o /usr/local/bin/composer.phar
        
        if [ ! -f /usr/local/bin/composer.phar ]; then
            echo "  Error: All download methods failed"
            exit 1
        fi
        echo "  ✓ Composer downloaded via direct download"
    fi
fi

# Make sure composer.phar is executable
chmod +x /usr/local/bin/composer.phar

# Create wrapper script
cat > /usr/local/bin/composer <<'COMPOSERWRAP'
#!/usr/bin/env bash
set -euo pipefail

export HOME="${HOME:-/root}"
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"

# CRITICAL: make Composer scripts (@php) use our php shim
export PHP_BINARY="/usr/local/bin/php"

exec /usr/local/bin/fpcli /usr/local/bin/composer.phar "$@"
COMPOSERWRAP

chmod +x /usr/local/bin/composer
echo "  ✓ Created composer wrapper at /usr/local/bin/composer"

# Update PATH hash
hash -r 2>/dev/null || true

echo ""
echo "========================================="
echo "Verification:"
echo "========================================="
echo ""
echo "PHP:"
which php
php -v | head -1
echo ""
echo "Composer:"
which composer
composer --version
echo ""
echo "✓ Full FrankenPHP integration complete!"
echo ""
echo "Both 'php' and 'composer' commands now use FrankenPHP."
echo "Laravel @php scripts will work correctly."
`
	case "option_c":
		// Option C: Create composer wrapper
		composerCmd = `
echo ""
echo "========================================="
echo "Setting up Composer with FrankenPHP..."
echo "========================================="
echo ""
echo "Option C: Creating Composer wrapper"
echo ""

# Check if composer.phar already exists at expected location
if [ -f /usr/local/bin/composer.phar ]; then
    echo "✓ composer.phar already exists at /usr/local/bin/composer.phar"
else
    echo "Downloading Composer..."
    echo ""
    
    # Download composer installer
    cd /tmp
    curl -sS https://getcomposer.org/installer -o composer-setup.php
    
    if [ ! -f composer-setup.php ]; then
        echo "Error: Failed to download composer installer"
        exit 1
    fi
    
    # Run installer with fpcli
    echo "Running composer installer with FrankenPHP..."
    /usr/local/bin/fpcli composer-setup.php --install-dir=/usr/local/bin --filename=composer.phar
    
    # Clean up installer
    rm -f composer-setup.php
    
    # Verify download
    if [ -f /usr/local/bin/composer.phar ]; then
        echo "✓ Composer downloaded to /usr/local/bin/composer.phar"
    else
        echo "Error: Composer installation failed"
        echo "Trying alternative download method..."
        
        # Alternative: download phar directly
        curl -sS https://getcomposer.org/download/latest-stable/composer.phar -o /usr/local/bin/composer.phar
        
        if [ ! -f /usr/local/bin/composer.phar ]; then
            echo "Error: All download methods failed"
            exit 1
        fi
        echo "✓ Composer downloaded via direct download"
    fi
fi

# Make sure composer.phar is executable
chmod +x /usr/local/bin/composer.phar

# Verify composer.phar exists
echo ""
echo "Verifying composer.phar..."
ls -la /usr/local/bin/composer.phar

# Test that composer.phar works with fpcli
echo ""
echo "Testing composer.phar with fpcli..."
/usr/local/bin/fpcli /usr/local/bin/composer.phar --version
if [ $? -ne 0 ]; then
    echo "Warning: composer.phar test returned non-zero, but may still work"
fi

# Create wrapper script
echo ""
echo "Creating composer wrapper script..."
cat > /usr/local/bin/composer <<'COMPOSERWRAP'
#!/usr/bin/env bash
set -euo pipefail

export HOME="${HOME:-/root}"
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"

# CRITICAL: make Composer scripts (@php) use our php shim
export PHP_BINARY="/usr/local/bin/php"

exec /usr/local/bin/fpcli /usr/local/bin/composer.phar "$@"
COMPOSERWRAP

chmod +x /usr/local/bin/composer
echo "✓ Created composer wrapper at /usr/local/bin/composer"

# Update PATH hash
hash -r 2>/dev/null || true

echo ""
echo "Final verification:"
echo "  composer location: $(which composer)"
echo "  composer.phar location: /usr/local/bin/composer.phar"
echo ""
composer --version
echo ""
echo "✓ Composer now runs through FrankenPHP!"
`
	case "skip":
		composerCmd = `
echo ""
echo "========================================="
echo "Skipping Composer setup"
echo "========================================="
echo ""
echo "You can configure Composer manually later:"
echo ""
echo "Option A - Replace PHP:"
echo "  sudo ln -sf /usr/local/bin/fpcli /usr/local/bin/php"
echo ""
echo "Option C - Wrap Composer:"
echo "  sudo mv /usr/local/bin/composer /usr/local/bin/composer.phar"
echo "  # Then create wrapper script (see docs)"
echo ""
`
	}

	// Combine site creation and composer setup
	fullCmd := siteCmd + composerCmd

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     fullCmd,
			Description: "Creating FrankenPHP site and configuring Composer",
		}
	}
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
	case "composer_setup":
		return m.viewComposerSetup()
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
	// Handle message display (e.g., manual installation instructions)
	if m.message != "" {
		header := m.theme.Title.Render("FrankenPHP Classic Mode")
		messageBox := m.theme.InfoStyle.Render(m.message)
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", messageBox)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

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

	// Error message
	messageSection := ""
	if m.err != nil {
		messageSection = m.theme.ErrorStyle.Render(m.err.Error())
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

// viewSiteSetup renders the site setup view with huh form
func (m FrankenPHPClassicModel) viewSiteSetup() string {
	// Handle message display
	if m.message != "" {
		header := m.theme.Title.Render("FrankenPHP Classic Mode")
		messageBox := m.theme.InfoStyle.Render(m.message)
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", messageBox, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	header := m.theme.Title.Render("FrankenPHP Classic Mode - Site Setup")

	// Binary info
	binaryInfo := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.SuccessStyle.Render("✓ FrankenPHP detected"),
		m.theme.Label.Render("Binary: ")+m.theme.InfoStyle.Render(m.binaryPath),
		m.theme.Label.Render("Version: ")+m.theme.DescriptionStyle.Render(m.binaryVersion),
	)

	// Architecture description
	archInfo := m.theme.DescriptionStyle.Render("Creates: systemd service + Nginx vhost + Unix socket + TCP port fallback")

	// Render the huh form
	formView := ""
	if m.form != nil {
		formView = m.form.View()
	}

	// Error message
	messageSection := ""
	if m.err != nil {
		messageSection = m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Help text with keyboard shortcuts
	help := m.theme.Help.Render("Tab: Next field • Shift+Tab: Previous • Enter: Submit • Esc: Cancel")

	sections := []string{header, "", binaryInfo, "", archInfo, "", formView}
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

	// Show summary of form values
	var summary []string
	summary = append(summary, m.theme.Subtitle.Render("Site Configuration Summary:"))
	summary = append(summary, "")

	// Display all form field values
	if m.formSiteRoot != "" {
		summary = append(summary, m.theme.Label.Render("Site Root: ")+m.theme.InfoStyle.Render(m.formSiteRoot))
	}
	if m.formSiteKey != "" {
		summary = append(summary, m.theme.Label.Render("Site Key: ")+m.theme.InfoStyle.Render(m.formSiteKey))
	}
	if m.formDocroot != "" {
		summary = append(summary, m.theme.Label.Render("Document Root: ")+m.theme.InfoStyle.Render(m.formDocroot))
	}
	if m.formDomains != "" {
		summary = append(summary, m.theme.Label.Render("Domain Names: ")+m.theme.InfoStyle.Render(m.formDomains))
	}
	if m.formPort != "" {
		summary = append(summary, m.theme.Label.Render("Port: ")+m.theme.InfoStyle.Render(m.formPort))
	}
	if m.formUser != "" {
		summary = append(summary, m.theme.Label.Render("Run as User: ")+m.theme.InfoStyle.Render(m.formUser))
	}
	if m.formGroup != "" {
		summary = append(summary, m.theme.Label.Render("Run as Group: ")+m.theme.InfoStyle.Render(m.formGroup))
	}

	// What will be created
	siteKey := m.formSiteKey
	port := m.formPort
	if port == "" {
		port = "8000"
	}

	summary = append(summary, "")
	summary = append(summary, m.theme.Subtitle.Render("Will create:"))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /etc/systemd/system/frankenphp-%s.service", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /etc/nginx/sites-available/%s.conf", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • /run/frankenphp/%s.sock (Unix socket)", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • 127.0.0.1:%s (TCP fallback)", port)))
	summary = append(summary, m.theme.DescriptionStyle.Render("  • /usr/local/bin/fpcli (FrankenPHP CLI wrapper)"))

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

// viewComposerSetup renders the composer setup options view
func (m FrankenPHPClassicModel) viewComposerSetup() string {
	header := m.theme.Title.Render("Composer Setup with FrankenPHP")

	description := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.DescriptionStyle.Render("Configure how Composer should use FrankenPHP's PHP."),
		"",
		m.theme.InfoStyle.Render("The fpcli CLI wrapper will be created at /usr/local/bin/fpcli"),
		m.theme.InfoStyle.Render("Choose how you want Composer to integrate with it:"),
	)

	// Options
	var items []string
	items = append(items, "")

	for i, opt := range m.composerOptions {
		cursor := "  "
		if i == m.composerCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.composerCursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, opt.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, opt.Name))
		}
		items = append(items, renderedItem)

		if i == m.composerCursor && opt.Description != "" {
			items = append(items, "    "+m.theme.DescriptionStyle.Render(opt.Description))
		}
		items = append(items, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Additional info based on selection
	var infoSection string
	if m.composerCursor < len(m.composerOptions) {
		opt := m.composerOptions[m.composerCursor]
		switch opt.ID {
		case "option_a":
			infoSection = lipgloss.JoinVertical(lipgloss.Left,
				m.theme.Subtitle.Render("What this does:"),
				m.theme.DescriptionStyle.Render("  • Creates symlink: /usr/local/bin/php → /usr/local/bin/fpcli"),
				m.theme.DescriptionStyle.Render("  • Composer automatically uses 'php' from PATH"),
				m.theme.DescriptionStyle.Render("  • Laravel @php scripts will work"),
				m.theme.DescriptionStyle.Render("  • System PHP still available at /usr/bin/php"),
			)
		case "option_both":
			infoSection = lipgloss.JoinVertical(lipgloss.Left,
				m.theme.Subtitle.Render("What this does:"),
				m.theme.SuccessStyle.Render("  ✓ Best option for Laravel/Symfony projects"),
				m.theme.DescriptionStyle.Render("  • Creates PHP symlink (php → fpcli)"),
				m.theme.DescriptionStyle.Render("  • Downloads & wraps Composer"),
				m.theme.DescriptionStyle.Render("  • Full compatibility with @php scripts"),
			)
		case "option_c":
			infoSection = lipgloss.JoinVertical(lipgloss.Left,
				m.theme.Subtitle.Render("What this does:"),
				m.theme.DescriptionStyle.Render("  • Downloads composer.phar"),
				m.theme.DescriptionStyle.Render("  • Creates wrapper script that uses fpcli"),
				m.theme.WarningStyle.Render("  ⚠ Laravel @php scripts won't work"),
			)
		}
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	sections := []string{header, "", description, menu}
	if infoSection != "" {
		sections = append(sections, infoSection)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
