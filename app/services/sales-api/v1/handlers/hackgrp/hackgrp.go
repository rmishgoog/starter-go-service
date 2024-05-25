package hackgrp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rmishgoog/starter-go-service/foundations/web"
)

func Want(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func Hack(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	status := struct {
		Status string
	}{
		Status: "OK",
	}
	fmt.Println("called here in logger")
	return web.Respond(ctx, w, status, http.StatusOK)

}
