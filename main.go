// Command site builds the astwerk documentation site.
//
// It depends on a published tag of astwerk, the same way any other consumer
// does, rather than on a working tree next door. That makes it a check on the
// release rather than on the branch: if this builds, the API as tagged is
// usable as tagged. To try an unreleased library change, use a Go workspace
// (`go work init . ../astwerk`) instead of editing go.mod.
//
//	go run .              # build into build/
//	go run . --dev        # keep the existing output, overwrite in place
//	go run . --serve      # build, serve on localhost:8080, rebuild on change
//	go run . --base ""    # build for a domain root instead of /astwerk-website
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/LukasDerBaum42/astwerk-website/pages"
	"github.com/LukasDerBaum42/astwerk-website/templates"
	"github.com/LukasDerBaum42/astwerk/content"
	"github.com/LukasDerBaum42/astwerk/devserver"
	"github.com/LukasDerBaum42/astwerk/ssg"
	"github.com/a-h/templ"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

// DefaultBase is where GitHub project Pages serve this repo from. Override it
// with -base "" if the site ever moves to a domain root.
const DefaultBase = "/astwerk-website"

// FrontMatter is the shape of the TOML block in content/docs/*.md.
type FrontMatter struct {
	Title       string
	Description string
	Order       int
	Group       string // sidebar section; pages default to one shared group
}

// doc is one loaded documentation page.
type doc struct {
	Slug string
	FM   FrontMatter
	HTML string
	Subs []templates.SubItem // this page's own h2 headings, for the sidebar
}

func main() {
	dev := flag.Bool("dev", false, "keep the existing build directory and overwrite in place")
	serve := flag.Bool("serve", false, "serve the build directory and rebuild on change")
	addr := flag.String("addr", devserver.DefaultAddr, "address the dev server listens on")
	base := flag.String("base", DefaultBase, "path the site is served from")
	out := flag.String("out", "build", "output directory")
	flag.Parse()

	if *serve {
		runServer(*out, *base, *addr)
		return
	}

	n, err := buildSite(*out, *base, *dev)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("built %d pages into %s/ (relative URLs, canonical base %q)", n, *out, *base)
}

// runServer builds the site on every change and reloads the browser.
//
// The rebuild runs out of process — templ generate, then this same program
// again without --serve. That is not indirection for its own sake: every page
// here is a compiled templ component, and a running Go program cannot load new
// code. Rebuilding in-process would re-render the components baked into the
// binary that started the server, so editing a .templ would reload the browser
// and change nothing.
//
// The inner build runs in dev mode: overwriting in place keeps a page's assets
// readable while the rebuild is halfway through.
func runServer(outDir, base, addr string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	templ, templArgs := templCommand()

	err := devserver.Run(ctx, devserver.Config{
		Dir:  outDir,
		Addr: addr,
		Build: devserver.Steps(
			devserver.Command(templ, templArgs...),
			devserver.Command("go", "run", ".", "--out", outDir, "--base", base, "--dev"),
		),
	})
	if err != nil {
		log.Fatal(err)
	}
}

// TemplVersion is the templ release this site generates with. Keep it in step
// with the version the CI workflows pin.
const TemplVersion = "v0.3.1020"

// templCommand is how to run templ generate, preferring an installed binary.
//
// `go run github.com/a-h/templ/cmd/templ@version` always works and needs
// nothing installed, but it costs about three and a half seconds even warm —
// too slow to sit in a save-and-look loop. The installed binary does the same
// job in about two tenths of a second, so it is worth looking for, with the
// go run form kept as the fallback that cannot fail to exist.
func templCommand() (name string, args []string) {
	candidates := []string{"templ"}
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		candidates = append(candidates, filepath.Join(gobin, "templ"))
	}
	if out, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		if gopath := strings.TrimSpace(string(out)); gopath != "" {
			candidates = append(candidates, filepath.Join(gopath, "bin", "templ"))
		}
	}

	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path, []string{"generate"}
		}
	}
	return "go", []string{"run", "github.com/a-h/templ/cmd/templ@" + TemplVersion, "generate"}
}

// buildSite generates the site and returns how many pages it wrote. Separated
// from main so tests can build into a temp directory and inspect the output.
func buildSite(outDir, base string, dev bool) (int, error) {
	docs, err := loadDocs("content/docs")
	if err != nil {
		return 0, err
	}
	pager := make([]templates.NavItem, 0, len(docs))
	summaries := make([]pages.DocSummary, 0, len(docs))
	for _, d := range docs {
		pager = append(pager, templates.NavItem{Label: d.FM.Title, Path: d.Path()})
		summaries = append(summaries, pages.DocSummary{
			Title: d.FM.Title, Description: d.FM.Description, Path: d.Path(),
		})
	}
	sidebar := groupNav(docs)

	root := ssg.Node{
		Title: "astwerk",
		Page:  func(c ssg.Ctx) templ.Component { return pages.Home(c) },
		Files: map[string]ssg.PageFunc{
			// GitHub Pages runs Jekyll unless told not to, and Jekyll drops files
			// and directories beginning with an underscore.
			".nojekyll": empty,
			"404.html":  func(c ssg.Ctx) templ.Component { return pages.NotFound(c) },
		},
		Children: map[string]ssg.Node{
			"docs": {
				Title: "Documentation",
				Page: func(c ssg.Ctx) templ.Component {
					return pages.DocsIndex(c, sidebar, summaries)
				},
				Generate: func() map[string]ssg.Node { return docNodes(docs, sidebar, pager) },
			},
			"tutorial": tutorialRoot(),
			"examples": exampleRoot(),
			"style":    {CopyFrom: "style"},
			"public":   {CopyFrom: "public"},
			"scripts":  {CompileFrom: "scripts"},
		},
	}

	if err := ssg.Build(root, ssg.BuildOptions{
		OutDir: outDir,
		Dev:    dev,
		// BaseURL only feeds Ctx.URL, for canonical tags. Every in-page link and
		// asset is relative, so this output works opened from disk, served at a
		// localhost root, and deployed under /astwerk-website/ — without a rebuild.
		BaseURL:      base,
		RelativeURLs: true,
	}); err != nil {
		return 0, err
	}
	return len(docs) + len(flattenNav(pages.ExampleNav)) + len(flattenNav(pages.TutorialNav)) + 4, nil
}

// exampleRoot builds the examples/ subtree: one page per entry in
// pages.ExampleNav, in the same order, with prev/next wired from that order —
// the same pattern navRoot uses for the tutorial and docNodes for the docs.
func exampleRoot() ssg.Node {
	renderers := []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component{
		pages.ExampleOverview,
		pages.ExampleXCounter,
		pages.ExampleXTodo,
		pages.ExampleXForms,
		pages.ExampleXBindings,
		pages.ExampleXAsync,
		pages.ExampleXCanvas,
		pages.ExampleXFaq,
		pages.ExampleXUnderTheHood,
		pages.ExampleCounter,
		pages.ExampleList,
		pages.ExampleForms,
		pages.ExampleBindings,
		pages.ExampleAsync,
		pages.ExampleCanvas,
		pages.ExampleStopwatch,
		pages.ExampleFilter,
		pages.ExampleTemplBasics,
		pages.ExampleTreeAndContent,
		pages.ExampleScriptsAndRouting,
		pages.ExampleWasmwrapAndTempl,
	}
	return navRoot("examples", pages.ExampleNav, renderers)
}

// tutorialRoot builds the tutorial/ subtree: an overview and ten steps, with
// prev/next wired from the step order.
func tutorialRoot() ssg.Node {
	renderers := []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component{
		pages.TutorialOverview,
		pages.TutorialSetup,
		pages.TutorialTree,
		pages.TutorialLayout,
		pages.TutorialMorePages,
		pages.TutorialCollection,
		pages.TutorialStyles,
		pages.TutorialLocale,
		pages.TutorialInteractivity,
		pages.TutorialDevloop,
		pages.TutorialShipIt,
	}
	return navRoot("tutorial", pages.TutorialNav, renderers)
}

// navRoot builds a paged subtree from a sidebar: one page per entry after the
// first (which is the section's own index), with prev/next wired from the
// sidebar's order. The slug of each child page is the last path segment of its
// entry.
func navRoot(section string, nav []templates.NavGroup, renderers []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component) ssg.Node {
	items := flattenNav(nav)
	if len(items) != len(renderers) {
		log.Fatalf("%s: %d nav entries but %d renderers — nav and renderers are out of sync",
			section, len(items), len(renderers))
	}

	children := make(map[string]ssg.Node, len(items)-1)
	for i := 1; i < len(items); i++ {
		i, render := i, renderers[i]
		prev := &items[i-1]
		var next *templates.NavItem
		if i < len(items)-1 {
			next = &items[i+1]
		}

		parts := strings.Split(strings.Trim(items[i].Path, "/"), "/")
		children[parts[len(parts)-1]] = ssg.Node{
			Title: items[i].Label,
			Page:  func(c ssg.Ctx) templ.Component { return render(c, prev, next) },
		}
	}

	var overviewNext *templates.NavItem
	if len(items) > 1 {
		overviewNext = &items[1]
	}
	overview := renderers[0]

	return ssg.Node{
		Title:    section,
		Page:     func(c ssg.Ctx) templ.Component { return overview(c, nil, overviewNext) },
		Children: children,
	}
}

// flattenNav flattens a grouped sidebar into reading order, for wiring a
// previous/next pager the same way docNodes does for docs.
func flattenNav(groups []templates.NavGroup) []templates.NavItem {
	var out []templates.NavItem
	for _, g := range groups {
		out = append(out, g.Items...)
	}
	return out
}

// docNodes turns loaded markdown into one Node per page, wiring up the
// previous/next pager from reading order (docs' own Order field, independent
// of how the sidebar groups them).
func docNodes(docs []doc, sidebar []templates.NavGroup, pager []templates.NavItem) map[string]ssg.Node {
	out := make(map[string]ssg.Node, len(docs))

	for i, d := range docs {
		var prev, next *templates.NavItem
		if i > 0 {
			prev = &pager[i-1]
		}
		if i < len(docs)-1 {
			next = &pager[i+1]
		}

		body, desc := d.HTML, d.FM.Description
		out[d.Slug] = ssg.Node{
			Title: d.FM.Title,
			Page: func(c ssg.Ctx) templ.Component {
				return pages.DocPage(c, sidebar, desc, body, prev, next)
			},
		}
	}
	return out
}

// groupNav clusters docs into sidebar sections by FrontMatter.Group, in the
// order each group first appears — so the docs' own Order field decides
// section order too, with no separate list to keep in sync.
func groupNav(docs []doc) []templates.NavGroup {
	var groups []templates.NavGroup
	index := map[string]int{}

	for _, d := range docs {
		i, ok := index[d.FM.Group]
		if !ok {
			i = len(groups)
			index[d.FM.Group] = i
			groups = append(groups, templates.NavGroup{Title: d.FM.Group})
		}
		groups[i].Items = append(groups[i].Items, templates.NavItem{Label: d.FM.Title, Path: d.Path(), Sub: d.Subs})
	}
	return groups
}

var headingRe = regexp.MustCompile(`(?s)<h([23])>(.*?)</h[23]>`)
var tagRe = regexp.MustCompile(`<[^>]+>`)

// withHeadingAnchors gives every h2/h3 in html an id (goldmark doesn't add
// one on its own) and collects them as sidebar subsections — h2s top-level,
// h3s nested under them. The h3s are the API names a docs page is organised
// by, so listing them keeps "on this page" useful.
func withHeadingAnchors(html string) (string, []templates.SubItem) {
	var subs []templates.SubItem
	seen := map[string]int{}

	patched := headingRe.ReplaceAllStringFunc(html, func(m string) string {
		parts := headingRe.FindStringSubmatch(m)
		level, inner := parts[1], parts[2]
		text := strings.TrimSpace(tagRe.ReplaceAllString(inner, ""))

		slug := slugify(text)
		if n := seen[slug]; n > 0 {
			slug = fmt.Sprintf("%s-%d", slug, n)
		}
		seen[slug]++

		if level == "2" || level == "3" {
			subs = append(subs, templates.SubItem{Label: text, Anchor: "#" + slug, Deep: level == "3"})
		}
		return fmt.Sprintf(`<h%s id="%s">%s</h%s>`, level, slug, inner, level)
	})
	return patched, subs
}

// slugify makes a heading's text safe as a URL fragment: lowercase, letters
// and digits kept, everything else collapsed to a single hyphen.
func slugify(s string) string {
	var b strings.Builder
	dash := true // starts true so a leading run of punctuation is dropped, not hyphenated
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// loadDocs reads the docs directory and returns pages in sidebar order.
func loadDocs(dir string) ([]doc, error) {
	pages, err := content.NewLoader(markdown()).LoadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no markdown found in %s", dir)
	}

	// Slugs gives a stable alphabetical base; Order then decides the reading
	// sequence, so two pages sharing an Order still come out deterministically.
	docs := make([]doc, 0, len(pages))
	for _, slug := range content.Slugs(pages) {
		fm, err := content.Decode[FrontMatter](pages[slug])
		if err != nil {
			return nil, fmt.Errorf("front matter in %s/%s.md: %w", dir, slug, err)
		}
		if fm.Title == "" {
			return nil, fmt.Errorf("%s/%s.md has no Title in its front matter", dir, slug)
		}
		html, subs := withHeadingAnchors(pages[slug].HTML)
		docs = append(docs, doc{Slug: slug, FM: fm, HTML: html, Subs: subs})
	}

	for i := 1; i < len(docs); i++ {
		for j := i; j > 0 && docs[j].FM.Order < docs[j-1].FM.Order; j-- {
			docs[j], docs[j-1] = docs[j-1], docs[j]
		}
	}
	return docs, nil
}

// Path is the doc's site path, matching where the tree puts it.
func (d doc) Path() string { return "docs/" + d.Slug + "/" }

// markdown is the renderer for docs pages: syntax highlighting via chroma,
// GitHub-flavoured tables and strikethrough, anchored headings, and raw HTML
// passed through so a page can drop into markup.
func markdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle(templates.StyleName),
				// Classes, not inline styles, so style/code.css drives both these
				// blocks and the ones views.HighlightGo emits.
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
	)
}

// empty writes nothing, for marker files like .nojekyll that must exist but
// stay empty.
func empty(ssg.Ctx) templ.Component {
	return templ.ComponentFunc(func(context.Context, io.Writer) error { return nil })
}
