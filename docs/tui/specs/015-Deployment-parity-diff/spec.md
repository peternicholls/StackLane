# Feature Specification: Deployment Parity & Known Differences (Docs-First)

**Feature Branch**: `015-deployment-parity-known-differences`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟡 High  
**Input**: User description: "Document what this Docker stack replicates from 20i shared hosting, what it does not replicate, and how to validate so deployments are not surprised"

## Product Contract *(mandatory)*

This feature is **documentation-only** in the MVP.

- The 20i Docker stack aims for **development parity**, not a perfect production clone.
- The documentation MUST be explicit about **what matches**, **what differs**, and **what to test**.
- The project MUST NOT claim full production equivalence.

### Scope

- Applies to local development using the 20i Docker stack.
- Describes differences versus typical 20i shared hosting behaviour.
- Provides actionable validation guidance.

### Non-goals

- This feature does NOT add new runtime components (e.g. mail server) in MVP.
- This feature does NOT introduce HTTPS/TLS in MVP.
- This feature does NOT attempt full OS or kernel parity.
- This feature does NOT replace staging or real production testing.

## Parity Promise *(what we replicate)*

The stack SHOULD replicate these aspects reliably:

- Core topology: **Nginx + PHP-FPM + MariaDB + phpMyAdmin**
- PHP execution model (FPM behind Nginx)
- Project entry point conventions (e.g. `public_html/`)
- Local environment variable patterns (`.env`, `.20i-local`, etc. where supported)
- Typical request routing and basic headers through Nginx
- Database connectivity patterns (localhost port mapping, container network connectivity)

## Known Differences *(what we do not replicate)*

This section lists common differences and why they matter.

### 1) HTTPS / TLS

**Difference**: Local stack runs HTTP by default. 20i production deployments commonly terminate TLS at the provider edge or via hosting configuration.

**20i notes**:

- 20i provides free SSL certificates via the My20i control panel (SSL/TLS section).
- 20i provides a “Force HTTPS” option to redirect visitors to HTTPS once SSL is active.

**Why it matters**:

- Apps may behave differently when `HTTPS` is present (redirects, cookies, secure session flags, mixed-content rules).
- Reverse proxy headers (`X-Forwarded-Proto`) often change app behaviour.

**How to validate**:

- Ensure your app respects proxy headers and does not hardcode HTTP assumptions.
- Test secure cookies and redirect logic in a production-like mode (see Optional Emulation section).

**Risk**: High (common cause of “works locally, breaks in prod”).

### 1a) Reverse proxy headers and client IP

**Difference**: In production, requests may pass through TLS termination and proxies before reaching PHP, which can affect scheme (`http`/`https`) and client IP reporting.

**20i notes**:

- 20i documentation does not clearly specify which proxy headers are set in front of hosted sites (e.g. `X-Forwarded-Proto`, `X-Forwarded-For`). Treat this as environment-dependent.

**Why it matters**:

- Apps that build absolute URLs, enforce HTTPS, or apply security rules based on client IP can behave differently behind a proxy.
- Incorrect proxy trust configuration can cause security issues (spoofed headers) or broken redirects.

**How to validate**:

- Ensure your framework is configured to trust only known proxies and to interpret `X-Forwarded-Proto` / `X-Forwarded-For` safely.
- Verify what your app receives in production for scheme and client IP (e.g. log `REMOTE_ADDR` and relevant headers) and adjust your trust proxy settings accordingly.

**Risk**: High (commonly causes redirect loops and incorrect IP-based logic).

### 2) Email delivery

**Difference**: No real mail server is provided by default in the dev stack.

**20i notes**:

- If you host email with 20i, outbound SMTP settings use `smtp.stackmail.com` on port `465` (SSL) or `587` (TLS/STARTTLS).
- Many applications (e.g. WordPress/Laravel) should be configured to use SMTP rather than relying on a local sendmail binary.
- 20i imposes sending limits from web servers (PHP `mail()`), including a daily message cap and a per-message size limit.
- If mail sending fails or is rate-limited, production behaviour may differ from local dev unless your app uses SMTP explicitly.

**Why it matters**:

- Password reset flows, sign-up verification, and notification emails may fail silently if mail isn’t configured.
- Apps may assume `sendmail`/SMTP exists.
- Provider limits and anti-abuse controls can cause intermittent failures; relying on `mail()` without visibility makes this hard to diagnose.

**How to validate**:

- Configure your application to use a dev mail sink or mock transport.
- Ensure mail failures are visible in logs.

**Risk**: Medium to High (app dependent).

### 3) Shared hosting restrictions

**Difference**: Shared hosting may enforce limits that do not exist locally (CPU, memory, execution time, concurrent processes).

**20i notes**:

- Scheduled tasks (“cron jobs”) are configured through the My20i control panel under Scheduled Tasks.
- Some platform-level restrictions can disable mail/cron for safety in specific cases (e.g. compromised sites); do not assume these facilities are always available.

**Why it matters**:

- Long-running requests, heavy image processing, and background tasks can succeed locally but fail in production due to resource limits.
- Shared hosting constraints can require different strategies (queues, chunking, async processing).

**How to validate**:

- Verify your app behaves correctly under realistic PHP limits (timeouts, memory cap) and hosting constraints.
- Prefer background jobs for heavy work and ensure user-facing requests stay fast.

**Risk**: High (performance and reliability).

### 3a) PHP defaults and configurable limits

**Difference**: Local PHP settings may not match the defaults (or configured limits) in 20i hosting.

**20i notes**:

- 20i documents a default PHP maximum upload file size of `128M` (configurable per package via PHP Configuration).
- 20i documents a standard PHP memory limit of `128M` (configurable; higher values are possible but uncommon for typical sites).
- 20i allows editing PHP configuration per hosting package, including `max_execution_time` (script time limit).

**Why it matters**:

- Upload-heavy workflows can pass locally but fail in production due to `upload_max_filesize` / `post_max_size`.
- Memory-sensitive tasks (image manipulation, large imports, Composer operations) can behave differently under tighter limits.
- If `max_execution_time` differs, long operations (imports, backups, Composer tasks) may terminate unexpectedly in production.

**How to validate**:

- Set local limits to match your target hosting profile when testing deployment-critical flows.
- Verify your app fails gracefully (clear error paths) when limits are exceeded.
- Check and document the configured `max_execution_time` for your target 20i package and mirror it locally when testing critical flows.

**Risk**: Medium to High (workload dependent).

### 3b) Disabled PHP functions and security hardening

**Difference**: Production hosting can disable certain PHP functions for security reasons.

**20i notes**:

- 20i documents a set of PHP functions disabled on WordPress hosting for security reasons (e.g. `exec`, `shell_exec`, `system`, `proc_open`).

**Why it matters**:

- Apps or plugins that rely on shell execution or process control may work locally but fail in production.

**How to validate**:

- Avoid relying on disabled functions for core functionality.
- If you must use system calls, verify your target hosting plan permits them.

**Risk**: Medium to High (application dependent).

### 4) Filesystem permissions and ownership

**Difference**: Docker volume mounts (especially on macOS) do not perfectly emulate shared hosting file ownership and permission constraints.

**Why it matters**:

- Uploads, cache writes, session files, and framework storage directories may behave differently.

**How to validate**:

- Ensure your application writes only to intended writable directories.
- Confirm framework storage and cache directories are writable in production constraints.

**Risk**: High.

### 5) CPU architecture differences (ARM vs x86_64)

**Difference**: On Apple Silicon, containers may run on ARM64. Production may be x86_64.

**Why it matters**:

- Native extensions and certain binaries can behave differently.
- Multi-arch image availability may affect performance.

**How to validate**:

- Prefer multi-arch images.
- If using native dependencies, confirm they exist for both architectures.

**Risk**: Medium.

### 6) Environment separation and secrets

**Difference**: Local `.env` patterns differ from production secret management.

**Why it matters**:

- Apps can accidentally rely on local-only environment files.
- Secrets handling differs by hosting provider.

**How to validate**:

- Treat `.env` files as development conveniences.
- Ensure production configuration can be supplied without changing the stack.

**Risk**: High.

### 7) Background jobs, cron, and long-running workers

**Difference**: Local dev may support long-running workers freely; shared hosting may constrain background processes.

**Why it matters**:

- Laravel queues, schedulers, and workers need deployment-appropriate strategies.

**How to validate**:

- Document how your app schedules tasks on 20i (cron vs HTTP triggers).
- Ensure job queues degrade gracefully.

**Risk**: Medium to High.

### 8) Database version and defaults

**Difference**: MariaDB version, charset, collation, and SQL modes may differ.

**Why it matters**:

- Queries may behave differently; migrations can fail.

**How to validate**:

- Pin MariaDB version where possible.
- Test migrations and seed scripts against the production target version.

**Risk**: Medium.

## Risk Summary *(quick guide)*

- **High risk**: HTTPS assumptions, filesystem permissions, shared hosting limits, secrets/environment separation
- **Medium risk**: Email delivery, background workers, CPU architecture, DB defaults
- **Low risk**: Cosmetic header differences, minor Nginx config variance

## Recommended Validation Checklist *(before deploy)*

1. Run the app in a production-like mode (e.g. `APP_ENV=production`) locally.
2. Confirm cookies, redirects, and session behaviour do not require real HTTPS locally.
3. Verify framework storage/cache directories are writable and only expected paths are used.
4. Confirm email flows are testable via a dev mail sink or mocked transport.
5. Run database migrations against the target MariaDB version and settings.
6. Validate any scheduled tasks and background jobs have a viable production strategy.
7. Review PHP limits (memory, execution time) and ensure heavy work is not request-bound.

## Optional Emulation (Documentation Only)

These are suggested approaches a user may adopt; they are not implemented by the stack in MVP:

- Local TLS termination via a reverse proxy (self-signed cert) to test HTTPS-only logic.
- A local mail sink service (e.g. a lightweight SMTP catcher) to observe emails.
- Production-like PHP ini limits to surface timeouts and memory issues.

## Implementation Notes *(future)*

Future enhancements MAY provide:

- A documented “prod-like” toggle that adjusts only safe, deterministic settings.
- Optional services via Compose profiles (see spec 007) to add a mail sink or TLS proxy.

These MUST remain opt-in and MUST not compromise the simplicity of the core stack.

## Success Criteria *(mandatory)*

- **SC-001**: Users can identify and understand the top causes of local-vs-prod differences.
- **SC-002**: Documentation provides actionable steps to validate common failure modes.
- **SC-003**: The project does not claim full parity and avoids encouraging unsafe assumptions.

---
## References (20i documentation)

- SSL and Force HTTPS: My20i SSL/TLS tools and Force HTTPS option
- Cron / Scheduled Tasks: My20i Scheduled Tasks documentation
- PHP configuration (upload size): PHP max upload file size documentation
- PHP configuration (memory limit): PHP memory limit guidance
- SMTP settings (20i mail): Email SMTP settings documentation
- Email sending limits (PHP mail): Web server daily and message size limits
- PHP configuration: `max_execution_time` and related limits
- Disabled PHP functions (WordPress hosting): security hardening list