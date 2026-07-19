package agenthelp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gleanwork/glean-cli/internal/schema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestTree builds a synthetic cobra tree with a registered schema entry,
// a hidden command, and a nested subcommand.
func newTestTree(t *testing.T) *cobra.Command {
	t.Helper()
	schema.Register(schema.CommandSchema{
		Command:     "widgets",
		Description: "Manage widgets.",
		WhenToUse:   "Use for all widget lifecycle operations.",
		Surface:     schema.SurfacePlatform,
		Flags: map[string]schema.FlagSchema{
			"--output": {Type: "enum", Enum: []string{"json", "text"}},
			"--json":   {Type: "string", Required: true},
		},
		Example: "glean widgets list",
	})

	root := &cobra.Command{Use: "glean"}
	widgets := &cobra.Command{Use: "widgets", Short: "Manage widgets"}
	list := &cobra.Command{Use: "list", Short: "List widgets", Run: func(*cobra.Command, []string) {}}
	list.Flags().String("output", "json", "Output format")
	list.Flags().String("json", "", "JSON request body")
	widgets.AddCommand(list)
	hidden := &cobra.Command{Use: "secret", Hidden: true}
	root.AddCommand(widgets, hidden)
	return root
}

func TestCollect_SkipsHiddenAndMergesRegistry(t *testing.T) {
	docs := Collect(newTestTree(t))

	require.Len(t, docs, 1, "hidden commands are excluded")
	w := docs[0]
	assert.Equal(t, "widgets", w.Path)
	assert.Equal(t, "Use for all widget lifecycle operations.", w.WhenToUse)
	assert.Equal(t, schema.SurfacePlatform, w.Surface)
	assert.Equal(t, "glean widgets list", w.Example)

	require.Len(t, w.Subcommands, 1)
	list := w.Subcommands[0]
	assert.Equal(t, "widgets list", list.Path)
	assert.Equal(t, "Use for all widget lifecycle operations.", list.WhenToUse, "subcommands inherit top-level registry semantics")

	// cobra is truth for flag existence/type; registry enriches enum/required.
	outFlag, ok := list.Flags["--output"]
	require.True(t, ok)
	assert.Equal(t, "string", outFlag.Type)
	assert.Equal(t, []string{"json", "text"}, outFlag.Enum)
	jsonFlag := list.Flags["--json"]
	assert.True(t, jsonFlag.Required)
}

func TestFind(t *testing.T) {
	docs := Collect(newTestTree(t))

	doc, err := Find(docs, []string{"widgets", "list"})
	require.NoError(t, err)
	assert.Equal(t, "widgets list", doc.Path)

	_, err = Find(docs, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestRender_OverviewText(t *testing.T) {
	docs := Collect(newTestTree(t))
	buf := &bytes.Buffer{}
	ctx := Context{Version: "test", ServerURL: "https://acme-be.glean.com", Authenticated: true, AuthType: "api_token"}
	require.NoError(t, Render(buf, ctx, docs, nil, false))

	out := buf.String()
	assert.Contains(t, out, "source of truth")
	assert.Contains(t, out, "version: test")
	assert.Contains(t, out, "authenticated: yes (api_token)")
	assert.Contains(t, out, "widgets")
	assert.Contains(t, out, "Use for all widget lifecycle operations.")
}

func TestRender_UnauthenticatedGuidance(t *testing.T) {
	buf := &bytes.Buffer{}
	require.NoError(t, Render(buf, Context{Version: "test"}, nil, nil, false))
	assert.Contains(t, buf.String(), "glean auth login")
}

func TestRender_JSON(t *testing.T) {
	docs := Collect(newTestTree(t))
	buf := &bytes.Buffer{}
	ctx := Context{Version: "test", LegacyMode: true}
	require.NoError(t, Render(buf, ctx, docs, []string{"widgets"}, true))

	var envelope struct {
		Context struct {
			Version    string `json:"version"`
			LegacyMode bool   `json:"legacy_mode"`
		} `json:"context"`
		Command struct {
			Path    string `json:"path"`
			Surface string `json:"surface"`
		} `json:"command"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	assert.Equal(t, "test", envelope.Context.Version)
	assert.True(t, envelope.Context.LegacyMode)
	assert.Equal(t, "widgets", envelope.Command.Path)
	assert.Equal(t, schema.SurfacePlatform, envelope.Command.Surface)
}

func TestBuildContext_NeverLeaksTokenMaterial(t *testing.T) {
	t.Setenv("GLEAN_SERVER_URL", "https://acme-be.glean.com")
	t.Setenv("GLEAN_API_TOKEN", "super-secret-token-value")
	t.Setenv("X_GLEAN_INCLUDE_EXPERIMENTAL", "false")

	ctx := BuildContext("test")
	assert.True(t, ctx.Authenticated)
	assert.Equal(t, "api_token", ctx.AuthType)
	assert.True(t, ctx.ExperimentalDisabledByEnv)

	serialized, err := json.Marshal(ctx)
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(serialized), "super-secret"), "context must never contain token material")
}
