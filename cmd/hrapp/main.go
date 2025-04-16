package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	h := server.Default()
	h.GET("/", func(c context.Context, rc *app.RequestContext) {
		rc.String(consts.StatusOK, "Hello, world!")
	})

	h.Spin()
}
