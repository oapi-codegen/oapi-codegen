package api

import (
	"net/http"

	"github.com/kataras/iris/v12"
)

// ensure that we've conformed to the `ServerInterface` with a compile-time check
var _ ServerInterface = (*Server)(nil)

type Server struct{}

func NewServer() Server {
	return Server{}
}

// (GET /ping)
func (Server) GetPing(ctx iris.Context) {
	resp := Pong{
		Ping: "pong",
	}

	ctx.StatusCode(http.StatusOK)
	_ = ctx.JSON(resp)
}
