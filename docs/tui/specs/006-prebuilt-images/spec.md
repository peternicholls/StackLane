# Feature Specification: Docker Distribution Pre-built Images

**Feature Branch**: `006-prebuilt-images`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Publish pre-built PHP-FPM images to Docker Hub/GHCR for faster stack startup"

## Product Contract *(mandatory)*

Pre-built images are an **optimization**, not a new stack behaviour. The stack MUST remain functionally equivalent whether images are pulled or built locally.

### Determinism and reproducibility

- Given the same stack release and the same `PHP_VERSION`, users MUST be able to obtain an equivalent runtime image via:
  - pulling a tagged pre-built image, OR
  - building locally from the same Dockerfile
- Image selection MUST be deterministic: for a given config, the same image reference MUST be chosen.

### Registry strategy

- The system MAY publish to Docker Hub and/or GHCR.
- Documentation MUST clearly state the canonical registry (primary) and any mirrors.
- If authentication is required, it MUST be optional and documented.

### Module compatibility *(mandatory)*

Pre-built images MUST remain compatible with the planned stack **module** architecture.

- Modules may provide their own Compose definitions and assets, but image selection MUST still follow the deterministic rules in this spec.
- The launcher (CLI/TUI) SHOULD supply image-selection environment variables (e.g. `PHP_VERSION`, optional `PMA_IMAGE`) consistently across modules.
- User-installed modules MAY live in `~/.20i/modules/`, but MUST NOT require embedding registry credentials or secrets inside modules or projects.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start Stack with Pre-built Images (Priority: P1)

As a developer, I want the stack to use pre-built images by default so that startup is fast without waiting for image builds.

**Why this priority**: Faster startup directly improves developer experience and is the core benefit of pre-built images.

**Independent Test**: Run `20i start` with default configuration, verify images are pulled from registry (not built locally), and stack reaches a “ready” state without local image build steps.

**Acceptance Scenarios**:

1. **Given** default stack configuration, **When** the user runs `20i start`, **Then** pre-built PHP-FPM images are pulled from registry  
2. **Given** images already cached locally, **When** the user runs `20i start`, **Then** stack starts in under 15 seconds  
3. **Given** a new PHP version is specified, **When** the user runs `20i start`, **Then** the corresponding pre-built image is pulled  
4. **Given** the user is offline or the registry is unreachable, **When** the user runs `20i start`, **Then** the system offers or automatically applies the configured local-build fallback  

#### Cadence note

- Pre-built images are guaranteed for official stack releases.
- On feature branches or before the next release is cut, configuration may reference a PHP version for which no pre-built image exists yet; in this case, the fallback rules in this spec define the expected behaviour.

---

### User Story 2 - Multi-Architecture Support (Priority: P2)

As a developer using Apple Silicon or ARM-based Linux, I want pre-built images to support my architecture so that I get native performance.

**Why this priority**: ARM support is essential for modern Mac users and cloud deployments.

**Independent Test**: Pull the image on ARM64 machine, verify it runs natively without emulation warnings.

**Acceptance Scenarios**:

1. **Given** an ARM64 machine (M1/M2 Mac), **When** the stack starts, **Then** ARM64 native image is pulled and runs without emulation  
2. **Given** an x64 machine, **When** the stack starts, **Then** AMD64 image is pulled and runs natively  
3. **Given** manifest inspection, **When** viewing image details, **Then** both AMD64 and ARM64 architectures are listed  

---

### User Story 3 - Fall Back to Local Build (Priority: P3)

As a developer, I want the option to build images locally so that I can customize the image or work offline.

**Why this priority**: Fallback ensures users aren't blocked if registry is unavailable or customization is needed.

**Independent Test**: Set configuration to use local build, run `20i start`, verify Dockerfile is built locally.

**Acceptance Scenarios**:

1. **Given** pre-built images are disabled (e.g. `USE_PREBUILT=false`), **When** the user runs `20i start`, **Then** images are built from the local Dockerfile(s)  
2. **Given** the registry is unavailable, **When** pulling fails, **Then** the system falls back to local build if allowed, otherwise fails with an actionable error  
3. **Given** custom Dockerfile modifications, **When** user builds locally, **Then** modifications are included in the running container  

#### Fallback contract

- Fallback MUST be configurable:
  - `USE_PREBUILT=true|false`
  - `ALLOW_LOCAL_BUILD_FALLBACK=true|false`
- If `USE_PREBUILT=true` and `ALLOW_LOCAL_BUILD_FALLBACK=true`, the system MUST fall back automatically (no interactive prompt required).
- If `USE_PREBUILT=true` and `ALLOW_LOCAL_BUILD_FALLBACK=false`, the system MUST fail with an actionable error explaining how to enable fallback.

---

### User Story 4 - Image Versioning Aligned with Stack Releases (Priority: P4)

As a maintainer, I want images published for each stack release so that users can pin to specific versions.

**Why this priority**: Version pinning enables reproducible environments and safe upgrades.

**Independent Test**: Pull image tagged with specific version (e.g., `v2.0.0`), verify it matches the stack configuration from that release.

**Acceptance Scenarios**:

1. **Given** release v2.0.0 is published, **When** images are built, **Then** images are tagged with `v2.0.0` and `latest`  
2. **Given** `PHP_VERSION: 8.4`, **When** viewing available images, **Then** the `php-8.4` tag is available  
3. **Given** the `latest` tag, **When** pulling, **Then** it points to the most recent stable stack release  

#### Tagging strategy *(mandatory)*

Images MUST be published with tags that support both “pinning” and “convenience” use:

- **Release pin**: `vX.Y.Z` (stack release tag)  
- **PHP line**: `php-8.5` (or equivalent) for quick selection  
The canonical PHP-line tag format MUST be `php-<major>.<minor>` (e.g. `php-8.4`, `php-8.5`).  
- **Latest**: `latest` MUST point to the most recent stable stack release  

If both Docker Hub and GHCR are used, tags MUST be consistent across registries.

Where possible, documentation SHOULD encourage pinning by stack release tag (e.g. `vX.Y.Z`) for reproducibility.

---

### Edge Cases

- Requested PHP version has no published pre-built image (must fall back or fail with a clear message)  
- Registry rate limiting or temporary outage (must fall back or fail with actionable guidance)  
- Private registry authentication missing/invalid (must provide actionable guidance)  
- Digest mismatch or corrupted pull (must retry and/or fail safely)  
- Local build succeeds but pre-built image would differ (must document equivalence expectations and supported deltas)  
- Security patch release required out of band (must support republishing images without changing stack behaviour)  

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST publish pre-built PHP-FPM images to a primary registry (Docker Hub and/or GHCR), with optional mirror support  
- **FR-002**: System MUST support multiple PHP versions (8.1, 8.2, 8.3, 8.4, 8.5)  
- **FR-003**: System MUST build multi-architecture images (AMD64 and ARM64)  
- **FR-004**: System MUST tag images using the defined tagging strategy (release pin, PHP line, and latest)  
- **FR-005**: System MUST support configurable fallback to local build when pre-built images are unavailable  
- **FR-006**: System MUST update images on every stack release  
- **FR-007**: Documentation MUST explain pre-built vs custom build options  
- **FR-008**: System MUST provide image checksums/digests for verification  
- **FR-009**: Release documentation MUST provide the image digest for each published tag (or a verifiable reference to it)  

### Key Entities

- **Pre-built Image**: Docker image published to registry, ready to pull without local build  
- **Image Tag**: Version identifier (e.g., `8.4`, `v2.0.0`, `latest`) for pulling specific image versions  
- **Multi-arch Manifest**: Docker manifest listing available architectures for a single image tag  

## Non-goals *(mandatory)*

- This feature does NOT change runtime semantics of the stack; it only changes how images are obtained.  
- This feature does NOT publish private secrets or production credentials.  
- This feature does NOT require users to have registry authentication unless they choose a private registry.  
- This feature does NOT define the module system itself; it only guarantees that pre-built image selection remains module-agnostic and deterministic.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: First stack startup time reduced by 80% compared to local build (from ~5 minutes to ~1 minute)  
- **SC-002**: Subsequent startups complete in under 15 seconds with cached images  
- **SC-003**: 100% of supported PHP versions have pre-built images available  
- **SC-004**: Images available for both AMD64 and ARM64 architectures  
- **SC-005**: Image size is tracked per release; regressions are identified and justified (target < 700MB per PHP version unless extensions require more)  
- **SC-006**: Images are updated within 24 hours of stack releases  

## Assumptions

- Docker Hub or GHCR provides sufficient bandwidth for image pulls  
- Users have internet connectivity for initial image pull  
- Multi-arch build process is reliable and reproducible  
- Image tag naming follows Docker conventions
