package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gleanwork/glean-cli/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// driftAllowlist names visible top-level commands that deliberately have no
// schema registry entry: they are interactive or maintenance commands, not
// agent-invocable API operations. Everything else MUST be registered — the
// registry powers agent-help, so an unregistered command is invisible
// guidance-wise.
var driftAllowlist = map[string]bool{
	"auth":   true, // interactive login flow
	"update": true, // self-update maintenance
}

func TestAgentHelp_RegistryDriftBidirectional(t *testing.T) {
	root := NewCmdRoot()

	visible := map[string]bool{}
	for _, cmd := range root.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		visible[cmd.Name()] = true
		if driftAllowlist[cmd.Name()] {
			continue
		}
		_, err := schema.Get(cmd.Name())
		assert.NoError(t, err,
			"visible command %q has no schema registry entry — register it in cmd/schema.go (with WhenToUse and Surface) so agent-help can describe it",
			cmd.Name())
	}

	// Reverse direction: every registry entry must name a live command
	// (hidden commands count — `schema` stays registered as a hidden alias
	// would, but stale entries for deleted commands must fail).
	all := map[string]bool{}
	for _, cmd := range root.Commands() {
		all[cmd.Name()] = true
	}
	for _, name := range schema.List() {
		assert.True(t, all[name],
			"schema registry entry %q does not correspond to any command — remove the stale entry from cmd/schema.go", name)
	}
}

func TestAgentHelp_Overview(t *testing.T) {
	usePlatformAPIs(t)
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"agent-help"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "source of truth")
	assert.Contains(t, out, "COMMAND")
	assert.Contains(t, out, "search")
	assert.Contains(t, out, "agents")
	assert.Contains(t, out, "platform")
	assert.NotContains(t, out, "generate-skills", "hidden commands stay out of the map")
}

func TestAgentHelp_WorksUnauthenticated(t *testing.T) {
	// No credentials anywhere: agent-help must still succeed and say so.
	t.Setenv("GLEAN_SERVER_URL", "")
	t.Setenv("GLEAN_API_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"agent-help"})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "authenticated: no")
	assert.Contains(t, buf.String(), "glean auth login")
}

func TestAgentHelp_CommandDetail(t *testing.T) {
	usePlatformAPIs(t)
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"agent-help", "agents", "run"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "agents run")
	assert.Contains(t, out, "--json")
}

func TestAgentHelp_JSON(t *testing.T) {
	usePlatformAPIs(t)
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"agent-help", "search", "--json"})
	require.NoError(t, root.Execute())

	var envelope struct {
		Context map[string]any `json:"context"`
		Command struct {
			Path    string                    `json:"path"`
			Surface string                    `json:"surface"`
			Flags   map[string]map[string]any `json:"flags"`
		} `json:"command"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	assert.Equal(t, "search", envelope.Command.Path)
	assert.Equal(t, schema.SurfacePlatform, envelope.Command.Surface)
	assert.Contains(t, envelope.Command.Flags, "--page-size")
}

func TestAgentHelp_UnknownCommand(t *testing.T) {
	root := NewCmdRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"agent-help", "frobnicate"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestSchemaCommand_HiddenAliasStillWorks(t *testing.T) {
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"schema", "search"})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), `"command": "search"`)

	// And it is hidden from help.
	for _, cmd := range root.Commands() {
		if cmd.Name() == "schema" {
			assert.True(t, cmd.Hidden, "schema must be a hidden alias")
		}
	}
}
