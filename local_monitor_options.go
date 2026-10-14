package bisibility

import (
	"fmt"
	"net/url"
	"regexp"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

type LocalMonitorReadOptions struct{ AsOf string }
type LocalMonitorHistoryOptions struct {
	AsOf, From, To string
	Cursor         *string
}
type LocalMonitorExportOptions struct{ AsOf, Version, Format string }

var localInstantPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$`)

func localReadQuery(query url.Values, asOf, from, to, cursor string, hasCursor bool) error {
	for key, value := range map[string]string{"as_of": asOf, "from": from, "to": to} {
		if value == "" {
			continue
		}
		at, err := time.Parse("2006-01-02T15:04:05.000Z", value)
		if err != nil || !localInstantPattern.MatchString(value) || at.UnixMilli() < 0 || at.After(time.Now()) {
			return fmt.Errorf("invalid Local %s instant", key)
		}
		query.Set(key, value)
	}
	end := to
	if end == "" {
		end = asOf
	}
	if (from != "" && end != "" && from > end) || (to != "" && asOf != "" && to > asOf) {
		return fmt.Errorf("invalid Local history interval")
	}
	if hasCursor {
		if !utf8.ValidString(cursor) || len(utf16.Encode([]rune(cursor))) < 1 || len(utf16.Encode([]rune(cursor))) > 8000 {
			return fmt.Errorf("Local cursor must contain 1..8000 characters")
		}
		query.Set("cursor", cursor)
	}
	return nil
}

func validateLocalRoutePublicIDs(parts []string, position int) error {
	if err := validateRoutePart(parts, position+1, "monitorID", PublicIDPrefixLocalMonitor); err != nil {
		return err
	}
	for i := position + 2; i+1 < len(parts); i++ {
		if parts[i] == "runs" {
			if err := validateRoutePart(parts, i+1, "runID", PublicIDPrefixLocalRun); err != nil {
				return err
			}
		}
		if parts[i] == "points" {
			if err := validateRoutePart(parts, i+1, "targetID", PublicIDPrefixLocalTarget); err != nil {
				return err
			}
		}
	}
	return nil
}
