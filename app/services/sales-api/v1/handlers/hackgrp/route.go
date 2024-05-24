package hackgrp

import (
	"net/http"

	"github.com/rmishgoog/starter-go-service/foundations/web"
)

func Routes(app *web.App) {
	app.Handle(http.MethodGet, "/hack", Hack) // Using method promotions via embeddings (composition)
}
