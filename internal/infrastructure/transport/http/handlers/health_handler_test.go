package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	Healthcheck(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Empty(t, w.Body.String())
}
