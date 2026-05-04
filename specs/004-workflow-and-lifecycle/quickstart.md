# Quickstart: Workflow And Lifecycle Validation

## Goal

Validate the 004 lifecycle contract end to end: project-local bootstrap after readiness, rollback on bootstrap failure, explicit failure classification, `.env.stageserve` as the stack-owned defaults file, and shortened project-scoped runtime naming.

## Preconditions

- macOS with Docker Desktop available
- If the repository working copy is not the copy you run, sync the relevant changes into the deployed stack copy under `$HOME/docker/20i-stack` before validation
- `stage-bin` rebuilt from the current branch
- Local DNS already bootstrapped with `stage dns-setup --site-suffix develop`
- One representative application that requires bootstrap and one additional project for multi-project validation

## Configuration Setup

1. Create or update the stack-owned defaults file as `<stack-home>/.env.stageserve`.
2. Confirm project-local runtime settings live in project-root `.env.stageserve`.
3. If the representative app needs bootstrap, set `STAGESERVE_POST_UP_COMMAND` in project-root `.env.stageserve`.
4. Confirm project `.env` remains application-owned rather than a generic StageServe config file.

## Happy-Path Validation

1. From the representative application directory, run `stage up`.
2. Confirm StageServe reports shared-gateway and project readiness first, then completes the bootstrap command.
3. Run `stage status`.
4. Confirm DNS routing resolves the project hostname to the local stack as documented.
5. Confirm the reported hostname, routes, runtime details, and Docker identities match the running project.
6. Confirm runtime env injection and database provisioning alignment by checking the app or container environment for the expected `DB_*` / `MYSQL_*` values used by the representative app.
7. Inspect Docker resources and confirm both project-scoped and shared StageServe-owned runtime names use the `stage-` prefix, including the shared compose project/network (`stage-shared`) and gateway service alias (`stage-gateway`).
8. Open the project route and confirm the app reaches the expected post-bootstrap state.
9. When validating in Safari or VS Code Simple Browser, enter the full URL with scheme, for example `http://my-project.develop/`, rather than only the hostname.

## Bootstrap Failure Validation

1. Change `STAGESERVE_POST_UP_COMMAND` in project-root `.env.stageserve` only — setting it via shell env, stack-home `.env.stageserve`, or project `.env` is silently ignored — to a command that fails deterministically.
2. Run `stage up` again.
3. Confirm StageServe reports a named bootstrap lifecycle failure under the `post-up-hook` step.
4. Run `stage status`.
5. Confirm the project was rolled back, no record is left as `attached`, and no phantom running state remains.
6. Confirm unrelated attached projects, if any, retain their routes, recorded state, and reported attachment unchanged.
7. Optionally repeat the test using `Ctrl-C` mid-bootstrap to confirm cancellation also rolls the project back under the same step name.
8. Fix the project-local bootstrap command or underlying application issue, then rerun `stage up`; if a forced clean stop is needed first, run `stage down` before retrying.

## Multi-Project Validation

1. Start the first project with `stage up`. (`up` registers and attaches the first project; the explicit `attach` step in step 2 is what proves multi-project routing.)
2. From the second project, run `stage attach` explicitly.
3. Confirm both routes work through the shared gateway.
4. Use `.develop` hostnames for the operator-facing validation examples unless the test explicitly targets another allowed suffix.
5. Confirm shared-gateway readiness is healthy while both projects are attached.
6. Confirm Docker resource names leave enough room to distinguish the project slugs in listings.
7. Run `stage status` and verify both projects report the correct hostnames, routes, and recorded identities.

## Teardown Validation

1. Run `stage down` in the representative application.
2. Confirm only that project stops.
3. If a second project remains attached, confirm its route still works.
4. Run `stage status` and confirm the recorded state matches reality.

## Documentation Validation

1. Review `README.md` and `docs/runtime-contract.md`.
2. Confirm they both point operators to `.env.stageserve` as the stack-owned defaults file.
3. Confirm they both describe `stage-` as the project-scoped runtime prefix.
4. Confirm they both describe bootstrap failure as a rollback-triggering lifecycle failure.

## Validation Notes

- Record the representative applications used.
- Record whether the bootstrap command exercised migration-only behavior or a broader setup command.
- Record whether validation ran from the repository working copy or from the deployed copy under `$HOME/docker/20i-stack`.
- Record any real-daemon gap that was not rerun during implementation.