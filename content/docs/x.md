+++
Title = "x — compiled reactive markup"
Description = "Markers that generate both the HTML and the JavaScript for a page at build time — no WASM, no framework."
Order = 110
Group = "Add interactivity"
+++

`x` is interactive UI without WASM and without a JavaScript framework. Markers
are ordinary Go functions that render where they are used, and each one owns the
element it produces — `@count` writes the value **and** the JavaScript that
keeps it live, in the same call. No build pass, no selector to keep in sync.

The idea — move reactivity to build time and ship the compiled result — is the
one Svelte made famous; `x` is the Go-shaped version, with no `.svelte` file
format and no Node toolchain. See the [comparison](../comparison/#astwerk-vs-svelte)
for the full contrast.

Everything below is opt-in: outside a `Document` the markers render plain
static HTML, so a page with no markers ships no script at all.

Markers that change the document's structure or run scripts of their own —
`Effect`, `When`, `List`, `Script` — are on [x — structure & scripts](../x-structure/).

## Documentation

### Signal

```go
var count = x.NamedNumSignal("count", 0) // named, so scripts can read count.get()
var open  = x.NewBoolSignal(false)
```

State is one of four split types, each exposing only the operations meaningful
for its value:

| Constructor | Kind |
| --- | --- |
| `NewNumSignal(v)` / `NamedNumSignal(name, v)` | `int`, `int64`, `float64` |
| `NewBoolSignal(v)` / `NamedBoolSignal(name, v)` | `bool` |
| `NewStrSignal(v)` / `NamedStrSignal(name, v)` | `string` |
| `NewListSignal(v)` / `NamedListSignal(name, v)` | `[]T` for a keyed list |

A signal holds a value that exists both as a build-time Go value and as a live
JavaScript state cell. `Named` fixes the signal's name in the generated
JavaScript; auto-named signals are fine for bindings, which never need to know
the name.

<details>
<summary>Why the name matters</summary>

The name is the bridge to hand-written JavaScript: a script can read and write
`count.get()` / `count.set(...)` because the generated code declares `count` as
a real variable. An auto-named signal has no stable name, so only the compiled
bindings can reach it.

</details>

### Text

```templ
@count
```

Renders the current value and emits the binding that keeps it in sync. A signal
is itself a component — `@count` and `@count.Text()` are the same thing. The
marker owns the `<span>`, so there is no selector to keep in sync by hand.

### El

```templ
@x.El("button", x.On("click", count.Inc(1))) {
	+
}
```

Renders the element with its bindings attached. Every binding owns one property
of one element.

### Bindings

```templ
@x.El("a", x.Attr("href", url))        { link }
@x.El("span", x.Class("on", active))   { … }
@x.El("button", x.Disabled(busy))      { save }
@x.El("p", x.Show(visible))            { … }
@x.El("p", x.Style("color", colour))   { … }
@x.El("button", x.On("click", action)) { + }
```

| Binding | Effect |
| --- | --- |
| `Attr(name, value)` | sets a string attribute, removes it when the value is empty |
| `Class(name, on)` | toggles one class |
| `Disabled(on)` | toggles the boolean attribute — present or absent, not true or false |
| `Show(on)` | toggles visibility |
| `Style(prop, value)` | sets one CSS property through `style.setProperty`, so a value can't break out of its property |
| `On(event, action)` | binds an event handler whose action is built from the signal methods |

### Actions

```go
count.SetTo(0)   // store a constant
count.Inc(1)     // add one (Inc and Dec take a literal or another signal)
count.Dec(step)  // subtract a signal's current value
open.Toggle()    // flip a bool
```

`Set(expr)` is the general form; `SetTo`, `Inc`, `Dec` and `Toggle` are the
one-word forms for the common cases. Handlers that need more than one
statement compose with `x.Seq(...)` and `x.If(...)` — see
[x — structure & scripts](../x-structure/).

### Expressions

```go
x.Add(price, 1)      // arithmetic: Add, Sub, Mul, Div
x.Mul(price, qty)    // … another signal of the same type
x.Concat(full, " Ltd.") // strings have their own combinator
x.Gt(count, 0)       // comparisons: Gt, Lt, Gte, Lte, Eq, Ne
x.Not(open)          // negation
```

The set is deliberately small — every operator is a step toward a language.
When an expression stops reading like the JavaScript it produces, that's the
boundary; use the [script escape hatch](../x-structure/#script) instead.

### Computed

```go
var price = x.NewNumSignal(10.0)
var qty   = x.NewNumSignal(2.0)
var total = x.Computed(x.Mul(price, qty))
```

```templ
@x.Number(price)
@x.Number(qty)
@x.Text(total)
```

Derived state is an expression built from combinators. It is evaluated at build
time to render the static value, and compiled to JavaScript that recomputes it
when its dependencies change.

### Input, Number, Checkbox

```templ
@x.Input(name)
@x.Number(age)
@x.Checkbox(subscribed)
```

All three are two-way — typing writes the signal, setting the signal updates the
field — without feeding back on themselves. `Number` ignores input that isn't a
number rather than clobbering the signal with zero.

<details>
<summary>Why two-way doesn't feed back</summary>

Typing fires an input event that writes the signal; the reverse direction is a
plain value assignment on the field. Assigning an input's `.value` does not
dispatch an input event, so the two directions never chase each other.

</details>

### Document

```templ
@x.Document(counter())
```

The page's entry point: it renders its body and emits the runtime and state
declarations before the markup, the bindings after it. Outside a `Document` the
markers render plain static HTML — which is what makes reactivity opt-in, and
why a page with no markers ships no script at all.

```templ
@x.Document(counter(), x.Minified()) // compact production output
```

Without `Minified`, the generated JavaScript is indented and readable.

<details>
<summary>How a marker works, and how this differs from reactive</summary>

`@x.Text(count)` renders `<span data-x="7">0</span>` into the page and appends
`aw.text("7", s0)` to a script buffer. The marker owns the span, so it can stamp
it with an internal id at the exact spot it was written — there is no second
pass and no way for the wiring to drift out of sync with the markup.

State that is never used in markup generates no JavaScript, and the runtime
itself is assembled from fragments — a counter page ships roughly the runtime
for a counter, not a framework.

`x` and the WASM `reactive` package are two implementations of the same
reactive model — signals with tracked dependencies, fine-grained bindings,
keyed lists. `x` is the lightweight default for the common 90%; `reactive` is
the complete path for anything that needs full Go in the browser — goroutines,
the standard library, the router.

</details>

## Accessibility

Markers render real elements, so native behaviour comes along: `x.El("button", …)`
is a real `<button>` with keyboard focus and Enter/Space activation — nothing to
fake with roles. For a toggle-style widget, mirror the state in `aria-pressed`
next to the `on` class; the site's own demos set both from the same effect, and
that is the pattern:

```templ
@x.El("button", x.On("click", open.Toggle()), x.Class("on", open)) {
	toggle
}
@x.Effect() {
	<script>
		document.querySelector("[data-toggle]").setAttribute("aria-pressed", String(open.get()));
	</script>
}
```

## Example

```go
var count = x.NamedNumSignal("count", 0)
```

```templ
@x.Document(counter())

templ counter() {
	<div>
		@count
		@x.El("button", x.On("click", count.Inc(1))) {
			+
		}
	</div>
}
```

The whole model in one component: named state, a text binding, and a handler
whose action is one word.
