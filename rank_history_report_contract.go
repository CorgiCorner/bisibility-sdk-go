package bisibility

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"time"
)

//go:embed rank_history_report_contract.json
var reportContractJSON []byte

type reportSchema struct {
	Type       string                  `json:"type"`
	Format     string                  `json:"format"`
	Ref        string                  `json:"$ref"`
	Properties map[string]reportSchema `json:"properties"`
	Required   []string                `json:"required"`
	Items      *reportSchema           `json:"items"`
	AnyOf      []reportSchema          `json:"anyOf"`
	OneOf      []reportSchema          `json:"oneOf"`
	Const      json.RawMessage         `json:"const"`
	Enum       []json.RawMessage       `json:"enum"`
	Pattern    string                  `json:"pattern"`
	MaxLength  *int                    `json:"maxLength"`
	MinLength  *int                    `json:"minLength"`
	MaxItems   *int                    `json:"maxItems"`
	MinItems   *int                    `json:"minItems"`
	Maximum    *float64                `json:"maximum"`
	Minimum    *float64                `json:"minimum"`
}

// Validate the published required fields while allowing future additional fields.
func validReportContract(raw []byte, definition string) bool {
	var definitions map[string]reportSchema
	if json.Unmarshal(reportContractJSON, &definitions) != nil {
		return false
	}
	schema, ok := definitions[definition]
	for name, definition := range definitions {
		definitions[name] = rankOutputSchema(definition, "")
	}
	return ok && validReportValue(raw, rankOutputSchema(schema, ""), definitions)
}

func validReportValue(raw []byte, s reportSchema, defs map[string]reportSchema) bool {
	if s.Ref != "" {
		target, ok := defs[s.Ref[strings.LastIndex(s.Ref, "/")+1:]]
		return ok && validReportValue(raw, target, defs)
	}
	if len(s.AnyOf) > 0 {
		for _, v := range s.AnyOf {
			if validReportValue(raw, v, defs) {
				return true
			}
		}
		return false
	}
	if len(s.OneOf) > 0 {
		count := 0
		for _, v := range s.OneOf {
			if validReportValue(raw, v, defs) {
				count++
			}
		}
		return count == 1
	}
	raw = bytes.TrimSpace(raw)
	if s.Type == "null" {
		return bytes.Equal(raw, []byte("null"))
	}
	if bytes.Equal(raw, []byte("null")) {
		return false
	}
	if len(s.Const) > 0 && !bytes.Equal(raw, s.Const) {
		return false
	}
	if len(s.Enum) > 0 {
		found := false
		for _, v := range s.Enum {
			if bytes.Equal(raw, v) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	switch s.Type {
	case "object":
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || fields == nil {
			return false
		}
		for _, key := range s.Required {
			if _, ok := fields[key]; !ok {
				return false
			}
		}
		for key, value := range fields {
			if property, ok := s.Properties[key]; ok && !validReportValue(value, property, defs) {
				return false
			}
		}
	case "array":
		var values []json.RawMessage
		if json.Unmarshal(raw, &values) != nil || values == nil || s.Items == nil {
			return false
		}
		if (s.MaxItems != nil && len(values) > *s.MaxItems) || (s.MinItems != nil && len(values) < *s.MinItems) {
			return false
		}
		for _, value := range values {
			if !validReportValue(value, *s.Items, defs) {
				return false
			}
		}
	case "string":
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return false
		}
		if (s.MaxLength != nil && len([]rune(value)) > *s.MaxLength) || (s.MinLength != nil && len([]rune(value)) < *s.MinLength) {
			return false
		}
		if s.Format == "date-time" {
			if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
				return false
			}
		}
		if s.Pattern != "" {
			pattern, err := regexp.Compile(s.Pattern)
			if err != nil || !pattern.MatchString(value) {
				return false
			}
		}
	case "number", "integer":
		var value float64
		if json.Unmarshal(raw, &value) != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
		if s.Type == "integer" && math.Trunc(value) != value {
			return false
		}
		if s.Maximum != nil && value > *s.Maximum {
			return false
		}
		if s.Minimum != nil && value < *s.Minimum {
			return false
		}
	case "boolean":
		var value bool
		if json.Unmarshal(raw, &value) != nil {
			return false
		}
	default:
		return false
	}
	return true
}
