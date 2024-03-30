# HX

A Go package for working with HTMX.

Currently implements helper functions for working with HTMX response headers.

## Examples

The primary method for setting HTMX headers is the `SetHeaders` function.
This function takes the response writer and any number of utility functions for setting HTMX headers.

When `SetHeaders` is called, it loops through all of the given utility functions and applies the headers to the response writer. If any of these utility functions returns an error, the loop is broken and the error is returned.

Many of the utility functions do not return errors, but consistently return a nil error instead.

```go
import "github.com/thisisthemurph/hx"

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
    _ := hx.SetHeaders(w, hx.Retarget("#new-target"), hx.Reswap(hx.SwapOuterHTML))
}
```
The above example shows the setting of the `HX-Retarget` and `HX-Reswap` headers, a common pattern for overwriting the `hx-target` and `hx-swap` attributes set in the request. Neither of these utility functions returns an error, so the error can be ignored in this case.

### Trigger

HTMX has three response headers that can be used to trigger events in the front end; `HX-Trigger`, `HX-Trigger-After-Settle`, and `HX-Trigger-After-Swap`. Please read the [official HTMX documentation](https://htmx.org/headers/hx-trigger/) to better understand these concepts.

In hx, there are 6 primary functions for setting HX trigger headers, two for each trigger header; `HX-Trigger`, `HX-Trigger-After-Settle`, and `HX-Trigger-After-Swap`.

The following three functions allow adding of event names that have no associated detail:

- hx.Trigger(eventNames ...string)
- hx.TriggerAfterSettle(eventNames ...string)
- hx.TriggerAfterSwap(eventNames ...string)

The following three functions allow adding of events with both a name and detail:

- hx.TriggerWithDetail(events ...TriggerEvent)
- hx.TriggerAfterSettleWithDetail(events ...TriggerEvent)
- hx.TriggerAfterSwapWithDetail(events ...TriggerEvent)

The following example shows how the first thee functions can be used:

```go
err := hx.SetHeaders(w, hx.Trigger("event1", "event2"))
```

The following example shows how the latter three functions can be used:

```go
event := hx.NewTriggerEvent("event1", myStruct)
err := hx.SetHeaders(w, hx.TriggerWithDetail(event))
```
