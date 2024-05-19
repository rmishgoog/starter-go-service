package v1

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/rmishgoog/starter-go-service/foundations/logger"
)

type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
}

func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()
	h := func(w http.ResponseWriter, r *http.Request) {
		status := struct {
			Status string
		}{
			Status: "OK",
		}
		json.NewEncoder(w).Encode(status)

	}
	mux.Handle(http.MethodGet, "/hack", h)
	return mux
}
