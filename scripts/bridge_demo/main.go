//go:build js && wasm

// Command bridge-demo is the shared-cell demo: one named x signal, two
// runtimes. The page declares x.NamedNumSignal("shared", 0) and renders its
// own counter with compiled markup; this module mirrors that cell with
// reactive.XNum and renders the same counter in Go. Clicking either side
// updates both — the terminology matrix's proof that x and reactive share one
// state model.
package main

import (
	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func main() {
	reactive.Hydrate("#bridge-wasm", func(root wasmwrap.Element) {
		shared := reactive.XNum[int]("shared")

		reactive.BindText(root.Find("[data-bridge-count]"), shared)
		reactive.On(root.Find("[data-bridge-dec]"), "click", func(wasmwrap.Event) {
			shared.Dec(1)
		})
		reactive.On(root.Find("[data-bridge-inc]"), "click", func(wasmwrap.Event) {
			shared.Inc(1)
		})
	})

	// Keep the module alive: main returning would tear down every handler.
	select {}
}
