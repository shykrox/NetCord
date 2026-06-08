package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerRoutesRequireAuth(t *testing.T) {
	router := NewRouter(nil, nil, nil, nil, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/servers", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
