+++
Title = "Testing"
Description = "Unit-test reactive and wasmwrap bindings with jsdom — no browser, no headless driver."
Order = 390
Group = "Extensions (astwerk-live)"
Extension = true
+++

`GOOS=js GOARCH=wasm go test` runs under node, which has no DOM. The `jsdom`
package is a small fake DOM that astwerk's own tests run against, exported so
yours can too.

## Documentation

### Install

```go
//go:build js && wasm

package main

import "github.com/LukasDerBaum42/astwerk-live/jsdom"

func init() { jsdom.Install() }
```

`Install` defines `document`, location, history and `requestAnimationFrame` on
`globalThis`; calling it twice is harmless.

### Reset and NewElement

```go
func TestCounterIncrements(t *testing.T) {
	jsdom.Reset() // a clean document.body for this test
	root := wasmwrap.Wrap(jsdom.NewElement("div"))
	// mount your bindings into root, then assert against it
}
```

`Reset` restores a clean document between tests. `NewElement` makes a detached
element to mount bindings into.

### Running the tests

```sh
export PATH="$PATH:$(go env GOROOT)/lib/wasm"
GOOS=js GOARCH=wasm go test ./...
```

The toolchain's own wasm runner needs to be on `PATH`. The build constraint on
your test file keeps it out of ordinary host builds.

<details>
<summary>What the fake DOM implements</summary>

- elements: `tagName`, children, `appendChild`, `insertBefore`, `remove`,
  `cloneNode`, `textContent`, `innerHTML`, attributes, `classList`, `style`,
  `value`/`checked`, `getBoundingClientRect`;
- selectors, limited to `#id`, `.class`, `[attr]` and a bare tag name, for
  `querySelector`, `querySelectorAll` and `closest`;
- events: `addEventListener`, `removeEventListener`, and a dispatch helper;
- a 2D canvas context that records its calls;
- `location`, `history` and `requestAnimationFrame`, so the router has
  something to drive.

And extras a real DOM doesn't have, for assertions:

```go
el.listenerCount(type)  // number of handlers bound for an event
el._tree                // comma-joined text of the subtree, for list ordering
el._rect                // the object getBoundingClientRect returns
el.dispatch(type, prop) // fire an event with extra properties
el._ctxCalls            // recorded 2D canvas calls, as "fillRect(1,2,3,4)"
```

It's not a browser and it will never grow into one — layout, the CSS cascade,
focus and paint are out of scope by design. It answers "did my Go code reach
for the right property with the right arguments", which is the question a
binding test is actually asking. Anything that only a real DOM can settle needs
a real browser.

</details>

## Example

```go
//go:build js && wasm

package main

import (
	"testing"

	"github.com/LukasDerBaum42/astwerk-live/jsdom"
	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func init() { jsdom.Install() }

func TestCounterIncrements(t *testing.T) {
	jsdom.Reset()
	root := wasmwrap.Wrap(jsdom.NewElement("div"))

	count := reactive.Num(0)
	label := wasmwrap.Create("span")
	reactive.BindText(label, count)
	root.Append(label)

	count.Set(3)
	if got := label.Text(); got != "3" {
		t.Errorf("label = %q, want \"3\"", got)
	}
}
```

A signal, a binding, a write, an assertion — no browser, no headless driver, no
separate test runner.
