# Implementation Plan: Guided Experience And Runtime Hardening

**Branch**: `011-guided-experience-and-runtime-hardening` | **Date**: 2026-06-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-guided-experience-and-runtime-hardening/spec.md`

## Summary

Harden StageServe runtime and guided-entry correctness first, then complete the guided UX follow-up in bounded phases. Implement merge-gate work to eliminate missing-asset startup failures, machine-not-ready misclassification, silent lifecycle error handling, and password flag safety risk; then deliver shared presentation parity, recovery surfaces, running-project day-2 actions, and async confirmation/progress behavior. Keep the simple-first contract from spec 007 intact and preserve automation-safe direct command semantics.

Preserve the retained spec 007 root-routing contract during hardening work: machine-not-ready, project-missing-config, ready-to-run, running-project, and recovery-needed states must remain deterministic and validation-backed.

## Technical Context

**Language/Version**: Go 1.26 (toolchain go1.26.2)  
**Primary Dependencies**: Cobra, Bubble Tea, Lip Gloss, Bubbles, Huh, Docker SDK, internal lifecycle/guidance/state seams  
**Storage**: Filesystem state under `.stageserve-state` and project-local `.env.stageserve`  
**Testing**: `go test` package-focused suites, targeted terminal/manual validation for TTY/non-TTY flows  
**Target Platform**: macOS primary operator environment, Linux-compatible runtime checks where already supported  
**Project Type**: Go CLI/TUI application  
**Performance Goals**: Long-running guided actions display activity feedback within 250 ms of confirmation  
**Constraints**: Preserve non-TTY and JSON purity guarantees; no new persistent config surface; keep direct command behavior automation-safe  
**Scale/Scope**: Spec-local implementation across guidance, commands, lifecycle, compose, config, docs/help, and installer bundling contract

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Ease-of-use impact is documented. This plan keeps bare `stage` as the simple entrypoint, adds clearer failure/recovery paths, and avoids adding first-level operator friction.
- [x] Reliability expectations are explicit. Runtime asset preflight, deterministic failure remedies, password-source boundaries, and state/route consistency are defined in spec requirements and contracts.
- [x] Robustness boundaries are defined. Partial failure, stale registry, gateway route sync, and missing runtime assets are treated as first-class failure modes with explicit operator outcomes.
- [x] Documentation surfaces are identified: `README.md`, `docs/installer-onboarding.md`, `docs/runtime-contract.md`, command help text, `.env.stageserve.example`, and spec-local validation docs.
- [x] Validation scope is explicit: startup, status/inspection, teardown, and failure/recovery paths are required for merge-gate and follow-up completion.

Post-design re-check: **PASS**. Phase 0 and Phase 1 artifacts preserve all constitution gates without requiring exceptions.

## Project Structure

### Documentation (this feature)

```text
specs/011-guided-experience-and-runtime-hardening/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── guided-runtime-hardening-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/stage/commands/
core/config/
core/guidance/
core/lifecycle/
core/onboarding/
core/state/
infra/compose/
observability/status/
docs/
install.sh
```

**Structure Decision**: Keep the existing monorepo CLI structure and implement through existing seams. Do not introduce new top-level packages for this spec.

## Phase Plan

### Phase 0: Research And Decisions

- Confirm bundled installer artifact as the merge-gate runtime-asset fix path.
- Confirm preflight checks happen before compose startup and are surfaced as StageServe-native remedies.
- Confirm consistent severity/presentation model strategy across guided/text/direct paths.
- Confirm minimum async action slice for Bubble Tea execution (`up`, `down`, `attach`, `detach`, recovery retries).
- Confirm service-selection fallback rules for logs/restart flows.

Output: [research.md](./research.md)

### Phase 1: Design And Contracts

- Define entities for runtime assets, readiness summaries, lifecycle failure envelopes, service action targets, and async action state.
- Define contract for merge-gate versus follow-up behavior boundaries.
- Define quickstart verification matrix with merge-gate and follow-up command sets.

Outputs:

- [data-model.md](./data-model.md)
- [contracts/guided-runtime-hardening-contract.md](./contracts/guided-runtime-hardening-contract.md)
- [quickstart.md](./quickstart.md)

### Phase 2: Implementation Sequencing (Reference)

- Runtime and lifecycle hardening first (`T004a`-`T019`) before UX expansion.
- Retained root-routing, non-TTY or JSON safety, and recovery-ordering obligations land during the same hardening slice (`T007`, `T010`-`T014`, `T023a`, `T043`-`T045`).
- Shared presentation parity (`T020`-`T024`) before day-2 surface expansion.
- Running-project expansion and async confirmation/progress (`T025`-`T037`).
- Dedicated TUI design polish (`T038`-`T040`).
- Docs/help + validation at merge gate and follow-up closure (`T041`-`T050`).

## Complexity Tracking

No constitution violations accepted for this plan.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
