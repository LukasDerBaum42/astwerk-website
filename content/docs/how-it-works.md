+++
Title = "How it works"
Description = "The walker, locale derivation and content loading — the three mechanisms, in full."
Order = 230
Group = "Reference"
+++

astwerk has three moving parts worth understanding. None of them is large; all of
them are readable in one sitting.

## The three mechanisms

| Mechanism | Where it lives | What it does |
| --- | --- | --- |
| The walker | `ssg/build.go` | depth-first recursion that writes the output |
| Locale derivation | `ssg/locale.go` | mirrors the base tree per locale, applies overrides |
| Content loading | `content/collection.go` | markdown and front matter become typed items |

## The walker

`Build` is a depth-first recursion over `Node`. There is no plugin pipeline, no
lifecycle, no pass ordering to reason about — one function, roughly a hundred
lines, and this is all it does per node:

```go
func buildNode(n Node, dir, sitePath string, loc *localeCtx, opts BuildOptions) error {
	// 1. make the directory, if this node has any output
	// 2. render Page to index.html, then each Files entry
	// 3. copy CopyFrom
	// 4. compile CompileFrom
	// 5. merge Generate() into Children — generated keys win
	// 6. recurse, appending each key to dir and sitePath
}
```

Two paths are threaded down in parallel: `dir` is where files land on disk, and
`sitePath` is what the page believes its URL to be. Keeping them separate is what
lets `BaseURL` change every generated link without moving a single output file.

Children are visited in sorted key order. That's deliberate: a build is
reproducible, and when something fails you get the same node reported every time
rather than whichever the map handed over first.

### Parallelism

`BuildOptions.Parallel` fans siblings out across goroutines instead. Nodes write
to disjoint paths and share nothing, so the output is identical either way — and
errors are still collected by index and returned in sorted order, so a failure
names the same node whether you built in parallel or not.

It's opt-in for one reason: it moves *your* code onto other goroutines. Rendering
a templ component is safe; a `Generate` closure appending to a captured slice is
not. Making that a silent default would turn a routine upgrade into a data race,
so you say when you want it.

`CompileScripts` doesn't ask. Each `.wasm` is built by its own `go build`
process writing its own file — nothing of yours runs in astwerk's address
space — so there's no race to opt out of, and it's always parallel.

### Failure

Errors propagate immediately, wrapped with the path being written:

```
ssg: render build/projects/thing/index.html: …
```

No partial-success reporting, no continuing past a broken page. A site that
half-built is worse than one that didn't.

## Content loading

`content` is deliberately a separate package from `ssg`: the walker does not
know what markdown is. `LoadCollection[T]` reads a directory, decodes each
front matter block into `T`, and keeps the page body for `Item.Content(ctx)`.
The site decides what to do with the items — usually one index page plus one
child per item through `Node.Generate`.

That split is why markdown is optional. A site made only of templ pages never
imports `content` at all.

## Locale derivation

`BuildLocales` copies the base tree and mounts the copy under the locale's code.
The copy carries `Page`, `Files`, `Generate` and `Children` — and deliberately
drops `CopyFrom` and `CompileFrom`.

That omission is the interesting part. Stylesheets and scripts are shared across
languages and served from the site root, so inheriting them would produce
`build/de/style/` duplicating every asset per language. The rule is: a locale
subtree mirrors *pages*, not assets. Branches that end up with nothing to render
are pruned.

Overrides then patch nodes at a path, applied shallowest-first so an override of
`projects` lands before one of `projects/foo`. Each override is a
`func(ssg.Node) ssg.Node` handed the derived node, which is why patching one
field doesn't drop the rest and why the locale root needs no special case —
`n` already holds the inherited children. Locale identity is carried on an
unexported field of the locale's root node and inherited downward by the walker,
which is how a page five levels deep still knows its prefix.

## The build step

`CopyFrom` and `CompileFrom` are the two fields that write files without a
page. `CopyFrom` copies a directory verbatim, which is how `style/` and
`public/` reach the output. `CompileFrom` hands a scripts directory to
`ssg.CompileScripts`: wasm-tagged Go files become `.wasm` modules, `.ts` files
go to `tsc`, and anything else is ignored so helper packages can live
alongside.

Both are build-time concerns, not runtime ones. A site with neither ships
exactly the HTML its templates produce.

<details>
<summary>Reading the source</summary>

Roughly in order of how much they'll teach you:

| File | What's in it |
| --- | --- |
| `ssg/build.go` | the walker |
| `ssg/node.go` | `Node`, `Ctx`, `BuildOptions` |
| `ssg/locale.go` | tree derivation and override application |
| `content/collection.go` | `Collection[T]`, `Item[T]`, sorting and filtering |
| `content/markdown.go` | front matter, markdown rendering, components |
| `devserver/devserver.go` | watch, rebuild, reload |
| `assets/assets.go` | sidecar metadata |

This site is built with astwerk, source at
[astwerk-website](https://github.com/LukasDerBaum42/astwerk-website). Writing it
found three gaps that nothing else had: no base-path support (every link broke
under `/astwerk/`), no way to configure the markdown renderer (so no syntax
highlighting), and no locale-independent path for a language switcher. All
three are in the library now. That is what a real test is for.

</details>
