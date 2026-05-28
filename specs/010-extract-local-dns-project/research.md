# Research: Standalone DNS Project Extraction

## Current StageServe DNS Surface

The current extraction candidate is already fairly tight.

### Code That Looks Extractable

- `platform/dns/types.go`
- `platform/dns/common.go`
- `platform/dns/macos.go`
- `platform/dns/linux.go`

### StageServe-Owned Integration Points

- `cmd/stage/commands/dnssetup.go` builds settings and runs bootstrap
- `core/onboarding/readiness_dns_darwin.go` runs status for readiness output
- `core/config.LocalDNS` resolves user and stack configuration

### StageServe-Specific Coupling That Must Be Removed Or Parameterized

- Managed config naming currently uses `stage-<suffix>.conf`
- Preview artifacts currently live under `.stageserve-state/shared`
- Comments and messages reference StageServe in code comments and workflow assumptions
- The current API shape uses `StateDir`, which is really a caller-owned preview root

## Extraction Options

| Option | Shape | Advantages | Risks | Verdict |
|---|---|---|---|---|
| A | Extract a library only | Smallest implementation change for StageServe | Does not satisfy the goal of making DNS useful on its own | Rejected |
| B | Extract a standalone CLI only and make StageServe shell out | Strong standalone story | Adds a second required binary for StageServe users, complicates install/version skew | Rejected for phase 1 |
| C | Extract a new repo that ships both a Go module and a small CLI | Preserves easy StageServe integration and creates a real standalone product | Slightly more setup work in the new repo | Recommended |

## Recommended Approach

Create a new repo that exposes:

- a small Go module for status/bootstrap/preview operations
- a standalone CLI for direct use outside StageServe
- neutral branding and generic caller-owned configuration

StageServe should consume the module directly inside its existing binary. That keeps `stage dns-setup` stable and avoids pushing a second install step onto current users. The standalone CLI then provides the independent-product value the split is meant to unlock.

## Recommended Boundary

### New DNS Project Owns

- status codes and result vocabulary
- preview artifact generation
- OS-specific bootstrap/inspection behavior
- provider selection and unsupported-platform semantics
- standalone DNS CLI
- standalone docs, CI, releases, and distribution strategy

### StageServe Continues To Own

- reading `.env.stageserve` and stack defaults
- deciding `LOCAL_DNS_*` values
- readiness sequencing and remediation copy
- `.stageserve-state` layout and preview-root location
- StageServe command names and help text

## Naming Guidance

The new project should not carry StageServe branding in its public API unless the strategic decision is to make it a StageServe subproject. A neutral working name is preferable because the stated goal is utility beyond StageServe.

Recommended naming rule:

- repo/module/CLI name should describe local development DNS, not StageServe
- ownership metadata should be caller-supplied so StageServe can still mark its managed files cleanly

## API Design Observations

The current `Settings` type is close to reusable, but it still exposes StageServe assumptions through `StateDir` and the managed file naming convention.

The extracted contract should instead center on a generic request such as:

```go
type Request struct {
	Provider      string
	Suffix        string
	IP            string
	Port          int
	PreviewRoot   string
	ManagedPrefix string
}
```

The extracted library can then generate preview files and determine managed file paths without learning anything about StageServe itself.

## Risk Review

### Risk: Generic API Still Smells Like StageServe

If the new API keeps `StateDir`, `stage-*.conf`, or StageServe-specific copy, the split will be only nominal.

Mitigation: rename those concepts before the external dependency is tagged.

### Risk: StageServe Gets A Hard External Runtime Dependency

If StageServe shells out to a separate binary in phase 1, installation and support complexity increase immediately.

Mitigation: keep the first integration path as a Go module dependency; standalone CLI is additive.

### Risk: Status Code Drift Breaks Readiness Mapping

StageServe's onboarding surfaces already depend on discrete DNS states.

Mitigation: freeze the result-code vocabulary in a contract document and treat changes as versioned API work.

## Recommendation Summary

Proceed with a dual-surface extraction:

1. New repo with generic DNS library and standalone CLI
2. StageServe consumes the library directly
3. StageServe keeps its existing command and readiness UX
4. Tag a prerelease only after StageServe has validated the dependency through a local module replacement