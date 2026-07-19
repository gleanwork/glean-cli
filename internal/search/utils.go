package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	glean "github.com/gleanwork/api-client-go"
	"github.com/gleanwork/api-client-go/models/components"
)

// GetTimezoneOffset returns the current timezone offset in minutes
func GetTimezoneOffset() int {
	_, offset := time.Now().Zone()
	return offset / 60
}

// AddFacetFilter adds a facet filter to the search options
func AddFacetFilter(opts *Options, fieldName string, values []string) {
	filter := FacetFilter{
		FieldName: fieldName,
		Values:    make([]FilterValue, len(values)),
	}
	for i, value := range values {
		filter.Values[i] = FilterValue{
			Value:        value,
			RelationType: "EQUALS",
		}
	}
	opts.RequestOptions.FacetFilters = append(opts.RequestOptions.FacetFilters, filter)
}

// BuildSearchRequest converts Options into a components.SearchRequest without executing it.
// Used by --dry-run to show the exact request that would be sent.
func BuildSearchRequest(opts *Options) components.SearchRequest {
	pageSize := int64(opts.PageSize)
	maxSnippet := int64(opts.MaxSnippetSize)
	timeout := int64(opts.TimeoutMillis)

	req := components.SearchRequest{
		Query:             opts.Query,
		PageSize:          &pageSize,
		MaxSnippetSize:    &maxSnippet,
		TimeoutMillis:     &timeout,
		DisableSpellcheck: &opts.DisableSpellcheck,
	}

	if opts.RequestOptions != nil {
		ro := opts.RequestOptions
		tzOffset := int64(ro.TimezoneOffset)
		facetBucketSize := int64(ro.FacetBucketSize)

		sdkOpts := &components.SearchRequestOptions{
			TimezoneOffset:               &tzOffset,
			FacetBucketSize:              facetBucketSize,
			DisableQueryAutocorrect:      &ro.DisableQueryAutocorrect,
			FetchAllDatasourceCounts:     &ro.FetchAllDatasourceCounts,
			QueryOverridesFacetFilters:   &ro.QueryOverridesFacetFilters,
			ReturnLlmContentOverSnippets: &ro.ReturnLlmContentOverSnippets,
		}

		for _, ff := range ro.FacetFilters {
			// The Glean API has a dedicated DatasourcesFilter field that must
			// be used for datasource filtering — the generic FacetFilters
			// mechanism does not filter by datasource correctly.
			if ff.FieldName == datasourceField {
				for _, v := range ff.Values {
					sdkOpts.DatasourcesFilter = append(sdkOpts.DatasourcesFilter, v.Value)
				}
				continue
			}
			name := ff.FieldName
			sdkFF := components.FacetFilter{FieldName: &name}
			for _, v := range ff.Values {
				val := v.Value
				relType := components.RelationType(v.RelationType)
				sdkFF.Values = append(sdkFF.Values, components.FacetFilterValue{
					Value:        &val,
					RelationType: &relType,
				})
			}
			sdkOpts.FacetFilters = append(sdkOpts.FacetFilters, sdkFF)
		}

		for _, hint := range ro.ResponseHints {
			sdkOpts.ResponseHints = append(sdkOpts.ResponseHints, components.ResponseHint(hint))
		}

		req.RequestOptions = sdkOpts
	}

	return req
}

// RunSearchSDK executes a search and returns the raw SDK response for the caller to format.
func RunSearchSDK(ctx context.Context, opts *Options, sdk *glean.Glean) (*components.SearchResponse, error) {
	return performSearch(ctx, sdk, opts)
}

// RunSearch executes a search and writes the results as JSON to w.
func RunSearch(ctx context.Context, opts *Options, sdk *glean.Glean, w io.Writer) error {
	resp, err := performSearch(ctx, sdk, opts)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling search results: %w", err)
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}

// performSearch executes a search request using the Glean SDK.
func performSearch(ctx context.Context, sdk *glean.Glean, opts *Options) (*components.SearchResponse, error) {
	req := BuildSearchRequest(opts)
	result, err := sdk.Client.Search.Query(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	return result.SearchResponse, nil
}
