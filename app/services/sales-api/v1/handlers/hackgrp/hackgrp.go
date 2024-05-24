package hackgrp

import (
	"context"
	"encoding/json"
	"net/http"
)

func Want(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func Hack(w http.ResponseWriter, r *http.Request) {

	status := struct {
		Status string
	}{
		Status: "OK",
	}
	json.NewEncoder(w).Encode(status)

}
