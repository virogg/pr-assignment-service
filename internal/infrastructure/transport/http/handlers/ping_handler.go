package handlers

import (
	"net/http"

	pkg "github.com/virogg/pr-assignment-service/pkg/http"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	pkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "pong",
	})
}
