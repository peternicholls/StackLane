# Feature Specification: Extract Local DNS Into A Standalone Project

**Feature Branch**: `010-extract-local-dns-project`  
**Created**: 2026-05-28  
**Status**: Draft  
**Input**: User description: "Split the DNS side of StageServe into its own project so StageServe depends on it, while making the DNS system useful on its own for users who do not use StageServe."

## Decision Overrides

- Phase 1 uses a versioned Go module dependency inside the StageServe binary. A separately installed DNS binary may exist, but it is not a required dependency for StageServe users.
- StageServe keeps `stage dns-setup` as its user-facing command surface. The extraction changes the implementation boundary, not the StageServe command contract.
- The new DNS project's public API and CLI must use neutral naming. StageServe-specific file names, comments, and UX text must not leak into the generic public surface.
- macOS feature parity is required for the first extraction milestone. Linux and other platforms may preserve the current explicit unsupported behavior if that behavior is documented and tested.

## Problem Statement

StageServe currently owns host-level DNS bootstrap code under `platform/dns`, but most of that code is not inherently tied to StageServe's project lifecycle. The subsystem configures local development DNS via `dnsmasq`, writes resolver artifacts, and reports typed readiness states. That capability is valuable on its own, yet it is trapped inside the StageServe repository and naming model.

The extraction must produce a reusable DNS project without causing StageServe to regress. StageServe should still own project configuration, onboarding flow, remediation copy, and state layout. The new dependency should own DNS provider logic, artifact generation, OS-specific bootstrap, and standalone CLI behavior.

## User Scenarios And Testing

### User Story 1 - StageServe Depends On A Narrow DNS Module (Priority: P1)

As a StageServe maintainer, I can replace the in-repo DNS implementation with a versioned dependency so DNS behavior can evolve independently.

**Why this priority**: This is the core architectural goal. Without it, the split has not actually happened.

**Independent Test**: StageServe compiles and its DNS-related tests pass after removing the in-repo `platform/dns` implementation in favor of the dependency.

**Acceptance Scenarios**:

1. **Given** StageServe loads `LOCAL_DNS_*` settings, **When** it runs `stage dns-setup`, **Then** it builds a request for the dependency rather than calling an in-repo provider implementation.
2. **Given** StageServe runs machine readiness on macOS, **When** DNS status is checked, **Then** the readiness surface consumes the dependency's typed status/result codes.
3. **Given** StageServe is built against a tagged dependency version, **When** the build runs, **Then** no copied DNS implementation remains under `platform/dns`.

### User Story 2 - StageServe Operator Workflow Stays Stable (Priority: P1)

As a StageServe operator, I keep the same StageServe command and readiness flow even though the DNS implementation moved out.

**Why this priority**: The split should not create user-visible churn for StageServe's current workflow.

**Independent Test**: On macOS, `stage dns-setup` and the DNS readiness check produce the same practical behavior before and after the extraction.

**Acceptance Scenarios**:

1. **Given** local DNS is already configured, **When** the operator runs `stage dns-setup`, **Then** StageServe reports that DNS is already ready.
2. **Given** local DNS is not configured, **When** the operator runs `stage dns-setup`, **Then** StageServe bootstraps local DNS through the dependency and reports success or a typed failure.
3. **Given** DNS readiness is missing during onboarding, **When** StageServe reports the blocker, **Then** remediation still points to `stage dns-setup --site-suffix <suffix>` or the agreed guided setup flow.

### User Story 3 - External Users Can Use The DNS Project Without StageServe (Priority: P2)

As a developer who does not use StageServe, I can use the standalone DNS project directly to inspect and bootstrap local development DNS.

**Why this priority**: This is the business and product reason for the split.

**Independent Test**: A user can install the standalone CLI, point it at a suffix/IP/port, and use it without any StageServe config or state.

**Acceptance Scenarios**:

1. **Given** the standalone DNS CLI is installed, **When** the user runs its status command, **Then** it reports typed local DNS state without needing StageServe files.
2. **Given** the user runs the standalone bootstrap command, **When** the platform is supported, **Then** it writes or installs the required DNS artifacts using generic naming.
3. **Given** the platform is unsupported, **When** the standalone CLI runs, **Then** it returns a clear unsupported result instead of silently succeeding.

### User Story 4 - DNS Can Release Independently (Priority: P2)

As a maintainer, I can version, test, and release the DNS project independently from StageServe.

**Why this priority**: Independent release cadence is one of the main benefits of the split.

**Independent Test**: The DNS project can publish a tagged release, and StageServe can upgrade to that tag with a normal dependency update.

**Acceptance Scenarios**:

1. **Given** the DNS project tags a new release, **When** StageServe updates its module version, **Then** the change is isolated to the dependency update and any planned adapter changes.
2. **Given** the DNS project needs a breaking API change, **When** that change is proposed, **Then** it uses a major-version contract rather than silent drift.

## Edge Cases

- StageServe and the DNS project drift on status code names.
- The extracted API still assumes StageServe-owned file naming such as `stage-<suffix>.conf`.
- The standalone CLI requires StageServe state paths or config files.
- Unsupported-platform behavior diverges between StageServe and the standalone project.
- StageServe users are forced to install a second binary to keep existing commands working.
- The dependency needs caller-provided preview paths, but the contract hard-codes `.stageserve-state`.
- Version skew causes StageServe readiness logic to mis-handle new or renamed DNS result codes.
- The extraction accidentally widens scope into TLS, proxy, or generic onboarding work.

## Operational Impact

### StageServe Impact

- Affected StageServe surfaces: `stage dns-setup`, DNS readiness checks, docs mentioning local DNS bootstrap, and any tests covering DNS provider behavior.
- StageServe remains the owner of `LOCAL_DNS_PROVIDER`, `LOCAL_DNS_IP`, `LOCAL_DNS_PORT`, and `LOCAL_DNS_SUFFIX` resolution.
- StageServe remains the owner of project remediation language and onboarding flow.

### Standalone DNS Project Impact

- New public surface: versioned Go module and standalone CLI.
- New docs: install, supported platforms, bootstrap semantics, and privilege expectations.
- New release work: semantic versioning, CI, and optional distribution channels such as Homebrew.

## Requirements

### Functional Requirements

- **FR-001**: The new DNS project MUST own the host DNS provider abstraction, status codes, preview artifact generation, and OS-specific bootstrap logic currently implemented under `platform/dns`.
- **FR-002**: StageServe MUST consume the extracted DNS project as a versioned dependency rather than keeping a copied provider implementation in-tree.
- **FR-003**: StageServe MUST preserve `stage dns-setup` as a first-party StageServe command that delegates to the dependency.
- **FR-004**: StageServe DNS readiness checks MUST consume the dependency's typed status contract rather than reaching into provider-specific internals.
- **FR-005**: The extracted DNS project MUST expose a standalone CLI so non-StageServe users can inspect and bootstrap supported DNS setups.
- **FR-006**: The extracted DNS project's public API MUST use generic naming and caller-supplied ownership metadata instead of StageServe-specific names.
- **FR-007**: The first extracted release MUST preserve current macOS behavior: Homebrew/dnsmasq inspection, managed config generation, resolver installation, and typed readiness states.
- **FR-008**: Unsupported-platform behavior in the extracted project MUST remain explicit and testable.
- **FR-009**: Preview artifact locations and managed-file prefixes MUST be caller-configurable so StageServe can keep its state layout without making that layout part of the generic DNS contract.
- **FR-010**: StageServe MUST retain ownership of `.env.stageserve`, stack defaults, project state, onboarding sequencing, and user-facing remediation copy.
- **FR-011**: Both repositories MUST document the split clearly: the DNS repo documents its standalone usage, and StageServe documents that DNS behavior is dependency-backed.
- **FR-012**: The dependency contract MUST be versioned with semantic versioning and a stable code vocabulary.
- **FR-013**: StageServe MUST be able to develop against the new DNS repo locally before release by using a local module replacement workflow.
- **FR-014**: The extraction plan MUST include a rollback path that can restore the in-repo implementation if the split reveals blocking integration issues before the dependency is tagged.

### Non-Functional Requirements

- **NFR-001**: StageServe's DNS-related command latency SHOULD remain materially unchanged after the dependency swap.
- **NFR-002**: The extracted DNS project SHOULD be testable in isolation without requiring StageServe config files or project directories.
- **NFR-003**: The split SHOULD minimize StageServe file churn by concentrating integration changes in a small adapter seam.
- **NFR-004**: The extracted public API SHOULD remain small enough that StageServe can upgrade it with low-friction version bumps.

## Out Of Scope

- Expanding Linux automation beyond the current explicit unsupported surface.
- Folding TLS or mkcert work into the DNS project.
- Reworking StageServe's broader onboarding/TUI flow as part of this split.
- Replacing `stage dns-setup` with a differently named StageServe command.
- Shipping a background daemon or long-running service in the first extraction milestone.

## Key Entities

- **DNS Request**: The caller-provided desired local DNS configuration, including suffix, IP, port, provider, preview root, and ownership metadata.
- **DNS Result Code**: Stable typed state returned by the dependency, such as `ready`, `dnsmasq-missing`, or `unsupported-os`.
- **Managed Artifact**: A generated config file or resolver file preview that the dependency creates or installs.
- **StageServe DNS Adapter**: The thin integration layer that converts `core/config.LocalDNS` into dependency requests and maps dependency results back into StageServe command or readiness surfaces.
- **Standalone DNS CLI**: The new project's direct operator surface for status/bootstrap/preview behavior outside StageServe.

## Success Criteria

- **SC-001**: StageServe no longer contains the active DNS provider implementation in-tree and instead builds against a tagged external dependency.
- **SC-002**: `stage dns-setup` retains the same practical operator behavior on macOS after the extraction.
- **SC-003**: The standalone DNS CLI can inspect and bootstrap supported local DNS behavior without StageServe files.
- **SC-004**: The DNS dependency can be versioned and upgraded independently from StageServe.