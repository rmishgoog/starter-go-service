package mid

import (
	"context"
	"net/http"

	"github.com/rmishgoog/starter-go-service/business/web/v1/response"
	"github.com/rmishgoog/starter-go-service/foundations/logger"
	"github.com/rmishgoog/starter-go-service/foundations/web"
)

func Errors(log *logger.Logger) web.MiddleWare {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				log.Error(ctx, "message", "msg", err)

				var er response.ErrorDocument
				var status int
				switch {
				case response.IsError(err):
					reqErr := response.GetError(err)
					er = response.ErrorDocument{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = response.ErrorDocument{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError

				}

				if err := web.Respond(ctx, w, er, status); err != nil {
					return err
				}

				if web.IsShutdown(err) {
					return err
				}
			}
			return nil
		}
		return h
	}
	return m
}
