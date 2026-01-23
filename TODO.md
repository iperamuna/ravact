# Ravact TODO

## 🔧 High Priority - UI Screen Refactoring

The following screen files need to be refactored to match the existing Model pattern:

### Files to Fix:
- [ ] `internal/ui/screens/mysql_management.go`
- [ ] `internal/ui/screens/postgresql_management.go`
- [ ] `internal/ui/screens/phpfpm_management.go`
- [ ] `internal/ui/screens/supervisor_management.go`
- [ ] `internal/ui/screens/text_display.go`

### Issue:
These screens currently have compilation errors because they use:
- ❌ `theme.HeaderStyle` (global package usage - doesn't exist)
- ❌ `NewMainMenuScreen()` (function doesn't exist)
- ❌ Direct theme imports without Model pattern

### Required Pattern (from existing screens):
```go
type MyScreenModel struct {
    theme  *theme.Theme  // Theme as struct field
    width  int
    height int
    // ... other fields
}

func NewMyScreenModel() MyScreenModel {
    return MyScreenModel{
        theme: theme.DefaultTheme(),  // Initialize theme
        // ...
    }
}

func (m MyScreenModel) View() string {
    header := m.theme.Title.Render("My Screen")  // Use m.theme.*
    // ...
}

// Navigation
return m, func() tea.Msg {
    return NavigateMsg{Screen: ConfigMenuScreen}  // Not NewMainMenuScreen()
}
```

### Reference Files:
- ✅ `internal/ui/screens/redis_config_screen.go` - Perfect example
- ✅ `internal/ui/screens/nginx_config.go` - Another good example
- ✅ `internal/ui/screens/quick_commands.go` - Shows command execution

### What Works:
- ✅ System layer implementation (mysql.go, postgresql.go, phpfpm.go, supervisor.go)
- ✅ Configuration menu integration (screen types defined)
- ✅ Comprehensive documentation
- ✅ All business logic

### What's Needed:
- Fix UI screens to use Model pattern with `m.theme.*`
- Update navigation to use NavigateMsg
- Test compilation after fixes
- Verify in VM

---

## 📋 Medium Priority

### Testing
- [ ] Test Quick Commands feature thoroughly
- [ ] Test Redis configuration (working baseline)
- [ ] Test Nginx configuration (working baseline)
- [ ] Test User Management (working baseline)
- [ ] Automated test suite for new features

### Documentation
- [x] Database management guide
- [x] PHP-FPM & Supervisor guide
- [x] M1 Mac testing guide
- [x] AMD64 testing guide
- [ ] Add screenshots/demos to documentation

### Features
- [ ] Database backup/restore functionality
- [ ] PHP-FPM pool templates
- [ ] Supervisor program templates
- [ ] Bulk operations

---

## 🔮 Low Priority

### Enhancements
- [ ] Configuration file syntax highlighting
- [ ] Real-time log viewing
- [ ] Service dependency management
- [ ] Multi-site SSL management
- [ ] Automated backups scheduler

### CI/CD
- [ ] Add GitHub Actions for automated testing
- [ ] Cross-compile releases (AMD64 + ARM64)
- [ ] Automated documentation deployment

---

## ✅ Completed

- [x] Core system layer for MySQL management
- [x] Core system layer for PostgreSQL management
- [x] Core system layer for PHP-FPM management
- [x] Enhanced Supervisor management with XML-RPC
- [x] Configuration menu integration
- [x] Navigation system updates
- [x] Comprehensive documentation (4 new guides)
- [x] Documentation cleanup (removed 16 duplicate files)
- [x] M1 Mac testing setup with Multipass
- [x] Service installation in test VM

---

## 📝 Notes

**Current Status:**
- 27 commits total
- 114 tracked files
- All services tested and working on Ubuntu 24.04 ARM64
- Main blocker: UI screen refactoring to match existing pattern

**Estimated Effort:**
- Screen refactoring: ~2-3 hours
- Testing after refactor: ~1 hour
- Total: ~3-4 hours

**Next Steps:**
1. Refactor the 5 screen files to use Model pattern
2. Test compilation
3. Deploy to VM and test all features
4. Create final commit
5. Update README with status
