+++
Title = "x — structure & scripts"
Description = "Conditional and keyed subtrees, effects, and the JavaScript escape hatch."
Order = 320
Group = "Extensions (astwerk-live)"
Extension = true
+++

The markers that change the document's structure or run code of their own:
conditional and keyed subtrees, effects, and raw scripts. Everything here
builds on the state and binding markers from [x](../x/).

## Documentation

### Effect

```templ
@x.Effect() {
	<script>
		document.title = "Count: " + count.get();
	</script>
}
```

Runs once on load and again whenever a signal it reads changes. Dependencies
are discovered by execution, so nothing is declared. The script is written in
templ, so its content is raw — no escaping, no quoting.

### When

```templ
@x.El("button", x.On("click", open.Toggle())) {
	toggle
}
@x.When(open, secretBody())
```

Mounts and unmounts a subtree as the condition changes. The subtree's HTML is
captured at build time, and remounting rebuilds it.

<details>
<summary>Why remounting loses focus</summary>

`When` rebuilds the subtree from the captured HTML each time it mounts, so any
input focus, scroll position or transient state inside is discarded. If the
subtree only needs to appear and disappear — a panel, a dropdown — `Show` keeps
the same nodes and only toggles visibility, so state survives.

That cuts both ways: an entrance animation on every mount is fine, because the
subtree is fresh anyway — but the animation must respect
`prefers-reduced-motion`. The site-wide rule is that reduced motion collapses
animations to instant state changes: drive the transition with a class toggle
and let the stylesheet's reduced-motion block handle the rest, rather than
animating in JavaScript.

</details>

### List

```go
type Todo struct {
	ID   string `json:"id"` // the JSON field that keys each item
	Text string `json:"text"`
}

var todos = x.NamedListSignal("todos", []Todo{{ID: "1", Text: "Read the docs"}})
var draft = x.NamedStrSignal("draft", "")
```

```templ
@x.Input(draft)
@x.List(todos, "id", todoRow, "todoRow")
```

Renders a signal of items as a keyed list: the JSON key identifies each item,
`todoRow` is the Go renderer for rows present at build time, and `"todoRow"` is
the name of a JavaScript function that renders rows added in the browser.
Existing rows are moved, inserted or detached — never rebuilt.

<details>
<summary>Why the key is a JSON field</summary>

`List` works from the JSON form of the signal: rows are matched by the named
field, and rows created in the browser are plain objects carrying that field.
The field has to be stable across an item's lifetime — an item whose key
changes is treated as a new item.

</details>

### Script

```templ
templ todoScripts() {
	<script>
		document.querySelector("#add").addEventListener("click", () => {
			const text = draft.get();
			if (!text) return;
			todos.set([...todos.get(), { id: String(Date.now()), text }]);
			draft.set("");
		});
	</script>
}
```

Combinators cover increment, toggle, reset and arithmetic. Anything more —
append to a list, validate a field, fetch data — is a `<script>` reading the
named signals. Scripts have the whole runtime: `aw.signal`, `aw.computed`,
`aw.effect`, and every signal's `get`/`set`/`sub`.

```templ
@x.Script() {
	/* raw JavaScript, wrapped in <script> tags */
}
```

### JS

```templ
@x.JS(handlers()) // a script authored as a templ component, emitted verbatim
```

`Script` is for raw markup; `JS` takes a templ component and emits its output
verbatim.

## Example

```go
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

var todos = x.NamedListSignal("todos", []Todo{{ID: "1", Text: "Read the docs"}})
var draft = x.NamedStrSignal("draft", "")
```

```templ
templ todoRow(t Todo) {
	<li>{ t.Text }</li>
}

templ todo() {
	<div>
		@x.Input(draft)
		<button id="add">Add</button>
		@x.List(todos, "id", todoRow, "todoRow")
		@x.Script() {
			<script>
				document.querySelector("#add").addEventListener("click", () => {
					const text = draft.get();
					if (!text) return;
					todos.set([...todos.get(), { id: String(Date.now()), text }]);
					draft.set("");
				});
			</script>
		}
	</div>
}
```

A keyed list fed by a two-way input, with the script escape hatch doing the one
thing combinators can't: appending to the array. The same widget ships in
`astwerk-live/starter/widgets/todo.templ`.
