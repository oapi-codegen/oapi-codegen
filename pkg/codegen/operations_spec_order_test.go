package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func paths(ops []OperationDefinition) []string {
	out := make([]string, len(ops))
	for i, op := range ops {
		out[i] = op.Path
	}
	return out
}

// TestSortRoutesBySpecOrder covers the route-registration ordering used to fix
// issue #1887: routes are registered in spec declaration order (by source
// line), which the user controls, rather than sorted.
func TestSortOperationsBySpecOrder(t *testing.T) {
	t.Run("orders by spec source line, not alphabetically", func(t *testing.T) {
		in := []OperationDefinition{
			// Deliberately not in path-sorted order; SpecOrder is the spec's
			// declaration order.
			{Path: "/zebra", SpecOrder: 4},
			{Path: "/apple", SpecOrder: 8},
			{Path: "/templates/{visibility}/shortcuts", SpecOrder: 6},
		}
		got := sortOperationsBySpecOrder(in)
		assert.Equal(t, []string{"/zebra", "/templates/{visibility}/shortcuts", "/apple"}, paths(got))
	})

	t.Run("stable within a shared path preserves method order", func(t *testing.T) {
		in := []OperationDefinition{
			{Path: "/things/{id}", Method: "GET", SpecOrder: 10},
			{Path: "/things/{id}", Method: "DELETE", SpecOrder: 10},
			{Path: "/things/{id}", Method: "PUT", SpecOrder: 10},
		}
		got := sortOperationsBySpecOrder(in)
		assert.Equal(t, []string{"GET", "DELETE", "PUT"}, []string{got[0].Method, got[1].Method, got[2].Method})
	})

	t.Run("unavailable source locations leave order unchanged", func(t *testing.T) {
		in := []OperationDefinition{
			{Path: "/b", SpecOrder: 0},
			{Path: "/a", SpecOrder: 0},
		}
		got := sortOperationsBySpecOrder(in)
		assert.Equal(t, []string{"/b", "/a"}, paths(got))
	})

	t.Run("does not mutate input", func(t *testing.T) {
		in := []OperationDefinition{
			{Path: "/b", SpecOrder: 20},
			{Path: "/a", SpecOrder: 10},
		}
		_ = sortOperationsBySpecOrder(in)
		assert.Equal(t, []string{"/b", "/a"}, paths(in))
	})
}
