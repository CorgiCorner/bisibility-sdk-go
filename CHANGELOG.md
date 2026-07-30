# Changelog

## 0.5.1 - 2026-07-30

- Improved package metadata to describe the SDK's SEO rank-tracking, keyword, and ranking-history
  capabilities.

## 0.5.0 - 2026-07-29

- Breaking for API consumers: resource identifiers now use the public ID v3 prefix registry, and
  identifiers with retired v2 prefixes are rejected before a request is sent.
- Breaking for authentication: use the namespaced `bsb_key_live_`, `bsb_key_test_`, and
  `bsb_pat_live_` credential prefixes; retired `bsk_` and `bsp_` credentials are rejected locally.
- Breaking for cloud-import consumers: require schema v5 packages and sessions, and document v3
  cursors as opaque continuation values that must be passed back unchanged.

## 0.4.1 - 2026-07-28

- Added `ranking_url` to `KeywordMatch` responses, identifying the URL that ranked at `latest_position` in the last completed check or returning null when no completed check exists.

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
