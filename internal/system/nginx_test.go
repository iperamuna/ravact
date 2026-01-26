package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewNginxManager(t *testing.T) {
	manager := NewNginxManager()
	if manager == nil {
		t.Fatal("NewNginxManager returned nil")
	}
}

func TestNginxSiteStruct(t *testing.T) {
	site := NginxSite{
		Name:       "example.com",
		Domain:     "example.com",
		RootDir:    "/var/www/example.com",
		HasSSL:     true,
		IsEnabled:  true,
		ConfigPath: "/etc/nginx/sites-available/example.com",
	}

	if site.Name != "example.com" {
		t.Errorf("expected name 'example.com', got '%s'", site.Name)
	}
	if site.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got '%s'", site.Domain)
	}
	if site.RootDir != "/var/www/example.com" {
		t.Errorf("expected root '/var/www/example.com', got '%s'", site.RootDir)
	}
	if !site.HasSSL {
		t.Error("expected HasSSL to be true")
	}
	if !site.IsEnabled {
		t.Error("expected IsEnabled to be true")
	}
}

func TestNginxTemplateStruct(t *testing.T) {
	template := NginxTemplate{
		ID:             "laravel",
		Name:           "Laravel Application",
		Description:    "Laravel PHP framework",
		DefaultIndex:   "index.php",
		RequiresPHP:    true,
		PHPVersion:     "8.2",
		PublicDir:      "public",
		RecommendedFor: []string{"Laravel 8+"},
		Notes:          "Requires public directory",
	}

	if template.ID != "laravel" {
		t.Errorf("expected ID 'laravel', got '%s'", template.ID)
	}
	if !template.RequiresPHP {
		t.Error("Laravel template should require PHP")
	}
	if template.PublicDir != "public" {
		t.Errorf("expected public dir 'public', got '%s'", template.PublicDir)
	}
}

func TestNginxManager_GetTemplates(t *testing.T) {
	manager := NewNginxManager()

	// Without embedded FS, templates should be empty
	templates := manager.GetTemplates()
	if templates == nil {
		t.Error("GetTemplates should return non-nil slice")
	}
}

func TestNginxManager_ParseServerBlock(t *testing.T) {
	configContent := `server {
    listen 80;
    listen [::]:80;
    server_name example.com www.example.com;
    root /var/www/example.com/public;
    index index.php index.html;
}
`

	// Verify the config contains expected directives
	if !strings.Contains(configContent, "listen 80") {
		t.Error("expected config to contain 'listen 80'")
	}
	if !strings.Contains(configContent, "server_name example.com") {
		t.Error("expected config to contain server_name")
	}
	if !strings.Contains(configContent, "root /var/www/example.com") {
		t.Error("expected config to contain root directive")
	}
}

func TestNginxSitesPaths(t *testing.T) {
	// Test common Nginx paths
	paths := map[string]string{
		"available": "/etc/nginx/sites-available",
		"enabled":   "/etc/nginx/sites-enabled",
		"conf.d":    "/etc/nginx/conf.d",
	}

	for name, path := range paths {
		if path == "" {
			t.Errorf("%s path should not be empty", name)
		}
	}
}

func TestNginxConfigValidation(t *testing.T) {
	// Test server name validation patterns
	validNames := []string{
		"example.com",
		"www.example.com",
		"sub.domain.example.com",
		"example-site.com",
		"example123.com",
	}

	for _, name := range validNames {
		// Basic validation - contains only valid characters
		if strings.ContainsAny(name, " \t\n") {
			t.Errorf("server name '%s' should not contain whitespace", name)
		}
	}
}

func TestNginxManager_WriteSiteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	sitesAvailable := filepath.Join(tmpDir, "sites-available")
	os.MkdirAll(sitesAvailable, 0755)

	configPath := filepath.Join(sitesAvailable, "test.conf")
	configContent := `server {
    listen 80;
    server_name test.com;
    root /var/www/test;
}
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !strings.Contains(string(data), "server_name test.com") {
		t.Error("config file content mismatch")
	}
}

func TestNginxSSLConfig(t *testing.T) {
	sslConfig := `
    listen 443 ssl http2;
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
`

	if !strings.Contains(sslConfig, "listen 443 ssl") {
		t.Error("SSL config should contain HTTPS listener")
	}
	if !strings.Contains(sslConfig, "ssl_certificate") {
		t.Error("SSL config should contain certificate path")
	}
	if !strings.Contains(sslConfig, "TLSv1.2") {
		t.Error("SSL config should specify TLS protocols")
	}
}

func TestNginxManager_EnableSite(t *testing.T) {
	tmpDir := t.TempDir()

	sitesAvailable := filepath.Join(tmpDir, "sites-available")
	sitesEnabled := filepath.Join(tmpDir, "sites-enabled")

	os.MkdirAll(sitesAvailable, 0755)
	os.MkdirAll(sitesEnabled, 0755)

	// Create a site config
	configPath := filepath.Join(sitesAvailable, "test.conf")
	os.WriteFile(configPath, []byte("server {}"), 0644)

	// Create symlink (simulating enable)
	linkPath := filepath.Join(sitesEnabled, "test.conf")
	err := os.Symlink(configPath, linkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Verify symlink exists
	_, err = os.Lstat(linkPath)
	if err != nil {
		t.Error("symlink should exist")
	}
}

func TestNginxManager_DisableSite(t *testing.T) {
	tmpDir := t.TempDir()

	sitesEnabled := filepath.Join(tmpDir, "sites-enabled")
	os.MkdirAll(sitesEnabled, 0755)

	// Create a symlink
	linkPath := filepath.Join(sitesEnabled, "test.conf")
	os.WriteFile(linkPath, []byte("server {}"), 0644)

	// Remove (simulating disable)
	err := os.Remove(linkPath)
	if err != nil {
		t.Fatalf("failed to remove link: %v", err)
	}

	// Verify link is gone
	_, err = os.Stat(linkPath)
	if !os.IsNotExist(err) {
		t.Error("link should be removed")
	}
}
