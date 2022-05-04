package types

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.UUID(u).String())
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return errors.New("uuid: failed to pass validation")
	}
	*u = UUID(parsed)
	return nil
}
