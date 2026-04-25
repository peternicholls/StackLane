# Quickstart: Workflow And Lifecycle Validation

## Goal

Validate the 004 lifecycle contract end to end: project-local bootstrap after readiness, rollback on bootstrap failure, explicit failure classification, `.env.stacklane` as the stack-owned defaults file, and shortened project-scoped runtime naming.

## Preconditions

- macOS with Docker Desktop available
- If the repository working copy is not the copy you run, sync the relevant changes into the deployed stack copy under `$HOME/docker/20i-stack` before validation
- `stacklane-bin` rebuilt from the current branch
- Local DNS already bootstrapped with `stacklane dns-setup`
- One representative application that requires bootstrap and one additional project for multi-project validation

## Configuration Setup

1. Create or update the stack-owned defaults file as `.env.stacklane`.
2. Confirm project-local runtime settings live in `.stacklane-local`.
3. If the representative app needs bootstrap, set `STACKLANE_POST_UP_COMMAND` in `.stacklane-local`.
4. Confirm project `.env` remains application-owned rather than a generic Stacklane config file.

## Happy-Path Validation

1. From the representative application directory, run `stacklane up`.
2. Confirm Stacklane reports shared-gateway and project readiness first, then completes the bootstrap command.
3. Run `stacklane status`.
4. Confirm DNS routing resolves the project hostname to the local stack as documented.
5. Confirm the reported hostname, routes, runtime details, and Docker identities match the running project.
6. Confirm runtime env injection and database provisioning alignment by checking the app or container environment for the expected `DB_*` / `MYSQL_*` values used by the representative app.
7. Inspect Docker resources and confirm project-scoped runtime names use the `stln-` prefix while shared infrastructure remains `stacklane-shared`.
8. Open the project route and confirm the app reaches the expected post-bootstrap state.

## Bootstrap Failure Validation

1. Change `STACKLANE_POST_UP_COMMAND` to a command that fails deterministically.
2. Run `stacklane up` again.
3. Confirm Stacklane reports a named bootstrap lifecycle failure.
4. Run `stacklane status`.
5. Confirm the project was rolled back and no phantom running state remains.
6. Confirm unrelated attached projects, if any, retain their routes, recorded state, and reported attachment unchanged.
7. Fix the project-local bootstrap command or underlying application issue, then rerun `stacklane up`; if a forced clean stop is needed first, run `stacklane down` before retrying.

## Multi-Project Validation

1. Start the first project with `stacklane up`.
2. From the second project, run `stacklane attach` explicitly.
3. Confirm both routes work through the shared gateway.
4. Confirm shared-gateway readiness is healthy while both projects are attached.
5. Confirm Docker resource names leave enough room to distinguish the project slugs in listings.
6. Run `stacklane status` and verify both projects report the correct hostnames, routes, and recorded identities.

## Teardown Validation

1. Run `stacklane down` in the representative application.
2. Confirm only that project stops.
3. If a second project remains attached, confirm its route still works.
4. Run `stacklane status` and confirm the recorded state matches reality.

## Documentation Validation

1. Review `README.md` and `docs/runtime-contract.md`.
2. Confirm they both point operators to `.env.stacklane` as the stack-owned defaults file.
3. Confirm they both describe `stln-` as the project-scoped runtime prefix.
4. Confirm they both describe bootstrap failure as a rollback-triggering lifecycle failure.

## Validation Notes

- Record the representative applications used.
- Record whether the bootstrap command exercised migration-only behavior or a broader setup command.
- Record whether validation ran from the repository working copy or from the deployed copy under `$HOME/docker/20i-stack`.
- Record any real-daemon gap that was not rerun during implementation.