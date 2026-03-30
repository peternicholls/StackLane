# 20i Stack - Docker Development Environment

## Overview
A reusable, centralized Docker development stack for PHP projects using:
- **PHP 8.4** with FPM on Alpine Linux and a global `phpunit` binary
- **Nginx** as reverse proxy
- **MariaDB** for database
- **phpMyAdmin** for database management

## Quick Start

### Shell Commands (Recommended)
```bash
# From any project directory:
dcu          # Start stack (uses current directory)
dcd          # Stop stack
20i          # View status
20i-gui      # Interactive menu
```

### Manual Usage
```bash
cd /path/to/your/project
export CODE_DIR=$(pwd)
export COMPOSE_PROJECT_NAME=$(basename $(pwd))
docker compose -f $HOME/docker/20i-stack/docker-compose.yml up -d
```

## Features

✅ **Centralized Stack** - One stack serves any project  
✅ **Project Isolation** - Each project gets isolated containers  
✅ **Environment Variables** - Fully configurable via .env or .20i-local  
✅ **Shell Integration** - Convenient aliases and functions  
✅ **GUI Interface** - Interactive menu system  
✅ **Live Reloading** - Volume mounting for development  

## Access Points

- **Website**: http://localhost (or custom HOST_PORT)
- **phpMyAdmin**: http://localhost:8081
- **Database**: localhost:3306

## Default Credentials

- **MySQL Root**: `root` / `root`
- **MySQL User**: `devuser` / `devpass`
- **Default DB**: `devdb`

## Configuration

### Global Settings (.env.example)
```bash
HOST_PORT=80
PHP_VERSION=8.4
MYSQL_VERSION=10.6
MYSQL_PORT=3306
PMA_PORT=8081
```

### Per-Project Settings (.20i-local)
Create in your project root:
```bash
export HOST_PORT=8080
export MYSQL_DATABASE=myproject_db
export MYSQL_USER=projectuser
export MYSQL_PASSWORD=projectpass
```

## Architecture

- **Nginx (Port 80)**: Front-end web server and reverse proxy
- **Apache/PHP-FPM (Port 9000)**: PHP processing engine
- **MariaDB (Port 3306)**: Database server
- **phpMyAdmin (Port 8081)**: Database management interface

## Files Structure

```
20i-stack/
├── docker/
│   ├── apache/
│   │   ├── Dockerfile          # PHP 8.4 + extensions
│   │   └── php.ini            # PHP configuration
│   └── nginx.conf.tmpl        # Nginx reverse proxy config
├── docker-compose.yml         # Main stack definition
├── 20i-gui                   # Interactive CLI menu
├── .env.example              # Default configuration
└── README.md                 # This file
```

## Shell Integration

Add to your `.zshrc`:

```bash
# 20i stack configuration
STACK_HOME="${STACK_HOME:-$HOME/docker/20i-stack}"

# Functions (see copy of zshrc.txt for full implementations)
20i-up() { ... }     # Start stack
20i-down() { ... }   # Stop stack
20i-status() { ... } # View status
20i-logs() { ... }   # View logs

# Aliases
alias 20i='20i-status'
alias dcu='20i-up'
alias dcd='20i-down'
alias 20i-gui='$STACK_HOME/20i-gui'
```

## Workflow Examples

### Start New Project
```bash
cd /path/to/new-project
dcu                    # Starts stack for this project
# Site available at http://localhost
```

### Switch Projects
```bash
dcd                    # Stop current stack
cd /path/to/other-project
dcu                    # Start stack for new project
```

### Interactive Management
```bash
20i-gui               # Opens menu with start/stop/status/logs options
```

## Troubleshooting

### Port Conflicts
```bash
# Use custom port
export HOST_PORT=8080
dcu
```

### Database Issues
```bash
# Reset database
dcd
docker volume rm $(docker volume ls -q | grep db_data)
dcu
```

### View Logs
```bash
20i-logs              # Follow all logs
20i-gui               # Menu option for specific service logs
```

### PHPUnit
```bash
# Rebuild the PHP image after pulling stack changes
docker compose -f $HOME/docker/20i-stack/docker-compose.yml build apache

# Run PHPUnit inside the PHP container
docker compose -f $HOME/docker/20i-stack/docker-compose.yml exec apache phpunit --version
```

## Requirements

- Docker Desktop for Mac
- Bash/Zsh shell
- Optional: `dialog` package for prettier GUI menus

## License

MIT License - Use freely for development purposes.

---

**Perfect for**: PHP development, Laravel projects, WordPress development, prototyping, and any web project needing a quick, reliable development environment.
