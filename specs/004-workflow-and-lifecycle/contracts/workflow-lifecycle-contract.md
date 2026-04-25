# Workflow And Lifecycle Contract

## Operator Entry Points

- Primary commands in scope:
  - `stacklane up`
  - `stacklane attach`
  - `stacklane status`
  - `stacklane down`
  - multi-project workflows that explicitly exercise `attach`

## Bootstrap Phase Contract

- Stacklane defines one bootstrap phase only: `post-up`.
- Stacklane runs the bootstrap phase only after Stacklane-owned readiness succeeds.
- Stacklane sources bootstrap configuration only from project-root `.env.stacklane`.
- Stacklane runs the bootstrap command inside the `apache` service container.
- If no bootstrap command is configured, Stacklane completes the normal runtime lifecycle without adding implicit framework-specific behavior.

## Failure Contract

- If the bootstrap command fails, Stacklane must:
  - report a named bootstrap lifecycle failure
  - keep that failure distinct from gateway, DNS, and container-health failures
  - roll the current project runtime back
  - leave unrelated attached projects untouched
  - direct the operator to the documented recovery path: fix the project-local bootstrap command or application issue, then rerun `stacklane up`; use `stacklane down` if the operator needs to force a clean stopped state first
- If Stacklane-owned readiness fails before bootstrap starts, Stacklane must report that infrastructure failure under the relevant readiness step rather than as a bootstrap failure.
- If the application remains broken after bootstrap succeeds, Stacklane docs and validation guidance must classify that as application-owned follow-up work unless the defect proves a Stacklane infrastructure issue.

## Configuration Contract

### Canonical config surfaces

- Project-local Stacklane config: project-root `.env.stacklane`
- Stack-owned defaults: `<stack-home>/.env.stacklane`
- Application-owned config: project `.env`
- Machine-generated runtime envfiles: `<stack-home>/.stacklane-state/envfiles/*.env`

### Precedence order

1. CLI flags
2. Project-root `.env.stacklane`
3. Shell environment
4. Stack-home `.env.stacklane`
5. Built-in defaults

Project `.env` is not a generic Stacklane config surface.

## Naming Contract

### Project-scoped runtime names

- Compose project default: `stln-<slug>`
- Runtime network default: `<compose-project>-runtime`
- Database volume default: `<compose-project>-db-data`
- Derived route/upstream names must stay consistent with the project-scoped compose name.

### Shared resources

- Shared-gateway compose project and shared network use `stln-shared`.
- The gateway service network alias on the shared network uses `stln-gateway`.
- Shared resources must remain distinguishable from project-scoped resources in both docs and status output.

### Bootstrap source restriction

- `STACKLANE_POST_UP_COMMAND` is honored only when set in project-root `.env.stacklane`.
- Setting `STACKLANE_POST_UP_COMMAND` in shell environment, stack-home `.env.stacklane`, or project `.env` is silently ignored.
- This is enforced in the config loader, not in the lifecycle orchestrator.

### Bootstrap timeout and cancellation

- There is no separate bootstrap-execution timeout in spec 004. `STACKLANE_WAIT_TIMEOUT` covers Stacklane-owned readiness only.
- Operator-initiated cancellation (`Ctrl-C`) during bootstrap aborts the hook and triggers project rollback under the same `post-up-hook` step name.
- Background or non-foreground operation of `stacklane up` is out of scope for spec 004.

## Validation Contract

- The completion path for this feature must include:
  - one representative single-project validation scenario
  - one representative multi-project validation scenario
  - at least one bootstrap failure and rollback check
- Each validation scenario must check:
  - DNS routing
  - shared gateway readiness
  - runtime env injection
  - database provisioning alignment
  - bootstrap execution outcome
  - status output after success or rollback
  - teardown behavior

- The multi-project validation scenario must execute `attach` explicitly rather than treating it as optional shorthand for startup.

## Documentation Contract

- `README.md` and `docs/runtime-contract.md` must describe the same lifecycle and naming contract.
- Example env files must point operators to `.env.stacklane`.
- If validation runs from a deployed stack copy instead of the repository working copy, operator docs and validation notes must call out the sync point under `$HOME/docker/20i-stack` explicitly.
- Operator-facing docs must not describe `.stackenv` or `stacklane-` as the supported contract once this feature lands.