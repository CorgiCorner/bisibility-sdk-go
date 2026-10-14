package bisibility

import (
	"encoding/json"
	"fmt"
)

// RankHistoryReportRecord is exactly one observation, check attempt, retained run or item.
// Historical enum values remain strings; null never becomes a zero position.
type RankHistoryReportRecord struct {
	Observation *RankHistoryReportObservation
	Attempt     *RankHistoryReportAttempt
	Run         *RankHistoryReportRun
	RunItem     *RankHistoryReportRunItem
	Extensions  map[string]json.RawMessage
}

func (r *RankHistoryReportRecord) UnmarshalJSON(data []byte) error {
	if !validReportContract(data, "RankHistoryReportRecord") {
		return fmt.Errorf("invalid required rank history record fields")
	}
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(data, &shape); err != nil {
		return err
	}
	var kind string
	if shape["kind"] != nil {
		if string(shape["kind"]) == "null" {
			return fmt.Errorf("rank history record kind must not be null")
		}
		if err := json.Unmarshal(shape["kind"], &kind); err != nil {
			return err
		}
		if kind == "" {
			return fmt.Errorf("rank history record kind must not be empty")
		}
	}
	var decoded RankHistoryReportRecord
	var known []byte
	var err error
	switch kind {
	case "logical_rank_check":
		var value RankHistoryReportAttempt
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		if err := ValidatePublicIDPrefix(value.ID, PublicIDPrefixCheck); err != nil {
			return err
		}
		if err := ValidatePublicIDPrefix(value.KeywordID, PublicIDPrefixKeyword); err != nil {
			return err
		}
		known, err = json.Marshal(value)
		decoded.Attempt = &value
	case "rank_check_run":
		var value RankHistoryReportRun
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		known, err = json.Marshal(value)
		decoded.Run = &value
	case "rank_check_run_item":
		var value RankHistoryReportRunItem
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		known, err = json.Marshal(value)
		decoded.RunItem = &value
	case "":
		if shape["evidence"] == nil {
			return fmt.Errorf("missing rank history evidence")
		}
		var value RankHistoryReportObservation
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		if err := ValidatePublicIDPrefix(value.ID, PublicIDPrefixCheck); err != nil {
			return err
		}
		if err := ValidatePublicIDPrefix(value.KeywordID, PublicIDPrefixKeyword); err != nil {
			return err
		}
		known, err = json.Marshal(value)
		decoded.Observation = &value
	default:
		return fmt.Errorf("unsupported rank history record kind")
	}
	if err != nil {
		return err
	}
	extensions, err := reportExtensions(data, known)
	if err != nil {
		return err
	}
	decoded.Extensions = extensions
	*r = decoded
	return nil
}

func (r RankHistoryReportRecord) MarshalJSON() ([]byte, error) {
	count := 0
	if r.Observation != nil {
		count++
	}
	if r.Attempt != nil {
		count++
	}
	if r.Run != nil {
		count++
	}
	if r.RunItem != nil {
		count++
	}
	if count != 1 {
		return nil, fmt.Errorf("rank history record must contain exactly one variant")
	}
	var known []byte
	var err error
	switch {
	case r.Observation != nil:
		known, err = json.Marshal(r.Observation)
	case r.Attempt != nil:
		known, err = json.Marshal(r.Attempt)
	case r.Run != nil:
		known, err = json.Marshal(r.Run)
	case r.RunItem != nil:
		known, err = json.Marshal(r.RunItem)
	}
	if err != nil {
		return nil, err
	}
	return mergeReportExtensions(known, r.Extensions)
}

func reportExtensions(data, known []byte) (map[string]json.RawMessage, error) {
	var original, fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &original); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(known, &fields); err != nil {
		return nil, err
	}
	for name := range fields {
		delete(original, name)
	}
	return original, nil
}
func mergeReportExtensions(known []byte, extensions map[string]json.RawMessage) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(known, &fields); err != nil {
		return nil, err
	}
	for name, value := range extensions {
		if _, exists := fields[name]; !exists {
			fields[name] = value
		}
	}
	return json.Marshal(fields)
}
