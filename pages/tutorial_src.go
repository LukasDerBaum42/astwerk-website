package pages

// Source snippets for the tutorial steps. In a plain .go file: templ rewrites
// raw strings containing templ syntax into generated Go, and a snippet is
// meant to show the source.

const tutStep2Main = `package main

import (
	"log"

	"github.com/LukasDerBaum42/astwerk/ssg"
	"example.com/mysite/pages"
)

func main() {
	root := ssg.Node{
		Title: "My Site",
		Page:  ssg.Templ(pages.Home),
	}
	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
		log.Fatal(err)
	}
}`

const tutStep2Home = `package pages

templ Home(title string, path string, prefix string) {
	<h1>My Site</h1>
	<p>A portfolio built with astwerk.</p>
}`

const tutLayout = `package pages

// Layout is the shell every page wraps: html/head/body, the nav, and
// { children... } where the page's own content goes.
templ Layout(title string, path string, prefix string) {
	<!DOCTYPE html>
	<html lang={ lang(prefix) }>
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>{ title }</title>
		</head>
		<body>
			<nav>
				<a href={ prefix + "/" }>My Site</a>
				<a href={ prefix + "/projects/" }>Projects</a>
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
		<h1>My Site</h1>
		<p>A portfolio built with astwerk.</p>
	}
}`

const tutAbout = `package pages

templ About(title string, path string, prefix string) {
	@Layout(title, path, prefix) {
		<h1>About</h1>
		<p>This site is built with astwerk.</p>
	}
}`

const tutTreeAddAbout = `	root := ssg.Node{
		Title: "My Site",
		Page:  ssg.Templ(pages.Home),
		Children: map[string]ssg.Node{
			"about": {Title: "About", Page: ssg.Templ(pages.About)},
		},
	}`

const tutProjects = `package pages

// Tile is one entry in the projects index.
type Tile struct {
	Title       string
	Description string
	Slug        string
}

templ Projects(title string, path string, prefix string, tiles []Tile) {
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
}

// Project renders one markdown file. The body is markdown the build already
// rendered to HTML, so it is trusted markup.
templ Project(title string, path string, prefix string, body string) {
	@Layout(title, path, prefix) {
		<h1>{ title }</h1>
		@templ.Raw(body)
	}
}`

const tutMain = `package main

import (
	"log"

	"github.com/LukasDerBaum42/astwerk/content"
	"github.com/LukasDerBaum42/astwerk/ssg"
	"github.com/a-h/templ"
	"example.com/mysite/pages"
)

// FrontMatter is the shape of the TOML block in content/projects/*.md.
type FrontMatter struct {
	Title       string
	Description string
}

func main() {
	root := ssg.Node{
		Title: "My Site",
		Page:  ssg.Templ(pages.Home),
		Children: map[string]ssg.Node{
			"about":    {Title: "About", Page: ssg.Templ(pages.About)},
			"projects": projects(),
		},
	}
	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
		log.Fatal(err)
	}
}

// projects reads content/projects and returns the collection subtree: an
// index page and one page per markdown file, generated at build time.
func projects() ssg.Node {
	md, err := content.LoadDir("content/projects")
	if err != nil {
		log.Fatal(err)
	}

	var tiles []pages.Tile
	children := map[string]ssg.Node{}
	for _, slug := range content.Slugs(md) {
		fm, _ := content.Decode[FrontMatter](md[slug])
		body := md[slug].HTML

		children[slug] = ssg.Node{
			Title: fm.Title,
			Page: func(c ssg.Ctx) templ.Component {
				return pages.Project(c.Title, c.Path, c.Prefix, body)
			},
		}
		tiles = append(tiles, pages.Tile{Title: fm.Title, Description: fm.Description, Slug: slug})
	}

	return ssg.Node{
		Title: "Projects",
		Page: func(c ssg.Ctx) templ.Component {
			return pages.Projects(c.Title, c.Path, c.Prefix, tiles)
		},
		Children: children,
	}
}`

const tutProjectMD = `+++
Title = "Hello"
Description = "The first project."
+++

The first thing this site can show. Markdown body, rendered by the build.`

const tutStyleLink = `	<link rel="stylesheet" href={ prefix + "/style/style.css" }/>`

const tutStyleTree = `		Children: map[string]ssg.Node{
			"about":    {Title: "About", Page: ssg.Templ(pages.About)},
			"projects": projects(),
			"style":    {CopyFrom: "style"},
		},`

const tutLocale = `	return ssg.BuildLocales(root, []ssg.Locale{{Code: "de"}})`

const tutHomeWithCounter = `package pages

import x "github.com/LukasDerBaum42/astwerk/x"

// count is the site's one piece of state. It is named so scripts can read it;
// every page that renders counter() shares it.
var count = x.Named("count", 0)

templ Home(title string, path string, prefix string) {
	@Layout(title, path, prefix) {
		<h1>My Site</h1>
		<p>A portfolio built with astwerk.</p>
		@x.Document(counter())
	}
}

// counter is a live widget: the markers below render the value and generate
// the JavaScript that keeps it in sync.
templ counter() {
	<div class="counter">
		@x.Text(count)
		@x.El("button", x.On("click", count.Set(count.Add(1)))) {
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
	"example.com/mysite/pages"
)

type FrontMatter struct {
	Title       string
	Description string
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
		Title: "My Site",
		Page:  ssg.Templ(pages.Home),
		Children: map[string]ssg.Node{
			"about":    {Title: "About", Page: ssg.Templ(pages.About)},
			"projects": projects(),
			"style":    {CopyFrom: "style"},
		},
	}
	return ssg.BuildLocales(root, []ssg.Locale{{Code: "de"}})
}

// projects is unchanged from step 5.`
