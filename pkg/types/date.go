package types

import (
	"encoding/json"
	"time"
)

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#DateFormat
const DateFormat = "2006-01-02"

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#Date
type Date struct {
	time.Time
}

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#MarshalJSON
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(DateFormat))
}

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#UnmarshalJSON
func (d *Date) UnmarshalJSON(data []byte) error {
	var dateStr string
	err := json.Unmarshal(data, &dateStr)
	if err != nil {
		return err
	}
	parsed, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#String
func (d Date) String() string {
	return d.Time.Format(DateFormat)
}

// Deprecated: This has been replaced by github.com/oapi-codegen/runtime/types#UnmarshalText
func (d *Date) UnmarshalText(data []byte) error {
	parsed, err := time.Parse(DateFormat, string(data))
	if err != nil {
		return err
	}
	d.Time = parsed
	return nil
}
