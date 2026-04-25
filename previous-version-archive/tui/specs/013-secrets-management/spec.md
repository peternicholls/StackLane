# Feature Specification: Secrets Management Guidance

**Feature Branch**: `013-secrets-management`  
**Created**: 2025-12-28  
**Status**: Concept / Documentation-Oriented  
**Priority**: 🟡 High  
**Input**: User description: "Clarify best practices for managing secrets in 20i-style projects without the stack owning or resolving secrets"

## Product Contract *(mandatory)*

Secrets management is **outside the responsibility of the 20i stack runtime**. The stack provides defaults and expects the application to source secrets appropriately.

### Scope

- Secrets MUST live in the **application/project layer**, not in the 20i stack implementation.
- The 20i stack MUST NOT fetch, decrypt, or resolve secrets from external providers at runtime.
- The stack MAY expose environment variable pass-through mechanisms only.

### Environment separation

- Developers SHOULD use environment-specific files such as `.env`, `.env.dev`, `.env.local`, or framework-specific mechanisms.
- Production secrets handling is the responsibility of the hosting platform (e.g. 20i shared hosting, CI/CD, provider dashboards).

### Security stance

- The 20i project SHOULD document secure patterns but MUST NOT become a secret manager.

## User Scenarios & Guidance *(mandatory)*

### User Story 1 - Reference Secrets from External Provider (Priority: P1)

> This scenario describes recommended usage patterns and is not implemented by the 20i stack itself.

As a security-conscious developer, I want to reference secrets from an external provider so that credentials are never stored in plain text in my project files.

**Why this priority**: This is the core security value - eliminating plain-text credentials from project files.

**Acceptance Scenarios**:

1. **Given** config references `op://vault/item/field`, **When** stack starts, **Then** actual secret value is fetched and used  
2. **Given** external secret is updated, **When** stack restarts, **Then** new secret value is used  
3. **Given** secret reference is invalid, **When** stack starts, **Then** clear error message indicates which secret failed

---

### User Story 2 - Support Multiple Secret Providers (Priority: P2)

> This scenario describes recommended usage patterns and is not implemented by the 20i stack itself.

As a developer on a team, I want to use my team's preferred secret provider so that I can integrate with existing security infrastructure.

**Why this priority**: Flexibility in providers enables adoption across different team setups.

**Acceptance Scenarios**:

1. **Given** 1Password CLI is configured, **When** secrets reference 1Password, **Then** secrets are fetched via `op` CLI  
2. **Given** AWS credentials are configured, **When** secrets reference AWS Secrets Manager, **Then** secrets are fetched via AWS SDK  
3. **Given** provider is not available, **When** stack starts, **Then** error suggests installing/configuring the provider

---

### User Story 3 - Fallback to Local .env (Priority: P3)

> This scenario describes recommended usage patterns and is not implemented by the 20i stack itself.

As a developer working offline, I want the system to fall back to local .env values so that I can work without network access to secret providers.

**Why this priority**: Fallback ensures developers aren't blocked when providers are unavailable.

**Acceptance Scenarios**:

1. **Given** external provider is unavailable, **When** fallback is configured, **Then** local .env values are used with warning  
2. **Given** no fallback configured, **When** provider is unavailable, **Then** stack fails to start with clear error  
3. **Given** `--offline` flag, **When** stack starts, **Then** provider is not contacted and local values are used

---

### User Story 4 - Encrypted Local Secrets (Priority: P4)

> This scenario describes recommended usage patterns and is not implemented by the 20i stack itself.

As a solo developer, I want to encrypt my .env file so that credentials are protected without needing an external provider.

**Why this priority**: Local encryption provides security benefits for users without external infrastructure.

**Acceptance Scenarios**:

1. **Given** encrypted .env file, **When** stack starts, **Then** user is prompted for decryption passphrase  
2. **Given** correct passphrase entered, **When** decryption succeeds, **Then** stack starts with decrypted values  
3. **Given** `20i secrets encrypt`, **When** command runs, **Then** plain-text .env is encrypted and original is removed

---

### Edge Cases

- Secrets accidentally committed to version control (mitigate with `.gitignore` and secret scanning)
- Divergence between development and production secret handling
- Lost or rotated credentials requiring application reconfiguration
- Team members using different secret providers locally

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST document recommended patterns for managing secrets in development and production  
- **FR-002**: Project MUST clearly state that the 20i stack does not own or resolve secrets  
- **FR-003**: Documentation MUST include examples of environment-specific `.env` usage  
- **FR-004**: Documentation MUST include guidance on avoiding committing secrets to version control  

### Key Entities

- **Application Secrets**: Credentials and sensitive values consumed by the web application  
- **Environment Files**: Files such as `.env`, `.env.dev`, `.env.local` used by the application  
- **Hosting Provider Secrets**: Secrets managed externally by hosting platforms or CI/CD systems  

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Documentation clearly explains where secrets should live and why  
- **SC-002**: Developers can configure local and production secrets without modifying the stack  
- **SC-003**: No secret-handling code exists inside the 20i stack runtime  

## Assumptions

- Developers manage secrets via application frameworks or hosting providers  
- Different environments (dev, staging, production) use different secret sources  
- The 20i stack is used purely for local parity and does not replicate provider secret systems  
