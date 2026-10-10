package bisibility

import (
	"encoding/json"
	"testing"
)

// TestAITrackingSourceConfigurationMarshalsExplicitNullModel covers
// contract-audit finding #5 for the specific "model: null" shape the API
// requires on a tracking payload: a nil Model must serialize as an explicit
// JSON null, never be omitted, because the server-side schema is nullable
// but required.
func TestAITrackingSourceConfigurationMarshalsExplicitNullModel(t *testing.T) {
	configuration := AITrackingSourceConfiguration{
		Provider:   "dataforseo",
		Endpoint:   "serp/organic",
		Engine:     "google",
		Source:     "consumer_scrape",
		Model:      nil,
		Parameters: map[string]any{},
	}
	raw, err := json.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	modelRaw, ok := value["model"]
	if !ok {
		t.Fatalf("tracking source configuration omitted the required nullable model field: %s", raw)
	}
	if string(modelRaw) != "null" {
		t.Fatalf("tracking source configuration serialized model = %s, want null", modelRaw)
	}
}

// TestNullableStringThreeStateMarshal covers contract-audit finding #5 for the
// patch payloads whose fields are both optional AND nullable: callers must be
// able to pick between "omit" (do not change), "null" (clear) and a string
// value (set), and the SDK must send only those bytes to the server.
func TestNullableStringThreeStateMarshal(t *testing.T) {
	cases := []struct {
		name  string
		patch AITrackingPromptPatch
		want  string
	}{
		{
			name:  "zero patch omits every optional field",
			patch: AITrackingPromptPatch{},
			want:  `{}`,
		},
		{
			name:  "clearing topic sends explicit null, label stays omitted",
			patch: AITrackingPromptPatch{TopicID: NullString()},
			want:  `{"topic_id":null}`,
		},
		{
			name:  "setting topic sends the string value",
			patch: AITrackingPromptPatch{TopicID: StringValue("ait_a00000000000000000000000")},
			want:  `{"topic_id":"ait_a00000000000000000000000"}`,
		},
		{
			name: "clearing label and setting topic co-exist in one patch",
			patch: AITrackingPromptPatch{
				TopicID: StringValue("ait_a00000000000000000000000"),
				Label:   NullString(),
			},
			want: `{"label":null,"topic_id":"ait_a00000000000000000000000"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.patch)
			if err != nil {
				t.Fatal(err)
			}
			assertJSONEqual(t, string(raw), tc.want)
		})
	}
}

// TestNullableStringAccessors confirms the inspection helpers callers rely
// on when round-tripping a decoded patch.
func TestNullableStringAccessors(t *testing.T) {
	var unset NullableString
	if unset.IsSet() {
		t.Fatal("zero NullableString must not report IsSet")
	}
	if unset.IsNull() {
		t.Fatal("zero NullableString must not report IsNull")
	}
	if unset.Value() != nil {
		t.Fatal("zero NullableString must have a nil value pointer")
	}

	cleared := NullString()
	if !cleared.IsSet() || !cleared.IsNull() {
		t.Fatal("NullString must be set and null")
	}
	if cleared.Value() != nil {
		t.Fatal("NullString must have a nil value pointer")
	}

	stored := StringValue("corgi")
	if !stored.IsSet() || stored.IsNull() {
		t.Fatal("StringValue must be set and non-null")
	}
	if ptr := stored.Value(); ptr == nil || *ptr != "corgi" {
		t.Fatalf("stored.Value = %v, want pointer to \"corgi\"", ptr)
	}

	// Decoded JSON null round-trips to the cleared state so a client that
	// reads the server response can re-encode it without losing the
	// three-state semantics.
	var decoded NullableString
	if err := json.Unmarshal([]byte("null"), &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.IsSet() || !decoded.IsNull() {
		t.Fatal("decoded null must round-trip to IsSet && IsNull")
	}
	var decodedValue NullableString
	if err := json.Unmarshal([]byte(`"corgi"`), &decodedValue); err != nil {
		t.Fatal(err)
	}
	if !decodedValue.IsSet() || decodedValue.IsNull() || *decodedValue.Value() != "corgi" {
		t.Fatal("decoded string must round-trip to the stored value")
	}
}
