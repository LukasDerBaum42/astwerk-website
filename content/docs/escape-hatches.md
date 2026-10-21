+++
Title = "No abstraction is a prison"
Description = "The design law in practice: every convenience has a path back to plain code."
Order = 20
Group = "Introduction"
+++

The design law is **abstract only the repeating boilerplate**. Its other half
is the principle that gives it teeth: an abstraction should remove boilerplate,
*not capability*. Anything a marker or a helper can't say, you say in plain
code — the framework never becomes a wall.

## The ladder

```go
// framework abstraction
@count

// escape hatch
<script> … </script>

// ultimate escape hatch
syscall/js
```

Each rung is a path down, not a cliff. `x` compiles the common cases; when an
expression stops reading like the JavaScript it produces, you write the
JavaScript. When JavaScript isn't enough, `wasmwrap` wraps `syscall/js` so Go
runs in the browser. Nothing is trapped at the top rung.

## The same shape, everywhere

The ladder isn't unique to interactivity — it's the pattern of the whole
framework:

| Convenience | Escape hatch |
| --- | --- |
| `x.Text(signal)`, `x.When`, `x.List` | a `<script>` reading the named signals |
| `content.LoadCollection[T]` | `LoadDir` + `Decode` + `Slugs` + `Generate` by hand |
| `Node.Generate` | a hand-written `Generate` closure doing anything Go can do |
| `starter/` templates | copy the files and edit them — they're your code |
| `ssg.Templ` adapter | a `Page` closure taking `ssg.Ctx` directly |

## Why this matters

A framework with no escape hatch has to grow a feature for every user need —
which is how frameworks become languages. astwerk's answer is the opposite:
the marker set stays deliberately small, and everything past it is ordinary
Go, ordinary HTML, ordinary JavaScript. The boundary is documented, not
guarded.

> **An abstraction should remove boilerplate, not remove capability.**

If you ever find yourself fighting the framework — a widget that won't
compile, a page the tree can't express — that's not the end of the road, it's
the point where you're supposed to drop a level. That's by design.
