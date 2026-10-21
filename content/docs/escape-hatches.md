+++
Title = "No abstraction is a prison"
Description = "The design law in practice: every convenience has a path back to plain code."
Order = 20
Group = "Introduction"
+++

The design law is **abstract only the repeating boilerplate**. Its other half
is the principle that gives it teeth: an abstraction should remove boilerplate,
*not capability*. Anything a helper can't say, you say in plain code — the
framework never becomes a wall.

## The ladder

```go
// convenience: the common case, one line
"about": {Title: "About", Page: ssg.Templ(views.About)},

// escape hatch: the same page, written out
"about": {Title: "About", Page: func(c ssg.Ctx) templ.Component {
	return views.About(c.Title, c.Path, c.Prefix)
}},

// ultimate escape hatch: the page is just Go
"about": {Title: "About", Page: func(c ssg.Ctx) templ.Component {
	return templ.Raw(renderAboutWithAnything(c))
}},
```

Each rung is a path down, not a cliff. `ssg.Templ` fits the common signature
and gets out of the way; a `Page` closure takes exactly the context you need;
and if the page isn't HTML at all, it's a `templ.Component` you built yourself.

## The same shape, everywhere

The ladder isn't unique to pages — it's the pattern of the whole framework:

| Convenience | Escape hatch |
| --- | --- |
| `ssg.Templ(views.Page)` | a `Page` closure taking `ssg.Ctx` directly |
| `content.LoadCollection[T]` | `LoadDir` + `Decode` + `Slugs` + `Generate` by hand |
| `Node.Generate` | a hand-written `Generate` closure doing anything Go can do |
| `CopyFrom` | copy the files in your own Go code, or generate them |
| markdown pages | raw HTML in markdown, or a templ component instead |
| `starter/` templates | copy the files and edit them — they're your code |

## Why this matters

A framework with no escape hatch has to grow a feature for every user need —
which is how frameworks become languages. astwerk's answer is the opposite:
the helper set stays deliberately small, and everything past it is ordinary
Go, ordinary HTML, ordinary templ. The boundary is documented, not guarded.

> **An abstraction should remove boilerplate, not remove capability.**

If you ever find yourself fighting the framework — a page the tree can't
express, a helper that doesn't fit — that's not the end of the road, it's the
point where you're supposed to drop a level. That's by design.
