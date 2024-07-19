package v1

import (
	"os"

	"github.com/rmishgoog/starter-go-service/business/web/v1/auth"
	"github.com/rmishgoog/starter-go-service/business/web/v1/mid"
	"github.com/rmishgoog/starter-go-service/foundations/logger"
	"github.com/rmishgoog/starter-go-service/foundations/web"
)

type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	Auth     *auth.Auth
}

type RouteAdder interface {
	Add(app *web.App, cfg APIMuxConfig)
}

func APIMux(cfg APIMuxConfig, routeAdder RouteAdder) *web.App {
	app := web.NewApp(cfg.Shutdown, mid.Logger(cfg.Log), mid.Metrics(), mid.Errors(cfg.Log), mid.Panics())
	routeAdder.Add(app, cfg)
	return app
}
