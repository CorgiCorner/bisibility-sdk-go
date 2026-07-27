# Changelog

## Unreleased

## 0.4.0 - 2026-07-27

- Added Backlinks analysis and load-more row operations.
- Added `GetProjectOverview` to read project rank-performance summaries with device, date-range, and tag filters.
- Added `MatchProjectKeywords` to match normalized request texts across stored project keywords and markets, preserving normalized `matched_text` separately from stored `text` and reporting partial results in `meta.truncated_texts`.

## 0.3.0 - 2026-07-26

- Changed the Go module and import path to `bisibility.com/sdk-go`; use `go get bisibility.com/sdk-go` and import `bisibility.com/sdk-go`.

## 0.2.0 - 2026-07-26

- Added `GetProjectDefaults` to read the effective default market and schedule settings for a project.
- Matched project defaults to the current API contract by adding SERP and source fields and removing obsolete auto-schedule fields.

## 0.1.0 - 2026-07-24

- Initial release.
