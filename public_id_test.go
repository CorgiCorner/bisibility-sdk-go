package bisibility

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	publicIDSuffix  = "a00000000000000000000000"
	legacyProjectID = "prj" + "_1"
	legacyKeywordID = "kw" + "_1"
)

func strictID(prefix PublicIDPrefix) string {
	return string(prefix) + "_" + publicIDSuffix
}

func TestPublicIDPrefixRegistry(t *testing.T) {
	want := []string{
		"al", "alr", "audit", "check", "cmp", "conn", "dwh", "ferry", "imp", "inv", "key", "kw",
		"mbr", "ntf", "pat", "prj", "sid", "sig", "svkw", "tag", "usr", "viw", "we",
	}
	got := make([]string, 0, len(publicIDPrefixes))
	for prefix := range publicIDPrefixes {
		got = append(got, string(prefix))
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("public ID prefixes = %v, want %v", got, want)
	}

	for _, prefix := range got {
		value := string(prefix) + "_" + publicIDSuffix
		if err := ValidatePublicID(value); err != nil {
			t.Fatalf("ValidatePublicID(%q): %v", value, err)
		}
		if err := ValidatePublicIDPrefix(value, PublicIDPrefix(prefix)); err != nil {
			t.Fatalf("ValidatePublicIDPrefix(%q, %q): %v", value, prefix, err)
		}
	}
}

func TestPublicIDsRejectLegacyRawAndMixedCaseValues(t *testing.T) {
	for _, value := range []string{
		legacyProjectID,
		"cmf7v7r2e0000a0r2b3c4d5e6",
		"prj_A00000000000000000000000",
		"PRJ_a00000000000000000000000",
		"unknown_a00000000000000000000000",
		"prj_a0000000000000000000000",
		"prj_a000000000000000000000000",
	} {
		if IsPublicID(value) {
			t.Fatalf("IsPublicID(%q) = true, want false", value)
		}
		if err := ValidatePublicID(value); err == nil {
			t.Fatalf("ValidatePublicID(%q) succeeded", value)
		}
	}

	if err := ValidatePublicIDPrefix(strictID(PublicIDPrefixProject), PublicIDPrefixKeyword); err == nil {
		t.Fatal("expected a namespace mismatch error")
	}
}

func TestPublicIDsRejectRetiredV2Prefixes(t *testing.T) {
	for _, prefix := range []string{
		"alert", "rule", "comp", "hook", "invite", "job", "member", "mtok", "notif", "ses", "skw", "view", "webhook",
	} {
		value := prefix + "_" + publicIDSuffix
		if IsPublicID(value) {
			t.Fatalf("IsPublicID(%q) = true, want false", value)
		}
	}
}

func TestTypedPublicIDInputsFailBeforeHTTP(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api/v1")
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "path project",
			call: func() error {
				_, err := client.GetProject(context.Background(), legacyProjectID)
				return err
			},
		},
		{
			name: "path keyword",
			call: func() error {
				_, err := client.GetKeyword(context.Background(), "raw-keyword-id")
				return err
			},
		},
		{
			name: "query connection",
			call: func() error {
				_, err := client.ResearchKeywords(context.Background(), strictID(PublicIDPrefixProject), ResearchKeywordsOptions{ConnectionID: strictID(PublicIDPrefixProject), Seed: "test"})
				return err
			},
		},
		{
			name: "body keyword IDs",
			call: func() error {
				_, err := client.BulkUpdateKeywords(context.Background(), KeywordBulkInput{KeywordIDs: []string{strictID(PublicIDPrefixProject)}, Operation: KeywordBulkOperationAddTags})
				return err
			},
		},
		{
			name: "body alert target",
			call: func() error {
				_, err := client.CreateAlertRule(context.Background(), strictID(PublicIDPrefixProject), CreateAlertRuleInput{TargetType: AlertTargetTypeKeyword, TargetIDs: []string{strictID(PublicIDPrefixProject)}})
				return err
			},
		},
		{
			name: "body alert recipient",
			call: func() error {
				_, err := client.CreateAlertRule(context.Background(), strictID(PublicIDPrefixProject), CreateAlertRuleInput{RecipientIDs: []string{strictID(PublicIDPrefixProject)}})
				return err
			},
		},
		{
			name: "cloud import package",
			call: func() error {
				_, err := client.ImportCloudExport(context.Background(), "mig_secret_do_not_log", CloudImportPackage{ProjectID: strictID(PublicIDPrefixKeyword)})
				return err
			},
		},
		{
			name: "cloud import session path",
			call: func() error {
				_, err := client.UploadCloudImportChunk(context.Background(), "mig_secret_do_not_log", strictID(PublicIDPrefixProject), 0, CloudImportKeywordsChunk{})
				return err
			},
		},
		{
			name: "path API key",
			call: func() error {
				_, err := client.RevokeAPIKey(context.Background(), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path rank check",
			call: func() error {
				_, err := client.GetRankCheckResult(context.Background(), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path alert rule",
			call: func() error {
				_, err := client.DeleteAlertRule(context.Background(), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path triggered alert",
			call: func() error {
				_, err := client.MuteTriggeredAlert(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path team member",
			call: func() error {
				_, err := client.RemoveTeamMember(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path team invite",
			call: func() error {
				_, err := client.RevokeProjectTeamInvite(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path saved view",
			call: func() error {
				_, err := client.DeleteProjectSavedView(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path competitor",
			call: func() error {
				_, err := client.RemoveProjectCompetitor(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path migration token",
			call: func() error {
				_, err := client.RevokeProjectMigrationToken(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path personal access token",
			call: func() error {
				_, err := client.RevokeMyToken(context.Background(), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path webhook",
			call: func() error {
				_, err := client.DeleteWebhook(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixProject))
				return err
			},
		},
		{
			name: "path sitemap monitor",
			call: func() error {
				_, err := client.UpdateSitemapMonitor(context.Background(), strictID(PublicIDPrefixProject), strictID(PublicIDPrefixKeyword), UpdateSitemapMonitorInput{})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			var configurationErr *ConfigurationError
			if !errors.As(err, &configurationErr) {
				t.Fatalf("error = %T (%v), want ConfigurationError", err, err)
			}
		})
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
}

func TestProjectHeaderRequiresStrictProjectID(t *testing.T) {
	if _, err := NewClient(WithProjectID(legacyProjectID)); err == nil {
		t.Fatal("WithProjectID accepted a legacy project ID")
	}
	if _, err := NewClient(WithProjectID(strictID(PublicIDPrefixProject))); err != nil {
		t.Fatalf("WithProjectID rejected a strict project ID: %v", err)
	}
}

func TestPublicIDResponseContractRejectsLegacyValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{"id": legacyProjectID})
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").GetProject(context.Background(), strictID(PublicIDPrefixProject))
	var responseErr *ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("error = %T (%v), want ResponseError", err, err)
	}
	if !strings.Contains(responseErr.Cause.Error(), "public ID response contract") {
		t.Fatalf("response error cause = %v, want public ID contract failure", responseErr.Cause)
	}
}

func TestPublicIDResponseContractRejectsMissingRequiredID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{"id": ""})
	}))
	defer server.Close()

	_, err := newTestClient(t, server.URL+"/api/v1").GetProject(context.Background(), strictID(PublicIDPrefixProject))
	var responseErr *ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("error = %T (%v), want ResponseError", err, err)
	}
	if !strings.Contains(responseErr.Cause.Error(), "response.id is required") {
		t.Fatalf("response error cause = %v, want required ID failure", responseErr.Cause)
	}
}

func TestPublicIDResponseContractAllowsOptionalIDsOnlyWhenAbsent(t *testing.T) {
	if err := validateResponsePublicIDs(CompetitorColumn{}); err != nil {
		t.Fatalf("empty optional string response ID was rejected: %v", err)
	}

	empty := ""
	view := SavedView{ID: strictID(PublicIDPrefixView)}
	if err := validateResponsePublicIDs(view); err != nil {
		t.Fatalf("nil optional pointer response ID was rejected: %v", err)
	}
	view.CreatedByID = &empty
	if err := validateResponsePublicIDs(view); err == nil {
		t.Fatal("present empty pointer response ID was accepted")
	}
}

func TestAlertRuleResponsePublicIDs(t *testing.T) {
	rule := AlertRule{
		ID:           strictID(PublicIDPrefixRule),
		RecipientIDs: []string{strictID(PublicIDPrefixUser)},
		TargetIDs:    []string{strictID(PublicIDPrefixKeyword)},
		TargetType:   AlertTargetTypeKeyword,
	}
	if err := validateResponsePublicIDs(rule); err != nil {
		t.Fatalf("valid alert rule response was rejected: %v", err)
	}

	rule.TargetIDs = []string{strictID(PublicIDPrefixTag)}
	if err := validateResponsePublicIDs(rule); err == nil {
		t.Fatal("tag target was accepted for a keyword alert rule")
	}

	rule.TargetIDs = []string{strictID(PublicIDPrefixTag)}
	rule.TargetType = AlertTargetTypeTag
	if err := validateResponsePublicIDs(rule); err != nil {
		t.Fatalf("valid tag alert rule response was rejected: %v", err)
	}

	rule.RecipientIDs = []string{strictID(PublicIDPrefixKeyword)}
	if err := validateResponsePublicIDs(rule); err == nil {
		t.Fatal("non-user recipient was accepted")
	}
}

func TestPublicIDResponseSchemas(t *testing.T) {
	contracts := map[reflect.Type]map[string]PublicIDPrefix{
		reflect.TypeOf(Project{}):                          {"id": PublicIDPrefixProject},
		reflect.TypeOf(ProjectDefaults{}):                  {"project_id": PublicIDPrefixProject},
		reflect.TypeOf(ProjectOverview{}):                  {"project_id": PublicIDPrefixProject},
		reflect.TypeOf(APIKey{}):                           {"id": PublicIDPrefixKey},
		reflect.TypeOf(Keyword{}):                          {"id": PublicIDPrefixKeyword, "project_id": PublicIDPrefixProject},
		reflect.TypeOf(KeywordMatch{}):                     {"keyword_id": PublicIDPrefixKeyword},
		reflect.TypeOf(KeywordBulkItemResult{}):            {"keyword_id": PublicIDPrefixKeyword},
		reflect.TypeOf(RankCheck{}):                        {"id": PublicIDPrefixCheck, "keyword_id": PublicIDPrefixKeyword},
		reflect.TypeOf(RankHistoryExportRow{}):             {"id": PublicIDPrefixCheck, "keyword_id": PublicIDPrefixKeyword},
		reflect.TypeOf(AlertRule{}):                        {"id": PublicIDPrefixRule},
		reflect.TypeOf(TriggeredAlert{}):                   {"id": PublicIDPrefixAlert},
		reflect.TypeOf(TeamMember{}):                       {"id": PublicIDPrefixMember},
		reflect.TypeOf(TeamInvite{}):                       {"id": PublicIDPrefixInvite},
		reflect.TypeOf(CreatedTeamInvite{}):                {"id": PublicIDPrefixInvite},
		reflect.TypeOf(RevokeTeamInviteResult{}):           {"id": PublicIDPrefixInvite},
		reflect.TypeOf(SitemapMonitor{}):                   {"id": PublicIDPrefixProject, "project_id": PublicIDPrefixProject},
		reflect.TypeOf(ProviderConnection{}):               {"id": PublicIDPrefixConn, "project_id": PublicIDPrefixProject},
		reflect.TypeOf(SavedView{}):                        {"id": PublicIDPrefixView, "created_by_id": PublicIDPrefixUser},
		reflect.TypeOf(ManagedCompetitor{}):                {"id": PublicIDPrefixComp},
		reflect.TypeOf(Competitor{}):                       {"id": PublicIDPrefixComp},
		reflect.TypeOf(CompetitorColumn{}):                 {"id": PublicIDPrefixComp},
		reflect.TypeOf(CompetitorShare{}):                  {"id": PublicIDPrefixComp},
		reflect.TypeOf(NotificationPreferences{}):          {"project_id": PublicIDPrefixProject},
		reflect.TypeOf(UpdatedNotificationPreferences{}):   {"project_id": PublicIDPrefixProject},
		reflect.TypeOf(MigrationToken{}):                   {"id": PublicIDPrefixMToken},
		reflect.TypeOf(CloudImportJob{}):                   {"id": PublicIDPrefixJob},
		reflect.TypeOf(IssuedMigrationToken{}):             {"id": PublicIDPrefixMToken},
		reflect.TypeOf(RevokedMigrationToken{}):            {"id": PublicIDPrefixMToken},
		reflect.TypeOf(Signal{}):                           {"id": PublicIDPrefixSignal, "keyword_id": PublicIDPrefixKeyword, "project_id": PublicIDPrefixProject, "public_id": PublicIDPrefixSignal},
		reflect.TypeOf(MeProject{}):                        {"id": PublicIDPrefixProject},
		reflect.TypeOf(Me{}):                               {"id": PublicIDPrefixUser},
		reflect.TypeOf(PersonalAccessToken{}):              {"id": PublicIDPrefixPAT},
		reflect.TypeOf(Webhook{}):                          {"id": PublicIDPrefixWebhook},
		reflect.TypeOf(RankedKeywordConnection{}):          {"id": PublicIDPrefixConn},
		reflect.TypeOf(TeamMemberRoleResult{}):             {"id": PublicIDPrefixMember},
		reflect.TypeOf(TeamMemberMutationResult{}):         {"id": PublicIDPrefixMember},
		reflect.TypeOf(TeamInviteResendResult{}):           {"id": PublicIDPrefixInvite},
		reflect.TypeOf(AnalyticsConnection{}):              {"id": PublicIDPrefixConn},
		reflect.TypeOf(TrafficSyncRun{}):                   {"connection_id": PublicIDPrefixConn},
		reflect.TypeOf(TrafficSyncSummary{}):               {"project_id": PublicIDPrefixProject},
		reflect.TypeOf(CloudImportFinalizeResponse{}):      {"job_id": PublicIDPrefixJob},
		reflect.TypeOf(CloudImportSessionCreateResponse{}): {"session_id": PublicIDPrefixJob},
	}
	if !reflect.DeepEqual(responsePublicIDFieldPrefixes, contracts) {
		t.Fatalf("response public ID contract = %#v, want %#v", responsePublicIDFieldPrefixes, contracts)
	}

	for responseType, fields := range contracts {
		for jsonName, prefix := range fields {
			field, ok := responseFieldByJSONName(responseType, jsonName)
			if !ok {
				t.Fatalf("%s has no %q response field", responseType.Name(), jsonName)
			}
			fieldType := field.Type
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() != reflect.String {
				t.Fatalf("%s.%s is %s, want string public ID", responseType.Name(), field.Name, field.Type)
			}
			if err := ValidatePublicIDPrefix(strictID(prefix), prefix); err != nil {
				t.Fatalf("invalid schema prefix %q for %s.%s: %v", prefix, responseType.Name(), field.Name, err)
			}
		}
	}
}
