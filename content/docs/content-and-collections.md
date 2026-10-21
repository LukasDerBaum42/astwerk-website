+++
Title = "Content & collections"
Description = "Markdown with TOML front matter, and the list-plus-detail pattern."
Order = 80
Group = "Add content"
+++

The `content` package reads markdown files with TOML front matter. Combined with
`Node.Generate`, that gives you a collection: one index page plus one page per
file.

## Documentation

### Page

```go
type Page struct {
	Meta map[string]any // front matter between +++ markers, or nil
	HTML string         // the rendered markdown body
}
```

`Page` is what every loader returns. The front matter fence is
`content.Delimiter` (`"+++"`); a document without a leading fence is all body.
`Meta` is the raw front matter map — `Decode` turns it into your own struct.

### Load and Parse

```go
func Load(path string) (Page, error)
func Parse(src string) (Page, error)
```

`Load` reads a file, `Parse` takes a string. Both render through goldmark with
raw HTML passed through, so a page can drop into hand-written markup.

### LoadDir and LoadTree

```go
func LoadDir(dir string) (map[string]Page, error)  // non-recursive, keyed by slug
func LoadTree(dir string) (map[string]Page, error) // recursive, keys like "2024/hello"
```

`LoadDir` reads the `*.md` files in one directory, keyed by filename without
the extension — the slug. `LoadTree` is the recursive variant, keyed by path
relative to `dir`: `content/blog/2024/hello.md` becomes `2024/hello`. Keys are
slash-separated on every platform and may nest, so they drop straight into
`Node.Children`:

```go
Generate: func() map[string]ssg.Node {
	pages, _ := content.LoadTree("content/blog")
	out := map[string]ssg.Node{}
	for _, slug := range content.Slugs(pages) { // "2024/hello"
		out[slug] = ssg.Node{Page: post(pages[slug])}
	}
	return out
},
```

### Slugs

```go
func Slugs(pages map[string]Page) []string // sorted — always range this
```

Ranging the map directly gives a different order on every build, which shuffles
any index page built from it.

<details>
<summary>How this bit the first site</summary>

The first site built with astwerk produced byte-identical output on every page
*except* the collection index, which shuffled its tiles on each run. `Slugs`
exists so that failure mode is one explicit call away from impossible.

</details>

### Decode

```markdown
+++
Title = "Dice Dungeon"
Description = "A roguelike"
Tags = ["game", "go"]
+++

Body here.
```

```go
type FrontMatter struct {
	Title       string
	Description string
	Tags        []string
}

fm, err := content.Decode[FrontMatter](page)
```

Front matter is deliberately a generic map, not a fixed struct — every site
wants different fields, so `Decode` unmarshals into whatever you define. A page
with no front matter decodes to the zero value without error.

<details>
<summary>Why Meta is a map, not a struct</summary>

There is no front matter schema the library could know — every site wants
different fields. So `Page.Meta` stays a generic map and `Decode` is the
bridge to your own struct: the map is the storage, your struct is the
interface.

</details>

### Collection

`LoadDir + Decode + Slugs + Generate` is the repeating pattern, so it's
abstracted once: a `Collection[T]` is a loaded set of pages whose front matter
is decoded into your type, sorted by slug by default.

```go
posts, err := content.LoadCollection[Post]("content/posts")
```

The methods return new collections rather than mutating, so they chain —
filter drafts, sort by date, take a page at a time:

```go
latest := posts.Filter(func(it content.Item[Post]) bool {
	return !it.Meta.Draft
}).Sort(func(a, b content.Item[Post]) bool {
	return a.Meta.Date.After(b.Meta.Date)
})
```

And `Nodes` turns the collection into one `ssg.Node` per item, keyed by slug,
ready for `Generate`:

```go
"blog": {Generate: func() map[string]ssg.Node {
	return latest.Nodes(func(c ssg.Ctx, it content.Item[Post]) templ.Component {
		return views.Post(c, it)
	})
}},
```

`Nodes` returns the same `map[string]ssg.Node` `Generate` uses, so it composes
with hand-written nodes — one odd item can still be built by hand and merged
into the result.

<details>
<summary>What Collection does under the hood</summary>

A collection is a thin layer over the primitives on this page: `LoadDir`
loads, `Decode` reads each page's front matter into `T`, `Slugs` keeps the
order deterministic, and `Nodes` is the `Generate` loop at the bottom of this
page. The primitives stay public — reach for them when a site needs to deviate.

</details>

### Chunk

```go
func Chunk[T any](items []T, size int) [][]T
```

Splits a slice into groups of at most `size`, in order, for paginating a
collection. It's deliberately just the arithmetic — turning the groups into
pages is three lines, and every site wants them slightly differently.

<details>
<summary>Paging a collection</summary>

```go
"blog": {
	Page: index(pages[0]), // page one is the collection index
	Generate: func() map[string]ssg.Node {
		out := map[string]ssg.Node{}
		for i, group := range content.Chunk(posts, 10) {
			if i > 0 {
				out["page/"+strconv.Itoa(i+1)] = ssg.Node{Page: index(group)}
			}
			for _, p := range group {
				out[p.Slug] = ssg.Node{Title: p.Title, Page: post(p)}
			}
		}
		return out
	},
},
```

A `size` of zero or less returns one group holding everything, so a site that
hasn't decided on a page size yet doesn't have to special-case it.

</details>

### Custom markdown

The default renderer is goldmark with raw HTML enabled. To add extensions,
build your own goldmark and pass it to a `Loader`:

```go
md := goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

pages, err := content.NewLoader(md).LoadDir("content/docs")
```

The docs you're reading are rendered exactly that way — highlighting happens at
build time via chroma, so no JavaScript highlighter ships to the browser.

**Raw HTML passes through unsanitized.** goldmark doesn't filter it, so a
markdown file can drop into markup — and untrusted content can inject it. Only
enable raw HTML for content you control, and treat `SetHTML`-style output the
same way: never feed user input through it unescaped.

## Example

```go
func projects(dir string) ssg.Node {
	pages, err := content.LoadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var tiles []templ.Component
	children := map[string]ssg.Node{}
	for _, slug := range content.Slugs(pages) {
		fm, _ := content.Decode[FrontMatter](pages[slug])
		body := pages[slug].HTML

		children[slug] = ssg.Node{
			Title: fm.Title,
			Page:  func(c ssg.Ctx) templ.Component {
				return views.ProjectPage(c.Title, c.Path, c.Prefix, body)
			},
		}
		tiles = append(tiles, views.Tile(fm.Title, fm.Description, slug))
	}

	return ssg.Node{
		Title:    "Projects",
		Page:     func(c ssg.Ctx) templ.Component { return views.ProjectIndex(c.Title, c.Path, c.Prefix, tiles) },
		Children: children,
	}
}

// "projects": projects("content/projects"),
```

Each child gets its path from the walker, so adding a markdown file adds a page
and a tile with nothing else to update. Tile links should be bare slugs
(`href="thing/"`) so they resolve relative to whichever index page they're on.
