//go:build js && wasm

// The bridge demo, driven under the jsdom fake DOM with the real x runtime:
// render the actual page (markup + emitted scripts), eval the x program,
// mirror the named cell with XNum, and click both sides. This is the
// reproduction for the "wasm doesn't affect x" report — and the
// minified-runtime check that caught the minifier swallowing the signal's
// set method behind a // comment.
package main

import (
	"context"
	"regexp"
	"strings"
	"syscall/js"
	"testing"

	"github.com/LukasDerBaum42/astwerk-live/jsdom"
	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
	"github.com/LukasDerBaum42/astwerk-website/pages/examples"
	"github.com/LukasDerBaum42/astwerk-website/templates"
	"github.com/LukasDerBaum42/astwerk/ssg"
)

func init() { jsdom.Install() }

var scriptRe = regexp.MustCompile(`(?s)<script>\n(.*?)</script>`)

// renderBridge renders the real bridge page with the given MinifiedX setting
// and returns the page HTML.
func renderBridge(t *testing.T, min bool) string {
	t.Helper()
	templates.MinifiedX = min
	defer func() { templates.MinifiedX = false }()

	var buf strings.Builder
	if err := examples.ExampleBridge(ssg.Ctx{Title: "Bridge"}, nil, nil).Render(context.Background(), &buf); err != nil {
		t.Fatalf("render bridge page: %v", err)
	}
	return buf.String()
}

// evalHeader runs the runtime and declaration script — the part that must
// exist before the markup is mounted, because its bindings capture the
// runtime's closure. Top-level `const shared = …` becomes `var` so the footer
// eval can see the declaration, which mirrors how browser script elements
// share the global lexical scope.
func evalHeader(t *testing.T, html string) {
	t.Helper()
	for _, m := range scriptRe.FindAllStringSubmatch(html, -1) {
		body := m[1]
		if !strings.Contains(body, "signal(") || strings.Contains(body, "aw.on(") || strings.Contains(body, "aw.text(") {
			continue // the footer's bindings, or a layout script — not the header
		}
		body = regexp.MustCompile(`(?m)^const (shared|s\d+) = `).ReplaceAllString(body, "var $1 = ")
		js.Global().Call("eval", body)
	}
}

// evalFooter runs the binding script — after the markup exists, so the
// el() lookups in aw.on/aw.text find their nodes. Eval'ing it earlier makes
// the handlers silently skip (the `if (n)` guards), which is exactly the
// mount-order flaw this harness was written to rule out.
func evalFooter(t *testing.T, html string) {
	t.Helper()
	for _, m := range scriptRe.FindAllStringSubmatch(html, -1) {
		body := m[1]
		if !strings.Contains(body, "aw.on(") && !strings.Contains(body, "aw.text(") {
			continue
		}
		js.Global().Call("eval", body)
	}
}

// mountXSide builds the x half of the page — the same markup bridgeXDemo
// ships: a dec button, the value span, an inc button, each stamped with its
// data-x id in build order — and returns the value span and the inc button.
func mountXSide() (out, inc wasmwrap.Element) {
	jsdom.Reset()
	row := wasmwrap.Wrap(jsdom.NewElement("div"))
	dec := wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-x", "0")
	out = wasmwrap.Wrap(jsdom.NewElement("span")).SetAttr("data-x", "1").SetText("0")
	inc = wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-x", "2")
	row.Append(dec, out, inc)
	wasmwrap.Body().Append(row)
	return out, inc
}

// mountWasmSide builds the wasm half's markup and hydrates it with the real
// bridge logic (XNum mirror over the named cell).
func mountWasmSide() (out, inc wasmwrap.Element) {
	widget := wasmwrap.Wrap(jsdom.NewElement("div")).SetAttr("id", "bridge-wasm")
	out = wasmwrap.Wrap(jsdom.NewElement("span")).SetAttr("data-bridge-count", "").SetText("0")
	dec := wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-bridge-dec", "")
	inc = wasmwrap.Wrap(jsdom.NewElement("button")).SetAttr("data-bridge-inc", "")
	widget.Append(out, dec, inc)
	wasmwrap.Body().Append(widget)

	shared := reactive.XNum[int]("shared")
	reactive.BindText(widget.Find("[data-bridge-count]"), shared)
	reactive.On(widget.Find("[data-bridge-dec]"), "click", func(wasmwrap.Event) { shared.Dec(1) })
	reactive.On(widget.Find("[data-bridge-inc]"), "click", func(wasmwrap.Event) { shared.Inc(1) })
	return out, inc
}

// TestBridgeBothDirections is the user's report, mechanised: the x side
// updates the wasm side, and the wasm side updates the x side.
func TestBridgeBothDirections(t *testing.T) {
	html := renderBridge(t, false) // readable runtime, as the dev server ships
	evalHeader(t, html)
	xOut, xInc := mountXSide()
	wasmOut, wasmInc := mountWasmSide()
	evalFooter(t, html)

	// x → wasm
	xInc.Value().Call("dispatch", "click", js.Null())
	if got := wasmOut.Text(); got != "1" {
		t.Errorf("wasm side after x click = %q, want 1", got)
	}

	// wasm → x
	wasmInc.Value().Call("dispatch", "click", js.Null())
	if got := xOut.Text(); got != "2" {
		t.Errorf("x side after wasm click = %q, want 2", got)
	}
}

// TestMinifiedRuntimeKeepsSet is the production check: the minified runtime
// must still carry the signal's set method — the regression test for the
// minifier swallowing set: behind a // comment.
func TestMinifiedRuntimeKeepsSet(t *testing.T) {
	html := renderBridge(t, true)
	evalHeader(t, html)
	xOut, xInc := mountXSide()
	wasmOut, wasmInc := mountWasmSide()
	evalFooter(t, html)

	xInc.Value().Call("dispatch", "click", js.Null())
	if got := wasmOut.Text(); got != "1" {
		t.Errorf("minified: wasm side after x click = %q, want 1 (x set must work)", got)
	}
	wasmInc.Value().Call("dispatch", "click", js.Null())
	if got := xOut.Text(); got != "2" {
		t.Errorf("minified: x side after wasm click = %q, want 2 (cell set must work)", got)
	}
}
