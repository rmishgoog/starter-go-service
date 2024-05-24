package web

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
)

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []MiddleWare
}

func NewApp(shutdown chan os.Signal, mw ...MiddleWare) *App {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/trusted.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers.
	// https://w3c.github.io/trace-context/

	//mux := http.NewServeMux()

	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, path string, handler Handler, mw ...MiddleWare) {

	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		// ADD ANY LOGIC HERE BEFORE CALLING THE REAL FRIKIN MUX
		if err := handler(r.Context(), w, r); err != nil {
			fmt.Println(err) // To be implemented later
		}
	}

	a.ContextMux.Handle(method, path, h)
}
