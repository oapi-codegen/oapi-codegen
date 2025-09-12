package anyofallofoneof

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
