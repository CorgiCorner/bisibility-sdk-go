package bisibility

// CheckNotificationSetting is the server's check-summary preference. Future
// response values remain readable. Manual app runs summarize under All and Important.
type CheckNotificationSetting string

const (
	CheckNotificationSettingAll       CheckNotificationSetting = "all"
	CheckNotificationSettingImportant CheckNotificationSetting = "important"
	CheckNotificationSettingNone      CheckNotificationSetting = "none"
)

// CheckNotificationSettingInput selects all nonempty app windows, important
// changes only (the server default), or none. Only these three input values are valid.
type CheckNotificationSettingInput string

const (
	CheckNotificationSettingInputAll       CheckNotificationSettingInput = "all"
	CheckNotificationSettingInputImportant CheckNotificationSettingInput = "important"
	CheckNotificationSettingInputNone      CheckNotificationSettingInput = "none"
)

// NotificationPreferenceWarning describes an ignored deprecated preference write.
type NotificationPreferenceWarning struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func validateCheckNotificationPreferences(setting *CheckNotificationSettingInput, topN, positionMove *int) error {
	if setting != nil {
		switch *setting {
		case CheckNotificationSettingInputAll, CheckNotificationSettingInputImportant, CheckNotificationSettingInputNone:
		default:
			return &ConfigurationError{Message: "check_notification_setting must be all, important, or none."}
		}
	}
	for _, field := range []struct {
		name  string
		value *int
	}{{"check_notification_top_n", topN}, {"check_notification_position_move", positionMove}} {
		if field.value != nil && (*field.value < 1 || *field.value > 100) {
			return &ConfigurationError{Message: field.name + " must be between 1 and 100."}
		}
	}
	return nil
}
