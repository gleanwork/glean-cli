package search

import (
	"context"
	"fmt"
	"time"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
)

// datasourceField is the facet filter field that both API surfaces route to
// a dedicated datasource filter instead of the generic filter mechanism.
const datasourceField = "datasource"

// BuildPlatformSearchRequest converts Options into the platform API's
// components.PlatformSearchRequest. The platform surface is deliberately
// narrower than the classic SearchRequest: flags without a platform
// equivalent are simply not mapped (cmd/search reports those to the user via
// flags.Changed, which knows what was explicitly set).
func BuildPlatformSearchRequest(opts *Options) components.PlatformSearchRequest {
	req := components.PlatformSearchRequest{
		Query: opts.Query,
	}

	if opts.PageSize > 0 {
		pageSize := int64(opts.PageSize)
		req.PageSize = &pageSize
	}
	if opts.Cursor != "" {
		cursor := opts.Cursor
		req.Cursor = &cursor
	}

	if opts.RequestOptions != nil {
		for _, ff := range opts.RequestOptions.FacetFilters {
			// Mirror the classic builder: datasource filtering uses the
			// dedicated field, not the generic filter mechanism.
			if ff.FieldName == datasourceField {
				for _, v := range ff.Values {
					req.Datasources = append(req.Datasources, v.Value)
				}
				continue
			}
			filter := components.PlatformFilter{Field: ff.FieldName}
			for _, v := range ff.Values {
				filter.Values = append(filter.Values, v.Value)
			}
			req.Filters = append(req.Filters, filter)
		}
	}

	return req
}

// RunPlatformSearch executes a search against the platform API
// (POST /api/search) and returns the raw platform response. The classic
// request carries timeoutMillis in the body; the platform API does not, so
// --timeout is honored via a context deadline instead.
func RunPlatformSearch(ctx context.Context, opts *Options, sdk *glean.Glean) (*components.PlatformSearchResponse, error) {
	if opts.TimeoutMillis > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.TimeoutMillis)*time.Millisecond)
		defer cancel()
	}

	result, err := sdk.Search.Query(ctx, BuildPlatformSearchRequest(opts))
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	return result.PlatformSearchResponse, nil
}
