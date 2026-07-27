package platform

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gleanwork/api-client-go/models/apierrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// legacyResult is the sentinel legacyFn return value in Run tests.
const legacyResult = "legacy"

func TestLegacy(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"1", true},
		{"true", true},
		{"TRUE", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.value), func(t *testing.T) {
			t.Setenv(EnvLegacy, tt.value)
			assert.Equal(t, tt.want, Legacy())
		})
	}
}

func TestIsGateClosed(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
		{"problem detail 404", &apierrors.PlatformProblemDetailError{Status: http.StatusNotFound}, true},
		{"problem detail 404 wrapped", fmt.Errorf("search failed: %w", &apierrors.PlatformProblemDetailError{Status: http.StatusNotFound}), true},
		{"problem detail 401", &apierrors.PlatformProblemDetailError{Status: http.StatusUnauthorized}, false},
		{"problem detail 500", &apierrors.PlatformProblemDetailError{Status: http.StatusInternalServerError}, false},
		{"api error 404", &apierrors.APIError{StatusCode: http.StatusNotFound}, true},
		{"api error 403", &apierrors.APIError{StatusCode: http.StatusForbidden}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsGateClosed(tt.err))
		})
	}
}

func TestRun_PlatformSuccess(t *testing.T) {
	ResetWarnings()
	stderr := &bytes.Buffer{}
	result, viaLegacy, err := Run(context.Background(), stderr, "/api/search",
		func(context.Context) (any, error) { return "platform", nil },
		func(context.Context) (any, error) { t.Fatal("legacyFn must not run"); return nil, nil },
	)
	require.NoError(t, err)
	assert.Equal(t, "platform", result)
	assert.False(t, viaLegacy)
	assert.Empty(t, stderr.String())
}

func TestRun_GateClosedFallsBack(t *testing.T) {
	ResetWarnings()
	stderr := &bytes.Buffer{}
	gateErr := &apierrors.PlatformProblemDetailError{Status: http.StatusNotFound}
	result, viaLegacy, err := Run(context.Background(), stderr, "/api/search",
		func(context.Context) (any, error) { return nil, gateErr },
		func(context.Context) (any, error) { return legacyResult, nil },
	)
	require.NoError(t, err)
	assert.Equal(t, legacyResult, result)
	assert.True(t, viaLegacy)
	assert.Contains(t, stderr.String(), "platform API /api/search is unavailable")
	assert.Contains(t, stderr.String(), EnvLegacy)
}

func TestRun_WarnsOnlyOnce(t *testing.T) {
	ResetWarnings()
	stderr := &bytes.Buffer{}
	gateErr := &apierrors.APIError{StatusCode: http.StatusNotFound}
	for range 2 {
		_, _, err := Run(context.Background(), stderr, "/api/search",
			func(context.Context) (any, error) { return nil, gateErr },
			func(context.Context) (any, error) { return legacyResult, nil },
		)
		require.NoError(t, err)
	}
	assert.Equal(t, 1, bytes.Count(stderr.Bytes(), []byte("Warning:")))
}

func TestRun_WarnsPerEndpoint(t *testing.T) {
	ResetWarnings()
	stderr := &bytes.Buffer{}
	gateErr := &apierrors.APIError{StatusCode: http.StatusNotFound}
	for _, endpoint := range []string{"/api/search", "/api/agents/search"} {
		_, _, err := Run(context.Background(), stderr, endpoint,
			func(context.Context) (any, error) { return nil, gateErr },
			func(context.Context) (any, error) { return legacyResult, nil },
		)
		require.NoError(t, err)
	}
	assert.Equal(t, 2, bytes.Count(stderr.Bytes(), []byte("Warning:")), "each endpoint warns once")
}

func TestRun_NonGateErrorNoFallback(t *testing.T) {
	ResetWarnings()
	stderr := &bytes.Buffer{}
	authErr := &apierrors.PlatformProblemDetailError{Status: http.StatusUnauthorized}
	_, viaLegacy, err := Run(context.Background(), stderr, "/api/search",
		func(context.Context) (any, error) { return nil, authErr },
		func(context.Context) (any, error) {
			t.Fatal("legacyFn must not run on non-404 errors")
			return nil, nil
		},
	)
	require.Error(t, err)
	assert.False(t, viaLegacy)
	assert.Empty(t, stderr.String())
}

func TestRun_LegacyEnvBypassesPlatform(t *testing.T) {
	ResetWarnings()
	t.Setenv(EnvLegacy, "1")
	stderr := &bytes.Buffer{}
	result, viaLegacy, err := Run(context.Background(), stderr, "/api/search",
		func(context.Context) (any, error) { t.Fatal("platformFn must not run in legacy mode"); return nil, nil },
		func(context.Context) (any, error) { return legacyResult, nil },
	)
	require.NoError(t, err)
	assert.Equal(t, legacyResult, result)
	assert.True(t, viaLegacy)
	assert.Empty(t, stderr.String(), "legacy env opt-out is deliberate, no warning")
}

func TestGateClosedErr(t *testing.T) {
	base := &apierrors.PlatformProblemDetailError{Status: http.StatusNotFound}
	err := GateClosedErr("/api/search", base)
	assert.Contains(t, err.Error(), "/api/search")
	assert.Contains(t, err.Error(), EnvLegacy)
	assert.True(t, errors.Is(err, error(base)) || IsGateClosed(err), "must preserve the wrapped error")
}
