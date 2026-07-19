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
		Command:     "version",
		Description: "Print the glean CLI version string.",
		Flags:       map[string]schema.FlagSchema{},
		Example:     `glean version`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "shortcuts",
		Description: "Manage Glean shortcuts (go-links). Subcommands: list, get, create, update, delete.",
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
		Description: "Manage and run Glean agents. Subcommands: list, get, schemas, run.",
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body"},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean agents list | jq '.[].id'
glean agents run --json '{"agentId":"my-agent","input":{"query":"test"}}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "documents",
		Description: "Retrieve and summarize Glean documents. Subcommands: get, get-by-facets, get-permissions, summarize.",
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
		Flags: map[string]schema.FlagSchema{
			"--json":   {Type: "string", Description: "JSON request body", Required: true},
			"--output": {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
		},
		Example: `glean entities read-people --json '{"query":"smith"}' | jq '.[].name'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "collections",
		Description: "Manage Glean collections. Subcommands: create, delete, update, add-items, delete-item.",
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
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (required)", Required: true},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean activity report --json '{"events":[{"action":"VIEW","url":"https://example.com"}]}'`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "insights",
		Description: "Retrieve Glean usage insights. Subcommands: get.",
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
		Flags: map[string]schema.FlagSchema{
			"--json":   {Type: "string", Description: "JSON request body (required)", Required: true},
			"--output": {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
		},
		Example: `glean messages get --json '{"messageId":"MSG_ID"}' | jq .`,
	})

	schema.Register(schema.CommandSchema{
		Command:     "announcements",
		Description: "Manage Glean announcements. Subcommands: create, update, delete.",
		Flags: map[string]schema.FlagSchema{
			"--json":    {Type: "string", Description: "JSON request body (required)", Required: true},
			"--output":  {Type: "enum", Enum: []string{"json", "ndjson", "text"}, Default: "json"},
			"--dry-run": {Type: "boolean", Default: false},
		},
		Example: `glean announcements create --json '{"title":"Company Update","body":"..."}'`,
	})
}

// NewCmdSchema creates and returns the schema command.
func NewCmdSchema() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [command]",
		Short: "Show JSON schema for a command's flags and request format",
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
