package middleware

import (
	"context"
	"net/http"
)

const (
	headerBoosted               string = "HX-Boosted"
	headerRequest               string = "HX-Request"
	headerCurrentURL            string = "HX-Current-URL"
	headerHistoryRestoreRequest string = "HX-History-Restore-Request"
	headerTarget                string = "HX-Target"
	headerTrigger               string = "HX-Trigger"
	headerTriggerName           string = "HX-Trigger-Name"
)

type ContextKey string

const HTMXRequestKey ContextKey = "HTMXRequest"

// HTMXRequest is a struct detailing HTMX request header values.
// HTMX documentation: https://htmx.org/reference/#request_headers
type HTMXRequest struct {
	CurrentURL              string // The current URL of the browser.
	IsBoosted               bool   // Indicates that the request is via an element using hx-boost.
	IsHistoryRestoreRequest bool   // Indicates if the request is for history restoration after a miss in the local history cache.
	IsHTMXRequest           bool   // Indicates if the request was a HTMX request; false if the HX-Request header is not present.
	Target                  string // The id of the triggering element, if it exists.
	Trigger                 string // The id of the triggered element, if it exists.
	TriggerName             string // The name of the triggering element, if it exists.
}

// WithHTMX is a middleware function for interpreting the HTMX request headers and making
// them available within the handler's context. If the request is not a HTMX request, the
// HTMXRequest result will take all default values.
func WithHTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		htmxRequest := HTMXRequest{
			CurrentURL:              r.Header.Get(headerCurrentURL),
			IsBoosted:               r.Header.Get(headerBoosted) == "true",
			IsHistoryRestoreRequest: r.Header.Get(headerHistoryRestoreRequest) == "true",
			IsHTMXRequest:           r.Header.Get(headerRequest) == "true",
			Target:                  r.Header.Get(headerTarget),
			Trigger:                 r.Header.Get(headerTrigger),
			TriggerName:             r.Header.Get(headerTriggerName),
		}

		ctx := context.WithValue(r.Context(), HTMXRequestKey, htmxRequest)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestHeaders extracts the HTMXRequest headers from the provided HTTP request.
// It retrieves the HTMXRequest object stored in the request's context.
// Parameters:
//
//	r (*http.Request): The HTTP request from which to extract the HTMXRequest headers.
//
// Returns:
//
//	HTMXRequest: The HTMXRequest headers if found; returns a blank struct if not found.
//	bool: A boolean indicating whether the HTMXRequest headers were successfully extracted (true) or not (false).
//
// This is shorthand for:
//
//	htmxRequest, ok := r.Context().Value(middleware.HTMXRequestKey).(middleware.HTMXRequest)
func GetRequestHeaders(r *http.Request) (HTMXRequest, bool) {
	htmxRequest, ok := r.Context().Value(HTMXRequestKey).(HTMXRequest)
	return htmxRequest, ok
}
