// Package agenthelp derives agent-facing usage documentation from the live
// cobra command tree, enriched with the schema registry's semantics
// (when-to-use guidance, API surface, examples). Because the output is
// generated from the installed binary at runtime, it is version-accurate by
// construction — agents should treat it as the source of truth when it
// differs from static documentation.
package agenthelp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/config"
	"github.com/gleanwork/glean-cli/internal/output"
	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/gleanwork/glean-cli/internal/schema"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Context describes the environment the CLI is running in, so agents can
// adapt (e.g. run `glean auth login` before anything else, or expect legacy
// response shapes when the platform opt-out is active).
type Context struct {
	Version       string `json:"version"`
	ServerURL     string `json:"server_url,omitempty"`
	Authenticated bool   `json:"authenticated"`
	// AuthType is "api_token" or "oauth" when authenticated.
	AuthType string `json:"auth_type,omitempty"`
	// LegacyMode reports the GLEAN_LEGACY_APIS opt-out: platform-first
	// commands call the classic APIs directly and emit classic shapes.
	LegacyMode bool `json:"legacy_mode"`
	// ExperimentalDisabledByEnv reports that X_GLEAN_INCLUDE_EXPERIMENTAL is
	// explicitly false in the environment, which suppresses the experimental
	// header inside the SDK and forces platform-first commands into their
	// legacy fallback.
	ExperimentalDisabledByEnv bool `json:"experimental_disabled_by_env,omitempty"`
}

// BuildContext inspects config and environment. It never fails and never
// requires authentication: missing credentials simply yield
// Authenticated=false.
func BuildContext(version string) Context {
	c := Context{Version: version, LegacyMode: platform.Legacy()}
	if v, ok := os.LookupEnv("X_GLEAN_INCLUDE_EXPERIMENTAL"); ok && strings.EqualFold(v, "false") {
		c.ExperimentalDisabledByEnv = true
	}
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return c
	}
	c.ServerURL = cfg.GleanServerURL
	token, authType := client.ResolveToken(cfg)
	if token == "" {
		return c
	}
	c.Authenticated = true
	if authType == "" {
		c.AuthType = "api_token"
	} else {
		c.AuthType = "oauth"
	}
	return c
}

// FlagDoc describes one flag of one command. Existence, type, and default
// come from cobra (the binary is the source of truth); enum values and
// required-ness are enriched from the schema registry where available.
type FlagDoc struct {
	Type        string   `json:"type"`
	Default     string   `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

// CommandDoc describes one command (and, recursively, its subcommands).
type CommandDoc struct {
	Path        string             `json:"path"`
	Short       string             `json:"short,omitempty"`
	Long        string             `json:"long,omitempty"`
	WhenToUse   string             `json:"when_to_use,omitempty"`
	Surface     string             `json:"surface,omitempty"`
	Example     string             `json:"example,omitempty"`
	Flags       map[string]FlagDoc `json:"flags,omitempty"`
	Subcommands []CommandDoc       `json:"subcommands,omitempty"`
}

// Collect walks the visible commands under root and merges cobra's ground
// truth (structure, flags, defaults) with the schema registry's semantics.
func Collect(root *cobra.Command) []CommandDoc {
	var docs []CommandDoc
	for _, cmd := range root.Commands() {
		if skipCommand(cmd) {
			continue
		}
		docs = append(docs, collectCommand(cmd, cmd.Name()))
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })
	return docs
}

func skipCommand(cmd *cobra.Command) bool {
	return cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion"
}

func collectCommand(cmd *cobra.Command, path string) CommandDoc {
	doc := CommandDoc{
		Path:  path,
		Short: cmd.Short,
		Long:  cmd.Long,
		Flags: map[string]FlagDoc{},
	}

	// Registry semantics are keyed by top-level command name.
	topLevel := strings.SplitN(path, " ", 2)[0]
	var registered *schema.CommandSchema
	if s, err := schema.Get(topLevel); err == nil {
		registered = &s
		doc.WhenToUse = s.WhenToUse
		doc.Surface = s.Surface
		if path == topLevel {
			doc.Example = s.Example
		}
	}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fd := FlagDoc{
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		if registered != nil {
			if rs, ok := registered.Flags["--"+f.Name]; ok {
				fd.Enum = rs.Enum
				fd.Required = rs.Required
			}
		}
		doc.Flags["--"+f.Name] = fd
	})

	for _, sub := range cmd.Commands() {
		if skipCommand(sub) {
			continue
		}
		doc.Subcommands = append(doc.Subcommands, collectCommand(sub, path+" "+sub.Name()))
	}
	sort.Slice(doc.Subcommands, func(i, j int) bool { return doc.Subcommands[i].Path < doc.Subcommands[j].Path })
	return doc
}

// Find returns the CommandDoc for a command path like ["agents", "run"].
func Find(docs []CommandDoc, target []string) (CommandDoc, error) {
	if len(target) == 0 {
		return CommandDoc{}, fmt.Errorf("empty command path")
	}
	current := docs
	var found *CommandDoc
	for depth, name := range target {
		found = nil
		for i := range current {
			parts := strings.Fields(current[i].Path)
			if parts[len(parts)-1] == name {
				found = &current[i]
				break
			}
		}
		if found == nil {
			return CommandDoc{}, fmt.Errorf("unknown command %q — run 'glean agent-help' for the command map", strings.Join(target[:depth+1], " "))
		}
		current = found.Subcommands
	}
	return *found, nil
}

// jsonEnvelope is the --json output shape.
type jsonEnvelope struct {
	Context  Context      `json:"context"`
	Commands []CommandDoc `json:"commands,omitempty"`
	Command  *CommandDoc  `json:"command,omitempty"`
}

// Render writes the overview (empty target) or a single command's detail.
func Render(w io.Writer, ctx Context, docs []CommandDoc, target []string, asJSON bool) error {
	if len(target) == 0 {
		if asJSON {
			return writeJSON(w, jsonEnvelope{Context: ctx, Commands: docs})
		}
		renderOverview(w, ctx, docs)
		return nil
	}
	doc, err := Find(docs, target)
	if err != nil {
		return err
	}
	if asJSON {
		return writeJSON(w, jsonEnvelope{Context: ctx, Command: &doc})
	}
	renderCommand(w, doc)
	return nil
}

func writeJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func renderContext(w io.Writer, ctx Context) {
	auth := "no — run 'glean auth login' or set GLEAN_API_TOKEN"
	if ctx.Authenticated {
		auth = "yes (" + ctx.AuthType + ")"
	}
	mode := "platform-first (falls back to legacy per endpoint)"
	if ctx.LegacyMode {
		mode = "legacy (GLEAN_LEGACY_APIS is set)"
	} else if ctx.ExperimentalDisabledByEnv {
		mode = "legacy (X_GLEAN_INCLUDE_EXPERIMENTAL=false suppresses platform endpoints)"
	}
	fmt.Fprintf(w, "version: %s\n", ctx.Version)
	if ctx.ServerURL != "" {
		fmt.Fprintf(w, "server: %s\n", ctx.ServerURL)
	}
	fmt.Fprintf(w, "authenticated: %s\n", auth)
	fmt.Fprintf(w, "api mode: %s\n", mode)
}

func renderOverview(w io.Writer, ctx Context, docs []CommandDoc) {
	fmt.Fprintln(w, "glean — agent-facing usage generated from this binary.")
	fmt.Fprintln(w, "Treat this output as the source of truth when it differs from static docs.")
	fmt.Fprintln(w)
	renderContext(w, ctx)
	fmt.Fprintln(w)

	rows := make([][]string, 0, len(docs))
	for _, d := range docs {
		use := d.WhenToUse
		if use == "" {
			use = d.Short
		}
		rows = append(rows, []string{d.Path, use, d.Surface})
	}
	_ = output.WriteTable(w, []string{"COMMAND", "WHEN TO USE", "API"}, rows)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next: `glean agent-help <command> [subcommand]` for flags and payload shapes; add --json for machine-readable output.")
}

func renderCommand(w io.Writer, doc CommandDoc) {
	fmt.Fprintf(w, "glean %s — %s\n", doc.Path, doc.Short)
	if doc.Surface != "" {
		fmt.Fprintf(w, "api surface: %s\n", doc.Surface)
	}
	if doc.WhenToUse != "" {
		fmt.Fprintf(w, "when to use: %s\n", doc.WhenToUse)
	}
	if doc.Long != "" && doc.Long != doc.Short {
		fmt.Fprintf(w, "\n%s\n", strings.TrimSpace(doc.Long))
	}
	if len(doc.Flags) > 0 {
		fmt.Fprintln(w, "\nflags:")
		names := make([]string, 0, len(doc.Flags))
		for name := range doc.Flags {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			f := doc.Flags[name]
			line := fmt.Sprintf("  %s %s", name, f.Type)
			if f.Required {
				line += " (required)"
			}
			if len(f.Enum) > 0 {
				line += " [" + strings.Join(f.Enum, "|") + "]"
			}
			if f.Default != "" && f.Default != "false" && f.Default != "0" && f.Default != "[]" {
				line += " (default " + f.Default + ")"
			}
			if f.Description != "" {
				line += " — " + f.Description
			}
			fmt.Fprintln(w, line)
		}
	}
	if len(doc.Subcommands) > 0 {
		fmt.Fprintln(w, "\nsubcommands:")
		for _, sub := range doc.Subcommands {
			fmt.Fprintf(w, "  %s — %s\n", sub.Path, sub.Short)
		}
	}
	if doc.Example != "" {
		fmt.Fprintln(w, "\nexamples:")
		for _, line := range strings.Split(doc.Example, "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
	}
}
