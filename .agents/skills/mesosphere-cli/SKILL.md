# Mesosphere CLI (Agent Skill)

Use this skill to operate Mesosphere from an AI assistant. All project docs and tasks live in git repositories listed in `~/.mesosphere/config.yaml` (override home with `MESOSPHERE_HOME`).

## Invocation

```
mesosphere [command]
```

- **No arguments** or `mesosphere serve [--addr host:port]`: start the local UI/API server.
- **`--help`**: Cobra help for any command.

## Working context (`--repo`)

Repo-scoped commands (`doc`, `task`) resolve the repository as follows:

1. If `--repo <id-or-path>` is set, match a configured repository ID or filesystem path.
2. Otherwise, the current working directory must be inside a configured repository path.
3. If neither applies, the command **errors** (no silent fallback).

## Commands

### Repositories

```
mesosphere repo list
mesosphere repo list --json
```

**Output (default):** table columns `ID`, `NAME`, `PATH`.  
**Output (`--json`):** JSON array of repository objects (`id`, `name`, `path`, `remote_url`, `credentials`).

Example:

```
mesosphere repo list --json
```

### Documents

```
mesosphere doc list [--repo <id>]
mesosphere doc versions <doc-path> [--repo <id>]
mesosphere doc get <doc-path> [--version <commit-hash>] [--repo <id>]
```

| Command | Inputs | Output |
| --- | --- | --- |
| `doc list` | optional `--repo` | TSV: relative path, title (from first heading/line). `(no documents)` if empty. |
| `doc versions` | required `doc-path`, optional `--repo` | TSV: commit hash, date (`YYYY-MM-DD`), message. Newest first. |
| `doc get` | required `doc-path`, optional `--version`, optional `--repo` | Raw file text. Omit `--version` for the working tree; supply a hash for a historical revision. |

Examples:

```
mesosphere doc list --repo mesosphere-main
mesosphere doc versions docs/architecture.md --repo mesosphere-main
mesosphere doc get docs/architecture.md --repo mesosphere-main
mesosphere doc get docs/architecture.md --version abcdef123 --repo mesosphere-main
```

### Tasks

```
mesosphere task list [--repo <id>]
mesosphere task get <task-id> [--repo <id>]
```

Both commands print **JSON**.

- `task list`: array of `CommonTask` objects (`card_id`, `title`, `scope`, `validation_gate`, `is_blocked`, `blocker_ids`, `status`, `raw_content`, `file_path`, `repo_id`, …).
- `task get <task-id>`: single task; errors if the id is unknown.

Examples:

```
mesosphere task list --repo mesosphere-main
mesosphere task get TASK-101 --repo mesosphere-main
```

## Typical agent workflows

1. **Discover repos:** `mesosphere repo list --json`
2. **Browse docs:** `doc list` → `doc get <path>` → optional `doc versions` + `doc get --version`
3. **Inspect work:** `task list` → `task get <id>` for raw JSON and status
4. **Stay in a configured clone:** run without `--repo` from the repo working tree; otherwise always pass `--repo`.

Do not invent repository IDs. If context resolution fails, ask the user to configure the repo or pass `--repo`.
