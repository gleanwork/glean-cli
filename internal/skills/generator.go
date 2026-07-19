// Package skills generates a single thin SKILL.md from the CLI's schema
// registry. The skill deliberately contains almost no command detail: its job
// is to route agents to `glean agent-help`, which generates version-accurate
// usage from the installed binary and is therefore the source of truth.
//
// Usage: glean generate-skills [--output-dir skills/]
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gleanwork/glean-cli/internal/schema"
)

// rootSkillName is the directory and frontmatter name for the root discovery skill.
const rootSkillName = "glean-cli"

// skillPrefix matches legacy per-command skill directory names
// (e.g. "glean-cli-search") for cleanup.
const skillPrefix = rootSkillName + "-"

var rootTmpl = template.Must(template.New("root").Parse(`---
name: {{ .RootSkill }}
description: "Glean CLI: access company knowledge, search documents, chat with Glean Assistant, look up people, and manage enterprise content. Use when the user asks about internal docs, company information, people, policies, or enterprise data."
compatibility: Requires the glean binary on $PATH. Install via brew install gleanwork/tap/glean-cli
---

# Glean CLI

The ` + "`glean`" + ` command-line tool provides authenticated access to your company's Glean instance.

## Source of truth: ask the binary, not this file

This skill is intentionally thin. The CLI describes itself — always start with:

` + "```bash" + `
glean agent-help                     # environment context + command map with when-to-use guidance
glean agent-help <command> [sub]     # exact flags, payload shapes, examples for one command
glean agent-help <command> --json    # machine-readable
` + "```" + `

The output is generated from the installed binary, so it matches the version
on this machine exactly. **When anything below (or in any other document)
disagrees with ` + "`glean agent-help`" + `, trust ` + "`glean agent-help`" + `.**
It also reports whether the environment is authenticated and which API surface
(platform vs legacy) is active.

## Authentication

` + "```bash" + `
glean auth login                     # browser-based OAuth (interactive)
glean auth status                    # verify credentials

# CI/scripting
export GLEAN_API_TOKEN=your-token
export GLEAN_SERVER_URL=<your Glean server URL>
` + "```" + `

## Command map

| Command | When to use | API |
|---------|-------------|-----|
{{ range .Commands -}}
| ` + "`glean {{ .Name }}`" + ` | {{ .WhenToUse }} | {{ .Surface }} |
{{ end }}
(Regenerate this table's live equivalent anytime with ` + "`glean agent-help`" + `.)

## Ground rules

- **Never** output API tokens or secrets directly
- **Always** use --dry-run before write/delete operations in automated pipelines
- All errors go to stderr; stdout contains only structured output. Exit code 0 = success.

## Previously installed per-command skills?

Earlier versions of this project shipped one skill per command (` + "`glean-cli-search`" + `, ` + "`glean-cli-pins`" + `, etc.) and static per-command reference files. Those are superseded by ` + "`glean agent-help`" + `. If you still have the old skills installed, remove them with:

` + "```bash" + `
npx -y skills remove -g -y \
  glean-cli-activity glean-cli-agents glean-cli-announcements \
  glean-cli-answers glean-cli-api glean-cli-chat glean-cli-collections \
  glean-cli-documents glean-cli-entities glean-cli-insights \
  glean-cli-messages glean-cli-pins glean-cli-search glean-cli-shortcuts \
  glean-cli-tools glean-cli-verification
` + "```" + `
`))

// CommandEntry is a row in the skill's command map, sourced from the schema
// registry (the same data agent-help serves live).
type CommandEntry struct {
	Name      string
	WhenToUse string
	Surface   string
}

// Generate writes the thin SKILL.md to outputDir.
func Generate(outputDir string) error {
	// Remove any legacy per-command skill directories from earlier versions
	// of this project so the generator output is always idempotent.
	if err := cleanStaleSkillDirs(outputDir); err != nil {
		return fmt.Errorf("cleaning stale skill dirs: %w", err)
	}

	var entries []CommandEntry
	for _, name := range schema.List() {
		s, err := schema.Get(name)
		if err != nil {
			continue
		}
		use := s.WhenToUse
		if use == "" {
			use = s.Description
		}
		entries = append(entries, CommandEntry{Name: name, WhenToUse: use, Surface: s.Surface})
	}

	if err := writeRootSkill(outputDir, entries); err != nil {
		return fmt.Errorf("writing root skill: %w", err)
	}
	fmt.Fprintf(os.Stderr, "  wrote %s/SKILL.md\n", rootSkillName)

	fmt.Fprintf(os.Stderr, "\nDone. Skills written to %s/\n", outputDir)
	return nil
}

// cleanStaleSkillDirs removes any per-command skill directories from earlier
// layouts of this project (skills/glean-cli-<cmd>/) and the retired
// per-command reference directory (skills/glean-cli/reference/), while
// leaving the root skill directory and anything unrelated intact.
func cleanStaleSkillDirs(outputDir string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// Match legacy "glean-cli-<cmd>" directories only. Root "glean-cli"
		// has no trailing hyphen after the rootSkillName, so it's safe.
		if !strings.HasPrefix(name, skillPrefix) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(outputDir, name)); err != nil {
			return fmt.Errorf("removing legacy skill dir %s: %w", name, err)
		}
	}
	// Static per-command reference files are superseded by `glean agent-help`.
	if err := os.RemoveAll(filepath.Join(outputDir, rootSkillName, "reference")); err != nil {
		return fmt.Errorf("removing retired reference dir: %w", err)
	}
	return nil
}

func writeRootSkill(outputDir string, commands []CommandEntry) error {
	dir := filepath.Join(outputDir, rootSkillName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		RootSkill string
		Commands  []CommandEntry
	}{RootSkill: rootSkillName, Commands: commands}
	return rootTmpl.Execute(f, data)
}
