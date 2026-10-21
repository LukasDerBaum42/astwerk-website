+++
Title = "Client-side tools"
Description = "Optional astwerk-live packages: compiled markup, signals, direct DOM access and jsdom. Still in heavy redesign."
Order = 300
Group = "Extensions (astwerk-live)"
Extension = true
+++

**This page is not core astwerk.** Everything below lives in the separate
[astwerk-live](https://github.com/LukasDerBaum42/astwerk-live) module. It is
optional, still in heavy redesign, and very likely to change or be removed.
Core astwerk — `ssg`, `content`, `assets`, `devserver`, `starter` — has no
client runtime at all.

With that said, there are currently four ways to put behaviour in a page:

## The decision table

| I want to… | Use |
| --- | --- |
| Generate HTML | core [`ssg`](../the-node-tree/) |
| Generate pages from Markdown | core [`content`](../content-and-collections/) |
| Compile a Go or TypeScript file during the build | core [`CompileFrom`](../scripts-and-wasm/) |
| Add a button / toggle / simple form | `x` (compiled markup) |
| Build a widget with local state | `x` |
| Build a client-routed SPA | `reactive` |
| Manipulate the DOM directly | `wasmwrap` |
| Run arbitrary Go in the browser | `wasmwrap` |
| Test browser bindings without a browser | `jsdom` |
| Write it by hand | a plain `<script>` tag |

## Why there are several ways to make a button

The buttons all look the same; what differs is what runs them:

- **`x`** — the handler is a small expression compiled to JavaScript at build
  time. Zero download beyond a few kilobytes of generated script.
- **`reactive`** — the handler is real Go running in the browser, via WASM.
- **`wasmwrap`** — the raw DOM layer, in Go. Right when you want `syscall/js`
  without the ceremony.
- **plain `<script>`** — the escape hatch. Sometimes the answer isn't a
  framework at all.

`x` and `reactive` are two implementations of one reactive model: signals with
tracked dependencies, fine-grained bindings, keyed lists. The concepts you
learn on one transfer to the other — only the runtime changes.

## The x ↔ reactive terminology matrix

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
a tax on every page.

When the extension settles, this page will move back into the main docs. Until
then, treat the API on the following pages as a moving target.
