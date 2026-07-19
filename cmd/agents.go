package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
	"github.com/gleanwork/glean-cli/internal/cmdutil"
	"github.com/gleanwork/glean-cli/internal/output"
	"github.com/spf13/cobra"
)

// agentIDRequest is a CLI-only request struct for commands whose only input
// is an agent ID. The snake_case tag matches the platform input shape;
// cmdutil transparently normalizes camelCase agentId.
type agentIDRequest struct {
	AgentID string `json:"agent_id"`
}

// agentRunRequest is the CLI request for `agents run`. The platform route
// takes the agent ID as a path parameter (POST /api/agents/{agent_id}/runs),
// so the CLI folds it into the --json body alongside the run inputs.
type agentRunRequest struct {
	AgentID  string                       `json:"agent_id"`
	Input    map[string]any               `json:"input,omitempty"`
	Messages []components.PlatformMessage `json:"messages,omitempty"`
	Metadata map[string]any               `json:"metadata,omitempty"`
}

func NewCmdAgents() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage and run Glean agents",
		Long: `Manage and run Glean agents.

Agents are AI-powered workflows that can search, reason, and act on your company's knowledge.

Agent commands are served by the platform API (/api/agents/...) and responses
use its snake_case shape. list/get/schemas fall back to the classic API with a
warning when the platform API is not enabled; run does not fall back
automatically because its request body differs between the two surfaces. Set
GLEAN_LEGACY_APIS=1 to use the classic API directly (run then expects the
classic messages/fragments body).

Example:
  glean agents list
  glean agents get --json '{"agent_id":"<id>"}'
  glean agents schemas --json '{"agent_id":"<id>"}'
  glean agents run --json '{"agent_id":"<id>","input":{"query":"summarize Q1 results"}}'
  glean agents run --json '{"agent_id":"<id>","messages":[{"role":"user","content":[{"type":"text","text":"summarize Q1 results"}]}]}'`,
	}
	cmd.AddCommand(
		newAgentsListCmd(),
		newAgentsGetCmd(),
		newAgentsSchemasCmd(),
		newAgentsRunCmd(),
	)
	return cmd
}

// agentsListTextFn renders whichever agents list shape the request produced:
// the platform response (default) or the classic response (legacy fallback).
func agentsListTextFn(w io.Writer, v any) error {
	header := []string{"ID", "NAME", "DESCRIPTION"}
	switch resp := v.(type) {
	case *components.PlatformAgentsSearchResponse:
		rows := make([][]string, len(resp.Agents))
		for i, a := range resp.Agents {
			desc := ""
			if a.Description != nil {
				desc = output.Truncate(*a.Description, 60)
			}
			rows[i] = []string{a.AgentID, a.Name, desc}
		}
		return output.WriteTable(w, header, rows)
	case *components.SearchAgentsResponse:
		rows := make([][]string, len(resp.Agents))
		for i, a := range resp.Agents {
			desc := ""
			if a.Description != nil {
				desc = output.Truncate(*a.Description, 60)
			}
			rows[i] = []string{a.AgentID, a.Name, desc}
		}
		return output.WriteTable(w, header, rows)
	default:
		return output.WriteJSON(w, v)
	}
}

func newAgentsListCmd() *cobra.Command {
	return cmdutil.Build(cmdutil.Spec[components.PlatformAgentsSearchRequest]{
		Use:      "list",
		Short:    "List available agents",
		Endpoint: "/api/agents/search",
		TextFn:   agentsListTextFn,
		Run: func(ctx context.Context, sdk *glean.Glean, req components.PlatformAgentsSearchRequest) (any, error) {
			resp, err := sdk.Agents.Search(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.PlatformAgentsSearchResponse, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			var req components.SearchAgentsRequest
			if len(rawJSON) > 0 {
				if err := json.Unmarshal(rawJSON, &req); err != nil {
					return nil, fmt.Errorf("invalid --json: %w", err)
				}
			}
			resp, err := sdk.Client.Agents.List(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.SearchAgentsResponse, nil
		},
	})
}

// parseAgentIDRequest is the shared LegacyRun payload parse for get/schemas:
// the input is a bare agent ID, identical on both surfaces.
func parseAgentIDRequest(rawJSON []byte) (agentIDRequest, error) {
	var req agentIDRequest
	if err := json.Unmarshal(rawJSON, &req); err != nil {
		return req, fmt.Errorf("invalid --json: %w", err)
	}
	return req, nil
}

func newAgentsGetCmd() *cobra.Command {
	return cmdutil.Build(cmdutil.Spec[agentIDRequest]{
		Use:          "get",
		Short:        "Get an agent by ID",
		JSONRequired: true,
		Endpoint:     "/api/agents/{agent_id}",
		Run: func(ctx context.Context, sdk *glean.Glean, req agentIDRequest) (any, error) {
			resp, err := sdk.Agents.Get(ctx, req.AgentID)
			if err != nil {
				return nil, err
			}
			return resp.PlatformAgentGetResponse, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			req, err := parseAgentIDRequest(rawJSON)
			if err != nil {
				return nil, err
			}
			resp, err := sdk.Client.Agents.Retrieve(ctx, req.AgentID, nil, nil)
			if err != nil {
				return nil, err
			}
			return resp.Agent, nil
		},
	})
}

func newAgentsSchemasCmd() *cobra.Command {
	return cmdutil.Build(cmdutil.Spec[agentIDRequest]{
		Use:          "schemas",
		Short:        "Get the schemas for an agent",
		JSONRequired: true,
		Endpoint:     "/api/agents/{agent_id}/schemas",
		Run: func(ctx context.Context, sdk *glean.Glean, req agentIDRequest) (any, error) {
			resp, err := sdk.Agents.GetSchemas(ctx, req.AgentID, nil)
			if err != nil {
				return nil, err
			}
			return resp.PlatformAgentSchemasResponse, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			req, err := parseAgentIDRequest(rawJSON)
			if err != nil {
				return nil, err
			}
			resp, err := sdk.Client.Agents.RetrieveSchemas(ctx, req.AgentID, nil, nil)
			if err != nil {
				return nil, err
			}
			return resp.AgentSchemas, nil
		},
	})
}

func newAgentsRunCmd() *cobra.Command {
	return cmdutil.Build(cmdutil.Spec[agentRunRequest]{
		Use:          "run",
		Short:        "Run an agent (synchronous)",
		JSONRequired: true,
		Endpoint:     "/api/agents/{agent_id}/runs",
		// The run body is user-authored and shaped differently on the two
		// surfaces (platform input/messages vs classic messages/fragments),
		// so gate-closed errors instead of silently replaying the payload.
		FallbackMode: cmdutil.FallbackEnvOnly,
		Run: func(ctx context.Context, sdk *glean.Glean, req agentRunRequest) (any, error) {
			if req.AgentID == "" {
				return nil, fmt.Errorf("agent_id is required in the --json payload")
			}
			resp, err := sdk.Agents.CreateRun(ctx, req.AgentID, components.PlatformAgentRunCreateRequest{
				Input:    req.Input,
				Messages: req.Messages,
				Metadata: req.Metadata,
			})
			if err != nil {
				return nil, err
			}
			return resp.PlatformAgentRunWaitResponse, nil
		},
		LegacyRun: func(ctx context.Context, sdk *glean.Glean, rawJSON []byte) (any, error) {
			var req components.AgentRunCreate
			if err := json.Unmarshal(rawJSON, &req); err != nil {
				return nil, fmt.Errorf("invalid --json: %w", err)
			}
			resp, err := sdk.Client.Agents.Run(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.AgentRunWaitResponse, nil
		},
	})
}
