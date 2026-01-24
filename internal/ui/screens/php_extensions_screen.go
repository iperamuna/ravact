package screens

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PHPExtension represents a PHP extension
type PHPExtension struct {
	Name        string
	Description string
	Selected    bool
}

// PHPExtensionsModel represents the PHP extensions selection screen
type PHPExtensionsModel struct {
	theme             *theme.Theme
	width             int
	height            int
	cursor            int
	extensions        []PHPExtension
	filteredIndices   []int
	searchQuery       string
	searchMode        bool
	installedVersions []string
	selectedVersion   string
	versionCursor     int
	mode              string // "version_select", "extensions", "confirm"
	err               error
	success           string
	scrollOffset      int
	maxVisible        int
}

// Available PHP extensions (non-default ones that users might want to install)
var availablePHPExtensions = []PHPExtension{
	{Name: "redis", Description: "Redis client extension"},
	{Name: "memcached", Description: "Memcached client extension"},
	{Name: "mongodb", Description: "MongoDB driver"},
	{Name: "imagick", Description: "ImageMagick image processing"},
	{Name: "xdebug", Description: "Debugging and profiling"},
	{Name: "apcu", Description: "APC User Cache for data caching"},
	{Name: "uuid", Description: "UUID generation functions"},
	{Name: "yaml", Description: "YAML parsing and emitting"},
	{Name: "igbinary", Description: "Binary serialization"},
	{Name: "msgpack", Description: "MessagePack serialization"},
	{Name: "swoole", Description: "Coroutine-based async framework"},
	{Name: "grpc", Description: "gRPC client library"},
	{Name: "protobuf", Description: "Protocol Buffers"},
	{Name: "mailparse", Description: "Email message parsing"},
	{Name: "imap", Description: "IMAP email protocol support"},
	{Name: "ldap", Description: "LDAP directory access"},
	{Name: "snmp", Description: "SNMP protocol support"},
	{Name: "ssh2", Description: "SSH2 protocol bindings"},
	{Name: "tidy", Description: "HTML/XHTML tidying"},
	{Name: "xsl", Description: "XSL transformations"},
	{Name: "enchant", Description: "Spell checking"},
	{Name: "pspell", Description: "Spell checking with Pspell"},
	{Name: "gmp", Description: "GNU Multiple Precision arithmetic"},
	{Name: "bz2", Description: "Bzip2 compression"},
	{Name: "lz4", Description: "LZ4 compression"},
	{Name: "zstd", Description: "Zstandard compression"},
	{Name: "raphf", Description: "Resource and persistent handles factory"},
	{Name: "http", Description: "Extended HTTP support (pecl_http)"},
	{Name: "ast", Description: "Abstract Syntax Tree"},
	{Name: "ds", Description: "Data Structures extension"},
	{Name: "decimal", Description: "Arbitrary precision decimal"},
	{Name: "pcov", Description: "Code coverage driver"},
	{Name: "ev", Description: "Event loop extension"},
	{Name: "event", Description: "Event library bindings"},
	{Name: "ffi", Description: "Foreign Function Interface"},
	{Name: "sockets", Description: "Low-level socket interface"},
	{Name: "pdo-firebird", Description: "PDO Firebird driver"},
	{Name: "pdo-odbc", Description: "PDO ODBC driver"},
	{Name: "odbc", Description: "ODBC database access"},
	{Name: "dba", Description: "Database abstraction layer"},
	{Name: "interbase", Description: "Firebird/InterBase database"},
	{Name: "sybase", Description: "Sybase database support"},
}

// NewPHPExtensionsModel creates a new PHP extensions model
func NewPHPExtensionsModel() PHPExtensionsModel {
	installedVersions := detectInstalledPHPVersions()

	// Copy extensions list
	extensions := make([]PHPExtension, len(availablePHPExtensions))
	copy(extensions, availablePHPExtensions)

	// Initialize filtered indices to show all
	filteredIndices := make([]int, len(extensions))
	for i := range extensions {
		filteredIndices[i] = i
	}

	selectedVersion := ""
	if len(installedVersions) > 0 {
		selectedVersion = installedVersions[0]
	}

	return PHPExtensionsModel{
		theme:             theme.DefaultTheme(),
		cursor:            0,
		extensions:        extensions,
		filteredIndices:   filteredIndices,
		installedVersions: installedVersions,
		selectedVersion:   selectedVersion,
		mode:              "version_select",
		maxVisible:        15,
	}
}

// Init initializes the PHP extensions screen
func (m PHPExtensionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for PHP extensions
func (m PHPExtensionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust max visible based on height
		m.maxVisible = (m.height - 20) / 2
		if m.maxVisible < 5 {
			m.maxVisible = 5
		}
		return m, nil

	case tea.KeyMsg:
		// Handle search mode
		if m.searchMode {
			switch msg.String() {
			case "esc":
				m.searchMode = false
				return m, nil
			case "enter":
				m.searchMode = false
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filterExtensions()
					m.cursor = 0
					m.scrollOffset = 0
				}
			default:
				char := msg.String()
				if len(char) == 1 && ((char[0] >= 'a' && char[0] <= 'z') || (char[0] >= 'A' && char[0] <= 'Z') || (char[0] >= '0' && char[0] <= '9') || char[0] == '-' || char[0] == '_') {
					m.searchQuery += strings.ToLower(char)
					m.filterExtensions()
					m.cursor = 0
					m.scrollOffset = 0
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			if m.mode == "extensions" {
				m.mode = "version_select"
				return m, nil
			} else if m.mode == "confirm" {
				m.mode = "extensions"
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigateMsg{Screen: PHPInstallScreen}
			}

		case "/":
			if m.mode == "extensions" {
				m.searchMode = true
				m.searchQuery = ""
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			maxIdx := m.getMaxIndex()
			if m.cursor < maxIdx {
				m.cursor++
				// Adjust scroll
				if m.cursor >= m.scrollOffset+m.maxVisible {
					m.scrollOffset = m.cursor - m.maxVisible + 1
				}
			}

		case " ":
			// Toggle selection in extensions mode
			if m.mode == "extensions" && len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
				idx := m.filteredIndices[m.cursor]
				m.extensions[idx].Selected = !m.extensions[idx].Selected
			}

		case "enter":
			return m.executeAction()
		}
	}

	return m, nil
}

// getMaxIndex returns the maximum cursor index based on mode
func (m PHPExtensionsModel) getMaxIndex() int {
	switch m.mode {
	case "version_select":
		return len(m.installedVersions) - 1
	case "extensions":
		return len(m.filteredIndices) - 1
	case "confirm":
		return 1 // Yes/No
	}
	return 0
}

// filterExtensions filters the extensions based on search query
func (m *PHPExtensionsModel) filterExtensions() {
	m.filteredIndices = []int{}
	query := strings.ToLower(m.searchQuery)

	for i, ext := range m.extensions {
		if query == "" || strings.Contains(strings.ToLower(ext.Name), query) || strings.Contains(strings.ToLower(ext.Description), query) {
			m.filteredIndices = append(m.filteredIndices, i)
		}
	}
}

// executeAction handles the selected action
func (m PHPExtensionsModel) executeAction() (PHPExtensionsModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	switch m.mode {
	case "version_select":
		if m.cursor < len(m.installedVersions) {
			m.selectedVersion = m.installedVersions[m.cursor]
			m.mode = "extensions"
			m.cursor = 0
			m.scrollOffset = 0
			// Reset selections
			for i := range m.extensions {
				m.extensions[i].Selected = false
			}
		}

	case "extensions":
		// Count selected
		selected := m.getSelectedExtensions()
		if len(selected) == 0 {
			m.err = fmt.Errorf("no extensions selected. Use Space to select extensions")
			return m, nil
		}
		m.mode = "confirm"
		m.cursor = 0

	case "confirm":
		if m.cursor == 0 {
			// Yes - install
			selected := m.getSelectedExtensions()
			cmd := m.buildInstallCommand(selected)
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     cmd,
					Description: fmt.Sprintf("Installing %d extensions for PHP %s", len(selected), m.selectedVersion),
				}
			}
		} else {
			// No - go back
			m.mode = "extensions"
			m.cursor = 0
		}
	}

	return m, nil
}

// getSelectedExtensions returns list of selected extension names
func (m PHPExtensionsModel) getSelectedExtensions() []string {
	var selected []string
	for _, ext := range m.extensions {
		if ext.Selected {
			selected = append(selected, ext.Name)
		}
	}
	return selected
}

// buildInstallCommand creates the apt command to install extensions
func (m PHPExtensionsModel) buildInstallCommand(extensions []string) string {
	var packages []string
	for _, ext := range extensions {
		packages = append(packages, fmt.Sprintf("php%s-%s", m.selectedVersion, ext))
	}

	return fmt.Sprintf(`apt-get update && apt-get install -y %s
systemctl restart php%s-fpm 2>/dev/null || true
echo "Extensions installed successfully!"
php%s -m | grep -E "(%s)"`, strings.Join(packages, " "), m.selectedVersion, m.selectedVersion, strings.Join(extensions, "|"))
}

// getInstalledExtensions returns extensions installed for a PHP version
func getInstalledExtensions(version string) []string {
	cmd := exec.Command(fmt.Sprintf("php%s", version), "-m")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}
	
	lines := strings.Split(string(output), "\n")
	var extensions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "[") {
			extensions = append(extensions, strings.ToLower(line))
		}
	}
	return extensions
}

// View renders the PHP extensions screen
func (m PHPExtensionsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.mode {
	case "version_select":
		return m.viewVersionSelect()
	case "extensions":
		return m.viewExtensions()
	case "confirm":
		return m.viewConfirm()
	}

	return "Unknown mode"
}

// viewVersionSelect renders the version selection view
func (m PHPExtensionsModel) viewVersionSelect() string {
	header := m.theme.Title.Render("PHP Extensions - Select Version")

	if len(m.installedVersions) == 0 {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			m.theme.WarningStyle.Render("No PHP versions installed."),
			m.theme.DescriptionStyle.Render("Please install a PHP version first."),
			"",
			m.theme.Help.Render("Esc: Back • q: Quit"),
		)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	var items []string
	items = append(items, "")
	items = append(items, m.theme.Subtitle.Render("Select PHP version to add extensions:"))
	items = append(items, "")

	for i, version := range m.installedVersions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		label := fmt.Sprintf("PHP %s", version)
		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, label))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, label))
		}
		items = append(items, renderedItem)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", menu, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewExtensions renders the extensions selection view
func (m PHPExtensionsModel) viewExtensions() string {
	header := m.theme.Title.Render(fmt.Sprintf("PHP %s Extensions", m.selectedVersion))

	// Search bar
	searchBar := ""
	if m.searchMode {
		searchBar = m.theme.SelectedItem.Render(fmt.Sprintf("Search: %s_", m.searchQuery))
	} else if m.searchQuery != "" {
		searchBar = m.theme.InfoStyle.Render(fmt.Sprintf("Filter: %s (press / to search)", m.searchQuery))
	} else {
		searchBar = m.theme.DescriptionStyle.Render("Press / to search extensions")
	}

	// Count selected
	selectedCount := len(m.getSelectedExtensions())
	selectedInfo := m.theme.Label.Render(fmt.Sprintf("Selected: %d extensions", selectedCount))

	// Extensions list
	var items []string
	items = append(items, "")

	if len(m.filteredIndices) == 0 {
		items = append(items, m.theme.WarningStyle.Render("No extensions match your search"))
	} else {
		// Show visible range
		start := m.scrollOffset
		end := start + m.maxVisible
		if end > len(m.filteredIndices) {
			end = len(m.filteredIndices)
		}

		for i := start; i < end; i++ {
			extIdx := m.filteredIndices[i]
			ext := m.extensions[extIdx]

			cursor := "  "
			if i == m.cursor {
				cursor = m.theme.KeyStyle.Render("▶ ")
			}

			checkbox := "[ ]"
			if ext.Selected {
				checkbox = m.theme.SuccessStyle.Render("[✓]")
			}

			label := fmt.Sprintf("%s %s %s", cursor, checkbox, ext.Name)
			var renderedItem string
			if i == m.cursor {
				renderedItem = m.theme.SelectedItem.Render(label)
				items = append(items, renderedItem)
				items = append(items, "      "+m.theme.DescriptionStyle.Render(ext.Description))
			} else {
				if ext.Selected {
					renderedItem = m.theme.SuccessStyle.Render(label)
				} else {
					renderedItem = m.theme.MenuItem.Render(label)
				}
				items = append(items, renderedItem)
			}
		}

		// Scroll indicator
		if len(m.filteredIndices) > m.maxVisible {
			scrollInfo := fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.filteredIndices))
			items = append(items, "")
			items = append(items, m.theme.DescriptionStyle.Render(scrollInfo))
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Install button hint
	installHint := ""
	if selectedCount > 0 {
		installHint = m.theme.SuccessStyle.Render(fmt.Sprintf("Press Enter to install %d extension(s)", selectedCount))
	}

	// Messages
	messageSection := ""
	if m.err != nil {
		messageSection = m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Space: Toggle • /: Search • Enter: Install • Esc: Back")

	sections := []string{header, "", searchBar, selectedInfo, menu}
	if installHint != "" {
		sections = append(sections, "", installHint)
	}
	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewConfirm renders the confirmation view
func (m PHPExtensionsModel) viewConfirm() string {
	header := m.theme.Title.Render("Confirm Installation")

	selected := m.getSelectedExtensions()
	extList := strings.Join(selected, ", ")

	info := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Label.Render(fmt.Sprintf("PHP Version: %s", m.selectedVersion)),
		"",
		m.theme.Label.Render("Extensions to install:"),
		m.theme.InfoStyle.Render(extList),
	)

	var items []string
	items = append(items, "")
	options := []string{"Yes, install", "No, go back"}
	for i, opt := range options {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}
		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, opt))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, opt))
		}
		items = append(items, renderedItem)
	}
	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Confirm • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", info, menu, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
