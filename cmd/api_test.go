package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPICommand_StdinNotTTY_ReturnsQuickly(t *testing.T) {
	// Simulate a non-TTY stdin with no data (like piping from /dev/null)
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	pw.Close() // EOF immediately — no data
	oldStdin := os.Stdin
	os.Stdin = pr
	defer func() { os.Stdin = oldStdin; pr.Close() }()

	// No mock needed — command should error before hitting API
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"api", "users/me"})
	// This must complete without hanging
	err = root.Execute()
	// We expect an error because no body was provided and stdin was empty
	assert.Error(t, err)
}

func TestAPICommand_Preview_WritesToCmdOut(t *testing.T) {
	root := NewCmdRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"api", "search", "--method", "POST", "--raw-field", `{"query":"test"}`, "--preview"})
	_ = root.Execute()
	// Preview output must appear in buf (cmd.OutOrStdout())
	// It should contain the HTTP method and endpoint info
	assert.NotEmpty(t, buf.String(), "preview output should be written to cmd.OutOrStdout()")
}

func TestAPICommandPreviewShowsAuthHeader(t *testing.T) {
	// Inject a test token via the env var that config.LoadConfig actually reads.
	t.Setenv("GLEAN_API_TOKEN", "test-token-for-preview")
	t.Setenv("GLEAN_SERVER_URL", "https://test.glean.com")

	b := bytes.NewBufferString("")
	cmd := NewCmdAPI()
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--preview", "--method", "POST", "--raw-field", `{"query":"test"}`, "search"})
	err := cmd.Execute()
	require.NoError(t, err)
	out := b.String()
	assert.Contains(t, out, "Authorization:", "auth header must appear in preview")
	assert.NotContains(t, out, "Bearer \n", "auth token must not be empty")
}

func TestResolveAPIPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"search", "/rest/api/v1/search"},
		{"/search", "/rest/api/v1/search"},
		{"users/me", "/rest/api/v1/users/me"},
		{"/rest/api/v1/search", "/rest/api/v1/search"},
		{"rest/api/v1/search", "/rest/api/v1/search"},
		{"/api/search", "/api/search"},
		{"api/search", "/api/search"},
		{"/api/agents/search", "/api/agents/search"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveAPIPath(tt.in))
		})
	}
}

func TestSendExperimentalHeader(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		force    bool
		want     bool
	}{
		{"classic path, no flag", "search", false, false},
		{"classic path, forced", "search", true, true},
		{"platform path, no flag", "/api/search", false, true},
		{"platform path without slash", "api/search", false, true},
		{"explicit rest path, no flag", "/rest/api/v1/search", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sendExperimentalHeader(APIOptions{experimental: tt.force}, tt.endpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAPICommandPreviewExperimentalHeader(t *testing.T) {
	t.Setenv("GLEAN_API_TOKEN", "test-token-for-preview")
	t.Setenv("GLEAN_SERVER_URL", "https://test.glean.com")

	tests := []struct {
		name       string
		args       []string
		wantHeader bool
		wantURL    string
	}{
		{
			name:       "platform path sends header automatically",
			args:       []string{"--preview", "--method", "POST", "--raw-field", `{"query":"test"}`, "/api/search"},
			wantHeader: true,
			wantURL:    "https://test.glean.com/api/search",
		},
		{
			name:       "classic path omits header",
			args:       []string{"--preview", "--method", "POST", "--raw-field", `{"query":"test"}`, "search"},
			wantHeader: false,
			wantURL:    "https://test.glean.com/rest/api/v1/search",
		},
		{
			name:       "classic path with --experimental sends header",
			args:       []string{"--preview", "--method", "POST", "--raw-field", `{"query":"test"}`, "--experimental", "search"},
			wantHeader: true,
			wantURL:    "https://test.glean.com/rest/api/v1/search",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			cmd := NewCmdAPI()
			cmd.SetOut(b)
			cmd.SetArgs(tt.args)
			require.NoError(t, cmd.Execute())
			out := b.String()
			assert.Contains(t, out, tt.wantURL)
			if tt.wantHeader {
				assert.Contains(t, out, "X-Glean-Include-Experimental: true")
			} else {
				assert.NotContains(t, out, "X-Glean-Include-Experimental")
			}
		})
	}
}

func TestApiCmd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no endpoint provided",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			cmd := NewCmdAPI()
			cmd.SetOut(b)
			cmd.SetErr(b)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
