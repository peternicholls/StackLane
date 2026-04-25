# Data Model: Workflow And Lifecycle Hardening

## Bootstrap Contract

- Purpose: Represents the project-scoped declaration of whether Stacklane runs a post-up bootstrap command and how that command participates in lifecycle success or failure.
- Fields:
  - `source`: fixed to `.stacklane-local`
  - `command`: string value from `STACKLANE_POST_UP_COMMAND`
  - `phase`: fixed to `post-up`
  - `execution_target`: fixed to the `apache` service container
  - `working_directory`: project site root inside the container
  - `failure_mode`: fixed to rollback
  - `failure_step_name`: lifecycle step label used in operator-visible errors
- Relationships:
  - Belongs to one project runtime.
  - Depends on Stacklane-owned readiness succeeding first.
  - Produces either a successful post-bootstrap runtime or a rollback outcome.
- Validation rules:
  - The contract must not be sourced from stack-wide config.
  - Absence of a bootstrap command must not trigger implicit framework-specific behavior.
  - Failure must be classified separately from gateway, DNS, and container-health failures.

## Stack Defaults Contract

- Purpose: Represents stack-owned defaults that apply across projects without taking ownership of project application config.
- Fields:
  - `file_name`: `.env.stacklane`
  - `scope`: stack-owned defaults only
  - `precedence_rank`: below shell environment, above built-in defaults
  - `allowed_keys`: stack runtime defaults such as shared gateway, DNS, and runtime defaults
- Relationships:
  - Feeds the config loader.
  - Must remain distinct from both `.stacklane-local` and project `.env`.
- Validation rules:
  - The file must be the only stack-owned defaults file in the supported contract.
  - It must not be documented as application config.

## Runtime Naming Contract

- Purpose: Represents the names Stacklane derives for project-scoped runtime resources.
- Fields:
  - `project_prefix`: `stln-`
  - `compose_project_name`: `stln-<slug>` by default
  - `runtime_network`: `<compose-project>-runtime`
  - `database_volume`: `<compose-project>-db-data`
  - `web_network_alias`: derived from the compose project name and service role
- Relationships:
  - Derived from the project slug and config loader defaults.
  - Used by compose invocations, gateway upstream routing, Docker label lookup, and operator-facing status output.
- Validation rules:
  - Project-scoped runtime names must use the shortened prefix consistently.
  - Shared resources require their own explicit naming rule and must not be conflated with project-scoped naming.

## Validation Scenario

- Purpose: Represents a documented real-world workflow used to prove lifecycle behavior.
- Fields:
  - `scenario_type`: `single-project` or `multi-project`
  - `projects_under_test`: list of representative repos or fixtures
  - `commands`: ordered lifecycle commands under test
  - `checks`: DNS, gateway, runtime env, database, bootstrap, status, teardown, rollback
  - `expected_outcomes`: observable operator-visible results for each check
  - `evidence`: test notes, command output summary, or recorded validation artifact
- Relationships:
  - Exercises the bootstrap contract and runtime naming contract together.
  - Feeds quickstart validation and plan completion criteria.
- Validation rules:
  - At least one scenario must cover bootstrap success.
  - At least one scenario must cover bootstrap failure and rollback.
  - At least one scenario must cover attached multi-project routing.

## Failure Classification

- Purpose: Represents the operator-facing boundary between Stacklane infrastructure failures and application-owned failures.
- Fields:
  - `class`: gateway, DNS, readiness, bootstrap, application-follow-up
  - `owner`: Stacklane or application
  - `recovery_path`: rerun, inspect logs, fix app, or reroute through documented workflow
  - `status_effect`: whether runtime remains up or is rolled back
- Relationships:
  - Attached to lifecycle step errors and validation notes.
  - Determines how docs and status output explain what failed.
- Validation rules:
  - Bootstrap failure must be a Stacklane lifecycle class with rollback.
  - Application route defects after readiness must not be misreported as gateway or DNS failures.