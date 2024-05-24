package hackgrp

import (
	"net/http"

	"github.com/dimfeld/httptreemux/v5"
)

func Routes(mux *httptreemux.ContextMux) {
	mux.Handle(http.MethodGet, "/hack", Hack)
}
