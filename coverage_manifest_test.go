package bisibility

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

const skipCoverageOperation = ""

var clientMethodOperationIDs = map[string]string{
	"AddCompetitor":                   "addCompetitor",
	"AddKeywords":                     "addKeywords",
	"BaseURL":                         skipCoverageOperation,
	"BulkUpdateKeywords":              "bulkUpdateKeywords",
	"ConnectProvider":                 "connectProvider",
	"CreateAPIKey":                    "createApiKey",
	"CreateCloudImportSession":        "createCloudImportSession",
	"CreateAlertRule":                 "createAlertRule",
	"CreateKeywords":                  "addKeywords",
	"CreateMyToken":                   "createPersonalAccessToken",
	"CreateProject":                   "createProject",
	"CreateProjectAPIKey":             "createProjectApiKey",
	"CreateSavedView":                 "createSavedView",
	"CreateSignal":                    "createSignal",
	"CreateTeamInvite":                "createTeamInvite",
	"CreateWebhook":                   "createWebhookEndpoint",
	"DeleteAlertRule":                 "deleteAlertRule",
	"DeleteKeyword":                   "deleteKeyword",
	"DeleteProject":                   "deleteProject",
	"DeleteProjectSavedView":          "deleteProjectSavedView",
	"DeleteSavedView":                 "deleteSavedView",
	"DeleteWebhook":                   "deleteWebhookEndpoint",
	"DisconnectProvider":              "disconnectProvider",
	"ExportRankHistory":               "exportRankHistory",
	"FinalizeCloudImportSession":      "finalizeCloudImportSession",
	"GetCapabilities":                 "getCapabilities",
	"GetCloudImportCompatibility":     "getCloudImportCompatibility",
	"GetCostEstimate":                 "getCostEstimate",
	"GetHealth":                       "getHealth",
	"GetKeyword":                      "getKeyword",
	"GetKeywordMetrics":               "getKeywordMetrics",
	"GetLLMSText":                     "getLlmsTxt",
	"GetMe":                           "getMe",
	"GetNotificationPreferences":      "getNotificationPreferences",
	"GetOpenAPI":                      "getOpenApi",
	"GetProject":                      "getProject",
	"GetProviderRates":                "getProviderRates",
	"GetRankCheckResult":              "getRankCheckResult",
	"ImportCloudExport":               "importCloudExport",
	"IterateAPIKeys":                  skipCoverageOperation,
	"IterateAlertRules":               skipCoverageOperation,
	"IterateCompetitors":              skipCoverageOperation,
	"IterateKeywords":                 skipCoverageOperation,
	"IterateMigrationTokens":          skipCoverageOperation,
	"IterateProjectAPIKeys":           skipCoverageOperation,
	"IterateProviders":                skipCoverageOperation,
	"IterateRankChecks":               skipCoverageOperation,
	"IterateRankHistory":              skipCoverageOperation,
	"IterateSavedViews":               skipCoverageOperation,
	"IterateSignals":                  skipCoverageOperation,
	"IterateTeamInvites":              skipCoverageOperation,
	"IterateTeamMembers":              skipCoverageOperation,
	"IterateTriggeredAlerts":          skipCoverageOperation,
	"IterateWebhooks":                 skipCoverageOperation,
	"KeywordsCreate":                  skipCoverageOperation,
	"KeywordsList":                    skipCoverageOperation,
	"ListAPIKeys":                     "listApiKeys",
	"ListAlertRules":                  "listAlertRules",
	"ListCompetitors":                 "listCompetitors",
	"ListKeywords":                    "listKeywords",
	"ListMigrationTokens":             "listMigrationTokens",
	"ListMyTokens":                    "listPersonalAccessTokens",
	"ListProjectAPIKeys":              "listProjectApiKeys",
	"ListProjects":                    "listProjects",
	"ListProviders":                   "listProviders",
	"ListRankChecks":                  "listRankChecks",
	"ListRankedKeywordSuggestions":    "listRankedKeywordSuggestions",
	"ListSavedViews":                  "listSavedViews",
	"ListSearchPerformanceQueryStats": "listSearchPerformanceQueryStats",
	"ListSignals":                     "listSignals",
	"ListSitemapMonitors":             "listSitemapMonitors",
	"ListTeamInvites":                 "listTeamInvites",
	"ListTeamMembers":                 "listTeamMembers",
	"ListTrafficSnapshots":            "listTrafficSnapshots",
	"ListTriggeredAlerts":             "listTriggeredAlerts",
	"ListWebhooks":                    "listWebhookEndpoints",
	"MarkProjectAlertsRead":           "markProjectAlertsRead",
	"MintMigrationToken":              "mintMigrationToken",
	"MuteTriggeredAlert":              "muteTriggeredAlert",
	"Projects":                        skipCoverageOperation,
	"RankHistory":                     skipCoverageOperation,
	"RemoveCompetitor":                "removeCompetitor",
	"RemoveProjectCompetitor":         "removeProjectCompetitor",
	"RemoveTeamMember":                "removeTeamMember",
	"ResearchKeywords":                "researchKeywords",
	"ResendTeamInvite":                "resendTeamInvite",
	"RevokeAPIKey":                    "revokeApiKey",
	"RevokeMigrationToken":            "revokeMigrationToken",
	"RevokeMyToken":                   "revokePersonalAccessToken",
	"RevokeProjectMigrationToken":     "revokeProjectMigrationToken",
	"RevokeProjectTeamInvite":         "revokeProjectTeamInvite",
	"RevokeTeamInvite":                "revokeTeamInvite",
	"RunCheck":                        skipCoverageOperation,
	"RunRankCheck":                    "runRankCheck",
	"SearchLocations":                 "searchLocations",
	"SetKeywordTargetURL":             skipCoverageOperation,
	"SetPrimaryProvider":              skipCoverageOperation,
	"SetProviderEnabled":              skipCoverageOperation,
	"SetProviderPriority":             skipCoverageOperation,
	"SyncProjectTraffic":              "syncProjectTraffic",
	"TestProviderConnection":          "testProviderConnection",
	"UpdateAlertRule":                 "updateAlertRule",
	"UpdateKeyword":                   "setKeywordTargetUrl",
	"UpdateMe":                        "updateMe",
	"UpdateNotificationPreferences":   "updateNotificationPreferences",
	"UpdateProject":                   "updateProject",
	"UpdateProjectDefaults":           "updateProjectDefaults",
	"UpdateProviderSettings":          "updateProviderSettings",
	"UpdateSitemapMonitor":            "updateSitemapMonitor",
	"UpdateTeamMemberRole":            "updateTeamMemberRole",
	"UpdateWebhook":                   "updateWebhookEndpoint",
	"UploadCloudImportChunk":          "uploadCloudImportChunk",
	"UploadCloudImportChunkRaw":       skipCoverageOperation,
}

func TestCoverageManifest(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "method map covers exported client methods",
			run: func(t *testing.T) {
				assertStringSlicesEqual(t, reflectedClientMethods(), sortedMapKeys(clientMethodOperationIDs))
			},
		},
		{
			name: "manifest matches method operation map",
			run: func(t *testing.T) {
				assertStringSlicesEqual(t, readCoverageManifest(t), expectedCoverageOperations())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func reflectedClientMethods() []string {
	clientType := reflect.TypeOf((*Client)(nil))
	methods := make([]string, 0, clientType.NumMethod())
	for i := 0; i < clientType.NumMethod(); i++ {
		methods = append(methods, clientType.Method(i).Name)
	}
	sort.Strings(methods)
	return methods
}

func readCoverageManifest(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile("coverage/operations.json")
	if err != nil {
		t.Fatalf("read coverage manifest: %v", err)
	}

	var operations []string
	if err := json.Unmarshal(data, &operations); err != nil {
		t.Fatalf("parse coverage manifest: %v", err)
	}
	return operations
}

func expectedCoverageOperations() []string {
	seen := make(map[string]bool)
	operations := make([]string, 0, len(clientMethodOperationIDs))
	for _, operationID := range clientMethodOperationIDs {
		if operationID == skipCoverageOperation || seen[operationID] {
			continue
		}
		seen[operationID] = true
		operations = append(operations, operationID)
	}
	sort.Strings(operations)
	return operations
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertStringSlicesEqual(t *testing.T, got, want []string) {
	t.Helper()

	if reflect.DeepEqual(got, want) {
		return
	}
	t.Fatalf("string slices differ\ngot:  %#v\nwant: %#v", got, want)
}
