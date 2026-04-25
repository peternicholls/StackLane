## Plan: Multi-site Attachable Stack

Refactor the current localhost-centric workflow into a shared front-door model: one persistent gateway and local DNS layer in front of isolated per-project runtimes. `stacklane` is the canonical entrypoint; `stacklane up` becomes "ensure shared infra exists, start this project, register its hostname". `stacklane attach` and `stacklane detach` then manage additional repos against that same shared layer.

I’m recommending `.test` for the first stage, not `.dev`. You said `.dev` is preferred only if it stays low-friction, and on macOS `.dev` becomes awkward fast because of HSTS and the implied need for local TLS.

## Outcome

- `stacklane up` works from a project root and exposes that project at a stable local hostname.
- Multiple projects can coexist concurrently through one shared gateway and DNS layer.
- `stacklane attach` and `stacklane detach` manage project registration without breaking other attached projects.
- Monitoring reports both Docker state and logical attachment state.

## Principles

- Optimize for ease of use in the common shell-first workflow.
- Make repeated runs reliable and predictable across the CLI entry point.
- Keep the runtime robust under partial failure, drift, and multi-project isolation pressure.
- Remove pinch points and user friction before adding optional complexity.

## Stories

### Story 1: Shared runtime bootstrap

As a developer, I want `stacklane up` to ensure the shared gateway and DNS layer exist so I do not need to manually prepare the environment before starting a project.

### Story 2: Project-specific hostname

As a developer, I want each repo to resolve to its own hostname so I can work on more than one site without relying on generic `localhost` routing.

### Story 3: Concurrent attachment

As a developer, I want to attach a second repo while the stack is already running so I can keep multiple sites live at the same time.

### Story 4: Safe detach

As a developer, I want to detach one project without disturbing others so I can stop work on one site without tearing down the whole environment.

### Story 5: Reliable visibility

As a developer, I want status commands to show what is attached, where it lives, and whether routing is healthy so operational state is obvious.

### Story 6: Low-friction current workflow

As a developer, I want the current `stacklane <subcommand>` workflow to be predictable and documented.

## Task List

### Phase 1: Runtime contract and CLI semantics

- [x] Define exact command semantics for `stacklane up`, `stacklane attach`, `stacklane detach`, `stacklane down`, and global teardown.
- [x] Decide the canonical hostname derivation rule: folder name by default, project-root `.env.stacklane` override when set.
- [x] Define the first-stage suffix as `.test` and record `.dev` as a later HTTPS-capable option.
- [x] Extend the project-root `.env.stacklane` contract with site name override, document root override, PHP version override, and project database settings.
- [x] Define the expected state transitions for attached, detached, down, and global teardown.
- [x] Document behavior for running `stacklane up` in a single project with no other attachments.

### Phase 2: Shared infrastructure split

- [x] Separate shared services from project-scoped services in the compose model.
- [x] Move host ports `80/443` to a shared gateway layer.
- [x] Create a shared Docker network for gateway-to-project routing.
- [x] Remove direct host web port publishing from normal per-project runtimes.
- [x] Decide whether phpMyAdmin is deferred, centralized, or exposed per project in the first milestone.
- [x] Make sure the shared layer can start once and remain stable while projects are added and removed.

### Phase 3: Project runtime isolation

- [x] Namespace project containers, volumes, and networks so multiple repos can coexist cleanly.
- [x] Keep code mounting project-specific and preserve current `CODE_DIR` behavior.
- [x] Add document root override support so projects are not forced to use only one fixed layout.
- [x] Move database state to project-scoped storage and ensure no cross-project leakage.
- [x] Confirm PHP version override remains project-specific.
- [x] Define how project runtime names map back to repo paths and hostnames.

### Phase 4: Registry and orchestration

- [x] Add a registry/state file under the stack home to record attachments.
- [x] Store repo path, project name, hostname, document root, runtime settings, and live container identity.
- [x] Update `stacklane up` to write registration state and validate it after startup.
- [x] Implement `stacklane attach` as attach-or-bootstrap behavior.
- [x] Implement `stacklane detach` to remove routing and stop only the targeted project runtime.
- [x] Update `stacklane down` to remain project-local by default.
- [x] Add explicit global teardown behavior such as `stacklane down --all`.

### Phase 5: Gateway routing

- [x] Replace single-site `localhost` routing with hostname-aware gateway configuration.
- [x] Generate or template route definitions from the registry.
- [x] Reload the gateway safely after attach and detach operations.
- [x] Validate that one bad project registration cannot break routing for all attached projects.
- [x] Ensure the gateway can surface a clear error when a project runtime is down but still registered.

### Phase 6: Local DNS service integration

- [x] Choose the concrete local DNS service implementation for macOS.
- [x] Add bootstrap/setup logic for the DNS service and resolver configuration.
- [x] Support wildcard resolution for the chosen suffix.
- [x] Add health checks so CLI status can report DNS readiness.
- [x] Add failure handling for missing resolver setup, missing privileges, or stopped DNS service.

### Phase 7: Monitoring and status

- [x] Update status output to show shared gateway health.
- [x] Show local DNS health separately from Docker health.
- [x] Show attached project name, repo path, hostname, document root, and project container state.
- [x] Detect and report drift between registry state and live Docker state.
- [x] Make logs and status project-aware rather than only compose-project-aware.

### Phase 8: Documentation and migration

- [x] Update README examples away from `localhost` toward project hostnames.
- [x] Document project-root `.env.stacklane` additions and override precedence.
- [x] Add docs for attach, detach, shared teardown, and concurrent project workflows.
- [x] Add a migration section explaining old versus new behavior.
- [x] Mark GUI support as deferred or partial if CLI ships first.
- [x] Tidy up the project structure and docs to reflect the new multi-project focus and follow good practices and patterns for project organization.

## Gates

### Gate A: Contract locked

Pass criteria:

- [x] Command semantics are documented and unambiguous.
- [x] Config precedence is defined.
- [x] Suffix choice for stage one is fixed.
- [x] No unresolved disagreement remains on project isolation model.

### Gate B: Shared infrastructure viable

Pass criteria:

- [x] Shared gateway starts independently of any one project.
- [x] A single project can be started behind the gateway without using `localhost`.
- [x] No per-project web host port is required for normal access.

### Gate C: Multi-project runtime proven

Pass criteria:

- [x] Two projects can run concurrently.
- [x] Each project has distinct routing and isolated runtime state.
- [x] Detaching one project does not interrupt the other.

### Gate D: Operational visibility complete

Pass criteria:

- [x] Status shows gateway, DNS, attachments, and drift.
- [x] Failure modes are visible without manually inspecting raw Docker output.
- [x] Logs can be scoped to an attached project reliably.

### Gate E: Ready to adopt

Pass criteria:

- [x] Core docs are updated.
- [x] Single-project workflow still feels familiar.
- [x] A clean-machine bootstrap path is documented and validated on macOS.

## Checkpoints

### Checkpoint 1: Single-project validation

- [x] From a clean state, run `stacklane up` in one repo.
- [x] Confirm shared services bootstrap automatically.
- [x] Confirm the project is reachable at its hostname, not `localhost`.
- [x] Confirm database connectivity and existing dev workflow still work.

### Checkpoint 2: Concurrent attachment

- [x] Run `stacklane attach` in a second repo.
- [x] Confirm both sites stay reachable simultaneously.
- [x] Confirm project A and project B route to the correct mounted codebases.
- [x] Confirm both projects preserve isolated database state.

### Checkpoint 3: Safe detach and local down

- [x] Run `stacklane detach` in one repo.
- [x] Confirm its hostname stops resolving or routing.
- [x] Confirm the other project stays healthy.
- [ ] Run `stacklane down` from the remaining repo and confirm only that project stops.

### Checkpoint 4: Global teardown and recovery

- [x] Run the global teardown command.
- [x] Confirm all shared infrastructure and registrations are removed cleanly.
- [ ] Re-run `stacklane up` and confirm the environment can rebuild from scratch.
- [ ] Reattach a previously used repo and confirm its project database persists correctly.

### Checkpoint 5: Failure-path validation

- [x] Validate behavior when the DNS service is unavailable.
- [x] Validate behavior when registry state and Docker state diverge.
- [ ] Validate behavior when one project runtime fails while others remain healthy.
- [ ] Validate behavior when a hostname collision is attempted.

**Relevant files**

- `docker-compose.yml` — current all-in-one runtime definition that needs to be split conceptually into shared infra and project-scoped runtime.
- `docker/nginx.conf.tmpl` — current single-site `localhost` routing template to evolve into hostname-aware behavior.
- `previous-version-archive/legacy GUI script` — legacy command semantics and status patterns extended with attach/detach and registry-backed monitoring.
- `previous-version-archive/` archived AppleScript entrypoint — kept aligned with revised command behavior in `previous-version-archive/`.
- `README.md` — current user contract describing localhost and one-project switching.
- `previous-version-archive/AUTOMATION-README.md` — automation docs that currently assume stop/start project switching.
- `.env.example` — environment contract to update for shared-layer and project-layer settings.
- `previous-version-archive/GUI-HELP.md` — help text that still assumes one active project at a time (archived).

**Verification**

1. From a clean state, bootstrap the local DNS setup and verify wildcard resolution before any project is attached.
2. Run `stacklane up` in one repo and confirm it is reachable by hostname rather than `localhost`.
3. Run `stacklane attach` in a second repo and confirm both sites remain reachable concurrently.
4. Run monitoring/status and confirm it reports attached repo path, hostname, container health, and DNS/gateway health together.
5. Run `stacklane detach` in one repo and verify only that project disappears while the other stays live.
6. Run the global teardown path and verify shared infra and registrations are removed cleanly.
7. Reattach a previously detached project and verify its database data remains isolated and intact.

**Decisions**

- Included now: CLI/runtime architecture, attach/detach semantics, shared gateway, local DNS integration, monitoring/status output, and shell docs.
- Excluded unless you want them pulled in now: full GUI parity, local TLS/cert management for `.dev`, and a full redesign of database admin UX.
- Recommended hostname policy: folder name by default, override via project-root `.env.stacklane`.
- Recommended suffix policy: ship `.test` first, leave `.dev` for a later HTTPS-capable phase.

## Recommended delivery order

1. Lock command semantics and config contract.
2. Split shared gateway from project runtime.
3. Make one project work behind hostname-based routing.
4. Add project registry and attach/detach flows.
5. Add DNS bootstrap and health reporting.
6. Prove multi-project behavior.
7. Finish monitoring and docs.
