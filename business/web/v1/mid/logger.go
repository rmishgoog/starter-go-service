package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/rmishgoog/starter-go-service/foundations/logger"
	"github.com/rmishgoog/starter-go-service/foundations/web"
)

func Logger(log *logger.Logger) web.MiddleWare {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v := web.GetValues(ctx)
			log.Info(ctx, "request started", "method", r.Method, "path", r.URL.Path)
			err := handler(ctx, w, r)
			log.Info(ctx, "request ended", "method", r.Method, "path", r.URL.Path, "time", time.Since(v.Now), "status", v.StatusCode)
			return err
		}

		return h
	}
	return m
}
