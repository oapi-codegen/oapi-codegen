package optionsfiltertags

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// filter/tags: with -include-tags included-tag1,included-tag2 the generated
// ServerInterface exposes exactly IncludedOperation1 and IncludedOperation2 (the
// filtered-tag operation is excluded). The struct below satisfying ServerInterface is
// the compile-time proof; the NoError calls exercise the generated handler signatures.
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
