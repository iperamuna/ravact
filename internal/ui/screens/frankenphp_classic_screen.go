package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/stubs"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FrankenPHPClassicModel represents the FrankenPHP Classic Mode screen
type FrankenPHPClassicModel struct {
	theme         *theme.Theme
	width         int
	height        int
	cursor        int
	mode          string // "install_options", "site_setup", "confirm", "review_files", "custom_url_input", "composer_setup"
	binaryPath    string
	binaryVersion string
	binaryFound   bool

	// Install options (used when binary not found)
	installOptions []FrankenPHPInstallOption
	customURL      string

	// Current directory (for auto-detection)
	currentDir string

	// Form fields for site setup (huh form)
	form            *huh.Form
	formSiteRoot    string
	formSiteKey     string
	formDocroot     string
	formDomains     string
	formConnType    string // "socket" or "port"
	formPort        string
	formUser        string
	formGroup       string
	formNumThreads  string
	formMaxThreads  string
	formMaxWaitTime string

	// PHP INI fields
	formPHPMemoryLimit              string
	formPHPMaxExecutionTime         string
	formPHPOpcacheEnable            bool
	formPHPOpcacheEnableCli         bool
	formPHPOpcacheMemoryConsumption string
	formPHPOpcacheInternedStrings   string
	formPHPOpcacheMaxFiles          string
	formPHPOpcacheValidate          bool
	formPHPOpcacheRevalidateFreq    string
	formPHPOpcacheJit               bool
	formPHPOpcacheJitBufferSize     string
	formPHPRealpathCacheSize        string
	formPHPRealpathCacheTtl         string
	formPHPMaxUploadSize            string

	// Composer setup options
	composerOptions []ComposerSetupOption
	composerCursor  int

	// Review files state
	generatedFiles []GeneratedFile
	fileCursor     int

	// UI state
	detector *system.Detector
	err      error
	message  string
}

// GeneratedFile represents a config file to be reviewed
type GeneratedFile struct {
	Path    string
	Content string
	Name    string
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

// NewFrankenPHPClassicModelWithSite creates a new FrankenPHP Classic Mode model from an existing Nginx site
func NewFrankenPHPClassicModelWithSite(site system.NginxSite) FrankenPHPClassicModel {
	m := NewFrankenPHPClassicModelWithDir(site.RootDir)
	m.formSiteKey = site.Name
	m.formDomains = site.Domain
	m.formSiteRoot = site.RootDir
	// We skip the installation step if binary is already found
	if m.binaryPath != "" {
		m.mode = "site_setup"
		m.buildSiteSetupForm()
	}
	return m
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
		formDocroot:     "", // Default empty
		formConnType:    "socket",
		formUser:        "www-data",
		formGroup:       "www-data",
		formPort:        "8000",
		formNumThreads:  strconv.Itoa(runtime.NumCPU() * 2),
		formMaxThreads:  "auto",
		formMaxWaitTime: "15",

		// PHP INI defaults
		formPHPMemoryLimit:              "256M",
		formPHPMaxExecutionTime:         "30",
		formPHPOpcacheEnable:            true,
		formPHPOpcacheEnableCli:         true,
		formPHPOpcacheMemoryConsumption: "512",
		formPHPOpcacheInternedStrings:   "32",
		formPHPOpcacheMaxFiles:          "100000",
		formPHPOpcacheValidate:          false,
		formPHPOpcacheRevalidateFreq:    "0",
		formPHPOpcacheJit:               false,
		formPHPOpcacheJitBufferSize:     "0",
		formPHPRealpathCacheSize:        "4096K",
		formPHPRealpathCacheTtl:         "600",
		formPHPMaxUploadSize:            "20",
		detector:                        system.NewDetector(),
	}

	// Default docroot to 'public' if it exists
	publicPath := filepath.Join(m.formSiteRoot, "public")
	if _, err := exec.Command("test", "-d", publicPath).Output(); err == nil {
		m.formDocroot = "public"
	}

	// Build the huh form for site setup
	m.form = m.buildSiteSetupForm()

	return m
}

// buildSiteSetupForm creates the huh form for site configuration
func (m FrankenPHPClassicModel) buildSiteSetupForm() *huh.Form {
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
				Title("Web Directory (relative)").
				Description("Relative path from site root (e.g., 'public' for Laravel). Leave blank to use site root.").
				Placeholder("public").
				Value(&m.formDocroot),

			huh.NewInput().
				Key("domains").
				Title("Domain Names").
				Description("Space-separated domain names for Nginx proxy").
				Placeholder("mysite.com www.mysite.com").
				Value(&m.formDomains),

			huh.NewSelect[string]().
				Key("connType").
				Title("Connection Type").
				Description("How Nginx connects to FrankenPHP").
				Options(
					huh.NewOption("Unix Socket (recommended)", "socket"),
					huh.NewOption("TCP Port", "port"),
				).
				Value(&m.formConnType),

			huh.NewInput().
				Key("port").
				Title("Port").
				Description("Port for FrankenPHP (used when connection type is Port)").
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
					if m.detector != nil && m.detector.IsPortInUse(port) {
						return fmt.Errorf("warning: port %d is already in use by another process", port)
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
		huh.NewGroup(
			huh.NewInput().
				Key("numThreads").
				Title("Number of Threads").
				Description("Suggestion: System Threads * 2").
				Placeholder(strconv.Itoa(runtime.NumCPU()*2)).
				Validate(func(s string) error {
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}).
				Value(&m.formNumThreads),

			huh.NewInput().
				Key("maxThreads").
				Title("Max Threads").
				Description("Accepts positive integer > Number of Threads, or 'auto'").
				Placeholder("auto").
				Validate(func(s string) error {
					if s == "auto" {
						return nil
					}
					v, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("must be a number or 'auto'")
					}
					num, _ := strconv.Atoi(m.formNumThreads)
					if v <= num {
						return fmt.Errorf("must be greater than Number of Threads (%d)", num)
					}
					return nil
				}).
				Value(&m.formMaxThreads),

			huh.NewInput().
				Key("maxWaitTime").
				Title("Max Wait Time").
				Description("Max time to wait for a thread (in seconds)").
				Placeholder("15").
				Validate(func(s string) error {
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}).
				Value(&m.formMaxWaitTime),
		).Title("Performance Tuning"),

		huh.NewGroup(
			huh.NewInput().
				Key("memoryLimit").
				Title("PHP memory_limit").
				Placeholder("256M").
				Value(&m.formPHPMemoryLimit),

			huh.NewInput().
				Key("maxExecTime").
				Title("PHP max_execution_time").
				Placeholder("30").
				Value(&m.formPHPMaxExecutionTime),

			huh.NewInput().
				Key("maxUploadSize").
				Title("Max Upload Size (MB)").
				Description("Accepts positive integer. post_max_size will be +10MB.").
				Placeholder("20").
				Validate(func(s string) error {
					v, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("must be a positive integer")
					}
					if v <= 0 {
						return fmt.Errorf("must be greater than 0")
					}
					return nil
				}).
				Value(&m.formPHPMaxUploadSize),

			huh.NewConfirm().
				Key("opcacheEnable").
				Title("Enable OPcache").
				Value(&m.formPHPOpcacheEnable),

			huh.NewConfirm().
				Key("opcacheCli").
				Title("Enable OPcache CLI").
				Value(&m.formPHPOpcacheEnableCli),

			huh.NewInput().
				Key("opcacheMemory").
				Title("OPcache Memory Consumption (MB)").
				Placeholder("512").
				Value(&m.formPHPOpcacheMemoryConsumption),

			huh.NewInput().
				Key("opcacheStrings").
				Title("OPcache Interned Strings Buffer").
				Placeholder("32").
				Value(&m.formPHPOpcacheInternedStrings),

			huh.NewInput().
				Key("opcacheMaxFiles").
				Title("OPcache Max Accelerated Files").
				Placeholder("100000").
				Value(&m.formPHPOpcacheMaxFiles),

			huh.NewConfirm().
				Key("opcacheValidate").
				Title("OPcache Validate Timestamps").
				Description("Set to false for production optimization").
				Value(&m.formPHPOpcacheValidate),

			huh.NewInput().
				Key("opcacheFreq").
				Title("OPcache Revalidate Frequency").
				Placeholder("0").
				Value(&m.formPHPOpcacheRevalidateFreq),

			huh.NewConfirm().
				Key("jit").
				Title("Enable JIT").
				Value(&m.formPHPOpcacheJit),

			huh.NewInput().
				Key("jitBuffer").
				Title("JIT Buffer Size").
				Placeholder("0").
				Value(&m.formPHPOpcacheJitBufferSize),

			huh.NewInput().
				Key("realpathSize").
				Title("Realpath Cache Size").
				Placeholder("4096K").
				Value(&m.formPHPRealpathCacheSize),

			huh.NewInput().
				Key("realpathTtl").
				Title("Realpath Cache TTL").
				Placeholder("600").
				Value(&m.formPHPRealpathCacheTtl),
		).Title("PHP INIT - Core & Opcashe & Realpath"),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// IdentifyExistingFrankenPHPSetup checks if any FrankenPHP classic mode services exist
func IdentifyExistingFrankenPHPSetup() bool {
	cmd := exec.Command("bash", "-c", `ls /etc/systemd/system/frankenphp-*.service 2>/dev/null | grep -q .`)
	err := cmd.Run()
	return err == nil
}

// IdentifyExistingFrankenPHPSetupForDir checks if a FrankenPHP classic mode service exists for the given directory
func IdentifyExistingFrankenPHPSetupForDir(dir string) bool {
	if dir == "" {
		return false
	}
	// Normalize dir: remove trailing slash
	dir = strings.TrimSuffix(dir, "/")

	// Use grep -E to handle potential quotes and escape special characters in dir
	// We look for WorkingDirectory=/path/to/dir or WorkingDirectory="/path/to/dir"
	escapedDir := strings.ReplaceAll(dir, "/", "\\/")
	// Matches WorkingDirectory=/path/to/dir, WorkingDirectory="/path/to/dir", with optional trailing slash
	pattern := fmt.Sprintf(`WorkingDirectory=(")?%s(\/)?(")?$`, escapedDir)

	cmd := exec.Command("bash", "-c", fmt.Sprintf(`grep -Er '%s' /etc/systemd/system/frankenphp-*.service 2>/dev/null | grep -q .`, pattern))
	err := cmd.Run()
	return err == nil
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
	// Check if a service already exists for this directory
	if IdentifyExistingFrankenPHPSetupForDir(m.currentDir) {
		return func() tea.Msg {
			return NavigateMsg{
				Screen: FrankenPHPServicesScreen,
				Data: map[string]interface{}{
					"filterDir": m.currentDir,
				},
			}
		}
	}

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
					// Yes - create the site, then show review files
					m = m.generateConfigFiles()
					m.mode = "review_files"
					m.fileCursor = 0
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

		// Handle review_files mode
		if m.mode == "review_files" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.mode = "confirm"
				m.cursor = 0
				return m, nil
			case "up", "k":
				if m.fileCursor > 0 {
					m.fileCursor--
				}
			case "down", "j":
				if m.fileCursor < len(m.generatedFiles)-1 {
					m.fileCursor++
				}
			case "v", "enter":
				// View file content internally
				m.mode = "view_file"
				return m, nil
			case "e":
				// Edit file content with nano or vi
				if m.fileCursor < len(m.generatedFiles) {
					file := m.generatedFiles[m.fileCursor]
					// Write to temp file
					tmpFile := filepath.Join(os.TempDir(), "ravact-"+file.Name)
					os.WriteFile(tmpFile, []byte(file.Content), 0644)

					return m, func() tea.Msg {
						return NavigateMsg{
							Screen: EditorSelectionScreen,
							Data: map[string]interface{}{
								"file":        tmpFile,
								"description": fmt.Sprintf("Editing %s", file.Name),
							},
						}
					}
				}
			case "d":
				// Proceed to deploy confirmation
				m.mode = "confirm_deploy"
				return m, nil
			}
			return m, nil
		}

		// Handle view_file mode (internal preview)
		if m.mode == "view_file" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "enter", "v", "backspace":
				m.mode = "review_files"
				return m, nil
			case "d":
				m.mode = "confirm_deploy"
				return m, nil
			}
			return m, nil
		}

		// Handle confirm_deploy mode
		if m.mode == "confirm_deploy" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace", "n":
				m.mode = "review_files"
				return m, nil
			case "enter", "y", "d":
				// All done, proceed to deploy and then composer setup
				m.mode = "composer_setup"
				m.composerCursor = 0
				return m, nil
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
	case EditorCompleteMsg:
		if msg.Error == "" && m.mode == "review_files" && m.fileCursor < len(m.generatedFiles) {
			file := &m.generatedFiles[m.fileCursor]
			tmpFile := filepath.Join(os.TempDir(), "ravact-"+file.Name)
			if content, err := os.ReadFile(tmpFile); err == nil {
				file.Content = string(content)
				// Clean up
				os.Remove(tmpFile)
			}
		}
		return m, nil
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
			// Docroot can be empty (means use site root)
			m.formDocroot = m.form.GetString("docroot")
			if v := m.form.GetString("domains"); v != "" {
				m.formDomains = v
			}
			if v := m.form.GetString("connType"); v != "" {
				m.formConnType = v
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
			m = m.autoFillFields()
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
func (m FrankenPHPClassicModel) autoFillFields() FrankenPHPClassicModel {
	if m.formSiteRoot == "" {
		return m
	}

	// Auto-fill site key from site root
	if m.formSiteKey == "" {
		m.formSiteKey = suggestSiteKey(m.formSiteRoot)
	}

	// Auto-fill domains
	if m.formDomains == "" && m.formSiteKey != "" {
		m.formDomains = m.formSiteKey + ".test"
	}

	// Default connection type
	if m.formConnType == "" {
		m.formConnType = "socket"
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
	return m
}

// getFullDocroot returns the full document root path
func (m FrankenPHPClassicModel) getFullDocroot() string {
	if m.formDocroot == "" {
		// If no docroot specified, use site root
		return m.formSiteRoot
	}
	// If docroot is already absolute, use it as-is
	if strings.HasPrefix(m.formDocroot, "/") {
		return m.formDocroot
	}
	// Otherwise, join with site root
	return filepath.Join(m.formSiteRoot, m.formDocroot)
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

func (m FrankenPHPClassicModel) buildCreateSiteCommand() string {
	// Get values from form fields
	siteKey := m.formSiteKey
	siteRoot := m.formSiteRoot
	user := m.formUser
	group := m.formGroup
	binaryPath := m.binaryPath
	if binaryPath == "" {
		binaryPath = "/usr/local/bin/frankenphp"
	}

	var script strings.Builder
	script.WriteString("#!/bin/bash\nset -e\n\n")

	script.WriteString(fmt.Sprintf("echo \"Creating FrankenPHP Classic Mode site: %s\"\n", siteKey))
	script.WriteString(fmt.Sprintf("echo \"  Site Root: %s\"\n", siteRoot))
	script.WriteString("echo \"\"\n")

	// Determine the system user (owner)
	systemUser := getGitSystemUser()
	if systemUser == "" {
		systemUser = os.Getenv("USER")
	}

	// Create directories and set permissions
	script.WriteString(fmt.Sprintf("sudo mkdir -p /etc/frankenphp/%s\n", siteKey))
	script.WriteString("sudo mkdir -p /run/frankenphp\n")
	script.WriteString(fmt.Sprintf("sudo chown %s:%s /run/frankenphp\n", user, group))

	// Base /var/lib/caddy setup
	script.WriteString("sudo mkdir -p /var/lib/caddy\n")
	script.WriteString(fmt.Sprintf("sudo chown -R %s:%s /var/lib/caddy\n", user, group))
	script.WriteString("sudo chmod -R 750 /var/lib/caddy\n")

	// Ensure system user is in web group
	script.WriteString(fmt.Sprintf("if ! groups %s | grep -q \"\\b%s\\b\"; then\n", systemUser, group))
	script.WriteString(fmt.Sprintf("    sudo usermod -a -G %s %s\n", group, systemUser))
	script.WriteString("fi\n")

	// Create site-specific storage directory structure
	script.WriteString(fmt.Sprintf("sudo mkdir -p /var/lib/caddy/%s/config\n", siteKey))
	script.WriteString(fmt.Sprintf("sudo mkdir -p /var/lib/caddy/%s/data\n", siteKey))
	script.WriteString(fmt.Sprintf("sudo mkdir -p /var/lib/caddy/%s/tls\n", siteKey))

	// Set site-specific permissions
	script.WriteString(fmt.Sprintf("sudo chown -R %s:%s /var/lib/caddy/%s\n", systemUser, group, siteKey))
	script.WriteString(fmt.Sprintf("sudo chmod -R 775 /var/lib/caddy/%s\n", siteKey))

	// Write generated files (this includes Caddyfile, Service, php.ini, Nginx, fpcli)
	for _, file := range m.generatedFiles {
		script.WriteString(fmt.Sprintf("\nif [ -f \"%s\" ]; then\n", file.Path))
		script.WriteString(fmt.Sprintf("    echo \"Backing up existing %s...\"\n", file.Path))
		script.WriteString(fmt.Sprintf("    cp \"%s\" \"%s.bak\"\n", file.Path, file.Path))
		script.WriteString("fi\n")
		// Use heredoc to write content safely
		script.WriteString(fmt.Sprintf("cat > \"%s\" <<'EOF'\n", file.Path))
		script.WriteString(file.Content)
		script.WriteString("\nEOF\n")
	}

	// Fix permissions and enable services
	script.WriteString("\n# Fix permissions and enable services\n")
	caddyfilePath := fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", siteKey)
	script.WriteString(fmt.Sprintf("%s fmt --overwrite %s\n", binaryPath, caddyfilePath))

	// Ensure config permission
	script.WriteString(fmt.Sprintf("sudo chown -R %s:%s /etc/frankenphp/%s\n", user, group, siteKey))

	serviceName := fmt.Sprintf("frankenphp-%s", siteKey)
	script.WriteString("sudo systemctl daemon-reload\n")
	script.WriteString(fmt.Sprintf("sudo systemctl enable --now %s\n", serviceName))
	script.WriteString(fmt.Sprintf("echo \"✓ Service %s enabled and started\"\n", serviceName))

	// Set executable bit for fpcli
	script.WriteString("\nchmod +x /usr/local/bin/fpcli 2>/dev/null || true\n")
	script.WriteString(fmt.Sprintf("chown -R %s:%s /etc/frankenphp/%s\n", user, group, siteKey))

	script.WriteString("\n# Verification phase\n")
	script.WriteString("set +e\n")
	script.WriteString("echo \"\"\n")
	script.WriteString("echo \"=========================================\"\n")
	script.WriteString("echo \"🔍 Final Verification\"\n")
	script.WriteString("echo \"=========================================\"\n")
	script.WriteString("echo \"Checking service status...\"\n")
	script.WriteString("sleep 1\n")
	script.WriteString(fmt.Sprintf("\nif sudo systemctl is-active --quiet \"%s\"; then\n", serviceName))
	script.WriteString("    echo \"✓ FrankenPHP service is active\"\n")
	script.WriteString("else\n")
	script.WriteString("    echo \"✗ FrankenPHP service is NOT active!\"\n")
	script.WriteString(fmt.Sprintf("    echo \"    Diagnostic: sudo systemctl status %s --no-pager -l\"\n", serviceName))
	script.WriteString(fmt.Sprintf("    sudo systemctl status %s --no-pager -l\n", serviceName))
	script.WriteString("fi\n")

	script.WriteString("\necho \"Checking PHP configuration...\"\n")
	phpIniPath := fmt.Sprintf("/etc/frankenphp/%s/app-php.ini", siteKey)
	script.WriteString(fmt.Sprintf("if [ -f \"%s\" ]; then\n", phpIniPath))
	script.WriteString(fmt.Sprintf("    RAW_INI_OUTPUT=$(%s php-cli -c %s --ini 2>&1)\n", binaryPath, phpIniPath))
	script.WriteString("    LOADED_INI=$(echo \"$RAW_INI_OUTPUT\" | grep \"Loaded Configuration File\" | awk '{print $NF}')\n")
	script.WriteString(fmt.Sprintf("    if [ \"$LOADED_INI\" = \"%s\" ]; then\n", phpIniPath))
	script.WriteString("        echo \"  ✓ Custom PHP INI loaded correctly\"\n")
	script.WriteString("    else\n")
	script.WriteString("        echo \"  ✗ Custom PHP INI NOT loaded\"\n")
	script.WriteString("        echo \"    Output: $LOADED_INI\"\n")
	script.WriteString("        if [ -z \"$LOADED_INI\" ]; then\n")
	script.WriteString("            echo \"    Error Details: $RAW_INI_OUTPUT\"\n")
	script.WriteString("        fi\n")
	script.WriteString("    fi\n")
	script.WriteString("else\n")
	script.WriteString("    echo \"  ✗ PHP INI template not found at $phpIniPath\"\n")
	script.WriteString("fi\n")

	return script.String()
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
echo "⚙️  Setting up PHP Symlink"
echo "========================================="
set -e
# Backup existing php if it exists and is not a symlink
if [ -f /usr/local/bin/php ] && [ ! -L /usr/local/bin/php ]; then
    echo "  Backing up /usr/local/bin/php to php.bak"
    mv /usr/local/bin/php /usr/local/bin/php.bak
fi
ln -sf /usr/local/bin/fpcli /usr/local/bin/php
hash -r 2>/dev/null || true
echo "  ✓ Created /usr/local/bin/php -> /usr/local/bin/fpcli"
set +e
echo ""
echo "Verification:"
ls -la /usr/local/bin/php
php -v | head -n 1
`
	case "option_both":
		// Option A+C: Both PHP symlink and Composer wrapper
		composerCmd = `
echo ""
echo "========================================="
echo "🚀 Setting up Full Composer Integration"
echo "========================================="
set -e
# Part 1: PHP Symlink
if [ -f /usr/local/bin/php ] && [ ! -L /usr/local/bin/php ]; then
    mv /usr/local/bin/php /usr/local/bin/php.bak
fi
ln -sf /usr/local/bin/fpcli /usr/local/bin/php
echo "  ✓ PHP symlink created"

# Part 2: Composer Wrapper
if [ ! -f /usr/local/bin/composer.phar ]; then
    echo "  Downloading Composer..."
    curl -sS https://getcomposer.org/installer | /usr/local/bin/fpcli - -- --install-dir=/usr/local/bin --filename=composer.phar
fi
chmod +x /usr/local/bin/composer.phar

cat > /usr/local/bin/composer <<'COMPWRAP'
#!/usr/bin/env bash
set -e
export PHP_BINARY="/usr/local/bin/php"
exec /usr/local/bin/fpcli /usr/local/bin/composer.phar "$@"
COMPWRAP
chmod +x /usr/local/bin/composer
echo "  ✓ Composer wrapper created"
hash -r 2>/dev/null || true
set +e
echo ""
echo "Verification:"
composer --version
`
	case "option_c":
		// Option C: Create composer wrapper only
		composerCmd = `
echo ""
echo "========================================="
echo "⚙️  Setting up Composer Wrapper"
echo "========================================="
set -e
if [ ! -f /usr/local/bin/composer.phar ]; then
    echo "  Downloading Composer..."
    curl -sS https://getcomposer.org/installer | /usr/local/bin/fpcli - -- --install-dir=/usr/local/bin --filename=composer.phar
fi
chmod +x /usr/local/bin/composer.phar

cat > /usr/local/bin/composer <<'COMPWRAP'
#!/usr/bin/env bash
set -e
exec /usr/local/bin/fpcli /usr/local/bin/composer.phar "$@"
COMPWRAP
chmod +x /usr/local/bin/composer
echo "  ✓ Composer wrapper created"
hash -r 2>/dev/null || true
set +e
echo ""
echo "Verification:"
composer --version
`
	case "skip":
		composerCmd = `
echo ""
echo "========================================="
echo "⏭️  Skipping Composer Setup"
echo "========================================="
echo " FrankenPHP site is ready."
`
	}

	// Combine site creation and composer setup
	fullCmd := siteCmd + composerCmd

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     fullCmd,
			Description: "Setting up FrankenPHP Site and Composer integration",
		}
	}
}

// generateConfigFiles generates the content for the required config files
func (m FrankenPHPClassicModel) generateConfigFiles() FrankenPHPClassicModel {
	m.generatedFiles = []GeneratedFile{}

	id := m.formSiteKey

	// 1. Caddyfile
	caddyTemplate := m.generateCaddyfileContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "Caddyfile",
		Path:    fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", id),
		Content: caddyTemplate,
	})

	// 2. Systemd Service
	serviceTemplate := m.generateServiceFileContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "Systemd Service",
		Path:    fmt.Sprintf("/etc/systemd/system/frankenphp-%s.service", id),
		Content: serviceTemplate,
	})

	// 3. fpcli Wrapper
	fpcliTemplate := m.generateFpcliContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "fpcli Wrapper",
		Path:    "/usr/local/bin/fpcli",
		Content: fpcliTemplate,
	})
	return m
}

func (m FrankenPHPClassicModel) generateCaddyfileContent() string {
	id := m.formSiteKey
	docroot := m.getFullDocroot()
	port := m.formPort
	if port == "" {
		port = "8000"
	}

	numThreads := m.formNumThreads
	maxThreads := m.formMaxThreads
	maxWaitTime := m.formMaxWaitTime

	var bindLine string
	if m.formConnType == "socket" {
		bindLine = fmt.Sprintf("bind unix//run/frankenphp/%s.sock", id)
	} else {
		bindLine = fmt.Sprintf("bind 127.0.0.1:%s", port)
	}

	// Calculate upload sizes
	uploadMax := m.formPHPMaxUploadSize
	if uploadMax == "" {
		uploadMax = "20"
	}
	uploadInt, _ := strconv.Atoi(uploadMax)
	postMax := strconv.Itoa(uploadInt + 10)

	// Build PHP directives
	var phpDirectives strings.Builder
	settings := map[string]string{
		"memory_limit":                    m.formPHPMemoryLimit,
		"max_execution_time":              m.formPHPMaxExecutionTime,
		"upload_max_filesize":             uploadMax + "M",
		"post_max_size":                   postMax + "M",
		"opcache.enable":                  "0",
		"opcache.enable_cli":              "0",
		"opcache.memory_consumption":      m.formPHPOpcacheMemoryConsumption,
		"opcache.interned_strings_buffer": m.formPHPOpcacheInternedStrings,
		"opcache.max_accelerated_files":   m.formPHPOpcacheMaxFiles,
		"opcache.validate_timestamps":     "0",
		"opcache.revalidate_freq":         m.formPHPOpcacheRevalidateFreq,
		"opcache.jit":                     "0",
		"opcache.jit_buffer_size":         m.formPHPOpcacheJitBufferSize,
		"realpath_cache_size":             m.formPHPRealpathCacheSize,
		"realpath_cache_ttl":              m.formPHPRealpathCacheTtl,
	}

	if m.formPHPOpcacheEnable {
		settings["opcache.enable"] = "1"
	}
	if m.formPHPOpcacheEnableCli {
		settings["opcache.enable_cli"] = "1"
	}
	if m.formPHPOpcacheValidate {
		settings["opcache.validate_timestamps"] = "1"
	}
	if m.formPHPOpcacheJit {
		settings["opcache.jit"] = "1255"
	}

	keys := []string{
		"memory_limit", "max_execution_time", "upload_max_filesize", "post_max_size", "opcache.enable", "opcache.enable_cli",
		"opcache.memory_consumption", "opcache.interned_strings_buffer", "opcache.max_accelerated_files",
		"opcache.validate_timestamps", "opcache.revalidate_freq", "opcache.jit",
		"opcache.jit_buffer_size", "realpath_cache_size", "realpath_cache_ttl",
	}

	for _, k := range keys {
		if v, ok := settings[k]; ok && v != "" {
			phpDirectives.WriteString(fmt.Sprintf("\t\tphp_ini %s %s\n", k, v))
		}
	}

	requestBody := fmt.Sprintf("request_body {\n\t\tmax_size %sMB\n\t}", uploadMax)

	content, err := stubs.LoadAndReplace("caddyfile", map[string]string{
		"SITE_KEY":       id,
		"NUM_THREADS":    numThreads,
		"MAX_THREADS":    maxThreads,
		"MAX_WAIT_TIME":  maxWaitTime,
		"PORT":           port,
		"BIND_LINE":      bindLine,
		"REQUEST_BODY":   requestBody,
		"DOCROOT":        docroot,
		"PHP_DIRECTIVES": strings.TrimSpace(phpDirectives.String()),
	})
	if err != nil {
		return fmt.Sprintf("Error loading caddyfile stub: %v", err)
	}

	return content
}

func (m FrankenPHPClassicModel) generateServiceFileContent() string {
	id := m.formSiteKey
	siteRoot := m.formSiteRoot
	user := m.formUser
	group := m.formGroup
	binary := m.binaryPath
	if binary == "" {
		binary = "/usr/local/bin/frankenphp"
	}

	var preStart string
	var postStart string
	if m.formConnType == "socket" {
		preStart = fmt.Sprintf("ExecStartPre=/usr/bin/rm -f /run/frankenphp/%s.sock\n", id)
		postStart = fmt.Sprintf("ExecStartPost=/bin/sh -c 'for i in $(seq 1 50); do [ -S /run/frankenphp/%s.sock ] && chmod 0660 /run/frankenphp/%s.sock && exit 0; sleep 0.1; done; echo \"Socket not created: /run/frankenphp/%s.sock\" >&2; exit 1'\n", id, id, id)
	}

	caddyfile := fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", id)

	content, err := stubs.LoadAndReplace("service", map[string]string{
		"ID":                id,
		"USER":              user,
		"GROUP":             group,
		"WORKING_DIRECTORY": siteRoot,
		"APP_BASE_PATH":     siteRoot,
		"PRE_START":         preStart,
		"BINARY":            binary,
		"CADDYFILE":         caddyfile,
		"POST_START":        postStart,
	})
	if err != nil {
		return fmt.Sprintf("Error loading service stub: %v", err)
	}

	return content
}

// generateFpcliContent generates the fpcli CLI wrapper script
func (m FrankenPHPClassicModel) generateFpcliContent() string {
	binary := m.binaryPath
	if binary == "" {
		binary = "/usr/local/bin/frankenphp"
	}

	content, err := stubs.LoadAndReplace("fpcli", map[string]string{
		"BINARY": binary,
	})
	if err != nil {
		return fmt.Sprintf("Error loading fpcli stub: %v", err)
	}

	return content
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
	case "review_files":
		return m.viewReviewFiles()
	case "view_file":
		return m.viewFileContent()
	case "confirm_deploy":
		return m.viewConfirmDeploy()
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
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewInstallOptions renders the installation options view
func (m FrankenPHPClassicModel) viewInstallOptions() string {
	// Handle message display (e.g., manual installation instructions)
	if m.message != "" {
		header := m.theme.Title.Render("FrankenPHP Classic Mode")
		messageBox := m.theme.InfoStyle.Render(m.message)
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", messageBox)
		bordered := m.theme.RenderBox(content)
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
	bordered := m.theme.RenderBox(content)
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
		bordered := m.theme.RenderBox(content)
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
	archInfo := m.theme.DescriptionStyle.Render("Creates: systemd service + FrankenPHP Caddyfile + Unix socket + TCP port fallback")

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
	bordered := m.theme.RenderBox(content)
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

	// Performance Tuning
	summary = append(summary, "")
	summary = append(summary, m.theme.Subtitle.Render("Performance Tuning:"))
	summary = append(summary, m.theme.Label.Render("Threads:       ")+m.theme.InfoStyle.Render(fmt.Sprintf("%s (max: %s)", m.formNumThreads, m.formMaxThreads)))
	summary = append(summary, m.theme.Label.Render("Max Wait:      ")+m.theme.InfoStyle.Render(m.formMaxWaitTime+"s"))

	// PHP INI
	summary = append(summary, "")
	summary = append(summary, m.theme.Subtitle.Render("PHP Configuration:"))
	summary = append(summary, m.theme.Label.Render("Memory Limit:  ")+m.theme.InfoStyle.Render(m.formPHPMemoryLimit))
	summary = append(summary, m.theme.Label.Render("Max Exec Time: ")+m.theme.InfoStyle.Render(m.formPHPMaxExecutionTime+"s"))
	summary = append(summary, m.theme.Label.Render("Max Upload:    ")+m.theme.InfoStyle.Render(m.formPHPMaxUploadSize+"MB"))
	opcacheStatus := "Enabled"
	if !m.formPHPOpcacheEnable {
		opcacheStatus = "Disabled"
	}
	summary = append(summary, m.theme.Label.Render("OPcache:       ")+m.theme.InfoStyle.Render(opcacheStatus))

	// What will be created
	siteKey := m.formSiteKey
	port := m.formPort
	if port == "" {
		port = "8000"
	}

	summary = append(summary, "")
	summary = append(summary, m.theme.Subtitle.Render("Will generate and deploy:"))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s", m.theme.Label.Render("systemd service: "))+fmt.Sprintf("/etc/systemd/system/frankenphp-%s.service", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s", m.theme.Label.Render("FrankenPHP Caddyfile: "))+fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s", m.theme.Label.Render("Custom app-php.ini: "))+fmt.Sprintf("/etc/frankenphp/%s/app-php.ini", siteKey)))
	summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s", m.theme.Label.Render("CLI wrapper script: "))+"/usr/local/bin/fpcli"))

	if m.formConnType == "socket" {
		summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s /run/frankenphp/%s.sock", m.theme.Label.Render("Unix Socket:"), siteKey)))
	} else {
		summary = append(summary, m.theme.DescriptionStyle.Render(fmt.Sprintf("  • %s 127.0.0.1:%s", m.theme.Label.Render("TCP Port:"), port)))
	}

	summarySection := lipgloss.JoinVertical(lipgloss.Left, summary...)

	// Yes/No options
	var options []string
	options = append(options, "")
	choices := []string{"Review and Confirm Configuration files", "No, go back"}
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

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", summarySection, optionsSection, "", help)
	bordered := m.theme.RenderBox(content)
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
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewReviewFiles renders the file review view
func (m FrankenPHPClassicModel) viewReviewFiles() string {
	header := m.theme.Title.Render("Review Configuration Files")

	description := m.theme.DescriptionStyle.Render("Review and optionally edit the files that will be created.")

	var items []string
	items = append(items, "")
	for i, file := range m.generatedFiles {
		cursor := "  "
		if i == m.fileCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.fileCursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, file.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, file.Name))
		}
		items = append(items, renderedItem)
		items = append(items, "    "+m.theme.DescriptionStyle.Render(file.Path))
		items = append(items, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	statusInfo := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.theme.Subtitle.Render("Actions:"),
		m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s: View/Preview file content", m.theme.KeyStyle.Render("Enter/v"))),
		m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s: Edit file (select editor)", m.theme.KeyStyle.Render("e"))),
		m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s: Proceed to Deployment", m.theme.KeyStyle.Render("d"))),
	)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: View • e: Edit • d: Deploy • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", description, "", menu, statusInfo, "", help)
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewFileContent renders the content of a single generated file
func (m FrankenPHPClassicModel) viewFileContent() string {
	if m.fileCursor >= len(m.generatedFiles) {
		return "No file selected"
	}

	file := m.generatedFiles[m.fileCursor]
	header := m.theme.Title.Render(fmt.Sprintf("Preview: %s", file.Name))
	path := m.theme.DescriptionStyle.Render(file.Path)

	// Wrap content in a style
	content := m.theme.MenuItem.Render(file.Content)

	help := m.theme.Help.Render("Esc/Enter/v: Back to List • d: Proceed to Deployment • q: Quit")

	sections := []string{
		header,
		path,
		"",
		content,
		"",
		help,
	}

	contentSection := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.RenderBox(contentSection)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewConfirmDeploy renders the final deployment confirmation
func (m FrankenPHPClassicModel) viewConfirmDeploy() string {
	header := m.theme.Title.Render("Final Deployment Confirmation")

	message := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Subtitle.Render("Are you sure you want to deploy the site now?"),
		"",
		m.theme.DescriptionStyle.Render("This will:"),
		m.theme.DescriptionStyle.Render("  • Create all configuration files"),
		m.theme.DescriptionStyle.Render("  • Run systemctl daemon-reload"),
		m.theme.DescriptionStyle.Render("  • Enable and start the systemd service"),
		m.theme.DescriptionStyle.Render("  • Create Nginx symbolic link and test config"),
		m.theme.DescriptionStyle.Render("  • Configure Composer integration"),
		m.theme.SuccessStyle.Render("  • Run final verification checks"),
		m.theme.WarningStyle.Render("  • (Nginx reload must be done manually if needed)"),
		"",
		m.theme.InfoStyle.Render("You can still review the verification results after deployment."),
	)

	choices := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.theme.SuccessStyle.Render("  Enter/d/y: Yes, Deploy now"),
		m.theme.DescriptionStyle.Render("  Esc/n: No, back to review"),
	)

	help := m.theme.Help.Render("Enter: Confirm Deployment • Esc: Cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", message, choices, "", help)
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
