package types

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type UUID string

func (u UUID) MarshalJSON() ([]byte, error) {
	if _, err := uuid.Parse(string(u)); err != nil {
		return nil, errors.New("uuid: failed to pass validation")
	}
	return json.Marshal(string(u))
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if _, err := uuid.Parse(s); err != nil {
		return errors.New("uuid: failed to pass validation")
	}
	*u = UUID(s)
	return nil
}
