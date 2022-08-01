package plog

import (
	"encoding/json"
	"fmt"
)

var (
	severityStrings = map[Severity]string{
		SevDebug: "debug",
		SevInfo:  "info",
		SevError: "error",
	}
)

type Severity int

func (s Severity) String() string {
	return severityStrings[s]
}

func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Severity) UnmarshalJSON(b []byte) error {
	var str *string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("severity must be a string: %w", err)
	}

	if s == nil {
		*s = SevInfo
		return nil
	}

	for k, v := range severityStrings {
		if v == *str {
			*s = k
			return nil
		}
	}

	*s = SevInfo
	return nil
}

const (
	SevInfo  = 0
	SevDebug = 1
	SevError = 2
)
