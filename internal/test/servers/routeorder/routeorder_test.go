package routeorder

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/stretchr/testify/assert"

	"github.com/oapi-codegen/testutil"
)

// server records which operation the router dispatched to.
type server struct{ hit string }

func (s *server) PrivateTemplateGet(c *fiber.Ctx, id string) error {
	s.hit = "PrivateTemplateGet:" + id
	return c.SendStatus(http.StatusOK)
}

func (s *server) TemplateShortcutGetAll(c *fiber.Ctx, templateVisibility string) error {
	s.hit = "TemplateShortcutGetAll:" + templateVisibility
	return c.SendStatus(http.StatusOK)
}

// TestFiberRouteSpecOrder verifies the issue-1887 fix: because Fiber matches
// routes in registration order, and registration now follows spec declaration
// order, the user can put /templates/{visibility}/shortcuts before
// /templates/privates/{id} in the spec to have it matched first. Before the
// fix, registration was alphabetically sorted (privates/{id} first regardless
// of spec order), so /templates/privates/shortcuts was mis-dispatched to
// PrivateTemplateGet with id="shortcuts".
func TestFiberRouteSpecOrder(t *testing.T) {
	cases := []struct {
		name, path, wantHit string
	}{
		{"ambiguous path resolves to the earlier-declared route", "/api/v1/templates/privates/shortcuts", "TemplateShortcutGetAll:privates"},
		{"other visibility shortcuts", "/api/v1/templates/publics/shortcuts", "TemplateShortcutGetAll:publics"},
		{"private template by id", "/api/v1/templates/privates/42", "PrivateTemplateGet:42"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &server{}
			app := fiber.New()
			RegisterHandlers(app, s)

			rr := testutil.NewRequest().Get(tc.path).GoWithHTTPHandler(t, adaptor.FiberApp(app)).Recorder
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tc.wantHit, s.hit)
		})
	}
}
