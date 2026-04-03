**StackLane Modular Architecture Rewrite — Research Report**  
**Branch**: 003-rewrite-language-choices (SHA 8cde1d7)**Codebase snapshot**: lib/stacklane-common.sh (2,213 lines), docker-compose.yml, docker-compose.shared.yml, docker/apache/Dockerfile, docker/nginx.conf.tmpl, legacy archive scripts, and supporting docs.  
  
**1. Current Architecture Map**  
**1.1 Top-Level Entry Points**  

| File | Role |  |  |  |  |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | - | - | - | - |
| stacklane | Primary CLI entry; sources lib/stacklane-common.sh and calls stacklane_main "$@" |  |  |  |  |
| 20i-up, 20i-down, etc. | Deprecated thin shims; each calls twentyi_legacy_forward <action> (common.sh L1612) which re-invokes stacklane_main with the equivalent --flag |  |  |  |  |
| lib/stacklane-common.sh | Single 2,213-line monolith containing every abstraction layer |  |  |  |  |
  
****1.2 Functional Areas Inside stacklane-common.sh****  
The file is logically partitioned by function-name prefix but has no enforced module boundary:  

| Concern | Representative functions | Lines (approx.) |  |  |  |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | - | - | - |
| Bootstrap / defaults | twentyi_init_defaults, twentyi_load_stack_and_project_config | 1–200 |  |  |  |
| Argument parsing | twentyi_parse_initial_args, twentyi_parse_args, twentyi_validate_stacklane_action_selection | ~1750–1860 |  |  |  |
| Config resolution | twentyi_load_env_file, twentyi_finalize_context, twentyi_resolve_docroot, twentyi_resolve_hostname, twentyi_resolve_ports | ~25–380 |  |  |  |
| Port management | twentyi_port_in_use, twentyi_port_reserved, twentyi_find_available_port, twentyi_resolve_shared_gateway_ports, twentyi_validate_requested_ports | ~175–430 |  |  |  |
| State I/O | twentyi_write_state, twentyi_load_state_file, twentyi_remove_state, twentyi_refresh_registry | ~840–970 |  |  |  |
| Docker orchestration | twentyi_compose, twentyi_shared_compose, twentyi_ensure_shared_infra, twentyi_up_like, twentyi_down_like, twentyi_down_all | ~1040–1090, ~1870–2050 |  |  |  |
| Gateway config gen | twentyi_write_gateway_config, twentyi_gateway_route_lines, twentyi_gateway_block_for_route, twentyi_update_gateway_route | ~1100–1380 |  |  |  |
| TLS / cert management | twentyi_ensure_tls_cert, twentyi_tls_available, twentyi_tls_cert_file, twentyi_tls_key_file | ~1390–1500 |  |  |  |
| DNS setup | twentyi_dns_setup, twentyi_dns_status, twentyi_write_dns_support_files, twentyi_dnsmasq_* | ~760–840 |  |  |  |
| Health probes | twentyi_wait_for_gateway_ready, twentyi_wait_for_route_target, twentyi_wait_for_gateway_route | ~1000–1040 |  |  |  |
| Status / monitoring | twentyi_status, twentyi_docker_status, twentyi_registry_drift_status, twentyi_live_container_summary | ~1050, ~2080–2140 |  |  |  |
| Runtime identity | twentyi_capture_runtime_identity, twentyi_reset_runtime_identity, twentyi_validate_runtime_registration | ~800–850 |  |  |  |
| Collision detection | twentyi_validate_collision, twentyi_validate_requested_ports | ~430–580 |  |  |  |
| Logging | twentyi_logs | ~2150 |  |  |  |
  
****1.3 Key Coupling Hotspots****  
**Finding:** The entire runtime is a single flat Bash process. Every function in stacklane-common.sh runs in (or inherits from) the same shell process and shares global variables (PROJECT_SLUG, HOSTNAME, COMPOSE_PROJECT_NAME, etc.). This creates tight coupling at three specific points:  
1. **Config resolution → state I/O**: twentyi_finalize_context (L~~1880) populates dozens of global variables that are then consumed by both twentyi_export_runtime_env (L~~580) and every downstream step. Changing any variable name or adding a new one requires tracking all call sites manually.  
2. **Collision detection restores globals by hand** (twentyi_validate_requested_ports, L~~430–520): to avoid clobbering the current project's variables when scanning other state files, the function manually saves and restores 22 separate global variables. This pattern repeats in twentyi_validate_collision (L~~540–580) and twentyi_status. Any new tracked field requires updating all three save/restore blocks.  
3. **Gateway config generation couples directly to the TSV registry format**: twentyi_gateway_route_lines (L~~1105) reads the registry TSV line-by-line with positional field extraction. The column order is hard-coded and any change to twentyi_refresh_registry (L~~880) that reorders columns silently breaks gateway generation.  
**1.4 Docker Infrastructure**  

| Compose file | Purpose |  |  |  |  |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | - | - | - | - |
| docker-compose.shared.yml | Shared nginx gateway on ports 80/443; uses external Docker network twentyi-shared |  |  |  |  |
| docker-compose.yml | Per-project stack: nginx (static/fastcgi), apache/PHP-FPM, MariaDB, phpMyAdmin; connects to both <slug>-runtime and twentyi-shared networks |  |  |  |  |
  
**Finding:** Docker Compose is used as an imperative execution engine (docker compose up -d) rather than for its declarative reconciliation capabilities. Labels (io.20i.*) are applied to containers for introspection, which is a good pattern but currently only queried by docker ps --filter label=... one container at a time.  
  
**2. Docker Integration Points and Advanced Capabilities**  
**2.1 Current Usage**  
* docker network create / docker network inspect — shared network lifecycle (twentyi_ensure_shared_infra, L~1060)  
* docker compose up/down via docker compose -f ... -p ... — project and gateway lifecycle  
* docker ps --filter label=com.docker.compose.project=... — container introspection  
* docker compose exec -T gateway wget — gateway health probe (twentyi_wait_for_gateway_ready, L~1000)  
* Docker labels (io.20i.project.*, io.20i.service) on containers for identity tracking  
* Docker's embedded DNS resolver 127.0.0.11 is explicitly leveraged in generated nginx configs for dynamic upstream resolution (L~1220–1235) — this is a strong and correct pattern  
**2.2 Underutilised Advanced Capabilities**  
**Recommendation:** A rewrite should leverage these Docker capabilities more deeply rather than reimplementing them in application code:  

| Docker capability | Current gap | Recommended use |  |  |  |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- | - | - | - |
| Docker healthchecks | No HEALTHCHECK in compose files; readiness is polled externally via wget in a retry loop (L~1000–1020) | Define HEALTHCHECK in docker-compose.yml for nginx/apache/mariadb; use docker compose up --wait to block until healthy |  |  |  |
| Compose profiles | Single monolithic docker-compose.yml; phpMyAdmin always starts | Use Compose profiles (--profile debug) to opt-in services like phpMyAdmin |  |  |  |
| Compose depends_on.condition: service_healthy | Not used; post-start probes are bash polling loops | Replace bash polling with depends_on: condition: service_healthy |  |  |  |
| Docker label filtering | Labels exist but queries are per-service and sequential | Use docker ps --filter label=io.20i.project.slug=<slug> for atomic project-wide queries |  |  |  |
| Docker Engine API (HTTP) | Not used; only CLI subprocess calls | For a compiled rewrite, the Docker Engine SDK provides typed access without spawning subprocesses |  |  |  |
| Compose Watch | Not used | Could replace manual volume mounts for hot-reload workflows |  |  |  |
| Network aliases as service discovery | Used correctly for WEB_NETWORK_ALIAS → nginx upstream; Docker DNS 127.0.0.11 used in gateway nginx | Extend: use structured labels + alias conventions for zero-config service discovery |  |  |  |
| Multi-platform images | docker/apache/Dockerfile uses FROM php:${PHP_VERSION}-fpm-alpine without --platform | Add --platform linux/amd64 or use Buildx for consistent cross-platform image builds |  |  |  |
  
**Core principle to encode in the rewrite:** StackLane's job is to be an intelligent orchestration layer *above* Docker, not to re-implement container lifecycle management. Every place where bash is polling, parsing, or guarding what Docker can express declaratively is a rewrite candidate.  
  
**3. Platform-Specific Assumptions and Portability Risks**  
**3.1 Explicit macOS-Only Code**  

| Finding | Location | Risk |  |  |  |
| ---------------------------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | - | - | - |
| [[ "$(uname -s)" != "Darwin" ]] && exit 1 | twentyi_dns_setup (L~800) | --dns-setup fails immediately on Linux/Windows |  |  |  |
| brew --prefix, brew list dnsmasq, brew services restart dnsmasq | twentyi_dnsmasq_conf_dir, twentyi_dns_setup (L~760–840) | Hard dependency on Homebrew; no Linux (systemd-resolved, /etc/NetworkManager) or Windows (WSL2 resolve.conf, Acrylic DNS) equivalent |  |  |  |
| osascript -e "do shell script ... with administrator privileges" | twentyi_dns_setup (L~840) | AppleScript; macOS-only privilege escalation |  |  |  |
| mkcert via Homebrew | twentyi_ensure_tls_cert (L~1410) | Homebrew path assumption; mkcert exists on other platforms but installation differs |  |  |  |
| /etc/resolver/<suffix> | twentyi_dns_resolver_file (L~830) | macOS-specific resolver file system; Linux uses /etc/resolv.conf or systemd-resolved drop-ins; Windows uses registry DNS settings |  |  |  |
  
****3.2 Shell-Specific Assumptions****  

| Finding | Location | Risk |  |  |  |
| ------------------------------------------------------------------ | -------------------------------------------------------------------- | ----------------------------------------------------------------- | - | - | - |
| BASH_SOURCE[0] | twentyi_script_dir (L~8) | Bash 4+; Zsh and POSIX sh differ |  |  |  |
| set -euo pipefail | top of stacklane-common.sh | Not fully POSIX; pipefail behaviour differs between Bash versions |  |  |  |
| printf -v varname for dynamic variable assignment | twentyi_capture_runtime_identity (L~820), twentyi_parse_args | Bash 3.1+; not available in sh/dash |  |  |  |
| ${!var} indirect expansion | twentyi_validate_requested_ports (L~390), twentyi_validate_collision | Bash-only |  |  |  |
| readarray/mapfile not used but array pattern += and "${arr[@]}" is | various | Bash 4+; not POSIX |  |  |  |
| [[ -h "$source_path" ]] | twentyi_script_dir (L~8) | Bash conditional expression, not sh |  |  |  |
  
****3.3 Filesystem Assumptions****  

| Finding | Location | Risk |  |  |  |
| ---------------------------------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | - | - | - |
| Container mount path hardcoded as /home/sites/<slug> | twentyi_finalize_context (L~1920), docker-compose.yml | Assumes POSIX host/container filesystem layout |  |  |  |
| State directory under $STACK_HOME/.20i-state | twentyi_init_defaults (L~1840) | On Windows (native, not WSL), paths with . prefixes in certain tools behave differently |  |  |  |
| cd "$path" && pwd -P for path resolution | twentyi_abs_dir (L~70) | Requires a working pwd -P and the directory to exist; silent failure if directory is missing |  |  |  |
| mktemp "${config_file}.tmp.XXXXXX" | twentyi_write_gateway_config (L~1110) | mktemp flags differ between GNU and BSD; XXXXXX suffix works on both but path-based template may not |  |  |  |
| cp "$preview_resolver" "$resolver_file" | twentyi_dns_setup | Assumes POSIX file operations |  |  |  |
  
****3.4 Networking Assumptions****  

| Finding | Location | Risk |  |  |  |
| ------------------------------------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | - | - | - |
| lsof -nP -iTCP:<port> -sTCP:LISTEN | twentyi_port_in_use (L~195) | lsof is not available by default on many Linux distros; fallback to netstat is attempted but netstat is also absent from minimal systems |  |  |  |
| 127.0.0.11 as Docker DNS | twentyi_gateway_block_for_route (L~1220) | Docker's embedded resolver address is 127.0.0.11 in Linux-based Docker; valid on Docker Desktop for Mac/Windows too |  |  |  |
| Ports 80/443 as default gateway ports | .env.example, twentyi_init_defaults | Requires root or CAP_NET_BIND_SERVICE on Linux for ports below 1024; Docker Desktop on macOS handles this transparently |  |  |  |
| dnsmasq port 53535 | .env.example | High-port workaround for non-root DNS on macOS; Linux would typically use port 53 with proper privilege |  |  |  |
  
****3.5 Process Dependencies****  

| Tool | Context | Portability risk |  |  |  |
| ------------------------------ | --------------------------- | ----------------------------------------------------------------------- | - | - | - |
| docker CLI (compose v2 plugin) | all orchestration | Must be docker compose (not docker-compose); version detection absent |  |  |  |
| brew | DNS setup only | macOS-only |  |  |  |
| lsof, netstat | port checking | Absent on minimal images; ss is the Linux-preferred tool |  |  |  |
| mkcert | TLS cert generation | Requires manual install on Linux; no equivalent Windows path documented |  |  |  |
| openssl x509 | cert expiry display | Availability varies; cosmetic use only |  |  |  |
| sed -E | gateway config gen, slugify | GNU vs BSD sed -E differences (macOS ships BSD sed) |  |  |  |
| tr, awk, grep -E, grep -Fq | various | Generally portable but BSD/GNU flag differences can surface |  |  |  |
  
**4. Recommended Modular Target Architecture**  
**4.1 Design Principles (from Constitution + Product Goal)**  
1. StackLane enhances and orchestrates Docker — it must never replace Docker's own lifecycle or networking capabilities.  
2. The same input produces the same result (Constitution Principle II).  
3. Failures in one project must not corrupt others (Constitution Principle III).  
4. The operator UX stays shell-first and memorisable (Constitution Principle I).  
**4.2 Module Boundaries**  
  
Code  
  
```
stacklane/
├── cmd/                    # CLI entry points & argument parsing
│   └── stacklane/          # single binary, subcommands or flags
├── core/
│   ├── config/             # config loading, precedence chain, validation
│   ├── project/            # project identity: slug, hostname, paths
│   ├── state/              # state read/write, registry, drift detection
│   └── lifecycle/          # up/down/attach/detach orchestration
├── infra/
│   ├── docker/             # Docker Engine SDK client wrapper
│   ├── compose/            # compose file templating and invocation
│   └── gateway/            # nginx config generation, gateway lifecycle
├── platform/
│   ├── dns/                # DNS setup abstraction
│   │   ├── macos.go        # dnsmasq + /etc/resolver
│   │   ├── linux.go        # systemd-resolved / dnsmasq
│   │   └── windows.go      # WSL2 DNS config stub
│   ├── tls/                # mkcert integration
│   └── ports/              # port availability checking (platform-aware)
├── observability/
│   ├── status/             # status rendering, drift reporting
│   └── logs/               # log streaming
└── migration/              # legacy command forwarding, deprecation notices

```
**4.3 Interface Definitions**  
**ConfigProvider** — isolates config resolution from all consumers:  
  
Code  
  
```
type ProjectConfig struct {
    Name, Slug, Dir, Hostname, Suffix  string
    Docroot, ContainerSiteRoot         string
    PHPVersion, MySQLVersion           string
    Ports                              PortSet
    // ...
}

type ConfigLoader interface {
    Load(projectDir string, flags CLIFlags) (ProjectConfig, error)
}

```
Implementation stacks: CLI flags → .20i-local → shell env → .env → defaults. This replaces the 22-variable global save/restore pattern.  
**StateStore** — isolates all file I/O from orchestration logic:  
**StateStore** — isolates all file I/O from orchestration logic:  
  
Code  
  
```
type StateStore interface {
    Save(cfg ProjectConfig, state AttachmentState) error
    Load(slug string) (ProjectConfig, AttachmentState, error)
    List() ([]ProjectRecord, error)
    Delete(slug string) error
    Registry() ([]RegistryRow, error)
}

```
**DockerClient** — wraps the Docker Engine SDK, abstracting away CLI subprocess calls:  
  
Code  
  
```
type DockerClient interface {
    NetworkExists(name string) (bool, error)
    CreateNetwork(name string) error
    ComposeUp(projectName, composeFile string, env []string) error
    ComposeDown(projectName, composeFile string) error
    ListContainersByLabel(labelFilter map[string]string) ([]Container, error)
}

```
**GatewayManager** — owns gateway config generation and reload:  
  
Code  
Code  
  
```
type GatewayManager interface {
    WriteConfig(routes []Route) error
    Reload() error
    Health() (HealthState, error)
}

```
**DNSProvider** — platform-adaptive interface:  
  
Code  
  
```
type DNSProvider interface {
    Bootstrap(suffix, ip string, port int) error
    Status() DNSStatus
}

```
**4.4 Orchestration Layer**  
The lifecycle module coordinates across interfaces. The up flow becomes:  
  
Code  
  
```
1. ConfigLoader.Load(...)           → ProjectConfig
2. StateStore.List()                → port collision check
3. DNSProvider.Status()             → warn if not ready
4. DockerClient.CreateNetwork(...)  → idempotent
5. GatewayManager.WriteConfig(...)  → no-route placeholder
6. DockerClient.ComposeUp(...)      → per-project stack
7. DockerClient.ListContainersByLabel → capture identity
8. StateStore.Save(...)             → write state + registry
9. GatewayManager.WriteConfig(...)  → with new route
10. GatewayManager.Reload()         → hot reload

```
Each step is independently testable; failures at steps 6–9 can roll back via StateStore and GatewayManager without corrupting other projects.  
**4.5 Where Domain Logic Lives**  
* **Project identity and slug derivation**: core/project — pure functions, no I/O, fully unit-testable  
* **Config precedence chain**: core/config — deterministic, no side effects  
* **Port allocation**: platform/ports — platform-aware but injectable for testing  
* **Gateway nginx template**: infra/gateway — use text/template or a Go templating approach; make route structures typed, not string-interpolated TSV  
* **DNS bootstrapping**: platform/dns/<os> — compile-time platform selection via build tags (Go) or #[cfg(target_os)] (Rust)  
* **Docker API calls**: infra/docker — single place that knows about Docker; everything else uses the interface  
  
**5. Safety, Robustness, and Engineering Considerations**  
**5.1 Error Handling**  
**Finding:** The current codebase uses set -euo pipefail with exit 1 on fatal errors. This works for a linear bash script but has limits:  
* || true is scattered throughout to suppress errors from Docker commands that might fail safely (e.g., docker network rm ... || true, L~2025)  
* Post-start validation (twentyi_validate_runtime_registration, L~950) captures container identity but cannot easily roll back the compose up that preceded it  
**Recommendation:** In a compiled language, every operation that can fail should return a typed error. Orchestration steps should be structured as a transaction with an explicit rollback path — if StateStore.Save fails after DockerClient.ComposeUp succeeds, the next stacklane --status should report drift rather than corrupt the registry.  
**5.2 State Management**  
**Finding:** State is stored as printf '%q'-escaped shell files (twentyi_write_state, L~1000). This is readable and simple but:  
* Sourcing these files for every registry refresh is O(n) file reads with shell process spawning overhead at scale  
* The registry TSV at .20i-state/registry.tsv is written atomically (via > "$registry_file" then append), but the projects directory's individual .env files are written with truncate-then-write which is not atomic  
**Recommendation:** Use a single structured file (JSON or TOML) per project state and a single registry file, written via atomic rename (write to temp, rename). In Go this is os.Rename after writing to a temp file in the same directory.  
**5.3 Concurrency / Process Management**  
**Finding:** There is no locking. Two concurrent stacklane --up invocations in different terminals can race on twentyi_find_available_port, both pick the same free port, and both succeed until Docker port binding fails.  
**Recommendation:** Use a filesystem lock (e.g., flock on Linux/macOS, LockFileEx on Windows) or a simple lock file under .20i-state/ as a coarse mutex around port allocation and state writes. A compiled binary can use OS-level advisory locks more cleanly than a shell script.  
**Recommendation:** Use a filesystem lock (e.g., flock on Linux/macOS, LockFileEx on Windows) or a simple lock file under .20i-state/ as a coarse mutex around port allocation and state writes. A compiled binary can use OS-level advisory locks more cleanly than a shell script.  
**5.4 Configuration Validation**  
**Finding:** Validation happens late — after all config is resolved, twentyi_validate_requested_ports and twentyi_validate_collision run, but only for up/attach. Passing an invalid PHP version or non-existent docroot path only fails at Docker build/startup time, not at config parse time.  
**Recommendation:** A ConfigLoader should validate all known fields immediately after resolution and return structured validation errors before any Docker call is made. A --dry-run mode (already implemented in skeleton) should exercise the full validation path and report all errors at once.  
**Recommendation:** A ConfigLoader should validate all known fields immediately after resolution and return structured validation errors before any Docker call is made. A --dry-run mode (already implemented in skeleton) should exercise the full validation path and report all errors at once.  
**5.5 Secrets Handling**  
**Finding:** Database credentials (MYSQL_ROOT_PASSWORD, MYSQL_PASSWORD) default to root/devpass and are passed through environment variables in the compose file. The .env.example explicitly shows MYSQL_ROOT_PASSWORD=root. State files written to .20i-state/projects/*.env include MYSQL_PASSWORD=%q in plaintext.  
**Recommendation:**  
* Use Docker secrets or compose secrets: for credentials in production-adjacent workflows  
* In local dev, accept that credentials are low-security but document the threat model explicitly  
* State files should not store credentials; they should store only the compose project name so credentials can be re-resolved from .20i-local on each invocation  
* Never commit secrets to the registry TSV  
**5.6 Logging and Observability**  
**Finding:** All user-facing output uses printf to stdout/stderr. There is no structured logging, no log levels, and no machine-readable output format. The --status output is human-readable text that would be difficult to parse by tooling.  
**Recommendation:**  
**Recommendation:**  
* Add --json / --quiet output flags for status so the output can be consumed by scripts or dashboards  
* Add a --verbose flag that enables debug-level logging (e.g., the resolved docker compose command before execution)  
* Use structured logging internally (e.g., slog in Go, tracing in Rust) with a human-readable default renderer  
**5.7 Rollback and Recovery**  
**Finding:** twentyi_update_gateway_route (L~1310) does implement a backup-and-restore for the gateway config file — it cps the current config to a .bak file, writes the new config, and restores if the gateway reload fails. This is a good pattern.  
**Recommendation:** Extend this to project state: before writing a new state file, preserve the previous version. Add a stacklane --reset-project <slug> command that removes drift by stopping the docker project and deleting the state file, with user confirmation.  
**Recommendation:** Extend this to project state: before writing a new state file, preserve the previous version. Add a stacklane --reset-project <slug> command that removes drift by stopping the docker project and deleting the state file, with user confirmation.  
**5.8 Deterministic Behavior**  
**Finding:** twentyi_find_available_port (L~225) scans for available ports starting at a configurable base. In a multi-project concurrent startup scenario the same port could be allocated to two projects if both start simultaneously. The check is not atomic.  
**Recommendation:** Port allocation should be serialised through a lock. In the longer term, consider removing per-project port publishing entirely for MariaDB and phpMyAdmin in favour of internal-network-only access via docker exec or a gateway-proxied admin path.  
  
**6. Technical Debt Risks and Mitigation**  
**6.1 Current Debt Inventory**  

| Debt item | Severity | Evidence |  |  |  |
| ---------------------------------------------------- | ------------------ | ---------------------------------------------------------------------------- | - | - | - |
| Single 2,213-line bash file for all logic | High | lib/stacklane-common.sh — no module boundaries possible in bash |  |  |  |
| Global variable state shared between all functions | High | 22-variable save/restore in twentyi_validate_requested_ports (L~430) |  |  |  |
| State files in shell-sourceable format | Medium | twentyi_write_state (L~1000); tightly couples state format to bash |  |  |  |
| TSV registry with positional column order | Medium | twentyi_gateway_route_lines (L~1105); column order change = silent breakage |  |  |  |
| Port allocation race condition | Medium | twentyi_find_available_port (L~225); no lock |  |  |  |
| macOS-only DNS implementation | High (portability) | twentyi_dns_setup (L~800); explicit uname guard |  |  |  |
| Docker MariaDB volumes (not host-bound) | Medium | docker-compose.yml volumes: db_data: named volume; README notes deferred fix |  |  |  |
| phpMyAdmin always started | Low | docker-compose.yml; no profiles: opt-in |  |  |  |
| No config validation before Docker invocation | Medium | Config errors surface at Docker layer, not at parse time |  |  |  |
| Bash version dependency (arrays, printf -v, ${!var}) | Medium | Prevents sh compatibility and complicates cross-platform packaging |  |  |  |
| Credentials in state files | Medium | twentyi_write_state persists MYSQL_PASSWORD |  |  |  |
| No structured logging | Low | All output via printf |  |  |  |
| set -euo pipefail with scattered ` |  | true` |  |  |  |
  
****6.2 Practices to Minimise Future Debt****  
1. **Typed configuration structs** — no global variables; pass config explicitly through function signatures  
2. **Interface-driven design** — platform adapters implement a common interface; swap implementations for testing without mocking the OS  
3. **Atomic state writes** — write-to-temp, rename; never partial-write a state file  
4. **Schema-versioned state files** — include a version field in JSON state; add a migration function for each version increment  
5. **Dependency injection for Docker** — the Docker client is a constructor parameter, enabling unit tests that don't require a real Docker daemon  
6. **Explicit ADR (Architecture Decision Records)** — the specs/ directory already contains good planning artifacts; extend with one ADR per significant design decision  
7. **CI on all platforms** — GitHub Actions matrix across ubuntu-latest, macos-latest, windows-latest; fail early on platform-specific regressions  
8. **Contract tests for state format** — golden file tests that verify state files can be round-tripped across versions  
  
**7. Language/Runtime Choices**  
**7.1 Evaluation Criteria**  
From the product principle: StackLane must be distributable as a standalone tool, use the Docker Engine SDK or CLI, run cross-platform, and be maintainable by a small team.  
**7.2 Comparison**  

| Criterion | Go | Rust | Python | Node.js / TypeScript | Current Bash |
| --------------------------- | --------------------------------- | --------------------------------- | --------------------------------- | ---------------------------- | ----------------------------- |
| Single-binary distribution | ✅ native | ✅ native | ❌ requires runtime or PyInstaller | ❌ requires runtime or pkg | ✅ (just scripts) |
| Cross-platform compile | ✅ GOOS=windows/linux/darwin | ✅ targets require cross toolchain | N/A | N/A | ❌ macOS-only in practice |
| Docker Engine SDK | ✅ docker/docker-client (official) | ⚠️ bollard (active, unofficial) | ✅ docker SDK (official) | ✅ dockerode (community) | ❌ CLI subprocess only |
| Performance | ✅ excellent startup | ✅ excellent | ⚠️ acceptable for CLI | ⚠️ acceptable; slower startup | ✅ instant (no parse overhead) |
| Portability risk | Low | Low | Medium (version conflicts) | Medium (node_modules) | Very high |
| Type safety | ✅ strong | ✅ very strong | ⚠️ optional | ✅ TypeScript | ❌ none |
| Error handling | ✅ explicit error returns | ✅ Result<T,E> | ⚠️ exceptions | ⚠️ exceptions / async errors | ❌ exit codes only |
| Ecosystem for CLI | ✅ cobra, viper, kong | ✅ clap, indicatif | ✅ click, typer | ✅ commander, oclif | N/A |
| Concurrency | ✅ goroutines | ✅ async/threads | ⚠️ GIL limits | ⚠️ event loop | ❌ none |
| Learning curve | Low-medium | High | Low | Low-medium | Already known |
| Developer ergonomics | ✅ good tooling | ⚠️ steep initially | ✅ rapid iteration | ✅ rapid iteration | ✅ immediate |
| Template generation (nginx) | ✅ text/template | ✅ tera/minijinja | ✅ Jinja2 | ✅ Handlebars/nunjucks | ⚠️ heredoc + printf |
| Packaging/installation | ✅ brew install, apt, scoop | ✅ same | ⚠️ pip install or binary | ⚠️ npm global install | git clone + PATH |
  
****7.3 Recommendation****  
**Go** is the strongest choice for this rewrite, for the following reasons:  
1. **Official Docker SDK**: github.com/docker/docker/client is the same SDK Docker itself uses. It provides typed structs for containers, networks, and volumes — no shell subprocess overhead, no output parsing.  
2. **Single binary, all platforms**: GOOS=darwin GOARCH=arm64 go build produces a self-contained binary with zero runtime dependencies. This directly solves the current "clone a bash script and add to PATH" distribution story.  
3. **Strong error handling**: Go's error return convention combined with errors.Is/errors.As enables structured error chains that map cleanly to the failure modes catalogued above.  
4. **Goroutines for concurrent operations**: gateway reload, health probes, and multi-project teardown can be parallelised naturally.  
5. **Standard library sufficiency**: text/template for nginx config generation, encoding/json for state files, os for atomic file operations, net for port checks.  
6. **Homebrew tap distribution**: Go binaries are straightforward to distribute via a Homebrew tap or GitHub Releases, which fits the macOS-first current audience.  
**If the team prefers scripting ergonomics during transition:** a **hybrid approach** is viable — rewrite the core modules in Go as a library/binary, but keep a thin bash wrapper for the initial migration period that calls the Go binary for heavy operations. The current legacy wrapper pattern already establishes this separation.  
**Rust** is a worthy alternative if memory safety and maximum performance are priorities, but the learning curve and the unofficial Docker SDK (bollard) add risk for a small team. Revisit after Go proves out the architecture.  
**Rust** is a worthy alternative if memory safety and maximum performance are priorities, but the learning curve and the unofficial Docker SDK (bollard) add risk for a small team. Revisit after Go proves out the architecture.  
**Python** and **Node.js** are unsuitable as primary choices because they require a runtime to be installed, complicating the distribution story on platforms where they are not standard.  
**Python** and **Node.js** are unsuitable as primary choices because they require a runtime to be installed, complicating the distribution story on platforms where they are not standard.  
  
**8. Migration Strategy**  
**8.1 Guiding Constraint**  
The current stacklane binary must continue to work throughout the migration. Operators running stacklane --up today must be able to run the same command against the new binary on day one of each phase.  
**8.2 Phase Plan**  
**Phase M1 — Extract and test the config resolution layer**  
* Implement core/config in Go: ConfigLoader that replicates the exact precedence chain (twentyi_finalize_context, L~1875–1955)  
* Write golden-file tests against a set of .20i-local fixtures derived from the current bash behaviour  
* No user-visible change; no bash files modified  
* Gate: 100% test coverage for config resolution edge cases documented in twentyi_finalize_context  
**Phase M2 — Implement the state store**  
* Implement core/state with JSON state files; write a migration reader that can load the existing shell-format state files and re-serialize to JSON  
* The migration reader runs once on first invocation of the new binary; existing .20i-state/ is transparently upgraded  
* Gate: stacklane --status produces identical output from both old bash and new binary against the same state directory  
**Phase M3 — Docker adapter and port allocation**  
* Implement infra/docker using the Docker Engine SDK  
* Implement platform/ports with OS-specific port checks (ss/lsof/netstat per platform)  
* Gate: stacklane --up --dry-run and stacklane --down --dry-run produce the correct Docker commands without executing them  
**Phase M4 — Gateway config generation**  
* Implement infra/gateway with typed Route structs replacing TSV positional parsing  
* Replace the bash twentyi_write_gateway_config / twentyi_gateway_block_for_route functions  
* Gate: generated nginx configs are byte-for-byte identical to current bash output for the same route set  
**Phase M5 — DNS platform adapters**  
* Implement platform/dns/macos first (feature parity with current bash)  
* Implement platform/dns/linux (systemd-resolved or dnsmasq)  
* Implement platform/dns/windows (WSL2 stub initially)  
* Gate: stacklane --dns-setup passes all existing DNS status checks on macOS; Linux stub returns unsupported-platform with actionable message  
**Phase M6 — Full orchestration and binary release**  
* Wire all modules into cmd/stacklane; the new Go binary is now the canonical implementation  
* Bash stacklane script becomes a thin wrapper: exec /usr/local/bin/stacklane-bin "$@" (or distributes the binary directly)  
* Remove lib/stacklane-common.sh in the next release; existing 20i-* wrappers forward to the new binary  
* Gate: all existing checkpoint validations from docs/plan.md pass against the new binary  
**Phase M7 — Debt cleanup**  
* Remove legacy 20i-* wrappers (end of migration window)  
* Move MariaDB to host-bound mounts (as noted in README and constitution)  
* Add Compose healthchecks and depends_on.condition: service_healthy  
* Add --json output to --status  
* Gate: clean stacklane --status --json | jq output; CI passes on Linux, macOS, Windows  
**8.3 Minimising Disruption**  
* State file format migration is automatic and silent; the old format is always readable as a fallback  
* The registry TSV remains for the migration window so existing bash scripts that parse it continue to work  
* Every new binary release includes a compatibility test against a fixture set derived from the current bash state format  
* The --dry-run flag is tested in every phase to preserve the debugging escape hatch  
  
**Summary of Key Findings vs Recommendations**  

| Area | Finding | Recommendation |  |  |  |
| ---------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------- | - | - | - |
| Architecture | Single 2,213-line bash monolith with global state | Modular typed architecture with explicit interfaces and no global state |  |  |  |
| Docker usage | CLI subprocess calls; no Docker healthchecks; no SDK | Use Docker Engine SDK; add HEALTHCHECK + depends_on.condition: service_healthy |  |  |  |
| Platform support | Hard macOS-only DNS; Homebrew and AppleScript dependencies | Platform adapter pattern with per-OS implementations; Go build tags |  |  |  |
| State management | Shell-sourceable files; non-atomic writes | JSON state files; atomic write-rename |  |  |  |
| Concurrency | No locking; port allocation race | File-level advisory lock for port allocation and state writes |  |  |  |
| Secrets | Credentials persisted in state files | Remove credentials from state; re-resolve per invocation from .20i-local |  |  |  |
| Language | Bash | Go (primary recommendation): official Docker SDK, single binary, cross-platform |  |  |  |
| Migration | Big-bang risky | Phase-by-phase with feature parity gates and state-format auto-migration |  |  |  |
| Technical debt | TSV with positional columns; manual 22-var save/restore | Typed structs; column-independent queries; dependency injection |  |  |  |
  
You asked  
save the report as a mark down file in //Users/peternicholls/Dev/stacklane/specs/003-rewrite-language-choices  
