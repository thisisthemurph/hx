package hx_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thisisthemurph/hx"
)

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

	fn := hx.Trigger("event1", "event2")
	err := fn(w)

	headerValue := w.Header().Get("HX-Trigger")
	assert.NoError(t, err)
	assert.Equal(t, "event1, event2", headerValue)
}

func TestTrigger_RetainsPreviousEvents(t *testing.T) {
	testCases := []struct {
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
				w.Header().Set("HX-Trigger", tc.previousHeader)
			}

			fn := hx.Trigger(tc.events...)
			err := fn(w)

			headerValue := w.Header().Get("HX-Trigger")
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, headerValue)
		})
	}

}

func TestTrigger_UsesEventWithDetailIfExistingEventsAreJSON(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set(hx.HeaderTrigger, "{\"msg\":\"some message\"}")

	fn := hx.Trigger("new-event")
	err := fn(w)

	expectedHeaderValue := "{\"msg\":\"some message\",\"new-event\":null}"
	assert.NoError(t, err)
	assert.JSONEq(t, expectedHeaderValue, w.Header().Get(hx.HeaderTrigger))
}

func TestTriggerWithDetail(t *testing.T) {
	testCases := []struct {
		name   string
		events []hx.TriggerEvent
	}{
		{
			name: "with string detail",
			events: []hx.TriggerEvent{
				hx.NewTriggerEvent("stringEvent", "event1-detail"),
			},
		}, {
			name: "with struct detail",
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
			events: []hx.TriggerEvent{
				hx.NewTriggerEvent("sliceEvent", []string{"a", "b", "c"}),
			},
		}, {
			name: "with multiple events",
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

			fn := hx.TriggerWithDetail(tc.events...)
			err := fn(w)

			expectedMap := make(map[string]any)
			for _, ev := range tc.events {
				expectedMap[ev.Name] = ev.Detail
			}
			expectedJSON, _ := json.Marshal(expectedMap)

			headerValue := w.Header().Get("HX-Trigger")
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedJSON), headerValue)
		})
	}
}

func TestTriggerWithDetail_RetainsEventsIfAlreadyPresent(t *testing.T) {
	w := httptest.NewRecorder()
	existingEventJSON := "{\"event1\":{\"msg\":\"this is only a test\"}}"
	w.Header().Set(hx.HeaderTrigger, existingEventJSON)

	event := hx.NewTriggerEvent("event2", "event2-data")
	fn := hx.TriggerWithDetail(event)
	err := fn(w)

	headerValue := w.Header().Get("HX-Trigger")
	expectedHeaderValue := "{\"event1\":{\"msg\":\"this is only a test\"},\"event2\":\"event2-data\"}"

	assert.NoError(t, err)
	assert.JSONEq(t, expectedHeaderValue, headerValue)

}

func TestTriggerWithDetail_ConvertsEventNamesIfAlreadyPresent(t *testing.T) {
	headers := []string{
		hx.HeaderTrigger,
		hx.HeaderTriggerAfterSettle,
		hx.HeaderTriggerAfterSwap,
	}

	testCases := []struct {
		name                string
		existingEvents      string
		expectedHeaderValue map[string]any
	}{
		{
			name:           "single existing event",
			existingEvents: "event1",
			expectedHeaderValue: map[string]any{
				"event1":    nil,
				"new-event": "this is the detail",
			},
		}, {
			name:           "multiple existing events",
			existingEvents: "event1, event2, event3",
			expectedHeaderValue: map[string]any{
				"event1":    nil,
				"event2":    nil,
				"event3":    nil,
				"new-event": "this is the detail",
			},
		},
	}

	for _, header := range headers {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s %s", header, tc.name), func(t *testing.T) {
				w := httptest.NewRecorder()
				w.Header().Set(header, tc.existingEvents)

				event := hx.TriggerEvent{
					Name:   "new-event",
					Detail: "this is the detail",
				}

				var err error
				switch header {
				case hx.HeaderTriggerAfterSettle:
					err = hx.TriggerAfterSettleWithDetail(event)(w)
				case hx.HeaderTriggerAfterSwap:
					err = hx.TriggerAfterSwapWithDetail(event)(w)
				default:
					err = hx.TriggerWithDetail(event)(w)
				}

				assert.NoError(t, err)

				var actualHeaderValue map[string]any
				err = json.Unmarshal([]byte(w.Header().Get(header)), &actualHeaderValue)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedHeaderValue, actualHeaderValue)
			})
		}
	}
}
