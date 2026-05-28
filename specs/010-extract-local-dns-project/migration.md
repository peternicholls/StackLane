# Migration Strategy: Moving DNS Out Of StageServe

## Objective

Move DNS provider logic out of StageServe without breaking current StageServe workflows and without forcing StageServe users to install a second binary.

## Migration Stages

### Stage 1 - Stabilize The Current StageServe Boundary

Goal: isolate DNS calls behind one StageServe-owned adapter while still using the in-repo implementation.

Actions:

1. Add a small StageServe adapter/helper around the current `platform/dns` package.
2. Update `cmd/stage/commands/dnssetup.go` to use the adapter.
3. Update `core/onboarding/readiness_dns_darwin.go` to use the adapter.
4. Keep all behavior unchanged.

Rollback point: stop here and keep the in-repo implementation if the boundary turns out to be wider than expected.

### Stage 2 - Create The Standalone DNS Repo

Goal: stand up the new repo with a generic public surface.

Actions:

1. Create the new repo and module path.
2. Port the current DNS logic into a generic library.
3. Replace StageServe-specific names with generic request fields.
4. Add the standalone CLI and docs.

Rollback point: the new repo can remain private or unpublished until local StageServe integration proves the API.

### Stage 3 - Develop Through A Local Dependency Replacement

Goal: prove the split before any public tag is cut.

Actions:

1. Add a local `replace` directive in StageServe's `go.mod`.
2. Point the StageServe adapter at the new module.
3. Run focused StageServe tests and macOS smoke checks.
4. Adjust the dependency contract while it is still cheap to change.

Rollback point: restore the old import path and keep the new repo as a draft if the contract is not yet stable.

### Stage 4 - Tag And Adopt The Dependency

Goal: move from local development wiring to a real versioned dependency.

Actions:

1. Tag the DNS repo.
2. Replace the local `replace` directive with the tagged version.
3. Re-run StageServe validations.
4. Remove the old `platform/dns` code from StageServe.

Rollback point: revert to the previous StageServe commit if the tagged dependency exposes a release blocker.

### Stage 5 - Clean Up And Document

Goal: make the split durable for maintainers and users.

Actions:

1. Update StageServe docs to explain that local DNS is dependency-backed.
2. Add standalone docs in the DNS repo.
3. Decide whether to add packaging channels such as Homebrew now or later.
4. Record future-platform work as separate follow-up planning.

## Cutover Rules

- Do not delete `platform/dns` before StageServe passes local integration checks against the new repo.
- Do not tag the new repo before StageServe has consumed it through a local replacement.
- Do not change StageServe command names during the cutover.

## Suggested Rollout Order

1. Boundary prep in StageServe
2. New repo scaffold and port
3. Local replacement integration
4. Tag prerelease
5. StageServe tagged dependency adoption
6. Old code removal
7. Docs and distribution follow-up

## Success Markers Per Stage

- Stage 1 success: StageServe has one clear DNS adapter seam.
- Stage 2 success: the new repo builds and tests independently.
- Stage 3 success: StageServe works through the local replacement.
- Stage 4 success: StageServe works against a tagged dependency and no longer uses in-repo provider code.
- Stage 5 success: both repos explain the split clearly and can evolve independently.