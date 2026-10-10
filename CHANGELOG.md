# Changelog

## Unreleased

## 0.16.0 - 2026-10-10

- Add typed project AI tracking operations for topics, immutable prompt revisions, schedules,
  consented runs, samples, history, trends, evidence exports, and accepted suggestions.
- Add free AI research catalog reads and prompt comparison options for current models, web search,
  locale hints, and explicitly acknowledged provider actual-cost mode.
- Add reviewed model suggestion preview and generation with a frozen preview, explicit paid consent,
  and a stable idempotency UUID. Accepted edits retain trusted generation or provider dataset references.
- Preserve unknown costs and observation times as null, and expose optional evidence export
  completeness and resume metadata.
- Retained rank coverage, latest-check evidence and history availability during typed response decoding and JSON encoding.
- Stopped automatic retries of potentially paid requests, separated omitted from null fields, made schedule updates partial and kept trends baseline, category and strata.

## 0.13.0 - 2026-10-07

- Add typed project context, agent reports, AI visibility, prompt comparison, and site-audit
  methods, including strict report IDs and cursor iteration.
- Keep queued rank-check polling active until the matching check is completed or failed while
  preserving context cancellation.
- Preserve rank-check cost caps without a provider override and add `MaxCostCentsOverride` for
  explicit zero ceilings, taking precedence over the existing integer field.
- Support cloud-import package and session versions 6 and 7, including current ranking-history
  metadata, canonical location keys, and alert severity. Preserve required null history values,
  accept server-advertised integer compatibility versions, and reject ambiguous legacy history.
  Package and session `Version` fields default to 7 when zero; legacy version 5 packages may
  contain metadata only.
- Correct product, public-ID, migration, and feature-pricing guidance.

## 0.12.0 - 2026-09-27

- Added saved research report reads and separate own-key and credit provider budget methods.

- **Breaking for typed consumers:** `AnalyzeBacklinks` returns a `BacklinksResponse` whose
  `Data` is a discriminated union with `Estimate` and `Snapshot` pointers. An `EstimateOnly`
  request decodes into the new cost-only `BacklinksEstimate`, and `BacklinksSnapshot` no longer
  carries `Estimate` or `EstimatedCostCents`. `LoadMoreBacklinkRows` still returns a snapshot.
- **Breaking for typed consumers:** `ResearchKeywords` returns a `KeywordResearchResponse` that is
  now a discriminated union with `Estimate` and `Result` pointers. An `EstimateOnly` request
  decodes into the new cost-only `KeywordResearchEstimate`, and the previous response shape is the
  new `KeywordResearchResult` without its `Estimate` field.
- Add `SerpDepth` to `ProjectDefaultsPatch`. Omitting it keeps the stored depth, and `10`, `20`,
  `50`, and `100` are accepted (`SerpDepths`); any other value is rejected locally.
- `ConnectProvider` sends `Priority` with the connect request instead of a follow-up PATCH, so a
  single call connects and orders the provider. `ProviderPrioritySyncError` is deprecated and is
  no longer returned.
- Document Plausible credentials (`Credentials.Login` is the site domain and defaults to the
  project domain; `Credentials.APIKey` is the Stats API token) and the `"Connected."` and
  `"Connected · <detail>."` provider test messages.
- Document the backlinks and provider methods in the README, including that a sitemap monitor ID
  is the project ID.

## 0.11.0 - 2026-09-20

- Add `MaxCostCents` to `RunRankCheckInput`; the server refuses the check with `cost_limit_exceeded`
  when its preflight estimate is higher.
- Send `X-Bisibility-Source: sdk` on every request so the API can report SDK usage separately;
  a `WithDefaultHeader("X-Bisibility-Source", ...)` value wins.

## 0.10.0 - 2026-09-05

- Model the queued rank-check contract: `RunRankCheck` returns a `RunRankCheckResult` carrying
  either the completed check or the run queued with 202, and rank checks carry `RunID`.
- Add `RunRankCheckAndWait`, which follows a queued run to its check and returns a `TimeoutError`
  at the deadline.
- Register the `rcr` rank-check-run public ID prefix.

## 0.9.0 - 2026-08-14

- Added language-qualified market fields to keyword, keyword-match, and location-search types.

## 0.8.0 - 2026-08-13

- Added typed Domain Overview estimate, report, history, ranked-keyword, and relevant-page methods.

## 0.7.0 - 2026-08-10

- Updated `ConnectProvider` for the current provider-priority API and corrected the default User-Agent version.

## 0.6.1 - 2026-08-04

- No SDK API or runtime behavior changes. This maintenance release updates public module
  validation and preserves the v0.6.0 surface.

## 0.6.0 - 2026-08-02

- **Breaking for typed consumers:** `GetHealth`, `GetLiveness`, and `GetReadiness` now return
  status-only responses; degraded health and readiness HTTP 503 responses return normally without
  retries, while other endpoints retain 503 retries.
- Added `ListSavedKeywords`, `IterateSavedKeywords`, `CreateSavedKeywords`, and
  `DeleteProjectSavedKeyword` with `svkw` public IDs.

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
