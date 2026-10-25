package bisibility

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const providerEvidenceJSON = `{"acceptance":"accepted","fulfillment":"partial","returned_depth":0,"accounting":"pending","result":"unresolved","cost":"unknown","original_checked_at":"2026-10-20T12:00:00Z","recovery":{"state":"scheduled","last_attempt_at":null,"next_attempt_at":"2026-10-21T12:00:00Z","reason":"provider_pending"}}`

func assertEvidenceJSONField(t *testing.T, raw []byte, field string, want any) {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if got, ok := body[field]; !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v (present %v), want %#v; JSON: %s", field, got, ok, want, raw)
	}
}

func TestEvidenceFieldsPreferencePatchTransport(t *testing.T) {
	for _, setting := range []CheckNotificationSettingInput{CheckNotificationSettingInputAll, CheckNotificationSettingInputImportant, CheckNotificationSettingInputNone} {
		for _, boundary := range []int{1, 100} {
			t.Run(string(setting)+"/"+strconv.Itoa(boundary), func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/projects/"+rankRunProject+"/notification-preferences" {
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					}
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					want := map[string]any{"check_notification_setting": string(setting), "check_notification_top_n": float64(boundary), "check_notification_position_move": float64(boundary), "provider_incident_in_app": false, "check_in_app": true}
					if !reflect.DeepEqual(body, want) {
						t.Errorf("patch = %#v, want %#v", body, want)
					}
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Deprecation", "true")
					w.Write([]byte(`{"project_id":"prj_aaaaaaaaaaaaaaaaaaaaaaaa","check_in_app":true,"check_notification_setting":"important","check_notification_top_n":10,"check_notification_position_move":5,"provider_incident_in_app":false,"warnings":[{"code":"deprecated_field_ignored","field":"check_in_app","message":"Use check_notification_setting."}]}`))
				}))
				defer server.Close()
				result, err := newTestClient(t, server.URL+"/api/v1").UpdateNotificationPreferences(context.Background(), rankRunProject, UpdateNotificationPreferencesInput{
					CheckNotificationSetting: &setting, CheckNotificationTopN: &boundary, CheckNotificationPositionMove: &boundary, ProviderIncidentInApp: boolPtr(false), CheckInApp: boolPtr(true),
				})
				if err != nil {
					t.Fatal(err)
				}
				if calls != 1 || result.CheckNotificationSetting != CheckNotificationSettingImportant || result.CheckNotificationTopN != 10 || result.CheckNotificationPositionMove != 5 || result.ProviderIncidentInApp == nil || *result.ProviderIncidentInApp || len(result.Warnings) != 1 || result.Warnings[0].Field != "check_in_app" || result.Warnings[0].Code != "deprecated_field_ignored" || result.Warnings[0].Message == "" {
					t.Fatalf("updated preferences or warning lost: %#v", result)
				}
			})
		}
	}
	raw, err := json.Marshal(UpdateNotificationPreferencesInput{})
	if err != nil || string(raw) != "{}" {
		t.Fatalf("omitted patch fields must stay omitted: %s, %v", raw, err)
	}
}

func TestEvidenceFieldsPreferenceValidation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer server.Close()
	client := newTestClient(t, server.URL+"/api/v1")
	for _, setting := range []CheckNotificationSettingInput{"", "future_value", "ALL"} {
		_, err := client.UpdateNotificationPreferences(context.Background(), rankRunProject, UpdateNotificationPreferencesInput{CheckNotificationSetting: &setting})
		var configurationError *ConfigurationError
		if !errors.As(err, &configurationError) {
			t.Errorf("invalid input enum accepted: %q: %v", setting, err)
		}
	}
	for _, value := range []int{-1, 0, 101} {
		for _, input := range []UpdateNotificationPreferencesInput{{CheckNotificationTopN: &value}, {CheckNotificationPositionMove: &value}} {
			if _, err := client.UpdateNotificationPreferences(context.Background(), rankRunProject, input); err == nil {
				t.Errorf("invalid threshold accepted: %#v", input)
			}
		}
	}
	if calls != 0 {
		t.Fatalf("invalid preferences sent %d requests", calls)
	}
	for _, raw := range []string{`{"check_notification_setting":"future_value"}`, `{"check_notification_top_n":0}`, `{"check_notification_position_move":101}`, `{"provider_incident_in_app":null}`, `{"check_notification_setting":null}`, `{"check_notification_top_n":null}`, `{"check_notification_position_move":null}`} {
		var preference CloudImportNotificationPreference
		if err := json.Unmarshal([]byte(raw), &preference); err == nil {
			t.Errorf("invalid cloud preference accepted: %s", raw)
		}
	}
	invalidSetting := CheckNotificationSettingInput("future_value")
	for _, input := range []CloudImportNotificationPreference{{CheckNotificationSetting: &invalidSetting}, {CheckNotificationTopN: intPtr(0)}, {CheckNotificationPositionMove: intPtr(101)}} {
		if _, err := json.Marshal(input); err == nil {
			t.Errorf("invalid cloud preference encoded: %#v", input)
		}
	}
	for _, setting := range []CheckNotificationSettingInput{CheckNotificationSettingInputAll, CheckNotificationSettingInputImportant, CheckNotificationSettingInputNone} {
		for _, boundary := range []int{1, 100} {
			input := CloudImportNotificationPreference{CheckNotificationSetting: &setting, CheckNotificationTopN: &boundary, CheckNotificationPositionMove: &boundary, ProviderIncidentInApp: boolPtr(true)}
			raw, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			var decoded CloudImportNotificationPreference
			if err := json.Unmarshal(raw, &decoded); err != nil || !reflect.DeepEqual(decoded, input) {
				t.Fatalf("valid cloud preference changed: %#v, %v", decoded, err)
			}
		}
	}
}

func TestEvidenceFieldsRejectInvalidRetainedContracts(t *testing.T) {
	validRun := strings.TrimSuffix(rankRunJSON, "}") + `,"evidence_projections":[{"check_id":"check_a00000000000000000000000",` + providerEvidenceJSON[1:] + `]}`
	for _, invalid := range []string{
		strings.Replace(validRun, `"check_id":"check_a00000000000000000000000"`, `"check_id":"invalid"`, 1),
		strings.Replace(validRun, `"returned_depth":0`, `"returned_depth":-1`, 1),
		strings.Replace(validRun, `"returned_depth":0`, `"returned_depth":1.5`, 1),
		strings.Replace(validRun, `"acceptance":"accepted",`, ``, 1),
		strings.Replace(validRun, `"last_attempt_at":null,`, ``, 1),
		strings.Replace(validRun, `"original_checked_at":"2026-10-20T12:00:00Z"`, `"original_checked_at":"invalid"`, 1),
	} {
		var run RankCheckRunRead
		if err := json.Unmarshal([]byte(invalid), &run); err == nil {
			t.Errorf("invalid run evidence accepted: %s", invalid)
		}
	}
	var page struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(reportFixture(t)["observations"], &page); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{
		strings.Replace(providerEvidenceJSON, `"returned_depth":0`, `"returned_depth":-1`, 1),
		strings.Replace(providerEvidenceJSON, `"acceptance":"accepted",`, ``, 1),
		strings.Replace(providerEvidenceJSON, `"last_attempt_at":null,`, ``, 1),
	} {
		page.Data[0]["evidence_projection"] = json.RawMessage(invalid)
		raw, _ := json.Marshal(page.Data[0])
		var record RankHistoryReportRecord
		if err := json.Unmarshal(raw, &record); err == nil {
			t.Errorf("invalid report evidence accepted: %s", raw)
		}
	}
}

func TestEvidenceFieldsOpenResponsesAndMissingValues(t *testing.T) {
	for _, completeness := range []ResultCompleteness{ResultCompletenessComplete, ResultCompletenessPartial, ResultCompletenessUnresolved, "future_value"} {
		var check RankCheck
		if err := json.Unmarshal([]byte(`{"result_completeness":"`+string(completeness)+`"}`), &check); err != nil || check.ResultCompleteness == nil || *check.ResultCompleteness != completeness {
			t.Fatalf("response enum lost: %#v, %v", check, err)
		}
	}
	for _, raw := range []string{`{}`, `{"comparison_allowed":null,"returned_depth":null,"evidence_projection":null}`} {
		var check RankCheck
		if err := json.Unmarshal([]byte(raw), &check); err != nil || check.ComparisonAllowed != nil || check.ReturnedDepth != nil || check.EvidenceProjection != nil || check.ResultCompleteness != nil {
			t.Fatalf("unknown evidence became known: %#v, %v", check, err)
		}
	}
	var preferences NotificationPreferences
	if err := json.Unmarshal([]byte(`{"check_notification_setting":"future_value"}`), &preferences); err != nil || preferences.CheckNotificationSetting != "future_value" || preferences.ProviderIncidentInApp != nil {
		t.Fatalf("open preference response lost: %#v, %v", preferences, err)
	}
	for _, raw := range []string{`{}`, `{"warnings":null}`, `{"warnings":[]}`} {
		var updated UpdatedNotificationPreferences
		if err := json.Unmarshal([]byte(raw), &updated); err != nil || len(updated.Warnings) != 0 {
			t.Fatalf("optional warnings rejected: %#v, %v", updated, err)
		}
	}
	var projection ProviderEvidenceProjection
	if err := json.Unmarshal([]byte(`{"acceptance":null,"fulfillment":null,"returned_depth":null,"accounting":null,"result":null,"cost":null,"original_checked_at":null,"recovery":{"state":null,"last_attempt_at":null,"next_attempt_at":null,"reason":null}}`), &projection); err != nil || projection.ReturnedDepth != nil || projection.OriginalCheckedAt != nil || projection.Recovery == nil || projection.Recovery.State != nil {
		t.Fatalf("nullable provider facts changed: %#v, %v", projection, err)
	}
}

func TestEvidenceFieldsRunsAndHistory(t *testing.T) {
	for _, evidence := range []string{"null", "[]", `[{"check_id":"check_a00000000000000000000000",` + providerEvidenceJSON[1:] + `]`} {
		raw := []byte(strings.TrimSuffix(rankRunJSON, "}") + `,"evidence_projections":` + evidence + `}`)
		var run RankCheckRunRead
		if err := json.Unmarshal(raw, &run); err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(evidence, "[{") && (len(run.EvidenceProjections) != 1 || run.EvidenceProjections[0].CheckID != "check_a00000000000000000000000" || run.EvidenceProjections[0].ReturnedDepth == nil || *run.EvidenceProjections[0].ReturnedDepth != 0) {
			t.Fatalf("typed run evidence lost: %#v", run)
		}
		encoded, err := json.Marshal(run)
		if err != nil {
			t.Fatal(err)
		}
		var want any
		json.Unmarshal([]byte(evidence), &want)
		assertEvidenceJSONField(t, encoded, "evidence_projections", want)
	}
	for _, stream := range []string{"observations", "full_attempts"} {
		var page struct {
			Data []map[string]json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(reportFixture(t)[stream], &page); err != nil {
			t.Fatal(err)
		}
		for _, projection := range []string{"null", providerEvidenceJSON} {
			record := page.Data[0]
			if stream == "full_attempts" {
				for _, candidate := range page.Data {
					if string(candidate["kind"]) == `"logical_rank_check"` {
						record = candidate
						break
					}
				}
				if string(record["kind"]) != `"logical_rank_check"` {
					t.Fatal("fixture has no logical check attempt")
				}
			}
			record["evidence_projection"] = json.RawMessage(projection)
			raw, _ := json.Marshal(record)
			var row RankHistoryExportRow
			if err := json.Unmarshal(raw, &row); err != nil {
				t.Fatal(err)
			}
			if row.Report == nil {
				t.Fatal("report variant lost")
			}
			var typed *ProviderEvidenceProjection
			if row.Report.Observation != nil {
				typed = row.Report.Observation.EvidenceProjection
			}
			if row.Report.Attempt != nil {
				typed = row.Report.Attempt.EvidenceProjection
			}
			if projection != "null" && (typed == nil || typed.Recovery == nil || typed.Recovery.State == nil) {
				t.Fatal("typed report evidence lost")
			}
			encoded, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			var want any
			json.Unmarshal([]byte(projection), &want)
			assertEvidenceJSONField(t, encoded, "evidence_projection", want)
		}
	}
	for _, projection := range []string{"null", providerEvidenceJSON} {
		var row RankHistoryExportRow
		if err := json.Unmarshal([]byte(`{"checked_at":"2026-10-20T12:00:00Z","evidence_projection":`+projection+`}`), &row); err != nil {
			t.Fatal(err)
		}
		if projection != "null" && row.EvidenceProjection == nil {
			t.Fatal("legacy export evidence lost")
		}
	}
}

func TestEvidenceFieldsApplicationResponseCompatibility(t *testing.T) {
	raw, err := os.ReadFile("testdata/evidence-response-compatibility.json")
	if err != nil {
		t.Fatal(err)
	}
	var captures []struct {
		Operation string          `json:"operationId"`
		Body      json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(raw, &captures); err != nil {
		t.Fatal(err)
	}
	for _, capture := range captures {
		t.Run(capture.Operation, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write(capture.Body)
			}))
			defer server.Close()
			client := newTestClient(t, server.URL+"/api/v1")
			switch capture.Operation {
			case "getNotificationPreferences":
				preferences, err := client.GetNotificationPreferences(context.Background(), rankRunProject)
				if err != nil || preferences.CheckNotificationSetting != CheckNotificationSettingImportant || preferences.CheckNotificationTopN != 10 || preferences.CheckNotificationPositionMove != 5 {
					t.Fatalf("preferences: %#v, %v", preferences, err)
				}
			case "getRankCheckResult":
				check, err := client.GetRankCheckResult(context.Background(), "check_a00000000000000000000000")
				if err != nil || check.EvidenceProjection != nil {
					t.Fatalf("check: %#v, %v", check, err)
				}
			case "listRankCheckRuns":
				runs, err := client.ListRankCheckRuns(context.Background(), "prj_a00000000000000000000000", nil)
				if err != nil || len(runs.Data) != 1 || runs.Data[0].EvidenceProjections != nil {
					t.Fatalf("runs: %#v, %v", runs, err)
				}
			default:
				t.Fatalf("unhandled capture %q", capture.Operation)
			}
		})
	}
}

func TestEvidenceFieldsRetainRankCheckProjection(t *testing.T) {
	raw := []byte(`{"checked_at":"2026-10-20T12:00:00Z","comparison_allowed":false,"result_completeness":"partial","returned_depth":0,"evidence_projection":` + providerEvidenceJSON + `}`)
	var check RankCheck
	if err := json.Unmarshal(raw, &check); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(check)
	if err != nil {
		t.Fatal(err)
	}
	assertEvidenceJSONField(t, encoded, "comparison_allowed", false)
	assertEvidenceJSONField(t, encoded, "result_completeness", "partial")
	assertEvidenceJSONField(t, encoded, "returned_depth", float64(0))
	var projection any
	json.Unmarshal([]byte(providerEvidenceJSON), &projection)
	assertEvidenceJSONField(t, encoded, "evidence_projection", projection)
}

func TestEvidenceFieldsCloudImportPreferenceRoundTrip(t *testing.T) {
	raw := []byte(`{"check_in_app":false,"check_notification_setting":"important","check_notification_top_n":1,"check_notification_position_move":100,"provider_incident_in_app":false}`)
	var preference CloudImportNotificationPreference
	if err := json.Unmarshal(raw, &preference); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(preference)
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	json.Unmarshal(encoded, &got)
	json.Unmarshal(raw, &want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cloud preference changed: %s", encoded)
	}
}
