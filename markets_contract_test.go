package bisibility

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMarketLanguageJSONTagsMatchContract(t *testing.T) {
	t.Parallel()

	assertJSONTagsEqual(t, reflect.TypeOf(Keyword{}), []string{
		"country", "created_at", "device", "id", "intent", "language_code", "language_label",
		"latest_position", "location", "location_key", "previous_position", "project_id", "ranking_url",
		"schedule", "tags", "target_url", "text", "topic", "updated_at",
	})
	assertJSONTagsEqual(t, reflect.TypeOf(KeywordMatchMarket{}), []string{
		"country_code", "device", "language_code", "language_label", "location", "location_key",
	})
	assertJSONTagsEqual(t, reflect.TypeOf(LocationSuggestion{}), []string{
		"city_name", "country_code", "display_name", "hl", "kind", "language_code", "language_label",
		"location_key", "region_code", "region_name",
	})
}

func TestLanguageQualifiedLocationKeysRoundTrip(t *testing.T) {
	t.Parallel()

	input := CreateKeywordInput{
		Keyword:     "rank tracker",
		LocationKey: "ES/Andalusia/Malaga@en",
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal keyword input: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("decode keyword input: %v", err)
	}
	assertEqual(t, payload["location_key"], "ES/Andalusia/Malaga@en")

	var keyword Keyword
	if err := json.Unmarshal([]byte(`{"location_key":"ES/Andalusia/Malaga@en","language_code":"en","language_label":"English"}`), &keyword); err != nil {
		t.Fatalf("decode keyword response: %v", err)
	}
	assertEqual(t, keyword.LocationKey, "ES/Andalusia/Malaga@en")
	assertEqual(t, keyword.LanguageCode, "en")
	assertEqual(t, keyword.LanguageLabel, "English")
}
