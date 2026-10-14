package bisibility

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Local retained reads never submit measurements or initialize schedules.
func localMonitorPath(projectID, monitorID string) (string, error) {
	if err := requirePublicID(projectID, "projectID", PublicIDPrefixProject); err != nil {
		return "", err
	}
	if err := requirePublicID(monitorID, "monitorID", PublicIDPrefixLocalMonitor); err != nil {
		return "", err
	}
	return projectResourcePath(projectID, "local-monitors/"+url.PathEscape(monitorID)), nil
}
func localRunPath(projectID, monitorID, runID, targetID string) (string, error) {
	path, err := localMonitorPath(projectID, monitorID)
	if err != nil {
		return "", err
	}
	if err := requirePublicID(runID, "runID", PublicIDPrefixLocalRun); err != nil {
		return "", err
	}
	path += "/runs/" + url.PathEscape(runID)
	if targetID != "" {
		if err := requirePublicID(targetID, "targetID", PublicIDPrefixLocalTarget); err != nil {
			return "", err
		}
		path += "/points/" + url.PathEscape(targetID)
	}
	return path, nil
}

func (c *Client) ListLocalMonitorRuns(ctx context.Context, projectID, monitorID string, input *LocalMonitorHistoryOptions, options ...RequestOption) (*LocalMonitorRunHistoryResponse, error) {
	config := newRequestConfig(options...)
	path, err := localMonitorPath(projectID, monitorID)
	if err != nil {
		return nil, err
	}
	path += "/runs"
	if input != nil {
		cursor := ""
		if input.Cursor != nil {
			cursor = *input.Cursor
		}
		if err := localReadQuery(config.query, input.AsOf, input.From, input.To, cursor, input.Cursor != nil); err != nil {
			return nil, err
		}
	}
	return requestJSON[LocalMonitorRunHistoryResponse](c, ctx, http.MethodGet, path, config)
}

func (c *Client) GetLocalMonitorRunResult(ctx context.Context, projectID, monitorID, runID string, input *LocalMonitorReadOptions, options ...RequestOption) (*LocalMonitorRunResultResponse, error) {
	config := newRequestConfig(options...)
	path, err := localRunPath(projectID, monitorID, runID, "")
	if err != nil {
		return nil, err
	}
	path += "/results"
	if input != nil {
		if err := localReadQuery(config.query, input.AsOf, "", "", "", false); err != nil {
			return nil, err
		}
	}
	return requestJSON[LocalMonitorRunResultResponse](c, ctx, http.MethodGet, path, config)
}

func (c *Client) ListLocalMonitorPointListings(ctx context.Context, projectID, monitorID, runID, targetID string, input *LocalMonitorReadOptions, options ...RequestOption) (*LocalMonitorPointListingsResponse, error) {
	config := newRequestConfig(options...)
	if err := requirePublicID(targetID, "targetID", PublicIDPrefixLocalTarget); err != nil {
		return nil, err
	}
	path, err := localRunPath(projectID, monitorID, runID, targetID)
	if err != nil {
		return nil, err
	}
	path += "/listings"
	if input != nil {
		if err := localReadQuery(config.query, input.AsOf, "", "", "", false); err != nil {
			return nil, err
		}
	}
	return requestJSON[LocalMonitorPointListingsResponse](c, ctx, http.MethodGet, path, config)
}

func (c *Client) ListLocalMonitorPointHistory(ctx context.Context, projectID, monitorID, runID, targetID string, input *LocalMonitorHistoryOptions, options ...RequestOption) (*LocalMonitorPointHistoryResponse, error) {
	config := newRequestConfig(options...)
	if err := requirePublicID(targetID, "targetID", PublicIDPrefixLocalTarget); err != nil {
		return nil, err
	}
	path, err := localRunPath(projectID, monitorID, runID, targetID)
	if err != nil {
		return nil, err
	}
	path += "/history"
	if input != nil {
		cursor := ""
		if input.Cursor != nil {
			cursor = *input.Cursor
		}
		if err := localReadQuery(config.query, input.AsOf, input.From, input.To, cursor, input.Cursor != nil); err != nil {
			return nil, err
		}
	}
	return requestJSON[LocalMonitorPointHistoryResponse](c, ctx, http.MethodGet, path, config)
}

func (c *Client) ExportLocalMonitorRun(ctx context.Context, projectID, monitorID, runID string, input *LocalMonitorExportOptions, options ...RequestOption) (*LocalMonitorExportResponse, error) {
	config := newRequestConfig(options...)
	path, err := localRunPath(projectID, monitorID, runID, "")
	if err != nil {
		return nil, err
	}
	path += "/export"
	if input != nil {
		if err := localReadQuery(config.query, input.AsOf, "", "", "", false); err != nil {
			return nil, err
		}
	}
	format := "json"
	if input != nil {
		if input.Version != "" && input.Version != "1" {
			return nil, fmt.Errorf("unsupported Local export version")
		}
		if input.Format != "" {
			format = input.Format
		}
	}
	if format != "json" && format != "csv" {
		return nil, fmt.Errorf("invalid Local export format")
	}
	config.query.Set("version", "1")
	config.query.Set("format", format)
	if format == "csv" {
		csv, err := requestText(c, ctx, http.MethodGet, path, config)
		if err != nil {
			return nil, err
		}
		return &LocalMonitorExportResponse{Format: format, CSV: csv}, nil
	}
	report, err := requestJSON[LocalMonitorPointReport](c, ctx, http.MethodGet, path, config)
	if err != nil {
		return nil, err
	}
	return &LocalMonitorExportResponse{Format: format, Report: report}, nil
}

// JSON preserves the retained report; CSV is the exact sanitized server text.
type LocalMonitorExportResponse struct {
	Format string
	CSV    string
	Report *LocalMonitorPointReport
}
