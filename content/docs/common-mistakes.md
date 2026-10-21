+++
Title = "Common mistakes"
Description = "The things that bite: the templ ctx collision, absolute URLs, map iteration order, Parallel races."
Order = 240
Group = "Reference"
+++

A short list of the mistakes that cost real time, in the order they're likely
to hit you. All of them are "works locally, breaks in a way that takes a while
to name" problems.

## Name a templ parameter `ctx`

templ's generated code declares its own `ctx context.Context`; a parameter or
variable with that name collides and the generated file won't compile. Name
the parameter `c` and pass `ssg.Ctx` as the first argument:

```templ
templ Page(c ssg.Ctx) { … }
```

## Write absolute URLs as literals

Every in-page link and asset must go through `Ctx.Asset` and `Ctx.Link`. A
literal `/style.css` breaks the moment the site is served from a base path —
GitHub Pages, for example — because the whole point of relative URLs is that
one build works everywhere:

```templ
<link rel="stylesheet" href={ c.Asset("style/style.css") }/>
<a href={ c.Link("docs/") }>Docs</a>
```

This site's build fails on an absolute or dead link (`site_test.go` walks
every page and resolves every link), so the mistake surfaces at build time
here — but your site should adopt the same habit: build the output and check
the links.

## Range a `map[string]Page` without `Slugs`

```go
// wrong — the order is random, so the index shuffles on every build:
for slug, page := range pages { … }

// right — sorted, deterministic:
for _, slug := range content.Slugs(pages) { … }
```

The first site built with astwerk produced byte-identical output on every page
except the collection index, which shuffled its tiles on each run. `Slugs`
exists so that failure mode is one explicit call away.

## Share mutable state in `Generate` closures under `Parallel: true`

`Parallel: true` moves your `Generate` and `Page` functions onto other
goroutines. Rendering a templ component is safe; a `Generate` closure appending
to a captured slice is not. Opt into `Parallel` deliberately, and treat shared
state the way you would in any concurrent program.

## Give `Generate` and `Children` the same key

Generated keys win — silently. If a `Generate` returns a key that also exists
in `Children`, the generated node replaces the hand-written one. Name them
deliberately, and test the output.

## Assume a page ships JavaScript

Core astwerk outputs HTML and CSS. A page has JavaScript only when you add a
`<script>` tag or a `CompileFrom` node. If a generated page contains a script
you did not write, look for a template or component that injected it.
