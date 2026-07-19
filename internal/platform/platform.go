// Package platform decides whether a command call is served by the new
// platform APIs (/api/*) or the classic client APIs (/rest/api/v1/*).
//
// Migrated commands are platform-first: they call the platform endpoint and
// fall back to the classic equivalent when the tenant has not enabled
// experimental platform endpoints, which surfaces as a hidden 404. Setting
// GLEAN_LEGACY_APIS=1 skips platform endpoints entirely.
package platform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gleanwork/api-client-go/models/apierrors"
)

// EnvLegacy is the environment variable that forces migrated commands onto
// the classic client APIs.
const EnvLegacy = "GLEAN_LEGACY_APIS"

// Legacy reports whether the user has opted out of platform APIs.
func Legacy() bool {
	v := os.Getenv(EnvLegacy)
	return v == "1" || strings.EqualFold(v, "true")
}

// IsGateClosed reports whether err looks like the hidden 404 returned when a
// tenant has not enabled experimental platform endpoints. The platform stack
// answers gated endpoints with 404 resource_not_found, which is
// indistinguishable from a genuinely missing resource — callers on GET-by-ID
// endpoints therefore accept one spurious fallback attempt.
func IsGateClosed(err error) bool {
	var problem *apierrors.PlatformProblemDetailError
	if errors.As(err, &problem) {
		return problem.Status == http.StatusNotFound
	}
	var apiErr *apierrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// GateClosedErr wraps a gate-closed error with guidance for call sites that
// cannot fall back automatically (user-authored request bodies are coupled to
// the endpoint's request shape, so silently replaying them against the legacy
// endpoint would mangle the request).
func GateClosedErr(endpoint string, err error) error {
	return fmt.Errorf("platform API %s is unavailable on this instance; re-run with %s=1 and a legacy-shaped request body: %w", endpoint, EnvLegacy, err)
}

var (
	warnMu          sync.Mutex
	warnedEndpoints = map[string]struct{}{}
)

// warnFallback prints the fallback warning at most once per endpoint per
// process, so a process that falls back on several endpoints names each one
// exactly once.
func warnFallback(stderr io.Writer, endpoint string) {
	warnMu.Lock()
	defer warnMu.Unlock()
	if _, ok := warnedEndpoints[endpoint]; ok {
		return
	}
	warnedEndpoints[endpoint] = struct{}{}
	fmt.Fprintf(stderr, "Warning: platform API %s is unavailable on this instance; falling back to the legacy API. Set %s=1 to skip platform endpoints.\n", endpoint, EnvLegacy)
}

// ResetWarnings re-arms the warn-once guards. Only for use in tests.
func ResetWarnings() {
	warnMu.Lock()
	defer warnMu.Unlock()
	warnedEndpoints = map[string]struct{}{}
}

// Run executes platformFn, falling back to legacyFn when GLEAN_LEGACY_APIS is
// set (silently) or when the platform endpoint is gated off (with a one-time
// stderr warning naming endpoint). Non-gate errors from platformFn — auth
// failures, validation errors, server errors — are returned as-is: only the
// hidden 404 triggers a fallback. viaLegacy tells the caller which surface
// produced the result, since the two surfaces return different shapes.
func Run(ctx context.Context, stderr io.Writer, endpoint string, platformFn, legacyFn func(context.Context) (any, error)) (result any, viaLegacy bool, err error) {
	if Legacy() {
		result, err = legacyFn(ctx)
		return result, true, err
	}
	result, err = platformFn(ctx)
	if err == nil || !IsGateClosed(err) {
		return result, false, err
	}
	warnFallback(stderr, endpoint)
	result, err = legacyFn(ctx)
	return result, true, err
}
