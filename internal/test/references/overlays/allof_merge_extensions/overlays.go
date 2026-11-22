package allofmergeextensions

// OverlayClient defines model for OverlayClient.
type OverlayClient struct {
	Name string `json:"name"`
}

// OverlayClientWithId defines model for OverlayClientWithId.
type OverlayClientWithId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// OverlayBaseOnly is the user-provided override for the BaseOnly schema.
// DerivedNoOverride composes BaseOnly via allOf and must NOT be aliased
// to this type — it has to remain its own struct so the Extra field is
// preserved.
type OverlayBaseOnly struct {
	Name string `json:"name"`
}
