# StageServe Installer Research Report (April 2026)

## Goal
> Detailed implementation specification: `specs/005-installer-and-onboarding/spec.md`

Define best-practice installer patterns, runtime setup behavior, and first-run onboarding for StageServe (a Docker-based local development CLI for shared-hosting emulation).

## Current StageServe Context
StageServe currently expects users to:
- install Docker Desktop,
- build or download `stage-bin`,
- put `stage` on `PATH`, and
- run `stage dns-setup` once on macOS.

This is simple for power users, but still leaves friction around prerequisites, permissions, and first-run confidence.

---

## Industry Patterns From Comparable Tools

### 1) Prefer package-managed installs and keep “manual binary” as fallback
**Observed:**
- Docker Compose docs explicitly recommend Docker Desktop as the easiest path and mark standalone installs as legacy/back-compat.
- DDEV calls Homebrew the easiest and most reliable install route.

**Pattern:**
- One primary install channel per OS (least surprise, easiest updates).
- Secondary channel for air-gapped/pinned/manual environments.

**Why this works:**
- Less installer logic in your project.
- Better upgrade/uninstall behavior via native package managers.

Sources:
- Docker Compose install overview: <https://docs.docker.com/compose/install/>
- DDEV installation docs: <https://docs.ddev.com/en/stable/users/install/ddev-installation/>

### 2) Be explicit about prerequisite support windows and machine requirements
**Observed:**
- Docker Desktop Mac docs define supported macOS window (“current and two previous major releases”), RAM minimums, and architecture-specific installs.

**Pattern:**
- Publish exact support matrix and enforce it in preflight checks.
- Fail fast with clear remediation when requirements are not met.

Source:
- Docker Desktop on Mac: <https://docs.docker.com/desktop/setup/install/mac-install/>

### 3) Use least-privilege by default and isolate privileged actions
**Observed:**
- Docker documents unprivileged run behavior with only limited privileged configuration points.
- StageServe’s current DNS flow already prints explicit `sudo` follow-up where needed.

**Pattern:**
- Keep install/start flow non-root whenever possible.
- For privileged steps, isolate to small explicit commands and explain why.

Source:
- Docker Desktop macOS permissions: <https://docs.docker.com/desktop/setup/install/mac-permission-requirements/>

### 4) Make onboarding a guided sequence with machine-readable checkpoints
**Observed:**
- AWS CLI onboarding is wizard-driven (`aws configure`, `aws configure sso`) and standardizes credential/config setup.
- GitHub CLI guides users through `gh auth login` flow and handles storage fallback behavior.

**Pattern:**
- Provide a guided `stage setup` wizard with deterministic steps.
- Persist setup state and allow safe reruns.

Sources:
- AWS CLI quickstart/setup: <https://docs.aws.amazon.com/cli/latest/userguide/getting-started-quickstart.html>
- GitHub CLI auth login: <https://cli.github.com/manual/gh_auth_login>

### 5) Treat trust and supply-chain integrity as first-class installer concerns
**Observed:**
- Terraform provides explicit checksum/signature verification workflow for binary archives.

**Pattern:**
- For direct-download installers, publish checksums/signatures and verify in scripted install path.

Source:
- Terraform archive verification: <https://developer.hashicorp.com/terraform/tutorials/cli/verify-archive>

### 6) Provide post-install “caveats” at install time, not buried in docs
**Observed:**
- Homebrew casks have a `caveats` pattern specifically to show install-time guidance (PATH, permissions, extra one-time steps).

**Pattern:**
- Ensure install output gives short, actionable next steps, especially for DNS/TLS and first project startup.

Source:
- Homebrew Cask Cookbook (`caveats`): <https://docs.brew.sh/Cask-Cookbook>

### 7) Accept ecosystem variability instead of assuming one provider
**Observed:**
- DDEV supports multiple Docker providers and gives tradeoff notes for each.

**Pattern:**
- Treat Docker provider as detectable capability rather than hard-coded assumption.
- Validate context and compatibility at runtime.

Source:
- DDEV Docker provider installation: <https://docs.ddev.com/en/stable/users/install/docker-installation/>

---

## Recommended Installer Behavior for StageServe

### A. Add first-class setup and project initialization commands
Introduce:
```bash
stage setup
stage init   # optional, run inside a project root
```

Proposed flow:
1. **Preflight**: detect OS/arch, Docker availability, Docker daemon readiness, and version constraints.
2. **Path/installation check**: validate executable location and active version.
3. **Project init handoff**: if in a repo without `.env.stageserve`, recommend `stage init`.
4. **DNS step**: run/check `dns-setup` for supported platforms.
5. **Optional TLS step (future `.dev`)**: mkcert presence + CA install checks.
6. **Health probes**: shared network/gateway prereq checks.
7. **Summary + next command**: print exactly what to run next (`stage init` if needed, then `stage up`).

### B. Make setup idempotent and resumable
- Every step should return one of: `ready`, `needs_action`, `error`.
- Re-running `stage setup` should skip already-complete steps.
- Add `--json` for automation and GUI integration.

### C. Package strategy
- **Primary (macOS):** Homebrew formula/cask (recommended path).
- **Secondary:** signed release binary + checksum verification helper script.
- Use installer output caveats to show post-install one-time commands.

### D. Improve failure UX
For each failed check, print:
- what failed,
- why it matters,
- exact fix command(s),
- recheck command.

Example:
```text
✖ Docker daemon not reachable
  Why: StageServe requires a running Docker provider.
  Fix: Open Docker Desktop (or start your configured provider), then run:
       stage setup
```

### E. Align privileges with principle of least astonishment
- Keep standard install/start non-privileged.
- For privileged DNS operations, continue explicit human-confirmed boundary.
- Never silently escalate privileges.

### F. Add clear support policy and compatibility matrix
Publish in README/docs and enforce in preflight:
- supported macOS versions,
- supported Docker providers/versions,
- architecture support (Intel/Apple Silicon),
- unsupported/experimental modes.

### G. Add post-install verification contract
Ship:
```bash
stage doctor
```
Checks should include:
- binary + version,
- Docker context/daemon,
- local DNS status,
- shared gateway health,
- writable state directories,
- port conflicts.

---

## Suggested Onboarding Design (User Journey)

### Stage 0: Install
- `brew install stage` (or signed binary script).
- Installer caveat: “Run `stage setup` next”.

### Stage 1: Machine setup
- `stage setup` runs full wizard.
- Outputs pass/fail per step.

### Stage 2: Project initialization and first attach
- From repo root (optional but recommended): `stage init` to generate/validate project `.env.stageserve`.
- Then run: `stage up`.
- `init` should avoid overwriting existing project config unless explicitly requested.

### Stage 3: Confidence checks
- Auto-run quick checks and print route URL.
- Offer `stage status` and `stage logs` tips.

### Stage 4: Recovery/repair
- `stage doctor` identifies drift and suggests one-command remediations.

---

## Prioritized Implementation Roadmap

### Phase 1 (High impact, low complexity)
1. `stage setup` command with preflight + DNS status.
2. Structured step results (`ready/needs_action/error`).
3. Better remediation messaging in setup and `dns-setup`.

### Phase 2 (Distribution hardening)
1. Homebrew-based install path.
2. Signed checksums for release binaries.
3. Install-time caveat text and docs alignment.

### Phase 3 (Operational maturity)
1. `stage doctor` diagnostics.
2. `--json` output for setup/doctor.
3. Optional non-interactive mode for CI/bootstrap scripts.

### Phase 4 (Future `.dev` + TLS onboarding)
1. mkcert detection/install guidance.
2. Browser trust checks where feasible.
3. Port collision and cert expiry diagnostics.

---

## Concrete Recommendations for This Repo (Next 2 Sprints)

1. Add `stage setup` and reuse existing DNS provider status/bootstrap calls.
2. Add optional `stage init` for project-root scaffolding and app-to-stack connection defaults.
3. Add a `setup/init` section in README as the canonical first-run path (`install -> setup -> init -> up`).
4. Define and document a minimum support matrix (OS + Docker provider versions).
5. Add a structured diagnostics command (`doctor`) before expanding feature surface.
6. Add release integrity docs: checksums, optional signature verification script.

If StageServe does only one installer improvement now, it should be **`stage setup` with idempotent checks + targeted remediation**. That closes the largest onboarding gap while fitting the current CLI-first architecture.
