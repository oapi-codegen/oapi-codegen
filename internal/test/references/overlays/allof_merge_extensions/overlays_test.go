package allofmergeextensions

import "testing"

func TestAllOfOverlay(t *testing.T) {
	var inner any = Client{}
	_, ok := inner.(OverlayClient)
	if !ok {
		t.Errorf("expected Client to be of type OverlayClient")
	}

	var outer any = ClientWithId{}
	_, ok = outer.(OverlayClientWithId)
	if !ok {
		t.Errorf("expected ClientWithId to be of type OverlayClientWithId")
	}
}

// TestBaseOnlyOverlay covers the harder regression path: when only the
// base schema has x-go-type (via overlay) and the derived allOf schema
// has no override of its own, the derived schema must still be emitted
// as a distinct struct containing all composed fields. The previous
// bug leaked BaseOnly's x-go-type up through the allOf merge, producing
// `type DerivedNoOverride = OverlayBaseOnly` and silently dropping the
// Extra field. The struct literal below would fail to compile under
// the buggy codegen because OverlayBaseOnly has no Extra field.
func TestBaseOnlyOverlay(t *testing.T) {
	d := DerivedNoOverride{Name: "x", Extra: "y"}

	var asAny any = d
	if _, ok := asAny.(OverlayBaseOnly); ok {
		t.Error("DerivedNoOverride must not be aliased to OverlayBaseOnly; composition would drop the Extra field")
	}

	if d.Name != "x" || d.Extra != "y" {
		t.Errorf("field values not preserved through composed struct: %+v", d)
	}
}
