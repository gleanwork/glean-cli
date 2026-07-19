// Package testutils provides testing utilities for the Glean CLI,
// including mock HTTP transports for injecting test responses into the SDK.
package testutils

import (
	"bytes"
	"io"
	"net/http"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/config"
)

// MockResponse is a canned response for a specific request path, used via
// MockTransport.Routes to give the platform and legacy endpoints different
// behavior (e.g. platform 404 → legacy 200 fallback tests).
type MockResponse struct {
	// StatusCode defaults to 200 when zero
	StatusCode int
	Body       []byte
	// ContentType defaults to the request's Accept header (see Do)
	ContentType string
}

// MockTransport implements http.RoundTripper (the Do method expected by glean.HTTPClient).
// It returns a predefined response body for every request, making it easy to test
// command output without making real network calls.
type MockTransport struct {
	// Body is returned for every request
	Body []byte
	// Err is returned instead of a response when non-nil
	Err error
	// StatusCode defaults to 200 when zero
	StatusCode int
	// ContentType defaults to "application/json" when empty
	ContentType string
	// Routes overrides the response per request URL path when the path is present
	Routes map[string]MockResponse
	// Requests records all requests received for inspection
	Requests []*http.Request
}

func (m *MockTransport) Do(req *http.Request) (*http.Response, error) {
	m.Requests = append(m.Requests, req)
	if m.Err != nil {
		return nil, m.Err
	}
	body := m.Body
	statusCode := m.StatusCode
	contentType := m.ContentType
	if route, ok := m.Routes[req.URL.Path]; ok {
		body = route.Body
		statusCode = route.StatusCode
		contentType = route.ContentType
	}
	if statusCode == 0 {
		statusCode = 200
	}
	// Mirror the Accept header the SDK sends so the SDK can parse its own response.
	// CreateStream sets Accept: text/plain; Create sets Accept: application/json.
	if contentType == "" {
		if accept := req.Header.Get("Accept"); accept != "" {
			contentType = accept
		} else {
			contentType = "application/json"
		}
	}
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

// SetupMockClient injects a MockTransport into the SDK client factory and
// returns the mock plus a cleanup function that restores the original factory.
func SetupMockClient(body []byte, err error) (*MockTransport, func()) {
	mock := &MockTransport{Body: body, Err: err}
	origFunc := client.NewFunc
	client.NewFunc = func(cfg *config.Config) (*glean.Glean, error) {
		return glean.New(
			glean.WithServerURL("https://test-company-be.glean.com"),
			glean.WithSecurity("test-token"),
			glean.WithClient(mock),
		), nil
	}
	return mock, func() {
		client.NewFunc = origFunc
	}
}
