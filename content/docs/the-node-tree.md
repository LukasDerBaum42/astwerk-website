+++
Title = "The Node tree"
Description = "Node, Build, BuildOptions — the shape of the tree and how it's written."
Order = 20
Group = "Start here"
+++

A `Node` describes one directory in the output. `Build` walks the tree from the
root and writes it. That is the whole mechanism — there is no config file, no
convention scanner, no code generation, no second system. What each page
receives when it renders is on [Ctx & URLs](../ctx/).

## Documentation

### Node

```go
type Node struct {
	Title       string
	Page        PageFunc
	Files       map[string]PageFunc
	Generate    func() map[string]Node
	CopyFrom    string
	CompileFrom string
	Children    map[string]Node
}
```

Every field is optional; the zero value only recurses into `Children`, so a node
can exist purely to group.

| Field | Effect at this node's directory |
| --- | --- |
| `Title` | passed to pages as `Ctx.Title` |
| `Page` | renders `index.html` |
| `Files` | renders extra files by name — `404.html`, `rss.xml`, `.nojekyll` |
| `Generate` | children computed at build time, merged into `Children` |
| `CopyFrom` | copies a directory's contents in as-is |
| `CompileFrom` | compiles a scripts directory in — see [Scripts & WebAssembly](../scripts-and-wasm/) |
| `Children` | subdirectories, keyed by name |

### Children keys

Keys may contain slashes to nest (`"assets/style"`), but may not escape the
output directory. Children are visited in sorted key order, so a build is
reproducible. A node with nothing to do creates no directory.

<details>
<summary>Why keys are validated</summary>

A key like `../secrets` or `/etc/passwd` would write outside the output
directory if it were used as a path. `..` and absolute paths are rejected up
front so a tree can't escape its `OutDir`.

</details>

<details>
<summary>Why children are visited in sorted order</summary>

Map iteration order in Go is random, and a build that shuffles its output
layout is a build you can't trust. Sorted keys make two things deterministic:
the output is byte-identical across builds, and when a build fails, the same
node is reported every time instead of whichever the map handed over first.

</details>

### Generate

`Generate` runs at build time and returns children to merge into `Children` —
generated keys win. This is how a directory of markdown becomes a set of pages:

```go
Generate: func() map[string]ssg.Node {
	pages, _ := content.LoadDir("content/projects")
	out := map[string]ssg.Node{}
	for _, slug := range content.Slugs(pages) {
		out[slug] = ssg.Node{Page: post(pages[slug])}
	}
	return out
},
```

The full collection pattern — index page, tiles, pagination — is on
[Content & collections](../content-and-collections/).

### PageFunc, Templ, Static

`Page` and every `Files` entry is a `PageFunc`:

```go
type PageFunc func(Ctx) templ.Component
```

Adapters cover the common shapes:

```go
"about": {Title: "About", Page: ssg.Templ(views.About)}, // (title, path, prefix)
Page: ssg.Static(views.Home),                            // needs nothing
Page: func(c ssg.Ctx) templ.Component {                  // anything else
	return views.Projects(c.Title, c.Path, c.Prefix, tiles)
},
```

`ssg.Templ` adapts the common `(title, path, prefix)` signature — the walker
fills in path and prefix, so only the title is yours to give. `ssg.Static` is
for components that need nothing. Anything else — extra arguments, computed
values — is a closure.

### Files

`Page` always writes `index.html`, because pretty URLs want a directory per
page. Some files can't work that way — hosts look for `404.html` at the root,
feed readers want `rss.xml`, and GitHub Pages needs `.nojekyll` or it runs the
output through Jekyll:

```go
Files: map[string]ssg.PageFunc{
	"404.html":  ssg.Templ(views.NotFound),
	"rss.xml":   feed,
	".nojekyll": empty,
},
```

`Files` entries render after `Page`, so a key of `"index.html"` overwrites it.

### BuildOptions

```go
type BuildOptions struct {
	OutDir       string // default "build"
	Dev          bool   // skip the clean, overwrite in place
	BaseURL      string // "/astwerk" when served from a subdirectory
	RelativeURLs bool   // emit URLs relative to each page
	Parallel     bool   // render sibling nodes concurrently
}
```

`OutDir` defaults to `"build"`. `BaseURL` and `RelativeURLs` control the URLs
pages generate — see [Ctx & URLs](../ctx/). `Dev` exists so a running dev server
doesn't have the ground pulled out from under it: without it, `Build` removes
`OutDir` first, so a renamed page can't leave a stale file behind.

<details>
<summary>Parallel, and the escape hatch</summary>

`Parallel` renders siblings across goroutines, bounded by `GOMAXPROCS`. The
output and error reporting are identical either way; only wall time changes.
It's opt-in because it runs *your* `Page`, `Files` and `Generate` functions
concurrently — rendering templ components is safe, a `Generate` closure
appending to a captured slice is not. `CompileScripts` is always parallel; each
`.wasm` is built by its own `go build` process, so there's nothing of yours to
race.

The tree won't always be enough, and that's fine:

```go
if err := ssg.Build(root, opts); err != nil {
	log.Fatal(err)
}
if err := writeSitemap("build/sitemap.xml", pages); err != nil {
	log.Fatal(err)
}
```

That's the intended escape hatch, not a workaround: the tree covers the
repeating shape, anything genuinely one-off stays plain Go.

</details>

## Example

```go
root := ssg.Node{
	Title: "Site",
	Page:  ssg.Templ(views.Home),
	Children: map[string]ssg.Node{
		"about": {Title: "About", Page: ssg.Templ(views.About)},
		"style": {CopyFrom: "style"},
	},
}
if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
	log.Fatal(err)
}
```

That writes `build/index.html`, `build/about/index.html`, and copies `style/`
in. `Page` renders the index; each `Children` key becomes a directory; `CopyFrom`
is copied verbatim.
