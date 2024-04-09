package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thisisthemurph/hx/middleware"
)

func TestWithHTMXMiddleware_WhenHTMXRequest(t *testing.T) {
	const url = "/path/to/resource"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set test HTMX headers on the request to be interpreted by middleware
	const currentURL = "https://domain.com/test/endpoint"
	req.Header.Set("HX-Current-URL", currentURL)
	req.Header.Set("HX-Boosted", "true")
	req.Header.Set("HX-History-Restore-Request", "true")
	req.Header.Set("HX-Request", "true")
	req.Header.Set("HX-Target", "confirm-btn")
	req.Header.Set("HX-Trigger", "notification-section")
	req.Header.Set("HX-Trigger-Name", "trigger-name")

	rr := httptest.NewRecorder()
	handler := middleware.WithHTMX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := middleware.GetRequestHeaders(r)

		assert.True(t, ok)

		// Assert the values are the expected set values
		assert.Equal(t, currentURL, h.CurrentURL)
		assert.True(t, h.IsBoosted)
		assert.True(t, h.IsHistoryRestoreRequest)
		assert.True(t, h.IsHTMXRequest)
		assert.Equal(t, "confirm-btn", h.Target)
		assert.Equal(t, "notification-section", h.Trigger)
		assert.Equal(t, "trigger-name", h.TriggerName)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestWithHTMXMiddleware_WhenNotHTMXRequest(t *testing.T) {
	const url = "/path/to/resource"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := middleware.WithHTMX(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := middleware.GetRequestHeaders(r)

		assert.True(t, ok)

		// Assert the values are the expected default values
		assert.Empty(t, h.CurrentURL)
		assert.False(t, h.IsBoosted)
		assert.False(t, h.IsHistoryRestoreRequest)
		assert.False(t, h.IsHTMXRequest)
		assert.Empty(t, h.Target)
		assert.Empty(t, h.Trigger)
		assert.Empty(t, h.TriggerName)

		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestWithHTMX_OkIsFalse_WhenMiddlewareNotConfigured(t *testing.T) {
	const url = "/path/to/resource"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := middleware.GetRequestHeaders(r)
		assert.False(t, ok)
		assert.Equal(t, middleware.HTMXRequest{}, h)

		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
