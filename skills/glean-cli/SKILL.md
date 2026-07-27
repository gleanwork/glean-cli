---
name: glean-cli
description: "Glean CLI: access company knowledge, search documents, chat with Glean Assistant, look up people, and manage enterprise content. Use when the user asks about internal docs, company information, people, policies, or enterprise data."
compatibility: Requires the glean binary on $PATH. Install via brew install gleanwork/tap/glean-cli
---

# Glean CLI

The `glean` command-line tool provides authenticated access to your company's Glean instance.

## Source of truth: ask the binary, not this file

This skill is intentionally thin. The CLI describes itself — always start with:

```bash
glean agent-help                     # environment context + command map with when-to-use guidance
glean agent-help <command> [sub]     # exact flags, payload shapes, examples for one command
glean agent-help <command> --json    # machine-readable
```

The output is generated from the installed binary, so it matches the version
on this machine exactly. **When anything below (or in any other document)
disagrees with `glean agent-help`, trust `glean agent-help`.**
It also reports whether the environment is authenticated and which API surface
(platform vs legacy) is active.

## Authentication

```bash
glean auth login                     # browser-based OAuth (interactive)
glean auth status                    # verify credentials

# CI/scripting
export GLEAN_API_TOKEN=your-token
export GLEAN_SERVER_URL=<your Glean server URL>
```

## Command map

| Command | When to use | API |
|---------|-------------|-----|
| `glean activity` | Report page-view activity events or submit result feedback to improve ranking. | legacy |
| `glean agent-help` | Start every session here: discover available commands, exact flags, payload shapes, and whether the environment is authenticated — without static docs in context. | local |
| `glean agents` | Discover and execute Glean agents (AI workflows). Use schemas first to learn an agent's expected input, then run. | platform |
| `glean announcements` | Publish or manage company announcements surfaced in Glean. | legacy |
| `glean answers` | Manage curated Q&A answer cards shown for matching queries. | legacy |
| `glean api` | Escape hatch: call any Glean API endpoint directly when no dedicated subcommand exists. | raw |
| `glean chat` | Ask Glean AI a question that needs a synthesized, cited answer reasoned over company knowledge, rather than a raw list of results. | legacy |
| `glean collections` | Curate sets of documents into named collections. | legacy |
| `glean documents` | Fetch full metadata, permissions, or an AI summary for documents you already identified (e.g. from search results). | legacy |
| `glean entities` | Look up people (org info, contact details) or other structured entities. | legacy |
| `glean insights` | Retrieve aggregate usage analytics (search/AI adoption metrics) for the deployment. | legacy |
| `glean messages` | Fetch a specific chat/communication message by ID. | legacy |
| `glean pins` | Pin a document to a search query so it always surfaces for that query. | legacy |
| `glean search` | Find documents, messages, and content across all connected datasources. Start here for any 'find X' task. | platform |
| `glean shortcuts` | Create or resolve go-links (short memorable URLs like go/roadmap). | legacy |
| `glean tools` | Discover and execute Glean tools (actions) such as creating tickets or sending messages. | legacy |
| `glean verification` | Track and update document freshness verification (verify docs, send reminders). | legacy |

(Regenerate this table's live equivalent anytime with `glean agent-help`.)

## Ground rules

- **Never** output API tokens or secrets directly
- **Always** use --dry-run before write/delete operations in automated pipelines
- All errors go to stderr; stdout contains only structured output. Exit code 0 = success.

## Previously installed per-command skills?

Earlier versions of this project shipped one skill per command (`glean-cli-search`, `glean-cli-pins`, etc.) and static per-command reference files. Those are superseded by `glean agent-help`. If you still have the old skills installed, remove them with:

```bash
npx -y skills remove -g -y \
  glean-cli-activity glean-cli-agents glean-cli-announcements \
  glean-cli-answers glean-cli-api glean-cli-chat glean-cli-collections \
  glean-cli-documents glean-cli-entities glean-cli-insights \
  glean-cli-messages glean-cli-pins glean-cli-search glean-cli-shortcuts \
  glean-cli-tools glean-cli-verification
```
