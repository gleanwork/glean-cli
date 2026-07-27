package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	gleanClient "github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/config"
	"github.com/gleanwork/glean-cli/internal/httputil"
	"github.com/gleanwork/glean-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// experimentalHeader opts a request into experimental (platform) API
// endpoints. Without it, experimental endpoints return a hidden 404.
const experimentalHeader = "X-Glean-Include-Experimental"

// APIOptions holds configuration for the API command.
type APIOptions struct {
	method       string
	requestBody  string
	inputFile    string
	preview      bool
	raw          bool
	noColor      bool
	experimental bool
}

// NewCmdAPI creates and returns the api command.
func NewCmdAPI() *cobra.Command {
	opts := APIOptions{}

	cmd := &cobra.Command{
		Use:   "api",
		Short: "Make an authenticated HTTP request to the Glean API",
		Long: heredoc.Doc(`
			Makes an authenticated HTTP request to the Glean API and prints the response.

			The endpoint argument should be a path of a Glean API endpoint. Bare paths
			default to the classic REST API under /rest/api/v1/:
			  glean api search
			  glean api users/me

			Paths starting with /rest/ or /api/ are used verbatim, so the new platform
			APIs are reachable directly. Requests to /api/* automatically send the
			X-Glean-Include-Experimental header (use --experimental to force it on
			classic endpoints):
			  glean api /api/search --method POST --raw-field '{"query": "test"}'

			The default HTTP request method is "GET". To use a different method,
			use the --method flag:
			  glean api --method POST search

			Request body can be provided via --raw-field, --field, or --input:
			  echo '{"query": "rust programming"}' | glean api --method POST search
			  glean api --method POST search --raw-field '{"query": "rust programming"}'
			  glean api --method POST search --input request.json

			Pass --preview to print the request details without actually sending it.
			Pass --no-color to disable colorized output (useful when piping to jq).
		`),
		Example: heredoc.Doc(`
			# Get the current user
			$ glean api users/me

			# Search with parameters
			$ glean api search --method POST --raw-field '{"query": "rust programming"}'

			# Platform API search (experimental header sent automatically)
			$ glean api /api/search --method POST --raw-field '{"query": "rust programming"}'

			# Search with parameters from a file
			$ glean api search --method POST --input search-params.json

			# Preview the request
			$ glean api search --method POST --raw-field '{"query": "test"}' --preview

			# Pipe to jq
			$ glean api search --no-color | jq .results
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			endpoint := args[0]

			var body map[string]interface{}
			if opts.requestBody != "" {
				if jsonErr := json.Unmarshal([]byte(opts.requestBody), &body); jsonErr != nil {
					return jsonErr
				}
			}
			if opts.inputFile != "" {
				data, readErr := os.ReadFile(opts.inputFile)
				if readErr != nil {
					return readErr
				}
				if jsonErr := json.Unmarshal(data, &body); jsonErr != nil {
					return jsonErr
				}
			}
			if opts.inputFile == "" && opts.requestBody == "" {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					stdinData, readErr := io.ReadAll(os.Stdin)
					if readErr != nil {
						return fmt.Errorf("reading stdin: %w", readErr)
					}
					if len(stdinData) > 0 {
						if jsonErr := json.Unmarshal(stdinData, &body); jsonErr != nil {
							return fmt.Errorf("invalid JSON from stdin: %w", jsonErr)
						}
					}
				} else {
					return fmt.Errorf("provide a request body via --raw-field, --input, or pipe from stdin")
				}
			}

			if opts.preview {
				return previewRequest(cmd, cfg, opts, endpoint, body)
			}

			useSpinner := term.IsTerminal(int(os.Stderr.Fd())) && !opts.raw && !opts.noColor
			var s *spinner.Spinner
			if useSpinner {
				s = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
				s.Suffix = " Making request to Glean API..."
				s.Writer = os.Stderr
				s.FinalMSG = "Request complete!\n"
				s.Start()
				defer s.Stop()
			}

			resp, err := rawAPIRequest(cmd.Context(), cfg, opts, endpoint, body)
			if err != nil {
				return err
			}

			return output.Write(os.Stdout, resp, output.Options{
				NoColor: opts.raw || opts.noColor,
				Format:  "json",
			})
		},
	}

	cmd.Flags().StringVarP(&opts.method, "method", "X", "GET", "The HTTP method for the request")
	cmd.Flags().StringVar(&opts.requestBody, "raw-field", "", "Add a JSON string as the request body")
	cmd.Flags().StringVarP(&opts.inputFile, "input", "F", "", "The file to use as body for the request")
	cmd.Flags().BoolVar(&opts.preview, "preview", false, "Preview the API request without sending it")
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "Print raw API response")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable colorized output")
	cmd.Flags().BoolVar(&opts.experimental, "experimental", false, "Send the "+experimentalHeader+" header (automatic for /api/* paths)")

	return cmd
}

// resolveAPIPath normalizes an endpoint argument to a server-relative path.
// Paths under /rest/ or /api/ are used verbatim (the platform APIs live at
// /api/*); anything else keeps the historical default prefix /rest/api/v1/.
func resolveAPIPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasPrefix(path, "/rest/") || strings.HasPrefix(path, "/api/") {
		return path
	}
	return "/rest/api/v1" + path
}

// apiFullURL returns the full API URL for an endpoint path.
func apiFullURL(cfg *config.Config, path string) string {
	return cfg.GleanServerURL + resolveAPIPath(path)
}

// sendExperimentalHeader reports whether the request should carry the
// X-Glean-Include-Experimental header: always for platform (/api/*) paths —
// without it experimental endpoints answer with a hidden 404 that reads like
// a typo'd path — or when forced via --experimental.
func sendExperimentalHeader(opts APIOptions, endpoint string) bool {
	return opts.experimental || strings.HasPrefix(resolveAPIPath(endpoint), "/api/")
}

// rawAPIRequest makes an authenticated HTTP request to the Glean API.
func rawAPIRequest(ctx context.Context, cfg *config.Config, opts APIOptions, endpoint string, body map[string]interface{}) ([]byte, error) {
	url := apiFullURL(cfg, endpoint)

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, opts.method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	token, authType := gleanClient.ResolveToken(cfg)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if authType != "" {
		req.Header.Set("X-Glean-Auth-Type", authType)
	}
	if sendExperimentalHeader(opts, endpoint) {
		req.Header.Set(experimentalHeader, "true")
	}

	httpClient := httputil.NewHTTPClient(30 * time.Second)
	httpResp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		var errResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			if errResp.Message != "" {
				return nil, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, errResp.Message)
			}
			if errResp.Error != "" {
				return nil, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, errResp.Error)
			}
		}
		return nil, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func previewRequest(cmd *cobra.Command, cfg *config.Config, opts APIOptions, endpoint string, body map[string]interface{}) error {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Request Method: %s\n", opts.method)
	fmt.Fprintf(w, "Request URL: %s\n", apiFullURL(cfg, endpoint))
	fmt.Fprintf(w, "\nRequest Headers:\n")
	fmt.Fprintf(w, "  Content-Type: application/json\n")
	token, authType := gleanClient.ResolveToken(cfg)
	if token != "" {
		fmt.Fprintf(w, "  Authorization: Bearer %s\n", config.MaskToken(token))
	}
	if authType != "" {
		fmt.Fprintf(w, "  X-Glean-Auth-Type: %s\n", authType)
	}
	if sendExperimentalHeader(opts, endpoint) {
		fmt.Fprintf(w, "  %s: true\n", experimentalHeader)
	}

	if body != nil {
		fmt.Fprintf(w, "\nRequest Body:\n")
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to format request body: %w", err)
		}
		return output.Write(w, bodyBytes, output.Options{
			NoColor: opts.noColor,
			Format:  "json",
		})
	}
	return nil
}
