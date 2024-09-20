package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (GET /ping)
func (Server) GetPing(ctx *gin.Context) {
	resp := Pong{
		Ping: "pong",
	}

	ctx.JSON(http.StatusOK, resp)
}
