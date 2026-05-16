package issue2367

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// stubServer satisfies ServerInterface and lets the test prove the generated
// package compiles end-to-end. The bug in issue 2367 emitted a constant
// referencing a context-key type that was never declared (because the spec's
// `security:` named a scheme not present under components/securitySchemes),
// so the generated file failed to compile. Instantiating the interface here
// would not be reachable if that regression returned.
type stubServer struct{}

func (stubServer) GetHello(c *gin.Context) {}

func TestGeneratedServerCompilesWithUndefinedSecurityScheme(t *testing.T) {
	var iface ServerInterface = stubServer{}
	assert.NotNil(t, iface)
}
