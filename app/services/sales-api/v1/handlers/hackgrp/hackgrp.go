package hackgrp

import (
	"context"
	"encoding/json"
	"net/http"
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
	return json.NewEncoder(w).Encode(status)

}
