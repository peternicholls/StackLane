# StageServe Integration Contract For The DNS Dependency

## Purpose

Define the minimum stable contract between StageServe and the extracted DNS project so the repos can evolve independently without accidental coupling.

## What StageServe Sends

StageServe provides the dependency with resolved DNS intent only. It does not ask the dependency to discover project settings or read StageServe-owned files.

Required request fields:

- `Provider`
- `Suffix`
- `IP`
- `Port`
- `PreviewRoot`
- `ManagedPrefix`

Optional future fields may include platform-specific overrides or alternate resolver directories, but those should be additive rather than required for the first release.

## What The Dependency Returns

The dependency returns typed DNS state using a stable code vocabulary plus an operator-facing message.

Required result fields:

- `Code`
- `Message`

Optional but recommended result fields:

- `Files.PreviewConfig`
- `Files.PreviewResolver`
- `Files.ManagedConfig`
- `Files.ResolverFile`

## Stable Code Vocabulary

The initial code vocabulary should preserve the current practical states:

- `ready`
- `unsupported-os`
- `brew-missing`
- `dnsmasq-missing`
- `dnsmasq-config-missing`
- `dnsmasq-stopped`
- `resolver-missing`
- `resolver-mismatch`
- `unknown`

StageServe is allowed to translate these into StageServe-specific UI copy, but it should not need to infer low-level state from freeform text.

## Explicit Non-Responsibilities Of The Dependency

The dependency must not:

- read `.env.stageserve`
- read stack metadata
- inspect StageServe project directories
- decide StageServe remediation text
- emit StageServe command suggestions

## StageServe Mapping Rules

StageServe should map dependency results into its own surfaces this way:

- `stage dns-setup`: delegate status/bootstrap and print StageServe command output
- readiness checks: convert dependency result into `onboarding.StepResult`
- docs/help: remain StageServe-owned and speak in StageServe terms

## Versioning Rules

- Adding a new result code is a contract change and requires coordinated StageServe review before adoption.
- Removing or renaming a result code requires a major version.
- Adding new optional request fields is a minor-version change.
- Changing required request semantics requires a major version.

## Development Workflow

Before the first tagged release, StageServe should integrate via a local `replace` directive so the real adapter can be tested without prematurely freezing the dependency API.