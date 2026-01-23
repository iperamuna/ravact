package main

import (
	"embed"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/screens"
)

var Version = "0.1.0"

//go:embed assets
var embeddedAssets embed.FS

// Model represents the root application model
type Model struct {
	currentScreen  screens.ScreenType
	splash         screens.SplashModel
	mainMenu       screens.MainMenuModel
	setupMenu      screens.SetupMenuModel
	setupAction    screens.SetupActionModel
	installedApps  screens.InstalledAppsModel
	userManagement screens.UserManagementModel
	userDetails    screens.UserDetailsModel
	addUser        screens.AddUserModel
	configMenu     screens.ConfigMenuModel
	nginxConfig    screens.NginxConfigModel
	addSite        screens.AddSiteModel
	siteDetails    screens.SiteDetailsModel
	sslOptions     screens.SSLOptionsModel
	sslManual      screens.SSLManualModel
	editorSelection screens.EditorSelectionModel
	redisConfig    screens.RedisConfigModel
	redisPassword  screens.RedisPasswordModel
	redisPort      screens.RedisPortModel
	mysqlManagement screens.MySQLManagementModel
	mysqlPassword   screens.MySQLPasswordModel
	mysqlPort       screens.MySQLPortModel
	postgresqlManagement screens.PostgreSQLManagementModel
	postgresqlPassword   screens.PostgreSQLPasswordModel
	postgresqlPort       screens.PostgreSQLPortModel
	phpfpmManagement screens.PHPFPMManagementModel
	supervisorManagement screens.SupervisorManagementModel
	supervisorXMLRPCConfig screens.SupervisorXMLRPCConfigModel
	supervisorAddProgram screens.SupervisorAddProgramModel
	quickCommands  screens.QuickCommandsModel
	execution      screens.ExecutionModel
	configEditorActive string // "add_site" or "site_details"
	width          int
	height         int
	scriptsDir     string
	configsDir     string
}

// NewModel creates a new application model
func NewModel() Model {
	// No need to extract - we'll read directly from embedded FS
	// Removed info message - silent operation
	
	return Model{
		currentScreen:  screens.SplashScreen,
		splash:         screens.NewSplashModel(),
		mainMenu:       screens.NewMainMenuModel(Version),
		setupMenu:      screens.NewSetupMenuModel("assets/scripts"),
		installedApps:  screens.NewInstalledAppsModel("assets/scripts"),
		userManagement: screens.NewUserManagementModel(),
		nginxConfig:    screens.NewNginxConfigModel(),
		quickCommands:  screens.NewQuickCommandsModel(),
		scriptsDir:     "assets/scripts",
		configsDir:     "assets/configs",
	}
}

// GetEmbeddedFS returns the embedded assets filesystem
func GetEmbeddedFS() embed.FS {
	return embeddedAssets
}

// No extraction needed - scripts run directly from embedded FS

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all application messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate size to all screens
		m.splash.SetSize(msg.Width, msg.Height)
		// No need to return here, let it propagate to current screen

	case tea.KeyMsg:
		// Global quit keys
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case screens.NavigateMsg:
		m.currentScreen = msg.Screen
		
		// Handle screen-specific initialization with data
		if msg.Screen == screens.SetupActionScreen && msg.Data != nil {
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if script, ok := data["script"].(models.SetupScript); ok {
					status := models.StatusUnknown
					if s, ok := data["status"].(models.ServiceStatus); ok {
						status = s
					}
					m.setupAction = screens.NewSetupActionModel(script, status)
				}
			}
		}
		
		// Initialize screen-specific models that need async loading or data
		var initCmd tea.Cmd
		switch msg.Screen {
		case screens.UserManagementScreen:
			// Reinitialize user management on navigation
			m.userManagement = screens.NewUserManagementModel()
			initCmd = m.userManagement.Init()
		
		case screens.UserDetailsScreen:
			// Initialize user details with user data
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if user, ok := data["user"].(system.User); ok {
						m.userDetails = screens.NewUserDetailsModel(user)
					}
				}
			}
		
		case screens.AddUserScreen:
			// Initialize add user screen
			m.addUser = screens.NewAddUserModel()
		
		case screens.ConfigMenuScreen:
			// Initialize config menu screen
			m.configMenu = screens.NewConfigMenuModel()
		
		case screens.NginxConfigScreen:
			// Initialize Nginx config screen
			m.nginxConfig = screens.NewNginxConfigModel()
		
		case screens.SSLOptionsScreen:
			// Initialize SSL options screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if site, ok := data["site"].(system.NginxSite); ok {
						m.sslOptions = screens.NewSSLOptionsModel(site)
					}
				}
			}
		
		case screens.SSLManualScreen:
			// Initialize SSL manual screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if site, ok := data["site"].(system.NginxSite); ok {
						m.sslManual = screens.NewSSLManualModel(site)
					}
				}
			}
		
		case screens.EditorSelectionScreen:
			// Initialize editor selection screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if site, ok := data["site"].(system.NginxSite); ok {
						m.editorSelection = screens.NewEditorSelectionModel(site)
					}
				}
			}
		
		case screens.RedisConfigScreen:
			// Initialize Redis config screen
			m.redisConfig = screens.NewRedisConfigModel()
		
		case screens.MySQLManagementScreen:
			// Initialize MySQL management screen
			m.mysqlManagement = screens.NewMySQLManagementModel()
			// Handle success message from sub-screens
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if success, ok := data["success"].(string); ok {
						m.mysqlManagement.SetSuccess(success)
					}
				}
			}
		
		case screens.MySQLPasswordScreen:
			// Initialize MySQL password screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.MySQLManager); ok {
						m.mysqlPassword = screens.NewMySQLPasswordModel(manager)
					}
				}
			}
		
		case screens.MySQLPortScreen:
			// Initialize MySQL port screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.MySQLManager); ok {
						config, _ := data["config"].(*system.MySQLConfig)
						m.mysqlPort = screens.NewMySQLPortModel(manager, config)
					}
				}
			}
		
		case screens.PostgreSQLManagementScreen:
			// Initialize PostgreSQL management screen
			m.postgresqlManagement = screens.NewPostgreSQLManagementModel()
			// Handle success message from sub-screens
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if success, ok := data["success"].(string); ok {
						m.postgresqlManagement.SetSuccess(success)
					}
				}
			}
		
		case screens.PostgreSQLPasswordScreen:
			// Initialize PostgreSQL password screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.PostgreSQLManager); ok {
						m.postgresqlPassword = screens.NewPostgreSQLPasswordModel(manager)
					}
				}
			}
		
		case screens.PostgreSQLPortScreen:
			// Initialize PostgreSQL port screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.PostgreSQLManager); ok {
						config, _ := data["config"].(*system.PostgreSQLConfig)
						m.postgresqlPort = screens.NewPostgreSQLPortModel(manager, config)
					}
				}
			}
		
		case screens.PHPFPMManagementScreen:
			// Initialize PHP-FPM management screen
			m.phpfpmManagement = screens.NewPHPFPMManagementModel()
		
		case screens.SupervisorManagementScreen:
			// Initialize Supervisor management screen
			m.supervisorManagement = screens.NewSupervisorManagementModel()
			// Handle success message from sub-screens
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if success, ok := data["success"].(string); ok {
						m.supervisorManagement.SetSuccess(success)
					}
				}
			}
		
		case screens.SupervisorXMLRPCConfigScreen:
			// Initialize XML-RPC config screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.SupervisorManager); ok {
						m.supervisorXMLRPCConfig = screens.NewSupervisorXMLRPCConfigModel(manager)
					}
				}
			}
		
		case screens.SupervisorAddProgramScreen:
			// Initialize add program screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if manager, ok := data["manager"].(*system.SupervisorManager); ok {
						m.supervisorAddProgram = screens.NewSupervisorAddProgramModel(manager)
					}
				}
			}
		
		case screens.RedisPasswordScreen:
			// Initialize Redis password screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if config, ok := data["config"].(*system.RedisConfig); ok {
						m.redisPassword = screens.NewRedisPasswordModel(config)
					}
				}
			}
		
		case screens.RedisPortScreen:
			// Initialize Redis port screen
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if config, ok := data["config"].(*system.RedisConfig); ok {
						m.redisPort = screens.NewRedisPortModel(config)
					}
				}
			}
		
		case screens.ConfigEditorScreen:
			// Initialize config editor (add site or edit site)
			if msg.Data != nil {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					if action, ok := data["action"].(string); ok {
						if action == "add_nginx_site" {
							m.addSite = screens.NewAddSiteModel()
							m.configEditorActive = "add_site"
						} else if action == "edit_nginx_site" {
							if site, ok := data["site"].(system.NginxSite); ok {
								m.siteDetails = screens.NewSiteDetailsModel(site)
								m.configEditorActive = "site_details"
							}
						}
					}
				}
			}
		}
		
		// Send window size to the new screen immediately after navigation
		if m.width > 0 && m.height > 0 {
			sizeMsg := tea.WindowSizeMsg{Width: m.width, Height: m.height}
			// Combine init cmd and size message
			return m, tea.Batch(initCmd, func() tea.Msg { return sizeMsg })
		}
		return m, initCmd
	
	case screens.ExecutionStartMsg:
		// Switch to execution screen and start execution
		m.currentScreen = screens.ExecutionScreen
		
		// Determine return screen based on current screen
		returnScreen := screens.MainMenuScreen
		switch m.currentScreen {
		case screens.SetupActionScreen:
			returnScreen = screens.SetupMenuScreen
		case screens.QuickCommandsScreen:
			returnScreen = screens.QuickCommandsScreen
		}
		
		m.execution = screens.NewExecutionModel(msg.Command, msg.Description, returnScreen)
		initCmd := m.execution.Init()
		
		// Send window size
		if m.width > 0 && m.height > 0 {
			sizeMsg := tea.WindowSizeMsg{Width: m.width, Height: m.height}
			return m, tea.Batch(initCmd, func() tea.Msg { return sizeMsg })
		}
		return m, initCmd

	case screens.QuitMsg:
		return m, tea.Quit
	}

	// Delegate to current screen
	var cmd tea.Cmd
	switch m.currentScreen {
	case screens.SplashScreen:
		var model tea.Model
		model, cmd = m.splash.Update(msg)
		m.splash = model.(screens.SplashModel)

	case screens.MainMenuScreen:
		var model tea.Model
		model, cmd = m.mainMenu.Update(msg)
		m.mainMenu = model.(screens.MainMenuModel)

	case screens.SetupMenuScreen:
		var model tea.Model
		model, cmd = m.setupMenu.Update(msg)
		m.setupMenu = model.(screens.SetupMenuModel)

	case screens.SetupActionScreen:
		var model tea.Model
		model, cmd = m.setupAction.Update(msg)
		m.setupAction = model.(screens.SetupActionModel)

	case screens.InstalledAppsScreen:
		var model tea.Model
		model, cmd = m.installedApps.Update(msg)
		m.installedApps = model.(screens.InstalledAppsModel)

	case screens.UserManagementScreen:
		var model tea.Model
		model, cmd = m.userManagement.Update(msg)
		m.userManagement = model.(screens.UserManagementModel)

	case screens.UserDetailsScreen:
		var model tea.Model
		model, cmd = m.userDetails.Update(msg)
		m.userDetails = model.(screens.UserDetailsModel)

	case screens.AddUserScreen:
		var model tea.Model
		model, cmd = m.addUser.Update(msg)
		m.addUser = model.(screens.AddUserModel)

	case screens.ConfigMenuScreen:
		var model tea.Model
		model, cmd = m.configMenu.Update(msg)
		m.configMenu = model.(screens.ConfigMenuModel)
	
	case screens.NginxConfigScreen:
		var model tea.Model
		model, cmd = m.nginxConfig.Update(msg)
		m.nginxConfig = model.(screens.NginxConfigModel)

	case screens.QuickCommandsScreen:
		var model tea.Model
		model, cmd = m.quickCommands.Update(msg)
		m.quickCommands = model.(screens.QuickCommandsModel)

	case screens.ExecutionScreen:
		var model tea.Model
		model, cmd = m.execution.Update(msg)
		m.execution = model.(screens.ExecutionModel)
	
	case screens.ConfigEditorScreen:
		// Determine which sub-screen to update based on flag
		if m.configEditorActive == "add_site" {
			var model tea.Model
			model, cmd = m.addSite.Update(msg)
			m.addSite = model.(screens.AddSiteModel)
		} else if m.configEditorActive == "site_details" {
			var model tea.Model
			model, cmd = m.siteDetails.Update(msg)
			m.siteDetails = model.(screens.SiteDetailsModel)
		}
	
	case screens.SSLOptionsScreen:
		var model tea.Model
		model, cmd = m.sslOptions.Update(msg)
		m.sslOptions = model.(screens.SSLOptionsModel)
	
	case screens.SSLManualScreen:
		var model tea.Model
		model, cmd = m.sslManual.Update(msg)
		m.sslManual = model.(screens.SSLManualModel)
	
	case screens.EditorSelectionScreen:
		var model tea.Model
		model, cmd = m.editorSelection.Update(msg)
		m.editorSelection = model.(screens.EditorSelectionModel)
	
	case screens.RedisConfigScreen:
		var model tea.Model
		model, cmd = m.redisConfig.Update(msg)
		m.redisConfig = model.(screens.RedisConfigModel)
	
	case screens.MySQLManagementScreen:
		var model tea.Model
		model, cmd = m.mysqlManagement.Update(msg)
		m.mysqlManagement = model.(screens.MySQLManagementModel)
	
	case screens.MySQLPasswordScreen:
		var model tea.Model
		model, cmd = m.mysqlPassword.Update(msg)
		m.mysqlPassword = model.(screens.MySQLPasswordModel)
	
	case screens.MySQLPortScreen:
		var model tea.Model
		model, cmd = m.mysqlPort.Update(msg)
		m.mysqlPort = model.(screens.MySQLPortModel)
	
	case screens.PostgreSQLManagementScreen:
		var model tea.Model
		model, cmd = m.postgresqlManagement.Update(msg)
		m.postgresqlManagement = model.(screens.PostgreSQLManagementModel)
	
	case screens.PostgreSQLPasswordScreen:
		var model tea.Model
		model, cmd = m.postgresqlPassword.Update(msg)
		m.postgresqlPassword = model.(screens.PostgreSQLPasswordModel)
	
	case screens.PostgreSQLPortScreen:
		var model tea.Model
		model, cmd = m.postgresqlPort.Update(msg)
		m.postgresqlPort = model.(screens.PostgreSQLPortModel)
	
	case screens.PHPFPMManagementScreen:
		var model tea.Model
		model, cmd = m.phpfpmManagement.Update(msg)
		m.phpfpmManagement = model.(screens.PHPFPMManagementModel)
	
	case screens.SupervisorManagementScreen:
		var model tea.Model
		model, cmd = m.supervisorManagement.Update(msg)
		m.supervisorManagement = model.(screens.SupervisorManagementModel)
	
	case screens.SupervisorXMLRPCConfigScreen:
		var model tea.Model
		model, cmd = m.supervisorXMLRPCConfig.Update(msg)
		m.supervisorXMLRPCConfig = model.(screens.SupervisorXMLRPCConfigModel)
	
	case screens.SupervisorAddProgramScreen:
		var model tea.Model
		model, cmd = m.supervisorAddProgram.Update(msg)
		m.supervisorAddProgram = model.(screens.SupervisorAddProgramModel)
	
	case screens.RedisPasswordScreen:
		var model tea.Model
		model, cmd = m.redisPassword.Update(msg)
		m.redisPassword = model.(screens.RedisPasswordModel)
	
	case screens.RedisPortScreen:
		var model tea.Model
		model, cmd = m.redisPort.Update(msg)
		m.redisPort = model.(screens.RedisPortModel)
	}

	return m, cmd
}

// View renders the current screen
func (m Model) View() string {
	switch m.currentScreen {
	case screens.SplashScreen:
		return m.splash.View()
	case screens.MainMenuScreen:
		return m.mainMenu.View()
	case screens.SetupMenuScreen:
		return m.setupMenu.View()
	case screens.SetupActionScreen:
		return m.setupAction.View()
	case screens.InstalledAppsScreen:
		return m.installedApps.View()
	case screens.UserManagementScreen:
		return m.userManagement.View()
	case screens.UserDetailsScreen:
		return m.userDetails.View()
	case screens.AddUserScreen:
		return m.addUser.View()
	case screens.ConfigMenuScreen:
		return m.configMenu.View()
	case screens.NginxConfigScreen:
		return m.nginxConfig.View()
	case screens.QuickCommandsScreen:
		return m.quickCommands.View()
	case screens.ExecutionScreen:
		return m.execution.View()
	case screens.ConfigEditorScreen:
		// Determine which sub-screen to render based on flag
		if m.configEditorActive == "add_site" {
			return m.addSite.View()
		} else if m.configEditorActive == "site_details" {
			return m.siteDetails.View()
		}
		// Fallback to prevent crash
		return "Loading configuration screen..."
	case screens.SSLOptionsScreen:
		return m.sslOptions.View()
	case screens.SSLManualScreen:
		return m.sslManual.View()
	case screens.EditorSelectionScreen:
		return m.editorSelection.View()
	case screens.RedisConfigScreen:
		return m.redisConfig.View()
	case screens.MySQLManagementScreen:
		return m.mysqlManagement.View()
	case screens.MySQLPasswordScreen:
		return m.mysqlPassword.View()
	case screens.MySQLPortScreen:
		return m.mysqlPort.View()
	case screens.PostgreSQLManagementScreen:
		return m.postgresqlManagement.View()
	case screens.PostgreSQLPasswordScreen:
		return m.postgresqlPassword.View()
	case screens.PostgreSQLPortScreen:
		return m.postgresqlPort.View()
	case screens.PHPFPMManagementScreen:
		return m.phpfpmManagement.View()
	case screens.SupervisorManagementScreen:
		return m.supervisorManagement.View()
	case screens.SupervisorXMLRPCConfigScreen:
		return m.supervisorXMLRPCConfig.View()
	case screens.SupervisorAddProgramScreen:
		return m.supervisorAddProgram.View()
	case screens.RedisPasswordScreen:
		return m.redisPassword.View()
	case screens.RedisPortScreen:
		return m.redisPort.View()
	default:
		return "Unknown screen"
	}
}

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("Ravact version %s\n", Version)
		os.Exit(0)
	}

	// Set embedded FS for screens to use
	screens.EmbeddedFS = embeddedAssets

	// Create and run the program
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
