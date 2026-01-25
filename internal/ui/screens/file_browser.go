package screens

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FileOperation represents a file operation type
type FileOperation int

const (
	OpNone FileOperation = iota
	OpCopy
	OpCut
	OpDelete
	OpRename
	OpNewFile
	OpNewDir
)

// FileBrowserMode represents the current mode of the file browser
type FileBrowserMode int

const (
	ModeNormal FileBrowserMode = iota
	ModeSearch
	ModeRename
	ModeNewFile
	ModeNewDir
	ModeConfirmDelete
	ModePreview
	ModeHelp
	ModeInfo
)

// FileEntry represents a file or directory entry
type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	// For symlinks
	IsSymlink  bool
	SymlinkDest string
}

// FileBrowserModel represents the file browser screen
type FileBrowserModel struct {
	theme           *theme.Theme
	width           int
	height          int
	
	// Current state
	currentPath     string
	entries         []FileEntry
	cursor          int
	scrollOffset    int
	maxVisibleItems int
	
	// Selection for operations
	selectedItems   map[string]bool
	clipboard       []FileEntry
	clipboardOp     FileOperation
	
	// Mode and input
	mode            FileBrowserMode
	inputBuffer     string
	inputCursor     int
	searchQuery     string
	filteredIndices []int
	
	// Preview
	previewContent  string
	previewScroll   int
	
	// History for back navigation
	history         []string
	historyIndex    int
	
	// Status messages
	statusMessage   string
	statusIsError   bool
	statusTimer     int
	
	// Settings
	showHidden      bool
	sortBy          string // "name", "size", "date"
	sortReverse     bool
	
	// Copied path indicator
	copied          bool
	copiedTimer     int
}

// NewFileBrowserModel creates a new file browser model
func NewFileBrowserModel() FileBrowserModel {
	// Determine the best starting directory
	startPath := determineStartPath()
	
	m := FileBrowserModel{
		theme:           theme.DefaultTheme(),
		currentPath:     startPath,
		selectedItems:   make(map[string]bool),
		history:         []string{startPath},
		historyIndex:    0,
		showHidden:      false,
		sortBy:          "name",
		maxVisibleItems: 20,
	}
	
	m.loadDirectory()
	return m
}

// determineStartPath finds the best starting directory for the file browser
func determineStartPath() string {
	// First, try the SUDO_USER's home directory (when running with sudo)
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" && sudoUser != "root" {
		sudoHome := filepath.Join("/home", sudoUser)
		if info, err := os.Stat(sudoHome); err == nil && info.IsDir() {
			if _, err := os.ReadDir(sudoHome); err == nil {
				return sudoHome
			}
		}
	}
	
	// Next, try the current user's home directory
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		// If home is /root, try to use /home instead if it has content
		if home == "/root" {
			// Check if /home has any user directories
			if entries, err := os.ReadDir("/home"); err == nil && len(entries) > 0 {
				// Use the first user's home directory
				for _, entry := range entries {
					if entry.IsDir() {
						userHome := filepath.Join("/home", entry.Name())
						if _, err := os.ReadDir(userHome); err == nil {
							return userHome
						}
					}
				}
			}
		}
		
		// Try to read the home directory
		if _, err := os.ReadDir(home); err == nil {
			return home
		}
	}
	
	// Try /home as fallback
	if info, err := os.Stat("/home"); err == nil && info.IsDir() {
		return "/home"
	}
	
	// Last resort: root directory
	return "/"
}

// NewFileBrowserModelWithPath creates a file browser starting at a specific path
func NewFileBrowserModelWithPath(path string) FileBrowserModel {
	m := NewFileBrowserModel()
	if path != "" {
		m.currentPath = path
		m.history = []string{path}
		m.loadDirectory()
	}
	return m
}

// loadDirectory loads the contents of the current directory
func (m *FileBrowserModel) loadDirectory() {
	m.entries = []FileEntry{}
	m.filteredIndices = []int{}
	
	dirEntries, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.setStatus(fmt.Sprintf("Error reading directory: %v", err), true)
		return
	}
	
	for _, entry := range dirEntries {
		name := entry.Name()
		
		// Skip hidden files if not showing them
		if !m.showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		
		fullPath := filepath.Join(m.currentPath, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		fe := FileEntry{
			Name:    name,
			Path:    fullPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
		}
		
		// Check for symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			fe.IsSymlink = true
			if dest, err := os.Readlink(fullPath); err == nil {
				fe.SymlinkDest = dest
			}
		}
		
		m.entries = append(m.entries, fe)
	}
	
	m.sortEntries()
	m.applyFilter()
	
	// Reset cursor if out of bounds
	if m.cursor >= len(m.getVisibleEntries()) {
		m.cursor = len(m.getVisibleEntries()) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.scrollOffset = 0
}

// sortEntries sorts the entries based on current sort settings
func (m *FileBrowserModel) sortEntries() {
	// Always put directories first
	sort.SliceStable(m.entries, func(i, j int) bool {
		// Directories first
		if m.entries[i].IsDir && !m.entries[j].IsDir {
			return true
		}
		if !m.entries[i].IsDir && m.entries[j].IsDir {
			return false
		}
		
		var result bool
		switch m.sortBy {
		case "size":
			result = m.entries[i].Size < m.entries[j].Size
		case "date":
			result = m.entries[i].ModTime.Before(m.entries[j].ModTime)
		default: // name
			result = strings.ToLower(m.entries[i].Name) < strings.ToLower(m.entries[j].Name)
		}
		
		if m.sortReverse {
			return !result
		}
		return result
	})
}

// applyFilter applies the search filter to entries
func (m *FileBrowserModel) applyFilter() {
	m.filteredIndices = []int{}
	query := strings.ToLower(m.searchQuery)
	
	for i, entry := range m.entries {
		if query == "" || strings.Contains(strings.ToLower(entry.Name), query) {
			m.filteredIndices = append(m.filteredIndices, i)
		}
	}
}

// getVisibleEntries returns the filtered entries
func (m *FileBrowserModel) getVisibleEntries() []FileEntry {
	if len(m.filteredIndices) == 0 && m.searchQuery == "" {
		return m.entries
	}
	
	result := make([]FileEntry, len(m.filteredIndices))
	for i, idx := range m.filteredIndices {
		result[i] = m.entries[idx]
	}
	return result
}

// getCurrentEntry returns the currently selected entry
func (m *FileBrowserModel) getCurrentEntry() *FileEntry {
	entries := m.getVisibleEntries()
	if m.cursor >= 0 && m.cursor < len(entries) {
		return &entries[m.cursor]
	}
	return nil
}

// navigateTo changes to a new directory
func (m *FileBrowserModel) navigateTo(path string) {
	// Verify path exists and is a directory
	info, err := os.Stat(path)
	if err != nil {
		m.setStatus(fmt.Sprintf("Cannot access: %v", err), true)
		return
	}
	if !info.IsDir() {
		m.setStatus("Not a directory", true)
		return
	}
	
	m.currentPath = path
	m.cursor = 0
	m.scrollOffset = 0
	m.searchQuery = ""
	m.loadDirectory()
	
	// Add to history
	if m.historyIndex < len(m.history)-1 {
		m.history = m.history[:m.historyIndex+1]
	}
	m.history = append(m.history, path)
	m.historyIndex = len(m.history) - 1
}

// goBack navigates to the parent directory
func (m *FileBrowserModel) goBack() {
	parent := filepath.Dir(m.currentPath)
	if parent != m.currentPath {
		m.navigateTo(parent)
	}
}

// goHistoryBack goes back in navigation history
func (m *FileBrowserModel) goHistoryBack() {
	if m.historyIndex > 0 {
		m.historyIndex--
		m.currentPath = m.history[m.historyIndex]
		m.loadDirectory()
	}
}

// goHistoryForward goes forward in navigation history
func (m *FileBrowserModel) goHistoryForward() {
	if m.historyIndex < len(m.history)-1 {
		m.historyIndex++
		m.currentPath = m.history[m.historyIndex]
		m.loadDirectory()
	}
}

// setStatus sets a status message
func (m *FileBrowserModel) setStatus(msg string, isError bool) {
	m.statusMessage = msg
	m.statusIsError = isError
	m.statusTimer = 3
}

// toggleSelection toggles selection of the current item
func (m *FileBrowserModel) toggleSelection() {
	entry := m.getCurrentEntry()
	if entry == nil {
		return
	}
	
	if m.selectedItems[entry.Path] {
		delete(m.selectedItems, entry.Path)
	} else {
		m.selectedItems[entry.Path] = true
	}
}

// selectAll selects all visible items
func (m *FileBrowserModel) selectAll() {
	entries := m.getVisibleEntries()
	for _, entry := range entries {
		m.selectedItems[entry.Path] = true
	}
}

// clearSelection clears all selections
func (m *FileBrowserModel) clearSelection() {
	m.selectedItems = make(map[string]bool)
}

// getSelectedEntries returns all selected entries
func (m *FileBrowserModel) getSelectedEntries() []FileEntry {
	var selected []FileEntry
	for _, entry := range m.entries {
		if m.selectedItems[entry.Path] {
			selected = append(selected, entry)
		}
	}
	return selected
}

// copyToClipboard copies selected items to internal clipboard
func (m *FileBrowserModel) copyToClipboard() {
	selected := m.getSelectedEntries()
	if len(selected) == 0 {
		if entry := m.getCurrentEntry(); entry != nil {
			selected = []FileEntry{*entry}
		}
	}
	
	if len(selected) > 0 {
		m.clipboard = selected
		m.clipboardOp = OpCopy
		m.setStatus(fmt.Sprintf("Copied %d item(s)", len(selected)), false)
	}
}

// cutToClipboard cuts selected items to internal clipboard
func (m *FileBrowserModel) cutToClipboard() {
	selected := m.getSelectedEntries()
	if len(selected) == 0 {
		if entry := m.getCurrentEntry(); entry != nil {
			selected = []FileEntry{*entry}
		}
	}
	
	if len(selected) > 0 {
		m.clipboard = selected
		m.clipboardOp = OpCut
		m.setStatus(fmt.Sprintf("Cut %d item(s)", len(selected)), false)
	}
}

// paste performs paste operation
func (m *FileBrowserModel) paste() error {
	if len(m.clipboard) == 0 {
		return fmt.Errorf("clipboard is empty")
	}
	
	for _, entry := range m.clipboard {
		destPath := filepath.Join(m.currentPath, entry.Name)
		
		// Check if destination exists
		if _, err := os.Stat(destPath); err == nil {
			// Add suffix to avoid overwrite
			base := strings.TrimSuffix(entry.Name, filepath.Ext(entry.Name))
			ext := filepath.Ext(entry.Name)
			destPath = filepath.Join(m.currentPath, fmt.Sprintf("%s_copy%s", base, ext))
		}
		
		if m.clipboardOp == OpCopy {
			if err := m.copyFile(entry.Path, destPath); err != nil {
				return err
			}
		} else if m.clipboardOp == OpCut {
			if err := os.Rename(entry.Path, destPath); err != nil {
				return err
			}
		}
	}
	
	if m.clipboardOp == OpCut {
		m.clipboard = nil
		m.clipboardOp = OpNone
	}
	
	m.loadDirectory()
	m.setStatus("Paste completed", false)
	return nil
}

// copyFile copies a file or directory
func (m *FileBrowserModel) copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if srcInfo.IsDir() {
		return m.copyDir(src, dst)
	}
	
	// Copy file
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, input, srcInfo.Mode())
}

// copyDir recursively copies a directory
func (m *FileBrowserModel) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if err := m.copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	
	return nil
}

// deleteSelected deletes selected items
func (m *FileBrowserModel) deleteSelected() error {
	selected := m.getSelectedEntries()
	if len(selected) == 0 {
		if entry := m.getCurrentEntry(); entry != nil {
			selected = []FileEntry{*entry}
		}
	}
	
	for _, entry := range selected {
		var err error
		if entry.IsDir {
			err = os.RemoveAll(entry.Path)
		} else {
			err = os.Remove(entry.Path)
		}
		if err != nil {
			return err
		}
	}
	
	m.clearSelection()
	m.loadDirectory()
	m.setStatus(fmt.Sprintf("Deleted %d item(s)", len(selected)), false)
	return nil
}

// createFile creates a new file
func (m *FileBrowserModel) createFile(name string) error {
	path := filepath.Join(m.currentPath, name)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Close()
	m.loadDirectory()
	m.setStatus(fmt.Sprintf("Created file: %s", name), false)
	return nil
}

// createDir creates a new directory
func (m *FileBrowserModel) createDir(name string) error {
	path := filepath.Join(m.currentPath, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	m.loadDirectory()
	m.setStatus(fmt.Sprintf("Created directory: %s", name), false)
	return nil
}

// renameEntry renames the current entry
func (m *FileBrowserModel) renameEntry(newName string) error {
	entry := m.getCurrentEntry()
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}
	
	newPath := filepath.Join(m.currentPath, newName)
	if err := os.Rename(entry.Path, newPath); err != nil {
		return err
	}
	
	m.loadDirectory()
	m.setStatus(fmt.Sprintf("Renamed to: %s", newName), false)
	return nil
}

// openFile opens a file with the system default application
func (m *FileBrowserModel) openFile(entry *FileEntry) error {
	var cmd *exec.Cmd
	
	switch {
	case isLinux():
		cmd = exec.Command("xdg-open", entry.Path)
	case isDarwin():
		cmd = exec.Command("open", entry.Path)
	default:
		return fmt.Errorf("unsupported platform")
	}
	
	return cmd.Start()
}

// loadPreview loads a preview of the current file
func (m *FileBrowserModel) loadPreview() {
	entry := m.getCurrentEntry()
	if entry == nil || entry.IsDir {
		m.previewContent = ""
		return
	}
	
	// Check file size - don't preview large files
	if entry.Size > 1024*1024 { // 1MB limit
		m.previewContent = "[File too large to preview]"
		return
	}
	
	content, err := os.ReadFile(entry.Path)
	if err != nil {
		m.previewContent = fmt.Sprintf("[Error reading file: %v]", err)
		return
	}
	
	// Check if binary
	if isBinary(content) {
		m.previewContent = "[Binary file]"
		return
	}
	
	m.previewContent = string(content)
	m.previewScroll = 0
}

// Helper functions
func isLinux() bool {
	return os.Getenv("XDG_CURRENT_DESKTOP") != "" || fileExists("/etc/os-release")
}

func isDarwin() bool {
	return fileExists("/System/Library/CoreServices/SystemVersion.plist")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isBinary(data []byte) bool {
	for _, b := range data[:min(512, len(data))] {
		if b == 0 {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// calculateDirSize calculates the total size of a directory (non-recursive for performance)
func calculateDirSize(path string) int64 {
	var totalSize int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			if info, err := entry.Info(); err == nil {
				totalSize += info.Size()
			}
		}
	}
	return totalSize
}

// formatSize formats a file size in human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() {
		if t.YearDay() == now.YearDay() {
			return t.Format("15:04")
		}
		return t.Format("Jan 02 15:04")
	}
	return t.Format("Jan 02 2006")
}

// Init initializes the file browser
func (m FileBrowserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the file browser
func (m FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.maxVisibleItems = (m.height - 12) / 1
		if m.maxVisibleItems < 5 {
			m.maxVisibleItems = 5
		}
		return m, nil

	case CopyTimerTickMsg:
		if m.copiedTimer > 0 {
			m.copiedTimer--
			if m.copiedTimer == 0 {
				m.copied = false
			} else {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}
		if m.statusTimer > 0 {
			m.statusTimer--
			if m.statusTimer == 0 {
				m.statusMessage = ""
			} else {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}

	case tea.KeyMsg:
		// Handle different modes
		switch m.mode {
		case ModeSearch:
			return m.handleSearchInput(msg)
		case ModeRename:
			return m.handleRenameInput(msg)
		case ModeNewFile:
			return m.handleNewFileInput(msg)
		case ModeNewDir:
			return m.handleNewDirInput(msg)
		case ModeConfirmDelete:
			return m.handleDeleteConfirm(msg)
		case ModePreview:
			return m.handlePreviewMode(msg)
		case ModeHelp:
			return m.handleHelpMode(msg)
		case ModeInfo:
			return m.handleInfoMode(msg)
		default:
			return m.handleNormalMode(msg)
		}
	}

	return m, nil
}

// handleNormalMode handles key input in normal mode
func (m FileBrowserModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	entries := m.getVisibleEntries()
	
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc":
		if m.searchQuery != "" {
			m.searchQuery = ""
			m.applyFilter()
			m.cursor = 0
			return m, nil
		}
		return m, func() tea.Msg {
			return NavigateMsg{Screen: MainMenuScreen}
		}

	// Navigation
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}

	case "down", "j":
		if m.cursor < len(entries)-1 {
			m.cursor++
			if m.cursor >= m.scrollOffset+m.maxVisibleItems {
				m.scrollOffset = m.cursor - m.maxVisibleItems + 1
			}
		}

	case "pgup", "ctrl+u":
		m.cursor -= m.maxVisibleItems
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.scrollOffset = m.cursor

	case "pgdown", "ctrl+d":
		m.cursor += m.maxVisibleItems
		if m.cursor >= len(entries) {
			m.cursor = len(entries) - 1
		}
		if m.cursor >= m.scrollOffset+m.maxVisibleItems {
			m.scrollOffset = m.cursor - m.maxVisibleItems + 1
		}

	case "home", "g":
		m.cursor = 0
		m.scrollOffset = 0

	case "end", "G":
		m.cursor = len(entries) - 1
		if m.cursor >= m.maxVisibleItems {
			m.scrollOffset = m.cursor - m.maxVisibleItems + 1
		}

	case "enter", "l", "right":
		entry := m.getCurrentEntry()
		if entry != nil {
			if entry.IsDir {
				m.navigateTo(entry.Path)
			} else {
				// Open file preview
				m.loadPreview()
				m.mode = ModePreview
			}
		}

	case "backspace", "h", "left":
		m.goBack()

	case "-":
		m.goHistoryBack()

	case "=", "+":
		m.goHistoryForward()

	// Selection
	case " ":
		m.toggleSelection()
		// Move to next item after selection
		if m.cursor < len(entries)-1 {
			m.cursor++
		}

	case "a":
		m.selectAll()
		m.setStatus("Selected all items", false)

	case "A":
		m.clearSelection()
		m.setStatus("Cleared selection", false)

	// File operations
	case "y":
		m.copyToClipboard()

	case "x":
		m.cutToClipboard()

	case "p":
		if err := m.paste(); err != nil {
			m.setStatus(fmt.Sprintf("Paste failed: %v", err), true)
		}

	case "d":
		if m.getCurrentEntry() != nil || len(m.selectedItems) > 0 {
			m.mode = ModeConfirmDelete
		}

	case "r":
		if m.getCurrentEntry() != nil {
			m.mode = ModeRename
			m.inputBuffer = m.getCurrentEntry().Name
			m.inputCursor = len(m.inputBuffer)
		}

	case "n":
		m.mode = ModeNewFile
		m.inputBuffer = ""
		m.inputCursor = 0

	case "N":
		m.mode = ModeNewDir
		m.inputBuffer = ""
		m.inputCursor = 0

	// Search and filter
	case "/":
		m.mode = ModeSearch
		m.inputBuffer = m.searchQuery
		m.inputCursor = len(m.inputBuffer)

	// View options
	case ".":
		m.showHidden = !m.showHidden
		m.loadDirectory()
		if m.showHidden {
			m.setStatus("Showing hidden files", false)
		} else {
			m.setStatus("Hiding hidden files", false)
		}

	case "s":
		// Cycle sort options
		switch m.sortBy {
		case "name":
			m.sortBy = "size"
			m.setStatus("Sort by: Size", false)
		case "size":
			m.sortBy = "date"
			m.setStatus("Sort by: Date", false)
		default:
			m.sortBy = "name"
			m.setStatus("Sort by: Name", false)
		}
		m.sortEntries()
		m.applyFilter()

	case "S":
		m.sortReverse = !m.sortReverse
		m.sortEntries()
		m.applyFilter()
		if m.sortReverse {
			m.setStatus("Sort: Reversed", false)
		} else {
			m.setStatus("Sort: Normal", false)
		}

	// Refresh
	case "R", "ctrl+r":
		m.loadDirectory()
		m.setStatus("Refreshed", false)

	// Copy path to system clipboard
	case "c":
		entry := m.getCurrentEntry()
		if entry != nil {
			err := clipboard.WriteAll(entry.Path)
			m.copied = true
			m.copiedTimer = 3
			if err == nil {
				m.setStatus("Path copied: "+entry.Path, false)
			} else {
				m.setStatus("Path: "+entry.Path+" (clipboard unavailable)", false)
			}
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return CopyTimerTickMsg{}
			})
		}

	// Open with system default
	case "o":
		entry := m.getCurrentEntry()
		if entry != nil {
			if err := m.openFile(entry); err != nil {
				m.setStatus(fmt.Sprintf("Failed to open: %v", err), true)
			} else {
				m.setStatus("Opening...", false)
			}
		}

	// Go to home directory
	case "~":
		if home, err := os.UserHomeDir(); err == nil {
			m.navigateTo(home)
		}

	// Go to root
	case "`":
		m.navigateTo("/")

	// Help screen
	case "?":
		m.mode = ModeHelp

	// Info/permissions screen
	case "i":
		if m.getCurrentEntry() != nil {
			m.mode = ModeInfo
		}
	}

	return m, nil
}

// handleHelpMode handles help screen input
func (m FileBrowserModel) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?", "enter", " ":
		m.mode = ModeNormal
	}
	return m, nil
}

// handleInfoMode handles info/permissions screen input
func (m FileBrowserModel) handleInfoMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "i", "enter", " ":
		m.mode = ModeNormal
	}
	return m, nil
}

// handleSearchInput handles input in search mode
func (m FileBrowserModel) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchQuery = m.inputBuffer
		m.applyFilter()
		m.cursor = 0
		m.scrollOffset = 0
		m.mode = ModeNormal

	case "esc":
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			// Live search
			m.searchQuery = m.inputBuffer
			m.applyFilter()
			m.cursor = 0
		}

	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
			// Live search
			m.searchQuery = m.inputBuffer
			m.applyFilter()
			m.cursor = 0
		}
	}
	return m, nil
}

// handleRenameInput handles input in rename mode
func (m FileBrowserModel) handleRenameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.inputBuffer != "" {
			if err := m.renameEntry(m.inputBuffer); err != nil {
				m.setStatus(fmt.Sprintf("Rename failed: %v", err), true)
			}
		}
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "esc":
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

// handleNewFileInput handles input in new file mode
func (m FileBrowserModel) handleNewFileInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.inputBuffer != "" {
			if err := m.createFile(m.inputBuffer); err != nil {
				m.setStatus(fmt.Sprintf("Create file failed: %v", err), true)
			}
		}
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "esc":
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

// handleNewDirInput handles input in new directory mode
func (m FileBrowserModel) handleNewDirInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.inputBuffer != "" {
			if err := m.createDir(m.inputBuffer); err != nil {
				m.setStatus(fmt.Sprintf("Create directory failed: %v", err), true)
			}
		}
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "esc":
		m.mode = ModeNormal
		m.inputBuffer = ""

	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
	}
	return m, nil
}

// handleDeleteConfirm handles delete confirmation
func (m FileBrowserModel) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if err := m.deleteSelected(); err != nil {
			m.setStatus(fmt.Sprintf("Delete failed: %v", err), true)
		}
		m.mode = ModeNormal

	case "n", "N", "esc":
		m.mode = ModeNormal
		m.setStatus("Delete cancelled", false)
	}
	return m, nil
}

// handlePreviewMode handles preview mode input
func (m FileBrowserModel) handlePreviewMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Split(m.previewContent, "\n")
	maxScroll := len(lines) - (m.height - 10)
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch msg.String() {
	case "esc", "q", "backspace":
		m.mode = ModeNormal
		m.previewContent = ""

	case "up", "k":
		if m.previewScroll > 0 {
			m.previewScroll--
		}

	case "down", "j":
		if m.previewScroll < maxScroll {
			m.previewScroll++
		}

	case "pgup", "ctrl+u":
		m.previewScroll -= m.height / 2
		if m.previewScroll < 0 {
			m.previewScroll = 0
		}

	case "pgdown", "ctrl+d":
		m.previewScroll += m.height / 2
		if m.previewScroll > maxScroll {
			m.previewScroll = maxScroll
		}

	case "home", "g":
		m.previewScroll = 0

	case "end", "G":
		m.previewScroll = maxScroll

	case "c":
		// Copy file content
		err := clipboard.WriteAll(m.previewContent)
		if err == nil {
			m.setStatus("Content copied to clipboard", false)
		} else {
			m.setStatus("Clipboard unavailable - install xclip", false)
		}

	case "o":
		// Open with system editor
		entry := m.getCurrentEntry()
		if entry != nil {
			m.openFile(entry)
		}
	}
	return m, nil
}

// View renders the file browser
func (m FileBrowserModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Handle special modes
	if m.mode == ModePreview {
		return m.renderPreview()
	}
	if m.mode == ModeHelp {
		return m.renderHelp()
	}
	if m.mode == ModeInfo {
		return m.renderInfo()
	}

	// Header with current path
	// Header with host info
	hostInfo := system.GetHostInfo()
	headerText := "File Browser"
	if hostInfo != "" {
		headerText = fmt.Sprintf("File Browser  %s", m.theme.DescriptionStyle.Render(hostInfo))
	}
	header := m.theme.Title.Render(headerText)
	
	// Path bar
	pathStyle := m.theme.InfoStyle.Copy().Bold(true)
	pathBar := pathStyle.Render(m.theme.Symbols.ArrowRight + " " + m.currentPath)

	// Search bar (if searching)
	searchBar := ""
	if m.searchQuery != "" || m.mode == ModeSearch {
		searchIcon := m.theme.Symbols.Info
		if m.mode == ModeSearch {
			searchBar = m.theme.WarningStyle.Render(searchIcon + " Search: " + m.inputBuffer + "_")
		} else {
			searchBar = m.theme.DescriptionStyle.Render(searchIcon + " Filter: " + m.searchQuery)
		}
	}

	// Input bar for other modes
	inputBar := ""
	switch m.mode {
	case ModeRename:
		inputBar = m.theme.WarningStyle.Render("Rename: " + m.inputBuffer + "_")
	case ModeNewFile:
		inputBar = m.theme.WarningStyle.Render("New file: " + m.inputBuffer + "_")
	case ModeNewDir:
		inputBar = m.theme.WarningStyle.Render("New directory: " + m.inputBuffer + "_")
	case ModeConfirmDelete:
		count := len(m.getSelectedEntries())
		if count == 0 {
			count = 1
		}
		inputBar = m.theme.ErrorStyle.Render(fmt.Sprintf("Delete %d item(s)? (y/n)", count))
	}

	// Padding values for the file browser
	paddingH := 10
	contentWidth := m.width - (paddingH * 2) - 10 // Account for padding and border
	
	// File list
	entries := m.getVisibleEntries()
	var fileList []string

	// Calculate visible range
	endIdx := m.scrollOffset + m.maxVisibleItems
	if endIdx > len(entries) {
		endIdx = len(entries)
	}

	// Column widths - use content width instead of full width
	nameWidth := contentWidth - 50  // Reserve space for size, date, permissions
	if nameWidth < 20 {
		nameWidth = 20
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		entry := entries[i]
		
		// Cursor and selection indicator
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
		}
		
		// Selection checkbox
		checkbox := m.theme.Symbols.Box + " "
		if m.selectedItems[entry.Path] {
			checkbox = m.theme.SuccessStyle.Render(m.theme.Symbols.BoxChecked + " ")
		}
		
		// File icon and name
		icon := m.getFileIcon(entry)
		name := entry.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-3] + "..."
		}
		
		// Size (show size for both files and directories)
		sizeStr := ""
		if !entry.IsDir {
			sizeStr = formatSize(entry.Size)
		} else {
			// Calculate total size of directory contents
			dirSize := calculateDirSize(entry.Path)
			sizeStr = formatSize(dirSize)
		}
		sizeStr = fmt.Sprintf("%8s", sizeStr)
		
		// Modified time
		timeStr := formatTime(entry.ModTime)
		
		// Permissions
		permStr := entry.Mode.String()[:10]
		
		// Build the line
		var line string
		if entry.IsDir {
			dirStyle := m.theme.InfoStyle.Copy().Bold(true)
			if i == m.cursor {
				line = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s%s %-*s %s %s %s",
					cursor, checkbox, icon, nameWidth, name, sizeStr, timeStr, permStr))
			} else {
				line = fmt.Sprintf("%s%s%s %-*s %s %s %s",
					cursor, checkbox, dirStyle.Render(icon+" "+name), nameWidth-len(name)-2, "", 
					m.theme.DescriptionStyle.Render(sizeStr),
					m.theme.DescriptionStyle.Render(timeStr),
					m.theme.DescriptionStyle.Render(permStr))
			}
		} else {
			if i == m.cursor {
				line = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s%s %-*s %s %s %s",
					cursor, checkbox, icon, nameWidth, name, sizeStr, timeStr, permStr))
			} else {
				line = fmt.Sprintf("%s%s%s %-*s %s %s %s",
					cursor, checkbox, m.theme.MenuItem.Render(icon+" "+name), nameWidth-len(name)-2, "",
					m.theme.DescriptionStyle.Render(sizeStr),
					m.theme.DescriptionStyle.Render(timeStr),
					m.theme.DescriptionStyle.Render(permStr))
			}
		}
		
		fileList = append(fileList, line)
	}

	// Empty directory message
	if len(entries) == 0 {
		fileList = append(fileList, m.theme.DescriptionStyle.Render("  (empty directory)"))
	}

	fileListStr := lipgloss.JoinVertical(lipgloss.Left, fileList...)

	// Status bar
	statusBar := m.renderStatusBar(entries)

	// Clipboard indicator
	clipboardInfo := ""
	if len(m.clipboard) > 0 {
		opStr := "copied"
		if m.clipboardOp == OpCut {
			opStr = "cut"
		}
		clipboardInfo = m.theme.InfoStyle.Render(fmt.Sprintf("%s %d item(s) %s", m.theme.Symbols.Copy, len(m.clipboard), opStr))
	}

	// Status message
	statusMsg := ""
	if m.statusMessage != "" {
		if m.statusIsError {
			statusMsg = m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark + " " + m.statusMessage)
		} else {
			statusMsg = m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark + " " + m.statusMessage)
		}
	}

	// Help bar
	help := m.renderHelpBar()

	// Combine all sections
	sections := []string{header, "", pathBar}
	
	if searchBar != "" {
		sections = append(sections, searchBar)
	}
	if inputBar != "" {
		sections = append(sections, inputBar)
	}
	
	sections = append(sections, "", fileListStr, "", statusBar)
	
	if clipboardInfo != "" {
		sections = append(sections, clipboardInfo)
	}
	if statusMsg != "" {
		sections = append(sections, statusMsg)
	}
	
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Add border
	bordered := m.theme.BorderStyle.Render(content)

	// Use same paddingH as defined above for content width
	paddingV := 2

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().
			Padding(paddingV, paddingH).
			Render(bordered),
	)
}

// renderPreview renders the file preview mode
func (m FileBrowserModel) renderPreview() string {
	entry := m.getCurrentEntry()
	if entry == nil {
		return "No file selected"
	}

	header := m.theme.Title.Render("File Preview: " + entry.Name)
	
	// File info
	info := m.theme.DescriptionStyle.Render(fmt.Sprintf("Size: %s | Modified: %s | Mode: %s",
		formatSize(entry.Size), formatTime(entry.ModTime), entry.Mode.String()))

	// Preview content with scrolling
	lines := strings.Split(m.previewContent, "\n")
	visibleLines := m.height - 12
	if visibleLines < 5 {
		visibleLines = 5
	}

	endLine := m.previewScroll + visibleLines
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var previewLines []string
	for i := m.previewScroll; i < endLine; i++ {
		line := lines[i]
		// Truncate long lines
		if len(line) > m.width-10 {
			line = line[:m.width-13] + "..."
		}
		// Add line numbers
		lineNum := m.theme.DescriptionStyle.Render(fmt.Sprintf("%4d ", i+1))
		previewLines = append(previewLines, lineNum+m.theme.MenuItem.Render(line))
	}

	previewContent := lipgloss.JoinVertical(lipgloss.Left, previewLines...)

	// Scroll position
	scrollInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Lines %d-%d of %d", 
		m.previewScroll+1, endLine, len(lines)))

	// Help
	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Scroll " +
		m.theme.Symbols.Bullet + " c: Copy content " +
		m.theme.Symbols.Bullet + " o: Open external " +
		m.theme.Symbols.Bullet + " Esc: Back")

	sections := []string{header, info, "", previewContent, "", scrollInfo, "", help}
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderStatusBar renders the bottom status bar
func (m FileBrowserModel) renderStatusBar(entries []FileEntry) string {
	// Count stats
	totalItems := len(entries)
	selectedCount := len(m.selectedItems)
	
	// Calculate total size of visible items
	var totalSize int64
	for _, entry := range entries {
		if !entry.IsDir {
			totalSize += entry.Size
		}
	}

	parts := []string{
		fmt.Sprintf("%d items", totalItems),
		formatSize(totalSize),
	}
	
	if selectedCount > 0 {
		parts = append(parts, fmt.Sprintf("%d selected", selectedCount))
	}
	
	// Sort indicator
	sortIndicator := "Name"
	switch m.sortBy {
	case "size":
		sortIndicator = "Size"
	case "date":
		sortIndicator = "Date"
	}
	if m.sortReverse {
		sortIndicator += m.theme.Symbols.ArrowDown
	} else {
		sortIndicator += m.theme.Symbols.ArrowUp
	}
	parts = append(parts, "Sort: "+sortIndicator)
	
	// Hidden files indicator
	if m.showHidden {
		parts = append(parts, "Hidden: On")
	}

	return m.theme.DescriptionStyle.Render(strings.Join(parts, " | "))
}

// renderHelpBar renders the help bar based on current mode
func (m FileBrowserModel) renderHelpBar() string {
	switch m.mode {
	case ModeSearch:
		return m.theme.Help.Render("Type to search " + m.theme.Symbols.Bullet + " Enter: Apply " + m.theme.Symbols.Bullet + " Esc: Cancel")
	case ModeRename, ModeNewFile, ModeNewDir:
		return m.theme.Help.Render("Type name " + m.theme.Symbols.Bullet + " Enter: Confirm " + m.theme.Symbols.Bullet + " Esc: Cancel")
	case ModeConfirmDelete:
		return m.theme.Help.Render("y: Confirm delete " + m.theme.Symbols.Bullet + " n/Esc: Cancel")
	default:
		return m.theme.Help.Render(
			m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " +
			m.theme.Symbols.Bullet + " Enter: Open " +
			m.theme.Symbols.Bullet + " Backspace: Up " +
			m.theme.Symbols.Bullet + " i: Info " +
			m.theme.Symbols.Bullet + " y/x/p: Copy/Cut/Paste " +
			m.theme.Symbols.Bullet + " ?: Help")
	}
}

// renderHelp renders the full help screen
func (m FileBrowserModel) renderHelp() string {
	header := m.theme.Title.Render("File Browser - Keyboard Shortcuts")

	// Define help sections
	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"↑/k, ↓/j", "Move cursor up/down"},
				{"Enter/l/→", "Open directory or preview file"},
				{"Backspace/h/←", "Go to parent directory"},
				{"PgUp/Ctrl+U", "Page up"},
				{"PgDown/Ctrl+D", "Page down"},
				{"Home/g", "Go to first item"},
				{"End/G", "Go to last item"},
				{"~", "Go to home directory"},
				{"`", "Go to root directory"},
				{"-", "Go back in history"},
				{"=/+", "Go forward in history"},
			},
		},
		{
			title: "Selection",
			keys: [][2]string{
				{"Space", "Toggle selection on current item"},
				{"a", "Select all items"},
				{"A", "Clear all selections"},
			},
		},
		{
			title: "File Operations",
			keys: [][2]string{
				{"y", "Copy selected items to clipboard"},
				{"x", "Cut selected items to clipboard"},
				{"p", "Paste from clipboard"},
				{"c", "Copy file path to system clipboard"},
				{"n", "Create new file"},
				{"N", "Create new directory"},
				{"r", "Rename current item"},
				{"d", "Delete selected items"},
				{"o", "Open with system default app"},
				{"i", "Show file info & permissions"},
			},
		},
		{
			title: "Search & View",
			keys: [][2]string{
				{"/", "Search/filter files"},
				{".", "Toggle hidden files"},
				{"s", "Cycle sort (Name → Size → Date)"},
				{"S", "Reverse sort order"},
				{"R/Ctrl+R", "Refresh directory"},
			},
		},
		{
			title: "Preview Mode",
			keys: [][2]string{
				{"↑/k, ↓/j", "Scroll up/down"},
				{"PgUp/PgDn", "Scroll page up/down"},
				{"c", "Copy file content"},
				{"o", "Open with external editor"},
				{"Esc/q", "Close preview"},
			},
		},
		{
			title: "General",
			keys: [][2]string{
				{"?", "Show/hide this help"},
				{"Esc", "Go back / Cancel"},
				{"q", "Quit to main menu"},
				{"Ctrl+C", "Quit application"},
			},
		},
	}

	var content []string
	content = append(content, header, "")

	// Render each section
	for _, section := range sections {
		// Section title
		sectionTitle := m.theme.CategoryStyle.Render(m.theme.Symbols.ArrowRight + " " + section.title)
		content = append(content, sectionTitle)

		// Key bindings
		for _, kv := range section.keys {
			key := m.theme.KeyStyle.Render(fmt.Sprintf("  %-16s", kv[0]))
			desc := m.theme.MenuItem.Render(kv[1])
			content = append(content, key+desc)
		}
		content = append(content, "")
	}

	// Footer
	footer := m.theme.Help.Render("Press Esc, ?, or Enter to close this help")
	content = append(content, footer)

	helpContent := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Add border
	bordered := m.theme.BorderStyle.Render(helpContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderInfo renders the file info/permissions popup
func (m FileBrowserModel) renderInfo() string {
	entry := m.getCurrentEntry()
	if entry == nil {
		return "No file selected"
	}

	header := m.theme.Title.Render("File Information")

	// Get file info using stat command for ownership
	var ownerInfo, groupInfo string
	cmd := exec.Command("stat", "-c", "%U:%G", entry.Path)
	if output, err := cmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), ":")
		if len(parts) == 2 {
			ownerInfo = parts[0]
			groupInfo = parts[1]
		}
	}
	if ownerInfo == "" {
		ownerInfo = "unknown"
		groupInfo = "unknown"
	}

	// Parse permissions into human readable format
	mode := entry.Mode
	permStr := mode.String()

	// Human readable permissions
	var permsReadable []string
	
	// Owner permissions
	ownerPerms := []string{}
	if mode&0400 != 0 {
		ownerPerms = append(ownerPerms, "read")
	}
	if mode&0200 != 0 {
		ownerPerms = append(ownerPerms, "write")
	}
	if mode&0100 != 0 {
		ownerPerms = append(ownerPerms, "execute")
	}
	if len(ownerPerms) > 0 {
		permsReadable = append(permsReadable, fmt.Sprintf("Owner (%s): %s", ownerInfo, strings.Join(ownerPerms, ", ")))
	} else {
		permsReadable = append(permsReadable, fmt.Sprintf("Owner (%s): none", ownerInfo))
	}

	// Group permissions
	groupPerms := []string{}
	if mode&0040 != 0 {
		groupPerms = append(groupPerms, "read")
	}
	if mode&0020 != 0 {
		groupPerms = append(groupPerms, "write")
	}
	if mode&0010 != 0 {
		groupPerms = append(groupPerms, "execute")
	}
	if len(groupPerms) > 0 {
		permsReadable = append(permsReadable, fmt.Sprintf("Group (%s): %s", groupInfo, strings.Join(groupPerms, ", ")))
	} else {
		permsReadable = append(permsReadable, fmt.Sprintf("Group (%s): none", groupInfo))
	}

	// Others permissions
	othersPerms := []string{}
	if mode&0004 != 0 {
		othersPerms = append(othersPerms, "read")
	}
	if mode&0002 != 0 {
		othersPerms = append(othersPerms, "write")
	}
	if mode&0001 != 0 {
		othersPerms = append(othersPerms, "execute")
	}
	if len(othersPerms) > 0 {
		permsReadable = append(permsReadable, fmt.Sprintf("Others: %s", strings.Join(othersPerms, ", ")))
	} else {
		permsReadable = append(permsReadable, "Others: none")
	}

	// Calculate size
	var sizeStr string
	if entry.IsDir {
		dirSize := calculateDirSize(entry.Path)
		// Count items
		items, _ := os.ReadDir(entry.Path)
		sizeStr = fmt.Sprintf("%s (%d items)", formatSize(dirSize), len(items))
	} else {
		sizeStr = formatSize(entry.Size)
	}

	// Build info content
	var content []string
	content = append(content, header, "")
	
	// File name
	content = append(content, m.theme.CategoryStyle.Render("Name"))
	content = append(content, "  "+m.theme.MenuItem.Render(entry.Name))
	content = append(content, "")

	// Full path
	content = append(content, m.theme.CategoryStyle.Render("Path"))
	content = append(content, "  "+m.theme.MenuItem.Render(entry.Path))
	content = append(content, "")

	// Type
	content = append(content, m.theme.CategoryStyle.Render("Type"))
	typeStr := "File"
	if entry.IsDir {
		typeStr = "Directory"
	}
	if entry.IsSymlink {
		typeStr = "Symbolic Link → " + entry.SymlinkDest
	}
	content = append(content, "  "+m.theme.MenuItem.Render(typeStr))
	content = append(content, "")

	// Size
	content = append(content, m.theme.CategoryStyle.Render("Size"))
	content = append(content, "  "+m.theme.MenuItem.Render(sizeStr))
	content = append(content, "")

	// Modified time
	content = append(content, m.theme.CategoryStyle.Render("Modified"))
	content = append(content, "  "+m.theme.MenuItem.Render(entry.ModTime.Format("Jan 02, 2006 15:04:05")))
	content = append(content, "")

	// Permissions
	content = append(content, m.theme.CategoryStyle.Render("Permissions"))
	content = append(content, "  "+m.theme.InfoStyle.Render(permStr))
	content = append(content, "")

	// Human readable permissions
	content = append(content, m.theme.CategoryStyle.Render("Access Rights"))
	for _, perm := range permsReadable {
		content = append(content, "  "+m.theme.MenuItem.Render(perm))
	}
	content = append(content, "")

	// Help
	help := m.theme.Help.Render("Press Esc, i, or Enter to close")
	content = append(content, help)

	infoContent := lipgloss.JoinVertical(lipgloss.Left, content...)
	bordered := m.theme.BorderStyle.Render(infoContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// getFileIcon returns an icon for the file type
func (m FileBrowserModel) getFileIcon(entry FileEntry) string {
	if entry.IsDir {
		if entry.IsSymlink {
			return m.theme.Symbols.ArrowRight
		}
		return m.theme.Symbols.ArrowDown
	}
	
	// Get extension
	ext := strings.ToLower(filepath.Ext(entry.Name))
	
	// Return appropriate icon based on extension
	switch ext {
	case ".go":
		return "Go"
	case ".js", ".ts", ".jsx", ".tsx":
		return "JS"
	case ".py":
		return "Py"
	case ".rb":
		return "Rb"
	case ".php":
		return "PHP"
	case ".html", ".htm":
		return "HTML"
	case ".css", ".scss", ".sass":
		return "CSS"
	case ".json":
		return "JSON"
	case ".xml":
		return "XML"
	case ".yaml", ".yml":
		return "YAML"
	case ".md", ".markdown":
		return "MD"
	case ".txt":
		return "TXT"
	case ".sh", ".bash", ".zsh":
		return "SH"
	case ".sql":
		return "SQL"
	case ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp":
		return "IMG"
	case ".mp3", ".wav", ".flac", ".ogg":
		return "AUD"
	case ".mp4", ".mkv", ".avi", ".mov":
		return "VID"
	case ".pdf":
		return "PDF"
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return "ZIP"
	case ".exe", ".bin", ".app":
		return "EXE"
	case ".env":
		return "ENV"
	case ".log":
		return "LOG"
	case ".lock":
		return "LCK"
	default:
		if entry.IsSymlink {
			return m.theme.Symbols.ArrowRight
		}
		return m.theme.Symbols.Bullet
	}
}
