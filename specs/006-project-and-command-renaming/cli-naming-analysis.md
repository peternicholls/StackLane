# CLI Command Name Analysis: Is `stage` the right name?

> A research-based analysis of the current command name and alternatives.  
> Updated to include the project-rename scenario: because the tool has no public release yet, renaming the project itself carries zero migration cost and opens up verb-first names that a binary rename alone cannot cleanly achieve.

---

## 1. The question

Homebrew ships as `brew`. Docker ships as `docker`. Git ships as `git`. Bun ships as `bun`. The current StageServe binary is invoked as `stage` — nine keystrokes, a noun-compound, and no verb energy.

Two distinct options exist:

1. **Binary rename only** — keep the project name "StageServe" but ship the CLI under a shorter name.
2. **Full project rename** — rename the project and CLI together, freeing up verb-first names that read naturally as commands (`hoist up`, `stage up`, `berth up`).

Because there is no public release yet, option 2 costs nothing today. This document evaluates both paths.

---

## 2. What the research says about CLI naming

### CLIG Guidelines (clig.dev)
The authoritative community standard for CLI design says, under the **Naming** section:

> *"Keep it short. Users will be typing it all the time. Don't make it too short: the very shortest commands are best reserved for the common utilities used all the time, such as `cd`, `ls`, `ps`."*
>
> *"Use only lowercase letters, and dashes if you really need to."*
>
> *"Make it a simple, memorable word. But not too generic, or you'll step on the toes of other commands and confuse users."*

### The Poetics of CLI Command Names (smallstep.com)
Carl Tashian's essay — referenced directly by the CLIG guidelines — offers additional heuristics:

- **Anti-patterns**: Never use `tool`, `kit`, `util`, `easy` in a name. Avoid requiring the shift key. Never embed a version number. Avoid names that are too generic (`convert`, `stack`). Don't make users fight their own keyboard.
- **Excellent names**: `curl` (verb, sounds like "see URL"), `vim` (the feeling you get), `cat`, `step` (smallstep's own CLI — short, easy, meaning-agnostic and therefore future-proof). The name should "feel like a soft breeze across the keys."
- **Keyboard ergonomics matter**: `sha256sum` "feels like gargling sand." Docker Compose was originally `plum`, renamed to `fig` because it was one-handed and flowed easily. `fig` was later superseded by the `docker compose` subcommand — but the ergonomics lesson stands.
- **Meaninglessness is fine**: If the tool's scope may evolve, a short, invented or tangentially-related word ages better than a literal, domain-specific description.
- **The more niche the tool, the longer the name can be.** Short names (2–4 chars) are for widely-used, every-session utilities. 5–7 chars is a reasonable sweet spot for a developer workflow tool.

### Precedents from comparable tools

| Tool | Full name | Command | Length | Notes |
|------|-----------|---------|--------|-------|
| Homebrew | Homebrew | `brew` | 4 | Derived from brand; thematic |
| Docker Compose | Docker Compose | `docker compose` | — | Subcommand of parent tool |
| Vagrant | Vagrant | `vagrant` | 7 | Full name used |
| Kubectl | Kubernetes CLI | `kubectl` | 6 | Frequently aliased to `k` |
| Terraform | Terraform | `terraform` | 9 | Often aliased to `tf` |
| Nix | Nix | `nix` | 3 | Short, abstract, iconic |
| Deno | Deno | `deno` | 4 | Invented portmanteau |
| Bun | Bun | `bun` | 3 | Short, invented, thematic (bun as in bread roll — fast) |
| Artisan | Laravel CLI | `artisan` | 7 | Craft metaphor |
| step | Smallstep CLI | `step` | 4 | Completely meaning-agnostic, easy to type |
| Direnv | Direnv | `direnv` | 6 | Descriptive compound |
| mkcert | mkcert | `mkcert` | 6 | `mk`-prefixed, clear domain |

**Key observation**: 4–7 characters is the practical sweet spot for developer tools that are used daily but not every single command invocation. Both ends of that range are well-represented by respected tools.

---

## 3. Assessing `stage` against those criteria

| Criterion | Score | Notes |
|-----------|-------|-------|
| Lowercase only | ✅ | No issues |
| No version number | ✅ | |
| Not generic | ✅ | "stage" is specific enough |
| Memorable | ⚠️ | Descriptive but long; two concepts fused |
| Easy to type | ⚠️ | 9 keystrokes; both hands required but no awkward combos |
| No major conflicts | ✅ | No well-known `stage` command exists |
| Future-proof | ⚠️ | "stack" implies the current 20i-style hosting stack; if StageServe supports other runtimes (`laravel`, `node`), the name implies more than it should |
| Tab-completion friendly | ✅ | Unique enough prefix that `sta<tab>` likely resolves |
| Script-stable | ✅ | Already in use; changing costs real migration effort |

**Summary**: `stage` is not a *bad* name. It's unambiguous, lowercase, has no known conflicts, and is currently in use. The main weaknesses are **length** (9 chars is on the long side for a daily-use tool) and a mild **future-proofing concern** if the "stack" concept evolves.

---

## 4. Proposals

Proposals are grouped into two tracks.

**Track A — Binary rename, keep "StageServe" as the project name.** Lower disruption; the brand stays. CLI gets shorter.

**Track B — Full project rename.** Unlocks verb-first names that read as natural commands. Zero migration cost at pre-release.

---

### Track A: Binary rename only

#### A0 — `stage` (status quo)

**Command**: `stage` | **Length**: 9  
**Risk**: None | **Cost**: None

Already in use. Unambiguous, lowercase, no conflicts. The 9-character length is a minor ergonomic inconvenience, not a blocking problem.

**Best if**: The brand matters more than CLI ergonomics, or first-impression polish is not a priority.

---

#### A1 — `lane`

**Command**: `lane` | **Length**: 4

Derived from the second half of "StageServe." A _lane_ is a channel through which traffic flows — an accurate metaphor for a routed, named local dev environment.

**Typing feel**: Very good. `l-a-n-e` — natural left-to-right roll.

**Conflicts**: No standalone `lane` CLI in Homebrew or common macOS toolchains. `fastlane` (mobile CI) uses `lane` as a subcommand concept, not a binary.

**Strengths**: Short (same length as `brew`, `bun`, `deno`); preserves half the product name; metaphor holds.

**Weaknesses**: Loses the "stack" concept from the identity; the command name has no verb quality.

---

#### A2 — `stln`

**Command**: `stln` | **Length**: 4

The Docker resource prefix already used internally (`stage-<slug>`, `stage-<slug>-runtime`). Adopting it as the CLI name unifies the internal and external identity.

**Typing feel**: Acceptable. `s-t-l-n` — all left-hand, somewhat awkward.

**Conflicts**: None.

**Strengths**: Internal consistency — what appears in `docker ps` matches what you type; professional and terse.

**Weaknesses**: Not pronounceable; no obvious mental hook; left-hand-heavy; feels like an abbreviation, not a name.

---

### Track B: Full project rename (verb-first)

A verb-first name changes both the project identity and the CLI command together. The command then reads as a natural action: `hoist up`, `stage up`, `berth up`. This is the model used by `step` (Smallstep), `spin` was considered (see conflicts below), and many other tools.

Key constraint from the research: **verbs are future-proof**. A noun-based name like "StageServe" describes the current scope (a hosting stack, a lane); a verb describes what you _do_ with the tool regardless of what hosting stacks it later supports.

---

#### B1 — `hoist`

**Command**: `hoist` | **Length**: 5  
**Homebrew formula**: none | **Local conflict**: none

To hoist is to raise something up — a flag, a sail, a load. Maps directly onto the primary use case (`hoist up`, `hoist down`). The nautical metaphor sits naturally in the Docker ecosystem without being confused with Docker itself.

**Typing feel**: Excellent. `h-o-i-s-t` — alternating hands, no stretching, smooth and satisfying.

**Conflicts**: No Homebrew formula. Not installed on test machine. Fermyon has no tool called `hoist`. No significant naming collision found.

**Subcommand read-aloud test**:
```
hoist up       ✅ natural
hoist down     ✅ natural
hoist status   ✅ fine
hoist attach   ✅ fine
hoist doctor   ✅ fine
```

**Strengths**:
- Strong action verb with an immediate physical image
- Completely future-proof — doesn't reference "stack", "lane", or "20i"
- 5 chars; same neighbourhood as `docker` (6)
- Nautical/cargo metaphor is thematically adjacent to Docker without clashing
- Clean namespace

**Weaknesses**:
- No continuity with "StageServe" — requires committing to the new identity
- Slightly less self-describing on first encounter

---

#### B2 — `stage` ⭐

**Command**: `stage` | **Length**: 5  
**Homebrew formula**: none | **Local conflict**: none

To stage something is to set it up, prepare it — a deeply familiar concept in development workflows (`git stage`, staging environments, stage management). The theatre metaphor also works: you're setting the stage for your application to perform.

**Typing feel**: Very good. `s-t-a-g-e` — smooth, common letter sequence.

**Conflicts**: No Homebrew formula. `git stage` is an alias for `git add` but that is a subcommand of `git`, not a standalone binary — no actual clash. No other `stage` binary in common macOS dev toolchains.

**Subcommand read-aloud test**:
```
stage up       ✅ natural
stage down     ✅ natural
stage status   ✅ fine
stage init     ✅ fine
stage doctor   ✅ fine
```

**Strengths**:
- Verb that developers already know and trust
- "Staging environment" is a universally understood concept — lowers the learning curve
- 5 chars, clean namespace
- Future-proof; doesn't encode the current hosting-stack scope

**Weaknesses**:
- "Staging environment" usually implies pre-production, not local development; could imply the wrong workflow tier to newcomers
- `git stage` creates mild mental ambiguity (though it is not a binary conflict)

---

#### B3 — `berth`

**Command**: `berth` | **Length**: 5  
**Homebrew formula**: none | **Local conflict**: none

A berth is where a vessel (container ship) docks. To berth is to bring a ship into its allocated slot. The metaphor maps cleanly: each project gets its own berth in the shared harbour. Docker's container/ship vocabulary makes this land without explanation.

**Typing feel**: Good. `b-e-r-t-h` — slightly less common sequence, but not awkward.

**Conflicts**: No Homebrew formula. No CLI tool using this name found.

**Subcommand read-aloud test**:
```
berth up       ✅ natural ("bring it into berth")
berth down     ✅ fine
berth status   ✅ fine
berth attach   ✅ fine
```

**Strengths**:
- Precisely accurate metaphor: each project gets a named berth in the shared harbour
- Docker-vocabulary-adjacent without any collision
- Clean namespace

**Weaknesses**:
- Less immediately obvious than `hoist` on first encounter
- "Berth" primarily means a place/slot; the verb form is less common than "hoist" or "stage"
- 5 chars but slightly less ergonomic than `hoist`

---

#### B4 — `loft`

**Command**: `loft` | **Length**: 4  
**Homebrew formula**: none | **Local conflict**: none

Evokes elevation (up/down), craft space, and lightness. Pairs naturally with `loft up`. Architectural connotation fits an infrastructure tool.

**Typing feel**: Very good. `l-o-f-t` — clean one-syllable roll.

**Conflicts**: No Homebrew formula. Loft.sh (a Kubernetes/vcluster SaaS) ships a CLI companion named `loft` — low overlap with the StageServe audience but worth checking on target machines with `which loft`.

**Weaknesses**: The Loft.sh CLI is a real, if low-prevalence, conflict risk; less verbally action-oriented than `hoist`.

---

### Ruled out

| Name | Reason |
|------|--------|
| `stack` | **Taken** — Haskell Tool Stack, widely installed (`brew install haskell-stack`) |
| `spin` | **Taken** — Homebrew formula (SPIN model checker, 374 installs/year) + Fermyon Spin (WebAssembly) |
| `moor` | **Taken** — Homebrew formula (pager tool, 7,275 installs/year) |
| `sl` | **Taken** — classic joke command (`brew install sl`); 2-char namespace reserved for POSIX utilities |
| `docker` / `dock` | Too close to Docker; would create persistent confusion |

---

## 5. Comparison matrix

| Name | Track | Length | Verb? | Typeable | Conflicts | Project rename needed |
|------|-------|--------|-------|----------|-----------|----------------------|
| `stage` | A0 | 9 | ❌ | ✅ | none | no |
| `lane` | A1 | 4 | ❌ | ✅✅ | none | no |
| `stln` | A2 | 4 | ❌ | ⚠️ | none | no |
| `hoist` | B1 | 5 | ✅✅ | ✅✅ | none | yes |
| **`stage`** | **B2** | **5** | **✅✅** | **✅✅** | **none** | **yes** |
| `berth` | B3 | 5 | ✅ | ✅ | none | yes |
| `loft` | B4 | 4 | ✅ | ✅✅ | low risk | yes |



---

## 6. Migration cost

**Before any public release**: cost is essentially zero. The binary name, GitHub repo name, env var prefix, and all docs can change in a single commit batch. The Docker resource prefix (`stage-*`) is independent of the CLI name and need not change.

**After a public release** (for reference):
- Installer (`install.sh`) and binary name on disk must change
- `STAGESERVE_*` env vars are independent of the binary — keep them or rename with a deprecation period
- Docker resource names (`stage-*`) are independent — no change needed
- Any `stage up` in CI configs or scripts needs a find-and-replace
- Transition path: ship both names, deprecate the old one over 1–2 releases

The current pre-release state means **now is the lowest-cost moment to rename**.

---

## 7. Recommendation

### Selected direction

Rename the project to **StageServe** and ship the CLI command as **`stage`**.

This keeps the strongest benefits of the verb-first track while using the name you preferred:

- `stage up` / `stage down` read naturally
- no Homebrew formula conflict was found
- 5-character command with strong typing ergonomics
- future-proof naming not tied to the current stack scope

### Why this works well

`stage` is familiar to developers, easy to remember, and immediately action-oriented. Pairing it with **StageServe** preserves a product-style project name while keeping the command short and verb-like.

### If keeping "StageServe" as the project name

Use **`lane`** as the CLI command. It's short (4 chars), conflict-free, preserves half the brand identity, and the routing metaphor is accurate. It does not feel like a verb, but it is crisp and memorable.

### If open to a full project rename

Use **`stage`**.

- It is a clean, strong verb — the command name _is_ the action
- `stage up` / `stage down` are immediately legible to anyone who reads them
- No Homebrew formula, no local conflict found
- 5 characters, excellent keyboard ergonomics
- Developers already understand the setup/preparation meaning
- Completely future-proof: no reference to "stack", "lane", "20i", or any current scope

`hoist` remains a strong alternative if you want a stronger physical metaphor. `berth` is the most precise Docker-vocabulary metaphor but relies on users knowing what a ship's berth is.

### Conflict check before committing

```bash
which stage   # should return nothing
which hoist   # should return nothing
which berth   # should return nothing
```

---

## 8. Key conflicts

- `stack` is already the command for **Haskell Tool Stack** (`brew install haskell-stack`) and would create immediate collisions on Haskell developer machines.
- `spin` is already taken by the SPIN model checker formula and is also strongly associated with Fermyon Spin.
- `moor` is already taken by a Homebrew formula and has non-trivial install footprint.

These conflicts are why the viable short verb candidates narrowed to `stage`, `hoist`, and `berth`.

---

## 9. Sources

- [Command Line Interface Guidelines — Naming section](https://clig.dev/#naming)
- [The Poetics of CLI Command Names — Carl Tashian, Smallstep](https://smallstep.com/blog/the-poetics-of-cli-command-names/)
- [12 Factor CLI Apps — Jeff Dickey](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)
- [Homebrew Formulae](https://formulae.brew.sh)
- Docker Compose naming history (`plum` → `fig` → `docker-compose` → `docker compose`)
- Homebrew (`Homebrew` → `brew`)
- Comparable tool survey: git, nix, deno, bun, vagrant, terraform, kubectl, step

---

## 10. StageServe + `stage` rename execution checklist

This checklist assumes the selected direction is:

- Project name: **StageServe**
- CLI command: **`stage`**
- Current status: pre-release (no public compatibility guarantees yet)

### A. Naming contract decisions (lock before editing files)

1. Decide whether internal prefixes remain stable:
- Keep `STAGESERVE_*` env vars for now, or rename to `STAGESERVE_*`.
- Keep `.env.stageserve` filename contract, or rename to `.env.stageserve`.
- Keep Docker prefixes (`stage-*`) unchanged unless there is explicit value in renaming.

2. Decide compatibility posture:
- **Hard switch** (pre-release clean break): only `stage` shipped.
- No compatibility shim or secondary command path should remain.

3. Decide brand surfaces:
- GitHub repo name (`stage` vs `stageserve`).
- Binary artifact names in releases (tarballs, checksums, install script labels).

### B. Code and build system changes

1. CLI binary and entrypoint:
- Update build/install targets so installed binary is `stage`.
- Do not keep any compatibility symlink or alternate command path.

2. Command help/version strings:
- Update usage banners, examples, and command synopsis to `stage`.
- Ensure `stage version` prints StageServe branding consistently.

3. Error/help output grep traps:
- Search for literal `stage` in user-facing errors.
- Replace references that users copy/paste directly.

4. Shell completion scripts:
- Regenerate completion for bash/zsh/fish under `stage`.
- Do not generate or document completions for any retired command name.

5. Makefile and installer:
- Update install destination filename and any chmod/symlink lines.
- Validate uninstall paths do not orphan stale `stage` binaries.

### C. Documentation and spec contract changes

1. Update active docs together:
- README quickstart examples
- onboarding docs
- runtime contract docs
- migration notes

2. Update examples and snippets:
- Any `stage up/down/status/logs` examples
- CI snippets, automation docs, troubleshooting sections

3. Keep archive boundaries explicit:
- Do not present `previous-version-archive/` wrappers as active behavior.

4. Keep branding singular:
- Do not add a "formerly Stacklane" tagline to active README branding.

### D. CI/CD and release pipeline niches

1. Pipeline command invocations:
- Replace `stage` command in CI jobs and smoke scripts.

2. Cache and artifact keys:
- Update cache keys that include command/repo name.
- Confirm old cache keys do not silently mask failures.

3. Release filenames and checksums:
- Ensure checksum files reference renamed artifacts.
- Verify install script matches released asset names exactly.

4. Container and package metadata:
- If publishing images, update tags/labels where name is embedded.

### E. OS and shell edge cases (easy to miss)

1. Shell command hash caches:
- zsh/bash may cache command path; validate after rename with `hash -r`/new shell.

2. PATH shadowing:
- Ensure old `stage` in `/usr/local/bin`, `~/.local/bin`, or custom toolchains does not shadow `stage` tests.

3. Case sensitivity:
- Verify scripts do not assume `StageServe`/`stageserve` path casing incorrectly on case-sensitive filesystems.

4. xattr/quarantine on macOS:
- If distributing binaries directly, re-check notarization/quarantine instructions still apply with renamed artifacts.

5. Completion cache residue:
- Old completion files in user directories can cause confusing stale behavior.

### F. User data and state safety

1. Runtime state folder policy:
- Decide whether `.stageserve-state` stays unchanged.
- If renamed later, require explicit migration logic, never silent destructive moves.

2. Config-file discovery:
- If config filename changes, support dual-read during transition:
	- read new file first
	- fall back to old file
	- emit one-time migration warning

3. DNS/TLS assets:
- Confirm local cert, host mapping, and resolver setup commands remain valid after rename.

### G. Interoperability and ecosystem checks

1. Command namespace recheck immediately before cut:
- `which stage`
- `brew search stage`
- spot-check npm/pip/cargo registry CLI name collisions

2. Searchability and discoverability:
- Use StageServe-only branding in active README and repo description.

3. Future Homebrew formula reservation:
- If planning formula publication, verify `stage` naming strategy early to avoid future formula conflict surprises.

### H. Compatibility posture

1. No shim:
- Do not ship any `stacklane` forwarding path.
- Keep `stage` as the only supported executable.

2. Completion policy:
- Generate completions only for `stage`.

3. Validation:
- Verify no active docs, scripts, or build targets still rely on `stacklane`.

### I. Validation plan (must-pass before merge)

1. Unit tests:
- Focus on command parsing/help text/version output and installer logic.

2. Integration smoke tests:
- `stage init`
- `stage up`
- `stage status`
- `stage logs`
- `stage down`

3. Fresh-machine install test:
- New shell, clean PATH, no prior StageServe binary, install and run success.

4. Dirty-machine upgrade test:
- Existing old binary present, run installer, verify expected precedence and no ambiguous behavior.

5. Docs copy/paste audit:
- Run commands exactly as documented to verify no stale references remain.

### J. Rollback plan

1. Keep rollback branch/tag before rename merge.
2. Preserve ability to ship previous binary name quickly if a blocker appears.
3. If a post-rename blocker occurs:
- restore prior installer target
- re-point docs quickstart
- publish short incident note and ETA

### K. Recommended execution order

1. Lock naming contract decisions (Section A)
2. Implement binary/install/help/completions (Sections B + E)
3. Update docs/specs in same change (Section C)
4. Update CI/release pipeline (Section D)
5. Run interoperability checks + validation matrix (Sections G + I)
6. Merge with rollback tag prepared (Section J)

This ordering minimizes partial-state risk where docs, binary name, and install behavior drift out of sync.
