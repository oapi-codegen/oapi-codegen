package types

import (
	"encoding/json"
	"reflect"
	"time"
)

const DateFormat = "2006-01-02"

var BaseDateType = reflect.TypeOf(Date{})

type Date struct {
	time.Time

	// Top level *Params structs may contain types that wrap this type,
	// and we need a way to identify those types as needing special
	// handling.  We will use reflect.ConvertibleTo to do this, assuming this
	// field will make this data structure unlikely to collide with types
	// accidentally.
	__oapiBaseDateType__ struct{}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format(DateFormat))
}

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
