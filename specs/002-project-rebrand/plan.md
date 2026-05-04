# Implementation Plan: StageServe Rebrand And `stage` Command Cutover

**Branch**: `002-project-rebrand` | **Date**: 2026-04-01 | **Spec**: `specs/002-project-rebrand/spec.md`
**Input**: Feature specification from `specs/002-project-rebrand/spec.md`

## Summary

Finalize the product rename to StageServe, establish `stage` as the canonical root command, and align maintained docs/specs to the current naming scheme in the same delivery unit. The implementation keeps runtime behavior scoped to the existing StageServe contract while removing stale prior-name references and compatibility-wrapper assumptions from active operator guidance.

## Technical Context

**Language/Version**: Go CLI with thin repo-root launcher, Markdown docs/specs, archived shell material for historical reference only  
**Primary Dependencies**: `cmd/stage`, `stage`, `stage-bin`, current docs/spec artifacts under `README.md`, `docs/`, and `specs/`  
**Storage**: project-local `.env.stageserve`, stack-home `.env.stageserve`, runtime state under `.stageserve-state/`, documentation files  
**Testing**: focused CLI validation, docs/spec sweeps, and targeted grep for stale naming  
**Target Platform**: macOS primary  
**Project Type**: CLI/runtime tool with operator docs and archived historical surfaces  
**Performance Goals**: preserve current operator responsiveness and keep command help/output consistent with the active StageServe contract  
**Constraints**: keep docs aligned to implemented behavior; do not reintroduce compatibility shims or legacy root commands as current behavior; document repo rename separately from manual local-folder handling  
**Scale/Scope**: one repository-wide rename sweep, one canonical CLI surface, and the maintained user-facing surfaces that still carry prior naming

## Constitution Check

- [x] Ease-of-use impact is documented, including the shorter canonical command path.
- [x] Reliability expectations are explicit, including current config precedence and state naming.
- [x] Robustness boundaries are defined: this is a naming and contract-alignment feature, not a runtime redesign.
- [x] Documentation surfaces requiring same-change updates are identified.
- [x] Validation covers command help/examples plus stale-reference sweeps.

### Constitution Check Notes

- Ease of use improves by converging on `stage <subcommand>` as the only supported root-command model.
- Reliability is preserved by keeping the current config and state contract: project `.env.stageserve`, stack-home `.env.stageserve`, and `.stageserve-state`.
- Robustness stays bounded because archived shell material remains archive-only and is not revived as current functionality.
- Same-change documentation scope includes top-level docs, migration docs, runtime-contract wording, and older maintained specs.
- Validation centers on targeted grep plus focused command/help checks.

## Project Structure

### Documentation (this feature)

```text
specs/002-project-rebrand/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli-contract.md
└── tasks.md
```

### Source Code (repository root)

```text
stage
stage-bin
cmd/
└── stage/
README.md
docs/
previous-version-archive/
docker-compose.20i.yml
docker-compose.shared.yml
core/
```

**Structure Decision**: Keep this feature centered on the active Go CLI, the thin `stage` launcher, and the maintained docs/spec surfaces. Archived material stays in `previous-version-archive/` and is never described as current behavior.

## Post-Design Constitution Check

- [x] Ease-of-use remains improved after design: one canonical `stage` entrypoint is explicit.
- [x] Reliability remains explicit after design: docs/specs point to the implemented config/state contract.
- [x] Robustness remains bounded after design: no compatibility layer or runtime redesign is introduced.
- [x] Documentation parity remains covered after design.
- [x] Validation remains sufficient after design: grep sweeps plus focused command/help checks are named.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
