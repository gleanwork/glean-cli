# <img src="demo/glean-logo.png" width="28" height="28" style="vertical-align: middle; margin-right: 4px"> Glean CLI

Your company's knowledge, search, and AI — from the command line.

Search across your company's knowledge, chat with Glean Assistant, manage the full Glean API surface, and integrate Glean into automated workflows — all without leaving the terminal.

<p>
  <a href="https://github.com/gleanwork/.github/blob/main/docs/repository-stability.md#prerelease"><img src="https://img.shields.io/badge/-Prerelease-F6F3EB?style=flat-square&logo=data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgMzIgMzIiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+CjxwYXRoIGQ9Ik0yNC4zMDA2IDIuOTU0MjdMMjAuNzY1NiAwLjE5OTk1MUwxNy45MDI4IDMuOTk1MjdDMTMuNTY1MyAxLjkzNDk1IDguMjMwMTkgMy4wODQzOSA1LjE5Mzk0IDcuMDA5ODNDMS42NTg4OCAxMS41NjQyIDIuNDgzIDE4LjExMzggNy4wMzczOCAyMS42NDg5QzguNzcyMzggMjIuOTkzNSAxMC43ODkzIDIzLjcwOTIgMTIuODI3OSAyMy44MTc3QzE2LjE0NjEgMjQuMDEyOCAxOS41MDc3IDIyLjYyNDggMjEuNjc2NSAxOS44MDU1QzI0LjczNDQgMTUuODggMjQuNTE3NSAxMC40MTQ4IDIxLjQ1OTYgNi43Mjc4OUwyNC4zMDA2IDIuOTU0MjdaTTE4LjExOTcgMTcuMDUxMkMxNi4xMDI4IDE5LjYzMiAxMi4zNzI1IDIwLjEwOTEgOS43NzAwMSAxOC4wOTIyQzcuMTg5MTkgMTYuMDc1MiA2LjcxMjA3IDEyLjMyMzMgOC43MjkwMSA5Ljc0MjQ2QzkuNzA0OTQgOC40ODQ1OCAxMS4xMTQ2IDcuNjgyMTQgMTIuNjc2MSA3LjQ4Njk2QzEzLjA0NDggNy40NDM1OCAxMy40MTM1IDcuNDIxOSAxMy43ODIyIDcuNDQzNThDMTQuOTc1IDcuNTA4NjUgMTYuMTI0NCA3Ljk0MjM5IDE3LjA3ODcgOC42Nzk3N0MxOS42NTk1IDEwLjcxODQgMjAuMTM2NiAxNC40NzAzIDE4LjExOTcgMTcuMDUxMloiIGZpbGw9IndoaXRlIi8+CjxwYXRoIGQ9Ik0yNC41MTc2IDIxLjY5MjJDMjMuOTMyIDIyLjQ1MTMgMjMuMjgxNCAyMy4xMjM2IDIyLjU2NTcgMjMuNzUyNUMyMS44NzE3IDI0LjMzODEgMjEuMTEyNyAyNC44ODAzIDIwLjMxMDIgMjUuMzM1N0MxOS41Mjk1IDI1Ljc2OTUgMTguNjgzNyAyNi4xMzgyIDE3LjgzNzggMjYuNDIwMUMxNi45OTIgMjYuNzAyIDE2LjEwMjggMjYuODk3MiAxNS4yMTM3IDI3LjAwNTdDMTQuMzI0NSAyNy4xMTQxIDEzLjQzNTMgMjcuMTU3NSAxMi41MjQ0IDI3LjA5MjRDMTEuNjEzNSAyNy4wMjczIDEwLjcyNDMgMjYuODc1NSA5Ljg1Njg0IDI2LjY1ODdMOS42NjE2NSAyNy4zNzQzTDguNzcyNDYgMzAuOTk2MkM5LjkwMDIxIDMxLjI5OTggMTEuMDQ5NyAzMS40NzMzIDEyLjIyMDggMzEuNTZDMTIuMjY0MiAzMS41NiAxMi4zMjkyIDMxLjU2IDEyLjM3MjYgMzEuNTZDMTMuNTAwMyAzMS42MjUxIDE0LjY0OTggMzEuNTgxNyAxNS43NTU4IDMxLjQ1MTZDMTYuOTI3IDMxLjI5OTggMTguMDk4MSAzMS4wMzk1IDE5LjIyNTggMzAuNjcwOEMyMC4zNTM2IDMwLjMwMjIgMjEuNDU5NyAyOS44MjUgMjIuNTAwNyAyOS4yMzk1QzIzLjU2MzQgMjguNjUzOSAyNC41NjEgMjcuOTM4MiAyNS40OTM1IDI3LjE1NzVDMjYuNDQ3OCAyNi4zNTUgMjcuMzE1MyAyNS40NDQyIDI4LjA3NDQgMjQuNDQ2NUMyOC4xODI4IDI0LjMxNjQgMjguMjY5NSAyNC4xNjQ2IDI4LjM3OCAyNC4wMTI4TDI0Ljc3NzkgMjEuMzQ1MkMyNC42Njk0IDIxLjQ1MzcgMjQuNjA0NCAyMS41ODM4IDI0LjUxNzYgMjEuNjkyMloiIGZpbGw9IndoaXRlIi8+Cjwvc3ZnPg==&labelColor=343CED" alt="Prerelease"></a>
  <a href="https://github.com/gleanwork/glean-cli/releases"><img src="https://img.shields.io/github/v/release/gleanwork/glean-cli" alt="latest release"></a>
  <a href="https://github.com/gleanwork/glean-cli/blob/main/LICENSE"><img src="https://img.shields.io/github/license/gleanwork/glean-cli" alt="license"></a>
  <a href="https://github.com/gleanwork/glean-cli/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/gleanwork/glean-cli/ci.yml?branch=main&label=CI" alt="CI status"></a>
</p>

## Contents

- [ Glean CLI](#-glean-cli)
  - [Contents](#contents)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
  - [Why Glean CLI?](#why-glean-cli)
  - [Authentication](#authentication)
    - [API Token (CI/CD)](#api-token-cicd)
    - [Credential resolution order](#credential-resolution-order)
  - [Interactive TUI](#interactive-tui)
    - [Keyboard Shortcuts](#keyboard-shortcuts)
    - [Slash Commands](#slash-commands)
    - [File Attachments](#file-attachments)
  - [Commands](#commands)
    - [Core](#core)
    - [`glean search`](#glean-search)
    - [`glean chat`](#glean-chat)
    - [`glean api`](#glean-api)
    - [API Namespace Commands](#api-namespace-commands)
      - [Example payloads](#example-payloads)
  - [Agent Workflow](#agent-workflow)
  - [Environment Variables](#environment-variables)
  - [Exit Codes](#exit-codes)
  - [Shell Completions](#shell-completions)
  - [Agent Skills](#agent-skills)
    - [Install](#install)
    - [Migrating from per-command skills](#migrating-from-per-command-skills)
  - [Contributing](#contributing)
  - [Acknowledgments](#acknowledgments)
  - [License](#license)

## Installation

```bash snippet=readme/snippet-01.sh
# Homebrew (recommended)
brew install gleanwork/tap/glean-cli

# Manual
curl -fsSL https://raw.githubusercontent.com/gleanwork/glean-cli/main/install.sh | sh
```

Pre-built binaries for macOS, Linux, and Windows are available on the [Releases](https://github.com/gleanwork/glean-cli/releases) page.

## Quick Start

```bash snippet=readme/snippet-02.sh
# 1. Authenticate
glean auth login                          # OAuth via browser (recommended)
# — OR set env vars for CI/CD:
# export GLEAN_SERVER_URL=https://your-server-url GLEAN_API_TOKEN=your-token

# 2. Search
glean search "vacation policy"

# 3. Chat
glean chat "Summarize our Q1 engineering goals"

# 4. Open the interactive TUI
glean
```

## Why Glean CLI?

If your workflow involves searching Glean, asking Glean AI questions, or wiring Glean data into scripts or agent pipelines — the CLI is faster than a browser tab and composable with everything else in your terminal.

Every command returns structured JSON. Use `--dry-run` to preview requests before they're sent. Use `glean schema <command>` to get machine-readable flag documentation. Results pipe cleanly to `jq`, scripts, or any tool that reads stdin.

```bash snippet=readme/snippet-03.sh
# Discover what commands are available
glean schema | jq '.commands'

# Preview a request before sending
glean search --dry-run --datasource confluence "Q1 planning"

# Parse results with jq — each result has .title and nested .document fields
glean search "onboarding" | jq '.results[].title'

# Stream results as NDJSON — one SearchResult object per line
glean search "engineering docs" --output ndjson | jq .title
```

## Authentication

```bash snippet=readme/snippet-04.sh
glean auth login    # interactive login (detects the best method automatically)
glean auth status   # verify credentials, host, and token expiry
glean auth logout   # remove all stored credentials
```

`glean auth login` detects the right authentication method for your environment automatically:

| Method | When it's used | What happens |
| --- | --- | --- |
| **Browser login** | Default for most Glean instances | Opens your browser, you approve, done |
| **Device code login** | Organizations using an external IdP (e.g. Okta) | Prints a URL and code — open the URL, enter the code |
| **API token** | Instances without OAuth support | Prompts you to paste a token from Glean Admin |

You don't need to choose — `auth login` tries each method in order and uses the first one that works. Tokens are stored securely in the system keyring and refreshed automatically.

### API Token (CI/CD)

For non-interactive environments, set credentials via environment variables:

```bash snippet=readme/snippet-05.sh
export GLEAN_API_TOKEN=your-token
export GLEAN_SERVER_URL=https://your-server-url
glean search "test"
```

API tokens are scoped to an individual user account. Generate one from **Glean Admin → Settings → API Tokens**.

### Credential resolution order

Credentials are resolved in this order: environment variables → system keyring → `~/.glean/config.json`.

## Interactive TUI

Running `glean` with no arguments opens a full-screen chat powered by Glean Assistant.

```bash snippet=readme/snippet-06.sh
glean            # open TUI
glean --continue # resume the most recent session
```

### Keyboard Shortcuts

| Key              | Action                                   |
| ---------------- | ---------------------------------------- |
| `enter`          | Send message                             |
| `↑` / `↓`        | Scroll history / navigate command picker |
| `ctrl+r`         | New session                              |
| `ctrl+l`         | Clear screen                             |
| `ctrl+y`         | Copy last response                       |
| `ctrl+h`         | Toggle help                              |
| `ctrl+c` / `esc` | Quit                                     |

### Slash Commands

Type `/` in the input to open the command picker:

| Command                      | Description                  |
| ---------------------------- | ---------------------------- |
| `/mode auto\|fast\|advanced` | Switch agent reasoning depth |
| `/clear`                     | Start a new session          |
| `/help`                      | Show keyboard shortcuts      |

### File Attachments

Type `@` followed by a filename to attach a local file to your message. The file's contents are sent as context to Glean AI.

```
@go.mod          # attach go.mod from current directory
@src/config.go   # attach a specific file
```

Use ↑/↓ to navigate matches, Enter to attach, Esc to dismiss.

## Commands

### Core

| Command                  | Description                                                 |
| ------------------------ | ----------------------------------------------------------- |
| `glean search <query>`   | Search across your company's knowledge                      |
| `glean chat [message]`   | Chat with Glean Assistant (non-interactive)                 |
| `glean api <endpoint>`   | Make a raw authenticated HTTP request to the Glean REST API |
| `glean schema [command]` | Show machine-readable JSON schema for any command           |
| `glean auth`             | Authenticate with Glean                                     |
| `glean --version`        | Print the CLI version                                       |

### `glean search`

```bash snippet=readme/snippet-07.sh
glean search "vacation policy"
glean search "Q1 planning" --datasource confluence --page-size 5
glean search "docs" --fields "results.document.title,results.document.url"
glean search "docs" --output ndjson | jq .title
glean search --json '{"query":"onboarding","pageSize":3}'
glean search --dry-run "test"
```

| Flag                    | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `--output` / `--format` | `json` (default), `ndjson` (one result per line), `text` |
| `--fields`              | Dot-path field projection — prefix paths with `results.` |
| `--datasource` / `-d`   | Filter by datasource (repeatable)                        |
| `--type` / `-t`         | Filter by document type (repeatable)                     |
| `--page-size`           | Results per page (default 10)                            |
| `--json`                | Raw SDK request body (overrides all flags)               |
| `--dry-run`             | Print request body without sending                       |

### `glean chat`

```bash snippet=readme/snippet-08.sh
glean chat "What are our company holidays?"
glean chat --timeout 120000 "Summarize all Q1 OKRs across teams"
glean chat --json '{"messages":[{"author":"USER","messageType":"CONTENT","fragments":[{"text":"What is Glean?"}]}]}'
glean chat --dry-run "test"
echo "What is Glean?" | glean chat
glean chat                                # interactive multiline input, Ctrl+D to send
```

When called without a message argument, `glean chat` reads from stdin until EOF (Ctrl+D).
This enables multiline messages and piping input from other commands.

| Flag        | Description                                     |
| ----------- | ----------------------------------------------- |
| `--timeout` | Request timeout in milliseconds (default 60000) |
| `--json`    | Raw SDK request body (overrides all flags)      |
| `--dry-run` | Print request body without sending              |
| `--save`    | Persist chat for continuation (default true)    |

### `glean api`

Raw authenticated HTTP access to any Glean REST API endpoint (relative to `/rest/api/v1/`).

```bash snippet=readme/snippet-09.sh
glean api search --method POST --raw-field '{"query":"rust","pageSize":3}'
glean api --preview search --method POST --raw-field '{"query":"test"}'
```

### API Namespace Commands

All namespace commands accept `--json`, `--output`, `--fields`, and `--dry-run`. Run `glean <command> --help` for full usage.

| Namespace             | Subcommands                                                             | Description                                    |
| --------------------- | ----------------------------------------------------------------------- | ---------------------------------------------- |
| `glean agents`        | `list`, `get`, `schemas`, `run`                                         | Manage and invoke Glean AI agents              |
| `glean answers`       | `list`, `get`, `create`, `update`, `delete`                             | Curated Q&A pairs                              |
| `glean announcements` | `create`, `update`, `delete`                                            | Time-bounded company announcements             |
| `glean collections`   | `list`, `get`, `create`, `update`, `delete`, `add-items`, `delete-item` | Curated document collections                   |
| `glean documents`     | `get`, `summarize`, `get-by-facets`, `get-permissions`                  | Document retrieval and summarization           |
| `glean entities`      | `list`, `read-people`                                                   | People, teams, and custom entities             |
| `glean insights`      | `get`                                                                   | Search and usage analytics                     |
| `glean messages`      | `get`                                                                   | Retrieve indexed messages (Slack, Teams, etc.) |
| `glean pins`          | `list`, `get`, `create`, `update`, `remove`                             | Promoted search results                        |
| `glean shortcuts`     | `list`, `get`, `create`, `update`, `delete`                             | Go-links / memorable short URLs                |
| `glean tools`         | `list`, `run`                                                           | Glean platform tools                           |
| `glean verification`  | `list`, `verify`, `remind`                                              | Document verification and review               |
| `glean activity`      | `report`, `feedback`                                                    | User activity reporting                        |

#### Example payloads

```bash snippet=readme/snippet-10.sh
# Retrieve a document by URL
glean documents get --json '{"documentSpecs":[{"url":"https://..."}]}'

# Summarize a document
glean documents summarize --json '{"documentSpecs":[{"url":"https://..."}]}'

# Look up people
glean entities list --json '{"entityType":"PEOPLE","query":"engineering"}'

# Create a go-link
glean shortcuts create --json '{"data":{"inputAlias":"onboarding","destinationUrl":"https://..."}}'

# Create a shortcut with a variable template
glean shortcuts create --json '{"data":{"inputAlias":"jira","urlTemplate":"https://jira.example.com/browse/{arg}"}}'

# Pin a result for a query
glean pins create --json '{"queries":["onboarding"],"documentId":"https://..."}'

# List available AI agents (table view)
glean agents list --output text

# List agents — filter to specific fields without jq
glean agents list --fields agents.agent_id,agents.name

# Get a specific agent
glean agents get --json '{"agentId":"<id>"}'

# Get schemas for an agent
glean agents schemas --json '{"agentId":"<id>"}'

# Run an agent
glean agents run --json '{"agentId":"<id>","messages":[{"author":"USER","fragments":[{"text":"summarize Q1 results"}]}]}'
```

## Agent Workflow

The CLI is designed as a first-class tool for AI coding agents. Every command returns JSON on stdout and errors on stderr with non-zero exit codes.

```bash snippet=readme/snippet-11.sh
# 1. Discover all available commands
glean schema | jq '.commands'

# 2. Understand a command's flags
glean schema search | jq '.flags | keys'
glean schema search | jq '.flags["--output"]'

# 3. Preview the exact request before sending
glean shortcuts create --dry-run \
  --json '{"data":{"inputAlias":"test","destinationUrl":"https://example.com"}}'

# 4. Execute and parse results
glean search "engineering values" | jq '.results[].title'

# 5. Stream NDJSON for large result sets — one SearchResult object per line
glean search "all docs" --output ndjson --page-size 50 | jq .title
```

## Environment Variables

| Variable          | Description                                               |
| ----------------- | --------------------------------------------------------- |
| `GLEAN_API_TOKEN`  | API token — overrides keyring and config file             |
| `GLEAN_SERVER_URL` | Full Glean server URL. See [developers.glean.com](https://developers.glean.com/get-started/authentication) for how to find yours. |

Environment variables take precedence over stored configuration.

## Exit Codes

| Code | Meaning                                                          |
| ---- | ---------------------------------------------------------------- |
| `0`  | Success                                                          |
| `1`  | General error (authentication failure, API error, invalid input) |

All error details are written to stderr. Stdout contains only structured output (JSON/NDJSON/text), making the CLI safe for piping.

## Shell Completions

```bash snippet=readme/snippet-12.sh
glean completion bash   # Bash
glean completion zsh    # Zsh
glean completion fish   # Fish
```

## Agent Skills

The `skills/` directory contains a single [Agent Skill](https://agentskills.io) — `glean-cli` — that teaches AI coding agents how to use the Glean CLI effectively. Skills are supported by Claude Code, Cursor, GitHub Copilot, VS Code, Gemini CLI, OpenAI Codex, Goose, Amp, Roo Code, Junie, and [many others](https://agentskills.io).

`SKILL.md` is a thin router: it lists the commands and tells the agent to run `glean agent-help` for flags, payload shapes, and examples. That output comes from the installed binary, so it always matches the agent's version.

### Install

Use [`npx skills`](https://github.com/agentskills/agentskills) to install into your agent:

```bash snippet=readme/snippet-13.sh
npx skills add https://github.com/gleanwork/glean-cli
```

### Migrating from per-command skills

Earlier versions of this project shipped a separate skill per command (`glean-cli-search`, `glean-cli-pins`, etc.). If you previously installed any of them, remove the orphans in one go:

```bash
npx -y skills remove -g -y \
  glean-cli-activity glean-cli-agents glean-cli-announcements \
  glean-cli-answers glean-cli-api glean-cli-chat glean-cli-collections \
  glean-cli-documents glean-cli-entities glean-cli-insights \
  glean-cli-messages glean-cli-pins glean-cli-search glean-cli-shortcuts \
  glean-cli-tools glean-cli-verification
npx -y skills add https://github.com/gleanwork/glean-cli
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding conventions, and how to submit pull requests.

## Acknowledgments

The Glean CLI's agent-first design — structured JSON output, `--dry-run` previews, `glean schema` introspection, and agent skills — was heavily inspired by the [Google Workspace CLI](https://github.com/googleworkspace/cli) built by [Justin Poehnelt](https://github.com/jpoehnelt). His work on `gws` and his writing on designing CLIs for AI agents shaped how we think about making command-line tools that work as well for agents as they do for humans.

## License

MIT — see [LICENSE](LICENSE).
