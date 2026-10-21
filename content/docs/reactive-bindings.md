+++
Title = "reactive — DOM bindings"
Description = "Fine-grained bindings: text, attributes, classes, forms, conditional and keyed subtrees."
Order = 340
Group = "Extensions (astwerk-live)"
Extension = true
+++

Bindings attach behaviour to elements. Each one is a one-liner over
[`Effect`](../reactive/) — "re-run this and poke one property" — and updates
exactly the property it owns. The state model behind them is on [reactive —
state](../reactive/).

## Documentation

### BindText, BindHTML, BindAttr, BindClass, BindStyle, BindShow

```go
reactive.BindText(el, count)        // formats numbers and strings itself
reactive.BindText(el, fullName)      // a *Memo binds wherever a signal would
reactive.BindHTML(el, renderedMarkdown) // unescaped — never user input
reactive.BindAttr(el, "disabled", reactive.Computed(func() string {
	if busy.Get() { return "disabled" }
	return "" // empty removes the attribute
}))
reactive.BindClass(el, "active", isActive)
reactive.BindStyle(el, "color", colour)
reactive.BindShow(el, visible)
```

Each binding takes the signal or memo directly — anything that satisfies
`Readable[T]` — rather than a getter function: the element knows how to render
the value, so the conversion dies here. `BindText` accepts a number or a string
value; `BindHTML` is the only unescaped binding.

`BindShow` toggles visibility and keeps the subtree's state; `BindWhen` (below)
builds and tears the subtree down.

### BindValue, BindNumber, BindChecked

```go
reactive.BindValue(input, draft)    // string
reactive.BindNumber(input, amount)  // float64, ignores unparseable input
reactive.BindChecked(box, done)     // bool
```

Two-way: typing sets the signal, and the effect that echoes it back into the
field stops because `Set` ignores an equal write — they don't feed back on
themselves. Use `reactive.On` rather than `Element.On` inside an effect; it
registers the unbind with the enclosing effect, so re-renders don't stack
handlers.

<details>
<summary>Why two-way bindings don't loop</summary>

Typing fires an input event that sets the signal; the echo back writes the same
value into the field. `Set` compares against the current value and skips an
equal write, so the echo produces no change and no further re-run — the loop
dies after one pass.

</details>

### BindWhen

```go
reactive.BindWhen(parent, loggedIn, func() wasmwrap.Element {
	return renderDashboard()
})
```

Builds the subtree when the condition becomes true and removes it when it goes
false, so effects inside it are genuinely fresh each time. `BindShow` only
toggles visibility and keeps state — pick it when the subtree holds nothing
worth discarding.

### BindList

```go
reactive.BindList(list, items,
	func(t Todo) string { return t.ID },
	func(t Todo, row *reactive.Row) wasmwrap.Element {
		el := wasmwrap.Create("li")
		row.Bool(false) // per-row state, disposed with the row
		return el.Append(…)
	})
```

`items` is a `*ListSignal[Todo]` — the keyed-list kind, not a plain slice
signal. An element is rendered once per key and then reused. Adding, removing or
reordering moves existing nodes, so input focus, scroll position and CSS
transitions survive. Two consequences: changing a field of an existing item does
not re-render it — make the parts that change signals — and the parent element
should be dedicated to the list, because ordering is done by index against its
children.

<details>
<summary>Why nodes are moved, not rebuilt</summary>

Rebuilding a row on every change would destroy whatever the row holds — focus
inside an input, scroll position, a transition mid-flight. `BindList`
reconciles by key instead: it figures out which rows to move, insert or detach
and touches only those, so everything else is untouched.

</details>

Testing your bindings — the `jsdom` fake DOM, running under node — is covered on
the [Testing](../testing/) page.

## Example

```go
open := reactive.Bool(false)

reactive.BindClass(nav, "open", open)
reactive.BindShow(panel, open)
reactive.On(toggle, "click", func(wasmwrap.Event) {
	open.Toggle()
})
```

A signal, two one-property bindings, and a handler — the nav-toggle widget.
For a complete widget hydrated onto build-time markup, see the
[Examples](../../examples/) page.
