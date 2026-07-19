package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/gleanwork/glean-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	platformSearchPath = "/api/search"
	legacySearchPath   = "/rest/api/v1/search"
)

// platformSearchResponse builds a minimal platform search response body with
// the given result titles.
func platformSearchResponse(titles ...string) []byte {
	type result struct {
		URL        string   `json:"url"`
		Title      string   `json:"title"`
		Datasource string   `json:"datasource"`
		Snippets   []string `json:"snippets,omitempty"`
	}
	rs := []result{}
	for _, title := range titles {
		rs = append(rs, result{URL: "https://docs.example.com", Title: title, Datasource: "gdrive"})
	}
	b, _ := json.Marshal(map[string]any{
		"results":     rs,
		"has_more":    false,
		"next_cursor": nil,
		"request_id":  "req-test",
	})
	return b
}

// legacySearchResponse builds a minimal classic SearchResponse JSON body with the given document titles.
func legacySearchResponse(titles ...string) []byte {
	type doc struct {
		Title string `json:"title"`
	}
	type result struct {
		Document doc `json:"document"`
	}
	var rs []result
	for _, title := range titles {
		rs = append(rs, result{Document: doc{Title: title}})
	}
	b, _ := json.Marshal(map[string]any{
		"results": rs,
	})
	return b
}

// gateClosedResponse is the hidden-404 problem detail the platform stack
// returns when experimental endpoints are not enabled for the tenant.
func gateClosedResponse() testutils.MockResponse {
	return testutils.MockResponse{
		StatusCode:  404,
		ContentType: "application/problem+json",
		Body:        []byte(`{"type":"about:blank","title":"Not Found","status":404,"detail":"resource not found","code":"resource_not_found","request_id":"req-gate"}`),
	}
}

// usePlatformAPIs pins the test to platform-first behavior regardless of the
// developer's own GLEAN_LEGACY_APIS setting.
func usePlatformAPIs(t *testing.T) {
	t.Helper()
	t.Setenv(platform.EnvLegacy, "")
}

func TestSearchCommand_BasicQuery_UsesPlatformAPI(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, platformSearchResponse("Vacation Policy", "Holiday Guide"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "vacation policy"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Vacation Policy")

	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, platformSearchPath, mock.Requests[0].URL.Path)
}

func TestSearchCommand_PlatformOutputNotCleansed(t *testing.T) {
	usePlatformAPIs(t)
	_, cleanup := testutils.SetupTestWithResponse(t, platformSearchResponse("Doc"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "doc"})
	require.NoError(t, root.Execute())

	// Platform responses are emitted as-is: snake_case fields survive.
	assert.Contains(t, buf.String(), "request_id")
	assert.Contains(t, buf.String(), "has_more")
}

func TestSearchCommand_GateClosedFallsBackToLegacy(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, nil)
	defer cleanup()
	mock.Routes = map[string]testutils.MockResponse{
		platformSearchPath: gateClosedResponse(),
		legacySearchPath:   {Body: legacySearchResponse("Legacy Doc")},
	}

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	root.SetArgs([]string{"search", "doc"})
	require.NoError(t, root.Execute())

	assert.Contains(t, buf.String(), "Legacy Doc")
	assert.Contains(t, errBuf.String(), "falling back to the legacy API")

	require.Len(t, mock.Requests, 2, "platform attempt then legacy fallback")
	assert.Equal(t, platformSearchPath, mock.Requests[0].URL.Path)
	assert.Equal(t, legacySearchPath, mock.Requests[1].URL.Path)
}

func TestSearchCommand_LegacyEnvSkipsPlatform(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")
	mock, cleanup := testutils.SetupTestWithResponse(t, legacySearchResponse("Legacy Doc"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(errBuf)
	root.SetArgs([]string{"search", "doc"})
	require.NoError(t, root.Execute())

	assert.Contains(t, buf.String(), "Legacy Doc")
	assert.NotContains(t, errBuf.String(), "Warning")
	require.Len(t, mock.Requests, 1)
	assert.Equal(t, legacySearchPath, mock.Requests[0].URL.Path)
}

func TestSearchCommand_MissingQuery(t *testing.T) {
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetErr(buf)
	root.SetArgs([]string{"search"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires a query argument")
}

func TestSearchCommand_DryRun(t *testing.T) {
	usePlatformAPIs(t)
	// Dry-run should not require credentials — SDK init is deferred until after the dry-run check.
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--dry-run", "test query"})
	err := root.Execute()
	require.NoError(t, err)
	var req map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &req), "dry-run output must be valid JSON")
	assert.Equal(t, "test query", req["query"])
	assert.Equal(t, float64(10), req["page_size"], "dry-run shows the platform request shape")
}

func TestSearchCommand_DryRun_LegacyEnv(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--dry-run", "test query"})
	require.NoError(t, root.Execute())
	var req map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &req))
	assert.Equal(t, "test query", req["query"])
	assert.Contains(t, buf.String(), "pageSize", "legacy dry-run shows the classic camelCase shape")
}

func TestSearchCommand_JSONPayload_Platform(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, platformSearchResponse("Engineering Docs"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--json", `{"query":"engineering","page_size":5}`})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Engineering Docs")
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, platformSearchPath, mock.Requests[0].URL.Path)
}

func TestSearchCommand_JSONPayload_GateClosedErrors(t *testing.T) {
	usePlatformAPIs(t)
	mock, cleanup := testutils.SetupTestWithResponse(t, nil)
	defer cleanup()
	mock.Routes = map[string]testutils.MockResponse{
		platformSearchPath: gateClosedResponse(),
	}

	root := NewCmdRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"search", "--json", `{"query":"engineering"}`})
	err := root.Execute()
	require.Error(t, err, "user-authored payloads must not silently fall back")
	assert.Contains(t, err.Error(), platform.EnvLegacy)
	require.Len(t, mock.Requests, 1, "no legacy retry for --json payloads")
}

func TestSearchCommand_JSONPayload_LegacyEnv(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")
	mock, cleanup := testutils.SetupTestWithResponse(t, legacySearchResponse("Engineering Docs"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--json", `{"query":"engineering","pageSize":5}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "Engineering Docs")
	require.NotEmpty(t, mock.Requests)
	assert.Equal(t, legacySearchPath, mock.Requests[0].URL.Path)
}

func TestSearchCommand_IgnoredFlagsNote(t *testing.T) {
	usePlatformAPIs(t)
	_, cleanup := testutils.SetupTestWithResponse(t, platformSearchResponse("Doc"))
	defer cleanup()

	root := NewCmdRoot()
	errBuf := &bytes.Buffer{}
	root.SetOut(&bytes.Buffer{})
	root.SetErr(errBuf)
	root.SetArgs([]string{"search", "--facet-bucket-size", "5", "--return-llm-content", "doc"})
	require.NoError(t, root.Execute())

	assert.Contains(t, errBuf.String(), "flags ignored by platform search")
	assert.Contains(t, errBuf.String(), "--facet-bucket-size")
	assert.Contains(t, errBuf.String(), "--return-llm-content")
}

func TestSearchCommand_OutputText(t *testing.T) {
	usePlatformAPIs(t)
	body, _ := json.Marshal(map[string]any{
		"results": []map[string]any{
			{
				"title":      "Vacation Policy",
				"datasource": "gdrive",
				"url":        "https://docs.example.com/vacation",
				"snippets":   []string{"All employees are entitled to 20 days PTO"},
			},
			{
				"title":      "Holiday Guide",
				"datasource": "confluence",
				"url":        "https://wiki.example.com/holidays",
			},
		},
		"has_more":   false,
		"request_id": "req-test",
	})
	_, cleanup := testutils.SetupTestWithResponse(t, body)
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--output", "text", "vacation"})
	err := root.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "TITLE")
	assert.Contains(t, out, "SOURCE")
	assert.Contains(t, out, "URL")
	assert.Contains(t, out, "SNIPPET")
	assert.Contains(t, out, "Vacation Policy")
	assert.Contains(t, out, "gdrive")
	assert.Contains(t, out, "Holiday Guide")
	assert.NotContains(t, out, "{")
}

func TestSearchCommand_OutputNDJSON(t *testing.T) {
	usePlatformAPIs(t)
	_, cleanup := testutils.SetupTestWithResponse(t, platformSearchResponse("Doc A", "Doc B"))
	defer cleanup()

	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"search", "--output", "ndjson", "test"})
	err := root.Execute()
	require.NoError(t, err)
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	assert.Greater(t, len(lines), 0)
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var obj map[string]any
		assert.NoError(t, json.Unmarshal(line, &obj))
	}
}
