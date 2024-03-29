package hx_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thisisthemurph/hx"
)

func TestReswap(t *testing.T) {
	w := httptest.NewRecorder()
	swap := hx.SwapAfterBegin
	err := hx.SetHeaders(w, hx.Reswap(swap))

	if err != nil {
		t.Errorf("Refresh returned an unexpected error: %v", err)
	}

	header := w.Header().Get("HX-Reswap")
	headerSwap, _ := hx.StringToSwap(header)
	if headerSwap != swap {
		t.Errorf("Expected header HX-Refresh to have value %s, got %s", swap, header)
	}
}

func TestSetHeader(t *testing.T) {
	testCases := []struct {
		key   string
		value string
	}{
		{
			key:   "HX-Reswap",
			value: "outerHTML",
		}, {
			key:   "HX-Retarget",
			value: "form#login",
		}, {
			key:   "HX-Redirect",
			value: "/login",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s: %s", tc.key, tc.value), func(t *testing.T) {
			w := httptest.NewRecorder()

			fn := hx.SetHeader(tc.key, tc.value)
			err := fn(w)

			headerValue := w.Header().Get(tc.key)

			assert.NoError(t, err)
			assert.Equal(t, tc.value, headerValue)
		})
	}

}

func TestTrigger(t *testing.T) {
	w := httptest.NewRecorder()

	fn := hx.Trigger(hx.TriggerAfterSwap, "event1", "event2")
	err := fn(w)

	headerValue := w.Header().Get(hx.TriggerAfterSwap.String())
	assert.NoError(t, err)
	assert.Equal(t, "event1, event2", headerValue)

	// Ensure previous headers are not overwritten
	fn = hx.Trigger(hx.TriggerAfterSwap, "event3")
	err = fn(w)

	headerValue = w.Header().Get(hx.TriggerAfterSwap.String())
	assert.NoError(t, err)
	assert.Equal(t, "event1, event2, event3", headerValue)
}

func TestTrigger_RetainsPreviousEvents(t *testing.T) {
	testCases := []struct {
		name           string
		previousHeader string
		events         []string
		expected       string
	}{
		{
			previousHeader: "event1",
			events:         []string{"event2"},
			expected:       "event1, event2",
		}, {
			previousHeader: "event1, event2",
			events:         []string{"event3"},
			expected:       "event1, event2, event3",
		}, {
			previousHeader: "event1,event2",
			events:         []string{"event3"},
			expected:       "event1, event2, event3",
		}, {
			previousHeader: "event1, event2",
			events:         []string{"event3", "event4"},
			expected:       "event1, event2, event3, event4",
		}, {
			previousHeader: "",
			events:         []string{"event1"},
			expected:       "event1",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test #%d", i), func(t *testing.T) {
			w := httptest.NewRecorder()

			if tc.previousHeader != "" {
				w.Header().Set(hx.TriggerAfterSwap.String(), tc.previousHeader)
			}

			fn := hx.Trigger(hx.TriggerAfterSwap, tc.events...)
			err := fn(w)

			headerValue := w.Header().Get(hx.TriggerAfterSwap.String())
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, headerValue)
		})
	}

}

func TestTriggerWithDetail(t *testing.T) {
	testCases := []struct {
		name   string
		when   hx.TriggerHeader
		events []hx.TriggerEvent
	}{
		{
			name: "with string detail",
			when: hx.TriggerAfterSettle,
			events: []hx.TriggerEvent{
				hx.NewTriggerEvent("stringEvent", "event1-detail"),
			},
		}, {
			name: "with struct detail",
			when: hx.TriggerAfterSwap,
			events: []hx.TriggerEvent{
				hx.NewTriggerEvent(
					"structEvent",
					struct {
						Message string `json:"msg"`
						Level   int    `json:"level"`
					}{
						Message: "this is only a test",
						Level:   1,
					},
				),
			},
		}, {
			name: "with slice detail",
			when: hx.TriggerAfterSwap,
			events: []hx.TriggerEvent{
				hx.NewTriggerEvent("sliceEvent", []string{"a", "b", "c"}),
			},
		}, {
			name: "with multiple events",
			when: hx.TriggerImmediately,
			events: []hx.TriggerEvent{
				{
					Name: "event1",
					Detail: struct {
						Message string `json:"msg"`
						Level   int    `json:"level"`
					}{
						Message: "message number 1",
						Level:   1,
					},
				}, {
					Name: "event2",
					Detail: struct {
						Message string `json:"msg"`
						Level   int    `json:"level"`
					}{
						Message: "message number 2",
						Level:   2,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			fn := hx.TriggerWithDetail(tc.when, tc.events...)
			err := fn(w)

			expectedMap := make(map[string]any)
			for _, ev := range tc.events {
				expectedMap[ev.Name] = ev.Detail
			}
			expectedJSON, _ := json.Marshal(expectedMap)

			headerValue := w.Header().Get(tc.when.String())
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), headerValue)

			// Ensure previous events are not overwritten
			var newEvents []hx.TriggerEvent
			for _, ev := range tc.events {
				newEvents = append(newEvents, hx.TriggerEvent{
					Name:   ev.Name + "_new",
					Detail: ev.Detail,
				})
			}

			fn = hx.TriggerWithDetail(tc.when, newEvents...)
			err = fn(w)

			// expectedMap = make(map[string]any)
			for _, ev := range newEvents {
				expectedMap[ev.Name] = ev.Detail
			}
			expectedJSON, _ = json.Marshal(expectedMap)

			headerValue = w.Header().Get(tc.when.String())
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), headerValue)
		})
	}
}

func TestTriggerWithDetail_RetainsEventsIfAlreadyPresent(t *testing.T) {
	w := httptest.NewRecorder()
	existingEventJSON := "{\"event1\":{\"msg\":\"this is only a test\"}}"
	w.Header().Set(hx.TriggerImmediately.String(), existingEventJSON)

	event := hx.NewTriggerEvent("event2", "event2-data")
	fn := hx.TriggerWithDetail(hx.TriggerImmediately, event)
	err := fn(w)

	headerValue := w.Header().Get(hx.TriggerImmediately.String())
	expectedHeaderValue := "{\"event1\":{\"msg\":\"this is only a test\"},\"event2\":\"event2-data\"}"

	assert.NoError(t, err)
	assert.Equal(t, expectedHeaderValue, headerValue)

}

func TestTriggerWithDetail_ConvertsEventNamesIfAlreadyPresent(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	w.Header().Set(hx.TriggerImmediately.String(), "event1, event2, event3")

	event := hx.TriggerEvent{
		Name:   "event4",
		Detail: "this is the detail",
	}

	// Act
	fn := hx.TriggerWithDetail(hx.TriggerImmediately, event)
	err := fn(w)

	// Assert
	headerValue := w.Header().Get(hx.TriggerImmediately.String())
	expectedJSON := "{\"event1\":null,\"event2\":null,\"event3\":null,\"event4\":\"this is the detail\"}"
	assert.NoError(t, err)
	assert.JSONEq(t, expectedJSON, headerValue)
}
