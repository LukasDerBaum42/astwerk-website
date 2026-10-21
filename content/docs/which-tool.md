+++
Title = "Which tool do I use?"
Description = "A decision table for the whole framework — static HTML, content, x, reactive, wasmwrap, jsdom."
Order = 100
Group = "Add interactivity"
+++

There are four ways to make a button in astwerk. That's a fair thing to ask
about, and the answer is one axis: *how much of the interactive half of the
page is a program?*

## The decision table

| I want to… | Use |
| --- | --- |
| Generate HTML | [`ssg`](../the-node-tree/) |
| Generate pages from Markdown | [`content`](../content-and-collections/) |
| Add a button / toggle / simple form | [`x`](../x/) |
| Build a widget with local state | [`x`](../x/) |
| Build a client-routed SPA | [`reactive`](../reactive/) |
| Manipulate the DOM directly | [`wasmwrap`](../wasmwrap/) |
| Run arbitrary Go in the browser | [`wasmwrap`](../wasmwrap/) |
| Test browser bindings without a browser | [`jsdom`](../testing/) |
| Compile TypeScript | [`CompileScripts`](../scripts-and-wasm/) |

## Why there are four ways to make a button

The buttons all look the same; what differs is what runs them:

- **`x`** — the handler is a small expression compiled to JavaScript at build
  time. Zero download beyond a few kilobytes of generated script. Right for
  anything that is state, bindings and structure.
- **`reactive`** — the handler is real Go running in the browser. Right when
  the widget *is* a program: a router, a canvas, async work, goroutines.
- **`wasmwrap`** — the raw DOM layer, in Go. Right when you want `syscall/js`
  without the ceremony, and the base `reactive` builds on.
- **plain `<script>`** — the escape hatch. Sometimes the answer isn't a
  framework at all.

`x` and `reactive` are [two implementations of one reactive model](../x/):
signals with tracked dependencies, fine-grained bindings, keyed lists. The
concepts you learn on one transfer to the other — only the runtime changes.

## The x &harr; reactive terminology matrix

The two implementations share one model, so every concept has a name on each
side. The bridge is the proof: the same named cell, readable and writable from
both runtimes — [see it live](../../examples/bridge/).

| Concept | `x` (compiled JS) | `reactive` (WASM) |
| --- | --- | --- |
| Mutable state | `x.NewNumSignal(0)` / `Named*` | `reactive.Num(0)` |
| Derived value | `x.Computed(x.Add(a, b))` | `reactive.Computed(func() T { … })` |
| Update | `count.Inc(1)`, `count.SetTo(0)`, `open.Toggle()` | `count.Inc(1)`, `s.Set(v)`, `open.Toggle()` |
| Text binding | `@count` | `reactive.BindText(el, count)` |
| Conditional | `x.When(cond, …)` | `reactive.BindWhen(parent, cond, …)` |
| Keyed list | `@items.List("id", row, …)` | `reactive.BindList(parent, items, key, row)` |
| Effects / cleanup | `x.Effect` | `reactive.Effect` + `OnCleanup` |
| Bridge | the named cell | `reactive.XNum("count")` |

## The rule of thumb

> Use `x` when you need a *little* interactivity. Use WASM when you actually
> need Go.

A page with no markers ships no script at all. A counter built with `x` ships
its generated script; the same counter as WASM ships the ~2 MB Go runtime. The
runtime is the price of one specific choice — running Go in the browser — not
a tax on every page. Most sites stay on the static + `x` side of that line.
