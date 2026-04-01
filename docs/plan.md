## Plan: Multi-site Attachable Stack

Refactor the current localhost-centric workflow into a shared front-door model: one persistent gateway and local DNS layer in front of isolated per-project runtimes. `20i-up` stays the main entrypoint, but under the hood it becomes "ensure shared infra exists, start this project, register its hostname". `20i-attach` and `20i-detach` then manage additional repos against that same shared layer.

I’m recommending `.test` for the first stage, not `.dev`. You said `.dev` is preferred only if it stays low-friction, and on macOS `.dev` becomes awkward fast because of HSTS and the implied need for local TLS.

## Outcome

- `20i-up` works from a project root and exposes that project at a stable local hostname.
- Multiple projects can coexist concurrently through one shared gateway and DNS layer.
- `20i-attach` and `20i-detach` manage project registration without breaking other attached projects.
- Monitoring reports both Docker state and logical attachment state.

## Principles

- Keep the current shell-first workflow intact.
- Prefer `.test` for the first shipping version.
- Reuse Nginx and the existing Docker-based approach rather than swapping in a new router stack.
- Keep project runtimes isolated, especially databases.
- Extend `.20i-local` instead of inventing a second project config format.

## Stories

### Story 1: Shared runtime bootstrap

As a developer, I want `20i-up` to ensure the shared gateway and DNS layer exist so I do not need to manually prepare the environment before starting a project.

### Story 2: Project-specific hostname

As a developer, I want each repo to resolve to its own hostname so I can work on more than one site without relying on generic `localhost` routing.

### Story 3: Concurrent attachment

As a developer, I want to attach a second repo while the stack is already running so I can keep multiple sites live at the same time.

### Story 4: Safe detach

As a developer, I want to detach one project without disturbing others so I can stop work on one site without tearing down the whole environment.

### Story 5: Reliable visibility

As a developer, I want status commands to show what is attached, where it lives, and whether routing is healthy so operational state is obvious.

### Story 6: Low-friction migration

As a developer, I want the new behavior to preserve existing `20i-up` and `20i-down` muscle memory so the migration from the current workflow is predictable.

## Task List

### Phase 1: Runtime contract and CLI semantics

- [x] Define exact command semantics for `20i-up`, `20i-attach`, `20i-detach`, `20i-down`, and global teardown.
- [x] Decide the canonical hostname derivation rule: folder name by default, `.20i-local` override when set.
- [x] Define the first-stage suffix as `.test` and record `.dev` as a later HTTPS-capable option.
- [x] Extend `.20i-local` contract with site name override, document root override, PHP version override, and project database settings.
- [x] Define the expected state transitions for attached, detached, down, and global teardown.
- [x] Document backward-compatible behavior for running `20i-up` in a single project with no other attachments.

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
- [x] Update `20i-up` to write registration state and validate it after startup.
- [x] Implement `20i-attach` as attach-or-bootstrap behavior.
- [x] Implement `20i-detach` to remove routing and stop only the targeted project runtime.
- [x] Update `20i-down` to remain project-local by default.
- [x] Add explicit global teardown behavior such as `20i-down --all`.

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

- [ ] Update README examples away from `localhost` toward project hostnames.
- [ ] Document `.20i-local` additions and override precedence.
- [ ] Add docs for attach, detach, shared teardown, and concurrent project workflows.
- [ ] Add a migration section explaining old versus new behavior.
- [ ] Mark GUI support as deferred or partial if CLI ships first.
- [ ] Tidy up the project structure and docs to reflect the new multi-project focus and follow good practices and patterns for project organization.

## Gates

### Gate A: Contract locked

Pass criteria:

- [x] Command semantics are documented and unambiguous.
- [x] Config precedence is defined.
- [x] Suffix choice for stage one is fixed.
- [x] No unresolved disagreement remains on project isolation model.

### Gate B: Shared infrastructure viable

Pass criteria:

- [ ] Shared gateway starts independently of any one project.
- [ ] A single project can be started behind the gateway without using `localhost`.
- [ ] No per-project web host port is required for normal access.

### Gate C: Multi-project runtime proven

Pass criteria:

- [ ] Two projects can run concurrently.
- [ ] Each project has distinct routing and isolated runtime state.
- [ ] Detaching one project does not interrupt the other.

### Gate D: Operational visibility complete

Pass criteria:

- [ ] Status shows gateway, DNS, attachments, and drift.
- [ ] Failure modes are visible without manually inspecting raw Docker output.
- [ ] Logs can be scoped to an attached project reliably.

### Gate E: Ready to adopt

Pass criteria:

- [ ] Core docs are updated.
- [ ] Single-project workflow still feels familiar.
- [ ] A clean-machine bootstrap path is documented and validated on macOS.

## Checkpoints

### Checkpoint 1: Single-project parity

- [ ] From a clean state, run `20i-up` in one repo.
- [ ] Confirm shared services bootstrap automatically.
- [ ] Confirm the project is reachable at its hostname, not `localhost`.
- [ ] Confirm database connectivity and existing dev workflow still work.

### Checkpoint 2: Concurrent attachment

- [ ] Run `20i-attach` in a second repo.
- [ ] Confirm both sites stay reachable simultaneously.
- [ ] Confirm project A and project B route to the correct mounted codebases.
- [ ] Confirm both projects preserve isolated database state.

### Checkpoint 3: Safe detach and local down

- [ ] Run `20i-detach` in one repo.
- [ ] Confirm its hostname stops resolving or routing.
- [ ] Confirm the other project stays healthy.
- [ ] Run `20i-down` from the remaining repo and confirm only that project stops.

### Checkpoint 4: Global teardown and recovery

- [ ] Run the global teardown command.
- [ ] Confirm all shared infrastructure and registrations are removed cleanly.
- [ ] Re-run `20i-up` and confirm the environment can rebuild from scratch.
- [ ] Reattach a previously used repo and confirm its project database persists correctly.

### Checkpoint 5: Failure-path validation

- [ ] Validate behavior when the DNS service is unavailable.
- [ ] Validate behavior when registry state and Docker state diverge.
- [ ] Validate behavior when one project runtime fails while others remain healthy.
- [ ] Validate behavior when a hostname collision is attempted.

**Relevant files**

- `/Users/peternicholls/Dev/20i-stack/docker-compose.yml` — current all-in-one runtime definition that needs to be split conceptually into shared infra and project-scoped runtime.
- `/Users/peternicholls/Dev/20i-stack/docker/nginx.conf.tmpl` — current single-site `localhost` routing template to evolve into hostname-aware behavior.
- `/Users/peternicholls/Dev/20i-stack/20i-gui` — current command semantics and status patterns to extend with attach/detach and registry-backed monitoring.
- `/Users/peternicholls/Dev/20i-stack/20i-stack-manager.scpt` — macOS automation entrypoint to keep aligned with revised command behavior.
- `/Users/peternicholls/Dev/20i-stack/README.md` — current user contract still describing localhost and one-project switching.
- `/Users/peternicholls/Dev/20i-stack/AUTOMATION-README.md` — automation docs that currently assume stop/start project switching.
- `/Users/peternicholls/Dev/20i-stack/.env.example` — environment contract to update for shared-layer and project-layer settings.
- `/Users/peternicholls/Dev/20i-stack/GUI-HELP.md` — help text that still assumes one active project at a time.

**Verification**

1. From a clean state, bootstrap the local DNS setup and verify wildcard resolution before any project is attached.
2. Run `20i-up` in one repo and confirm it is reachable by hostname rather than `localhost`.
3. Run `20i-attach` in a second repo and confirm both sites remain reachable concurrently.
4. Run monitoring/status and confirm it reports attached repo path, hostname, container health, and DNS/gateway health together.
5. Run `20i-detach` in one repo and verify only that project disappears while the other stays live.
6. Run the global teardown path and verify shared infra and registrations are removed cleanly.
7. Reattach a previously detached project and verify its database data remains isolated and intact.

**Decisions**

- Included now: CLI/runtime architecture, attach/detach semantics, shared gateway, local DNS integration, monitoring/status output, and shell docs.
- Excluded unless you want them pulled in now: full GUI parity, local TLS/cert management for `.dev`, and a full redesign of database admin UX.
- Recommended hostname policy: folder name by default, override via `.20i-local`.
- Recommended suffix policy: ship `.test` first, leave `.dev` for a later HTTPS-capable phase.

## Recommended delivery order

1. Lock command semantics and config contract.
2. Split shared gateway from project runtime.
3. Make one project work behind hostname-based routing.
4. Add project registry and attach/detach flows.
5. Add DNS bootstrap and health reporting.
6. Prove multi-project behavior.
7. Finish monitoring and docs.
