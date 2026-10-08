# Bisibility Go SDK

> Part of [bisibility](https://github.com/CorgiCorner/bisibility) - open-source SEO platform
> you can self-host and automate. This repository contains the Go SDK for
> the Bisibility REST API.
>
> [Docs](https://bisibility.com/docs) ·
> [API reference](https://bisibility.com/docs/api/overview) ·
> [Roadmap](https://bisibility.com/roadmap)
>
> Current versions are listed in the [module release tags](https://github.com/CorgiCorner/bisibility-sdk-go/tags).

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

All typed resource IDs accepted by client methods are strict typed public IDs:
`prefix_[a-z][a-z0-9]{23}`. The SDK rejects raw database IDs, legacy IDs, and
mixed-case values before it sends an HTTP request. Use `ValidatePublicID` or
`ValidatePublicIDPrefix` when validating values before calling the client.

The SDK namespaces are `agr`, `al`, `alr`, `audit`, `check`, `cmp`, `conn`, `dwh`, `ferry`,
`imp`, `inv`, `key`, `kw`, `mbr`, `ntf`, `pat`, `prj`, `rcr`, `sid`, `sig`, `svkw`, `tag`, `usr`,
`viw`, and `we`. Provider IDs and `location_key` values are not public resource IDs.
Migration-token secrets are credentials, while `ferry_` identifies the migration-token resource.

The examples below reuse `projectID` and `keywordID` values returned by the API,
as shown in the quickstart, instead of embedding synthetic resource IDs.

### Defaults

- Every request sends `X-Bisibility-Client: bisibility-sdk-go/<version>` and the same value as
  `User-Agent` (`bisibility.Version`). Override the user agent with
  `bisibility.WithDefaultHeader("User-Agent", "...")` or per request with
  `bisibility.WithRequestHeader`.
- Every request also declares its origin with `X-Bisibility-Source: sdk` for usage reporting;
  `WithDefaultHeader("X-Bisibility-Source", "cli")` overrides it.
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

`RunRankCheckInput.MaxCostCents` sends a ceiling without
a `ProviderID`. `MaxCostCentsOverride` takes an optional `*int`, takes precedence,
and can send an explicit zero ceiling. Existing provider-only requests omit the
ceiling.

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
`MaxCostCents` is a best-effort request guard. Partial auto-mode responses identify each source as
`ok`, `failed`, or `skipped` with an optional machine-readable reason. This method requires an API
key with write scope.

`KeywordResearchResponse` is a discriminated union: exactly one of `Estimate` and `Result` is set.
`EstimateOnly` returns a free, cache-aware `KeywordResearchEstimate` that carries per-source costs
only - never rows, a fetch time, or source statuses - so an estimate can never be mistaken for an
empty result. Every other request returns a `KeywordResearchResult` with the researched rows.

```go
estimate, err := client.ResearchKeywords(ctx, projectID, bisibility.ResearchKeywordsOptions{
	Seed:         "rank tracker",
	Mode:         bisibility.KeywordResearchModeAuto,
	ResultLimit:  100,
	EstimateOnly: true,
})
if err != nil {
	log.Fatal(err)
}
if estimate.Estimate == nil {
	log.Fatal("expected an estimate")
}
fmt.Printf("estimated %.2f cents across %d sources\n", estimate.Estimate.CostCents, len(estimate.Estimate.Sources))

research, err := client.ResearchKeywords(ctx, projectID, bisibility.ResearchKeywordsOptions{
	Seed:         "rank tracker",
	Mode:         bisibility.KeywordResearchModeAuto,
	ResultLimit:  100,
	MaxCostCents: 5,
})
if err != nil {
	log.Fatal(err)
}
if research.Result != nil {
	fmt.Printf("%d keywords charged %.2f cents\n", research.Result.TotalCount, research.Result.CostCents)
}
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

Use `AnalyzeDomainOverview` with `EstimateOnly: true` for the current feature estimate.

### Backlinks

`AnalyzeBacklinks` returns either a free, cache-aware estimate or a paid backlink snapshot.
`BacklinksResponse.Data` is a discriminated union: exactly one of `Estimate` and `Snapshot` is set.
`EstimateOnly` returns a cost-only `BacklinksEstimate` (`Estimate`, `EstimatedCostCents`,
`CostCents`, `Cached`, `CachedUntil`, `Provider`, and the normalized target) that never carries
`Summary`, `History`, `Rows`, or `FetchedAt`, so an estimate can never be mistaken for an empty
backlink profile. `CachedUntil` is nil when no unexpired snapshot exists. `MaxCostCents` is a
best-effort guard applied to the pre-estimate. Both methods require write scope because a cache
miss can spend provider budget.

```go
estimate, err := client.AnalyzeBacklinks(ctx, projectID, bisibility.AnalyzeBacklinksOptions{
	Target:       "example.com",
	TargetScope:  bisibility.BacklinkTargetScopeSite,
	EstimateOnly: true,
})
if err != nil {
	log.Fatal(err)
}
if estimate.Data.Estimate == nil {
	log.Fatal("expected an estimate")
}
fmt.Printf("estimated %.2f cents\n", estimate.Data.Estimate.EstimatedCostCents)

analysis, err := client.AnalyzeBacklinks(ctx, projectID, bisibility.AnalyzeBacklinksOptions{
	Target:       "example.com",
	TargetScope:  bisibility.BacklinkTargetScopeSite,
	ResultLimit:  100,
	MaxCostCents: 10,
})
if err != nil {
	log.Fatal(err)
}
if analysis.Data.Snapshot != nil {
	snapshot := analysis.Data.Snapshot
	fmt.Printf("%d of %d rows charged %.2f cents\n",
		snapshot.FetchedRowCount, snapshot.TotalRowsAvailable, snapshot.CostCents)
}
```

Use `AnalyzeBacklinks` with `EstimateOnly: true` for the current feature estimate.
`LoadMoreBacklinkRows`
appends paid rows to an unexpired snapshot and always returns a `BacklinksSnapshotResponse`,
never an estimate.

```go
more, err := client.LoadMoreBacklinkRows(ctx, projectID, bisibility.LoadMoreBacklinkRowsOptions{
	Target:      "example.com",
	TargetScope: bisibility.BacklinkTargetScopeSite,
	Limit:       100,
})
```

### Providers

`ConnectProvider` stores credentials and places the connection in the project fallback chain.
`Priority` is optional and runs from 0 to 1000: `0` promotes the provider and renumbers the chain,
any other value reorders it, and omitting it keeps a reconnected provider's place and appends a new
connection. `Primary` is SDK sugar for `Priority` 0 and wins when both are set. The priority now
travels with the connect request, so a single call both connects and orders the provider.

Credential fields are provider specific. For Plausible, `Credentials.Login` is the site domain
configured in Plausible (its `site_id`, such as `example.com`) and defaults to the project domain
when omitted, and `Credentials.APIKey` is the Stats API token.

```go
priority := 0
connection, err := client.ConnectProvider(ctx, projectID, bisibility.ProviderIDPlausible,
	bisibility.ConnectProviderInput{
		Credentials: &bisibility.ProviderCredentialsInput{APIKey: statsAPIToken, Login: "example.com"},
		Priority:    &priority,
	})
```

`TestProviderConnection` reports a successful probe with `Message` `"Connected."` for SERP
providers and `"Connected · <detail>."` for analytics providers, where the detail names the
verified property or site.

### Project defaults

`UpdateProjectDefaults` replaces the schedule fields (`Frequency`, `CronExpression`,
`JitterMinutes`, `Timezone`) as a whole, so omitted schedule fields fall back to server defaults.
`SerpDepth` and `SerpStopOnMatch` are independent of the schedule: omitting either keeps its
stored value. `SerpDepth` accepts `10`, `20`, `50`, or `100` (see `bisibility.SerpDepths`); any
other value is rejected locally with a `ConfigurationError`.

```go
serpDepth := 50
defaults, err := client.UpdateProjectDefaults(ctx, projectID, bisibility.ProjectDefaultsPatch{
	Frequency:   bisibility.RankCheckFrequencyDaily,
	LocationKey: "US/Texas/Austin",
	SerpDepth:   &serpDepth,
})
```

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
- Backlinks: `AnalyzeBacklinks`, `LoadMoreBacklinkRows`
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
- Sitemap monitors: `ListSitemapMonitors`, `UpdateSitemapMonitor` (a project has one
  sitemap monitor and its monitor ID is the project ID)

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

Typed methods cover project context, agent reports, AI visibility,
prompt comparison, and site audits. Agent reports use strict `agr_` IDs;
`IterateAgentReports` preserves report-kind filters across v3 cursor pages. AI methods
require `MaxCostCents` and distinguish cost estimates from saved results.

Cloud-import writes authenticate with a migration token minted by
`MintMigrationToken`, passed as the first argument rather than through the
client API key. `GetCloudImportCompatibility` is an unauthenticated preflight.
The client supports package versions 6 and 7; `Version` defaults to 7 when zero.
Version 7 requires `LocationKey` for every package keyword, while version 6 uses
legacy market names without that field. Both versions require history rows to carry
`NormalizationVersion`, `Provider`, `RequestedDepth`, `Position`, `PreviousPosition`,
and `RankingURL` alongside `CheckedAt`. Nullable values remain explicit JSON nulls.
Version 5 is accepted only without ranking history. Sessions require version 6 or 7.
Compatibility discovery returns the integer versions advertised by the server, even
if this SDK cannot send them yet.
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
        LocationKey: "US",
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

## Saved reports and provider budgets

Saved report reads never invoke a provider. Use `fresh_until` to determine freshness;
`state` is `fresh` or `stale`, while a domain report keeps its data outcome in `data_state`.
Own-key and credit budgets are independent. Omit a field to keep it and explicitly clear
a surface to remove its budget. Credit budgets always use cents.

```go
saved, err := client.ListStoredResearchReports(ctx, projectID)
report, err := client.GetStoredResearchReport(ctx, projectID, bisibility.StoredReportKeywordResearch, bisibility.StoredResearchReportOptions{Seed: "example", ResultLimit: 100})
budgets, err := client.ListProviderBudgets(ctx, projectID)
updated, err := client.UpdateProviderBudgets(ctx, projectID, bisibility.ProviderIDDataForSEO, bisibility.ProviderBudgetsUpdate{
    Own: &bisibility.ProviderBudgetPatch{App: bisibility.ClearProviderBudget()},
    Credits: &bisibility.ProviderBudgetPatch{Programmatic: bisibility.ProviderBudgetValue(bisibility.ProviderMonthlyBudget{AmountPerMonth: 500, Unit: "cents"})},
})
```

`report.Data` populates exactly one of `Backlinks`, `DomainOverview`, or `KeywordResearch`.
