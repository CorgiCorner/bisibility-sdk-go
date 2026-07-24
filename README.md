# Bisibility Go SDK

> Part of [bisibility](https://github.com/CorgiCorner/bisibility) - open-source keyword
> rank tracking you can self-host and automate. This repository contains the Go SDK for
> the Bisibility REST API.
>
> [Docs](https://bisibility.com/docs) ·
> [API reference](https://bisibility.com/docs/api/overview) ·
> [Roadmap](https://bisibility.com/roadmap)
>
> **Status:** In development.

Idiomatic Go client for the Bisibility REST API.

## Install

```sh
go get github.com/bisibility/bisibility-sdk-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	bisibility "github.com/bisibility/bisibility-sdk-go"
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

	check, err := client.RunRankCheck(ctx, created.Results[0].Keyword.ID, nil)
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
	bisibility.WithAPIKey("bsk_live_..."),
	bisibility.WithBaseURL("https://bisibility.com/api/v1"),
	bisibility.WithMaxRetries(2),
)
```

`WithBaseURL` should point at the API v1 root. Protected methods send
`Authorization: Bearer <apiKey>`. Write methods accept
`bisibility.WithIdempotencyKey("...")`, which maps to the server
`Idempotency-Key` header.

The client accepts project API keys (`bsk_live_...`) and personal access tokens
(`bsp_live_...`). For a PAT with multiple project memberships, add
`bisibility.WithProjectID("prj_abc123")` to send `X-Bisibility-Project` on
project-implicit routes. PAT methods include `GetMe`, `CreateProject`, token
self-management, project API-key minting, and webhook CRUD.

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

### Async rank checks

`RunRankCheck` waits for the check by default. Set `Async: true` to enqueue
the check instead; the API responds `202 Accepted` with a rank check in
status `running`, which you can poll with `GetRankCheckResult`:

```go
check, err := client.RunRankCheck(ctx, "kw_1", &bisibility.RunRankCheckInput{Async: true})
```

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

signals, err := client.ListSignals(ctx, "prj_1", &bisibility.ListSignalsOptions{
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
research, err := client.ResearchKeywords(ctx, "prj_1", bisibility.ResearchKeywordsOptions{
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
metrics, err := client.GetKeywordMetrics(ctx, "prj_1", bisibility.GetKeywordMetricsInput{
	Keywords: []string{"rank tracker", "seo api"},
})
```

## Methods

- Discovery: `GetHealth`, `GetOpenAPI`, `GetCapabilities`, `GetLLMSText`
- Public cost: `GetProviderRates`, `GetCostEstimate`
- Projects: `ListProjects`, `Projects`, `GetProject`, `UpdateProject`,
  `DeleteProject`, `UpdateProjectDefaults`
- API keys: `ListAPIKeys`, `CreateAPIKey`, `RevokeAPIKey`
- Keywords: `ListKeywords`, `KeywordsList`, `CreateKeywords`, `KeywordsCreate`,
  `AddKeywords`, `GetKeyword`, `UpdateKeyword`, `SetKeywordTargetURL`,
  `DeleteKeyword`, `BulkUpdateKeywords`, `ResearchKeywords`, `GetKeywordMetrics`
- Rank checks: `ListRankChecks`, `RankHistory`, `ExportRankHistory`,
  `IterateRankHistory`, `RunRankCheck`, `RunCheck`, `GetRankCheckResult`
- Alert rules: `ListAlertRules`, `CreateAlertRule`, `UpdateAlertRule`,
  `DeleteAlertRule`, `ListTriggeredAlerts`, `MuteTriggeredAlert`,
  `MarkProjectAlertsRead`
- Team: `ListTeamMembers`, `ListTeamInvites`, `CreateTeamInvite`,
  `RevokeTeamInvite`, `RevokeProjectTeamInvite`
- Providers: `ListProviders`, `ConnectProvider`, `TestProviderConnection`,
  `UpdateProviderSettings`, `SetProviderEnabled`, `SetProviderPriority`,
  `SetPrimaryProvider`, `DisconnectProvider`
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
return typed resources.

`ExportRankHistory` returns a cursor-paginated JSON page by default. Set
`Format: bisibility.RankHistoryExportFormatCSV` to receive the complete CSV document in the
response's `CSV` field.

Go 1.22 consumers can traverse every cursor list with the corresponding `Iterate*` method and a
`Pager`. Filters remain unchanged between pages:

```go
pager := client.IterateKeywords(ctx, "prj_1", &bisibility.ListKeywordsOptions{Tag: "api"})
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
`UploadCloudImportChunk` sends a typed `CloudImportUploadChunk` as JSON;
`UploadCloudImportChunkRaw` streams a pre-serialized JSON body from an
`io.Reader` and can set `Content-Encoding: gzip` for a compressed chunk.

## Errors

All SDK-defined errors implement `bisibility.BisibilityError`. Non-2xx API responses return
`*bisibility.APIError`; the original RFC problem details body is available on `err.Problem`.
`IsRateLimit`, `IsNotFound`, and `RetryAfter` provide common status helpers. Sensitive response
headers are removed before an API error is exposed.

```go
keyword, err := client.GetKeyword(ctx, "kw_missing")
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
