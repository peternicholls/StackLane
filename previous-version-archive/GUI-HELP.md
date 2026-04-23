# Stacklane GUI Manager

The GUI is still experimental and trails the CLI. Use `stacklane` for the implemented multi-project contract.

## 🚀 Usage

Use the installed app or Services workflow for the current GUI path:

```bash
open "$HOME/docker/20i-stack/Stacklane Manager.app"
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
   - Shows running Stacklane projects
   - Follow real-time logs for selected project
   - Press Ctrl+C to stop following

## 🎯 Current Use Cases:

- **Basic project switching** while CLI remains the authoritative workflow
- **Lightweight inspection** of what is running
- **Trying the experimental menu flow** if you do not need attach or detach yet

## 🛠 Integration with Existing Workflow:

Recommended command-line surface:
- `stacklane --up` - Start and attach the current project
- `stacklane --attach` - Attach an additional project concurrently
- `stacklane --down` - Stop the current project and retain state
- `stacklane --detach` - Stop the current project and remove its state
- `stacklane --status` - Show attachment state, hostname, and Docker status
- `stacklane --dns-setup` - One-time local DNS bootstrap (run once per machine)

For a full workflow walk-through including concurrent projects and migration from the old model, see [docs/migration.md](docs/migration.md).

If your live stack runs from a deployed copy such as `/Users/peternicholls/docker/20i-stack`, sync changes from your dev workspace before using the GUI wrappers. The GUI launches whatever copy is on disk at runtime.

> **Note on `legacy GUI wrapper`**: The `legacy GUI wrapper` script in the repo root is the original pre-shared-gateway GUI wrapper. It is kept for historical reference but does not integrate with the shared gateway, hostname routing, or the project registry. Prefer `stacklane` instead.

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
   stacklane --up --php-version 8.4
   stacklane --up version=8.4
   ```

- **From Anywhere**: The installed app and Services workflow can be launched outside the repo root

Legacy `legacy wrapper commands` wrappers still forward during migration, but the primary shell workflow is now `stacklane --action`.

Use it as a secondary option alongside the shell workflow while the GUI remains partial.
