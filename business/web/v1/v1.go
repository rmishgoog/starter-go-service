package v1

import (
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/rmishgoog/starter-go-service/foundations/logger"
)

type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
}

type RouteAdder interface {
	Add(mux *httptreemux.ContextMux, cfg APIMuxConfig)
}

func APIMux(cfg APIMuxConfig, routeAdder RouteAdder) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()
	routeAdder.Add(mux, cfg)
	return mux
}
