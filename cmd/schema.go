package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/gleanwork/glean-cli/internal/schema"
	"github.com/spf13/cobra"
)

func init() {
	// Register schemas for all commands at startup.
	schema.Register(schema.CommandSchema{
		Command:     "search",
		Description: "Search for content in your Glean instance via the platform API (POST /api/search); responses use its snake_case shape. Falls back to the classic API with a warning when the platform API is not enabled; GLEAN_LEGACY_APIS=1 forces the classic API. Results are JSON.",
		WhenToUse:   "Find documents, messages, and content across all connected datasources. Start here for any 'find X' task.",
		Surface:     schema.SurfacePlatform,
		Flags: map[string]schema.FlagSchema{
			"--json":                          {Type: "string", Description: "Complete JSON request body in the platform shape: query, page_size, cursor, datasources, datasource_instances, filters, time_range (overrides individual flags). With GLEAN_LEGACY_APIS=1, parsed as the classic shape", Required: false},
			"--query":                         {Type: "string", Description: "Search query (positional arg)", Required: true},
			"--page-size":                     {Type: "integer", Default: 10, Description: "Number of results per page"},
			"--max-snippet-size":              {Type: "integer", Default: 0, Description: "Maximum snippet size in characters (legacy API only)"},
			"--timeout":                       {Type: "integer", Default: 30000, Description: "Request timeout in milliseconds"},
			"--disable-spellcheck":            {Type: "boolean", Default: false, Description: "Disable spellcheck (legacy API only)"},
			"--datasource":                    {Type: "[]string", Description: "Filter by datasource (repeatable)"},
			"--type":                          {Type: "[]string", Description: "Filter by document type (repeatable)"},
			"--tab":                           {Type: "[]string", Description: "Filter by result tab IDs (repeatable, legacy API only)"},
			"--response-hints":                {Type: "[]string", Default: []string{"RESULTS", "QUERY_METADATA"}, Description: "Response hints (legacy API only)"},
			"--facet-bucket-size":             {Type: "integer", Default: 10, Description: "Maximum facet buckets per result (legacy API only)"},
			"--disable-query-autocorrect":     {Type: "boolean", Default: false, Description: "Disable automatic query corrections (legacy API only)"},
			"--fetch-all-datasource-counts":   {Type: "boolean", Default: false, Description: "Return counts for all datasources (legacy API only)"},
			"--query-overrides-facet-filters": {Type: "boolean", Default: false, Description: "Allow query operators to override facet filters (legacy API only)"},
			"--return-llm-content":            {Type: "boolean", Default: false, Description: "Return expanded LLM-friendly content (legacy API only)"},
			"--output":                        {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json", Description: "Output format"},
			"--dry-run":                       {Type: "boolean", Default: false, Description: "Print request body without sending"},
		},
		Example: `glean search "vacation policy" | jq '.results[].title'
glean search --json '{"query":"Q1 reports","page_size":5,"datasources":["confluence"]}' | jq .`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "chat",
		Description: "Have a conversation with Glean AI. Streams response to stdout.",
		WhenToUse:   "Ask Glean AI a question that needs a synthesized, cited answer reasoned over company knowledge, rather than a raw list of results.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "Complete JSON chat request body (overrides individual flags)"},
			"--message": {Type: "string", Description: "Chat message (positional arg)", Required: true},
			"--timeout": {Type: "integer", Default: 30000, Description: "Request timeout in milliseconds"},
			"--save":    {Type: "boolean", Default: true, Description: "Save the chat session"},
			"--dry-run": {Type: "boolean", Default: false, Description: "Print request body without sending"},
		},
		Example: `glean chat "What are the company holidays?"
glean chat --json '{"messages":[{"author":"USER","messageType":"CONTENT","fragments":[{"text":"What is Glean?"}]}]}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "api",
		Description: "Make a raw authenticated HTTP request to any Glean API endpoint. Bare paths default to /rest/api/v1/; /rest/* and /api/* paths are used verbatim. Requests to /api/* (platform APIs) automatically send the X-Glean-Include-Experimental header.",
		WhenToUse:   "Escape hatch: call any Glean API endpoint directly (classic /rest/api/v1/* or platform /api/*) when no dedicated subcommand exists.",
		Surface:     schema.SurfaceRaw,
		Flags: map[string]schema.FlagSchema{
			"--method":       {Type: "enum", Enum: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}, Default: "GET", Description: "HTTP method"},
			"--raw-field":    {Type: "string", Description: "JSON request body as a string"},
			"--input":        {Type: "string", Description: "Path to a JSON file to use as request body"},
			"--preview":      {Type: "boolean", Default: false, Description: "Print request details without sending"},
			"--raw":          {Type: "boolean", Default: false, Description: "Print raw response without syntax highlighting"},
			"--no-color":     {Type: "boolean", Default: false, Description: "Disable colorized output"},
			"--experimental": {Type: "boolean", Default: false, Description: "Send the X-Glean-Include-Experimental header (automatic for /api/* paths)"},
			"--dry-run":      {Type: "boolean", Default: false, Description: "Same as --preview"},
		},
		Example: `glean api search --method POST --raw-field '{"query":"test"}' --no-color | jq .results
glean api /api/search --method POST --raw-field '{"query":"test"}' | jq .results`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "shortcuts",
		Description: "Manage Glean shortcuts (go-links). Subcommands: list, get, create, update, delete.",
		WhenToUse:   "Create or resolve go-links (short memorable URLs like go/roadmap).",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (see Glean API docs for shape)"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false, Description: "Print request without sending"},
		},
		Example: `glean shortcuts list | jq '.results[].inputAlias'
glean shortcuts create --json '{"data":{"inputAlias":"test/link","destinationUrl":"https://example.com"}}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "agents",
		Description: "Manage and run Glean agents via the platform API (/api/agents/...). list/get/schemas fall back to the classic API when the platform API is not enabled; run does not (its body shape differs per surface). GLEAN_LEGACY_APIS=1 forces the classic API. Subcommands: list, get, schemas, run.",
		WhenToUse:   "Discover and execute Glean agents (AI workflows). Use schemas first to learn an agent's expected input, then run.",
		Surface:     schema.SurfacePlatform,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body. get/schemas: {\"agent_id\":\"<id>\"}. run: {\"agent_id\":\"<id>\"} plus input (map) or messages ([{role, content:[{type:\"text\",text}]}])"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean agents list | jq '.agents[].agent_id'
glean agents run --json '{"agent_id":"my-agent","input":{"query":"test"}}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "documents",
		Description: "Retrieve and summarize Glean documents. Subcommands: get, get-by-facets, get-permissions, summarize.",
		WhenToUse:   "Fetch full metadata, permissions, or an AI summary for documents you already identified (e.g. from search results).",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean documents summarize --json '{"documentId":"DOC_ID"}' | jq .summary`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "entities",
		Description: "List and read Glean entities and people. Subcommands: list, read-people.",
		WhenToUse:   "Look up people (org info, contact details) or other structured entities.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":   {Type: "string", Description: "JSON request body", Required: true},
			"--output": {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
		},
		Example: `glean entities read-people --json '{"query":"smith"}' | jq '.[].name'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "collections",
		Description: "Manage Glean collections. Subcommands: create, delete, update, add-items, delete-item.",
		WhenToUse:   "Curate sets of documents into named collections.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean collections create --json '{"name":"My Collection"}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "pins",
		Description: "Manage Glean pins. Subcommands: list, get, create, update, remove.",
		WhenToUse:   "Pin a document to a search query so it always surfaces for that query.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean pins list | jq '.[].id'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "answers",
		Description: "Manage Glean answers. Subcommands: list, get, create, update, delete.",
		WhenToUse:   "Manage curated Q&A answer cards shown for matching queries.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean answers list | jq '.[].id'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "tools",
		Description: "List and run Glean tools. Subcommands: list, run.",
		WhenToUse:   "Discover and execute Glean tools (actions) such as creating tickets or sending messages.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean tools list | jq '.[].name'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "verification",
		Description: "Manage document verification. Subcommands: list, verify, remind.",
		WhenToUse:   "Track and update document freshness verification (verify docs, send reminders).",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean verification list | jq '.[].document.title'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "activity",
		Description: "Report user activity and feedback. Subcommands: report, feedback.",
		WhenToUse:   "Report page-view activity events or submit result feedback to improve ranking.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (required)", Required: true},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean activity report --json '{"events":[{"action":"VIEW","url":"https://example.com"}]}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "insights",
		Description: "Retrieve Glean usage insights. Subcommands: get.",
		WhenToUse:   "Retrieve aggregate usage analytics (search/AI adoption metrics) for the deployment.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (required)", Required: true},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean insights get --json '{"insightTypes":["SEARCH"]}' | jq .`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "messages",
		Description: "Retrieve Glean messages. Subcommands: get.",
		WhenToUse:   "Fetch a specific chat/communication message by ID.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":   {Type: "string", Description: "JSON request body (required)", Required: true},
			"--output": {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
		},
		Example: `glean messages get --json '{"messageId":"MSG_ID"}' | jq .`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "agent-help",
		Description: "Show agent-facing usage generated from the installed glean binary: environment context plus a command map, or full detail for one command.",
		WhenToUse:   "Start every session here: discover available commands, exact flags, payload shapes, and whether the environment is authenticated — without static docs in context.",
		Surface:     schema.SurfaceLocal,
		Flags: map[string]schema.FlagSchema{
			"--json": {Type: "boolean", Default: false, Description: "Emit machine-readable JSON"},
		},
		Example: `glean agent-help
glean agent-help agents run --json`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "announcements",
		Description: "Manage Glean announcements. Subcommands: create, update, delete.",
		WhenToUse:   "Publish or manage company announcements surfaced in Glean.",
		Surface:     schema.SurfaceLegacy,
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (required)", Required: true},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean announcements create --json '{"title":"Company Update","body":"..."}'`,
	})
}

// NewCmdSchema creates and returns the schema command.
//
// Deprecated: superseded by `glean agent-help`, which derives structure from
// the live command tree and adds context awareness. Kept as a hidden alias
// for compatibility with existing agent skills that call `glean schema`.
func NewCmdSchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "schema [command]",
		Hidden: true,
		Short:  "Show JSON schema for a command's flags and request format (deprecated: use agent-help)",
		Long: `Show machine-readable JSON schema for any glean command.

Agents can call this before invoking a command to understand parameter
types, required fields, defaults, and example invocations — without
needing documentation in context.

Examples:
  glean schema             # list all commands with registered schemas
  glean schema search      # full schema for the search command
  glean schema chat        # full schema for the chat command`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// List all registered commands
				names := schema.List()
				out := map[string][]string{"commands": names}
				data, err := json.MarshalIndent(out, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			s, err := schema.Get(args[0])
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(s, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	return cmd
}
