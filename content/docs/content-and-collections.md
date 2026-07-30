+++
Title = "Content & collections"
Description = "Markdown with TOML front matter, and the list-plus-detail pattern."
Order = 30
Group = "Content & i18n"
+++

The `content` package reads markdown files with TOML front matter. Combined with
`Node.Generate`, that gives you a collection: one index page plus one page per
file.

## The API

```go
type Page struct {
	Meta map[string]any // front matter between +++ markers
	HTML string         // rendered markdown body
}

func Load(path string) (Page, error)
func Parse(src string) (Page, error)
func LoadDir(dir string) (map[string]Page, error)
func LoadTree(dir string) (map[string]Page, error)
func Slugs(pages map[string]Page) []string
func Decode[T any](p Page) (T, error)
func Chunk[T any](items []T, size int) [][]T
```

`LoadDir` is non-recursive and reads `*.md` only, keyed by filename without the
extension — the slug.

`LoadTree` is the recursive one, keyed by path relative to `dir` without the
extension: `content/blog/2024/hello.md` becomes `2024/hello`. Keys are
slash-separated on every platform, and `Node.Children` keys may contain
slashes — so a `LoadTree` result drops into a tree and nests itself:

```go
Generate: func() map[string]ssg.Node {
	pages, _ := content.LoadTree("content/blog")
	out := map[string]ssg.Node{}
	for _, slug := range content.Slugs(pages) { // "2024/hello"
		out[slug] = ssg.Node{Page: post(pages[slug])} // /blog/2024/hello/
	}
	return out
},
```

Nothing is special-cased: `blog/index.md` is the key `blog/index`, like any
other file.

## Front matter is yours

`Meta` is deliberately a generic map, not a fixed struct. Every site wants
different fields, so `Decode` unmarshals into whatever struct you define:

```go
type FrontMatter struct {
	Title       string
	Description string
	Tags        []string
	Draft       bool
}

fm, err := content.Decode[FrontMatter](page)
```

A file looks like this:

```markdown
+++
Title = "Dice Dungeon"
Description = "A roguelike built in a week"
Tags = ["game", "go"]
+++

# Dice Dungeon

Real markdown from here on, and <span>inline HTML</span> passes through.
```

A page with no front matter decodes to the zero value without error, so an
undecorated markdown file is still valid.

## Always range over Slugs

`LoadDir` returns a map, and Go randomises map iteration order. Range it
directly and your index page reshuffles on every build, producing a diff for no
reason:

```go
// Wrong — order changes every build:
for slug, p := range pages { … }

// Right:
for _, slug := range content.Slugs(pages) {
	p := pages[slug]
	…
}
```

This is worth the explicit call. It was found the hard way: the first site
migrated to astwerk produced byte-identical output on every page *except* the
collection index, which shuffled its tiles each run.

## A collection

`Generate` runs at build time and returns children to merge in. Point it at a
content directory:

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
			Page: func(c ssg.Ctx) templ.Component {
				return views.ProjectPage(c.Title, c.Path, c.Prefix, body)
			},
		}
		tiles = append(tiles, views.Tile(fm.Title, fm.Description, slug))
	}

	return ssg.Node{
		Title: "Projects",
		Page: func(c ssg.Ctx) templ.Component {
			return views.ProjectIndex(c.Title, c.Path, c.Prefix, tiles)
		},
		Children: children,
	}
}
```

Use it as one line in the tree:

```go
"projects": projects("content/projects"),
```

Notice there is no path arithmetic. Each child gets its path from the walker, so
adding a markdown file adds a page and a tile with nothing else to update.

Tile links should be bare slugs — `href="thing/"` — so they resolve relative to
whichever index page they're on. The same tile then works at `/projects/` and
`/de/projects/` with no prefix threading.

## Paging a long collection

`Chunk` splits a slice into groups of at most `size`, in order. That's all it
does, and that's on purpose:

```go
func Chunk[T any](items []T, size int) [][]T
```

Turning those groups into pages is three lines, and every site wants them
slightly differently — whether page one is the index or `/page/1/`, what the
pager looks like, whether the URL says "page" at all. There's no repetition left
to abstract there, so `Chunk` stops at the arithmetic:

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

## Markdown rendering

By default, goldmark with raw HTML enabled, so a page can drop into hand-written
markup where markdown runs out.

To add extensions, build your own goldmark and pass it to a `Loader`:

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

The docs you're reading are rendered exactly that way. Highlighting happens at
build time via chroma, so no JavaScript highlighter ships to your browser — the
`<span>` classes in the code blocks above are in the HTML.
