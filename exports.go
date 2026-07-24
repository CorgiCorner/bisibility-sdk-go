package bisibility

import (
	"context"
	"net/http"
)

// ExportRankHistory returns a cursor-paginated JSON page or the complete CSV export.
// Set Format to RankHistoryExportFormatCSV for CSV. CSV responses are not paginated.
func (c *Client) ExportRankHistory(ctx context.Context, projectID string, input *ExportRankHistoryOptions, options ...RequestOption) (*RankHistoryExportResponse, error) {
	config := newRequestConfig(options...)
	format := RankHistoryExportFormatJSON
	if input != nil {
		addQuery(config.query, "cursor", input.Cursor)
		addQuery(config.query, "format", string(input.Format))
		addQuery(config.query, "granularity", string(input.Granularity))
		for _, keywordID := range input.KeywordIDs {
			config.query.Add("keyword_id", keywordID)
		}
		addIntQuery(config.query, "limit", input.Limit)
		addQuery(config.query, "range", string(input.Range))
		if input.Format != "" {
			format = input.Format
		}
	}

	path := projectResourcePath(projectID, "exports/rank-history")
	if format == RankHistoryExportFormatCSV {
		csv, err := requestText(c, ctx, http.MethodGet, path, config)
		if err != nil {
			return nil, err
		}
		return &RankHistoryExportResponse{CSV: csv, Format: format}, nil
	}

	response, err := requestJSON[RankHistoryExportResponse](c, ctx, http.MethodGet, path, config)
	if err != nil || response == nil {
		return response, err
	}
	response.Format = format
	return response, nil
}
