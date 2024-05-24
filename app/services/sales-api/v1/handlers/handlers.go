package handlers

import (
	"github.com/rmishgoog/starter-go-service/app/services/sales-api/v1/handlers/hackgrp"
	v1 "github.com/rmishgoog/starter-go-service/business/web/v1"
	"github.com/rmishgoog/starter-go-service/foundations/web"
)

type Routes struct{}

func (Routes) Add(app *web.App, cfg v1.APIMuxConfig) {
	hackgrp.Routes(app)
}
