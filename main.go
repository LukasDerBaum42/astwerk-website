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
	"encoding/json"
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

	docspages "github.com/LukasDerBaum42/astwerk-website/pages/docs"
	"github.com/LukasDerBaum42/astwerk-website/pages/examples"
	"github.com/LukasDerBaum42/astwerk-website/pages/home"
	"github.com/LukasDerBaum42/astwerk-website/pages/tutorial"
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

// versionRe matches the astwerk require line in go.mod — the version the docs
// describe, not a working tree a go.work might redirect to.
var versionRe = regexp.MustCompile(`(?m)^\s*github\.com/LukasDerBaum42/astwerk\s+v([^\s]+)`)

// astwerkVersion returns the pinned astwerk version, or "dev" when go.mod
// can't be read or doesn't pin it.
func astwerkVersion() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "dev"
	}
	if m := versionRe.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return "dev"
}

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

	templates.Version = astwerkVersion()
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
	// Production output ships the compact x runtime; dev keeps it readable. The
	// pages read the flag through templates.XOptions, so the choice is one place.
	templates.MinifiedX = !dev

	docs, err := loadDocs("content/docs")
	if err != nil {
		return 0, err
	}
	pager := make([]templates.NavItem, 0, len(docs))
	summaries := make([]docspages.DocSummary, 0, len(docs))
	for _, d := range docs {
		pager = append(pager, templates.NavItem{Label: d.FM.Title, Path: d.Path()})
		summaries = append(summaries, docspages.DocSummary{
			Title: d.FM.Title, Description: d.FM.Description, Path: d.Path(),
		})
	}
	sidebar := groupNav(docs)

	// The site is its own i18n demo: English is the base tree, German is a
	// BuildLocales subtree whose docs overlay translates the beginner path. The
	// rest of the de subtree falls back to English behind a banner.
	deDocs, translated, err := loadDocsMerged("content/docs", "content/de/docs")
	if err != nil {
		return 0, err
	}
	templates.Translated = map[string]bool{"": true, "docs/": true}
	for slug := range translated {
		templates.Translated["docs/"+slug+"/"] = true
	}

	// Site-wide generated files. They live on the root node, so the de subtree
	// mirrors them too — which is what lets the de search box resolve its index
	// relative to its own page.
	entries := searchEntries(docs)
	origin := "https://lukasderbaum42.github.io" + base

	root := ssg.Node{
		Title: "astwerk",
		Page:  func(c ssg.Ctx) templ.Component { return home.Home(c) },
		Files: map[string]ssg.PageFunc{
			// GitHub Pages runs Jekyll unless told not to, and Jekyll drops files
			// and directories beginning with an underscore.
			".nojekyll":         empty,
			"404.html":          func(c ssg.Ctx) templ.Component { return docspages.NotFound(c) },
			"search/index.json": jsonFile(entries),
			"sitemap.xml":       textFile(sitemapXML(origin, docs)),
			"robots.txt": textFile("User-agent: *\nAllow: /\n\n" +
				"Sitemap: " + origin + "/sitemap.xml\n"),
		},
		Children: map[string]ssg.Node{
			"docs": {
				Title: "Documentation",
				Page: func(c ssg.Ctx) templ.Component {
					return docspages.DocsIndex(c, sidebar, summaries)
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

	root = ssg.BuildLocales(root, deLocales(deDocs, translated))

	if err := ssg.Build(root, ssg.BuildOptions{
		OutDir: outDir,
		Dev:    dev,
		// Sibling nodes render concurrently; the closures above were written
		// capture-safe, so the output is byte-identical to the sequential build.
		Parallel: true,
		// BaseURL only feeds Ctx.URL, for canonical tags. Every in-page link and
		// asset is relative, so this output works opened from disk, served at a
		// localhost root, and deployed under /astwerk-website/ — without a rebuild.
		BaseURL:      base,
		RelativeURLs: true,
	}); err != nil {
		return 0, err
	}
	return len(docs) + len(flattenNav(examples.ExampleNav)) + len(flattenNav(tutorial.TutorialNav)) + 4, nil
}

// deLocales is the German locale: the home page and the docs subtree are
// overridden with translated versions; everything else mirrors the base tree
// (and shows the "not translated yet" banner from the shared layout).
func deLocales(deDocs []doc, translated map[string]bool) []ssg.Locale {
	return []ssg.Locale{{
		Code: "de",
		Override: map[string]func(ssg.Node) ssg.Node{
			// The German home page: the same demo, translated.
			"": func(n ssg.Node) ssg.Node {
				n.Title = "astwerk"
				n.Page = func(c ssg.Ctx) templ.Component { return home.HomeDE(c) }
				return n
			},
			// The docs subtree: German front matter and content where a
			// translation exists, English otherwise — sidebar, index and pager
			// all built from the merged docs.
			"docs": func(n ssg.Node) ssg.Node {
				deSidebar := groupNav(deDocs)
				dePager := make([]templates.NavItem, 0, len(deDocs))
				deSummaries := make([]docspages.DocSummary, 0, len(deDocs))
				for _, d := range deDocs {
					dePager = append(dePager, templates.NavItem{Label: d.FM.Title, Path: d.Path()})
					deSummaries = append(deSummaries, docspages.DocSummary{
						Title: d.FM.Title, Description: d.FM.Description, Path: d.Path(),
					})
				}
				n.Page = func(c ssg.Ctx) templ.Component {
					return docspages.DocsIndex(c, deSidebar, deSummaries)
				}
				n.Generate = func() map[string]ssg.Node {
					return docNodes(deDocs, deSidebar, dePager)
				}
				return n
			},
		},
	}}
}

// exampleRoot builds the examples/ subtree: one page per entry in
// examples.ExampleNav, in the same order, with prev/next wired from that order —
// the same pattern navRoot uses for the tutorial and docNodes for the docs.
func exampleRoot() ssg.Node {
	renderers := []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component{
		examples.ExampleOverview,
		examples.ExampleXCounter,
		examples.ExampleXTodo,
		examples.ExampleXForms,
		examples.ExampleXBindings,
		examples.ExampleXAsync,
		examples.ExampleXFaq,
		examples.ExampleXUnderTheHood,
		examples.ExampleCounter,
		examples.ExampleList,
		examples.ExampleForms,
		examples.ExampleBindings,
		examples.ExampleAsync,
		examples.ExampleCanvas,
		examples.ExampleStopwatch,
		examples.ExampleFilter,
		examples.ExampleTemplBasics,
		examples.ExampleTreeAndContent,
		examples.ExampleScriptsAndRouting,
		examples.ExampleWasmwrapAndTempl,
		examples.ExampleRouterDemo,
		examples.ExampleBridge,
	}

	root := navRoot("examples", examples.ExampleNav, renderers)

	// The async demo fetches a file the site itself generated — the same three
	// posts the x-async twin shows, written as data next to the examples pages.
	root.Files = map[string]ssg.PageFunc{
		"data/posts.json": jsonFile([]struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}{
			{ID: "1", Title: "Signals track their own dependencies"},
			{ID: "2", Title: "Keyed lists move nodes instead of rebuilding them"},
			{ID: "3", Title: "Scripts are the escape hatch — this is one"},
		}),
	}

	// The router demo owns its own URL, so it is a standalone document — its
	// own <head>, the wasm loader, no site layout — embedded via iframe from
	// the Full apps page. It is deliberately not a nav entry: the frame's URL
	// belongs to the demo, not to the site's navigation.
	root.Children["router"] = ssg.Node{
		Title: "Router demo",
		Page:  func(c ssg.Ctx) templ.Component { return examples.RouterShell(c) },
	}

	return root
}

// tutorialRoot builds the tutorial/ subtree: the blog line (an overview and
// eleven steps, with prev/next wired from the step order) and the clicker
// line under tutorial/clicker/, its own paged subtree.
func tutorialRoot() ssg.Node {
	renderers := []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component{
		tutorial.TutorialOverview,
		tutorial.TutorialSetup,
		tutorial.TutorialTree,
		tutorial.TutorialLayout,
		tutorial.TutorialPosts,
		tutorial.TutorialCollection,
		tutorial.TutorialTags,
		tutorial.TutorialRss,
		tutorial.TutorialLocale,
		tutorial.TutorialInteractivity,
		tutorial.TutorialDevloop,
		tutorial.TutorialShipIt,
	}
	clickerRenderers := []func(ssg.Ctx, *templates.NavItem, *templates.NavItem) templ.Component{
		tutorial.TutorialClicker1,
		tutorial.TutorialClicker2,
		tutorial.TutorialClicker3,
		tutorial.TutorialClicker4,
		tutorial.TutorialClicker5,
		tutorial.TutorialClicker6,
		tutorial.TutorialClicker7,
	}

	root := navRoot("tutorial", tutorial.TutorialNav, renderers)

	// The clicker line: every quest is a child page (the group has no index
	// of its own — tutorial/clicker/ redirects to quest 1), with prev/next
	// wired from the line's own order.
	clickerChildren := make(map[string]ssg.Node, len(clickerRenderers))
	clickerItems := tutorial.ClickerNav[0].Items
	for i := range clickerRenderers {
		render := clickerRenderers[i]
		var prev, next *templates.NavItem
		if i > 0 {
			prev = &clickerItems[i-1]
		}
		if i < len(clickerItems)-1 {
			next = &clickerItems[i+1]
		}
		parts := strings.Split(strings.Trim(clickerItems[i].Path, "/"), "/")
		clickerChildren[parts[len(parts)-1]] = ssg.Node{
			Title: clickerItems[i].Label,
			Page:  func(c ssg.Ctx) templ.Component { return render(c, prev, next) },
		}
	}
	root.Children["clicker"] = ssg.Node{
		Title:    "Cookie Clicker",
		Page:     func(c ssg.Ctx) templ.Component { return tutorial.ClickerIndex(c) },
		Children: clickerChildren,
	}
	return root
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
				return docspages.DocPage(c, sidebar, desc, body, prev, next)
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
	docs, _, err := loadDocsMerged(dir, "")
	return docs, err
}

// loadDocsMerged reads dir as a Collection, overlaying translated pages from
// overlayDir when a slug exists in both. The returned set names the slugs that
// came from the overlay — the pages that are actually translated.
func loadDocsMerged(dir, overlayDir string) ([]doc, map[string]bool, error) {
	loader := content.NewLoader(markdown())
	docs, err := content.LoadCollectionWith[FrontMatter](loader, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	if len(docs.Items) == 0 {
		return nil, nil, fmt.Errorf("no markdown found in %s", dir)
	}

	translated := map[string]bool{}
	if overlayDir != "" {
		// The German overlay is the documented escape hatch, worked: load the
		// translated pages by hand with LoadDir and replace an item by slug —
		// the same shape as content-and-collections' "one odd item can still be
		// built by hand and merged" line.
		overlay, err := loader.LoadDir(overlayDir)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s: %w", overlayDir, err)
		}
		for slug, p := range overlay {
			meta, err := content.Decode[FrontMatter](p)
			if err != nil {
				return nil, nil, fmt.Errorf("front matter in %s/%s.md: %w", overlayDir, slug, err)
			}
			for i := range docs.Items {
				if docs.Items[i].Slug == slug {
					docs.Items[i].Meta = meta
					docs.Items[i].HTML = p.HTML
					break
				}
			}
			translated[slug] = true
		}
	}

	out := make([]doc, 0, len(docs.Items))
	for _, it := range docs.Items {
		if it.Meta.Title == "" {
			return nil, nil, fmt.Errorf("%s/%s.md has no Title in its front matter", dir, it.Slug)
		}
		html, subs := withHeadingAnchors(it.HTML)
		out = append(out, doc{Slug: it.Slug, FM: it.Meta, HTML: html, Subs: subs})
	}

	// Collection is sorted by slug; the sidebar and pager need reading order,
	// which is the docs' own Order field — the documented "sort by what the
	// site needs" case.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].FM.Order < out[j-1].FM.Order; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out, translated, nil
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

// searchEntry is one page in the build-time search index.
type searchEntry struct {
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Headings    string `json:"headings,omitempty"`
}

// searchEntries gathers every page the site generates: docs with their
// descriptions and headings, tutorial and example steps by title.
func searchEntries(docs []doc) []searchEntry {
	entries := []searchEntry{{Path: "", Title: "astwerk", Description: "A Go-first static web framework with an optional client runtime."}}
	for _, d := range docs {
		var headings []string
		for _, s := range d.Subs {
			headings = append(headings, s.Label)
		}
		entries = append(entries, searchEntry{
			Path: d.Path(), Title: d.FM.Title, Description: d.FM.Description,
			Headings: strings.Join(headings, " "),
		})
	}
	for _, g := range tutorial.TutorialNav {
		for _, it := range g.Items {
			entries = append(entries, searchEntry{Path: it.Path, Title: it.Label})
		}
	}
	for _, g := range tutorial.ClickerNav {
		for _, it := range g.Items {
			entries = append(entries, searchEntry{Path: it.Path, Title: it.Label})
		}
	}
	for _, g := range examples.ExampleNav {
		for _, it := range g.Items {
			entries = append(entries, searchEntry{Path: it.Path, Title: it.Label})
		}
	}
	return entries
}

// sitemapXML lists every page in both locales — the site's own SEO recipe.
func sitemapXML(origin string, docs []doc) string {
	paths := []string{""}
	for _, d := range docs {
		paths = append(paths, d.Path())
	}
	for _, g := range tutorial.TutorialNav {
		for _, it := range g.Items {
			paths = append(paths, it.Path)
		}
	}
	for _, g := range tutorial.ClickerNav {
		for _, it := range g.Items {
			paths = append(paths, it.Path)
		}
	}
	for _, g := range examples.ExampleNav {
		for _, it := range g.Items {
			paths = append(paths, it.Path)
		}
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, p := range paths {
		b.WriteString("  <url><loc>" + origin + "/" + p + "</loc></url>\n")
		b.WriteString("  <url><loc>" + origin + "/de/" + p + "</loc></url>\n")
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

// textFile is a Files entry writing a literal string.
func textFile(s string) ssg.PageFunc {
	return func(ssg.Ctx) templ.Component {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := io.WriteString(w, s)
			return err
		})
	}
}

// jsonFile is a Files entry marshaling v as indented JSON.
func jsonFile(v any) ssg.PageFunc {
	return func(ssg.Ctx) templ.Component {
		return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "\t")
			return enc.Encode(v)
		})
	}
}

// empty writes nothing, for marker files like .nojekyll that must exist but
// stay empty.
func empty(ssg.Ctx) templ.Component {
	return templ.ComponentFunc(func(context.Context, io.Writer) error { return nil })
}
