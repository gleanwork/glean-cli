package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/glean-cli/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPlatformSearchRequest_Basic(t *testing.T) {
	opts := &Options{Query: "vacation policy", PageSize: 25}
	req := BuildPlatformSearchRequest(opts)

	assert.Equal(t, "vacation policy", req.Query)
	require.NotNil(t, req.PageSize)
	assert.Equal(t, int64(25), *req.PageSize)
	assert.Nil(t, req.Cursor)
	assert.Empty(t, req.Datasources)
	assert.Empty(t, req.Filters)
}

func TestBuildPlatformSearchRequest_ZeroPageSizeOmitted(t *testing.T) {
	req := BuildPlatformSearchRequest(&Options{Query: "q"})
	assert.Nil(t, req.PageSize, "unset page size defers to the API default")
}

func TestBuildPlatformSearchRequest_Cursor(t *testing.T) {
	req := BuildPlatformSearchRequest(&Options{Query: "q", Cursor: "next-page-token"})
	require.NotNil(t, req.Cursor)
	assert.Equal(t, "next-page-token", *req.Cursor)
}

func TestBuildPlatformSearchRequest_DatasourceUsesDedicatedField(t *testing.T) {
	opts := &Options{Query: "q", RequestOptions: &RequestOptions{}}
	AddFacetFilter(opts, "datasource", []string{"confluence", "slack"})
	AddFacetFilter(opts, "type", []string{"document"})

	req := BuildPlatformSearchRequest(opts)

	assert.Equal(t, []string{"confluence", "slack"}, req.Datasources)
	require.Len(t, req.Filters, 1)
	assert.Equal(t, "type", req.Filters[0].Field)
	assert.Equal(t, []string{"document"}, req.Filters[0].Values)
}

func TestRunPlatformSearch_HitsPlatformEndpoint(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"results": []map[string]any{{
			"url":        "https://example.com/doc",
			"title":      "Example Doc",
			"datasource": "confluence",
			"snippets":   []string{"an excerpt"},
		}},
		"has_more":   false,
		"request_id": "req-123",
	})
	require.NoError(t, err)

	mock := &testutils.MockTransport{Body: body, ContentType: "application/json"}
	sdk := glean.New(
		glean.WithServerURL("https://test-company-be.glean.com"),
		glean.WithSecurity("test-token"),
		glean.WithClient(mock),
	)

	resp, err := RunPlatformSearch(context.Background(), &Options{Query: "example", TimeoutMillis: 5000}, sdk)
	require.NoError(t, err)

	require.Len(t, mock.Requests, 1)
	sent := mock.Requests[0]
	assert.Equal(t, "/api/search", sent.URL.Path)
	assert.Equal(t, http.MethodPost, sent.Method)

	sentBody, err := io.ReadAll(sent.Body)
	require.NoError(t, err)
	assert.Contains(t, string(sentBody), `"query":"example"`)

	require.NotNil(t, resp)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "Example Doc", resp.Results[0].Title)
	assert.Equal(t, "req-123", resp.RequestID)
}
