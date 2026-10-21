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
p, _   := content.Load("content/x.md")
p, _   := content.Parse(src)
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
		fm, _ := content.Decode[FrontMatter](pages[slug])
		body := pages[slug].HTML
		children[slug] = ssg.Node{
			Title: fm.Title,
			Page:  func(c ssg.Ctx) templ.Component { return views.Project(c.Title, body) },
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

import "github.com/LukasDerBaum42/astwerk-live/wasmwrap"

func main() {
	wasmwrap.Query("#toggle").On("click", func(e wasmwrap.Event) {
		e.PreventDefault()
		wasmwrap.Query("#nav").Class().Toggle("open")
	})
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

## x — compiled reactive markup

See [x](../x/) and [x — structure & scripts](../x-structure/).

```go
var count = x.NamedNumSignal("count", 0) // named, so scripts can read it
var open  = x.NewBoolSignal(false)
```

```templ
@x.Document(counter())

templ counter() {
	<div>
		@count
		@x.El("button", x.On("click", count.Inc(1))) {
			+
		}
		@x.El("button", x.On("click", count.SetTo(0))) {
			reset
		}
		@x.El("button", x.On("click", open.Toggle())) {
			toggle
		}
		@x.When(open, secret())
	</div>
}
```

Bindings:

```templ
@x.El("a", x.Attr("href", url))        { link }
@x.El("span", x.Class("on", active))   { toggled class }
@x.El("button", x.Disabled(busy))      { save }
@x.El("p", x.Show(visible))            { shown }
@x.El("p", x.Style("color", colour))   { styled }
```

Forms and computed values:

```go
var price = x.NewNumSignal(10.0)
var qty   = x.NewNumSignal(2.0)
var total = x.Computed(x.Mul(price, qty))
```

```templ
@x.Number(price)
@x.Number(qty)
@x.Text(total)
@x.Input(draft)
@x.Checkbox(subscribed)
```

Structure and effects:

```templ
@x.List(todos, "id", todoRow, "todoRow")
@x.Effect() {
	<script>document.title = "Count: " + count.get();</script>
}
@x.Script() {
	/* arbitrary JavaScript */
}
```

Actions: `Set`, `SetTo`, `Inc`, `Dec`, `Toggle`; expressions: `Add`, `Sub`,
`Mul`, `Div`, `Concat`, `Gt`/`Lt`/`Gte`/`Lte`/`Eq`/`Ne`, `Not`, `Lit`;
composition: `Seq`, `If`. Anything else is a `<script>` reading the named
signals — that is the escape hatch.

## wasmwrap

See [wasmwrap](../wasmwrap/) and [wasmwrap — canvas](../wasmwrap-canvas/).

```go
el := wasmwrap.Query("#box")
el.Style().Color("red").Display("block")
el.SetText("hi").SetAttr("data-id", "7").Class().Add("row")

wasmwrap.Create("li").Append(child)
btn.On("click", handler)        // returns an unsubscribe func
body, err := wasmwrap.Fetch("/api/thing") // blocks; safe outside handlers

canvas := wasmwrap.Query("#c").AsCanvas()
ctx := canvas.Context2D()
w, h := canvas.FitToDisplay(ctx)
ctx.FillStyle("#4f7cff").BeginPath().Arc(w/2, h/2, 40, 0, 2*math.Pi).Fill()

el.Value().Call("scrollIntoView", map[string]any{"behavior": "smooth"})
```

A failed `Query` returns an element where `Exists()` is false; readers return
zero values — missing is inert, not a panic.

## reactive

See [reactive — state](../reactive/), [DOM bindings](../reactive-bindings/),
[routing & async](../reactive-router/) and
[templ in the browser](../reactive-templ/).

```go
count := reactive.Num(0)          // split types: Num, Bool, Str, List
count.Get(); count.Peek(); count.Set(5)
count.Inc(1); count.Dec(1)        // the one-word forms
count.Subscribe(fn)               // manual subscription

full := reactive.Computed(func() string { return first.Get() + " " + last.Get() })
full.Dispose()                    // detach from inputs

reactive.Effect(func() { fmt.Println(count.Get()) }) // tracks deps automatically
reactive.Batch(func() { first.Set("Ada"); last.Set("Lovelace") })
reactive.Untracked(func() string { return count.Get() }) // read, no subscription
dispose := reactive.Scope(func() { /* effects owned here */ })
```

DOM bindings — signals and memos bind directly (they are `Readable`):

```go
reactive.BindText(el, count)                       // formats numbers too
reactive.BindHTML(el, md)                          // unescaped — never user input
reactive.BindAttr(el, "disabled", reactive.Computed(func() string { /* "" removes it */ }))
reactive.BindClass(el, "active", isActive)
reactive.BindStyle(el, "color", colour)
reactive.BindShow(el, visible)
reactive.BindWhen(parent, cond, renderDashboard)
reactive.BindValue(input, draft)                   // *StrSignal
reactive.BindNumber(input, amount)                 // *NumSignal, ignores unparseable input
reactive.BindChecked(box, done)                    // *BoolSignal
reactive.On(el, "click", handler)                  // auto-cleanup
```

Keyed lists:

```go
reactive.BindList(list, items,
	func(t Todo) string { return t.ID },
	func(t Todo, row *reactive.Row) wasmwrap.Element {
		el := wasmwrap.Create("li").SetText(t.Text)
		return el
	})
```

Routing and async:

```go
reactive.Router(root, []reactive.Route{
	{Path: "/", View: home},
	{Path: "/projects/:slug", View: func(p reactive.Params) wasmwrap.Element {
		return project(p.Get("slug"))
	}},
	{Path: "/*", View: notFound},
})
reactive.InterceptLinks(wasmwrap.Body())

posts := reactive.NewResource(func() ([]Post, error) {
	return reactive.FetchJSON[[]Post]("/api/posts")
})
posts.Loading.Get(); posts.Err.Get(); posts.Data.Get(); posts.Reload()
```

Using templ in the browser:

```go
reactive.Hydrate("#counter", func(root wasmwrap.Element) {
	reactive.BindText(root.Find("[data-count]"), count)
})
reactive.HydrateAll(".widget", mount)                    // hydrate every match
reactive.RenderTempl(views.Row(item))                     // component → string
reactive.TemplElement("li", views.Row(item))              // component → element
reactive.BindTempl(preview, func() templ.Component { return views.Markdown(src.Get()) })
```

## Testing

See [Testing](../testing/).

```go
func init() { jsdom.Install() }

func TestCounter(t *testing.T) {
	jsdom.Reset()
	root := wasmwrap.Wrap(jsdom.NewElement("div"))
	// mount, act, assert
}
```

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
