package hx

import "fmt"

const (
	HeaderLocation           = "HX-Location"             // HX-Location allows you to do a client-side redirect that does not do a full page reload.
	HeaderPushURL            = "HX-Push-Url"             // HX-Push-Url pushes a new url into the history stack.
	HeaderRedirect           = "HX-Redirect"             // HX-Redirect can be used to do a client-side redirect to a new location.
	HeaderRefresh            = "HX-Refresh"              // HX-Refresh if set to “true” the client-side will do a full refresh of the page.
	HeaderReplaceURL         = "HX-Replace-Url"          // HX-Replace-Url replaces the current URL in the location bar.
	HeaderReswap             = "HX-Reswap"               // HX-Reswap allows you to specify how the response will be swapped. See hx-swap for possible values.
	HeaderRetarget           = "HX-Retarget"             // HX-Retarget sets a CSS selector that updates the target of the content update to a different element on the page.
	HeaderReselect           = "HX-Reselect"             // HX-Reselect sets a CSS selector that allows you to choose which part of the response is used to be swapped in. Overrides an existing hx-select on the triggering element.
	HeaderTrigger            = "HX-Trigger"              // HX-Trigger header triggers events as soon as the response is received.
	HeaderTriggerAfterSettle = "HX-Trigger-After-Settle" // HX-Trigger-After-Settle triggers events after the settle step.
	HeaderTriggerAfterSwap   = "HX-Trigger-After-Swap"   // HX-Trigger-After-Swap triggers events after the swap step.
)

// Swap represents the type of content swap method used in HTMX.
// It enumerates different ways in which content can be swapped on the client-side
// without a full page reload. Each swap method has its own meaning and effect
// on how content is updated or manipulated.
//
// For more information see: https://htmx.org/attributes/hx-swap/
type Swap int

const (
	SwapInnerHTML Swap = iota
	SwapOuterHTML
	SwapBeforeBegin
	SwapAfterBegin
	SwapBeforeEnd
	SwapAfterEnd
	SwapDelete
	SwapNone
)

// String returns a string representation of the Swap value.
// If the Swap value is not recognized, it returns "innerHTML" by default.
func (s Swap) String() string {
	switch s {
	case SwapOuterHTML:
		return "outerHTML"
	case SwapBeforeBegin:
		return "beforebegin"
	case SwapAfterBegin:
		return "afterbegin"
	case SwapBeforeEnd:
		return "beforeend"
	case SwapAfterEnd:
		return "afterend"
	case SwapDelete:
		return "delete"
	case SwapNone:
		return "none"
	default:
		return "innerHTML"
	}
}

// SwapFromString converts a string representation to a Swap value.
// If the provided string does not match any known Swap values, it returns SwapInnerHTML by default
// along with an error indicating the invalid string value.
func SwapFromString(s string) (Swap, error) {
	switch s {
	case "innerHTML":
		return SwapInnerHTML, nil
	case "outerHTML":
		return SwapOuterHTML, nil
	case "beforebegin":
		return SwapBeforeBegin, nil
	case "afterbegin":
		return SwapAfterBegin, nil
	case "beforeend":
		return SwapBeforeEnd, nil
	case "afterend":
		return SwapAfterEnd, nil
	case "delete":
		return SwapDelete, nil
	case "none":
		return SwapNone, nil
	default:
		return SwapInnerHTML, fmt.Errorf("invalid Swap value: %q", s)
	}
}

// StringToSwap converts a string representation to a Swap value.
// If the provided string does not match any known Swap values, it returns SwapInnerHTML by default
// along with an error indicating the invalid string value.
func StringToSwap(s string) (Swap, error) {
	return SwapFromString(s)
}

// TriggerEvent represents an event to be added to one of the following trigger headers:
//
//   - HX-Trigger
//   - HX-Trigger-After-Settle
//   - HX-Trigger-After-Swap
type TriggerEvent struct {
	Name   string // Name of the event to be triggered.
	Detail any    // Detail associated with the event.
}

// NewTriggerEvent creates a new TriggerEvent struct with the given name and detail.
func NewTriggerEvent(name string, detail any) TriggerEvent {
	return TriggerEvent{
		Name:   name,
		Detail: detail,
	}
}
