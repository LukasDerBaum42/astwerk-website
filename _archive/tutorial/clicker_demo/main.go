//go:build js && wasm

// Command clicker-demo is quest 6 of the cookie clicker line: the same game
// the x markers built in steps 1–5, on the WASM path. The markup is written
// by hand on the page; this module hydrates it — reactive state for the
// counter, a wasmwrap.SetInterval for the auto-clicker, and Go's fmt for the
// formatted display. The x line kept those as named signals and a script;
// this is the honest "use Go" version, side by side with it.
package main

import (
	"fmt"

	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func main() {
	reactive.Hydrate("#clicker", func(root wasmwrap.Element) {
		cookies := reactive.Num(0.0)

		reactive.BindText(root.Find("[data-cookies]"), reactive.Computed(func() string {
			return fmtCookie(cookies.Get())
		}))
		reactive.On(root.Find("[data-bake]"), "click", func(wasmwrap.Event) {
			cookies.Inc(1)
		})

		// The same timer the x line kept as a script — here it is a
		// wasmwrap.SetInterval in Go.
		wasmwrap.SetInterval(func() { cookies.Inc(1) }, 1000)
	})

	// Keep the module alive: main returning would tear down every handler.
	select {}
}

// fmtCookie formats the counter — Go's strconv, the whole stdlib on tap.
func fmtCookie(n float64) string {
	if n >= 1e6 {
		return fmt.Sprintf("%.1fM", n/1e6)
	}
	if n >= 1e3 {
		return fmt.Sprintf("%.1fk", n/1e3)
	}
	return fmt.Sprintf("%.0f", n)
}
