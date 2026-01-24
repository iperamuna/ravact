# Ravact TODO

## ✅ Status: Feature Complete (v0.1.3)

All core features have been implemented and are production-ready. See [PROJECT_STATUS.md](docs/project/PROJECT_STATUS.md) for full details.

---

## 🎯 Future Enhancements

These are potential improvements that could be added in future versions:

### Database Management
- [ ] Database backup/restore functionality
- [ ] Database user management (create, delete, permissions)
- [ ] Query execution interface
- [ ] Database export/import with scheduling

### PHP-FPM Management
- [ ] Create/Edit/Delete pools via UI
- [ ] Pool templates (low traffic, high traffic, etc.)
- [ ] Real-time pool status monitoring
- [ ] Worker process statistics

### Supervisor Management
- [ ] Edit existing programs via UI
- [ ] Program logs viewer
- [ ] Start/Stop/Restart individual programs
- [ ] Process resource monitoring
- [ ] Group management

### Nginx Management
- [ ] Site templates expansion
- [ ] Configuration syntax validation
- [ ] Live config testing (nginx -t)
- [ ] Access/Error log viewer
- [ ] Rate limiting configuration

### Monitoring & Alerts
- [ ] Real-time resource monitoring dashboard
- [ ] Service health checks
- [ ] Email/SMS alerts for service failures
- [ ] Log analysis and search
- [ ] Performance metrics tracking

### Backup & Recovery
- [ ] Automated backup scheduling
- [ ] Configuration snapshots
- [ ] One-click restore
- [ ] Backup to remote storage (S3, FTP, etc.)

### Security Enhancements
- [ ] Firewall management (ufw/iptables)
- [ ] SSL certificate renewal automation
- [ ] Security audit reports
- [ ] Intrusion detection integration
- [ ] Password strength enforcement

### Multi-Server Management
- [ ] Connect to multiple servers
- [ ] Sync configurations across servers
- [ ] Load balancer configuration
- [ ] Cluster management

### Advanced Features
- [ ] Configuration file syntax highlighting in editor
- [ ] Diff viewer for configuration changes
- [ ] Git integration for config versioning
- [ ] Scheduled tasks management (cron)
- [ ] Package update management
- [ ] Docker container management

---

## 🐛 Known Issues

None currently reported! All features are working as expected.

---

## 📝 Notes for Contributors

### Development Setup
1. See `docs/development/DEVELOPMENT.md` for setup instructions
2. Testing: Use Multipass (M1 Mac) or real AMD64 hardware
3. Follow existing Model pattern for new screens
4. All PRs must include tests

### Code Standards
- Follow existing Model pattern (see `internal/ui/screens/redis_config_screen.go`)
- Use `m.theme.*` for all styling
- Navigate with `NavigateMsg`, not function calls
- Validate all user input
- Handle errors gracefully

### Testing Checklist
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Tested on ARM64 (Multipass)
- [ ] Tested on AMD64 (real hardware)
- [ ] Documentation updated
- [ ] No breaking changes

---

## 📞 Support

- **Issues:** Report on GitHub
- **Documentation:** See `docs/` directory
- **Testing:** See `docs/testing/COMPLETE_TEST_CASES.md`

---

**Last Updated:** January 2026
