# HX

A Go package for working with HTMX.

Currently implements helper functions for working with HTMX response headers.

## HTMX Headers

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

## HTMX Request Headers Middleware

If you would like to easily access the HTMX request headers, this can be done simply with the provided middleware.

```go
package main

import (
    "myproject/handler"
    "github.com/thisisthemurph/hx/middleware"
)


func main() {
    mux := http.NewServerMux()
    mux.Handle("/", middleware.WithHTMX(http.HandlerFunc(handler.LoginHandler)))

    http.ListenAndServe(":8080", mux)
}
```

In your handler you can access the HTMX request headers like so:

```go
package handler

import (
    "fmt"
    "github.com/thisisthemurph/hx/middleware"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    h, ok := middleware.GetRequestHeaders(r)
    if !ok {
        // This should should only happen if the middleware has not been configured
    }

    if h.IsHTMXRequest {
        fmt.Println("Hello HTMX!")
        fmt.Printf("The current URL is %s\n", h.CurrentURL)
    } else {
        fmt.Println("Hello standard request...")
    }
}
```

### Using a third-party framework such as Echo?

Using a third-party framework other than the standard library is as you would expect. The only difference is that the framework you are using may have a different method of setting up middleware and accessing the request. The following example demonstrates use with the Echo framework, but you should be able to figure out how to use this with your framework of choice.

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/thisisthemurph/hx/middleware"
)

func main() {
    e := echo.New()
    e.Use(middleware.WithHTMX)
    e.GET("/", handler.LoginHandler)

    e.start(":8080")
}
```

In your handler you can access the HTMX request headers like so:

```go
package handler

import (
    "fmt"
    "github.com/labstack/echo/v4"
    "github.com/thisisthemurph/hx/middleware"
)

func LoginHandler(c echo.Context) error {
    h, ok := middleware.GetRequestHeaders(c.Request())
    if !ok {
        // This should should only happen if the middleware has not been configured
    }

    if h.IsHTMXRequest {
        fmt.Println("Hello HTMX!")
        fmt.Printf("The current URL is %s\n", h.CurrentURL)
    } else {
        fmt.Println("Hello standard request...")
    }
}
```