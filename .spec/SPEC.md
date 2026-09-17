# Mesosphere Specification

## 1. Overview & Mission

Mesosphere is a lightweight, git-native alternative to SaaS product management suites (e.g. Jira, Confluence, Trello). It is built on the premise that emerging AI-driven development patterns — where all project context lives in the repository itself and AI assistants can operate directly on that context — remove the need for separate, hosted project management tooling.

Mesosphere provides a CLI, an Agent Skill (so AI assistants can interact with project docs/tasks programmatically), and a local UI, all operating directly on markdown/text documentation and task files stored in git repositories.

**Source:** Mesosphere.md (supplied product notes)

## 2. Problem Statement

Existing SaaS product management suites store project knowledge (docs, tasks, specs) outside the codebase, in proprietary hosted systems. This creates overhead and friction, especially for AI-driven development where an AI assistant benefits from having all context — code, docs, and tasks — colocated in the repository it operates on. Teams (and AI agents) need a way to manage docs and tasks that live natively in git, are human-reviewable, and are equally accessible to humans (via UI) and AI assistants (via CLI/Agent Skill).

## 3. Goals

- Enable docs and tasks to live as version-controlled markdown/text/JSON files within a project's git repository.
- Provide a unified CLI and UI for humans and AI assistants to read, browse, and (for humans) edit docs and tasks.
- Support multiple repositories and let a user view/compare content across repositories simultaneously.
- Render docs (including mermaid diagrams) and task boards in a way comparable in usability to existing tools (Confluence/Obsidian for docs, Jira/Trello for tasks).
- Make git the system of record: edits are committed and published (pushed) via git operations initiated through the UI/CLI.
- Support multiple task file schemas/frameworks by translating them into a common task representation.

## 4. Non-Goals

- Mesosphere is not a hosted/SaaS product management suite; it does not require or provide central server-hosted storage of project data (repos may be local or remote, but the system does not itself host the source of truth).
- Mesosphere does not define a mandatory task schema/framework; it translates existing/homegrown frameworks rather than replacing them.
- Mesosphere is not itself an AI assistant; it exposes capabilities (via CLI/Agent Skill) for AI assistants to use, but does not perform AI-driven development itself.

## 5. Users / Actors

- **Human contributor/reviewer:** browses and edits docs, reviews and updates tasks via the UI.
- **AI assistant/agent:** consumes CLI and Agent Skill capabilities to list, read, and retrieve docs and task details as part of AI-driven development workflows.
- **Project/repo maintainer:** configures which repos and settings Mesosphere operates against (via config file).

## 6. Functional Requirements

### FR-001 - Configuration File
**Statement:** The system shall support a config file that specifies repository locations and saves UI settings (e.g. active tab).
**Why:** Mesosphere needs to know which repositories to operate on and persist user UI preferences across sessions.
**Acceptance Criteria:**
- Config file can list one or more repo locations (local or remote).
- Config file can persist UI tab/settings state.
**Priority:** Must
**Status:** Accepted

### FR-002 - CLI: Launch UI
**Statement:** The system shall provide a CLI command to launch the UI.
**Priority:** Must
**Status:** Accepted

### FR-003 - CLI: List Repos
**Statement:** The CLI shall provide a command to list configured repositories.
**Priority:** Must
**Status:** Accepted

### FR-004 - CLI: List Repo Docs
**Statement:** The CLI shall provide a command to list the documents contained in a repository.
**Priority:** Must
**Status:** Accepted

### FR-005 - CLI: List Doc Versions
**Statement:** The CLI shall provide a command to list the versions of a given document.
**Priority:** Must
**Status:** Accepted

### FR-006 - CLI: Get Doc Contents
**Statement:** The CLI shall provide a command to retrieve the contents of a document, with an optional argument to retrieve a previous version of that document.
**Acceptance Criteria:**
- Command accepts an optional `version` argument.
- When `version` is omitted, the current version is returned.
- When `version` is supplied, the specified previous version is returned.
**Priority:** Must
**Status:** Accepted

### FR-007 - CLI: List Repo Tasks
**Statement:** The CLI shall provide a command to list the tasks contained in a repository.
**Priority:** Must
**Status:** Accepted

### FR-008 - CLI: Get Task Details
**Statement:** The CLI shall provide a command to retrieve details of a given task, optionally specifying which repo to use if it is not the current working directory.
**Acceptance Criteria:**
- Command accepts an optional `repo` argument.
- If `repo` is not specified, the current working directory is used to determine the repo.
**Priority:** Must
**Status:** Accepted

### FR-009 - CLI Working Context Resolution
**Statement:** The system shall determine the active repository from the current working directory by default, and shall raise an error if the current directory is not a repo recognized in the config file and no explicit repo argument is supplied.
**Why:** Prevents ambiguous or silent operation against the wrong repository.
**Acceptance Criteria:**
- Running a repo-scoped CLI command outside a configured repo directory, without an explicit repo argument, produces an error rather than a default/fallback behavior.
**Priority:** Must
**Status:** Accepted

### FR-010 - Agent Skill
**Statement:** The system shall provide an Agent Skill that exposes CLI details and capabilities to AI assistants.
**Why:** Enables AI assistants to discover and use Mesosphere's docs/tasks capabilities as part of AI-driven development.
**Priority:** Must
**Status:** Accepted

### FR-011 - UI: Document Review/Edit
**Statement:** The UI shall allow humans to review and edit markdown/text files within repos.
**Priority:** Must
**Status:** Accepted

### FR-012 - UI: Native Mermaid Rendering
**Statement:** The UI shall natively render mermaid diagrams embedded in documents.
**Priority:** Must
**Status:** Accepted

### FR-013 - UI: Navigation
**Statement:** The UI shall provide navigation between a "Docs" section and a "Tasks" section.
**Priority:** Must
**Status:** Accepted

### FR-014 - UI: Split View
**Statement:** The UI shall provide a "Split View" button that creates a second view pane, allowing the user to view two Repo+(Docs|Tasks) combinations simultaneously (including the same repo/section twice, e.g. Repo1.Docs + Repo1.Docs).
**Acceptance Criteria:**
- Layout is left/right when browser viewport width > height, top/bottom when viewport height > width.
- Each pane independently selects a repo and a Docs/Tasks mode.
- Any combination of repo and mode is supported per pane, including duplicate combinations across panes.
**Priority:** Should
**Status:** Accepted

### FR-015 - UI: Repo Selection Entry Point
**Statement:** Each view shall begin with a simple box-grid selector for choosing a repo; once selected, the repo name shall be displayed on a top bar within that view (left-aligned), with "Docs" and "Tasks" tabs right-aligned on the same top bar.
**Priority:** Must
**Status:** Accepted

### FR-016 - UI: Docs View Layout
**Statement:** The Docs View shall present a left-hand menu listing document names and directories, and shall render the currently selected document in the main pane.
**Acceptance Criteria:**
- Markdown renders natively.
- Images render natively.
- Mermaid diagrams render natively.
**Priority:** Must
**Status:** Accepted

### FR-017 - UI: Document Editing Workflow
**Statement:** The Docs View shall provide an edit affordance (pencil icon, bottom right) that switches the rendered document into a plain-text editable state and changes the icon to a Save icon.
**Priority:** Must
**Status:** Accepted

### FR-018 - UI: Document Save (Commit)
**Statement:** Saving an edited document shall perform a git commit behind the scenes.
**Why:** Git is the system of record; every saved change must be captured as a versioned commit.
**Priority:** Must
**Status:** Accepted

### FR-019 - UI: Document Publish (Push)
**Statement:** A "Publish" button shall appear once one or more documents have been saved. Activating it shall perform a `git pull` followed by a `git push` behind the scenes.
**Priority:** Must
**Status:** Accepted

### FR-020 - UI: Tasks View - Kanban Board
**Statement:** The Tasks View shall render a kanban-style board of tasks, reminiscent of Jira/Trello.
**Priority:** Must
**Status:** Accepted

### FR-021 - UI: Tasks View - Schema-Derived Columns
**Statement:** The board's columns (scrum statuses) shall be derived from the task file schema in use.
**Acceptance Criteria:**
- A task schema with binary/checkbox state produces exactly two columns: "TO DO" and "DONE".
- A task schema with explicit status values produces one column per distinct status.
- Tasks with a BLOCKED status remain in the "TO DO" column but display a blocked indicator rather than occupying a separate column.
**Priority:** Must
**Status:** Accepted

### FR-022 - UI: Blocked Task Indicator
**Statement:** A task shown with a blocked indicator shall reveal the name of its blocker on hover, and shall open a popup with the blocker task's information when the indicator is clicked.
**Priority:** Should
**Status:** Accepted

### FR-023 - UI: Task Card Editing
**Statement:** Clicking a task on the board shall open that task in an editable JSON format; a Save button shall appear when changes are made.
**Priority:** Must
**Status:** Accepted

### FR-024 - UI: Task Save (Commit)
**Statement:** Clicking the task Save button shall perform a git commit behind the scenes.
**Priority:** Must
**Status:** Accepted

### FR-025 - UI: Task Publish (Push)
**Statement:** A "Publish" button shall appear once one or more tasks have been saved. Activating it shall perform a `git pull` followed by a `git push` behind the scenes.
**Priority:** Must
**Status:** Accepted

### FR-026 - Core: Config Parsing
**Statement:** The system shall parse the config file to determine repository locations, supporting both local and remote repositories.
**Acceptance Criteria:**
- For a remote repo, the system uses configured git credentials by default, or explicitly specified credentials if present in config.
**Priority:** Must
**Status:** Accepted

### FR-027 - Core: Document Discovery
**Statement:** The system shall find all markdown and text files within configured repositories and shall retrieve their version history using git.
**Priority:** Must
**Status:** Accepted

### FR-029 - Core: Common Task Representation
**Statement:** The system shall use a common internal task representation consisting of: CardID, Scope, ValidationGate, and isBlocked.
**Why:** To provide a consistent UI across different task frameworks.
**Acceptance Criteria:**
- The system can map fields from supported frameworks (GitHub Spec Kit, OpenSpec, BMAD-METHOD, Amazon Kiro, Augment Cosmos) to this representation.
**Priority:** Must
**Status:** Accepted

### FR-030 - UI: Remote Change Notification
**Statement:** The UI shall perform a background `git fetch` every 3 minutes. If changes are detected, a refresh icon shall appear on the top-left of the view. Clicking this icon performs a `git pull` and updates the active view.
**Priority:** Should
**Status:** Accepted

### FR-034 - UI: Operation Feedback
**Statement:** The "Publish" button and "Refresh" icon shall provide visual feedback (e.g., spinning icon) and the "Publish" button shall be disabled while a git operation is in progress.
**Priority:** Must
**Status:** Accepted

### FR-032 - UI: Task Card - Scope Rendering
**Statement:** The Kanban card shall display the "Scope" field as text, truncated to fit the card dimensions.
**Priority:** Must
**Status:** Accepted

### FR-033 - UI: Task Detail - Raw View
**Statement:** The task detail view shall display the raw JSON from the task file.
**Priority:** Must
**Status:** Accepted

### FR-031 - UI: Conflict Handling
**Statement:** The UI shall display an error message if a "Publish" operation fails due to git conflicts.
**Priority:** Must
**Status:** Accepted

## 7. Non-Functional Requirements

- **Usability parity:** The Docs UI should be comparable in usability/familiarity to Confluence or Obsidian; the Tasks UI should be comparable to Jira or Trello. *(Status: Accepted, Source: notes)*
- **Responsiveness:** Split View layout must adapt to viewport orientation (width vs. height) dynamically. *(Status: Accepted, Source: notes)*

All other non-functional dimensions (performance targets, scalability limits, security/auth model, offline behavior, accessibility, supported platforms) are not specified in the source notes — see Open Questions.

## 8. Constraints & Invariants

- Git is the sole persistence/versioning mechanism for docs and tasks; there is no separate database of record described.
- A repo-scoped CLI operation must always resolve to exactly one repository, determined either by the current working directory (validated against config) or an explicit repo argument; ambiguous resolution is an error, not a silent default.
- Task board columns are derived from, not independent of, the task file schema — the system must not impose a fixed status taxonomy across all schemas.
- BLOCKED is always a display-time indicator within "TO DO," never a standalone board column.

## 9. External Dependencies / Integrations

- **Git:** required for version history, commit (save), and push (publish) operations, for both local and remote repositories.
- **Remote git hosting/credentials:** for remote repos, the system depends on configured git credentials (system-level or explicitly specified in config).
- **Mermaid:** required for native diagram rendering in the Docs UI.
- **AI assistant integration surface:** the Agent Skill is a dependency surface for external AI assistants/agents to consume; its consumers are out of Mesosphere's control.

## 10. Acceptance Criteria

High-level, system-wide acceptance (in addition to per-requirement criteria above):

- A user can configure one or more repos, launch the UI, select a repo, and browse both its docs and tasks.
- A user can edit a document, save it (git commit occurs), and publish it (git push occurs).
- A user can open a task, edit its JSON, save it (git commit occurs), and publish it (git push occurs).
- A user can open Split View and independently configure each pane with any repo/section combination, and the layout orientation follows viewport shape.
- An AI assistant, via the CLI or Agent Skill, can list repos, list a repo's docs, list a document's versions, retrieve document contents (current or a specific prior version), list a repo's tasks, and retrieve a task's details.
- Two different task file schemas (binary/checkbox and explicit-status) in two different repos produce correctly different kanban column layouts without code changes, via translation files.

## 11. Decisions & Rationale

- **Decision:** Project context (docs, tasks) lives directly in git repositories rather than in a separate hosted store.
  **Rationale:** Aligns with AI-driven development patterns where all context is baked into the repository, removing the overhead of separate SaaS PM tooling.
- **Decision:** Both a CLI and a UI are first-class interfaces, plus an Agent Skill layer.
  **Rationale:** Humans need reviewable/editable views (UI); AI assistants need programmatic access (CLI/Agent Skill) to the same underlying data.
- **Decision:** Saving is a two-step process (Save = local commit, Publish = push).
  **Rationale:** Lets edits accumulate as local commits before being shared, mirroring familiar git workflows.
- **Decision:** Task schemas are pluggable via translation files rather than a single fixed schema.
  **Rationale:** Different teams/projects already have or will invent their own task file frameworks; forcing one schema would reduce adoption.
- **Decision:** Use a common internal task representation.
  **Rationale:** Allows the UI to render disparate task frameworks (GitHub Spec Kit, OpenSpec, BMAD-METHOD, etc.) using a unified interface.
- **Decision:** No built-in authentication for the UI.
  **Rationale:** Mesosphere operates on locally cloned repositories; security is deferred to the local machine and git provider's existing auth.
- **Decision:** Single-user, local-file-based workflow.
  **Rationale:** Avoids the complexity of real-time collaboration by relying on git's decentralized nature.
- **Decision:** Web-based UI with cross-platform support.
  **Rationale:** Maximizes accessibility across Windows, Mac, and Linux while allowing both local and potentially hosted deployments.
- **Decision:** Perform a background `git fetch` every 3 minutes.
  **Rationale:** Balances the need for up-to-date information with network and resource efficiency.
- **Decision:** Disable the "Publish" button and show a spinning icon during operations.
  **Rationale:** Prevents race conditions and provides clear feedback to the user that a background git operation (pull/push/fetch) is in progress.
- **Decision:** Perform a `git pull` automatically when the "Publish" button is clicked.
  **Rationale:** Minimizes the likelihood of push rejection due to remote changes.
- **Decision:** Conflict resolution is out of scope for the initial version.
  **Rationale:** Simplifies initial implementation; users can resolve complex conflicts using standard git tools if the UI error occurs.
- **Decision:** Task Detail view shows raw JSON.
  **Rationale:** Provides full transparency into the underlying task data without requiring complex form generation for every possible schema.
- **Decision:** Top-left refresh icon for remote changes.
  **Rationale:** Provides a non-intrusive way for users to stay in sync with the remote repository.

## 12. Open Questions / Assumptions

- **Open Question:** How should the "ValidationGate" field be represented on the Kanban card (e.g., a progress bar, a count of criteria, or an icon)?
- **Open Question:** How should binary file conflicts (e.g., images) be handled if the Publish-time Pull encounters them?
- **Assumption:** "Remote repo" refers to a standard git remote (e.g., GitHub/GitLab-hosted) accessed via standard git protocols, not a bespoke remote storage system. Flagged as an assumption because the notes only say "assume configured git credentials" without naming a specific provider.
- **Assumption:** The config file format itself (YAML/JSON/TOML) is unspecified in the notes; left to DESIGN.md.

## 13. Traceability / Change Summary

- **2026-09-14 — Initial creation (greenfield).** Created SPEC.md from supplied product notes (`Mesosphere.md`). All functional requirements (FR-001–FR-028) derived directly from the notes' Config file, CLI, Agent Skill, UI, and Core sections. No prior SPEC existed. Open questions recorded for items the notes did not address (auth model, conflict handling, platform support, config file format, image storage, native task file format).
