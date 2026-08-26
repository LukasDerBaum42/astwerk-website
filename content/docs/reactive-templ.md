+++
Title = "reactive — templ in the browser"
Description = "Rendering and hydrating templ components client-side."
Order = 63
Group = "Client-side"
+++

templ is pure Go, so it compiles to wasm and runs in the browser unchanged —
the same component can render a page at build time and a fragment at runtime.
These helpers are how a script gets templ output into the DOM.

## Documentation

### Hydrate, HydrateAll

```go
reactive.Hydrate("#counter", func(root wasmwrap.Element) {
	reactive.BindText(root.Find("[data-count]"), func() string {
		return strconv.Itoa(count.Get())
	})
	reactive.On(root.Find("[data-inc]"), "click", func(wasmwrap.Event) {
		count.Update(func(v int) int { return v + 1 })
	})
})
```

Attaches behaviour to markup astwerk already generated — the recommended way to
use templ in the browser. Markup lives in one place, updates stay surgical, and
the page still reads with WebAssembly disabled. `Hydrate` does nothing if the
selector matches nothing, so one script can be included everywhere and act only
where its markup exists.

`HydrateAll` runs the callback once per element that matches, for repeated
widgets on one page.

<details>
<summary>Why hydration is the default over BindTempl</summary>

`ssg.Build` renders your templ at build time; a script attaches behaviour to
that markup. Nothing renders twice, so there is never a flash of unstyled
content, and the markup stays in one place. Re-rendering client-side is the
exception, not the rule.

</details>

### BindTempl

```go
reactive.BindTempl(preview, func() templ.Component {
	return views.Markdown(source.Get())
})
```

Runs a component in the browser and assigns the result as innerHTML. templ
produces a *string*, so there is nothing to patch against: everything inside the
target is destroyed and rebuilt, losing input focus, scroll position, selection,
and any handler bound inside. Good for a display-only panel that changes as a
unit; wrong for anything the user interacts with.

<details>
<summary>Why BindTempl loses state</summary>

`BindTempl` writes the rendered string as `innerHTML`, and innerHTML has no
concept of "what changed". The whole subtree is replaced, so anything the user
was holding — focus, selection, an input's value — is gone with the old nodes.

</details>

### RenderTempl, TemplElement

```go
html, err := reactive.RenderTempl(views.Row(item))    // component → string
el := reactive.TemplElement("li", views.Row(item))    // component → detached element
```

`RenderTempl` is the string form; `TemplElement` renders into a detached element
for handing to `BindList` or `BindWhen`. `BindTempl` is built on top of these.

## Example

```go
count := reactive.NewSignal(0)

reactive.Hydrate("#counter", func(root wasmwrap.Element) {
	reactive.BindText(root.Find("[data-count]"), func() string {
		return strconv.Itoa(count.Get())
	})
	reactive.On(root.Find("[data-inc]"), "click", func(wasmwrap.Event) {
		count.Update(func(v int) int { return v + 1 })
	})
})
```

Markup astwerk generated at build time, hydrated into a live counter.
