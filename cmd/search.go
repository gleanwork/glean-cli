package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gleanwork/api-client-go/models/components"
	gleanClient "github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/output"
	"github.com/gleanwork/glean-cli/internal/platform"
	"github.com/gleanwork/glean-cli/internal/search"
	"github.com/spf13/cobra"
)

// searchTextFn renders whichever search response shape the request produced:
// the platform response (default) or the classic response (legacy fallback).
func searchTextFn(w io.Writer, v any) error {
	switch resp := v.(type) {
	case *components.PlatformSearchResponse:
		return platformSearchTextFn(w, resp)
	case *components.SearchResponse:
		return legacySearchTextFn(w, resp)
	default:
		return output.WriteJSON(w, v)
	}
}

func platformSearchTextFn(w io.Writer, resp *components.PlatformSearchResponse) error {
	var rows [][]string
	for _, r := range resp.Results {
		snippet := ""
		if len(r.Snippets) > 0 {
			snippet = r.Snippets[0]
		}
		rows = append(rows, []string{
			output.Truncate(r.Title, 50),
			r.Datasource,
			r.URL,
			output.Truncate(snippet, 60),
		})
	}
	return output.WriteTable(w, []string{"TITLE", "SOURCE", "URL", "SNIPPET"}, rows)
}

func legacySearchTextFn(w io.Writer, resp *components.SearchResponse) error {
	var rows [][]string
	for _, r := range resp.Results {
		title, source, url, snippet := "", "", "", ""
		if r.Document != nil {
			if r.Document.Title != nil {
				title = *r.Document.Title
			}
			if r.Document.Datasource != nil {
				source = *r.Document.Datasource
			}
			if r.Document.URL != nil {
				url = *r.Document.URL
			}
		}
		if title == "" && r.Title != nil {
			title = *r.Title
		}
		if url == "" {
			url = r.URL
		}
		if len(r.Snippets) > 0 && r.Snippets[0].Text != nil {
			snippet = *r.Snippets[0].Text
		}
		rows = append(rows, []string{
			output.Truncate(title, 50),
			source,
			url,
			output.Truncate(snippet, 60),
		})
	}
	return output.WriteTable(w, []string{"TITLE", "SOURCE", "URL", "SNIPPET"}, rows)
}

// NewCmdSearch creates and returns the search command.
func NewCmdSearch() *cobra.Command {
	opts := &search.Options{
		RequestOptions: &search.RequestOptions{
			FacetBucketSize: 10,
			ResponseHints:   []string{"RESULTS", "QUERY_METADATA"},
		},
	}
	var jsonPayload string
	var outputFormat string
	var dryRun bool
	var fields string
	var raw bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for content in your Glean instance",
		Long: `Search for content in your Glean instance.

Search is served by the platform API (POST /api/search) and results use its
snake_case response shape. If the platform API is not enabled on your
instance, the command falls back to the classic API with a warning; set
GLEAN_LEGACY_APIS=1 to use the classic API directly.

Results are written as JSON to stdout by default, making the output easy to
pipe to jq or other tools.

Example:
  glean search "vacation policy"
  glean search "vacation policy" | jq '.results[].title'
  glean search --json '{"query":"Q1 reports","page_size":5}' | jq .
  glean search --output ndjson "engineering docs" | head -3 | jq .title
  glean search --dry-run "test"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonPayload == "" && len(args) == 0 {
				return fmt.Errorf("requires a query argument or --json payload")
			}

			// --json path: the payload shape is coupled to the endpoint, so a
			// user-authored body is never replayed against the other surface.
			// Platform shape by default; classic shape under GLEAN_LEGACY_APIS=1.
			if jsonPayload != "" {
				if platform.Legacy() {
					return runLegacyJSONSearch(cmd, jsonPayload, dryRun, outputFormat, raw)
				}
				var req components.PlatformSearchRequest
				if err := json.Unmarshal([]byte(jsonPayload), &req); err != nil {
					return fmt.Errorf("invalid --json payload: %w", err)
				}
				if dryRun {
					return output.WriteJSON(cmd.OutOrStdout(), req)
				}
				sdk, err := gleanClient.NewFromConfig()
				if err != nil {
					return err
				}
				resp, err := sdk.Search.Query(cmd.Context(), req)
				if err != nil {
					if platform.IsGateClosed(err) {
						return platform.GateClosedErr("/api/search", err)
					}
					return fmt.Errorf("search request failed: %w", err)
				}
				return output.WriteFormatted(cmd.OutOrStdout(), resp.PlatformSearchResponse, outputFormat, searchTextFn)
			}

			// flag-based path
			if datasources, flagErr := cmd.Flags().GetStringSlice("datasource"); flagErr == nil && len(datasources) > 0 {
				search.AddFacetFilter(opts, "datasource", datasources)
			}
			if types, flagErr := cmd.Flags().GetStringSlice("type"); flagErr == nil && len(types) > 0 {
				search.AddFacetFilter(opts, "type", types)
			}
			if tabs, flagErr := cmd.Flags().GetStringSlice("tab"); flagErr == nil && len(tabs) > 0 {
				opts.ResultTabIds = tabs
			}
			if opts.RequestOptions == nil {
				opts.RequestOptions = &search.RequestOptions{}
			}
			opts.RequestOptions.TimezoneOffset = search.GetTimezoneOffset()
			if size, flagErr := cmd.Flags().GetInt("facet-bucket-size"); flagErr == nil {
				opts.RequestOptions.FacetBucketSize = size
			}
			if hints, flagErr := cmd.Flags().GetStringSlice("response-hints"); flagErr == nil {
				opts.RequestOptions.ResponseHints = hints
			}
			if disable, flagErr := cmd.Flags().GetBool("disable-query-autocorrect"); flagErr == nil {
				opts.RequestOptions.DisableQueryAutocorrect = disable
			}
			if fetch, flagErr := cmd.Flags().GetBool("fetch-all-datasource-counts"); flagErr == nil {
				opts.RequestOptions.FetchAllDatasourceCounts = fetch
			}
			if override, flagErr := cmd.Flags().GetBool("query-overrides-facet-filters"); flagErr == nil {
				opts.RequestOptions.QueryOverridesFacetFilters = override
			}
			if llm, flagErr := cmd.Flags().GetBool("return-llm-content"); flagErr == nil {
				opts.RequestOptions.ReturnLlmContentOverSnippets = llm
			}

			opts.Query = args[0]

			if dryRun {
				if platform.Legacy() {
					return output.WriteJSON(cmd.OutOrStdout(), search.BuildSearchRequest(opts))
				}
				return output.WriteJSON(cmd.OutOrStdout(), search.BuildPlatformSearchRequest(opts))
			}

			sdk, err := gleanClient.NewFromConfig()
			if err != nil {
				return err
			}
			if !platform.Legacy() {
				if ignored := platformIgnoredFlags(cmd); len(ignored) > 0 {
					fmt.Fprintf(cmd.ErrOrStderr(),
						"Note: flags ignored by platform search (legacy-only): %s\n",
						strings.Join(ignored, ", "))
				}
			}

			result, viaLegacy, err := platform.Run(cmd.Context(), cmd.ErrOrStderr(), "/api/search",
				func(ctx context.Context) (any, error) { return search.RunPlatformSearch(ctx, opts, sdk) },
				func(ctx context.Context) (any, error) { return search.RunSearchSDK(ctx, opts, sdk) },
			)
			if err != nil {
				return err
			}
			// Text output operates on the raw SDK struct directly;
			// cleansing is redundant since the table selects its own fields.
			if outputFormat == output.OutputText {
				return output.WriteFormatted(cmd.OutOrStdout(), result, outputFormat, searchTextFn)
			}
			// The classic response carries superfluous properties and is
			// cleansed unless --raw; platform responses are emitted as-is.
			if viaLegacy && !raw {
				result, err = output.CleanseSearchResponse(result)
				if err != nil {
					return err
				}
			}
			if fields != "" {
				if viaLegacy && !raw {
					if stripped := output.WarnStrippedFields(fields); len(stripped) > 0 {
						fmt.Fprintf(cmd.ErrOrStderr(),
							"Warning: field(s) %s not available in cleansed output (use --raw for the full response)\n",
							strings.Join(stripped, ", "))
					}
				}
				return output.ProjectFields(cmd.OutOrStdout(), result, fields)
			}
			return output.WriteFormatted(cmd.OutOrStdout(), result, outputFormat, searchTextFn)
		},
	}

	cmd.Flags().StringVar(&jsonPayload, "json", "", "Complete JSON request body in the platform shape (overrides all other flags); with GLEAN_LEGACY_APIS=1, parsed as the classic shape")
	cmd.Flags().StringVar(&outputFormat, "output", "json", "Output format: json, ndjson, or text")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated dot-path fields to include (e.g. results.document.title,results.document.url). Results where all projected fields are missing appear as {}")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the request body without sending it")
	cmd.Flags().BoolVar(&raw, "raw", false, "Output the full SDK response without cleansing (classic API responses only; platform responses are never cleansed)")
	cmd.Flags().IntVar(&opts.PageSize, "page-size", 10, "Number of results per page")
	cmd.Flags().IntVar(&opts.MaxSnippetSize, "max-snippet-size", 0, "Maximum size of snippets")
	cmd.Flags().IntVar(&opts.TimeoutMillis, "timeout", 30000, "Request timeout in milliseconds")
	cmd.Flags().BoolVar(&opts.DisableSpellcheck, "disable-spellcheck", false, "Disable spellcheck")

	cmd.Flags().StringSliceP("datasource", "d", nil, "Filter by datasource (can be specified multiple times)")
	cmd.Flags().StringSliceP("type", "t", nil, "Filter by document type (can be specified multiple times)")
	cmd.Flags().StringSlice("tab", nil, "Filter by result tab IDs (can be specified multiple times)")

	cmd.Flags().Bool("disable-query-autocorrect", false, "Disable automatic query corrections")
	cmd.Flags().Bool("fetch-all-datasource-counts", false, "Return result counts for all supported datasources")
	cmd.Flags().Bool("query-overrides-facet-filters", false, "Let query operators override facet filters")
	cmd.Flags().Bool("return-llm-content", false, "Return expanded content for LLM usage")
	cmd.Flags().StringSlice("response-hints", []string{"RESULTS", "QUERY_METADATA"}, "Response hints (RESULTS, QUERY_METADATA, etc)")
	cmd.Flags().Int("facet-bucket-size", 10, "Maximum number of facet buckets to return")

	return cmd
}

// platformIgnoredFlags returns the explicitly-set search flags that have no
// platform API equivalent, so users learn their flag did nothing rather than
// silently getting unfiltered results.
func platformIgnoredFlags(cmd *cobra.Command) []string {
	legacyOnly := []string{
		"max-snippet-size",
		"disable-spellcheck",
		"tab",
		"response-hints",
		"facet-bucket-size",
		"disable-query-autocorrect",
		"fetch-all-datasource-counts",
		"query-overrides-facet-filters",
		"return-llm-content",
	}
	var ignored []string
	for _, name := range legacyOnly {
		if cmd.Flags().Changed(name) {
			ignored = append(ignored, "--"+name)
		}
	}
	return ignored
}

// runLegacyJSONSearch handles --json under GLEAN_LEGACY_APIS=1: the payload
// is parsed as the classic /rest/api/v1/search request shape and the
// response is cleansed unless --raw, exactly as before the platform
// migration.
func runLegacyJSONSearch(cmd *cobra.Command, jsonPayload string, dryRun bool, outputFormat string, raw bool) error {
	var req components.SearchRequest
	if err := json.Unmarshal([]byte(jsonPayload), &req); err != nil {
		return fmt.Errorf("invalid --json payload: %w", err)
	}
	if dryRun {
		return output.WriteJSON(cmd.OutOrStdout(), req)
	}
	sdk, err := gleanClient.NewFromConfig()
	if err != nil {
		return err
	}
	resp, err := sdk.Client.Search.Query(cmd.Context(), req, nil)
	if err != nil {
		return fmt.Errorf("search request failed: %w", err)
	}
	// Text output operates on the raw SDK struct directly;
	// cleansing is redundant since the table selects its own fields.
	if outputFormat == output.OutputText {
		return output.WriteFormatted(cmd.OutOrStdout(), resp.SearchResponse, outputFormat, searchTextFn)
	}
	var result any = resp.SearchResponse
	if !raw {
		result, err = output.CleanseSearchResponse(result)
		if err != nil {
			return err
		}
	}
	return output.WriteFormatted(cmd.OutOrStdout(), result, outputFormat, searchTextFn)
}
