package issue1469

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type impl struct{}

// (GET /test)
func (i *impl) Test(c *fiber.Ctx) error {
	panic("not implemented") // TODO: Implement
}

func TestIssue1469(t *testing.T) {
	server := &impl{}

	r := fiber.New()

	assert.NotPanics(t, func() {
		RegisterHandlers(r, server)
	})

	assert.NotPanics(t, func() {
		RegisterHandlersWithOptions(r, server, FiberServerOptions{
			Middlewares: []MiddlewareFunc{
				func(c *fiber.Ctx) error {
					return nil
				},
			},
		})
	})
}
