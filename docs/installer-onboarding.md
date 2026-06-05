# StageServe Installer And Guided Onboarding

This guide covers the supported install path, the guided first run, and the direct-command fallback for automation or power users.

## Install The Binary

Recommended install path:

```bash
curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

The installer:

1. Detects your OS and CPU architecture.
2. Downloads the matching `stage_<version>_<OS>_<arch>` release asset.
3. Verifies the matching SHA-256 checksum.
4. Downloads and extracts `stageserve-runtime_<version>.tar.gz` into `$HOME/docker/stageserve` by default.
5. Installs `stage` into `~/.local/bin` by default.
6. Warns when the install directory is not on `PATH`.
7. Hands interactive terminals to bare `stage`.
8. Prints explicit `stage setup`, `stage init`, and `stage up` steps for non-interactive installs.

Non-interactive install:

```bash
NONINTERACTIVE=1 curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

Custom install directory:

```bash
STAGESERVE_INSTALL_DIR="$HOME/bin" curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

Custom runtime asset home:

```bash
STAGESERVE_STACK_HOME="$HOME/docker/stageserve" curl -fsSL https://raw.githubusercontent.com/peternicholls/StageServe/master/install.sh | bash
```

## Guided First Run

From a project root:

```bash
cd /path/to/project
stage
```

What bare `stage` does depends on local context:

- If the machine is not ready, it opens the readiness flow and shows exact fixes.
- If the project has no `.env.stageserve`, it opens the guided project-settings form and previews the file before writing.
- If the project is configured but stopped, it offers the safe default action to run the project.
- If the project is already running, it keeps the default action non-destructive and makes logs, status, stop, and recovery available.

When you do not want the TUI, use the same planner in text form:

```bash
stage --notui
stage --cli
STAGESERVE_NO_TUI=1 stage
```

## Direct Command Path

Use direct commands when you already know the step you want:

```bash
stage setup
stage init
stage up
stage status
stage logs apache
stage down
stage doctor
```

Direct commands keep their normal help, flags, and automation behavior. `stage --help` and `stage <command> --help` do not enter the guided flow.

## Project Settings File

StageServe uses the same filename in two user-owned scopes:

- `<project>/.env.stageserve` for project-local settings.
- `<stack-home>/.env.stageserve` for stack-wide defaults.

The guided first run and `stage init` create or preview the project-local file. If you skip that and run `stage up` or `stage attach` first, StageServe assumes defaults for that run and writes a starter project file afterward.

If the file already exists, interactive `stage init` reopens the guided settings form and asks for confirmation before updating it. For automation or scripts, `stage init --force` remains the explicit overwrite path.

Machine-generated runtime env files live under `.stageserve-state/envfiles/` and are not user-editable inputs.

## DNS And TLS Choices

- `.test` is the default zero-friction hostname suffix.
- `.develop` is the preferred explicit local DNS example. Use `stage dns-setup --site-suffix develop` on macOS when you want the documented routed-hostname path.
- `.dev` enables local TLS on `https://<hostname>:8443` and requires `mkcert` plus local DNS readiness. `stage setup` and `stage doctor` check the prerequisites, and `stage up` or `stage attach` refresh the shared gateway certificate bundle.

## Build From Source

If you are working from a local checkout instead of an installed release artifact:

```bash
git clone https://github.com/peternicholls/StageServe.git ~/docker/stage
cd ~/docker/stage
make build
export STACK_HOME="$HOME/docker/stage"
export PATH="$STACK_HOME:$PATH"
```

Then move to a project root and run `stage` for the guided path, or use the direct commands listed above.