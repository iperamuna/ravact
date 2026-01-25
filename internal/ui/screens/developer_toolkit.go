package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// ToolkitCategory represents a category of toolkit commands
type ToolkitCategory int

const (
	LaravelCategory ToolkitCategory = iota
	WordPressCategory
	PHPCategory
	SecurityCategory
)

// ToolkitCommand represents a command in the developer toolkit
type ToolkitCommand struct {
	Name        string
	Description string
	Command     string
	Category    ToolkitCategory
	NeedsPath   bool // If true, command needs a project path
}

// DeveloperToolkitModel represents the developer toolkit screen
type DeveloperToolkitModel struct {
	theme           *theme.Theme
	width           int
	height          int
	cursor          int
	category        ToolkitCategory
	commands        []ToolkitCommand
	filteredCmds    []ToolkitCommand
	copied          bool
	copiedTimer     int
	copiedCommand   string
	scrollOffset    int
	maxVisibleItems int
}

// NewDeveloperToolkitModel creates a new developer toolkit model
func NewDeveloperToolkitModel() DeveloperToolkitModel {
	t := theme.DefaultTheme()

	commands := []ToolkitCommand{
		// Laravel Commands
		{
			Name:        "Tail Laravel Log",
			Description: "Watch Laravel log file in real-time",
			Command:     "tail -f storage/logs/laravel.log",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Clear Laravel Log",
			Description: "Truncate Laravel log file to zero bytes",
			Command:     "truncate -s 0 storage/logs/laravel.log",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Find Large Log Files",
			Description: "Find log files larger than 100MB",
			Command:     "find storage/logs -type f -size +100M -exec ls -lh {} \\;",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Fix Storage Permissions",
			Description: "Set correct permissions for storage & bootstrap/cache",
			Command:     "chmod -R 775 storage bootstrap/cache && chown -R www-data:www-data storage bootstrap/cache",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Generate APP_KEY",
			Description: "Generate a new Laravel APP_KEY (base64)",
			Command:     "echo \"base64:$(openssl rand -base64 32)\"",
			Category:    LaravelCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check .env File",
			Description: "Display environment configuration (hides sensitive values)",
			Command:     "cat .env | grep -E '^(APP_|DB_HOST|DB_DATABASE|CACHE_|QUEUE_|MAIL_MAILER)' | sed 's/=.*/=***/'",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},
		{
			Name:        "List Scheduled Tasks",
			Description: "Show crontab entries for Laravel scheduler",
			Command:     "crontab -l | grep -E 'artisan|schedule'",
			Category:    LaravelCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check Queue Workers",
			Description: "List running queue worker processes",
			Command:     "ps aux | grep -E 'queue:work|queue:listen' | grep -v grep",
			Category:    LaravelCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Find Recently Modified Files",
			Description: "Find files modified in the last 24 hours",
			Command:     "find . -type f -mtime -1 -not -path './vendor/*' -not -path './node_modules/*' | head -50",
			Category:    LaravelCategory,
			NeedsPath:   true,
		},

		// WordPress Commands
		{
			Name:        "Fix wp-content Permissions",
			Description: "Set correct permissions for wp-content directory",
			Command:     "find wp-content -type d -exec chmod 755 {} \\; && find wp-content -type f -exec chmod 644 {} \\; && chown -R www-data:www-data wp-content",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Find Large Uploads",
			Description: "Find uploaded files larger than 10MB",
			Command:     "find wp-content/uploads -type f -size +10M -exec ls -lh {} \\;",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Clear Cache Files",
			Description: "Remove all files from wp-content/cache",
			Command:     "rm -rf wp-content/cache/* && echo 'Cache cleared'",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Generate WP Salts",
			Description: "Fetch fresh security salts from WordPress API",
			Command:     "curl -s https://api.wordpress.org/secret-key/1.1/salt/",
			Category:    WordPressCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check wp-config.php",
			Description: "Display database and debug settings",
			Command:     "grep -E \"^define\\('(DB_|WP_DEBUG)\" wp-config.php",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "List Plugins",
			Description: "List all installed plugins with details",
			Command:     "ls -la wp-content/plugins/",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "List Themes",
			Description: "List all installed themes",
			Command:     "ls -la wp-content/themes/",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Check .htaccess",
			Description: "Display .htaccess contents",
			Command:     "cat .htaccess",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Find Modified Core Files",
			Description: "Find WordPress core files modified in last 7 days",
			Command:     "find . -type f -mtime -7 -not -path './wp-content/*' -name '*.php' | head -30",
			Category:    WordPressCategory,
			NeedsPath:   true,
		},

		// PHP Commands
		{
			Name:        "Check PHP Version",
			Description: "Display installed PHP version",
			Command:     "php -v",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "List PHP Modules",
			Description: "List all installed PHP modules",
			Command:     "php -m",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check PHP Memory Limit",
			Description: "Display PHP memory limit setting",
			Command:     "php -i | grep -E '^memory_limit'",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check PHP Upload Limits",
			Description: "Display upload and post size limits",
			Command:     "php -i | grep -E '^(upload_max_filesize|post_max_size|max_execution_time)'",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Find php.ini Location",
			Description: "Show loaded PHP configuration file path",
			Command:     "php --ini | grep 'Loaded Configuration'",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check OPcache Status",
			Description: "Display OPcache configuration",
			Command:     "php -i | grep -E '^opcache\\.'",
			Category:    PHPCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Test PHP Syntax",
			Description: "Check PHP files for syntax errors",
			Command:     "find . -name '*.php' -not -path './vendor/*' -exec php -l {} \\; 2>&1 | grep -v 'No syntax errors'",
			Category:    PHPCategory,
			NeedsPath:   true,
		},
		{
			Name:        "List PHP-FPM Pools",
			Description: "Show PHP-FPM pool configurations",
			Command:     "ls -la /etc/php/*/fpm/pool.d/",
			Category:    PHPCategory,
			NeedsPath:   false,
		},

		// Security Commands
		{
			Name:        "Scan for Malware Patterns",
			Description: "Search for common malware signatures in PHP files",
			Command:     "grep -r -l -E 'eval\\(base64_decode|eval\\(gzinflate|eval\\(str_rot13' --include='*.php' . 2>/dev/null | head -20",
			Category:    SecurityCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Find World-Writable Files",
			Description: "List files with dangerous 777 permissions",
			Command:     "find . -type f -perm 0777 2>/dev/null | head -20",
			Category:    SecurityCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Find World-Writable Dirs",
			Description: "List directories with 777 permissions",
			Command:     "find . -type d -perm 0777 2>/dev/null | head -20",
			Category:    SecurityCategory,
			NeedsPath:   true,
		},
		{
			Name:        "Check for Suspicious Files",
			Description: "Find PHP files in upload directories",
			Command:     "find wp-content/uploads -name '*.php' -o -name '*.phtml' 2>/dev/null",
			Category:    SecurityCategory,
			NeedsPath:   true,
		},
		{
			Name:        "List Failed SSH Logins",
			Description: "Show recent failed SSH authentication attempts",
			Command:     "grep 'Failed password' /var/log/auth.log 2>/dev/null | tail -20 || journalctl -u ssh --no-pager | grep 'Failed' | tail -20",
			Category:    SecurityCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check Open Ports",
			Description: "List all listening ports and services",
			Command:     "ss -tulpn | grep LISTEN",
			Category:    SecurityCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Check SSL Certificate",
			Description: "Display SSL certificate expiry for localhost",
			Command:     "echo | openssl s_client -servername localhost -connect localhost:443 2>/dev/null | openssl x509 -noout -dates",
			Category:    SecurityCategory,
			NeedsPath:   false,
		},
		{
			Name:        "Find SUID Files",
			Description: "List files with SUID bit set (potential security risk)",
			Command:     "find / -perm -4000 -type f 2>/dev/null | head -20",
			Category:    SecurityCategory,
			NeedsPath:   false,
		},
	}

	m := DeveloperToolkitModel{
		theme:           t,
		commands:        commands,
		category:        LaravelCategory,
		cursor:          0,
		maxVisibleItems: 10,
	}

	m.filterByCategory()
	return m
}

// filterByCategory filters commands by the current category
func (m *DeveloperToolkitModel) filterByCategory() {
	m.filteredCmds = []ToolkitCommand{}
	for _, cmd := range m.commands {
		if cmd.Category == m.category {
			m.filteredCmds = append(m.filteredCmds, cmd)
		}
	}
	m.cursor = 0
	m.scrollOffset = 0
}

// getCategoryName returns the display name for a category
func (m DeveloperToolkitModel) getCategoryName(cat ToolkitCategory) string {
	switch cat {
	case LaravelCategory:
		return "Laravel"
	case WordPressCategory:
		return "WordPress"
	case PHPCategory:
		return "PHP"
	case SecurityCategory:
		return "Security"
	default:
		return "Unknown"
	}
}

func (m DeveloperToolkitModel) Init() tea.Cmd {
	return nil
}

func (m DeveloperToolkitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust max visible items based on height
		m.maxVisibleItems = (m.height - 20) / 3
		if m.maxVisibleItems < 5 {
			m.maxVisibleItems = 5
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: MainMenuScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll offset if needed
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.filteredCmds)-1 {
				m.cursor++
				// Adjust scroll offset if needed
				if m.cursor >= m.scrollOffset+m.maxVisibleItems {
					m.scrollOffset = m.cursor - m.maxVisibleItems + 1
				}
			}

		case "tab", "right", "l":
			// Switch to next category
			m.category = (m.category + 1) % 4
			m.filterByCategory()

		case "shift+tab", "left", "h":
			// Switch to previous category
			if m.category == 0 {
				m.category = SecurityCategory
			} else {
				m.category--
			}
			m.filterByCategory()

		case "c":
			// Copy command to clipboard
			if len(m.filteredCmds) > 0 && m.cursor < len(m.filteredCmds) {
				cmd := m.filteredCmds[m.cursor]
				err := clipboard.WriteAll(cmd.Command)
				m.copied = true
				if err == nil {
					m.copiedCommand = cmd.Name
				} else {
					m.copiedCommand = "(clipboard unavailable - install xclip)"
				}
				m.copiedTimer = 3
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}

		case "enter", " ":
			// Execute command (navigate to execution screen)
			if len(m.filteredCmds) > 0 && m.cursor < len(m.filteredCmds) {
				cmd := m.filteredCmds[m.cursor]
				return m, func() tea.Msg {
					return ExecuteToolkitCommandMsg{
						Command:     cmd.Command,
						Description: cmd.Name + ": " + cmd.Description,
						NeedsPath:   cmd.NeedsPath,
					}
				}
			}
		}

	case CopyTimerTickMsg:
		if m.copiedTimer > 0 {
			m.copiedTimer--
			if m.copiedTimer == 0 {
				m.copied = false
				m.copiedCommand = ""
			} else {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}
	}

	return m, nil
}

func (m DeveloperToolkitModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	// Header with host info
	hostInfo := system.GetHostInfo()
	headerText := "Developer Toolkit"
	if hostInfo != "" {
		headerText = fmt.Sprintf("Developer Toolkit  %s", m.theme.DescriptionStyle.Render(hostInfo))
	}
	header := m.theme.Title.Render(headerText)
	subtitle := m.theme.Subtitle.Render("Essential commands for Laravel & WordPress maintenance")

	// Category tabs
	var tabs []string
	categories := []ToolkitCategory{LaravelCategory, WordPressCategory, PHPCategory, SecurityCategory}
	for _, cat := range categories {
		name := m.getCategoryName(cat)
		if cat == m.category {
			tabs = append(tabs, m.theme.SelectedItem.Render(" "+name+" "))
		} else {
			tabs = append(tabs, m.theme.MenuItem.Render(" "+name+" "))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	// Commands list
	var cmdItems []string

	// Calculate visible range
	endIdx := m.scrollOffset + m.maxVisibleItems
	if endIdx > len(m.filteredCmds) {
		endIdx = len(m.filteredCmds)
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		cmd := m.filteredCmds[i]
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
		}

		var nameStyle, descStyle, cmdStyle string
		if i == m.cursor {
			nameStyle = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, cmd.Name))
			descStyle = m.theme.DescriptionStyle.Render("    " + cmd.Description)
			cmdStyle = m.theme.InfoStyle.Render("    $ " + truncateCommand(cmd.Command, m.width-20))
		} else {
			nameStyle = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, cmd.Name))
			descStyle = m.theme.DescriptionStyle.Render("    " + cmd.Description)
			cmdStyle = m.theme.Help.Render("    $ " + truncateCommand(cmd.Command, m.width-20))
		}

		cmdItems = append(cmdItems, nameStyle, descStyle, cmdStyle, "")
	}

	// Scroll indicator
	scrollInfo := ""
	if len(m.filteredCmds) > m.maxVisibleItems {
		scrollInfo = m.theme.DescriptionStyle.Render(fmt.Sprintf("  Showing %d-%d of %d commands", m.scrollOffset+1, endIdx, len(m.filteredCmds)))
	}

	commandsList := lipgloss.JoinVertical(lipgloss.Left, cmdItems...)

	// Messages
	var messages []string
	if m.copied {
		messages = append(messages, m.theme.CopiedStyle.Render(m.theme.Symbols.Copy+" Copied: "+m.copiedCommand))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	// Help
	help := m.theme.Help.Render(
		m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " +
			m.theme.Symbols.Bullet + " Tab/"+m.theme.Symbols.ArrowLeft+"/"+m.theme.Symbols.ArrowRight+": Category " +
			m.theme.Symbols.Bullet + " c: Copy " +
			m.theme.Symbols.Bullet + " Enter: Run " +
			m.theme.Symbols.Bullet + " Esc: Back")

	// Combine all sections
	sections := []string{
		header,
		subtitle,
		"",
		tabBar,
		"",
		commandsList,
	}

	if scrollInfo != "" {
		sections = append(sections, scrollInfo)
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add border and center
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// truncateCommand truncates a command string if it's too long
func truncateCommand(cmd string, maxLen int) string {
	if maxLen < 10 {
		maxLen = 40
	}
	// Replace newlines with spaces
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	if len(cmd) > maxLen {
		return cmd[:maxLen-3] + "..."
	}
	return cmd
}

// ExecuteToolkitCommandMsg is sent when a toolkit command should be executed
type ExecuteToolkitCommandMsg struct {
	Command     string
	Description string
	NeedsPath   bool
}
