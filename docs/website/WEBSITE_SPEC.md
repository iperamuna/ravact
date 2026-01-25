# Ravact Website Specification

## AI Agent Instructions

This document provides comprehensive specifications for building the Ravact marketing website using Laravel. Follow these guidelines to create a modern, professional website that reflects the power and elegance of the Ravact TUI application.

---

## Brand Identity

### The Name: Ravact

**Ravact = Ravana + Act**

The name draws inspiration from **Ravana**, the legendary king from ancient Sri Lankan mythology. Ravana was:

- **A Powerful Ruler** - King of Lanka (ancient Sri Lanka), known for his immense strength and authority
- **A Scholar & Intellectual** - Master of the Vedas, accomplished in music (invented the Ravanhattha instrument), medicine, and astronomy
- **A Master Builder** - His kingdom Lanka was described as a golden city with advanced architecture
- **A Ten-Headed King** - The ten heads symbolize mastery over the ten directions and complete knowledge
- **A Devotee of Shiva** - Known for his devotion and the composition of the Shiva Tandava Stotram
- **Technologically Advanced** - Legends describe his Pushpaka Vimana (flying chariot), representing advanced engineering

**Brand Message**: Like Ravana who commanded his kingdom with absolute authority and technological prowess, Ravact gives sysadmins **powerful command** over their Linux servers through an elegant, intelligent interface.

### Brand Personality

- **Powerful** - Complete server control at your fingertips
- **Intelligent** - Smart detection, auto-configuration, contextual menus
- **Elegant** - Beautiful TUI that makes complex tasks simple
- **Authoritative** - Professional tool for serious server management
- **Modern** - Built with cutting-edge Go and Charm libraries

### Color Palette

```css
:root {
  /* Primary Colors */
  --primary: #FF6B35;        /* Vibrant Orange - Main accent, CTAs */
  --primary-dark: #E55A2B;   /* Darker Orange - Hover states */
  --primary-light: #FF8F66;  /* Lighter Orange - Backgrounds */
  
  /* Secondary Colors */
  --secondary: #004E89;      /* Deep Blue - Headers, navigation */
  --secondary-dark: #003D6B; /* Darker Blue - Hover states */
  --secondary-light: #0066B3;/* Lighter Blue - Links */
  
  /* Semantic Colors */
  --success: #2ECC71;        /* Green - Success states */
  --warning: #F39C12;        /* Yellow/Amber - Warnings */
  --error: #E74C3C;          /* Red - Errors */
  --info: #3498DB;           /* Light Blue - Information */
  
  /* Neutral Colors */
  --text-primary: #1A1A2E;   /* Dark navy - Main text */
  --text-secondary: #4A4A68; /* Medium gray - Secondary text */
  --text-muted: #7F8C8D;     /* Light gray - Muted text */
  --background: #FAFBFC;     /* Off-white - Page background */
  --surface: #FFFFFF;        /* White - Card backgrounds */
  --border: #E5E7EB;         /* Light gray - Borders */
  
  /* Dark Mode (for code blocks, terminal previews) */
  --terminal-bg: #1E1E2E;    /* Dark background */
  --terminal-text: #CDD6F4;  /* Light text */
  --terminal-green: #A6E3A1; /* Terminal green */
  --terminal-yellow: #F9E2AF;/* Terminal yellow */
  --terminal-orange: #FAB387;/* Terminal orange */
}
```

### Typography

```css
/* Font Stack */
--font-heading: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, sans-serif;
--font-body: 'Inter', 'SF Pro Text', -apple-system, BlinkMacSystemFont, sans-serif;
--font-mono: 'JetBrains Mono', 'SF Mono', 'Fira Code', monospace;

/* Font Sizes */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
--text-4xl: 2.25rem;   /* 36px */
--text-5xl: 3rem;      /* 48px */
--text-6xl: 3.75rem;   /* 60px */

/* Font Weights */
--font-normal: 400;
--font-medium: 500;
--font-semibold: 600;
--font-bold: 700;
```

---

## Design System

### Design Philosophy

Follow **Material Tailwind** inspired design with these principles:

1. **Flat Design** - No gradients, shadows are subtle and purposeful
2. **Small Rounded Corners** - Use `rounded-md` (6px) or `rounded-lg` (8px), NOT `rounded-xl` or larger
3. **Generous Whitespace** - Let content breathe, avoid cramped layouts
4. **Section Variety** - Alternate between full-width sections, cards, and open layouts
5. **Subtle Animations** - Smooth transitions, no flashy effects
6. **Clean Lines** - Use borders sparingly, rely on spacing for separation

### Border Radius Guidelines

```css
/* DO USE */
.rounded-sm { border-radius: 4px; }   /* Small elements, tags */
.rounded-md { border-radius: 6px; }   /* Buttons, inputs */
.rounded-lg { border-radius: 8px; }   /* Cards, containers */

/* DO NOT USE for main UI elements */
.rounded-xl { border-radius: 12px; }  /* Only for special callouts */
.rounded-2xl { border-radius: 16px; } /* Avoid */
.rounded-3xl { border-radius: 24px; } /* Avoid */
```

### Section Design Patterns

**Pattern 1: Open Section (No Boxes)**
```html
<section class="py-20 bg-white">
  <div class="max-w-6xl mx-auto px-6">
    <h2 class="text-3xl font-bold text-center mb-4">Features</h2>
    <p class="text-text-secondary text-center max-w-2xl mx-auto mb-16">
      Everything you need to manage your servers efficiently
    </p>
    <div class="grid md:grid-cols-3 gap-12">
      <!-- Feature items WITHOUT cards, just icon + text -->
      <div class="text-center">
        <div class="w-14 h-14 bg-primary/10 rounded-lg flex items-center justify-center mx-auto mb-4">
          <svg class="w-7 h-7 text-primary">...</svg>
        </div>
        <h3 class="font-semibold text-lg mb-2">Feature Name</h3>
        <p class="text-text-secondary">Feature description goes here.</p>
      </div>
    </div>
  </div>
</section>
```

**Pattern 2: Alternating Content**
```html
<section class="py-20 bg-background">
  <div class="max-w-6xl mx-auto px-6">
    <div class="flex flex-col md:flex-row items-center gap-12">
      <div class="md:w-1/2">
        <span class="text-primary font-medium text-sm uppercase tracking-wide">Category</span>
        <h2 class="text-3xl font-bold mt-2 mb-4">Section Title</h2>
        <p class="text-text-secondary mb-6">Description paragraph.</p>
        <ul class="space-y-3">
          <li class="flex items-start gap-3">
            <svg class="w-5 h-5 text-success mt-0.5">...</svg>
            <span>Feature point</span>
          </li>
        </ul>
      </div>
      <div class="md:w-1/2">
        <!-- Image or terminal preview -->
      </div>
    </div>
  </div>
</section>
```

**Pattern 3: Cards Grid (When Needed)**
```html
<section class="py-20 bg-white">
  <div class="max-w-6xl mx-auto px-6">
    <div class="grid md:grid-cols-3 gap-6">
      <div class="bg-background border border-border rounded-lg p-6 hover:border-primary/30 transition-colors">
        <h3 class="font-semibold mb-2">Card Title</h3>
        <p class="text-text-secondary text-sm">Card content.</p>
      </div>
    </div>
  </div>
</section>
```

### Button Styles

```html
<!-- Primary Button -->
<button class="bg-primary text-white px-6 py-3 rounded-md font-medium hover:bg-primary-dark transition-colors">
  Get Started
</button>

<!-- Secondary Button -->
<button class="border border-primary text-primary px-6 py-3 rounded-md font-medium hover:bg-primary/5 transition-colors">
  Learn More
</button>

<!-- Ghost Button -->
<button class="text-text-secondary px-4 py-2 rounded-md hover:text-primary hover:bg-primary/5 transition-colors">
  View Docs
</button>
```

---

## Page Structure

### 1. Homepage

#### Hero Section
- **Headline**: "Command Your Servers Like a King"
- **Subheadline**: "Ravact is a modern TUI application that gives you powerful, intuitive control over your Linux servers. Install software, manage services, and configure sites—all from a beautiful terminal interface."
- **CTA Buttons**: "Get Started" (primary), "View on GitHub" (secondary)
- **Hero Visual**: Animated terminal showing Ravact TUI demo (see Hero Animation Specification below)

#### Features Overview
Open layout (no cards), icons with descriptions:

1. **One-Click Installation**
   - Install 13+ server packages with a single command
   - Nginx, MySQL, PostgreSQL, Redis, PHP, Node.js, and more

2. **Smart Service Detection**
   - Automatically detects installed services
   - Shows real-time status (Running, Stopped, Failed)

3. **Complete Site Management**
   - 7 site templates (Laravel, WordPress, Static, etc.)
   - SSL with Let's Encrypt or manual certificates

4. **Developer Toolkit**
   - 34+ essential commands for Laravel & WordPress
   - Copy to clipboard, execute directly

5. **File Browser**
   - Full-featured terminal file manager
   - Preview, search, copy, paste, delete

6. **Modern UI/UX**
   - Beautiful forms, categorized menus
   - Works in web terminals (xterm.js)

#### Terminal Preview Section
Large terminal mockup showing:
```
┌─────────────────────────────────────────────────────────────┐
│  Ravact v0.2.1                    ravact-server (10.0.0.5)  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  📦 Package Management                                      │
│     ▶ Install Software                                      │
│       Installed Applications                                │
│                                                             │
│  ⚙️  Service Configuration                                  │
│       Service Settings                                      │
│                                                             │
│  🌐 Site Management                                         │
│       Site Commands                                         │
│       Developer Toolkit                                     │
│                                                             │
│  👥 System Administration                                   │
│       User Management                                       │
│       Quick Commands                                        │
│                                                             │
│  🔧 Tools                                                   │
│       File Browser                                          │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  Host: production-01 (192.168.1.100)                        │
│  OS: Ubuntu 24.04 LTS │ Arch: x86_64 │ RAM: 16 GB           │
│                                                             │
│  ↑/↓ Navigate  Enter Select  q Quit                         │
└─────────────────────────────────────────────────────────────┘
```

#### Installation Section
Clean, step-by-step:

```bash
# One-command install
curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash

# Then run
sudo ravact
```

Available for:
- Linux x86_64 (Intel/AMD)
- Linux ARM64 (Raspberry Pi, AWS Graviton)
- macOS (UI preview only)

#### Feature Deep-Dive Sections

**Software Setup & Management**
- List of 13+ packages with icons
- Screenshot of setup menu

**Database Management**
- MySQL & PostgreSQL features
- Port configuration, password management
- Performance tuning

**Nginx & SSL**
- Site templates showcase
- SSL options (Let's Encrypt, Manual)

**Developer Toolkit**
- Command categories (Laravel, WordPress, PHP, Security)
- Example commands with descriptions

**File Browser**
- Keyboard shortcuts highlight
- Screenshot of file browser

#### Testimonials/Use Cases (Optional)
Simple quotes without heavy card designs

#### Footer
- Links: Documentation, GitHub, Changelog
- Author: Indunil Peramuna
- Version badge

---

### 2. Documentation Page

Link to full docs or embed key sections:
- Quick Start Guide
- Feature Guides
- Keyboard Shortcuts
- Troubleshooting

---

### 3. Download Page

- Detect user's OS/architecture
- Highlight recommended download
- All download options listed
- Build from source instructions

---

## Components

### Navigation
```html
<nav class="fixed top-0 w-full bg-white/80 backdrop-blur-sm border-b border-border z-50">
  <div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
    <a href="/" class="flex items-center gap-2">
      <span class="text-2xl font-bold text-primary">Ravact</span>
    </a>
    <div class="hidden md:flex items-center gap-8">
      <a href="#features" class="text-text-secondary hover:text-primary transition-colors">Features</a>
      <a href="#installation" class="text-text-secondary hover:text-primary transition-colors">Install</a>
      <a href="/docs" class="text-text-secondary hover:text-primary transition-colors">Docs</a>
      <a href="https://github.com/iperamuna/ravact" class="text-text-secondary hover:text-primary transition-colors">GitHub</a>
      <a href="#installation" class="bg-primary text-white px-4 py-2 rounded-md font-medium hover:bg-primary-dark transition-colors">
        Get Started
      </a>
    </div>
  </div>
</nav>
```

### Terminal Component
```html
<div class="bg-terminal-bg rounded-lg overflow-hidden shadow-lg">
  <div class="flex items-center gap-2 px-4 py-3 bg-black/20">
    <div class="w-3 h-3 rounded-full bg-red-500"></div>
    <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
    <div class="w-3 h-3 rounded-full bg-green-500"></div>
    <span class="ml-2 text-terminal-text/60 text-sm font-mono">ravact</span>
  </div>
  <div class="p-6 font-mono text-sm text-terminal-text">
    <!-- Terminal content here -->
  </div>
</div>
```

### Feature Card (When Needed)
```html
<div class="group bg-white border border-border rounded-lg p-6 hover:border-primary/40 hover:shadow-sm transition-all">
  <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4 group-hover:bg-primary/20 transition-colors">
    <svg class="w-6 h-6 text-primary">...</svg>
  </div>
  <h3 class="font-semibold text-lg mb-2">Feature Title</h3>
  <p class="text-text-secondary text-sm leading-relaxed">
    Feature description that explains the benefit clearly.
  </p>
</div>
```

### Code Block
```html
<div class="bg-terminal-bg rounded-lg overflow-hidden">
  <div class="flex items-center justify-between px-4 py-2 bg-black/20">
    <span class="text-terminal-text/60 text-xs font-mono">bash</span>
    <button class="text-terminal-text/60 hover:text-terminal-text text-xs">Copy</button>
  </div>
  <pre class="p-4 overflow-x-auto"><code class="text-sm font-mono text-terminal-text">curl -sSL https://... | sudo bash</code></pre>
</div>
```

### Badge/Tag
```html
<span class="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-medium bg-primary/10 text-primary">
  v0.2.1
</span>

<span class="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-medium bg-success/10 text-success">
  Running
</span>
```

---

## Content Sections

### Supported Packages (for icons/grid)

| Package | Description |
|---------|-------------|
| Nginx | High-performance web server |
| MySQL | Popular relational database |
| PostgreSQL | Advanced open-source database |
| Redis | In-memory data store |
| PHP | Server-side scripting language |
| Node.js | JavaScript runtime |
| Git | Version control system |
| Supervisor | Process control system |
| Certbot | Let's Encrypt SSL automation |
| FrankenPHP | Modern PHP application server |
| Dragonfly | Redis-compatible cache |
| UFW | Uncomplicated Firewall |

### Site Templates

| Template | Use Case |
|----------|----------|
| Static HTML | Simple websites |
| PHP | Generic PHP applications |
| Laravel | Laravel framework projects |
| WordPress | WordPress installations |
| Symfony | Symfony framework projects |
| Node.js | Node.js applications |
| Reverse Proxy | Proxy to backend services |

### Developer Toolkit Commands

**Laravel (9 commands)**
- Tail Laravel Log
- Clear Laravel Log
- Find Large Log Files
- Fix Storage Permissions
- Generate APP_KEY
- Check .env File
- List Scheduled Tasks
- Check Queue Workers
- Find Recently Modified Files

**WordPress (9 commands)**
- Fix wp-content Permissions
- Find Large Uploads
- Clear Cache Files
- Generate WP Salts
- Check wp-config.php
- List Plugins
- List Themes
- Check .htaccess
- Find Modified Core Files

**PHP (8 commands)**
- Check PHP Version
- List PHP Modules
- Check PHP Memory Limit
- Check PHP Upload Limits
- Find php.ini Location
- Check OPcache Status
- Test PHP Syntax
- List PHP-FPM Pools

**Security (8 commands)**
- Scan for Malware Patterns
- Find World-Writable Files
- Find World-Writable Dirs
- Check for Suspicious Files
- List Failed SSH Logins
- Check Open Ports
- Check SSL Certificate
- Find SUID Files

### Keyboard Shortcuts Summary

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate |
| `Enter` | Select |
| `Esc` | Go back |
| `q` | Quit |
| `c` | Copy to clipboard |
| `?` | Help (File Browser) |
| `Tab` | Switch categories |
| `Space` | Toggle selection |

---

## SEO & Meta

```html
<title>Ravact - Modern Linux Server Management TUI</title>
<meta name="description" content="Ravact is a powerful terminal user interface for managing Linux servers. Install software, configure services, manage sites, and more with an elegant TUI.">
<meta name="keywords" content="linux, server management, tui, terminal, nginx, mysql, postgresql, redis, devops, sysadmin">

<!-- Open Graph -->
<meta property="og:title" content="Ravact - Command Your Servers Like a King">
<meta property="og:description" content="Modern TUI application for powerful Linux server management">
<meta property="og:image" content="/images/og-image.png">
<meta property="og:url" content="https://ravact.dev">

<!-- Twitter -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Ravact - Modern Linux Server Management TUI">
<meta name="twitter:description" content="Install software, manage services, configure sites—all from a beautiful terminal interface.">
```

---

## Assets Needed

### Images
1. `logo.svg` - Ravact logo (simple, clean)
2. `og-image.png` - Social media preview (1200x630)
3. `favicon.ico` - Browser favicon
4. `screenshot-main.png` - Main menu screenshot
5. `screenshot-setup.png` - Setup menu screenshot
6. `screenshot-toolkit.png` - Developer toolkit screenshot
7. `screenshot-filebrowser.png` - File browser screenshot

### Icons (Use Heroicons or similar)
- Package/Box icon
- Server icon
- Terminal icon
- Code icon
- Folder icon
- Shield/Security icon
- Database icon
- Globe/Web icon
- Users icon
- Lightning bolt icon

---

## Technical Implementation

### Laravel Setup
```bash
# Create Laravel project
composer create-project laravel/laravel ravact-website
cd ravact-website

# Install Tailwind CSS
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p

# Install Inter font
npm install @fontsource/inter
```

### Tailwind Config
```javascript
// tailwind.config.js
module.exports = {
  content: [
    "./resources/**/*.blade.php",
    "./resources/**/*.js",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#FF6B35',
          dark: '#E55A2B',
          light: '#FF8F66',
        },
        secondary: {
          DEFAULT: '#004E89',
          dark: '#003D6B',
          light: '#0066B3',
        },
        success: '#2ECC71',
        warning: '#F39C12',
        error: '#E74C3C',
        info: '#3498DB',
        'text-primary': '#1A1A2E',
        'text-secondary': '#4A4A68',
        'text-muted': '#7F8C8D',
        background: '#FAFBFC',
        surface: '#FFFFFF',
        border: '#E5E7EB',
        terminal: {
          bg: '#1E1E2E',
          text: '#CDD6F4',
          green: '#A6E3A1',
          yellow: '#F9E2AF',
          orange: '#FAB387',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      borderRadius: {
        'sm': '4px',
        'md': '6px',
        'lg': '8px',
      },
    },
  },
  plugins: [],
}
```

---

## Hero Animation Specification

The hero section features an animated terminal that demonstrates Ravact's TUI in action. This is a JavaScript-based animation that simulates the actual application flow.

### Animation Sequence

The animation plays in a loop with the following stages:

#### Stage 1: Splash Screen (3 seconds)
Display the Ravact splash screen with typing effect for the ASCII logo:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│                                                                         │
│     ██████╗  ██████╗ ██╗   ██╗ █████╗  ██████╗████████╗                 │
│     ██╔══██╗██╔═══██╗██║   ██║██╔══██╗██╔════╝╚══██╔══╝                 │
│     ██████╔╝███████║╚██╗ ██╔╝███████║██║        ██║                     │
│     ██╔══██╗██╔══██║ ╚████╔╝ ██╔══██║██║        ██║                     │
│     ██║  ██║██║  ██║  ╚██╔╝  ██║  ██║╚██████╗   ██║                     │
│     ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝   ╚═╝                     │
│                                                                         │
│                    Linux Server Management TUI                          │
│                                                                         │
│              Power and Control for Your Server Infrastructure           │
│                                                                         │
│                    Version 0.2.1 (linux/arm64)                          │
│                                                                         │
│                    Created by Indunil Peramuna                          │
│                 https://github.com/iperamuna/ravact                     │
│                                                                         │
│                      Press any key to continue...                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Stage 2: Simulate Key Press (0.5 seconds)
Show a brief key press indicator (highlight "Press any key..." text)

#### Stage 3: Main Menu (3 seconds)
Transition to main menu with host info displayed:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  Ravact v0.2.1                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  📦 Package Management                                                  │
│       Install Software                                                  │
│       Installed Applications                                            │
│                                                                         │
│  ⚙️  Service Configuration                                              │
│       Service Settings                                                  │
│                                                                         │
│  🌐 Site Management                                                     │
│       Site Commands                                                     │
│       Developer Toolkit                                                 │
│                                                                         │
│  👥 System Administration                                               │
│       User Management                                                   │
│       Quick Commands                                                    │
│                                                                         │
│  🔧 Tools                                                               │
│     ▶ File Browser                                                      │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  Host: ubuntu-server (192.168.1.100)                                    │
│  OS: Ubuntu 24.04 LTS │ Arch: x86_64 │ CPU: 4 cores │ RAM: 8 GB         │
│                                                                         │
│  ↑/↓ Navigate  Enter Select  q Quit                                     │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Stage 4: Navigate to File Browser (1.5 seconds)
Animate cursor moving down to "File Browser" with highlight effect:
- Show cursor moving through menu items (each item briefly highlights as cursor passes)
- Final position: "▶ File Browser" highlighted in orange

#### Stage 5: File Browser - /home (3 seconds)
Transition to File Browser showing /home directory:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  File Browser  ubuntu-server (192.168.1.100)                            │
├─────────────────────────────────────────────────────────────────────────┤
│  → /home                                                                │
│                                                                         │
│  ▶ [ ] ↓ ubuntu                    <DIR>     Jan 25 14:32  drwxr-xr-x   │
│    [ ] ↓ deploy                    <DIR>     Jan 20 09:15  drwxr-xr-x   │
│    [ ] ↓ www-data                  <DIR>     Jan 18 11:20  drwxr-xr-x   │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  3 items | 0 B | Sort: Name↑                                            │
│                                                                         │
│  ↑/↓: Navigate • Enter: Open • Space: Select • ?: Help                  │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Stage 6: Navigate into /home/ubuntu (1 second)
Show Enter key press, then transition to ubuntu's home directory

#### Stage 7: File Browser - /home/ubuntu (4 seconds)
Show ubuntu user's home directory with typical files:

```
┌─────────────────────────────────────────────────────────────────────────┐
│  File Browser  ubuntu-server (192.168.1.100)                            │
├─────────────────────────────────────────────────────────────────────────┤
│  → /home/ubuntu                                                         │
│                                                                         │
│  ▶ [ ] ↓ .config                   <DIR>     Jan 25 10:00  drwxr-xr-x   │
│    [ ] ↓ .ssh                      <DIR>     Jan 15 08:30  drwx------   │
│    [ ] ↓ projects                  <DIR>     Jan 24 16:45  drwxr-xr-x   │
│    [ ] ↓ backups                   <DIR>     Jan 22 03:00  drwxr-xr-x   │
│    [ ] • .bashrc                    3.5 KB   Jan 10 12:00  -rw-r--r--   │
│    [ ] • .profile                   807 B    Jan 10 12:00  -rw-r--r--   │
│    [ ] SH deploy.sh                 2.1 KB   Jan 23 14:20  -rwxr-xr-x   │
│    [ ] TXT notes.txt                  156 B  Jan 25 09:30  -rw-r--r--   │
│                                                                         │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  8 items | 6.5 KB | Sort: Name↑ | Hidden: On                            │
│  ✓ Path copied: /home/ubuntu                                            │
│  ↑/↓: Navigate • Enter: Open • Space: Select • ?: Help                  │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Stage 8: Pause and Loop (2 seconds)
Brief pause showing the file browser, then fade out and restart from Stage 1

### Animation Implementation

```javascript
// Hero Animation Configuration
const heroAnimation = {
  stages: [
    { name: 'splash', duration: 3000, content: splashScreen },
    { name: 'keypress', duration: 500, content: splashScreenHighlight },
    { name: 'mainmenu', duration: 3000, content: mainMenuScreen },
    { name: 'navigate', duration: 1500, content: mainMenuNavigating },
    { name: 'filebrowser-home', duration: 3000, content: fileBrowserHome },
    { name: 'enter-ubuntu', duration: 1000, content: fileBrowserTransition },
    { name: 'filebrowser-ubuntu', duration: 4000, content: fileBrowserUbuntu },
    { name: 'pause', duration: 2000, content: fileBrowserUbuntu }
  ],
  totalDuration: 18000, // 18 seconds per loop
  
  // Color scheme matching the TUI
  colors: {
    background: '#1E1E2E',
    text: '#CDD6F4',
    primary: '#FF6B35',      // Orange - highlights, cursor
    secondary: '#004E89',    // Blue - headers
    success: '#A6E3A1',      // Green - success messages
    warning: '#F9E2AF',      // Yellow - warnings
    muted: '#7F8C8D',        // Gray - descriptions
    border: '#404040',       // Border color
  }
};

// Terminal Component Structure
class TerminalAnimation {
  constructor(container) {
    this.container = container;
    this.currentStage = 0;
    this.isPlaying = true;
  }
  
  render(content) {
    // Render terminal frame with content
    // Use monospace font (JetBrains Mono)
    // Apply syntax highlighting based on content type
  }
  
  typeText(text, speed = 50) {
    // Typing animation effect
  }
  
  moveCursor(fromIndex, toIndex, items) {
    // Animate cursor movement through menu items
  }
  
  transition(fromContent, toContent) {
    // Smooth transition between screens
  }
  
  start() {
    this.playStage(0);
  }
  
  playStage(index) {
    const stage = heroAnimation.stages[index];
    this.render(stage.content);
    
    setTimeout(() => {
      const nextIndex = (index + 1) % heroAnimation.stages.length;
      this.playStage(nextIndex);
    }, stage.duration);
  }
}
```

### CSS Styling for Terminal

```css
.hero-terminal {
  background: #1E1E2E;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  max-width: 800px;
  margin: 0 auto;
}

.hero-terminal__header {
  background: rgba(0, 0, 0, 0.3);
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.hero-terminal__dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.hero-terminal__dot--red { background: #FF5F56; }
.hero-terminal__dot--yellow { background: #FFBD2E; }
.hero-terminal__dot--green { background: #27CA40; }

.hero-terminal__title {
  margin-left: 8px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
  font-family: 'JetBrains Mono', monospace;
}

.hero-terminal__content {
  padding: 20px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: #CDD6F4;
  min-height: 400px;
  white-space: pre;
  overflow: hidden;
}

/* Syntax highlighting classes */
.term-primary { color: #FF6B35; }
.term-secondary { color: #004E89; }
.term-success { color: #A6E3A1; }
.term-warning { color: #F9E2AF; }
.term-error { color: #E74C3C; }
.term-muted { color: #7F8C8D; }
.term-bold { font-weight: bold; }
.term-highlight {
  background: #FF6B35;
  color: #FFFFFF;
  padding: 0 4px;
}

/* Cursor blink animation */
.term-cursor {
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

/* Typing animation */
.typing-effect {
  overflow: hidden;
  white-space: nowrap;
  animation: typing 0.5s steps(40, end);
}

@keyframes typing {
  from { width: 0; }
  to { width: 100%; }
}

/* Fade transition between screens */
.screen-transition {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

### Responsive Considerations

```css
/* Mobile: Show static screenshot instead of animation */
@media (max-width: 768px) {
  .hero-terminal__content {
    font-size: 10px;
    min-height: 300px;
    padding: 12px;
  }
}

@media (max-width: 480px) {
  .hero-terminal {
    display: none; /* Hide on very small screens */
  }
  
  .hero-static-image {
    display: block; /* Show static image instead */
  }
}
```

### Accessibility

- Provide a "Pause Animation" button for users who prefer reduced motion
- Respect `prefers-reduced-motion` media query
- Include alt text describing the animation sequence
- Ensure the terminal content is not read by screen readers (decorative)

```css
@media (prefers-reduced-motion: reduce) {
  .hero-terminal__content * {
    animation: none !important;
    transition: none !important;
  }
}
```

---

## Animations

### Subtle Hover Effects
```css
.hover-lift {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.hover-lift:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
```

### Terminal Typing Animation
```css
@keyframes typing {
  from { width: 0; }
  to { width: 100%; }
}

.typing-effect {
  overflow: hidden;
  white-space: nowrap;
  animation: typing 2s steps(40, end);
}
```

### Fade In on Scroll
```javascript
// Use Intersection Observer for scroll animations
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('fade-in-visible');
    }
  });
}, { threshold: 0.1 });

document.querySelectorAll('.fade-in').forEach(el => observer.observe(el));
```

---

## Summary

This specification provides everything needed to build a professional, modern website for Ravact that:

1. **Reflects the brand** - Powerful, intelligent, elegant (like Ravana)
2. **Uses flat design** - Clean, no excessive shadows or gradients
3. **Has small rounded corners** - Professional, not playful
4. **Varies section layouts** - Mix of open layouts and cards
5. **Showcases features clearly** - Terminal previews, feature highlights
6. **Provides easy installation** - Clear CTAs, copy-paste commands
7. **Is SEO optimized** - Proper meta tags, semantic HTML

The website should feel like a professional tool for professional developers—powerful yet approachable, modern yet reliable.
