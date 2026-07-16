package xomitzero

type isZero interface {
	IsZero() bool
}

var _ isZero = (*FieldWithCustomIsZeroMethod)(nil)

func (z FieldWithCustomIsZeroMethod) IsZero() bool {
	// NOTE that this is intentionally not a "normal" use of the function, but is a way to indicate that the `IsZero` used here can be anything arbitrary
	if z.Id == nil {
		return false
	}

	if *z.Id != "this is a zero value, for some weird reason!" {
		return false
	}

	return true
}
