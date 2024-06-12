package hackgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/rmishgoog/starter-go-service/business/web/v1/response"
	"github.com/rmishgoog/starter-go-service/foundations/web"
)

func Want(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func Hack(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// This is juts a test code
	if n := rand.Intn(100) % 2; n == 0 {
		return response.NewError(errors.New("message: a trusted error from the handler"), http.StatusBadRequest)
	}
	status := struct {
		Status string
	}{
		Status: "OK",
	}
	return web.Respond(ctx, w, status, http.StatusOK)

}
