---
applyTo: "README.md,docs/**,specs/004-workflow-and-lifecycle/**,.env.stageserve.example"
---

Treat these files as the active operator and workflow contract.

When implementation changes the config, naming, compose-file layout, stack selection, lifecycle flow, or bootstrap behavior, update the corresponding docs/spec files in the same change so the written contract stays synchronized with the code.

Prefer current active surfaces and terminology:
- project-local config: `<project>/.env.stageserve`
- stack-wide defaults: `<stack-home>/.env.stageserve`
- active runtime compose file: `docker-compose.20i.yml`
- shared compose file: `docker-compose.shared.yml`
- archived material: `previous-version-archive/`

Do not describe archived TUI material or legacy Bash behavior as current functionality.

If a doc statement conflicts with code, align the doc to implemented behavior unless the task is explicitly to change the runtime.