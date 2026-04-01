# 20i Stack GUI Manager

The GUI is still experimental and currently trails the CLI. Use the shell commands for the implemented Phase 1 and Phase 2 multi-project contract.

## 🚀 Usage

From any project directory, simply run:

```bash
20i-gui
```

This currently gives you an interactive menu with these options:

### 📋 Menu Options:

1. **🚀 Start Stack (current directory)**
   - Uses the current directory as your project root
   - Still follows the older direct compose path
   - Does not yet surface the shared gateway, attach, and detach semantics from the new CLI contract

2. **🛑 Stop Stack**
   - Stops the selected compose project
   - Does not retain the richer attachment state that the CLI now tracks

3. **📊 View Status**
   - Shows Docker-oriented status only
   - Does not yet report shared gateway health, planned hostnames, project docroots, or attachment state

4. **📋 View Logs**
   - Shows running 20i stacks
   - Follow real-time logs for selected project
   - Press Ctrl+C to stop following

## 🎯 Current Use Cases:

- **Basic project switching** while CLI remains the authoritative workflow
- **Lightweight inspection** of what is running
- **Trying the experimental menu flow** if you do not need attach or detach yet

## 🛠 Integration with Existing Workflow:

Recommended command-line surface:
- `20i-up` - Start and attach the current project
- `20i-attach` - Attach an additional project concurrently
- `20i-down` - Stop the current project and retain state
- `20i-detach` - Stop the current project and remove its state
- `20i-status` - Show attachment state, hostname, and Docker status
- `20i-dns-setup` - One-time local DNS bootstrap (run once per machine)

For a full workflow walk-through including concurrent projects and migration from the old model, see [docs/migration.md](docs/migration.md).

> **Note on `20i-gui-depricated`**: The `20i-gui-depricated` script in the repo root is the original pre-shared-gateway GUI wrapper. It is kept for historical reference but does not integrate with the shared gateway, hostname routing, or the project registry. Prefer the CLI commands above.

## 💡 Pro Tips:

- **Dialog Support**: Install `dialog` package for prettier menus:
  ```bash
  brew install dialog
  ```

- **Project Settings**: Create `.20i-local` in your project root:

   ```bash
   export SITE_NAME=my-site
   export DOCROOT=public_html
   export PHP_VERSION=8.4
   export MYSQL_DATABASE=myproject_db
   ```

- **One-off CLI override**: Start with a different PHP version without editing project config:

   ```bash
   20i-up --php-version 8.4
   20i-up version=8.4
   ```

- **From Anywhere**: The `20i-gui` command works from any project directory

Use it as a secondary option alongside the shell workflow while the GUI remains partial.
