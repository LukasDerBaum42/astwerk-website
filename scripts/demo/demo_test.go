//go:build js && wasm

// The site starts testing what it ships: demo.go's widgets, driven under the
// jsdom fake DOM — the same recipe the docs' Testing page teaches.
//
//	GOOS=js GOARCH=wasm go test ./scripts
//
// demo.go's main is never run by the test binary; the demos' setup functions
// are called directly, exactly as main would.
package main

import (
	"syscall/js"
	"testing"

	"github.com/LukasDerBaum42/astwerk-live/jsdom"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func init() { jsdom.Install() }

// mountCounter builds the counter demo's markup — the same shape the Examples
// page ships — into the jsdom document, then hydrates it with counter().
func mountCounter(t *testing.T) (out, inc, dec, reset wasmwrap.Element) {
	t.Helper()
	jsdom.Reset()

	widget := wasmwrap.Wrap(jsdom.NewElement("div")).SetAttr("id", "counter")
	out = wasmwrap.Wrap(jsdom.NewElement("output")).SetAttr("data-count", "").SetText("0")
	doubled := wasmwrap.Wrap(jsdom.NewElement("strong")).SetAttr("data-doubled", "").SetText("0")
	dec = wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-dec", "")
	inc = wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-inc", "")
	reset = wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-reset", "")
	widget.Append(out, doubled, dec, inc, reset)
	wasmwrap.Body().Append(widget)

	counter()
	return out, inc, dec, reset
}

func TestCounterIncrements(t *testing.T) {
	out, inc, _, _ := mountCounter(t)
	if got := out.Text(); got != "0" {
		t.Fatalf("count = %q, want 0", got)
	}

	inc.Value().Call("dispatch", "click", js.Null())
	if got := out.Text(); got != "1" {
		t.Errorf("count = %q after one click, want 1", got)
	}
}

func TestCounterDecAndReset(t *testing.T) {
	out, inc, dec, reset := mountCounter(t)

	inc.Value().Call("dispatch", "click", js.Null())
	inc.Value().Call("dispatch", "click", js.Null())
	if got := out.Text(); got != "2" {
		t.Fatalf("count = %q after two clicks, want 2", got)
	}

	dec.Value().Call("dispatch", "click", js.Null())
	if got := out.Text(); got != "1" {
		t.Errorf("count = %q after decrement, want 1", got)
	}

	reset.Value().Call("dispatch", "click", js.Null())
	if got := out.Text(); got != "0" {
		t.Errorf("count = %q after reset, want 0", got)
	}
}
