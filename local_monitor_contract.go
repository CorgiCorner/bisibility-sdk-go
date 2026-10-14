package bisibility

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed local_monitor_contract.json
var localContractJSON []byte

// Metadata values may evolve; report/version discriminants and shapes remain strict.
func localOutputSchema(s reportSchema, field string) reportSchema {
	if field != "report" && field != "version" {
		s.Enum = nil
		s.Const = nil
	}
	for name, child := range s.Properties {
		s.Properties[name] = localOutputSchema(child, name)
	}
	if s.Items != nil {
		child := localOutputSchema(*s.Items, field)
		s.Items = &child
	}
	for i, child := range s.AnyOf {
		s.AnyOf[i] = localOutputSchema(child, field)
	}
	for i, child := range s.OneOf {
		s.OneOf[i] = localOutputSchema(child, field)
	}
	return s
}

func unmarshalLocalContract(raw []byte, definition string, into any) error {
	var definitions map[string]reportSchema
	if err := json.Unmarshal(localContractJSON, &definitions); err != nil {
		return err
	}
	schema, ok := definitions[definition]
	if !ok || !validReportValue(raw, localOutputSchema(schema, ""), definitions) {
		return fmt.Errorf("invalid required Local retained-read contract: %s", definition)
	}
	return json.Unmarshal(raw, into)
}

func (r *LocalMonitorRunResultResponse) UnmarshalJSON(raw []byte) error {
	type wire LocalMonitorRunResultResponse
	return unmarshalLocalContract(raw, "LocalMonitorRunResult", (*wire)(r))
}

func (r *LocalMonitorRunHistoryResponse) UnmarshalJSON(raw []byte) error {
	type wire LocalMonitorRunHistoryResponse
	return unmarshalLocalContract(raw, "LocalMonitorRunHistory", (*wire)(r))
}

func (r *LocalMonitorPointHistoryResponse) UnmarshalJSON(raw []byte) error {
	type wire LocalMonitorPointHistoryResponse
	return unmarshalLocalContract(raw, "LocalMonitorPointHistory", (*wire)(r))
}

func (r *LocalMonitorPointListingsResponse) UnmarshalJSON(raw []byte) error {
	type wire LocalMonitorPointListingsResponse
	return unmarshalLocalContract(raw, "LocalMonitorPointListings", (*wire)(r))
}

func (r *LocalMonitorPointReport) UnmarshalJSON(raw []byte) error {
	type wire LocalMonitorPointReport
	return unmarshalLocalContract(raw, "LocalMonitorPointReport", (*wire)(r))
}
