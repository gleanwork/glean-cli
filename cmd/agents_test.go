package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/gleanwork/glean-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentsHelp(t *testing.T) {
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Usage")
}

// list

func TestAgentsListDryRun(t *testing.T) {
	usePlatformAPIs(t)
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{}
`))
}

func TestAgentsListInvalidJSON(t *testing.T) {
	usePlatformAPIs(t)
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"list", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestAgentsListLive_UsesPlatformAPI(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{"agents":[],"request_id":"req-1"}`))
	defer cleanup()
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, "/api/agents/search", mock.Requests[0].URL.Path)
}

func TestAgentsListGateClosedFallsBack(t *testing.T) {
	usePlatformAPIs(t)
	platform.ResetWarnings()
	mock, cleanup := testutils.SetupTestWithResponse(t, nil)
	defer cleanup()
	mock.Routes = map[string]testutils.MockResponse{
		"/api/agents/search":         gateClosedResponse(),
		"/rest/api/v1/agents/search": {Body: []byte(`{"agents":[{"agent_id":"legacy-agent","name":"Legacy Agent","capabilities":{}}]}`)},
	}

	b := bytes.NewBufferString("")
	errBuf := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetErr(errBuf)
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, b.String(), "legacy-agent")
	assert.Contains(t, errBuf.String(), "falling back to the legacy API")
	require.Len(t, mock.Requests, 2)
	assert.Equal(t, "/api/agents/search", mock.Requests[0].URL.Path)
	assert.Equal(t, "/rest/api/v1/agents/search", mock.Requests[1].URL.Path)
}

func TestAgentsListLegacyEnv(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{"agents":[]}`))
	defer cleanup()
	cmd := NewCmdAgents()
	cmd.SetOut(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"list"})
	require.NoError(t, cmd.Execute())
	require.Len(t, mock.Requests, 1)
	assert.Equal(t, "/rest/api/v1/agents/search", mock.Requests[0].URL.Path)
}

func TestAgentsListFields(t *testing.T) {
	usePlatformAPIs(t)
	body, _ := json.Marshal(map[string]any{
		"agents": []map[string]any{
			{"agent_id": "agent-1", "name": "Research Agent", "capabilities": map[string]any{}},
			{"agent_id": "agent-2", "name": "Data Analyst", "capabilities": map[string]any{}},
		},
	})
	_, cleanup := testutils.SetupTestWithResponse(t, body)
	defer cleanup()

	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list", "--fields", "agents.agent_id,agents.name"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := b.String()
	assert.Contains(t, out, "agent-1")
	assert.Contains(t, out, "Research Agent")
	// capabilities should be filtered out
	assert.NotContains(t, out, "capabilities")
}

func TestAgentsListOutputText(t *testing.T) {
	usePlatformAPIs(t)
	body, _ := json.Marshal(map[string]any{
		"agents": []map[string]any{
			{"agent_id": "agent-1", "name": "Research Agent", "description": "Finds things", "capabilities": map[string]any{}},
		},
	})
	_, cleanup := testutils.SetupTestWithResponse(t, body)
	defer cleanup()

	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list", "--output", "text"})
	err := cmd.Execute()
	require.NoError(t, err)

	out := b.String()
	assert.Contains(t, out, "agent-1")
	assert.Contains(t, out, "Research Agent")
	assert.Contains(t, out, "Finds things")
}

// get

func TestAgentsGetDryRun(t *testing.T) {
	usePlatformAPIs(t)
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	// camelCase agentId input is normalized to the canonical snake_case shape.
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"get", "--dry-run", "--json", `{"agentId":"test-agent"}`})
	err := cmd.Execute()
	require.NoError(t, err)
	var req map[string]any
	require.NoError(t, json.Unmarshal(b.Bytes(), &req), "dry-run output must be valid JSON")
	assert.Equal(t, "test-agent", req["agent_id"])
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{
  "agent_id": "test-agent"
}
`))
}

func TestAgentsGetMissingJSON(t *testing.T) {
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"get"})
	err := cmd.Execute()
	assert.Error(t, err, "missing --json must return error")
	assert.Contains(t, err.Error(), "--json is required")
}

func TestAgentsGetInvalidJSON(t *testing.T) {
	usePlatformAPIs(t)
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"get", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestAgentsGetLive_UsesPlatformAPI(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{"agent":{"agent_id":"test-agent","name":"Test","capabilities":{}},"request_id":"req-1"}`))
	defer cleanup()
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"get", "--json", `{"agentId":"test-agent"}`})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, "/api/agents/test-agent", mock.Requests[0].URL.Path)
}

// schemas

func TestAgentsSchemasDryRun(t *testing.T) {
	usePlatformAPIs(t)
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"schemas", "--dry-run", "--json", `{"agentId":"test-agent"}`})
	err := cmd.Execute()
	require.NoError(t, err)
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{
  "agent_id": "test-agent"
}
`))
}

func TestAgentsSchemasMissingJSON(t *testing.T) {
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"schemas"})
	err := cmd.Execute()
	assert.Error(t, err, "missing --json must return error")
	assert.Contains(t, err.Error(), "--json is required")
}

func TestAgentsSchemasInvalidJSON(t *testing.T) {
	usePlatformAPIs(t)
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"schemas", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestAgentsSchemasLive_UsesPlatformAPI(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"schemas", "--json", `{"agentId":"test-agent"}`})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, "/api/agents/test-agent/schemas", mock.Requests[0].URL.Path)
}

// run

func TestAgentsRunDryRun(t *testing.T) {
	usePlatformAPIs(t)
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"run", "--dry-run", "--json", `{"agent_id":"test-agent","messages":[]}`})
	err := cmd.Execute()
	require.NoError(t, err)
	var req map[string]any
	require.NoError(t, json.Unmarshal(b.Bytes(), &req), "dry-run output must be valid JSON")
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{
  "agent_id": "test-agent"
}
`))
}

func TestAgentsRunMissingJSON(t *testing.T) {
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"run"})
	err := cmd.Execute()
	assert.Error(t, err, "missing --json must return error")
	assert.Contains(t, err.Error(), "--json is required")
}

func TestAgentsRunInvalidJSON(t *testing.T) {
	usePlatformAPIs(t)
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"run", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestAgentsRunMissingAgentID(t *testing.T) {
	usePlatformAPIs(t)
	_, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	cmd := NewCmdAgents()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"run", "--json", `{"input":{"query":"hi"}}`})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id is required")
}

func TestAgentsRunLive_UsesPlatformAPI(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	// The platform run op negotiates streaming in its Accept header; pin the
	// mock's Content-Type so the SDK parses the buffered JSON response.
	mock.ContentType = "application/json"
	b := bytes.NewBufferString("")
	cmd := NewCmdAgents()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"run", "--json", `{"agent_id":"test-agent","input":{"query":"hi"}}`})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, "/api/agents/test-agent/runs", mock.Requests[0].URL.Path)
}

func TestAgentsRunGateClosedErrors(t *testing.T) {
	usePlatformAPIs(t)
	platform.ResetWarnings()
	mock, cleanup := testutils.SetupTestWithResponse(t, nil)
	defer cleanup()
	mock.Routes = map[string]testutils.MockResponse{
		"/api/agents/test-agent/runs": gateClosedResponse(),
	}
	cmd := NewCmdAgents()
	cmd.SetOut(bytes.NewBufferString(""))
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"run", "--json", `{"agent_id":"test-agent","input":{"query":"hi"}}`})
	err := cmd.Execute()
	require.Error(t, err, "run must not silently fall back to the legacy body shape")
	assert.Contains(t, err.Error(), platform.EnvLegacy)
	require.Len(t, mock.Requests, 1, "no legacy retry for agents run")
}

func TestAgentsRunLegacyEnv(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")
	mock, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	cmd := NewCmdAgents()
	cmd.SetOut(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"run", "--json", `{"agent_id":"test-agent","messages":[{"author":"USER","fragments":[{"text":"hi"}]}]}`})
	err := cmd.Execute()
	require.NoError(t, err)
	require.Len(t, mock.Requests, 1)
	assert.Equal(t, "/rest/api/v1/agents/runs/wait", mock.Requests[0].URL.Path)
}
