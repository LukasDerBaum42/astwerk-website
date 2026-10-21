package tutorial

// Source snippets for the tutorial steps, and the project-explorer pictures.
// The snippets live in a plain .go file: templ rewrites raw strings that
// contain templ syntax into their generated Go, and a snippet is meant to show
// the source.

// The project-explorer pictures, one per step that shows them: the source tree
// and the build-output tree after that step, with the step's additions flagged.
var (
	expSetup      = pair("myblog", []string{"go.mod", "go.sum"}, []string{"go.mod", "go.sum"}, "build", nil, nil)
	expTree       = pair("myblog", []string{"go.mod", "main.go", "pages/home.templ"}, []string{"main.go", "pages/home.templ"}, "build", []string{"index.html"}, []string{"index.html"})
	expLayout     = pair("myblog", []string{"go.mod", "main.go", "pages/home.templ", "pages/layout.templ", "style/style.css"}, []string{"pages/layout.templ", "style/style.css"}, "build", []string{"index.html", "style/style.css"}, []string{"style/style.css"})
	expPosts      = pair("myblog", []string{"go.mod", "main.go", "pages/home.templ", "pages/layout.templ", "pages/post.templ", "content/posts/hello.md", "style/style.css"}, []string{"pages/post.templ", "content/posts/hello.md"}, "build", []string{"index.html", "style/style.css", "posts/hello/index.html"}, []string{"posts/hello/index.html"})
	expCollection = pair("myblog", []string{"go.mod", "main.go", "pages/home.templ", "pages/layout.templ", "pages/post.templ", "content/posts/hello.md", "style/style.css"}, nil, "build", []string{"index.html", "style/style.css", "posts/index.html", "posts/hello/index.html"}, []string{"posts/index.html"})
	expRss        = pair("myblog", []string{"go.mod", "main.go", "pages/home.templ", "pages/layout.templ", "pages/post.templ", "content/posts/hello.md", "style/style.css"}, nil, "build", []string{"index.html", "rss.xml", "sitemap.xml", "style/style.css", "posts/index.html", "posts/hello/index.html"}, []string{"rss.xml", "sitemap.xml"})
)

const tutStep2Main = `package main

import (
	"log"

	"github.com/LukasDerBaum42/astwerk/ssg"
	"example.com/myblog/pages"
)

func main() {
	root := ssg.Node{
		Title: "My Blog",
		Page:  ssg.Templ(pages.Home),
	}
	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
		log.Fatal(err)
	}
}`

const tutStep2Home = `package pages

templ Home(title string, path string, prefix string) {
	<h1>My Blog</h1>
	<p>Posts about Go, templ and static sites.</p>
}`

const tutLayout = `package pages

// Layout is the shell every page wraps: html/head/body, the nav, the
// stylesheet, and { children... } where the page's own content goes.
templ Layout(title string, path string, prefix string) {
	<!DOCTYPE html>
	<html lang={ lang(prefix) }>
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>{ title }</title>
			<link rel="stylesheet" href={ prefix + "/style/style.css" }/>
		</head>
		<body>
			<nav>
				<a href={ prefix + "/" }>My Blog</a>
				<a href={ prefix + "/posts/" }>Posts</a>
				<a href={ prefix + "/about/" }>About</a>
			</nav>
			<main>
				{ children... }
			</main>
		</body>
	</html>
}

// lang maps the locale prefix to an html lang attribute.
func lang(prefix string) string {
	if prefix == "" {
		return "en"
	}
	return prefix[1:] // "/de" -> "de"
}`

const tutHome = `package pages

templ Home(title string, path string, prefix string) {
	@Layout(title, path, prefix) {
		<h1>My Blog</h1>
		<p>Posts about Go, templ and static sites.</p>
	}
}`

const tutStyleTree = `		Children: map[string]ssg.Node{
			"about": {Title: "About", Page: ssg.Templ(pages.About)},
			"style": {CopyFrom: "style"},
		},`

const tutPostMD = `+++
Title = "Hello, blog"
Description = "The first post."
Date = "2026-08-28"
Tags = ["meta", "astwerk"]
+++

This is the first post. The front matter between the +++ markers is TOML;
the rest is markdown, rendered to HTML by the build.

## Why static

No database, no server, no runtime to patch. A static site is a folder.`

const tutPostPage = `package pages

// Post renders one markdown file. The body is markdown the build already
// rendered to HTML, so it is trusted markup.
templ Post(title string, path string, prefix string, body string) {
	@Layout(title, path, prefix) {
		<h1>{ title }</h1>
		@templ.Raw(body)
	}
}`

const tutPostsManual = `	"posts": {
		Generate: func() map[string]ssg.Node {
			posts, _ := content.LoadDir("content/posts")

			out := map[string]ssg.Node{}
			for _, slug := range content.Slugs(posts) {
				fm, _ := content.Decode[FrontMatter](posts[slug])
				out[slug] = ssg.Node{
					Title: fm.Title,
					Page: func(c ssg.Ctx) templ.Component {
						return pages.Post(c.Title, c.Path, c.Prefix, posts[slug].HTML)
					},
				}
			}
			return out
		},
	},`

const tutPostsCollection = `// FrontMatter is the shape of the TOML block in content/posts/*.md.
type FrontMatter struct {
	Title       string
	Description string
	Date        string
	Tags        []string
}

// posts returns the blog's posts subtree: an index and one page per file.
func posts() ssg.Node {
	collection, err := content.LoadCollection[FrontMatter]("content/posts")
	if err != nil {
		log.Fatal(err)
	}

	var tiles []pages.Tile
	for _, it := range collection.Items {
		tiles = append(tiles, pages.Tile{
			Title: it.Meta.Title, Description: it.Meta.Description, Slug: it.Slug,
		})
	}

	return ssg.Node{
		Title: "Posts",
		Page: func(c ssg.Ctx) templ.Component {
			return pages.Posts(c.Title, c.Path, c.Prefix, tiles)
		},
		Generate: func() map[string]ssg.Node {
			return collection.Nodes(func(c ssg.Ctx, it content.Item[FrontMatter]) templ.Component {
				return pages.Post(c.Title, c.Path, c.Prefix, it.Meta.Title, it.HTML)
			})
		},
	}
}`

const tutPostsIndex = `package pages

// Tile is one entry in the posts index.
type Tile struct {
	Title       string
	Description string
	Slug        string
}

// Posts is the index half of list + detail: one tile per post, linking to
// the detail page the collection generated.
templ Posts(title string, path string, prefix string, tiles []Tile) {
	@Layout(title, path, prefix) {
		<h1>{ title }</h1>
		if len(tiles) == 0 {
			<p>Nothing here yet.</p>
		}
		for _, t := range tiles {
			<div class="tile">
				<a href={ t.Slug + "/" }>{ t.Title }</a>
				<p>{ t.Description }</p>
			</div>
		}
	}
}`

const tutTagsFilter = `// A tag page is the collection filtered by one tag. The tags themselves come
// from the posts — a map built while loading, one subtree entry per tag.
func tagPosts(tag string) []pages.Tile {
	tagged := posts.Filter(func(it content.Item[FrontMatter]) bool {
		for _, t := range it.Meta.Tags {
			if t == tag {
				return true
			}
		}
		return false
	})

	var tiles []pages.Tile
	for _, it := range tagged.Items {
		tiles = append(tiles, pages.Tile{
			Title: it.Meta.Title, Description: it.Meta.Description, Slug: it.Slug,
		})
	}
	return tiles
}`

const tutPagination = `// Page one is the posts index; the rest are paginated copies of it. Chunk is
// just the arithmetic — turning the groups into pages is your three lines.
"posts": {
	Page: postsIndex(0),
	Generate: func() map[string]ssg.Node {
		out := map[string]ssg.Node{}
		for i, group := range posts.Chunk(5) {
			if i > 0 {
				out["page/"+strconv.Itoa(i+1)] = ssg.Node{Page: postsIndex(i)}
			}
			for _, it := range group {
				out[it.Slug] = ssg.Node{Page: postPage(it)}
			}
		}
		return out
	},
},`

const tutRssFiles = `	Files: map[string]ssg.PageFunc{
		"rss.xml":     rssFeed(posts),
		"sitemap.xml": sitemap(allPages),
	},`

const tutLocale = `	root = ssg.BuildLocales(root, []ssg.Locale{{
		Code: "de",
		Override: map[string]func(ssg.Node) ssg.Node{
			// Only the pages that genuinely differ — the home page, here.
			"": func(n ssg.Node) ssg.Node {
				n.Title = "Mein Blog"
				n.Page = func(c ssg.Ctx) templ.Component {
					return pages.HomeDE(c.Title, c.Path, c.Prefix)
				}
				return n
			},
		},
	}})`

const tutHomeWithCounter = `package pages

import x "github.com/LukasDerBaum42/astwerk-live/x"

// count is the site's one piece of state. It is named so scripts can read it;
// every page that renders counter() shares it.
var count = x.NamedNumSignal("count", 0)

templ Home(title string, path string, prefix string) {
	@Layout(title, path, prefix) {
		<h1>My Blog</h1>
		@x.Document(counter())
	}
}

// counter is a live widget: the markers below render the value and generate
// the JavaScript that keeps it in sync.
templ counter() {
	<div class="counter">
		@count
		@x.El("button", x.On("click", count.Inc(1))) {
			+
		}
	</div>
}`

const tutMainServe = `package main

import (
	"context"
	"log"
	"os"

	"github.com/LukasDerBaum42/astwerk/content"
	"github.com/LukasDerBaum42/astwerk/devserver"
	"github.com/LukasDerBaum42/astwerk/ssg"
	"github.com/a-h/templ"
	"example.com/myblog/pages"
)

type FrontMatter struct {
	Title       string
	Description string
	Date        string
	Tags        []string
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--serve" {
		devserver.Run(context.Background(), devserver.Config{
			Build: devserver.Steps(
				devserver.Command("templ", "generate"),
				devserver.Command("go", "run", ".", "--dev"),
			),
		})
		return
	}
	dev := len(os.Args) > 1 && os.Args[1] == "--dev"
	if err := ssg.Build(tree(), ssg.BuildOptions{Dev: dev}); err != nil {
		log.Fatal(err)
	}
}

func tree() ssg.Node {
	root := ssg.Node{
		Title: "My Blog",
		Page:  ssg.Templ(pages.Home),
		Children: map[string]ssg.Node{
			"about": {Title: "About", Page: ssg.Templ(pages.About)},
			"posts": posts(),
			"style": {CopyFrom: "style"},
		},
	}
	return ssg.BuildLocales(root, []ssg.Locale{{Code: "de"}})
}

// posts is unchanged from step 5.`

// samplePostSource is the markdown the step-4 browser frame renders with the
// site's own goldmark config at build time.
const samplePostSource = `+++
Title = "Hello, blog"
Description = "The first post."
Date = "2026-08-28"
Tags = ["meta", "astwerk"]
+++

This is the first post. The front matter between the +++ markers is TOML;
the rest is markdown, rendered to HTML by the build.

## Why static

No database, no server, no runtime to patch. A static site is a folder.

- one build, everywhere
- nothing to patch
- plain HTML out`
