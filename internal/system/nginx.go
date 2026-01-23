package system

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NginxSite represents an Nginx site configuration
type NginxSite struct {
	Name      string
	Domain    string
	RootDir   string
	IsEnabled bool
	HasSSL    bool
	ConfigPath string
}

// NginxTemplate represents a site template from JSON
type NginxTemplate struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	DefaultIndex   string   `json:"default_index"`
	RequiresPHP    bool     `json:"requires_php"`
	PHPVersion     string   `json:"php_version,omitempty"`
	PublicDir      string   `json:"public_dir,omitempty"`
	RecommendedFor []string `json:"recommended_for,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

// NginxTemplatesConfig holds all templates
type NginxTemplatesConfig struct {
	Templates []NginxTemplate `json:"templates"`
}

// NginxManager handles Nginx configuration operations
type NginxManager struct {
	sitesAvailable string
	sitesEnabled   string
	embeddedFS     *embed.FS
	templates      []NginxTemplate
}

// NewNginxManager creates a new Nginx manager
func NewNginxManager() *NginxManager {
	return &NginxManager{
		sitesAvailable: "/etc/nginx/sites-available",
		sitesEnabled:   "/etc/nginx/sites-enabled",
		embeddedFS:     nil,
		templates:      []NginxTemplate{},
	}
}

// SetEmbeddedFS sets the embedded filesystem for loading templates
func (nm *NginxManager) SetEmbeddedFS(fs *embed.FS) {
	nm.embeddedFS = fs
	nm.loadTemplates()
}

// loadTemplates loads nginx templates from embedded JSON
func (nm *NginxManager) loadTemplates() {
	if nm.embeddedFS == nil {
		return
	}

	data, err := nm.embeddedFS.ReadFile("assets/configs/nginx-templates.json")
	if err != nil {
		// Fallback to hardcoded templates if file not found
		return
	}

	var config NginxTemplatesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	nm.templates = config.Templates
}

// GetTemplates returns available nginx templates
func (nm *NginxManager) GetTemplates() []NginxTemplate {
	return nm.templates
}

// GetAllSites returns all available sites
func (nm *NginxManager) GetAllSites() ([]NginxSite, error) {
	entries, err := os.ReadDir(nm.sitesAvailable)
	if err != nil {
		if os.IsNotExist(err) {
			return []NginxSite{}, nil
		}
		return nil, err
	}

	var sites []NginxSite
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip default site
		if name == "default" {
			continue
		}

		configPath := filepath.Join(nm.sitesAvailable, name)
		
		// Check if enabled (symlink exists)
		isEnabled := false
		enabledPath := filepath.Join(nm.sitesEnabled, name)
		if _, err := os.Lstat(enabledPath); err == nil {
			isEnabled = true
		}

		// Parse config to get details
		domain, rootDir, hasSSL := nm.parseConfig(configPath)

		site := NginxSite{
			Name:       name,
			Domain:     domain,
			RootDir:    rootDir,
			IsEnabled:  isEnabled,
			HasSSL:     hasSSL,
			ConfigPath: configPath,
		}

		sites = append(sites, site)
	}

	return sites, nil
}

// parseConfig extracts basic info from nginx config
func (nm *NginxManager) parseConfig(configPath string) (domain, rootDir string, hasSSL bool) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", false
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Extract server_name
		if strings.HasPrefix(line, "server_name") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				domain = strings.TrimSuffix(parts[1], ";")
			}
		}
		
		// Extract root
		if strings.HasPrefix(line, "root ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				rootDir = strings.TrimSuffix(parts[1], ";")
			}
		}
		
		// Check for SSL
		if strings.Contains(line, "listen 443") || strings.Contains(line, "ssl_certificate") {
			hasSSL = true
		}
	}

	return domain, rootDir, hasSSL
}

// EnableSite enables a site by creating symlink
func (nm *NginxManager) EnableSite(siteName string) error {
	availablePath := filepath.Join(nm.sitesAvailable, siteName)
	enabledPath := filepath.Join(nm.sitesEnabled, siteName)

	// Check if site exists
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return fmt.Errorf("site not found: %s", siteName)
	}

	// Create symlink
	if err := os.Symlink(availablePath, enabledPath); err != nil {
		return fmt.Errorf("failed to enable site: %w", err)
	}

	return nil
}

// DisableSite disables a site by removing symlink
func (nm *NginxManager) DisableSite(siteName string) error {
	enabledPath := filepath.Join(nm.sitesEnabled, siteName)

	// Remove symlink
	if err := os.Remove(enabledPath); err != nil {
		return fmt.Errorf("failed to disable site: %w", err)
	}

	return nil
}

// TestConfig tests nginx configuration
func (nm *NginxManager) TestConfig() error {
	cmd := exec.Command("nginx", "-t")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", string(output))
	}

	return nil
}

// ReloadNginx reloads nginx configuration
func (nm *NginxManager) ReloadNginx() error {
	cmd := exec.Command("systemctl", "reload", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// CreateSite creates a new site configuration
func (nm *NginxManager) CreateSite(siteName, domain, rootDir, template string, useSSL, useCertbot bool) error {
	configPath := filepath.Join(nm.sitesAvailable, siteName)

	// Check if site already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("site already exists: %s", siteName)
	}

	// Generate config based on template and options
	config := nm.generateConfig(domain, rootDir, template, useSSL, useCertbot)

	// Write config file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// generateConfig generates nginx configuration based on parameters
func (nm *NginxManager) generateConfig(domain, rootDir, template string, useSSL, useCertbot bool) string {
	var config strings.Builder

	if !useSSL {
		// HTTP only
		config.WriteString(fmt.Sprintf(`server {
    listen 80;
    listen [::]:80;
    server_name %s;

    root %s;
    index index.html index.htm index.php;

    # Logging
    access_log /var/log/nginx/%s-access.log;
    error_log /var/log/nginx/%s-error.log;

`, domain, rootDir, domain, domain))

		// Add template-specific directives
		config.WriteString(nm.getTemplateDirectives(template))

		config.WriteString("}\n")
	} else if useCertbot {
		// HTTP with redirect (for certbot challenge and redirect)
		config.WriteString(fmt.Sprintf(`server {
    listen 80;
    listen [::]:80;
    server_name %s;

    # Allow certbot challenges
    location /.well-known/acme-challenge/ {
        root %s;
    }

    # Redirect all other HTTP traffic to HTTPS
    location / {
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name %s;

    root %s;
    index index.html index.htm index.php;

    # SSL Configuration (will be set by certbot)
    ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;
    
    # SSL Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Logging
    access_log /var/log/nginx/%s-access.log;
    error_log /var/log/nginx/%s-error.log;

`, domain, rootDir, domain, rootDir, domain, domain, domain, domain))

		// Add template-specific directives
		config.WriteString(nm.getTemplateDirectives(template))

		config.WriteString("}\n")
	} else {
		// HTTPS with placeholder certificates
		config.WriteString(fmt.Sprintf(`server {
    listen 80;
    listen [::]:80;
    server_name %s;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name %s;

    root %s;
    index index.html index.htm index.php;

    # SSL Configuration (UPDATE WITH YOUR CERTIFICATES)
    # ssl_certificate /path/to/your/certificate.crt;
    # ssl_certificate_key /path/to/your/private.key;
    
    # Uncomment and update the paths above, then remove this line:
    # For now, using self-signed or existing certificates
    
    # SSL Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Logging
    access_log /var/log/nginx/%s-access.log;
    error_log /var/log/nginx/%s-error.log;

`, domain, domain, rootDir, domain, domain))

		// Add template-specific directives
		config.WriteString(nm.getTemplateDirectives(template))

		config.WriteString("}\n")
	}

	return config.String()
}

// getTemplateDirectives returns nginx directives for specific templates
func (nm *NginxManager) getTemplateDirectives(template string) string {
	switch template {
	case "php":
		return `    # PHP Configuration
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php-fpm.sock;
    }

    location ~ /\.ht {
        deny all;
    }

`
	case "laravel":
		return `    # Laravel Configuration
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }

`
	case "wordpress":
		return `    # WordPress Configuration
    location / {
        try_files $uri $uri/ /index.php?$args;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php-fpm.sock;
    }

    location ~ /\.ht {
        deny all;
    }

    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }

    location = /robots.txt {
        allow all;
        log_not_found off;
        access_log off;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
        log_not_found off;
    }

`
	default:
		return `    # Static file configuration
    location / {
        try_files $uri $uri/ =404;
    }

`
	}
}

// DeleteSite deletes a site configuration
func (nm *NginxManager) DeleteSite(siteName string) error {
	// Disable first if enabled
	_ = nm.DisableSite(siteName)

	// Delete config file
	configPath := filepath.Join(nm.sitesAvailable, siteName)
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to delete site: %w", err)
	}

	return nil
}

// ObtainSSLCertificate obtains SSL certificate using certbot
func (nm *NginxManager) ObtainSSLCertificate(domain string) error {
	cmd := exec.Command("certbot", "--nginx", "-d", domain, "--non-interactive", "--agree-tos", "--email", "admin@"+domain)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("certbot failed: %s", string(output))
	}

	return nil
}

// AddSSLManual adds manual SSL certificates to a site
func (nm *NginxManager) AddSSLManual(siteName, certPath, keyPath, chainPath string) error {
	configPath := filepath.Join(nm.sitesAvailable, siteName)
	
	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read site config: %w", err)
	}
	
	config := string(content)
	
	// Check if already has SSL
	if strings.Contains(config, "ssl_certificate") {
		return fmt.Errorf("site already has SSL configured")
	}
	
	// Find server block and add SSL directives
	// Add listen 443 ssl
	config = strings.Replace(config, "listen 80;", "listen 80;\n    listen 443 ssl;", 1)
	config = strings.Replace(config, "listen [::]:80;", "listen [::]:80;\n    listen [::]:443 ssl;", 1)
	
	// Add SSL certificate paths after server_name
	sslDirectives := fmt.Sprintf("\n\n    # SSL Configuration\n    ssl_certificate %s;\n    ssl_certificate_key %s;", certPath, keyPath)
	if chainPath != "" {
		sslDirectives += fmt.Sprintf("\n    ssl_trusted_certificate %s;", chainPath)
	}
	sslDirectives += "\n    ssl_protocols TLSv1.2 TLSv1.3;\n    ssl_ciphers HIGH:!aNULL:!MD5;\n    ssl_prefer_server_ciphers on;"
	
	// Insert after server_name line
	lines := strings.Split(config, "\n")
	var newLines []string
	for _, line := range lines {
		newLines = append(newLines, line)
		if strings.Contains(line, "server_name") {
			newLines = append(newLines, sslDirectives)
		}
	}
	config = strings.Join(newLines, "\n")
	
	// Write updated config
	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// RemoveSSL removes SSL configuration from a site
func (nm *NginxManager) RemoveSSL(siteName string) error {
	configPath := filepath.Join(nm.sitesAvailable, siteName)
	
	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read site config: %w", err)
	}
	
	config := string(content)
	
	// Remove SSL listen directives
	config = strings.ReplaceAll(config, "listen 443 ssl;", "")
	config = strings.ReplaceAll(config, "listen [::]:443 ssl;", "")
	
	// Remove SSL certificate directives
	lines := strings.Split(config, "\n")
	var newLines []string
	skipSSLBlock := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip SSL configuration lines
		if strings.HasPrefix(trimmed, "# SSL Configuration") {
			skipSSLBlock = true
			continue
		}
		
		if skipSSLBlock {
			if strings.HasPrefix(trimmed, "ssl_") {
				continue
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				skipSSLBlock = false
			}
		}
		
		// Skip individual SSL directives
		if strings.HasPrefix(trimmed, "ssl_certificate") ||
			strings.HasPrefix(trimmed, "ssl_protocols") ||
			strings.HasPrefix(trimmed, "ssl_ciphers") ||
			strings.HasPrefix(trimmed, "ssl_prefer_server_ciphers") ||
			strings.HasPrefix(trimmed, "ssl_trusted_certificate") {
			continue
		}
		
		newLines = append(newLines, line)
	}
	
	config = strings.Join(newLines, "\n")
	
	// Clean up extra blank lines
	config = strings.ReplaceAll(config, "\n\n\n", "\n\n")
	
	// Write updated config
	err = os.WriteFile(configPath, []byte(config), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}
