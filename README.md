# Mesosphere

Git-native product management: docs and tasks live in your repositories. A single Go binary gives you a CLI, a local REST API, and an embedded web UI—no SaaS, no extra runtime.

Mesosphere is meant for humans and AI assistants that already work in git. Markdown/text documents and task files stay in the repo; **Save** commits locally and **Publish** pulls then pushes.

## Why

Hosted tools keep project knowledge away from the codebase. AI-driven workflows work better when specs, docs, and tasks sit next to the code they describe. Mesosphere does not replace those files with a new schema—it reads the ones you already have.

## Requirements

- **Go 1.23+** (module `github.com/mesosphere/mesosphere`)
- **Git** on `PATH` (all versioning and sync go through the system `git` CLI)
- **Node.js** only if you rebuild the frontend (`web/`)

The compiled binary has **no Node runtime dependency**. UI assets are embedded with `//go:embed`.

## Quick start

```bash
# From repo root
make binary          # or: bash scripts/build.sh
./mesosphere --help  # Windows: mesosphere.exe --help
```

With no arguments (or `mesosphere serve`), Mesosphere starts a local HTTP server (default `127.0.0.1:8080`) and opens the system browser.

```bash
mesosphere serve
mesosphere serve --addr 127.0.0.1:0
mesosphere serve --no-browser
```

## Configuration

Default file: `~/.mesosphere/config.yaml`  
Windows: `%USERPROFILE%\.mesosphere\config.yaml`  
Override directory with `MESOSPHERE_HOME` (then `config.yaml` lives in that folder).

If the file is missing, Mesosphere writes defaults on first load.

```yaml
version: "1.0"
settings:
  server:
    port: 8080
    auto_open_browser: true
  ui:
    active_tab: "docs"          # docs | tasks
    split_view_enabled: false
    panes: []
repositories:
  - id: "mesosphere-main"
    name: "Mesosphere Main Repo"
    path: "/path/to/local/repo"
    remote_url: "git@github.com:org/repo.git"
    credentials:
      ssh_key_path: "~/.ssh/id_rsa"
      # or token: "env:GITHUB_TOKEN"
```

CLI commands that need a repo use **`--repo <id-or-path>`**, or the current working directory if it is inside a configured `path`. Outside a configured tree and without `--repo`, the command errors (no silent fallback).

## CLI

| Command | Description |
| --- | --- |
| `mesosphere` / `mesosphere serve` | Start UI + API; optional `--addr`, `--no-browser` |
| `mesosphere repo list` | Configured repositories (table) |
| `mesosphere repo list --json` | Same as JSON |
| `mesosphere repo add --path <dir>` | Append a repo to config (`--id`, `--name`, `--remote` optional) |
| `mesosphere doc list [--repo]` | `.md` / `.txt` files |
| `mesosphere doc versions <path> [--repo]` | Git history for a document |
| `mesosphere doc get <path> [--version <hash>] [--repo]` | Working tree or historical content |
| `mesosphere task list [--repo]` | Tasks as JSON |
| `mesosphere task get <id> [--repo]` | One task as JSON |

AI assistants can use the skill at [`.agents/skills/mesosphere-cli/SKILL.md`](.agents/skills/mesosphere-cli/SKILL.md).

## Web UI

The UI is served from the same binary (`/` plus `/api/v1/...`).

- **Repo grid** → pick a configured repository
- **Docs / Tasks** tabs in the top bar
- **Docs:** file list, markdown (including ` ```mermaid ` as SVG), pencil to edit, Save (git commit), Publish (pull + push)
- **Tasks:** kanban with columns from task statuses (checkbox-style files → TO DO / DONE). Blocked tasks stay in TO DO with a red badge
- **Split View:** two panes; row if viewport is wider than tall, column otherwise
- **Refresh (↻):** shown top-left when the repo is **behind** upstream; click runs `git pull` and reloads

Publish shows a spinner while running and a conflict dialog on HTTP 409.

## REST API (`/api/v1`)

| Method | Path | Role |
| --- | --- | --- |
| GET/POST | `/config` | Read/update config + UI state |
| GET | `/repos` | List repositories |
| POST | `/repos` | Add a repository (201; 409 if id exists) |
| GET | `/repos/{id}/status` | Dirty tree, ahead/behind |
| POST | `/repos/{id}/sync` | `git pull` |
| POST | `/repos/{id}/publish` | pull then push (409 on conflict) |
| GET | `/repos/{id}/docs` | Document list |
| GET | `/repos/{id}/docs/detail?path=&version=` | Content |
| GET | `/repos/{id}/docs/versions?path=` | History |
| POST | `/repos/{id}/docs/save` | Write + commit |
| GET | `/repos/{id}/tasks` | Translated tasks |
| GET | `/repos/{id}/tasks/{taskID}` | One task |
| POST | `/repos/{id}/tasks/{taskID}/save` | Write raw JSON + commit |
| GET | `/health` | Liveness |

A background **git fetch** runs about every **3 minutes** for configured repos so status can show remote divergence.

## Tasks and frameworks

Task files (`.json`, `.yaml`, `.yml`, markdown checkboxes) are mapped into a common card (`card_id`, `title`, `scope`, `validation_gate`, `is_blocked`, `blocker_ids`, `status`, …).

Detectors include GitHub Spec Kit, OpenSpec, BMAD-METHOD, Amazon Kiro, Augment Cosmos, generic JSON/YAML, and `- [ ]` / `- [x]` markdown.

Save always writes the **raw** file and commits with a message like `update(mesosphere): task TASK-101`. Document saves use `update(mesosphere): <filename>` when no message is given.

## Development

```text
cmd/                 Cobra CLI
pkg/core/config      ~/.mesosphere/config.yaml
pkg/core/repo        Context resolution, sync/publish
pkg/core/docs        Discover / version / save markdown & text
pkg/core/tasks       Translators + save
pkg/git              os/exec wrapper around system git
pkg/server           Chi router, embed UI, REST, fetch ticker
pkg/browser          Open system browser
web/                 Static UI (copied into pkg/server/dist for embed)
```

```bash
make test            # go test ./...
make web             # npm run build in web/ → pkg/server/dist
make binary          # web + go build
cd web && npm run test:unit
```

Git must be installed for most package tests (they use temporary repositories).
