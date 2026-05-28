# Architecture: Target Split For The DNS Project

## Goal

Separate host DNS bootstrap behavior from StageServe's project orchestration so each repo owns a clean, durable responsibility.

## Ownership Boundary

| Area | Standalone DNS Project | StageServe |
|---|---|---|
| DNS provider abstraction | Owns | Consumes |
| `dnsmasq` and resolver inspection/bootstrap | Owns | Consumes |
| Result codes and status messages | Owns | Maps into StageServe UX |
| Preview artifact generation | Owns | Supplies preview root |
| `.env.stageserve` resolution | Does not know about it | Owns |
| Stack defaults and project config precedence | Does not know about it | Owns |
| Onboarding flow and remediation wording | Does not know about it | Owns |
| StageServe command names | Does not know about them | Owns |

## Proposed New Repo Shape

```text
<dns-repo>/
├── go.mod
├── cmd/
│   └── <dns-cli>/
│       └── main.go
├── pkg/
│   └── localdns/
│       ├── request.go
│       ├── result.go
│       ├── preview.go
│       ├── provider_darwin.go
│       └── provider_other.go
├── internal/
│   └── cli/
├── docs/
└── .github/
    └── workflows/
```

The exact repo name can change, but the project should separate the public library surface from CLI internals.

## Proposed StageServe Shape After The Split

```text
cmd/stage/commands/
├── dnssetup.go              # remains, but calls the dependency

core/onboarding/
├── readiness_dns_darwin.go  # remains, but maps dependency status to StepResult

core/config/
├── types.go                 # remains owner of LocalDNS settings

platform/
└── dns/                     # removed after successful dependency switch
```

## Public Library Shape

The dependency should keep its public surface small. A compact, stable contract is better than exposing platform-specific helpers.

### Proposed Types

```go
type Request struct {
	Provider      string
	Suffix        string
	IP            string
	Port          int
	PreviewRoot   string
	ManagedPrefix string
}

type Code string

type Result struct {
	Code    Code
	Message string
	Files   Files
}

type Files struct {
	PreviewConfig   string
	PreviewResolver string
	ManagedConfig   string
	ResolverFile    string
}

type Service interface {
	Status(Request) Result
	Bootstrap(Request) error
	WritePreview(Request) (Files, error)
}
```

### Why This Shape

- `PreviewRoot` replaces StageServe-specific `StateDir`
- `ManagedPrefix` replaces hard-coded `stage-`
- `Files` makes artifact paths observable for both StageServe and standalone CLI output
- the contract stays small enough to version cleanly

## StageServe Adapter Shape

StageServe should add a thin adapter layer that converts `core/config.LocalDNS` into dependency requests.

Responsibilities of the adapter:

- read StageServe config only once
- fill `PreviewRoot` from StageServe state layout
- set StageServe ownership metadata such as managed prefix
- map dependency results to `onboarding.StepResult` and command output

This keeps the dependency generic and prevents DNS logic from spreading across multiple StageServe packages.

## Design Rules

- The dependency must not read `.env.stageserve`.
- The dependency must not know where a StageServe project root lives.
- The dependency must not mention StageServe command names or remediation steps.
- StageServe must not reach into dependency internals beyond the public contract.

## Versioning Rules

- Result codes are contract surface and must be treated as versioned API.
- Additive fields are allowed in minor versions.
- Renaming or removing request fields, result codes, or service methods requires a major version.