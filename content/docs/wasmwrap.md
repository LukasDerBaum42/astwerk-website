+++
Title = "wasmwrap"
Description = "syscall/js made to feel like Go: DOM, style, events, fetch."
Order = 55
Group = "Client-side"
+++

`syscall/js` works, but every call is a stringly-typed `js.Value.Call` and
reading it feels like looking at JavaScript through a keyhole. `wasmwrap` covers
what you actually do in a browser with normal Go types. It exists only under
`GOOS=js GOARCH=wasm`.

Drawing is on [wasmwrap — canvas](../wasmwrap-canvas/).

## Documentation

### Finding elements

```go
nav := wasmwrap.Query("#nav")     // first match for a selector
rows := wasmwrap.QueryAll(".row") // every match
li := wasmwrap.Create("li")       // a detached element to build into
wasmwrap.Body()                   // document.body
wasmwrap.Doc()                    // the document itself
```

`QueryAll` returns a slice; the rest return a single `Element`. A failed
`Query` is inert rather than an error — see [Missing elements are
inert](#missing-elements-are-inert).

### Reading and writing

```go
el.SetText("hi").SetHTML("<b>raw</b>").SetAttr("data-id", "7")
text := el.Text()
html := el.HTML()
el.RemoveAttr("disabled")
value := el.InputValue()
```

`Text`/`SetText` handle text content, `HTML`/`SetHTML` raw markup (never user
input), `Attr`/`SetAttr`/`RemoveAttr` attributes, and `InputValue` the current
value of an input. Every setter returns its receiver, so calls chain.

<details>
<summary>Why setters return their receiver</summary>

Chaining is what makes the wrapper read like a sentence: each call hands the
element back so the next call can continue. Anything that would otherwise have
nothing to return returns the receiver instead.

</details>

### Structure

```go
el.Append(child)
el.InsertBefore(child, before)
parent := el.Parent()
children := el.Children()
first := el.ChildAt(0)
el.Remove()       // detach from its parent
clone := el.Clone()
el.Clear()        // remove all children
```

`Find` also works from an element: `el.Find(".row")` searches the subtree.

### Classes

```go
el.Class().Add("open").Remove("closed").Toggle("open")
el.Class().Has("open")
```

### Style

```go
el.Style().Color("red").Display("block").Set("gap", "1rem")
gap := el.Style().Get("gap")
width := el.Computed("width") // getComputedStyle, not the inline style
```

`Style` covers `Display`, `Color`, `Background`, `Opacity`, `Overflow`, `Size`,
`Position`, `Hide`, `Show` — each a thin `Set` — with `Set` for anything else
and `Computed` for the resolved value.

### Events

```go
off := btn.On("click", handler)
btn.Once("submit", handler) // fires once, then unbinds
```

`On` returns a function that unbinds the handler **and** releases the
underlying `js.Func`. For a handler bound once for the page's lifetime you can
ignore it; for anything bound repeatedly — per row of a re-rendered list —
dropping it leaks a `js.Func` every time. `reactive.On` wires this into effect
cleanup automatically.

```go
func handler(e wasmwrap.Event) {
	e.PreventDefault()
	e.StopPropagation()
	target := e.Target()
	key := e.Key()
	x, y := e.Pos()
}
```

`Event` carries `PreventDefault`, `StopPropagation`, `Target`, `Type`, `Key`,
`Pos` and `PosIn` (coordinates relative to an element).

### Timers

```go
cancel := wasmwrap.SetTimeout(func() { … }, 500)
stop := wasmwrap.SetInterval(func() { … }, 1000)
```

Both return a function that cancels the timer.

### Fetch

```go
body, err := wasmwrap.Fetch("/api/thing")
text, err := wasmwrap.FetchString("/api/thing")
```

Both block the calling goroutine, which is safe from `main` or any goroutine you
start: when every goroutine is blocked, the Go runtime hands control back to the
JavaScript event loop. They are **not** safe directly inside an event handler —
handlers run on a callback goroutine the event loop is waiting on, so start one:

```go
btn.On("click", func(wasmwrap.Event) {
	go func() {
		body, err := wasmwrap.Fetch("/api/thing")
		// …
	}()
})
```

<details>
<summary>Why blocking is safe — and where it isn't</summary>

On js/wasm the Go runtime and the browser's event loop share one thread. When
every goroutine is blocked on a promise, the runtime yields to the event loop
and the page keeps responding. But an event handler runs *on* a callback
goroutine that the event loop is waiting for — blocking there deadlocks the
page. That's the one place a `go func()` is required.

</details>

### Value — the escape hatch

```go
el.Value().Call("scrollIntoView", map[string]any{"behavior": "smooth"})
```

Anything not modelled — an obscure API, a browser quirk — drops to the raw
value. `Element.Value()`, `Event.Value()` and `Ctx2D.Value()` all return the
underlying `js.Value`, and `wasmwrap.Wrap` takes one back. The wrapper is a
shortcut for the common ninety percent, not a replacement for the platform.

<details>
<summary>Missing elements are inert</summary>

A failed `Query` — `querySelector` returned null — gives an element where
`Exists()` is false. Readers return zero values and `Remove` is a no-op, so a
missing element is inert rather than a panic three calls later.

</details>

## Example

```go
//go:build js && wasm

package main

import "github.com/LukasDerBaum42/astwerk/wasmwrap"

func main() {
	nav := wasmwrap.Query("#nav")
	wasmwrap.Query("#toggle").On("click", func(e wasmwrap.Event) {
		e.PreventDefault()
		nav.Class().Toggle("open")
	})
	select {} // keep the module alive
}
```

A nav toggle in three lines. `select {}` matters: when `main` returns, the WASM
instance exits and every handler dies with it.
