package bisibility

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestProjectDefaultsJSONTagsMatchContract(t *testing.T) {
	t.Parallel()

	assertJSONTagsEqual(t, reflect.TypeOf(ProjectDefaults{}), []string{
		"city",
		"country",
		"cron_expression",
		"device",
		"frequency",
		"jitter_minutes",
		"last_checked_at",
		"location_key",
		"next_check_at",
		"project_id",
		"serp_depth",
		"serp_stop_on_match",
		"source",
		"timezone",
		"updated_at",
	})
	assertJSONTagsEqual(t, reflect.TypeOf(ProjectDefaultsPatch{}), []string{
		"city",
		"country",
		"cron_expression",
		"device",
		"frequency",
		"jitter_minutes",
		"location_key",
		"serp_stop_on_match",
		"timezone",
	})
}

func TestProjectDefaultsSourceValues(t *testing.T) {
	t.Parallel()

	assertEqual(t, ProjectDefaultsSourceDerived, ProjectDefaultsSource("derived"))
	assertEqual(t, ProjectDefaultsSourceExplicit, ProjectDefaultsSource("explicit"))
	assertEqual(t, ProjectDefaultsSourceFallback, ProjectDefaultsSource("fallback"))
}

func assertJSONTagsEqual(t *testing.T, structType reflect.Type, want []string) {
	t.Helper()

	got := make([]string, 0, structType.NumField())
	for index := 0; index < structType.NumField(); index++ {
		tag := strings.Split(structType.Field(index).Tag.Get("json"), ",")[0]
		if tag != "" && tag != "-" {
			got = append(got, tag)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s JSON tags differ\ngot:  %v\nwant: %v", structType.Name(), got, want)
	}
}
