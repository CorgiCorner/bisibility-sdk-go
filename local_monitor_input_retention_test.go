package bisibility

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestLocalUTF16CursorBounds(t *testing.T) {
	query := url.Values{}
	if localReadQuery(query, "", "", "", strings.Repeat("😀", 4000), true) != nil {
		t.Error("valid capacity rejected")
	}
	if localReadQuery(query, "", "", "", strings.Repeat("😀", 4001), true) == nil {
		t.Error("oversized UTF16 accepted")
	}
}

func TestLocalRetainedFractionalZoom(t *testing.T) {
	f := localFixture(t)
	for _, key := range []string{"result", "export_json"} {
		t.Run(key, func(t *testing.T) {
			var raw map[string]any
			json.Unmarshal(f[key], &raw)
			data := raw
			if key == "result" {
				data = raw["data"].(map[string]any)
			}
			data["points"].([]any)[0].(map[string]any)["zoom"] = 5.5
			body, _ := json.Marshal(raw)
			if key == "result" {
				var result LocalMonitorRunResultResponse
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal(err)
				}
				if result.Data.Points[0].Zoom == nil || *result.Data.Points[0].Zoom != 5.5 {
					t.Error("fractional zoom lost")
				}
			} else {
				var result LocalMonitorPointReport
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal(err)
				}
				if result.Points[0].Zoom == nil || *result.Points[0].Zoom != 5.5 {
					t.Error("fractional zoom lost")
				}
			}
		})
	}
}
