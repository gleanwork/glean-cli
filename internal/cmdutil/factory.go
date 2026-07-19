// Package cmdutil provides the generic Build function for constructing
// SDK-backed CLI subcommands with consistent flags and protocol enforcement.
//
// Every namespace subcommand follows the same contract:
//   - Optional --json flag accepts a raw JSON payload
//   - Optional --output flag (json | ndjson | text)
//   - Optional --fields flag filters output to specific dot-path fields (e.g. agents.agent_id,agents.name)
//   - Optional --dry-run flag prints the request without sending it (never requires auth)
//   - camelCase input keys are transparently normalized to snake_case where the
//     SDK uses snake_case JSON tags (e.g. agentId → agent_id)
package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	glean "github.com/gleanwork/api-client-go"
	gleanClient "github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/output"
	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/spf13/cobra"
)

// FallbackMode controls what a platform-first command (Spec.LegacyRun != nil)
// does when the platform endpoint is gated off (hidden 404).
type FallbackMode int

const (
	// FallbackAuto retries via LegacyRun with a one-time warning. Use for
	// commands whose request the CLI built itself (or whose input is
	// surface-agnostic, like a bare ID).
	FallbackAuto FallbackMode = iota
	// FallbackEnvOnly never retries automatically: gate-closed produces an
	// actionable error naming GLEAN_LEGACY_APIS. Use for commands whose
	// --json body is user-authored and coupled to the endpoint's shape.
	FallbackEnvOnly
)

// Spec describes a single SDK-backed subcommand.
type Spec[Req any] struct {
	Use   string
	Short string
	Long  string

	// JSONRequired causes an error when --json is absent.
	// Set false for list/read commands where the request body is optional.
	JSONRequired bool

	// Endpoint is the platform API path used in fallback warnings and
	// gate-closed errors (e.g. "/api/agents/search"). Required when
	// LegacyRun is set.
	Endpoint string

	// FallbackMode selects the gate-closed behavior when LegacyRun is set.
	// The zero value is FallbackAuto.
	FallbackMode FallbackMode

	// LegacyRun, when non-nil, marks this command platform-first: Run is the
	// platform API call, and LegacyRun serves the classic API instead when
	// GLEAN_LEGACY_APIS=1 or as the gate-closed fallback (FallbackAuto).
	// It receives the normalized --json bytes (snake_case keys — the CLI's
	// canonical input shape) so it can unmarshal the classic request type,
	// which differs from the platform Req type.
	LegacyRun func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error)

	// ErrTransform optionally remaps errors returned by json.Unmarshal before
	// they are surfaced to the user. Use it to replace Go type names in SDK
	// enum validation errors with human-readable messages. The transform must
	// return an error (never nil); return fmt.Errorf("invalid --json: %w", err)
	// as the fallback to preserve the default behavior.
	ErrTransform func(err error) error

	// TextFn is the human-readable formatter called when --output text is set.
	// If nil, text output falls back to JSON. The value passed is the result
	// returned by Run (same type assertion rules apply as in Run).
	TextFn func(w io.Writer, v any) error

	// Run calls the SDK method and returns the response payload to serialize.
	// The payload must be the inner component field — NOT the operation wrapper
	// (e.g. return resp.ListCollectionsResponse, not resp).
	// Return (nil, nil) for operations with no response body (delete, etc.).
	// When LegacyRun is set, Run is the platform API call.
	Run func(ctx context.Context, sdk *glean.Glean, req Req) (any, error)
}

// Build constructs a *cobra.Command from spec, enforcing:
//  1. --json required check (before anything else, when JSONRequired is true)
//  2. Transparent camelCase → snake_case key normalization on --json input
//  3. --dry-run check before SDK init (never requires auth)
//  4. SDK init only when actually sending
//  5. Consistent --json / --output / --fields / --dry-run flag registration
func Build[Req any](spec Spec[Req]) *cobra.Command {
	var jsonPayload, outputFormat, fields string
	var dryRun bool

	aliases := buildAliases(reflect.TypeOf(new(Req)).Elem())

	cmd := &cobra.Command{
		Use:   spec.Use,
		Short: spec.Short,
		Long:  spec.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if spec.JSONRequired && jsonPayload == "" {
				return fmt.Errorf("--json is required\n\nRun '%s --help' for the expected payload format", cmd.CommandPath())
			}

			var normalized []byte
			if jsonPayload != "" {
				normalized = normalizeKeys([]byte(jsonPayload), aliases)
			}

			// In legacy mode the payload is parsed by LegacyRun against the
			// classic request type, so skip the platform-type parse (the two
			// shapes differ and a legacy payload may not satisfy it).
			legacyMode := spec.LegacyRun != nil && platform.Legacy()

			var req Req
			if len(normalized) > 0 && !legacyMode {
				if err := json.Unmarshal(normalized, &req); err != nil {
					if spec.ErrTransform != nil {
						return spec.ErrTransform(err)
					}
					return fmt.Errorf("invalid --json: %w", err)
				}
			}

			if dryRun {
				if legacyMode {
					// The classic request type is known only to LegacyRun;
					// show the normalized payload as-is.
					payload := map[string]any{}
					if len(normalized) > 0 {
						if err := json.Unmarshal(normalized, &payload); err != nil {
							return fmt.Errorf("invalid --json: %w", err)
						}
					}
					return output.WriteJSON(cmd.OutOrStdout(), payload)
				}
				return output.WriteJSON(cmd.OutOrStdout(), req)
			}

			sdk, err := gleanClient.NewFromConfig()
			if err != nil {
				return err
			}

			var result any
			switch {
			case spec.LegacyRun == nil:
				result, err = spec.Run(cmd.Context(), sdk, req)
			case legacyMode:
				result, err = spec.LegacyRun(cmd.Context(), sdk, normalized)
			case spec.FallbackMode == FallbackEnvOnly:
				result, err = spec.Run(cmd.Context(), sdk, req)
				if err != nil && platform.IsGateClosed(err) {
					return platform.GateClosedErr(spec.Endpoint, err)
				}
			default: // FallbackAuto
				result, _, err = platform.Run(cmd.Context(), cmd.ErrOrStderr(), spec.Endpoint,
					func(ctx context.Context) (any, error) { return spec.Run(ctx, sdk, req) },
					func(ctx context.Context) (any, error) { return spec.LegacyRun(ctx, sdk, normalized) },
				)
			}
			if err != nil {
				return err
			}
			if result == nil {
				return nil
			}
			if fields != "" {
				return output.ProjectFields(cmd.OutOrStdout(), result, fields)
			}
			return output.WriteFormatted(cmd.OutOrStdout(), result, outputFormat, spec.TextFn)
		},
	}

	jsonUsage := "JSON request body"
	if spec.JSONRequired {
		jsonUsage += " (required)"
	}
	cmd.Flags().StringVar(&jsonPayload, "json", "", jsonUsage)
	cmd.Flags().StringVar(&outputFormat, "output", "json", "Output format: json, ndjson, or text")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated dot-path fields to include (e.g. agents.agent_id,agents.name)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print request without sending")
	return cmd
}

// buildAliases reflects on t to find fields with snake_case JSON tags and
// returns a map of camelCase → snake_case for transparent input normalization.
// Only top-level struct fields are inspected.
func buildAliases(t reflect.Type) map[string]string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	aliases := map[string]string{}
	for i := 0; i < t.NumField(); i++ {
		tag := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if strings.Contains(tag, "_") {
			camel := snakeToCamel(tag)
			if camel != tag {
				aliases[camel] = tag
			}
		}
	}
	return aliases
}

// normalizeKeys rewrites camelCase keys in a flat JSON object to their
// snake_case equivalents according to aliases. Unknown keys are left as-is.
// If the JSON is not a flat object, data is returned unchanged.
func normalizeKeys(data []byte, aliases map[string]string) []byte {
	if len(aliases) == 0 {
		return data
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return data
	}
	changed := false
	for camel, snake := range aliases {
		if v, ok := m[camel]; ok {
			if _, exists := m[snake]; !exists {
				m[snake] = v
				delete(m, camel)
				changed = true
			}
		}
	}
	if !changed {
		return data
	}
	out, err := json.Marshal(m)
	if err != nil {
		return data
	}
	return out
}

// snakeToCamel converts a snake_case identifier to camelCase.
// Example: "agent_id" → "agentId".
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
