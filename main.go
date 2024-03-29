package hx

import (
	"encoding/json"
	"net/http"
	"strings"
)

type HeaderResponseWriter interface {
	Header() http.Header
}

type HeaderDecorator func(w HeaderResponseWriter) error

// SetHeaders sets custom HTTP headers in the provided http.ResponseWriter.
//
// This function applies one or more custom header decorators to the given response writer `w`.
// A decorator function modifies the response by adding or modifying HTTP headers as per the provided configuration.
//
// Parameters:
//
//	w: http.ResponseWriter - The response writer to which custom headers will be applied.
//	decorators: ...HeaderDecorator - Any number of HeaderDecorators for adding HTMX headers.
//	            Each HeaderDecorator is responsible for setting specific HTTP headers.
//
// Returns:
//
//	error: An error if any of the HeaderDecorators encounter an issue while setting the headers.
//	       It returns nil if all the decorators are applied successfully.
//		   Most HeaderDecorators always return a nill error.
//
// Example usage:
//
//	import "github.com/thisisthemurph/hx"
//	_ := hx.SetHeaders(w, hx.Retarget("/login"), hx.Reswap(hx.SwapOuterHTML))
//
// Note:
//
//	If an error is returned, the function will not add any of the remaining headers, but will leave all
//	previously set headers, it is your responsibility to remove these headers.
//
//	The order of decorators matters. Headers set by decorators earlier in the slice may be overwritten
//	by subsequent decorators.
func SetHeaders(w HeaderResponseWriter, funcs ...HeaderDecorator) error {
	for _, fn := range funcs {
		if err := fn(w); err != nil {
			return err
		}
	}
	return nil
}

// SetHeader returns a function for setting the header on the http.ResponseWriter.
// The returned function will always return a nil error.
func SetHeader(key, value string) HeaderDecorator {
	return func(w HeaderResponseWriter) error {
		w.Header().Set(key, value)
		return nil
	}
}

// Location allows you to do a client-side redirect that does not do a full page reload.
// https://htmx.org/headers/hx-location/
//
// Never returns an error.
func Location(location string) HeaderDecorator {
	return SetHeader("HX-Location", location)
}

// PushURL pushes a new url into the history stack.
// https://htmx.org/headers/hx-push-url/
func PushURL(url string) HeaderDecorator {
	return SetHeader("HX-Push-Url", url)
}

// PreventPushURL prevents the browser’s history from being updated by setting the HX-Push-Url header to "false".
// https://htmx.org/headers/hx-push-url/
func PreventPushURL() HeaderDecorator {
	return PushURL("false")
}

// Redirect can be used to do a client-side redirect to a new location.
// https://htmx.org/reference/#response_headers
//
// Never returns an error.
func Redirect(path string) HeaderDecorator {
	return SetHeader("HX-Redirect", path)
}

// Refresh forces the client-side to do a full refresh of the page.
// https://htmx.org/reference/#response_headers
func Refresh() HeaderDecorator {
	return SetHeader("HX-Refresh", "true")
}

// PreventRefresh prevents the client-side from doing a full refresh of the page by setting
// the HX-Refresh header to "false".
// https://htmx.org/reference/#response_headers
func PreventRefresh() HeaderDecorator {
	return SetHeader("HX-Refresh", "false")
}

// ReplaceURL replaces the current URL in the location bar.
// https://htmx.org/headers/hx-replace-url/
func ReplaceURL(url string) HeaderDecorator {
	return SetHeader("HX-Replace-Url", url)
}

// PreventReplaceURL prevents replacing the current URL in the location bar by setting the
// HX-Replace-Url header to "false".
// https://htmx.org/headers/hx-replace-url/
func PreventReplaceURL() HeaderDecorator {
	return ReplaceURL("false")
}

// Reselect a CSS selector that allows you to choose which part of the response is used to be swapped in.
// Overrides an existing hx-select on the triggering element
// https://htmx.org/reference/#response_headers
//
// Never returns an error.
func Reselect(selector string) HeaderDecorator {
	return SetHeader("HX-Reselect", selector)
}

// Reswap allows you to override how the response will be swapped.
// https://htmx.org/reference/#response_headers
//
// Never returns an error.
func Reswap(swap Swap) HeaderDecorator {
	return SetHeader("HX-Reswap", swap.String())
}

// Retarget a CSS selector that overrides the target of the content update to
// a different element on the page.
//
// Never returns an error.
func Retarget(target string) HeaderDecorator {
	return SetHeader("HX-Retarget", target)
}

// Trigger can be used to trigger client side actions on the target element within a response to HTMX.
// You can trigger a single event or as many uniquely named events as you would like.
//
// The header is determined by the value of the when parameter.
//
//   - TriggerImmediately -> HX-Trigger
//   - TriggerAfterSettle -> HX-Trigger-After-Settle
//   - TriggerAfterSwap   -> HX-Trigger-After-Swap
//
// If the header already includes values, these will be retained.
// If the current header value is a JSON object, the new headers will be added as part of the JSON object,
// rather than a list of comma separated event names; as with the TriggerWithDetail function.
//
// The returned function will return an error if existing headers require events to be JSON encoded and marshalling fails.
//
// Parameters:
//
//	when: TriggerHeader - Specifies which header should be used to trigger the event.
//	eventNames: ...string - Uniquely named events to be triggered.
//
// Example usage:
//
//	err := hx.SetHeaders(hx.Trigger(hx.TriggerAfterSettle, "myFirstEvent", "someOtherEvent"))
//
// Or passing a slice of events:
//
//	events := []string {"event1", "event2"}
//	err := hx.SetHeaders(hx.Trigger(hx.TriggerAfterSwap, events...))
//
// https://htmx.org/headers/hx-trigger/
func Trigger(when TriggerHeader, eventNames ...string) HeaderDecorator {
	return func(w HeaderResponseWriter) error {
		// eventMap := make(map[string]string)
		eventList := make([]string, 0)

		currentHeaderValues := w.Header().Get(when.String())
		if currentHeaderValues != "" {
			// If header has JSON data, the current event names must also be added as JSON.
			// Default to using the TriggerWithDetail function.
			var js interface{}
			err := json.Unmarshal([]byte(currentHeaderValues), &js)
			if err == nil {
				events := make([]TriggerEvent, 0)
				for _, event := range eventNames {
					events = append(events, TriggerEvent{
						Name:   event,
						Detail: nil,
					})
				}

				return TriggerWithDetail(when, events...)(w)
			}

			// If the data is not JSON, we must maintain the data and append the new
			for _, eventName := range strings.Split(currentHeaderValues, ",") {
				eventList = append(eventList, strings.TrimSpace(eventName))
			}
		}

		for _, eventName := range eventNames {
			eventList = append(eventList, strings.TrimSpace(eventName))
		}

		w.Header().Set(when.String(), strings.Join(eventList, ", "))
		return nil
	}
}

// TriggerWithDetail can be used to trigger client side actions on the target element within a response to HTMX.
// You can trigger a single event or as many uniquely named events as you would like.
//
// The header is determined by the value of the when parameter.
//
//   - TriggerImmediately -> HX-Trigger
//   - TriggerAfterSettle -> HX-Trigger-After-Settle
//   - TriggerAfterSwap   -> HX-Trigger-After-Swap
//
// If the header already includes values, these will be retained.
// If the current header value is a list of comma separated strings, these will be converted to
// JSON objects with null detail.
//
// The returned function will return an error if the provided detail cannot be serialized into JSON.
//
// Parameters:
//
//	when: TriggerHeader - Specifies which header should be used to trigger the event.
//	events: ...TriggerEvent - The events (name and detail) to be triggered.
//
// Example usage:
//
//	event := hx.NewTriggerEvent("eventName", myStruct)
//	err := hx.SetHeaders(hx.Trigger(hx.TriggerAfterSettle, event))
//
// https://htmx.org/headers/hx-trigger/
func TriggerWithDetail(when TriggerHeader, events ...TriggerEvent) HeaderDecorator {
	return func(w HeaderResponseWriter) error {
		triggerEvents := make(map[string]any)
		currentHeaderValue := w.Header().Get(when.String())

		// If the header already has events present, we want to maintain these.
		if currentHeaderValue != "" {
			// Attempt to parse the header value as a JSON object.
			if err := json.Unmarshal([]byte(currentHeaderValue), &triggerEvents); err != nil {
				// If not JSON, assume a comma separated list of event names.
				// Convert these to TriggerEvent structs.
				eventNames := strings.Split(currentHeaderValue, ",")
				for _, ev := range eventNames {
					eventName := strings.TrimSpace(ev)
					triggerEvents[eventName] = nil
				}
			}
		}

		// var triggerEvents = make(map[string]any)
		for _, event := range events {
			triggerEvents[event.Name] = event.Detail
		}

		data, err := json.Marshal(triggerEvents)
		if err != nil {
			return err
		}

		w.Header().Set(when.String(), string(data))
		return nil
	}
}
