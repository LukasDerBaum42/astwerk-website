+++
Title = "Quick reference"
Description = "The whole API on one page — signatures and copyable snippets, man-page style."
Order = 200
Group = "Reference"
+++

The entire API, in code. Everything below is copyable; the longer pages exist to
explain *why*, this one exists to be *looked up*. Each section links to its
canonical page.

## Synopsis

```go
package main

import (
	"log"

	"github.com/LukasDerBaum42/astwerk/ssg"
	"example.com/site/pages"
)

func main() {
	root := ssg.Node{
		Title: "Site",
		Page:  ssg.Templ(pages.Home),
		Children: map[string]ssg.Node{
			"about":   {Title: "About", Page: ssg.Templ(pages.About)},
			"style":   {CopyFrom: "style"},
			"scripts": {CompileFrom: "scripts"},
		},
	}
	if err := ssg.Build(root, ssg.BuildOptions{RelativeURLs: true}); err != nil {
		log.Fatal(err)
	}
}
```

```sh
templ generate && go run .
```

## Node

See [The Node tree](../the-node-tree/) and [Ctx & URLs](../ctx/).

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

| Field | Does |
| --- | --- |
| `Title` | becomes `Ctx.Title` |
| `Page` | renders `index.html` |
| `Files` | extra files: `404.html`, `rss.xml`, `.nojekyll` |
| `Generate` | children computed at build time; generated keys win |
| `CopyFrom` | copies a directory in as-is |
| `CompileFrom` | compiles a scripts dir (`.go`→wasm, `.ts`) |
| `Children` | subdirectories, keyed by name |

Keys may contain slashes (`"assets/style"`). `..` and absolute paths are rejected.

## BuildOptions

See [The Node tree](../the-node-tree/).

```go
ssg.Build(root, ssg.BuildOptions{
	OutDir:       "build",
	Dev:          false, // skip the clean, overwrite in place
	BaseURL:      "/site",
	RelativeURLs: true,  // URLs relative to each page — one build works everywhere
	Parallel:     false, // render siblings concurrently
})
```

## Ctx

See [Ctx & URLs](../ctx/).

```go
type Ctx struct {
	Title  string // the node's Title
	Path   string // "about/", "de/projects/thing/"
	Prefix string // "" or "/de"
	Locale string // "" or "de"
	Base   string // "" or "/site"
}

func (c Ctx) URL() string               // absolute, base-prefixed
func (c Ctx) Asset(path string) string  // base-aware asset URL
func (c Ctx) Link(path string) string   // base- and locale-aware
func (c Ctx) Rel() string               // locale-independent path
func (c Ctx) InLocale(prefix string) string
```

Always build URLs through `Asset`/`Link`, never as literals. Name the parameter
`c`, not `ctx` — templ's generated code declares its own `ctx`.

```templ
templ Layout(c ssg.Ctx) {
	<link rel="stylesheet" href={ c.Asset("style/style.css") }/>
	<a href={ c.Link("docs/") }>Docs</a>
}
```

## Adapters

```go
// Common (title, path, prefix) signature:
"about": {Title: "About", Page: ssg.Templ(views.About)},

// Needs nothing:
Page: ssg.Static(views.Home),

// Needs anything else — write the closure:
Page: func(c ssg.Ctx) templ.Component {
	return views.Projects(c.Title, c.Path, c.Prefix, tiles)
},
```

## Content

See [Content & collections](../content-and-collections/).

```go
p, _   := content.Load("content/notes.md")
p, _    = content.Parse(src)
pages, _ := content.LoadDir("content/docs")  // non-recursive, keyed by slug
pages, _ := content.LoadTree("content/blog") // recursive, keys like "2024/hello"

slugs := content.Slugs(pages)                 // sorted — always range this
fm, _ := content.Decode[FrontMatter](page)
```

```go
type FrontMatter struct {
	Title       string
	Description string
	Tags        []string
	Draft       bool
}
```

```markdown
+++
Title = "Dice Dungeon"
Description = "A roguelike"
Tags = ["game", "go"]
+++

Body here.
```

Collection (index + one page per file), the shortcut — load, chain, generate:

```go
posts, _ := content.LoadCollection[Post]("content/posts") // sorted by slug

latest := posts.Filter(func(it content.Item[Post]) bool {
	return !it.Meta.Draft
}).Sort(func(a, b content.Item[Post]) bool {
	return a.Meta.Date.After(b.Meta.Date)
})

"blog": {Generate: func() map[string]ssg.Node {
	return latest.Nodes(func(c ssg.Ctx, it content.Item[Post]) templ.Component {
		return views.Post(c, it)
	})
}},
```

…and what it does under the hood, if you need to deviate:

```go
func projects() ssg.Node {
	pages, _ := content.LoadDir("content/projects")

	var tiles []templ.Component
	children := map[string]ssg.Node{}
	for _, slug := range content.Slugs(pages) {
		page := pages[slug]
		fm, _ := content.Decode[FrontMatter](page)

		children[slug] = ssg.Node{
			Title: fm.Title,
			Page:  func(c ssg.Ctx) templ.Component { return views.Project(c.Title, page.Render(c)) },
		}
		tiles = append(tiles, views.Tile(fm.Title, fm.Description, slug))
	}
	return ssg.Node{
		Title:    "Projects",
		Page:     func(c ssg.Ctx) templ.Component { return views.Index(c.Title, tiles) },
		Children: children,
	}
}
// "projects": projects(),
```

Pagination: `content.Chunk(items, size)` splits a slice; build the pages yourself.

## Assets

Sidecar metadata pairs a directory of files with a `meta.json` or `meta.toml`
in the same directory:

```go
type Media struct {
	Credit  string `json:"credit" toml:"credit"`
	License string `json:"license" toml:"license"`
}

entries, err := assets.LoadTree[Media]("public/img", assets.Options[Media]{
	Include: []string{"*.png", "*.jpg", "*.webp"},
})

for _, e := range entries {
	// e.File.Path, e.Meta.Credit, e.Meta.License
}
```

`LoadTree` is recursive; `LoadDir` reads one directory. Files without a
sidecar entry are included with the zero `T`; `RequireMeta` drops them.

## Locales

See [Locales & i18n](../i18n/).

```go
root = ssg.BuildLocales(root, []ssg.Locale{{
	Code:   "de",
	Prefix: "/de", // defaults to "/" + Code
	Override: map[string]func(ssg.Node) ssg.Node{
		"": func(n ssg.Node) ssg.Node {
			n.Title, n.Page = "Meine Seite", ssg.Templ(views_de.Home)
			return n
		},
		"about": func(n ssg.Node) ssg.Node {
			n.Title = "Über mich"
			return n
		},
	},
}})
```

Everything not named in `Override` is mirrored automatically with the prefix
set. `CopyFrom`/`CompileFrom` are not inherited — assets are shared.

```templ
templ Nav(c ssg.Ctx) {
	<a href={ c.Link("projects/") }>Projects</a>
	<a href={ c.InLocale("") }>English</a>
	<a href={ c.InLocale("/de") }>Deutsch</a>
}
```

## Scripts & WebAssembly

See [Scripts & WebAssembly](../scripts-and-wasm/).

```go
"scripts": {CompileFrom: "scripts"},
```

```go
//go:build js && wasm

package main

import "syscall/js"

func main() {
	doc := js.Global().Get("document")
	doc.Call("querySelector", "#toggle").Call("addEventListener", "click",
		js.FuncOf(func(js.Value, []js.Value) any {
			doc.Call("querySelector", "#nav").Get("classList").Call("toggle", "open")
			return nil
		}))
	select {} // keep the module alive
}
```

Load it:

```templ
templ WasmScript(c ssg.Ctx, name string) {
	<script src={ c.Asset("scripts/wasm_exec.js") }></script>
	<script data-wasm={ c.Asset("scripts/" + name + ".wasm") }>
		(() => {
			const src = document.currentScript.dataset.wasm;
			const go = new Go();
			WebAssembly.instantiateStreaming(fetch(src), go.importObject)
				.then(res => go.run(res.instance));
		})();
	</script>
}
```

## Extensions

The client-side packages are not core astwerk. They live in the separate
[astwerk-live](https://github.com/LukasDerBaum42/astwerk-live) module, are
optional, and are still in heavy redesign. See
[client-side tools](../which-tool/) for the decision table and the extension
pages that follow it.

## Dev server

See [Dev server](../dev-server/).

```go
func main() {
	if *serve {
		devserver.Run(context.Background(), devserver.Config{
			Build: devserver.Steps(
				devserver.Command("templ", "generate"),
				devserver.Command("go", "run", ".", "--dev"),
			),
		})
		return
	}
	if err := ssg.Build(tree(), ssg.BuildOptions{Dev: *dev}); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go run . --serve
```

`Build` must be a subprocess — a running Go program can't load newly compiled
`.templ` code. Pass `--dev` to the inner build so it overwrites in place.
