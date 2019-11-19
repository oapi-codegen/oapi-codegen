package xmlutil

import "errors"

// RawMessage is a raw encoded XML value.
// It implements Marshaler and Unmarshaler and can
// be used to delay XML decoding or precompute a XML encoding.
type RawMessage []byte

// MarshalXML returns m as the XML encoding of m.
func (m RawMessage) MarshalXML() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalXML sets *m to a copy of data.
func (m *RawMessage) UnmarshalXML(data []byte) error {
	if m == nil {
		return errors.New("RawMessage: UnmarshalXML on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}
