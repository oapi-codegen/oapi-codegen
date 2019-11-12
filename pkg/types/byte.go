package types

import (
	"encoding/base64"
	"encoding/json"
)

type Byte struct {
	bytes []byte
}

func (b Byte) MarshalJSON() ([]byte, error) {
	enc := base64.StdEncoding.EncodeToString(b.bytes)
	return json.Marshal(enc)
}

func (b *Byte) UnmarshalJSON(data []byte) error {
	var strData string
	err := json.Unmarshal(data, &strData)
	if err != nil {
		return err
	}
	dec, err := base64.StdEncoding.DecodeString(strData)
	if err != nil {
		return err
	}
	b.bytes = dec
	return nil
}
