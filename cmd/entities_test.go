package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/gleanwork/glean-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntitiesHelp(t *testing.T) {
	b := bytes.NewBufferString("")
	cmd := NewCmdEntities()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Usage")
}

// list

func TestEntitiesListDryRun(t *testing.T) {
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	b := bytes.NewBufferString("")
	cmd := NewCmdEntities()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list", "--dry-run", "--json", `{"entityType":"PEOPLE"}`})
	err := cmd.Execute()
	require.NoError(t, err)
	var req map[string]any
	require.NoError(t, json.Unmarshal(b.Bytes(), &req), "dry-run output must be valid JSON")
	assert.Equal(t, "PEOPLE", req["entityType"])
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{
  "entityType": "PEOPLE",
  "requestType": "STANDARD"
}
`))
}

func TestEntitiesListMissingJSON(t *testing.T) {
	cmd := NewCmdEntities()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	assert.Error(t, err, "missing --json must return error")
	assert.Contains(t, err.Error(), "--json is required")
}

func TestEntitiesListInvalidJSON(t *testing.T) {
	cmd := NewCmdEntities()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"list", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestEntitiesListBadEnumShowsValidValues(t *testing.T) {
	cmd := NewCmdEntities()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"list", "--json", `{"entityType":"BADVALUE"}`})
	err := cmd.Execute()
	assert.Error(t, err)
	errMsg := err.Error()
	assert.NotContains(t, errMsg, "ListEntitiesRequestEntityType")
	assert.Contains(t, errMsg, "PEOPLE")
}

func TestEntitiesListLive(t *testing.T) {
	_, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	b := bytes.NewBufferString("")
	cmd := NewCmdEntities()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"list", "--json", `{"entityType":"PEOPLE"}`})
	err := cmd.Execute()
	require.NoError(t, err)
}

// read-people

func TestEntitiesReadPeopleDryRun(t *testing.T) {
	// Dry-run must not require auth — SDK init is deferred until after the dry-run check.
	b := bytes.NewBufferString("")
	cmd := NewCmdEntities()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"read-people", "--dry-run", "--json", `{"emailIds":["user@example.com"]}`})
	err := cmd.Execute()
	require.NoError(t, err)
	var req map[string]any
	require.NoError(t, json.Unmarshal(b.Bytes(), &req), "dry-run output must be valid JSON")
	assert.Contains(t, req, "emailIds")
	snaps.MatchInlineSnapshot(t, b.String(), snaps.Inline(`{
  "emailIds": [
    "user@example.com"
  ],
  "fallbackToAuthenticatedIdentity": false
}
`))
}

func TestEntitiesReadPeopleMissingJSON(t *testing.T) {
	cmd := NewCmdEntities()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"read-people"})
	err := cmd.Execute()
	assert.Error(t, err, "missing --json must return error")
	assert.Contains(t, err.Error(), "--json is required")
}

func TestEntitiesReadPeopleInvalidJSON(t *testing.T) {
	cmd := NewCmdEntities()
	cmd.SetErr(bytes.NewBufferString(""))
	cmd.SetArgs([]string{"read-people", "--json", "not valid json"})
	err := cmd.Execute()
	assert.Error(t, err, "invalid JSON must return error")
}

func TestEntitiesReadPeopleLive(t *testing.T) {
	_, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()
	b := bytes.NewBufferString("")
	cmd := NewCmdEntities()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"read-people", "--json", `{"emailIds":["user@example.com"]}`})
	err := cmd.Execute()
	require.NoError(t, err)
}
