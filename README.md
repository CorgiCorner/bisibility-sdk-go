# Bisibility Go SDK

> Part of [bisibility](https://github.com/CorgiCorner/bisibility) - open-source keyword
> rank tracking you can self-host and automate. This repository contains the Go SDK for
> the Bisibility REST API.
>
> [Docs](https://bisibility.com/docs) ·
> [API reference](https://bisibility.com/docs/api/overview) ·
> [Roadmap](https://bisibility.com/roadmap)
>
> **Status:** Published as v0.10.0.

Idiomatic Go client for the Bisibility REST API.

The [canonical SDK behavior contract](https://bisibility.com/docs/sdks/behavior)
defines the shared authentication, timeout, retry, cancellation, error, header,
and cursor semantics implemented by this client.

## Install

```sh
go get bisibility.com/sdk-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	bisibility "bisibility.com/sdk-go"
)

func main() {
	ctx := context.Background()

	client, err := bisibility.NewClient(
		bisibility.WithAPIKey(os.Getenv("BISIBILITY_API_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}

	projects, err := client.ListProjects(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(projects.Data) == 0 {
		return
	}

	created, err := client.CreateKeywords(ctx, projects.Data[0].ID, bisibility.CreateKeywordsInput{
		Keywords: []bisibility.CreateKeywordInput{
			{
				Keyword:   "rank tracker api",
				TargetURL: ptr("https://example.com/rank-tracker"),
				Tags:      []string{"api"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(created.Results) == 0 {
		return
	}

	check, err := client.RunRankCheckAndWait(ctx, created.Results[0].Keyword.ID, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("position=%v url=%v\n", check.Position, check.RankingURL)
}

func ptr(value string) *string {
	return &value
}
```

## Configuration

```go
client, err := bisibility.NewClient(
	bisibility.WithAPIKey("bsb_key_live_..."),
	bisibility.WithBaseURL("https://bisibility.com/api/v1"),
	bisibility.WithMaxRetries(2),
)
```

`WithBaseURL` should point at the API v1 root. Protected methods send
`Authorization: Bearer <apiKey>`. Write methods accept
`bisibility.WithIdempotencyKey("...")`, which maps to the server
`Idempotency-Key` header.

The client accepts project API keys (`bsb_key_live_...` or `bsb_key_test_...`) and personal access
tokens (`bsb_pat_live_...`). Retired `bsk_` and `bsp_` credentials are rejected locally. For a PAT
with multiple project memberships, pass a project ID
returned by `ListProjects` to `bisibility.WithProjectID(projectID)`. The client
sends it as `X-Bisibility-Project` on project-implicit routes. PAT methods
include `GetMe`, `CreateProject`, token self-management, project API-key minting,
and webhook CRUD.

### Public identifiers

All typed resource IDs accepted by client methods are strict public ID v3 values:
`prefix_[a-z][a-z0-9]{23}`. The SDK rejects raw database IDs, legacy IDs, and
mixed-case values before it sends an HTTP request. Use `ValidatePublicID` or
`ValidatePublicIDPrefix` when validating values before calling the client.

The registered namespaces are `al`, `alr`, `audit`, `check`, `cmp`, `conn`, `dwh`, `ferry`,
`imp`, `inv`, `key`, `kw`, `mbr`, `ntf`, `pat`, `prj`, `sid`, `sig`, `svkw`, `tag`, `usr`,
`viw`, and `we`. Provider IDs and `location_key` values are not public resource IDs.
Migration-token secrets are credentials, while `ferry_` identifies the migration-token resource.

The examples below reuse `projectID` and `keywordID` values returned by the API,
as shown in the quickstart, instead of embedding synthetic resource IDs.

### Defaults

- Every request sends `X-Bisibility-Client: bisibility-sdk-go/<version>` and the same value as
  `User-Agent` (`bisibility.Version`). Override the user agent with
  `bisibility.WithDefaultHeader("User-Agent", "...")` or per request with
  `bisibility.WithRequestHeader`.
- The default HTTP client uses a 30 second timeout. Supply your own client
  with `bisibility.WithHTTPClient(&http.Client{...})` to change the timeout,
  transport, or proxy behavior.
- Idempotent requests retry network errors and HTTP 429/503 responses twice by default. GET, HEAD,
  PUT, and DELETE are idempotent, as is any request carrying `WithIdempotencyKey`. Backoff starts
  at 500ms and honors `Retry-After` up to 60 seconds. `WithMaxRetries(0)` disables retries, and
  context cancellation interrupts retry sleeps.

### Queued rank checks

How a requested check executes belongs to the deployment, not to the call.
Where a background worker owns execution the API answers `202 Accepted` with
the queued run, and where checks run inline it answers `201 Created` with the
finished check. `RunRankCheck` returns both cases:

```go
started, err := client.RunRankCheck(ctx, keywordID, nil)
if err == nil && started.IsQueued() {
	fmt.Printf("queued as run %s\n", started.Queued.ID)
}
```

Every rank check carries the `RunID` of the run that produced it, which is how
a queued run is followed to its result. `RunRankCheckAndWait` does that polling
and returns the finished check, or a `*bisibility.TimeoutError` at the deadline:

```go
check, err := client.RunRankCheckAndWait(ctx, keywordID, nil, &bisibility.WaitForRankCheckOptions{
	Timeout: 2 * time.Minute,
})
```

`RunRankCheckInput.Async` is retained for compatibility and no longer changes
what the server does.

### Public cost estimates

`GetProviderRates` and `GetCostEstimate` are public like the discovery
methods and send no `Authorization` header:

```go
rates, err := client.GetProviderRates(ctx)

estimate, err := client.GetCostEstimate(ctx, bisibility.CostEstimateOptions{
	Keywords:  248,
	Frequency: bisibility.EstimateFrequencyDaily,
	Provider:  bisibility.ProviderIDDataForSEO,
	Option:    "standard",
})
fmt.Printf("monthly cost: $%.2f\n", estimate.Data.MonthlyCostUSD)
```

Flat rate cards (`PricingModel` `flat`) carry `Options`; plan rate cards
(`plan`) carry `Plans`.

### Signals

`CreateSignal` ingests deploy, CMS, or API events for the API key's project,
and `ListSignals` pages through them newest first:

```go
signal, err := client.CreateSignal(ctx, bisibility.CreateSignalInput{
	Source:  bisibility.SignalSourceDeploy,
	Type:    "deploy.completed",
	URL:     "https://example.com/releases/42",
	Payload: bisibility.JSONValue{"version": "1.2.3"},
})

signals, err := client.ListSignals(ctx, projectID, &bisibility.ListSignalsOptions{
	Source: bisibility.SignalSourceDeploy,
	From:   time.Now().AddDate(0, 0, -7),
})
```

`CreateSignal` only accepts the `deploy`, `cms`, and `api` sources; the other
`SignalSource` values are emitted by Bisibility and only appear in list
responses and list filters. Payloads must serialize to 8KB or less.

### Keyword research and metrics

`ResearchKeywords` runs one paid, cached DataForSEO Labs lookup for a single seed. Choose a
research mode and a result limit of 100, 300, or 500 before the request. There is no offset
pagination. `IncludeClickstream` requests clickstream-refined metrics and increases provider cost.
Use `EstimateOnly` for a free, cache-aware dry run and `MaxCostCents` for a best-effort request
guard. Partial auto-mode responses identify each source as `ok`, `failed`, or `skipped` with an
optional machine-readable reason. This method requires an API key with write scope.

```go
research, err := client.ResearchKeywords(ctx, projectID, bisibility.ResearchKeywordsOptions{
	Seed:         "rank tracker",
	Mode:         bisibility.KeywordResearchModeAuto,
	ResultLimit:  100,
	MaxCostCents: 5,
})
```

`GetKeywordMetrics` hydrates nullable volume, CPC, competition, difficulty, intent, and monthly
trend data for up to 700 keywords. The API caches each keyword independently and fetches only cache
misses unless `Fresh` is set. Split larger inputs into requests of at most 700 keywords. Set
`EstimateOnly` to inspect `CachedCount`, `FetchedCountEstimate`, and `EstimatedCostCents` without
spending. `MaxCostCents` rejects a paid lookup whose estimate is too high. This method requires an
API key with write scope.

```go
metrics, err := client.GetKeywordMetrics(ctx, projectID, bisibility.GetKeywordMetricsInput{
	Keywords: []string{"rank tracker", "seo api"},
})
```

### Saved keywords

`CreateSavedKeywords` persists researched keywords on a project so they survive
the research cache. Only `Keyword` is required; the API substitutes the project
default market when `Location` is empty and reports keywords already saved or
tracked as skipped instead of failing the request. Saved keywords carry `svkw_` public
IDs and nullable provider metrics:

```go
saved, err := client.CreateSavedKeywords(ctx, projectID, bisibility.CreateSavedKeywordsInput{
	Keywords: []bisibility.SavedKeywordItem{
		bisibility.SavedKeywordText("rank tracker"),
		bisibility.SavedKeywordInput{Keyword: "seo api", SourceSeed: "rank tracker"},
	},
})
fmt.Printf("saved %d, duplicates %d\n", saved.SavedCount, saved.DuplicateCount)

keywords, err := client.ListSavedKeywords(ctx, projectID, nil)

removed, err := client.DeleteProjectSavedKeyword(ctx, projectID, savedKeywordID)
```

### Domain overview

`AnalyzeDomainOverview` returns either a cache-aware estimate or a report with core metrics plus
ranked-keyword and relevant-page module outcomes. Estimate first, then pass an explicit
`MaxCostCents` pointer before any request that may spend provider budget. A zero cap makes the
operation cache-only. `Fresh` bypasses caches but never removes the explicit cap requirement.

```go
estimateOnly := true
estimate, err := client.AnalyzeDomainOverview(ctx, projectID, bisibility.AnalyzeDomainOverviewOptions{
	Target:       "example.com",
	LocationCode: 2840,
	LanguageCode: "en",
	EstimateOnly: &estimateOnly,
})
if err != nil {
	log.Fatal(err)
}
if estimate.Data.Estimate == nil {
	log.Fatal("expected an estimate")
}

maxCost := 10
report, err := client.AnalyzeDomainOverview(ctx, projectID, bisibility.AnalyzeDomainOverviewOptions{
	Target:       "example.com",
	LocationCode: 2840,
	LanguageCode: "en",
	MaxCostCents: &maxCost,
})
if err != nil {
	log.Fatal(err)
}
if report.Data.Report != nil {
	fmt.Printf("state=%s charged=%.4f cents\n", report.Data.Report.State, report.Data.Report.CostCents)
}
```

`LoadDomainOverviewHistory`, `LoadDomainOverviewKeywords`, and `LoadDomainOverviewPages` load
separately priced modules for an unexpired overview snapshot. Every input includes a required
`MaxCostCents` field, with zero used for cache-only attempts. The API returns failed top-level
operations as RFC problem responses; partial analysis reports keep typed success or failure
outcomes on their nested keyword and page modules. Decode `APIError.Problem.Errors` into
`DomainOverviewProblemErrors` when callers need the failure reason, charged cost, or reset time.

## Methods

- Discovery: `GetHealth`, `GetLiveness`, `GetReadiness`, `GetOpenAPI`, `GetCapabilities`,
  `GetLLMSText`
- Public cost: `GetProviderRates`, `GetCostEstimate`
- Projects: `ListProjects`, `Projects`, `GetProject`, `UpdateProject`,
  `DeleteProject`, `UpdateProjectDefaults`
- API keys: `ListAPIKeys`, `CreateAPIKey`, `RevokeAPIKey`
- Keywords: `ListKeywords`, `KeywordsList`, `CreateKeywords`, `KeywordsCreate`,
  `AddKeywords`, `GetKeyword`, `UpdateKeyword`, `SetKeywordTargetURL`,
  `DeleteKeyword`, `BulkUpdateKeywords`, `ResearchKeywords`, `GetKeywordMetrics`
- Rank checks: `ListRankChecks`, `RankHistory`, `ExportRankHistory`,
  `IterateRankHistory`, `RunRankCheck`, `RunRankCheckAndWait`, `RunCheck`,
  `GetRankCheckResult`
- Alert rules: `ListAlertRules`, `CreateAlertRule`, `UpdateAlertRule`,
  `DeleteAlertRule`, `ListTriggeredAlerts`, `MuteTriggeredAlert`,
  `MarkProjectAlertsRead`
- Team: `ListTeamMembers`, `ListTeamInvites`, `CreateTeamInvite`,
  `RevokeTeamInvite`, `RevokeProjectTeamInvite`
- Providers: `ListProviders`, `ConnectProvider`, `TestProviderConnection`,
  `UpdateProviderSettings`, `SetProviderEnabled`, `SetProviderPriority`,
  `SetPrimaryProvider`, `DisconnectProvider`
- Saved keywords: `ListSavedKeywords`, `IterateSavedKeywords`,
  `CreateSavedKeywords`, `DeleteProjectSavedKeyword`
- Domain overview: `AnalyzeDomainOverview`, `LoadDomainOverviewHistory`,
  `LoadDomainOverviewKeywords`, `LoadDomainOverviewPages`
- Saved views: `ListSavedViews`, `CreateSavedView`, `DeleteSavedView`,
  `DeleteProjectSavedView`
- Competitors: `ListCompetitors`, `AddCompetitor`, `RemoveCompetitor`,
  `RemoveProjectCompetitor`
- Notification preferences: `GetNotificationPreferences`,
  `UpdateNotificationPreferences`
- Migration tokens: `ListMigrationTokens`, `MintMigrationToken`,
  `RevokeMigrationToken`, `RevokeProjectMigrationToken`
- Cloud import: `GetCloudImportCompatibility`, `ImportCloudExport`,
  `CreateCloudImportSession`, `UploadCloudImportChunk`,
  `UploadCloudImportChunkRaw`, `FinalizeCloudImportSession`
- Signals: `CreateSignal`, `ListSignals`
- Sitemap monitors: `ListSitemapMonitors`, `UpdateSitemapMonitor`

List methods return `ListResponse[T]` with `Meta.NextCursor`. Resource methods
return typed resources. Cursor values are opaque: pass v3 API cursors back unchanged.

`ExportRankHistory` returns a cursor-paginated JSON page by default. Set
`Format: bisibility.RankHistoryExportFormatCSV` to receive the complete CSV document in the
response's `CSV` field.

Go 1.22 consumers can traverse every cursor list with the corresponding `Iterate*` method and a
`Pager`. Filters remain unchanged between pages:

```go
pager := client.IterateKeywords(ctx, projectID, &bisibility.ListKeywordsOptions{Tag: "api"})
for pager.Next() {
	keyword := pager.Item()
	fmt.Println(keyword.Text)
}
if err := pager.Err(); err != nil {
	log.Fatal(err)
}
```

Pagers cover keywords, rank checks, signals, API keys including project API keys, webhooks, alert
rules, triggered alerts, team members, team invites, providers, saved views, competitors, and
migration tokens.

`ListKeywords` filters include `Intent` and `Topic` (case-insensitive exact
matches, sent as `filter[intent]` and `filter[topic]`). Provider methods
accept the connectable provider ids `dataforseo`, `serpapi`, `gsc`, `ga4`,
and `plausible` (`bisibility.ProviderIDPlausible`); self-hosted providers
such as Plausible take their instance URL via
`ProviderCredentialsInput.Endpoint`.

Some list endpoints expose typed metadata beyond pagination. `ListCompetitors`
returns `ListCompetitorsResponse` with markets and suggestions, and
`ListMigrationTokens` returns `ListMigrationTokensResponse` with import job
status.

Cloud-import writes authenticate with a migration token minted by
`MintMigrationToken`, passed as the first argument rather than through the
client API key. `GetCloudImportCompatibility` is an unauthenticated preflight.
The SDK supports only protocol version 5 and writes that discriminator itself.
`CloudImportPackage` requires `project_id` plus non-nil `keywords`,
`alert_rules`, `competitors`, `notification_preferences`, and `saved_views`
collections. `CreateCloudImportSession` requires a strict `source_project_id`.
Although the route still says `sessions`, the returned ID and every chunk or
finalize path use the strict `imp_` public-ID namespace.

`UploadCloudImportChunk` accepts either `CloudImportKeywordsChunk` or
`CloudImportSectionsChunk`, so the `kind` discriminator is fixed by the Go
type. Alert-rule targets similarly use `CloudImportKeywordAlertTarget` or
`CloudImportTagAlertTarget`. The SDK rejects v4 payloads, raw IDs, camel-case
aliases for snake-case fields, and incomplete required shapes before sending a
request.
`UploadCloudImportChunkRaw` streams a pre-serialized JSON body from an
`io.Reader` and can set `Content-Encoding: gzip` for a compressed chunk.

```go
session, err := client.CreateCloudImportSession(ctx, migrationToken, bisibility.CloudImportSessionCreate{
	ChunkCount:      1,
	SourceProjectID: projectID,
})
if err != nil {
	log.Fatal(err)
}
_, err = client.UploadCloudImportChunk(ctx, migrationToken, session.SessionID, 0, bisibility.CloudImportKeywordsChunk{
	Checksum: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
	Keywords: []bisibility.CloudImportKeyword{{
		ID:       keywordID,
		Keyword:  "rank tracker",
		Device:   bisibility.DeviceDesktop,
		Location: "United States",
	}},
})
if err != nil {
	log.Fatal(err)
}
```

## Errors

All SDK-defined errors implement `bisibility.BisibilityError`. Non-2xx API responses return
`*bisibility.APIError`; the original RFC problem details body is available on `err.Problem`.
`IsRateLimit`, `IsNotFound`, and `RetryAfter` provide common status helpers. Sensitive response
headers are removed before an API error is exposed.

```go
keyword, err := client.GetKeyword(ctx, "kw_z9y8x7w6v5u4t3s2r1q0p9n8")
if err != nil {
	var apiErr *bisibility.APIError
	if errors.As(err, &apiErr) {
		log.Printf("status=%d detail=%s", apiErr.StatusCode, apiErr.Problem.Detail)
	}
	log.Fatal(err)
}
_ = keyword
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
