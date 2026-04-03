# Language Choices Research Report: StackLane Codebase Rewrite

**Feature Branch**: `003-rewrite-language-choices`  
**Created**: 2026-04-03  
**Status**: Draft

---

## Context

StackLane is currently implemented as a single Bash library (`lib/stacklane-common.sh`, ~2200 lines) with thin entry-point wrappers. It manages Docker Compose project lifecycles, shared gateway configuration, DNS bootstrapping, state persistence, and TLS certificate generation — all from the macOS command line.

The goal of this spec is to assess whether the codebase should be rewritten in a compiled or higher-level language, and if so, which language best satisfies the project's constitution principles.

---

## What the Codebase Does (Surface to Rewrite)

| Concern | Current implementation |
|---|---|
| Arg parsing | `case` blocks, `shift` |
| Config loading | `source`-based `.env` / `.20i-local` parsing |
| State persistence | Plain `.env` files in `~/.20i-state/projects/` |
| Port allocation | Loop over active state files + `lsof`/`netstat` |
| Docker orchestration | Subprocess calls to `docker compose` |
| Gateway config generation | Inline `heredoc` Nginx config writers |
| DNS management | Subprocess calls to `brew`, `osascript`, `dnsmasq` |
| TLS cert management | Subprocess call to `mkcert` |
| Registry | TSV file with tab-escaped values |
| Output | `printf` to stdout/stderr |

The logic is correct and well-structured, but Bash becomes brittle as complexity grows: no types, limited error handling, hard-to-test functions, and fragile string manipulation.

---

## Candidate Languages

### 1. Go

**Strengths**
- Compiles to a single static binary with no runtime dependencies — ideal for a CLI tool
- Excellent standard library for subprocess execution, file I/O, and string handling
- Strong ecosystem for CLI frameworks (`cobra`, `urfave/cli`), config parsing, and template rendering
- Trivially cross-compilable (macOS arm64/amd64)
- Built-in concurrency for parallel port checks or container probes
- Easy to ship and update (`go install`, Homebrew formula, or a release binary)
- Very fast startup time (sub-millisecond)
- Clear, typed, and testable — a direct answer to Bash's biggest weaknesses
- Used widely in the Docker/Kubernetes ecosystem; language familiarity with adjacent tools

**Weaknesses**
- More verbose than scripting languages; boilerplate for simple tasks
- No REPL; slower inner-loop iteration during initial development
- Requires a Go toolchain for contributors

**Constitution alignment**: High. Static binary = zero friction to install. Typed, testable code = reliability. Subprocess orchestration via `os/exec` is idiomatic.

---

### 2. Python

**Strengths**
- Familiar to most developers; fast to prototype
- Rich standard library (`subprocess`, `pathlib`, `os`, `shutil`)
- Good CLI libraries (`click`, `typer`, `argparse`)
- Easy to read and maintain
- Excellent test tooling (`pytest`)

**Weaknesses**
- Requires Python 3.x to be installed (not guaranteed on macOS without Homebrew)
- No single-binary distribution without additional tooling (`PyInstaller`, `Nuitka`, `shiv`) — adds distribution friction
- Slower startup than Go for CLI calls (100–300 ms CPython import overhead at cold start)
- Virtual environment management is an ongoing friction point for contributors and operators
- Type safety is opt-in (mypy)

**Constitution alignment**: Moderate. Development is fast, but distribution and startup friction conflict with Principle I (ease of use).

---

### 3. Rust

**Strengths**
- Single static binary, similar to Go
- Best-in-class performance and memory safety
- Growing CLI ecosystem (`clap`, `duct`)
- Zero runtime dependencies

**Weaknesses**
- Steep learning curve and slow compile times
- Relatively verbose for the kind of subprocess-heavy orchestration this tool does
- Smaller contributor pool
- Overengineered for a local DevOps CLI tool that isn't performance-critical

**Constitution alignment**: Moderate. The binary distribution model is excellent, but complexity cost is high relative to the functional gains. Violates the "simplest option that works" principle unless there's a specific need for systems-level performance.

---

### 4. Node.js / TypeScript

**Strengths**
- TypeScript provides strong typing and good tooling
- Broad developer familiarity
- Good CLI libraries (`commander`, `oclif`, `yargs`)
- Active ecosystem

**Weaknesses**
- Requires Node.js runtime — not always present, and version management (nvm/fnm) adds friction
- No single-binary distribution without packaging tools (`pkg`, `nexe`, `bun`)
- `node_modules` and dependency tree are notorious distribution friction points
- Slow startup for CLI tools (v8 JIT warm-up)

**Constitution alignment**: Low-moderate. Runtime dependency and distribution model work against ease of use and predictability.

---

### 5. Deno / Bun

**Strengths**
- Both support TypeScript natively
- Bun in particular compiles to a single executable
- Better startup times than Node.js

**Weaknesses**
- Less mature ecosystems compared to Go or Python
- Smaller contributor familiarity
- Bun's single-executable feature is relatively new and not yet battle-tested for distribution
- Adds unfamiliar runtime layer for operators and contributors

**Constitution alignment**: Low-moderate. Promising for the future but introduce unnecessary uncertainty right now.

---

## Evaluation Matrix

| Criterion | Go | Python | Rust | Node/TS | Deno/Bun |
|---|---|---|---|---|---|
| **No runtime dependency** | ✅ | ❌ | ✅ | ❌ | ⚠️ |
| **Single binary distribution** | ✅ | ⚠️ | ✅ | ⚠️ | ⚠️ |
| **Startup speed (CLI)** | ✅ | ⚠️ | ✅ | ❌ | ⚠️ |
| **Type safety** | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| **Testability** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Subprocess orchestration** | ✅ | ✅ | ⚠️ | ✅ | ✅ |
| **Contributor familiarity** | ✅ | ✅ | ❌ | ✅ | ❌ |
| **Ecosystem maturity** | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| **Complexity cost** | Low | Low | High | Low-Med | Med |
| **Distribution simplicity** | ✅ | ⚠️ | ✅ | ❌ | ⚠️ |

---

## Recommendation

**Go** is the recommended language for the rewrite.

It is the only candidate that satisfies all of the constitution's primary constraints simultaneously:

1. **Ease of use** — ships as a single binary; operators install it once with no runtime to manage
2. **Reliability** — strongly typed, thoroughly testable, deterministic argument parsing, no global-state side effects from `export` or `source`
3. **Robustness** — structured error handling with `error` interface; no silent failures from `set -euo pipefail` quirks
4. **Friction removal** — replaces fragile Bash string processing, heredoc template construction, and environment variable propagation with typed structs and explicit data flow

Python is a viable fallback if Go is rejected for contributor-preference reasons, with the explicit acknowledgement that distribution and startup performance will require mitigation work.

Rust is rejected as over-engineered for this problem domain.

Node.js and Deno/Bun are rejected on distribution and runtime-dependency grounds.

---

## Scope Boundaries for the Rewrite

The following concerns are **in scope**:

- Argument parsing and dispatch (`stacklane --up`, `--down`, etc.)
- Configuration loading (`.env`, `.20i-local`, env precedence)
- Port allocation and state file read/write
- Docker Compose subprocess invocation
- Gateway config file generation (Nginx heredoc → Go template)
- Registry TSV read/write
- DNS status checking and `dnsmasq` management
- TLS certificate management via `mkcert` subprocess

The following concerns are **out of scope** for the initial rewrite decision:

- Changes to Docker Compose file structure
- Changes to the shared gateway's Nginx configuration semantics
- Operator-visible command names or flags (migration contract is fixed in spec-002)
- New features not already present in the Bash implementation

---

## Open Questions

1. **Minimum Go version**: Should the rewrite target the current stable release (Go 1.22+) or a slightly older LTS-friendly version? Impact: module support and `slices`/`maps` standard library availability.
2. **Testing strategy**: Should integration tests shell out to the compiled binary, or should unit tests mock Docker/filesystem boundaries? Both are valid; the plan should define the boundary.
3. **Incremental vs big-bang rewrite**: Could the Go binary initially wrap `stacklane-common.sh` (thin shell-out) and progressively replace individual functions? This reduces risk but delays payoff.
4. **Homebrew formula**: Should distribution via a Homebrew tap be a day-one goal or deferred to a follow-up spec?
