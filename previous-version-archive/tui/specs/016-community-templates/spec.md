# Feature Specification: Community Templates Repository

**Feature Branch**: `016-community-templates`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "User-contributed project templates for frameworks like Drupal, Magento, and Shopware"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Use Community Template for Project (Priority: P1)

As a developer working with a less common framework, I want to use a community-contributed template so that I can get started quickly without creating configuration from scratch.

**Why this priority**: Using templates is the primary value - if templates can't be used, the feature provides no value.

**Independent Test**: Run `20i init --template community/drupal`, verify Drupal-specific configuration is created and stack starts successfully.

**Acceptance Scenarios**:

1. **Given** `20i init --template community/drupal`, **When** initialization completes, **Then** Drupal-optimized stack configuration is created
2. **Given** community template is used, **When** stack starts, **Then** all framework-specific services are available
3. **Given** `20i templates --community`, **When** listing templates, **Then** all available community templates are shown

---

### User Story 2 - Contribute New Template via PR (Priority: P2)

As an experienced developer, I want to contribute a template for my favorite framework so that others can benefit from my configuration.

**Why this priority**: Contribution flow enables template growth; without it, the repository stagnates.

**Independent Test**: Create a Shopware template following contribution guidelines, submit PR, verify template structure is validated.

**Acceptance Scenarios**:

1. **Given** contribution guidelines, **When** developer creates template, **Then** required structure and files are documented
2. **Given** template PR is submitted, **When** CI runs, **Then** template structure is automatically validated
3. **Given** template PR is merged, **When** next release occurs, **Then** template is available via CLI

---

### User Story 3 - View Template Documentation (Priority: P3)

As a developer evaluating templates, I want to read template documentation so that I understand what's included before using it.

**Why this priority**: Documentation helps users make informed decisions about template selection.

**Independent Test**: Run `20i templates community/magento --info`, verify template description, included services, and maintainer info are displayed.

**Acceptance Scenarios**:

1. **Given** a community template, **When** user requests info, **Then** description, included services, and requirements are shown
2. **Given** template README, **When** viewing on GitHub, **Then** setup instructions and configuration options are documented
3. **Given** template metadata, **When** listing templates, **Then** maintainer and last updated date are visible

---

### User Story 4 - Report Template Issues (Priority: P4)

As a template user encountering problems, I want to report issues to the template maintainer so that problems can be fixed.

**Why this priority**: Issue reporting enables community maintenance of templates.

**Independent Test**: Find issue template in repository, submit issue for specific community template, verify maintainer is notified.

**Acceptance Scenarios**:

1. **Given** issue with community template, **When** user opens GitHub issue, **Then** template-specific issue template is available
2. **Given** issue is submitted, **When** template maintainer is tagged, **Then** they receive notification
3. **Given** `20i templates community/drupal --report`, **When** command runs, **Then** link to issue creation page is provided

---

### Edge Cases

- What happens when a community template becomes unmaintained?
- How does the system handle templates that conflict with core functionality?
- What happens when a template's dependencies are no longer available?
- How are security issues in community templates handled?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support community templates in `templates/community/` directory
- **FR-002**: CLI MUST support `--template community/<name>` flag for initialization
- **FR-003**: CLI MUST support `--community` flag to list community templates
- **FR-004**: Each template MUST include README with setup instructions
- **FR-005**: Each template MUST include metadata file (maintainer, version, description)
- **FR-006**: Repository MUST include contribution guidelines for new templates
- **FR-007**: CI MUST validate template structure on PR submission
- **FR-008**: Templates MUST be versioned alongside stack releases
- **FR-009**: System MUST clearly distinguish community templates from official templates

### Key Entities

- **Community Template**: User-contributed configuration set for a specific framework, with attributes: name, maintainer, description, included services
- **Template Metadata**: Required information about a template including maintainer contact, version, and compatibility requirements
- **Contribution Guidelines**: Documentation describing how to create and submit new templates

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: At least 5 community templates available within 6 months of feature launch
- **SC-002**: Community templates function identically to official templates (100% feature parity)
- **SC-003**: Template contribution process takes less than 30 minutes for experienced users
- **SC-004**: 80%+ of submitted templates pass automated validation on first submission
- **SC-005**: Community template issues receive response within 1 week

## Assumptions

- Community members are willing to contribute and maintain templates
- GitHub PR workflow is familiar to potential contributors
- Template validation can be automated reliably
- Maintainers will respond to issues in reasonable timeframes

# Feature Specification: Working with 20i One‑Click Installs & Existing Sites (Docs‑First)

**Feature Branch**: `016-working-with-existing-sites`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟢 Medium  
**Input**: User description: "Guide users through bringing real 20i-hosted sites (including one‑click installs) into local development using the 20i stack"

## Product Contract *(mandatory)*

This feature is **documentation and guidance only** in the MVP.

- 20i itself remains the **canonical installer** for applications (e.g. one‑click installers, manual uploads).
- The 20i Docker stack does **not** re‑implement installers, deployment pipelines, or hosting automation.
- This documentation describes **supported, realistic workflows** for cloning and working on existing 20i sites locally.

### Scope

- Applies to applications installed on 20i shared hosting.
- Covers cloning files and databases into local development.
- Covers safe configuration changes required for local use.

### Non‑goals

- This feature does NOT provide framework templates or installers.
- This feature does NOT automate production deployments.
- This feature does NOT modify live production sites.
- This feature does NOT manage credentials or secrets.

## Supported Workflows *(mandatory)*

### Workflow A: One‑Click Install → Local Development

**Use case**: User installs an application (e.g. WordPress, Drupal) via the 20i one‑click installer and wants to work on it locally.

**Steps**:

1. Install the application on 20i using the My20i one‑click installer.
2. Access the site files via SSH or SFTP (typically under `public_html/`).
3. Download the application files to a local project directory.
4. Export the production database using `mysqldump` or phpMyAdmin.
5. Import the database into the local MariaDB instance.
6. Adjust application configuration for local development (database host, credentials, base URL).
7. Start the local stack using `20i`.

**Notes**:
- Treat the production site as read‑only during cloning.
- Avoid copying production secrets into local configuration.

---

### Workflow B: Existing Live Site → Local Development

**Use case**: User already has a live site hosted on 20i and wants a local development copy.

**Steps**:

1. Connect to the live site via SSH or SFTP.
2. Clone files using `scp`, `rsync`, or FTP.
3. Export the database.
4. Import files and database locally.
5. Update environment‑specific configuration.
6. Verify site runs correctly under the local stack.

**Notes**:
- URL rewriting may be required for CMS platforms.
- File permissions may differ between local and hosted environments.

---

### Workflow C: Ongoing Sync (Manual)

**Use case**: User needs to periodically refresh local data from production.

**Guidance**:

- Re‑export database snapshots as needed.
- Re‑sync uploads/media directories selectively.
- Avoid overwriting local configuration files.

Automation of this process is intentionally out of scope.

## SSH / File Transfer Guidance *(mandatory)*

Documentation SHOULD include examples of:

- `scp` and `rsync` commands
- SFTP client configuration
- Common directory layouts on 20i (e.g. `public_html/`)

Example (illustrative only):

```bash
rsync -avz user@server:/home/username/public_html/ ./site/
```

## Database Handling Guidance *(mandatory)*

Documentation SHOULD include:

- Exporting databases via `mysqldump`
- Importing into local MariaDB
- Handling large databases safely
- Adjusting application config for local DB access

## Configuration & Secrets *(mandatory)*

- Production credentials MUST NOT be committed to version control.
- Local development SHOULD use environment‑specific configuration (`.env`, framework config files).
- Production secrets remain managed by 20i and hosting providers.

(See spec 013 for detailed guidance.)

## Parity & Known Differences

This feature MUST reference **Deployment Parity & Known Differences** (spec 015) for:

- HTTPS behaviour
- Mail delivery
- Cron and background jobs
- Resource limits
- PHP defaults

## Edge Cases

- Sites installed outside `public_html/`
- Multiple applications under one hosting package
- Large media libraries
- CMS plugins/themes requiring special PHP extensions
- Legacy applications with hardcoded paths or URLs

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC‑001**: Users can successfully clone an existing 20i‑hosted site into local development.
- **SC‑002**: Documentation clearly distinguishes safe local changes from production‑only configuration.
- **SC‑003**: No automation is implied that could risk live production data.

## Assumptions

- Users have SSH or SFTP access to their 20i hosting.
- Users understand basic file transfer and database concepts.
- One‑click installers remain the preferred production install method on 20i.

---