# Feature Specification: Video Walkthroughs & Screen Scripts (Docs-First)

**Feature Branch**: `020-video-tutorials`  
**Created**: 2025-12-28  
**Status**: Draft  
**Priority**: 🟡 Low (Documentation & Education)  
**Input**: User description: "Screen-based walkthroughs and narrated demonstrations to complement written documentation"

## Product Contract *(mandatory)*

This feature is **documentation-first** and focuses on **screen script writing**, not software implementation.

- Videos are optional educational artefacts, not required for correct use of the tool.
- Written documentation remains the **canonical source of truth**.
- Videos MUST mirror documented workflows and MUST NOT introduce undocumented behaviour.
- This feature does NOT block releases or gate functionality.

## Scope

- Short, narrated screen recordings demonstrating real workflows
- Content derived directly from existing documentation and CLI/TUI behaviour
- Hosted externally and linked from documentation

## Non-goals *(mandatory)*

- This feature does NOT guarantee video coverage for all features
- This feature does NOT require professional production values
- This feature does NOT define analytics, engagement metrics, or success KPIs
- This feature does NOT replace written documentation

## Content Categories *(illustrative)*

### Category A: Getting Oriented

- What the tool is (and is not)
- Typical workflow overview
- Project detection and stack start

### Category B: Core Workflows
 
- Starting and stopping the stack
- Viewing status and logs
- Enabling/disabling optional services

### Category C: Realistic Scenarios

- Cloning an existing 20i site into local development (ref spec 016)
- Understanding parity differences (ref spec 015)
- Common mistakes and recovery paths

### Category D: Troubleshooting Narratives

- Docker not running
- Port conflicts
- Permission issues

## Screen Script Writing Guidance *(mandatory)*

All videos SHOULD be based on a **written screen script** before recording.

A screen script SHOULD include:

- Purpose of the video
- Starting state (filesystem, running/stopped stack)
- Step-by-step on-screen actions
- Expected output and narration cues
- Explicit call-outs when behaviour differs from production (parity notes)

Scripts SHOULD be stored alongside documentation (e.g. `docs/scripts/`).

## Versioning & Drift *(mandatory)*

- Each video MUST state the tool version it was recorded against.
- Documentation MUST note when a video may be outdated.
- Videos SHOULD be re-recorded only when workflows materially change.

## Accessibility *(mandatory)*

- Videos MUST include captions or subtitles where reasonably possible.
- Videos SHOULD be skippable and concise.
- Text alternatives MUST exist for all demonstrated workflows.

## Distribution & Linking

- Videos MAY be hosted on a public platform (e.g. YouTube).
- Documentation SHOULD link to videos contextually rather than embedding aggressively.
- Offline or text-only users MUST not be disadvantaged.

## Success Criteria *(qualitative)*

- **SC-001**: Videos help users understand workflows faster, without replacing docs.
- **SC-002**: Videos do not create confusion or contradict written guidance.
- **SC-003**: Maintaining videos does not block or slow feature development.

## Assumptions

- Maintainers are comfortable producing informal screen recordings.
- Users vary in learning preference; video is supplemental.
- Scripts can be reused or adapted for blog posts or talks.

---