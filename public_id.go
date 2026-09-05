package bisibility

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// PublicIDPrefix identifies the public resource namespace encoded in an ID.
// A public ID is always prefix_[a-z][a-z0-9]{23}.
type PublicIDPrefix string

const (
	PublicIDPrefixAlert   PublicIDPrefix = "al"
	PublicIDPrefixRule    PublicIDPrefix = "alr"
	PublicIDPrefixAudit   PublicIDPrefix = "audit"
	PublicIDPrefixCheck   PublicIDPrefix = "check"
	PublicIDPrefixComp    PublicIDPrefix = "cmp"
	PublicIDPrefixConn    PublicIDPrefix = "conn"
	PublicIDPrefixHook    PublicIDPrefix = "dwh"
	PublicIDPrefixMToken  PublicIDPrefix = "ferry"
	PublicIDPrefixJob     PublicIDPrefix = "imp"
	PublicIDPrefixInvite  PublicIDPrefix = "inv"
	PublicIDPrefixKey     PublicIDPrefix = "key"
	PublicIDPrefixKeyword PublicIDPrefix = "kw"
	PublicIDPrefixMember  PublicIDPrefix = "mbr"
	PublicIDPrefixNotif   PublicIDPrefix = "ntf"
	PublicIDPrefixPAT     PublicIDPrefix = "pat"
	PublicIDPrefixProject PublicIDPrefix = "prj"
	PublicIDPrefixRun     PublicIDPrefix = "rcr"
	PublicIDPrefixSession PublicIDPrefix = "sid"
	PublicIDPrefixSignal  PublicIDPrefix = "sig"
	PublicIDPrefixSKW     PublicIDPrefix = "svkw"
	PublicIDPrefixTag     PublicIDPrefix = "tag"
	PublicIDPrefixUser    PublicIDPrefix = "usr"
	PublicIDPrefixView    PublicIDPrefix = "viw"
	PublicIDPrefixWebhook PublicIDPrefix = "we"
)

var publicIDPrefixes = map[PublicIDPrefix]struct{}{
	PublicIDPrefixAlert: {}, PublicIDPrefixAudit: {}, PublicIDPrefixCheck: {},
	PublicIDPrefixComp: {}, PublicIDPrefixConn: {}, PublicIDPrefixHook: {},
	PublicIDPrefixInvite: {}, PublicIDPrefixJob: {}, PublicIDPrefixKey: {},
	PublicIDPrefixKeyword: {}, PublicIDPrefixMember: {}, PublicIDPrefixMToken: {},
	PublicIDPrefixNotif: {}, PublicIDPrefixPAT: {}, PublicIDPrefixProject: {},
	PublicIDPrefixRule: {}, PublicIDPrefixRun: {}, PublicIDPrefixSession: {},
	PublicIDPrefixSignal: {},
	PublicIDPrefixSKW:    {}, PublicIDPrefixTag: {}, PublicIDPrefixUser: {},
	PublicIDPrefixView: {}, PublicIDPrefixWebhook: {},
}

var publicIDSuffixPattern = regexp.MustCompile(`^[a-z][a-z0-9]{23}$`)
var cloudImportChecksumPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

const (
	cloudImportDeviceValidationError   = "%s.device must be desktop or mobile"
	cloudImportLocationValidationError = "%s.location must be a supported market"
)

// IsPublicID reports whether value is a strict Bisibility public ID from the
// canonical registry. It intentionally rejects legacy IDs and raw database IDs.
func IsPublicID(value string) bool {
	prefix, suffix, ok := strings.Cut(value, "_")
	if !ok || prefix == "" || !publicIDSuffixPattern.MatchString(suffix) {
		return false
	}
	_, ok = publicIDPrefixes[PublicIDPrefix(prefix)]
	return ok
}

// ValidatePublicID validates a strict public ID from any registered namespace.
func ValidatePublicID(value string) error {
	if !IsPublicID(value) {
		return fmt.Errorf("must be a strict Bisibility public ID (prefix_[a-z][a-z0-9]{23})")
	}
	return nil
}

// ValidatePublicIDPrefix validates a strict public ID in one resource namespace.
func ValidatePublicIDPrefix(value string, prefix PublicIDPrefix) error {
	if _, ok := publicIDPrefixes[prefix]; !ok {
		return fmt.Errorf("unsupported public ID prefix %q", prefix)
	}
	if !strings.HasPrefix(value, string(prefix)+"_") || !publicIDSuffixPattern.MatchString(strings.TrimPrefix(value, string(prefix)+"_")) {
		return fmt.Errorf("must be a strict %s public ID", prefix)
	}
	return nil
}

func requirePublicID(value, name string, prefix PublicIDPrefix) error {
	if err := ValidatePublicIDPrefix(value, prefix); err != nil {
		return &ConfigurationError{Message: fmt.Sprintf("%s %s.", name, err)}
	}
	return nil
}

func validateRequestIdentifiers(routePath string, query url.Values, body any) error {
	if err := validateRoutePublicIDs(routePath); err != nil {
		return err
	}
	if err := validateQueryPublicIDs(query); err != nil {
		return err
	}
	return validateBodyPublicIDs(body)
}

func validateProjectHeader(headers http.Header) error {
	for _, value := range headers.Values(projectHeader) {
		if err := requirePublicID(value, projectHeader, PublicIDPrefixProject); err != nil {
			return err
		}
	}
	return nil
}

func validateRoutePublicIDs(routePath string) error {
	parts := strings.Split(strings.Trim(routePath, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return nil
	}
	handled, err := validateTopLevelRoutePublicID(parts)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	return validateProjectRoutePublicIDs(parts)
}

func validateTopLevelRoutePublicID(parts []string) (bool, error) {
	switch parts[0] {
	case "projects":
		return false, validateRoutePart(parts, 1, "projectID", PublicIDPrefixProject)
	case "keywords":
		if len(parts) > 1 && parts[1] != "bulk" {
			return true, validateRoutePart(parts, 1, "keywordID", PublicIDPrefixKeyword)
		}
	case "api-keys":
		return true, validateRoutePart(parts, 1, "keyID", PublicIDPrefixKey)
	case "rank-checks":
		return true, validateRoutePart(parts, 1, "checkID", PublicIDPrefixCheck)
	case "alert-rules":
		return true, validateRoutePart(parts, 1, "ruleID", PublicIDPrefixRule)
	case "team":
		if len(parts) > 1 && parts[1] == "invites" {
			return true, validateRoutePart(parts, 2, "inviteID", PublicIDPrefixInvite)
		}
	case "saved-views":
		return true, validateRoutePart(parts, 1, "viewID", PublicIDPrefixView)
	case "competitors":
		return true, validateRoutePart(parts, 1, "competitorID", PublicIDPrefixComp)
	case "migration-tokens":
		return true, validateRoutePart(parts, 1, "tokenID", PublicIDPrefixMToken)
	case "me":
		if len(parts) > 2 && parts[1] == "tokens" && parts[2] != "current" {
			return true, validateRoutePart(parts, 2, "tokenID", PublicIDPrefixPAT)
		}
	case "cloud":
		if len(parts) > 3 && parts[1] == "import" && parts[2] == "sessions" {
			return true, validateRoutePart(parts, 3, "sessionID", PublicIDPrefixJob)
		}
	}
	return false, nil
}

func validateProjectRoutePublicIDs(parts []string) error {
	if parts[0] != "projects" || len(parts) < 3 {
		return nil
	}
	for index, part := range parts[2:] {
		position := index + 2
		handled, err := validateProjectRoutePart(parts, position, part)
		if err != nil {
			return err
		}
		if handled {
			return nil
		}
	}
	return nil
}

func validateProjectRoutePart(parts []string, position int, part string) (bool, error) {
	if position+1 >= len(parts) {
		return false, nil
	}
	switch part {
	case "webhooks":
		return true, validateRoutePart(parts, position+1, "webhookID", PublicIDPrefixWebhook)
	case "triggered-alerts":
		if parts[position+1] != "mark-read" {
			return true, validateRoutePart(parts, position+1, "alertID", PublicIDPrefixAlert)
		}
	case "sitemap-monitors":
		return true, validateRoutePart(parts, position+1, "monitorID", PublicIDPrefixProject)
	case "saved-keywords":
		return true, validateRoutePart(parts, position+1, "savedKeywordID", PublicIDPrefixSKW)
	case "saved-views":
		return true, validateRoutePart(parts, position+1, "viewID", PublicIDPrefixView)
	case "competitors":
		return true, validateRoutePart(parts, position+1, "competitorID", PublicIDPrefixComp)
	case "migration-tokens":
		return true, validateRoutePart(parts, position+1, "tokenID", PublicIDPrefixMToken)
	case "members":
		if position > 2 && parts[position-1] == "team" {
			return true, validateRoutePart(parts, position+1, "memberID", PublicIDPrefixMember)
		}
	case "invites":
		if position > 2 && parts[position-1] == "team" {
			return true, validateRoutePart(parts, position+1, "inviteID", PublicIDPrefixInvite)
		}
	}
	return false, nil
}

func validateRoutePart(parts []string, index int, name string, prefix PublicIDPrefix) error {
	if index >= len(parts) {
		return nil
	}
	value, err := url.PathUnescape(parts[index])
	if err != nil {
		return &ConfigurationError{Message: fmt.Sprintf("%s is not a valid URL path value.", name)}
	}
	return requirePublicID(value, name, prefix)
}

func validateQueryPublicIDs(query url.Values) error {
	for name, prefix := range map[string]PublicIDPrefix{
		"connection_id": PublicIDPrefixConn,
		"keyword_id":    PublicIDPrefixKeyword,
	} {
		for _, value := range query[name] {
			if value == "" {
				continue
			}
			if err := requirePublicID(value, name, prefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateBodyPublicIDs(body any) error {
	if body == nil {
		return nil
	}
	return validateBodyValue(reflect.ValueOf(body), "body")
}

func validateBodyValue(value reflect.Value, path string) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		return validateBodyReference(value, path)
	}
	handled, err := validateKnownBodyValue(value, path)
	if handled || err != nil {
		return err
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		return validateBodyValues(value, path)
	case reflect.Struct:
		return validateBodyStructFields(value, path)
	default:
		return nil
	}
}

func validateBodyReference(value reflect.Value, path string) error {
	if value.IsNil() {
		return nil
	}
	return validateBodyValue(value.Elem(), path)
}

func validateKnownBodyValue(value reflect.Value, path string) (bool, error) {
	switch input := value.Interface().(type) {
	case KeywordBulkInput:
		return true, validateIDs(input.KeywordIDs, "body.keyword_ids", PublicIDPrefixKeyword)
	case CreateAlertRuleInput:
		if err := validateAlertRuleTargetIDs(input.TargetType, input.TargetIDs, "body.target_ids"); err != nil {
			return true, err
		}
		return true, validateIDs(input.RecipientIDs, "body.recipient_ids", PublicIDPrefixUser)
	case CloudImportPackage:
		return true, validateCloudImportPackage(input, path)
	case CloudImportSessionCreate:
		return true, validateCloudImportSessionCreate(input, path)
	case CloudImportKeywordsChunk:
		return true, validateCloudImportKeywordsChunk(input, path)
	case CloudImportSectionsChunk:
		return true, validateCloudImportSectionsChunk(input, path)
	case CloudImportKeyword:
		return true, validateCloudImportKeyword(input, path)
	case CloudImportAlertRule:
		return true, validateCloudImportAlertRule(input, path)
	case CloudImportKeywordAlertTarget:
		return true, validateCloudImportKeywordAlertTarget(input, path)
	case CloudImportTagAlertTarget:
		return true, validateCloudImportTagAlertTarget(input, path)
	}
	return false, nil
}

func validateBodyValues(value reflect.Value, path string) error {
	for index := 0; index < value.Len(); index++ {
		if err := validateBodyValue(value.Index(index), fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateBodyStructFields(value reflect.Value, path string) error {
	typeOfValue := value.Type()
	for index := 0; index < value.NumField(); index++ {
		field := typeOfValue.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if err := validateBodyStructField(value.Field(index), field, path); err != nil {
			return err
		}
	}
	return nil
}

func validateBodyStructField(value reflect.Value, field reflect.StructField, path string) error {
	jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
	if jsonName == "" || jsonName == "-" {
		return validateBodyValue(value, path+"."+field.Name)
	}
	if prefix, ok := bodyIDFieldPrefixes[jsonName]; ok {
		return validateBodyFieldIDs(value, path+"."+jsonName, prefix)
	}
	return validateBodyValue(value, path+"."+jsonName)
}

var bodyIDFieldPrefixes = map[string]PublicIDPrefix{
	"connection_id": PublicIDPrefixConn,
	"keyword_id":    PublicIDPrefixKeyword,
	"keyword_ids":   PublicIDPrefixKeyword,
	"project_id":    PublicIDPrefixProject,
	"recipient_ids": PublicIDPrefixUser,
	"rule_id":       PublicIDPrefixRule,
	"tag_id":        PublicIDPrefixTag,
}

// responsePublicIDFieldPrefixes records every SDK response field that carries
// a public resource ID. Natural provider IDs and location keys are omitted.
var responsePublicIDFieldPrefixes = map[reflect.Type]map[string]PublicIDPrefix{
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
	reflect.TypeOf(SavedKeyword{}):                     {"id": PublicIDPrefixSKW},
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

func validateResponsePublicIDs(response any) error {
	return validateResponseValue(reflect.ValueOf(response), "response")
}

func validateResponseValue(value reflect.Value, path string) error {
	if !value.IsValid() {
		return nil
	}
	value, present := unwrapResponseValue(value)
	if !present {
		return nil
	}
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		return validateResponseValues(value, path)
	case reflect.Struct:
		return validateResponseStruct(value, path)
	default:
		return nil
	}
}

func unwrapResponseValue(value reflect.Value) (reflect.Value, bool) {
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	return value, true
}

func validateResponseValues(value reflect.Value, path string) error {
	for index := 0; index < value.Len(); index++ {
		if err := validateResponseValue(value.Index(index), fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateResponseStruct(value reflect.Value, path string) error {
	if err := validateRegisteredResponseFields(value, path); err != nil {
		return err
	}
	if value.Type() == reflect.TypeOf(AlertRule{}) {
		if err := validateAlertRuleResponseIDs(value.Interface().(AlertRule), path); err != nil {
			return err
		}
	}
	return validateResponseStructFields(value, path)
}

func validateRegisteredResponseFields(value reflect.Value, path string) error {
	if prefixes, ok := responsePublicIDFieldPrefixes[value.Type()]; ok {
		for jsonName, prefix := range prefixes {
			field, found := responseFieldByJSONName(value.Type(), jsonName)
			if !found {
				return fmt.Errorf("%s has no %s field", value.Type().Name(), jsonName)
			}
			fieldValue := value.FieldByIndex(field.Index)
			if err := validateResponseFieldID(fieldValue, field, path+"."+jsonName, prefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateResponseStructFields(value reflect.Value, path string) error {
	for index := 0; index < value.NumField(); index++ {
		field := value.Type().Field(index)
		if field.PkgPath != "" {
			continue
		}
		if err := validateResponseValue(value.Field(index), path+"."+field.Name); err != nil {
			return err
		}
	}
	return nil
}

func validateResponseFieldID(value reflect.Value, field reflect.StructField, name string, prefix PublicIDPrefix) error {
	value, presentPointer, skipped, err := unwrapResponseFieldIDValue(value, name)
	if err != nil {
		return err
	}
	if skipped {
		return nil
	}
	if !value.IsValid() || value.Kind() != reflect.String {
		return fmt.Errorf("%s must be a string public ID", name)
	}
	if value.String() == "" {
		return validateEmptyResponseFieldID(field, name, presentPointer)
	}
	if err := ValidatePublicIDPrefix(value.String(), prefix); err != nil {
		return fmt.Errorf("%s %w", name, err)
	}
	return nil
}

func unwrapResponseFieldIDValue(value reflect.Value, name string) (reflect.Value, bool, bool, error) {
	presentPointer := false
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			if value.Kind() == reflect.Pointer {
				return value, presentPointer, true, nil
			}
			return reflect.Value{}, presentPointer, false, fmt.Errorf("%s is required", name)
		}
		if value.Kind() == reflect.Pointer {
			presentPointer = true
		}
		value = value.Elem()
	}
	return value, presentPointer, false, nil
}

func validateEmptyResponseFieldID(field reflect.StructField, name string, presentPointer bool) error {
	if !presentPointer && strings.Contains(field.Tag.Get("json"), ",omitempty") {
		return nil
	}
	return fmt.Errorf("%s is required", name)
}

func responseFieldByJSONName(responseType reflect.Type, jsonName string) (reflect.StructField, bool) {
	for index := 0; index < responseType.NumField(); index++ {
		field := responseType.Field(index)
		if field.Anonymous {
			nestedType := field.Type
			if nestedType.Kind() == reflect.Pointer {
				nestedType = nestedType.Elem()
			}
			if nestedType.Kind() == reflect.Struct {
				if nested, ok := responseFieldByJSONName(nestedType, jsonName); ok {
					nested.Index = append([]int{index}, nested.Index...)
					return nested, true
				}
			}
			continue
		}
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag == jsonName {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func validateBodyFieldIDs(value reflect.Value, name string, prefix PublicIDPrefix) error {
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.String:
		if value.String() == "" {
			return nil
		}
		return requirePublicID(value.String(), name, prefix)
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if err := validateBodyFieldIDs(value.Index(index), fmt.Sprintf("%s[%d]", name, index), prefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateIDs(values []string, name string, prefix PublicIDPrefix) error {
	for index, value := range values {
		if err := requirePublicID(value, fmt.Sprintf("%s[%d]", name, index), prefix); err != nil {
			return err
		}
	}
	return nil
}

func validateAlertRuleTargetIDs(targetType AlertTargetType, values []string, name string) error {
	if len(values) == 0 {
		return nil
	}
	switch targetType {
	case AlertTargetTypeKeyword:
		return validateIDs(values, name, PublicIDPrefixKeyword)
	case AlertTargetTypeTag:
		return validateIDs(values, name, PublicIDPrefixTag)
	default:
		return &ConfigurationError{Message: fmt.Sprintf("%s requires TargetType keyword or tag.", name)}
	}
}

func validateAlertRuleResponseIDs(rule AlertRule, path string) error {
	if err := validateAlertRuleTargetIDs(rule.TargetType, rule.TargetIDs, path+".target_ids"); err != nil {
		return err
	}
	return validateIDs(rule.RecipientIDs, path+".recipient_ids", PublicIDPrefixUser)
}

func validateCloudImportPackage(input CloudImportPackage, path string) error {
	if err := validateCloudImportPackageFields(input, path); err != nil {
		return err
	}
	return validateCloudImportPackageItems(input, path)
}

func validateCloudImportPackageFields(input CloudImportPackage, path string) error {
	if err := requirePublicID(input.ProjectID, path+".project_id", PublicIDPrefixProject); err != nil {
		return err
	}
	if err := validateCloudImportRequiredSlice(input.Keywords, path+".keywords", 500); err != nil {
		return err
	}
	if err := validateCloudImportRequiredSlice(input.AlertRules, path+".alert_rules", 500); err != nil {
		return err
	}
	if err := validateCloudImportRequiredSlice(input.Competitors, path+".competitors", 500); err != nil {
		return err
	}
	if err := validateCloudImportRequiredSlice(input.NotificationPreferences, path+".notification_preferences", 50); err != nil {
		return err
	}
	if err := validateCloudImportRequiredSlice(input.SavedViews, path+".saved_views", 500); err != nil {
		return err
	}
	if input.Scope != "" && !isCloudImportScope(input.Scope) {
		return cloudImportConfigurationError("%s.scope must be current or history", path)
	}
	return nil
}

func validateCloudImportPackageItems(input CloudImportPackage, path string) error {
	if err := validateCloudImportPackageKeywords(input.Keywords, path+".keywords"); err != nil {
		return err
	}
	if err := validateCloudImportPackageAlertRules(input.AlertRules, path+".alert_rules"); err != nil {
		return err
	}
	if err := validateCloudImportPackageCompetitors(input.Competitors, path+".competitors"); err != nil {
		return err
	}
	return validateCloudImportPackageSavedViews(input.SavedViews, path+".saved_views")
}

func validateCloudImportPackageKeywords(keywords []CloudImportKeyword, path string) error {
	for index, keyword := range keywords {
		if err := validateCloudImportKeyword(keyword, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportPackageAlertRules(rules []CloudImportAlertRule, path string) error {
	for index, rule := range rules {
		if err := validateCloudImportAlertRule(rule, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportPackageCompetitors(competitors []CloudImportCompetitor, path string) error {
	for index, competitor := range competitors {
		if err := validateCloudImportCompetitor(competitor, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportPackageSavedViews(savedViews []CloudImportSavedView, path string) error {
	for index, savedView := range savedViews {
		if err := validateCloudImportSavedView(savedView, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportCompatibility(input CloudImportCompatibility) error {
	if input.SchemaVersionsSupported == nil {
		return cloudImportConfigurationError("CloudImportCompatibility.schema_versions_supported is required")
	}
	for index, version := range input.SchemaVersionsSupported {
		if version != CloudImportProtocolVersion {
			return cloudImportConfigurationError("CloudImportCompatibility.schema_versions_supported[%d] must be %d", index, CloudImportProtocolVersion)
		}
	}
	return nil
}

func validateCloudImportFinalizeResponse(input CloudImportFinalizeResponse) error {
	if input.Counts == nil {
		return cloudImportConfigurationError("CloudImportFinalizeResponse.counts is required")
	}
	if err := requirePublicID(input.JobID, "CloudImportFinalizeResponse.job_id", PublicIDPrefixJob); err != nil {
		return err
	}
	if input.State != CloudImportStateDone {
		return cloudImportConfigurationError("CloudImportFinalizeResponse.state must be done")
	}
	return nil
}

func validateCloudImportKeyword(input CloudImportKeyword, path string) error {
	if err := requirePublicID(input.ID, path+".id", PublicIDPrefixKeyword); err != nil {
		return err
	}
	if !hasCloudImportText(input.Keyword, 180) {
		return cloudImportConfigurationError("%s.keyword must contain 1 to 180 characters", path)
	}
	if !isCloudImportDevice(input.Device) {
		return cloudImportConfigurationError(cloudImportDeviceValidationError, path)
	}
	if !isCloudImportLocation(input.Location) {
		return cloudImportConfigurationError(cloudImportLocationValidationError, path)
	}
	if len(input.RankingHistory) > 5000 {
		return cloudImportConfigurationError("%s.rankingHistory must contain at most 5000 rows", path)
	}
	if len(input.Tags) > 12 {
		return cloudImportConfigurationError("%s.tags must contain at most 12 values", path)
	}
	for index, ranking := range input.RankingHistory {
		if err := validateCloudImportRankingHistory(ranking, fmt.Sprintf("%s.rankingHistory[%d]", path, index)); err != nil {
			return err
		}
	}
	for index, tag := range input.Tags {
		if !hasCloudImportText(tag, 48) {
			return cloudImportConfigurationError("%s.tags[%d] must contain 1 to 48 characters", path, index)
		}
	}
	if input.TargetURL != nil && len(*input.TargetURL) > 500 {
		return cloudImportConfigurationError("%s.target_url must contain at most 500 characters", path)
	}
	return nil
}

func validateCloudImportRankingHistory(input CloudImportRankingHistory, path string) error {
	if input.CheckedAt.IsZero() {
		return cloudImportConfigurationError("%s.checkedAt is required", path)
	}
	if err := validateCloudImportPositivePointer(input.Position, path+".position"); err != nil {
		return err
	}
	if err := validateCloudImportPositivePointer(input.PreviousPosition, path+".previousPosition"); err != nil {
		return err
	}
	if input.RankingURL != nil && len(*input.RankingURL) > 500 {
		return cloudImportConfigurationError("%s.rankingUrl must contain at most 500 characters", path)
	}
	return nil
}

func validateCloudImportAlertRule(input CloudImportAlertRule, path string) error {
	if err := validateCloudImportAlertRuleFields(input, path); err != nil {
		return err
	}
	if err := validateCloudImportAlertRuleChannels(input, path); err != nil {
		return err
	}
	if err := validateCloudImportAlertRulePositions(input, path); err != nil {
		return err
	}
	return validateCloudImportAlertRuleTargets(input, path)
}

func validateCloudImportAlertRuleFields(input CloudImportAlertRule, path string) error {
	if err := requirePublicID(input.ID, path+".id", PublicIDPrefixRule); err != nil {
		return err
	}
	if !hasCloudImportText(input.Name, 120) {
		return cloudImportConfigurationError("%s.name must contain 1 to 120 characters", path)
	}
	if input.ConditionType != "" && !isCloudImportAlertCondition(input.ConditionType) {
		return cloudImportConfigurationError("%s.condition_type is unsupported", path)
	}
	if input.TargetType != "" && input.TargetType != AlertTargetTypeAll && input.TargetType != AlertTargetTypeKeyword && input.TargetType != AlertTargetTypeTag {
		return cloudImportConfigurationError("%s.target_type must be all, keyword, or tag", path)
	}
	if len(input.Targets) > 1000 {
		return cloudImportConfigurationError("%s.targets must contain at most 1000 values", path)
	}
	return nil
}

func validateCloudImportAlertRuleChannels(input CloudImportAlertRule, path string) error {
	for index, channel := range input.Channels {
		if channel != AlertChannelEmail && channel != AlertChannelSlack && channel != AlertChannelWebhook {
			return cloudImportConfigurationError("%s.channels[%d] is unsupported", path, index)
		}
	}
	return nil
}

func validateCloudImportAlertRulePositions(input CloudImportAlertRule, path string) error {
	for _, position := range []struct {
		value *int
		name  string
	}{
		{input.DropPositions, "drop_positions"},
		{input.ThresholdPosition, "threshold_position"},
		{input.TopN, "top_n"},
	} {
		if err := validateCloudImportPositivePointer(position.value, path+"."+position.name); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportAlertRuleTargets(input CloudImportAlertRule, path string) error {
	for index, target := range input.Targets {
		if err := validateCloudImportAlertRuleTarget(target, fmt.Sprintf("%s.targets[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportAlertRuleTarget(target CloudImportAlertRuleTarget, path string) error {
	switch input := target.(type) {
	case CloudImportKeywordAlertTarget:
		return validateCloudImportKeywordAlertTarget(input, path)
	case *CloudImportKeywordAlertTarget:
		if input == nil {
			return cloudImportConfigurationError("%s must not be nil", path)
		}
		return validateCloudImportKeywordAlertTarget(*input, path)
	case CloudImportTagAlertTarget:
		return validateCloudImportTagAlertTarget(input, path)
	case *CloudImportTagAlertTarget:
		if input == nil {
			return cloudImportConfigurationError("%s must not be nil", path)
		}
		return validateCloudImportTagAlertTarget(*input, path)
	case nil:
		return cloudImportConfigurationError("%s is required", path)
	default:
		return cloudImportConfigurationError("%s must be a keyword or tag alert target", path)
	}
}

func validateCloudImportKeywordAlertTarget(input CloudImportKeywordAlertTarget, path string) error {
	if err := requirePublicID(input.KeywordID, path+".keyword_id", PublicIDPrefixKeyword); err != nil {
		return err
	}
	if input.Device != "" && !isCloudImportDevice(input.Device) {
		return cloudImportConfigurationError(cloudImportDeviceValidationError, path)
	}
	if input.Keyword != "" && !hasCloudImportText(input.Keyword, 180) {
		return cloudImportConfigurationError("%s.keyword must contain 1 to 180 characters", path)
	}
	if input.Location != "" && !isCloudImportLocation(input.Location) {
		return cloudImportConfigurationError(cloudImportLocationValidationError, path)
	}
	return nil
}

func validateCloudImportTagAlertTarget(input CloudImportTagAlertTarget, path string) error {
	if !hasCloudImportText(input.Tag, 80) {
		return cloudImportConfigurationError("%s.tag must contain 1 to 80 characters", path)
	}
	return nil
}

func validateCloudImportCompetitor(input CloudImportCompetitor, path string) error {
	if err := requirePublicID(input.ID, path+".id", PublicIDPrefixComp); err != nil {
		return err
	}
	if !isCloudImportDomain(input.Domain) {
		return cloudImportConfigurationError("%s.domain must be a canonical lowercase hostname", path)
	}
	if input.Label != nil && len(*input.Label) > 80 {
		return cloudImportConfigurationError("%s.label must contain at most 80 characters", path)
	}
	return nil
}

func validateCloudImportSavedView(input CloudImportSavedView, path string) error {
	if err := requirePublicID(input.ID, path+".id", PublicIDPrefixView); err != nil {
		return err
	}
	if !hasCloudImportText(input.Name, 120) {
		return cloudImportConfigurationError("%s.name must contain 1 to 120 characters", path)
	}
	if input.Surface != "" && !isCloudImportSavedViewSurface(input.Surface) {
		return cloudImportConfigurationError("%s.surface must be keywords or competitors", path)
	}
	return nil
}

func validateCloudImportSessionTotals(input CloudImportSessionTotals, path string) error {
	if input.Keywords < 0 || input.RankChecks < 0 {
		return cloudImportConfigurationError("%s values cannot be negative", path)
	}
	return nil
}

func validateCloudImportSessionCreate(input CloudImportSessionCreate, path string) error {
	if input.ChunkCount < 1 || input.ChunkCount > 500 {
		return cloudImportConfigurationError("%s.chunk_count must be between 1 and 500", path)
	}
	if err := requirePublicID(input.SourceProjectID, path+".source_project_id", PublicIDPrefixProject); err != nil {
		return err
	}
	if input.Totals != nil {
		return validateCloudImportSessionTotals(*input.Totals, path+".totals")
	}
	return nil
}

func validateCloudImportChunkLimits(input CloudImportChunkLimits, path string) error {
	if input.MaxBodyBytes < 1 || input.MaxHistoryRows < 1 || input.MaxKeywords < 1 {
		return cloudImportConfigurationError("%s values must be positive", path)
	}
	return nil
}

func validateCloudImportSessionCreateResponse(input CloudImportSessionCreateResponse) error {
	if err := requirePublicID(input.SessionID, "CloudImportSessionCreateResponse.session_id", PublicIDPrefixJob); err != nil {
		return err
	}
	if input.State != CloudImportStateReceiving {
		return cloudImportConfigurationError("CloudImportSessionCreateResponse.state must be receiving")
	}
	return validateCloudImportChunkLimits(input.ChunkLimits, "CloudImportSessionCreateResponse.chunk_limits")
}

func validateCloudImportSourceKeyword(input CloudImportSourceKeyword, path string) error {
	if !isCloudImportDevice(input.Device) {
		return cloudImportConfigurationError(cloudImportDeviceValidationError, path)
	}
	if !isCloudImportLocation(input.Location) {
		return cloudImportConfigurationError(cloudImportLocationValidationError, path)
	}
	return nil
}

func validateCloudImportSessionSections(input CloudImportSessionSections, path string) error {
	if err := validateCloudImportSessionSectionLimits(input, path); err != nil {
		return err
	}
	return validateCloudImportSessionSectionItems(input, path)
}

func validateCloudImportSessionSectionLimits(input CloudImportSessionSections, path string) error {
	if err := validateCloudImportOptionalSlice(input.AlertRules, path+".alert_rules", 500); err != nil {
		return err
	}
	if err := validateCloudImportOptionalSlice(input.Competitors, path+".competitors", 500); err != nil {
		return err
	}
	if err := validateCloudImportOptionalSlice(input.NotificationPreferences, path+".notification_preferences", 50); err != nil {
		return err
	}
	if err := validateCloudImportOptionalSlice(input.SavedViews, path+".saved_views", 500); err != nil {
		return err
	}
	return nil
}

func validateCloudImportSessionSectionItems(input CloudImportSessionSections, path string) error {
	if err := validateCloudImportSessionAlertRules(input.AlertRules, path+".alert_rules"); err != nil {
		return err
	}
	if err := validateCloudImportSessionCompetitors(input.Competitors, path+".competitors"); err != nil {
		return err
	}
	if err := validateCloudImportSessionSavedViews(input.SavedViews, path+".saved_views"); err != nil {
		return err
	}
	return validateCloudImportSessionSourceKeywords(input.SourceKeywordIDs, path+".source_keyword_ids")
}

func validateCloudImportSessionAlertRules(rules []CloudImportAlertRule, path string) error {
	for index, rule := range rules {
		if err := validateCloudImportAlertRule(rule, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportSessionCompetitors(competitors []CloudImportCompetitor, path string) error {
	for index, competitor := range competitors {
		if err := validateCloudImportCompetitor(competitor, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportSessionSavedViews(savedViews []CloudImportSavedView, path string) error {
	for index, savedView := range savedViews {
		if err := validateCloudImportSavedView(savedView, fmt.Sprintf("%s[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportSessionSourceKeywords(keywords map[string]CloudImportSourceKeyword, path string) error {
	for sourceID, keyword := range keywords {
		if err := validateCloudImportSourceKeyword(keyword, fmt.Sprintf("%s[%q]", path, sourceID)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportKeywordsChunk(input CloudImportKeywordsChunk, path string) error {
	if !isCloudImportChecksum(input.Checksum) {
		return cloudImportConfigurationError("%s.checksum must be a sha256 digest", path)
	}
	if err := validateCloudImportRequiredSlice(input.Keywords, path+".keywords", 500); err != nil {
		return err
	}
	for index, keyword := range input.Keywords {
		if err := validateCloudImportKeyword(keyword, fmt.Sprintf("%s.keywords[%d]", path, index)); err != nil {
			return err
		}
	}
	return nil
}

func validateCloudImportSectionsChunk(input CloudImportSectionsChunk, path string) error {
	if !isCloudImportChecksum(input.Checksum) {
		return cloudImportConfigurationError("%s.checksum must be a sha256 digest", path)
	}
	return validateCloudImportSessionSections(input.Sections, path+".sections")
}

func validateCloudImportChunkResponse(input CloudImportChunkResponse) error {
	if input.ChunkCount < 1 || input.ChunksReceived < 0 {
		return cloudImportConfigurationError("CloudImportChunkResponse counts are invalid")
	}
	if input.State != CloudImportStateReceiving {
		return cloudImportConfigurationError("CloudImportChunkResponse.state must be receiving")
	}
	return nil
}

func validateCloudImportPositivePointer(value *int, path string) error {
	if value != nil && *value < 1 {
		return cloudImportConfigurationError("%s must be positive", path)
	}
	return nil
}

func validateCloudImportRequiredSlice[T any](value []T, path string, maximum int) error {
	if value == nil {
		return cloudImportConfigurationError("%s is required", path)
	}
	return validateCloudImportOptionalSlice(value, path, maximum)
}

func validateCloudImportOptionalSlice[T any](value []T, path string, maximum int) error {
	if len(value) > maximum {
		return cloudImportConfigurationError("%s must contain at most %d values", path, maximum)
	}
	return nil
}

func isCloudImportChecksum(value string) bool {
	return cloudImportChecksumPattern.MatchString(value)
}

func isCloudImportDomain(value string) bool {
	if value == "" || len(value) > 253 || value != strings.ToLower(value) || strings.TrimSpace(value) != value || strings.HasPrefix(value, "www.") || strings.HasSuffix(value, ".") {
		return false
	}
	return !strings.ContainsAny(value, "/:@?#[\\] ")
}

func cloudImportConfigurationError(format string, values ...any) error {
	return &ConfigurationError{Message: fmt.Sprintf(format, values...) + "."}
}
