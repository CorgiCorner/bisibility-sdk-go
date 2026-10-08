package bisibility

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const (
	testMigrationToken  = "mig_secret_do_not_log"
	cloudImportChecksum = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

func cloudImportID(prefix PublicIDPrefix) string {
	return string(prefix) + "_a00000000000000000000000"
}

func validCloudImportPackage() CloudImportPackage {
	return CloudImportPackage{
		Version: 7,
		AlertRules: []CloudImportAlertRule{{
			ID:   cloudImportID(PublicIDPrefixRule),
			Name: "Position drop",
			Targets: []CloudImportAlertRuleTarget{CloudImportKeywordAlertTarget{
				KeywordID: cloudImportID(PublicIDPrefixKeyword),
			}},
		}},
		Competitors: []CloudImportCompetitor{{
			Domain: "example.com",
			ID:     cloudImportID(PublicIDPrefixComp),
		}},
		Keywords: []CloudImportKeyword{{
			Device:      DeviceDesktop,
			ID:          cloudImportID(PublicIDPrefixKeyword),
			Keyword:     "rank tracker",
			Location:    "United States",
			LocationKey: "US",
		}},
		NotificationPreferences: []CloudImportNotificationPreference{{}},
		ProjectID:               cloudImportID(PublicIDPrefixProject),
		SavedViews: []CloudImportSavedView{{
			ID:   cloudImportID(PublicIDPrefixView),
			Name: "Tracked keywords",
		}},
		Scope: CloudImportScopeCurrent,
	}
}

func validCloudImportSessionCreate() CloudImportSessionCreate {
	return CloudImportSessionCreate{
		Version:         7,
		ChunkCount:      2,
		SourceProjectID: cloudImportID(PublicIDPrefixProject),
		Totals:          &CloudImportSessionTotals{Keywords: 10, RankChecks: 100},
	}
}

func TestGetCloudImportCompatibilityV7(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertEqual(t, r.Method, http.MethodGet)
		assertEqual(t, r.URL.Path, "/api/v1/cloud/import/compatibility")
		assertEqual(t, r.Header.Get("Authorization"), "")
		writeJSON(t, w, http.StatusOK, map[string]any{
			"app_version":               "1.4.0",
			"latest_migration":          nil,
			"schema_versions_supported": []int{5},
		})
	}))
	defer server.Close()

	response, err := newTestClient(t, server.URL+"/api/v1").GetCloudImportCompatibility(context.Background())
	if err != nil {
		if responseErr, ok := err.(*ResponseError); ok {
			t.Fatalf("response decode cause: %v", responseErr.Cause)
		}
		t.Fatal(err)
	}
	assertEqual(t, response.AppVersion, "1.4.0")
	if !reflect.DeepEqual(response.SchemaVersionsSupported, []int{5}) {
		t.Fatalf("SchemaVersionsSupported = %#v, want []int{5}", response.SchemaVersionsSupported)
	}
}

func TestGetCloudImportCompatibilityAcceptsOlderServerVersion(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"app_version":               "1.3.0",
			"latest_migration":          nil,
			"schema_versions_supported": []int{4},
		})
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").GetCloudImportCompatibility(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestImportCloudExportWritesExactV7Package(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import")
		assertEqual(t, captured.Header.Get("Authorization"), "Bearer "+testMigrationToken)
		assertJSONEqual(t, captured.Body, fmt.Sprintf(`{
			"version":7,
			"project_id":%q,
			"scope":"current",
			"keywords":[{"id":%q,"keyword":"rank tracker","device":"desktop","location":"United States","location_key":"US"}],
			"alert_rules":[{"id":%q,"name":"Position drop","targets":[{"keyword_id":%q,"type":"keyword"}]}],
			"competitors":[{"id":%q,"domain":"example.com"}],
			"notification_preferences":[{}],
			"saved_views":[{"id":%q,"name":"Tracked keywords"}]
		}`,
			cloudImportID(PublicIDPrefixProject), cloudImportID(PublicIDPrefixKeyword), cloudImportID(PublicIDPrefixRule),
			cloudImportID(PublicIDPrefixKeyword), cloudImportID(PublicIDPrefixComp), cloudImportID(PublicIDPrefixView)))
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"counts": map[string]int{"keywords": 1},
			"job_id": cloudImportID(PublicIDPrefixJob),
			"state":  "done",
		})
	}))
	defer server.Close()

	response, err := newTestClient(t, server.URL+"/api/v1").ImportCloudExport(context.Background(), testMigrationToken, validCloudImportPackage())
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.JobID, cloudImportID(PublicIDPrefixJob))
	assertEqual(t, response.State, CloudImportStateDone)
}

func TestImportCloudExportRejectsInvalidV7PackageBeforeRequest(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")

	missingArrays := validCloudImportPackage()
	missingArrays.AlertRules = nil
	rawProjectID := validCloudImportPackage()
	rawProjectID.ProjectID = "1"
	for name, pkg := range map[string]CloudImportPackage{
		"missing required array": missingArrays,
		"raw project id":         rawProjectID,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := client.ImportCloudExport(context.Background(), testMigrationToken, pkg)
			if err == nil {
				t.Fatal("expected invalid package to fail")
			}
		})
	}
	if requests != 0 {
		t.Fatalf("invalid package made %d requests", requests)
	}
}

func TestCreateCloudImportSessionWritesV7SourceProject(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions")
		assertJSONEqual(t, captured.Body, fmt.Sprintf(`{"version":7,"chunk_count":2,"source_project_id":%q,"totals":{"keywords":10,"rank_checks":100}}`, cloudImportID(PublicIDPrefixProject)))
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"session_id": cloudImportID(PublicIDPrefixJob),
			"state":      "receiving",
			"chunk_limits": map[string]int{
				"max_body_bytes":   1048576,
				"max_history_rows": 5000,
				"max_keywords":     500,
			},
		})
	}))
	defer server.Close()

	response, err := newTestClient(t, server.URL+"/api/v1").CreateCloudImportSession(context.Background(), testMigrationToken, validCloudImportSessionCreate())
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.SessionID, cloudImportID(PublicIDPrefixJob))
	assertEqual(t, response.State, CloudImportStateReceiving)
}

func TestCreateCloudImportSessionRejectsInvalidSourceProjectBeforeRequest(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()
	input := validCloudImportSessionCreate()
	input.SourceProjectID = cloudImportID(PublicIDPrefixKeyword)
	_, err := newTestClient(t, server.URL+"/api/v1").CreateCloudImportSession(context.Background(), testMigrationToken, input)
	if err == nil {
		t.Fatal("expected invalid source_project_id to fail")
	}
	if requests != 0 {
		t.Fatal("invalid create-session input sent a request")
	}
}

func TestCloudImportSessionResponseRejectsRetiredJobID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusCreated, map[string]any{
			"session_id": "job_a00000000000000000000000",
			"state":      "receiving",
			"chunk_limits": map[string]int{
				"max_body_bytes":   1,
				"max_history_rows": 1,
				"max_keywords":     1,
			},
		})
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").CreateCloudImportSession(context.Background(), testMigrationToken, validCloudImportSessionCreate())
	if err == nil {
		t.Fatal("expected retired job_ session response to fail")
	}
}

func TestUploadCloudImportChunksUseImportPathAndDiscriminatedBodies(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPut)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions/"+cloudImportID(PublicIDPrefixJob)+"/chunks/0")
		assertJSONEqual(t, captured.Body, fmt.Sprintf(`{"checksum":%q,"kind":"keywords","keywords":[{"id":%q,"keyword":"rank tracker","device":"desktop","location":"United States","location_key":"US"}]}`, cloudImportChecksum, cloudImportID(PublicIDPrefixKeyword)))
		writeJSON(t, w, http.StatusOK, map[string]any{"state": "receiving", "chunks_received": 1, "chunk_count": 2})
	}))
	defer server.Close()

	chunk := CloudImportKeywordsChunk{Checksum: cloudImportChecksum, Keywords: validCloudImportPackage().Keywords}
	response, err := newTestClient(t, server.URL+"/api/v1").UploadCloudImportChunk(context.Background(), testMigrationToken, cloudImportID(PublicIDPrefixJob), 0, chunk)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.State, CloudImportStateReceiving)
}

func TestUploadCloudImportChunkRawRejectsRetiredJobID(t *testing.T) {
	t.Parallel()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()
	_, err := newTestClient(t, server.URL+"/api/v1").UploadCloudImportChunkRaw(context.Background(), testMigrationToken, "job_a00000000000000000000000", 0, bytes.NewBufferString(`{}`), false)
	if err == nil {
		t.Fatal("expected retired job_ path to fail")
	}
	if requests != 0 {
		t.Fatal("legacy session path sent a request")
	}
}

func TestFinalizeCloudImportSessionUsesImportPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := captureRequest(t, r)
		assertEqual(t, captured.Method, http.MethodPost)
		assertEqual(t, captured.Path, "/api/v1/cloud/import/sessions/"+cloudImportID(PublicIDPrefixJob)+"/finalize")
		writeJSON(t, w, http.StatusOK, map[string]any{"counts": map[string]int{}, "job_id": cloudImportID(PublicIDPrefixJob), "state": "done"})
	}))
	defer server.Close()

	response, err := newTestClient(t, server.URL+"/api/v1").FinalizeCloudImportSession(context.Background(), testMigrationToken, cloudImportID(PublicIDPrefixJob))
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, response.JobID, cloudImportID(PublicIDPrefixJob))
}

func TestCloudImportV7JSONRejectsLegacyAndIncompleteShapes(t *testing.T) {
	t.Parallel()

	valid := fmt.Sprintf(`{
		"version":7,"project_id":%q,"keywords":[{"id":%q,"keyword":"rank tracker","device":"desktop","location":"United States","location_key":"US"}],
		"alert_rules":[],"competitors":[],"notification_preferences":[],"saved_views":[]
	}`, cloudImportID(PublicIDPrefixProject), cloudImportID(PublicIDPrefixKeyword))

	cases := map[string]string{
		"v4":              stringsReplaceOnce(valid, `"version":7`, `"version":4`),
		"camel project":   stringsReplaceOnce(valid, `"project_id"`, `"projectId"`),
		"camel exported":  stringsReplaceOnce(valid, `"version":7`, `"version":7,"exportedAt":"2026-07-27T00:00:00Z"`),
		"null exported":   stringsReplaceOnce(valid, `"version":7`, `"version":7,"exported_at":null`),
		"top-level ranks": stringsReplaceOnce(valid, `"saved_views":[]`, `"saved_views":[],"rank_checks":[]`),
		"missing array":   stringsReplaceOnce(valid, `,"saved_views":[]`, ``),
		"raw keyword id":  stringsReplaceOnce(valid, cloudImportID(PublicIDPrefixKeyword), `"1"`),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var decoded CloudImportPackage
			if err := json.Unmarshal([]byte(body), &decoded); err == nil {
				t.Fatalf("expected %s package to fail", name)
			}
		})
	}
}

func TestCloudImportV7JSONRejectsInvalidDiscriminatorsAndResponses(t *testing.T) {
	t.Parallel()

	var rule CloudImportAlertRule
	err := json.Unmarshal([]byte(fmt.Sprintf(`{"id":%q,"name":"Rule","targets":[{"type":"keyword","keyword_id":%q,"tag":"wrong"}]}`, cloudImportID(PublicIDPrefixRule), cloudImportID(PublicIDPrefixKeyword))), &rule)
	if err == nil {
		t.Fatal("expected mixed alert target to fail")
	}

	var session CloudImportSessionCreate
	err = json.Unmarshal([]byte(fmt.Sprintf(`{"version":7,"chunk_count":1,"source_project_id":%q,"sourceProjectId":%q}`, cloudImportID(PublicIDPrefixProject), cloudImportID(PublicIDPrefixProject))), &session)
	if err == nil {
		t.Fatal("expected camel session alias to fail")
	}

	var response CloudImportSessionCreateResponse
	err = json.Unmarshal([]byte(`{"session_id":"job_a00000000000000000000000","state":"receiving","chunk_limits":{"max_body_bytes":1,"max_history_rows":1,"max_keywords":1}}`), &response)
	if err == nil {
		t.Fatal("expected retired job_ response to fail")
	}
}

func TestCloudImportV7PublicTypeContract(t *testing.T) {
	t.Parallel()

	var _ CloudImportUploadChunk = CloudImportKeywordsChunk{}
	var _ CloudImportUploadChunk = CloudImportSectionsChunk{}
	var _ CloudImportAlertRuleTarget = CloudImportKeywordAlertTarget{}
	var _ CloudImportAlertRuleTarget = CloudImportTagAlertTarget{}

	for _, field := range []string{"RankChecks"} {
		if _, ok := reflect.TypeOf(CloudImportPackage{}).FieldByName(field); ok {
			t.Fatalf("CloudImportPackage retains legacy %s field", field)
		}
	}
	if _, ok := reflect.TypeOf(CloudImportSessionCreate{}).FieldByName("Version"); !ok {
		t.Fatal("CloudImportSessionCreate must expose Version")
	}
	if field, ok := reflect.TypeOf(CloudImportPackage{}).FieldByName("ProjectID"); !ok || field.Tag.Get("json") != "project_id" {
		t.Fatal("CloudImportPackage.ProjectID must use project_id")
	}
	if field, ok := reflect.TypeOf(CloudImportSessionCreate{}).FieldByName("SourceProjectID"); !ok || field.Tag.Get("json") != "source_project_id" {
		t.Fatal("CloudImportSessionCreate.SourceProjectID must use source_project_id")
	}
	if _, ok := reflect.TypeOf(CloudImportPackage{}).FieldByName("ExportedAt"); !ok {
		t.Fatal("CloudImportPackage must expose ExportedAt")
	}

	sections, err := json.Marshal(CloudImportSectionsChunk{Checksum: cloudImportChecksum, Sections: CloudImportSessionSections{}})
	if err != nil {
		t.Fatal(err)
	}
	assertJSONEqual(t, string(sections), fmt.Sprintf(`{"checksum":%q,"kind":"sections","sections":{}}`, cloudImportChecksum))
}

func TestCloudImportPackageRoundTripUsesV7Protocol(t *testing.T) {
	t.Parallel()

	input := validCloudImportPackage()
	input.ExportedAt = timePtr(time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC))
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte(`"version":7`)) || !bytes.Contains(body, []byte(`"exported_at"`)) {
		t.Fatalf("v7 package JSON = %s", body)
	}
	var output CloudImportPackage
	if err := json.Unmarshal(body, &output); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(output, input) {
		t.Fatalf("round trip = %#v, want %#v", output, input)
	}
}

func TestCloudImportV7NestedJSONCodecs(t *testing.T) {
	t.Parallel()

	position := 7
	enabled := true
	alertEmail := true
	label := "Primary competitor"
	url := "https://example.com/rank-tracker"
	checkedAt := time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)
	keywordTarget := CloudImportKeywordAlertTarget{
		Device:    DeviceMobile,
		Keyword:   "rank tracker",
		KeywordID: cloudImportID(PublicIDPrefixKeyword),
		Location:  "Poland",
	}
	tagTarget := CloudImportTagAlertTarget{Tag: "priority"}
	rule := CloudImportAlertRule{
		ChangePct:         float64Ptr(10),
		Channels:          []AlertChannel{AlertChannelEmail, AlertChannelWebhook},
		CompetitorDomain:  &url,
		ConditionType:     CloudImportAlertConditionPositionDrop,
		DropPositions:     &position,
		Enabled:           &enabled,
		ID:                cloudImportID(PublicIDPrefixRule),
		Name:              "Rank change",
		SerpFeature:       &url,
		TargetType:        AlertTargetTypeKeyword,
		Targets:           []CloudImportAlertRuleTarget{keywordTarget, tagTarget},
		ThresholdPosition: &position,
		TopN:              &position,
	}
	ranking := CloudImportRankingHistory{NormalizationVersion: "v2", Provider: "serpapi", CheckedAt: checkedAt, Position: &position, PreviousPosition: &position, RankingURL: &url}
	keyword := CloudImportKeyword{
		Device: DeviceDesktop, ID: cloudImportID(PublicIDPrefixKeyword), Keyword: "rank tracker", Location: "United States",
		RankingHistory: []CloudImportRankingHistory{ranking}, Tags: []string{"priority"}, TargetURL: &url,
	}
	competitor := CloudImportCompetitor{Domain: "example.com", ID: cloudImportID(PublicIDPrefixComp), Label: &label}
	preference := CloudImportNotificationPreference{AlertEmail: &alertEmail, ReportEmail: &alertEmail}
	savedView := CloudImportSavedView{Config: map[string]any{"columns": []any{"keyword"}}, ID: cloudImportID(PublicIDPrefixView), Name: "Keywords", Surface: CloudImportSavedViewSurfaceKeywords}
	sections := CloudImportSessionSections{
		AlertRules: []CloudImportAlertRule{rule}, Competitors: []CloudImportCompetitor{competitor},
		NotificationPreferences: []CloudImportNotificationPreference{preference}, SavedViews: []CloudImportSavedView{savedView},
		SourceKeywordIDs: map[string]CloudImportSourceKeyword{"source-1": {Device: DeviceDesktop, Location: "United States", Text: "rank tracker"}},
	}

	assertCloudImportJSONRoundTrip(t, CloudImportCompatibility{AppVersion: "1.4.0", LatestMigration: &url, SchemaVersionsSupported: []int{5}})
	assertCloudImportJSONRoundTrip(t, CloudImportFinalizeResponse{Counts: CloudImportCounts{"keywords": 1}, JobID: cloudImportID(PublicIDPrefixJob), State: CloudImportStateDone})
	assertCloudImportJSONRoundTrip(t, keywordTarget)
	assertCloudImportJSONRoundTrip(t, tagTarget)
	assertCloudImportJSONRoundTrip(t, rule)
	assertCloudImportJSONRoundTrip(t, competitor)
	assertCloudImportJSONRoundTrip(t, ranking)
	assertCloudImportJSONRoundTrip(t, keyword)
	assertCloudImportJSONRoundTrip(t, preference)
	assertCloudImportJSONRoundTrip(t, savedView)
	assertCloudImportJSONRoundTrip(t, CloudImportSessionTotals{Keywords: 1, RankChecks: 2})
	assertCloudImportJSONRoundTrip(t, CloudImportSessionCreate{Version: 7, ChunkCount: 1, SourceProjectID: cloudImportID(PublicIDPrefixProject), Totals: &CloudImportSessionTotals{}})
	assertCloudImportJSONRoundTrip(t, CloudImportChunkLimits{MaxBodyBytes: 1, MaxHistoryRows: 1, MaxKeywords: 1})
	assertCloudImportJSONRoundTrip(t, CloudImportSessionCreateResponse{ChunkLimits: CloudImportChunkLimits{MaxBodyBytes: 1, MaxHistoryRows: 1, MaxKeywords: 1}, SessionID: cloudImportID(PublicIDPrefixJob), State: CloudImportStateReceiving})
	assertCloudImportJSONRoundTrip(t, CloudImportSourceKeyword{Device: DeviceDesktop, Location: "United States", Text: "rank tracker"})
	assertCloudImportJSONRoundTrip(t, sections)
	assertCloudImportJSONRoundTrip(t, CloudImportKeywordsChunk{Checksum: cloudImportChecksum, Keywords: []CloudImportKeyword{keyword}})
	assertCloudImportJSONRoundTrip(t, CloudImportSectionsChunk{Checksum: cloudImportChecksum, Sections: sections})
	assertCloudImportJSONRoundTrip(t, CloudImportChunkResponse{ChunkCount: 1, ChunksReceived: 0, State: CloudImportStateReceiving})
}

func assertCloudImportJSONRoundTrip[T any](t *testing.T, input T) {
	t.Helper()
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var output T
	if err := json.Unmarshal(body, &output); err != nil {
		t.Fatalf("unmarshal %T: %v", input, err)
	}
	if !reflect.DeepEqual(output, input) {
		t.Fatalf("round trip %T = %#v, want %#v", input, output, input)
	}
}

func stringsReplaceOnce(value, old, replacement string) string {
	index := bytes.Index([]byte(value), []byte(old))
	if index < 0 {
		panic("test fixture replacement not found")
	}
	return value[:index] + replacement + value[index+len(old):]
}

func timePtr(value time.Time) *time.Time { return &value }

func float64Ptr(value float64) *float64 { return &value }

func TestCloudImportCurrentVersions(t *testing.T) {
	for _, version := range []int{6, 7} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			locationKey := ""
			if version == 7 {
				locationKey = `,"location_key":"US/New York/New York"`
			}
			location := "United States"
			if version == 7 {
				location = "New York"
			}
			payload := fmt.Sprintf(`{"version":%d,"project_id":"prj_a00000000000000000000000",
                "keywords":[{"id":"kw_a00000000000000000000000","keyword":"rank tracker",
                "device":"desktop","location":%q%s,"rankingHistory":[{
                "checkedAt":"2026-10-01T00:00:00Z","position":null,"previousPosition":4,
                "rankingUrl":null,"provider":"serpapi","normalizationVersion":"v2","requestedDepth":20}]}],
                "alert_rules":[{"id":"alr_a00000000000000000000000","name":"Drop","severity":"warning"}],"competitors":[],"notification_preferences":[],"saved_views":[]}`, version, location, locationKey)
			var pkg CloudImportPackage
			if err := json.Unmarshal([]byte(payload), &pkg); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(pkg)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, string(encoded), payload)
			var session CloudImportSessionCreate
			if err := json.Unmarshal([]byte(fmt.Sprintf(`{"version":%d,"chunk_count":1,"source_project_id":"prj_a00000000000000000000000"}`, version)), &session); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCloudImportCompatibilityAcceptsServerVersions(t *testing.T) {
	var compatibility CloudImportCompatibility
	if err := json.Unmarshal([]byte(`{"app_version":"1.0.0","latest_migration":null,"schema_versions_supported":[7,6,8]}`), &compatibility); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(compatibility.SchemaVersionsSupported, []int{7, 6, 8}) {
		t.Fatal(compatibility)
	}
}

func TestCloudImportRejectsAmbiguousLegacyHistory(t *testing.T) {
	payload := `{"version":5,"project_id":"prj_a00000000000000000000000",
        "keywords":[{"id":"kw_a00000000000000000000000","keyword":"rank tracker","device":"desktop","location":"United States",
        "rankingHistory":[{"checkedAt":"2026-10-01T00:00:00Z","position":3}]}],
        "alert_rules":[],"competitors":[],"notification_preferences":[],"saved_views":[]}`
	var pkg CloudImportPackage
	if err := json.Unmarshal([]byte(payload), &pkg); err == nil {
		t.Fatal("legacy history must be rejected")
	}
}

func TestCloudImportVersionGuards(t *testing.T) {
	for name, mutate := range map[string]func(*CloudImportPackage){
		"v7 without key":         func(p *CloudImportPackage) { p.Keywords[0].LocationKey = "" },
		"v6 with key":            func(p *CloudImportPackage) { p.Version = 6 },
		"future request version": func(p *CloudImportPackage) { p.Version = 8 },
		"legacy history": func(p *CloudImportPackage) {
			p.Version = 5
			p.Keywords[0].LocationKey = ""
			p.Keywords[0].RankingHistory = []CloudImportRankingHistory{{CheckedAt: time.Now(), NormalizationVersion: "v2", Provider: "serpapi"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := validCloudImportPackage()
			mutate(&p)
			if _, err := json.Marshal(p); err == nil {
				t.Fatal("invalid package accepted")
			}
		})
	}
	for _, versions := range []string{`["7"]`, `[7.1]`, `[true]`, `[0]`, `[-1]`} {
		var compatibility CloudImportCompatibility
		if err := json.Unmarshal([]byte(`{"app_version":"1","latest_migration":null,"schema_versions_supported":`+versions+`}`), &compatibility); err == nil {
			t.Fatal("non-integer version accepted")
		}
	}
}

func TestCloudImportLocationIdentityRoundTrip(t *testing.T) {
	assertCloudImportJSONRoundTrip(t, CloudImportKeywordAlertTarget{KeywordID: cloudImportID(PublicIDPrefixKeyword), Location: "New York", LocationKey: "US/New York/New York"})
	assertCloudImportJSONRoundTrip(t, CloudImportSourceKeyword{Device: DeviceDesktop, Text: "rank tracker", Location: "New York", LocationKey: "US/New York/New York"})
	p := validCloudImportPackage()
	p.Version = 5
	p.Keywords[0].LocationKey = ""
	assertCloudImportJSONRoundTrip(t, p)
}
