package filteroperations

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type server struct{}

func (s server) IncludedOperation1(ctx echo.Context) error {
	return nil
}

func (s server) IncludedOperation2(ctx echo.Context) error {
	return nil
}

func TestServer(t *testing.T) {
	assert := assert.New(t)

	var s ServerInterface = server{}
	assert.NoError(s.IncludedOperation1(nil))
	assert.NoError(s.IncludedOperation2(nil))
}
