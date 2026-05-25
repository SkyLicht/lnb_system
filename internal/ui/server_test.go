package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerServesDashboard(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Body.Len() == 0 {
		t.Fatal("expected dashboard response body")
	}
}
