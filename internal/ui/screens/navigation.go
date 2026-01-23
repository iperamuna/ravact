package screens

// ScreenType represents different screens in the application
type ScreenType int

const (
	SplashScreen ScreenType = iota
	MainMenuScreen
	SetupMenuScreen
	SetupActionScreen
	InstalledAppsScreen
	UserManagementScreen
	UserDetailsScreen
	AddUserScreen
	NginxConfigScreen
	ConfigMenuScreen
	QuickCommandsScreen
	ConfigEditorScreen
	SSLOptionsScreen
	SSLManualScreen
	EditorSelectionScreen
	RedisConfigScreen
	RedisPasswordScreen
	RedisPortScreen
	ExecutionScreen
)

// NavigateMsg is sent when navigating between screens
type NavigateMsg struct {
	Screen ScreenType
	Data   interface{} // Optional data to pass to the next screen
}

// BackMsg is sent when going back to the previous screen
type BackMsg struct{}

// QuitMsg is sent when quitting the application
type QuitMsg struct{}

// ExecutionStartMsg is sent when starting execution
type ExecutionStartMsg struct {
	Command     string
	Description string
}

// ExecutionCompleteMsg is sent when execution completes
type ExecutionCompleteMsg struct {
	Success bool
	Output  string
	Error   error
}
