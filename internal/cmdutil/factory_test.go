package cmdutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/apierrors"
	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/gleanwork/glean-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeReq is a platform-shaped request type with a snake_case JSON tag so
// buildAliases produces the agentId → agent_id normalization.
type fakeReq struct {
	AgentID string `json:"agent_id"`
}

func gateClosed() error {
	return &apierrors.PlatformProblemDetailError{Status: http.StatusNotFound}
}

// execSpec builds the command from spec, executes it with args, and returns
// stdout, stderr, and the execution error.
func execSpec(t *testing.T, spec Spec[fakeReq], args ...string) (string, string, error) {
	t.Helper()
	_, cleanup := testutils.SetupTestWithResponse(t, []byte(`{}`))
	defer cleanup()

	cmd := Build(spec)
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestBuild_PlatformFirst_Success(t *testing.T) {
	platform.ResetWarnings()
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use:      "get",
		Endpoint: "/api/agents/{agent_id}",
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			return map[string]string{"via": "platform", "id": req.AgentID}, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			t.Fatal("LegacyRun must not run on platform success")
			return nil, nil
		},
	}
	out, errOut, err := execSpec(t, spec, "--json", `{"agentId":"a-1"}`)
	require.NoError(t, err)
	assert.Contains(t, out, `"via": "platform"`)
	assert.Contains(t, out, `"id": "a-1"`, "camelCase input must normalize to the snake_case platform field")
	assert.Empty(t, errOut)
}

func TestBuild_FallbackAuto_GateClosedUsesLegacy(t *testing.T) {
	platform.ResetWarnings()
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use:      "get",
		Endpoint: "/api/agents/{agent_id}",
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			return nil, gateClosed()
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			var m map[string]string
			require.NoError(t, json.Unmarshal(rawJSON, &m))
			return map[string]string{"via": "legacy", "id": m["agent_id"]}, nil
		},
	}
	out, errOut, err := execSpec(t, spec, "--json", `{"agentId":"a-1"}`)
	require.NoError(t, err)
	assert.Contains(t, out, `"via": "legacy"`)
	assert.Contains(t, out, `"id": "a-1"`, "LegacyRun receives the normalized snake_case payload")
	assert.Contains(t, errOut, "falling back to the legacy API")
}

func TestBuild_FallbackEnvOnly_GateClosedErrors(t *testing.T) {
	platform.ResetWarnings()
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use:          "run",
		Endpoint:     "/api/agents/{agent_id}/runs",
		FallbackMode: FallbackEnvOnly,
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			return nil, gateClosed()
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			t.Fatal("LegacyRun must not auto-run in EnvOnly mode")
			return nil, nil
		},
	}
	_, _, err := execSpec(t, spec, "--json", `{"agent_id":"a-1"}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), platform.EnvLegacy)
	assert.Contains(t, err.Error(), "/api/agents/{agent_id}/runs")
}

func TestBuild_FallbackEnvOnly_NonGateErrorPassesThrough(t *testing.T) {
	platform.ResetWarnings()
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use:          "run",
		Endpoint:     "/api/agents/{agent_id}/runs",
		FallbackMode: FallbackEnvOnly,
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			return nil, &apierrors.PlatformProblemDetailError{Status: http.StatusUnauthorized, Title: "auth required"}
		},
	}
	_, _, err := execSpec(t, spec, "--json", `{"agent_id":"a-1"}`)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), platform.EnvLegacy, "non-404 errors are not gate-closed guidance")
}

func TestBuild_LegacyEnv_UsesLegacyRunDirectly(t *testing.T) {
	platform.ResetWarnings()
	t.Setenv(platform.EnvLegacy, "1")

	spec := Spec[fakeReq]{
		Use:      "get",
		Endpoint: "/api/agents/{agent_id}",
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			t.Fatal("platform Run must not execute in legacy mode")
			return nil, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			return map[string]string{"via": "legacy"}, nil
		},
	}
	out, errOut, err := execSpec(t, spec, "--json", `{"agentId":"a-1"}`)
	require.NoError(t, err)
	assert.Contains(t, out, `"via": "legacy"`)
	assert.Empty(t, errOut, "explicit legacy opt-out produces no warning")
}

func TestBuild_NoLegacyRun_BehavesClassically(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use: "get",
		Run: func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) {
			return map[string]string{"id": req.AgentID}, nil
		},
	}
	out, _, err := execSpec(t, spec, "--json", `{"agent_id":"a-1"}`)
	require.NoError(t, err)
	assert.Contains(t, out, `"id": "a-1"`)
}

func TestBuild_DryRun_PlatformShape(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "")

	spec := Spec[fakeReq]{
		Use:       "get",
		Endpoint:  "/api/agents/{agent_id}",
		Run:       func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) { return nil, nil },
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) { return nil, nil },
	}
	out, _, err := execSpec(t, spec, "--json", `{"agentId":"a-1"}`, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, `"agent_id": "a-1"`)
}

func TestBuild_DryRun_LegacyEnvShowsNormalizedPayload(t *testing.T) {
	t.Setenv(platform.EnvLegacy, "1")

	spec := Spec[fakeReq]{
		Use:       "get",
		Endpoint:  "/api/agents/{agent_id}",
		Run:       func(ctx context.Context, sdk *glean.Glean, req fakeReq) (any, error) { return nil, nil },
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) { return nil, nil },
	}
	out, _, err := execSpec(t, spec, "--json", `{"agentId":"a-1"}`, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, `"agent_id": "a-1"`)
}
