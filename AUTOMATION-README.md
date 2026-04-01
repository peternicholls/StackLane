# 🚀 20i Stack Manager - macOS Automation

This automation provides experimental GUI interfaces to manage your 20i Docker stack on macOS. The CLI is now the primary interface for the implemented Phase 1 and Phase 2 project contract.

## 📱 What You Get

### 1. **20i Stack Manager.app** 
- **Location**: `/Users/peternicholls/docker/20i-stack/20i Stack Manager.app`
- **Usage**: Double-click to launch
- **Features**: 
  - 🚀 Start Stack (with folder picker and settings dialog)
  - 🛑 Stop Stack (with project selector)
  - 📊 View Status (shows running containers)
  - 📋 View Logs (follow logs in Terminal)

### 2. **Services Menu Integration**
- **Access**: Right-click anywhere → Services → "20i Stack Manager"
- **Usage**: Available system-wide in any application
- **Same features** as the standalone app

## 🎯 How It Works

### Starting a Stack:
1. **Select Project Folder**: Choose your project directory
2. **Optional Settings**: Set custom environment variables (e.g., `HOST_PORT=8080`)
3. **Auto-Detection**: Project name is automatically detected from folder name
4. **Terminal Launch**: Opens Terminal and runs the docker compose commands

### Smart Features:
- ✅ **Auto-detects running projects** for stop and logs operations
- ✅ **Proper environment isolation** using `COMPOSE_PROJECT_NAME`
- ✅ **Visual feedback** with notifications and dialogs
- ✅ **Terminal integration** for full command visibility
- ⚠️ **CLI leads GUI** for the shared gateway, attach, detach, retained state, and planned hostname reporting

## 🛠 Installation

The automation is already set up! Here's what was installed:

```bash
# Standalone App (ready to use)
~/docker/20i-stack/20i Stack Manager.app

# Services Menu (system-wide access)
~/Library/Services/20i Stack Manager.workflow
```

## 🚀 Quick Start

1. **Double-click** `20i Stack Manager.app`
2. **Choose "🚀 Start Stack"**
3. **Select your project folder**
4. **Optionally configure settings** (or just click "Skip")
5. **Watch Terminal** as your stack starts
6. **Access your site** at the URL printed in Terminal. The CLI now uses the shared gateway port rather than a per-project web port.

## 💡 Pro Tips

- **Services Menu**: Access from any app via right-click → Services
- **Multiple Projects**: Prefer the CLI for concurrent project workflows until GUI parity lands
- **Custom Ports**: The CLI owns the shared gateway web ports; GUI port overrides still follow the older direct-compose flow
- **Logs**: Use "📋 View Logs" to debug issues
- **Quick Stop**: The stop dialog shows only running projects

## 🔧 Environment Variables

You can set these in the settings dialog:

```bash
HOST_PORT=8080          # Custom web port
MYSQL_PORT=3307         # Custom database port  
PMA_PORT=8082          # Custom phpMyAdmin port
MYSQL_DATABASE=mydb    # Custom database name
```

## 🎨 Example Workflow

1. **Start a project**:
   - Start stack → Select `/path/to/project-a` → site becomes available through the shared gateway at `project-a.test`

2. **Add a concurrent project**:
   - The GUI does not yet expose `20i-attach`. Use the CLI: `cd /path/to/project-b && 20i-attach`
   - Both sites then run simultaneously. The GUI's stop and start actions operate on one project at a time and do not disrupt the other.

3. **Debug Issues**:
   - View Status → See all containers
   - View Logs → Follow real-time logs for a selected project

For full multi-project workflows, concurrent attach/detach, and global teardown see [docs/migration.md](docs/migration.md) and use the CLI commands directly.

Use the automation as a convenience layer, not the source of truth for the new runtime contract.
