package anyofallofoneof

// OverlayClient defines model for OverlayClient.
type OverlayClient struct {
	Name string `json:"name"`
}

// OverlayClientWithId defines model for OverlayClientWithId.
type OverlayClientWithId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
