+++
Title = "x — compiled reactive markup"
Description = "Markers that generate both the HTML and the JavaScript for a page at build time — no WASM, no framework."
Order = 45
Group = "Client-side"
+++

`x` is astwerk's compiled reactive markup: interactive UI without WASM and
without a JavaScript framework. Markers are ordinary Go functions that render
at the exact spot they are used, and each one owns the element it produces —
`@x.Text(count)` writes the value *and* the JavaScript that keeps it in sync,
in the same call. There is no build pass, no transpiler, nothing to re-render
client-side, and no selector to keep in sync by hand.

## A counter

The whole model in one component:

```go
var count = x.NewSignal(0)
```

```templ
templ Counter(title string, path string, prefix string) {
	@layout.BaseLayout(title, path, prefix) {
		@x.Document(counter())
	}
}

templ counter() {
	<div>
		@x.Text(count)
		@x.El("button", x.On("click", count.Set(count.Add(1)))) {
			+
		}
	</div>
}
```

`x.Document` is the page's entry point: it renders its body and emits the
generated JavaScript around it — a small runtime and the state declarations
before the markup, the bindings after it. This page's markup stays one source:
`<span>0</span><button>+</button>` with a `<script>` behind it that keeps them
alive. View source on any example page to see it.

Outside a `Document` the markers render plain static HTML, which is what makes
reactivity opt-in: a page reads fine with JavaScript disabled, and a page with
no markers ships no script at all.

## How a marker works

`@x.Text(count)` renders `<span data-x="7">0</span>` into the page and appends
`aw.text("7", s0)` to the script buffer hanging off the context. The marker
owns the span, so it can stamp it with an internal id at the exact spot it was
written — there is no second pass and no way for the wiring to drift out of
sync with the markup.

The same marker set covers the other spots a value can live:

```templ
@x.El("a", x.Attr("href", url)) {
	link text
}
@x.El("button", x.Disabled(busy)) {
	save
}
@x.El("p", x.Class("on", active), x.Show(visible)) {
	…
}
@x.El("p", x.Style("color", colour)) {
	…
}
@x.El("button", x.On("click", handler)) {
	+
}
```

Each binding renders its static contribution to the element's start tag and
records its wiring against the element's id. `Attr` removes the attribute when
the value is empty; `Class` toggles one class; `Show` toggles visibility;
`Style` sets one CSS property through `style.setProperty`, so a value cannot
break out of its property.

## Forms

Two-way form fields bind in both directions without feeding back on
themselves — typing writes the signal, setting the signal updates the field:

```templ
@x.Input(name)
@x.Number(age)
@x.Checkbox(subscribed)
```

`Number` ignores input that isn't a number rather than clobbering the signal
with zero.

## Computed values

Derived state is an expression built from combinators. It is evaluated at
build time to render the static HTML, and compiled to JavaScript that
recomputes it when its dependencies change:

```go
var price = x.NewSignal(10.0)
var qty = x.NewSignal(2.0)
var total = x.Computed(price.Mul(qty))
```

```templ
@x.Number(price)
@x.Number(qty)
@x.Text(total)
```

The combinator set is deliberately small — `Add`, `Sub`, `Mul`, `Div` over
numbers (and `+` over strings), `x.Lit` for constants, `x.Not` for toggling —
because every operator is a step toward a language. When an expression stops
reading like the JavaScript it produces, that is the boundary; see the escape
hatch below.

## Effects

An effect is a `<script>` that runs once on load and again whenever a signal it
reads changes. Dependencies are discovered by execution, so nothing is
declared:

```go
var count = x.Named("count", 0)
```

```templ
@x.Effect() {
	<script>
		document.title = "Count: " + count.get();
	</script>
}
```

`x.Named` is what makes this possible: it fixes the signal's name in the
generated JavaScript, so hand-written scripts and effects can read it as
`count.get()`. Auto-named signals (`s0`, `s1`, …) are fine for bindings, which
never need to know the name.

## Structure

`x.When` mounts and unmounts a subtree as a condition changes — the subtree's
HTML is captured at build time and embedded in the script, and remounting
rebuilds it, so input focus inside it does not survive. Use `x.Show` when only
visibility should change.

```templ
@x.El("button", x.On("click", open.Set(x.Not(open)))) {
	toggle
}
@x.When(open, secretBody())
```

`x.List` renders a signal of items as a keyed list: rows present at build time
are static HTML, and at runtime the list is kept in sync by moving existing
nodes, inserting new ones and removing gone ones — existing rows are never
rebuilt, so input focus inside them survives.

```go
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

var todos = x.Named("todos", []Todo{{ID: "1", Text: "Read the docs"}})
var draft = x.Named("draft", "")
```

```templ
@x.Input(draft)
@x.List(todos, "id", todoRow, "todoRow")
```

`"id"` is the JSON field that keys each item; `todoRow` is the Go renderer for
rows present at build time; `"todoRow"` is the name of a JavaScript function
that renders rows added in the browser — items a build-time renderer cannot
know about:

```templ
templ todoScripts() {
	<script>
		const esc = (s) => String(s).replace(/[&<>"]/g, (c) => "&#" + c.charCodeAt(0) + ";");
		function todoRow(item) {
			return `<li>${esc(item.text)}</li>`;
		}
		document.querySelector("#add").addEventListener("click", () => {
			const text = draft.get();
			if (!text) return;
			todos.set([...todos.get(), { id: String(Date.now()), text }]);
			draft.set("");
		});
	</script>
}
```

Subtrees and rows are static: markers inside them render their static form and
are not re-wired when they remount.

## The escape hatch

Combinators cover increment, toggle, reset and arithmetic. Anything more —
append to a list, validate a field, fetch data — is a `<script>` reading the
named signals, exactly as in the list example above. Scripts have the whole
runtime: `aw.signal`, `aw.computed`, `aw.effect`, and every signal's
`get`/`set`/`sub`. When a handler stops being a binding, the answer is a
script, not another combinator.

`x.Script()` wraps a script in children; `x.JS(comp)` emits a script component
verbatim — plain `<script>` tags in templ work unchanged, and both forms can
read the named signals.

## Static by default

State that is never used in markup generates no JavaScript, a fully static page
emits no script, and the runtime itself is assembled from fragments — a
counter page ships roughly the runtime for a counter, not a framework. Pass
`x.Minified()` to `Document` for compact production output; the development
form is indented and readable.

## x and reactive

`x` and the WASM `reactive` package are two implementations of the same
reactive model — signals with tracked dependencies, fine-grained bindings,
keyed lists. `x` is the lightweight default for the common 90%; `reactive` is
the complete path for anything that needs full Go in the browser — goroutines,
the standard library, the router. Anything past what a script can say is a
normal `CompileFrom` WASM script, exactly as before.
