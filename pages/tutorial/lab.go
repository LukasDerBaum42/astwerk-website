package tutorial

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"

	"github.com/a-h/templ"
)

// The step lab (06-step-lab.md): each quest's "whole change" rendered as an
// editable frame whose slots validate against expected answers baked in at
// build time. The lab never compiles Go in the browser — it analyzes what the
// reader chose against the line it would render, compared to the expected
// result the page shows. The canonical answer is the reveal value and the
// render comparison's source; it is never compared as a raw string.
//
// The lab tests the quest's logic, not memory: order files rebuild a
// scrambled change, pick slots offer the plausible candidates to drop in, and
// every slot with a render template shows a live preview of "what your code
// would make" as the reader works. Only a value the reader must compute (a
// group size, a start value) stays a typed input.

// labRelate is the relate tier's declared relationship: the value must
// compare to With (another slot's {{key}}, or a literal) the way Kind says.
type labRelate struct {
	Kind  string // eq | neq | contains
	With  string // another slot's key as {{key}}, or a literal
	Label string // the relationship, named for the feedback line
}

// labSlot is one editable blank in a quest's lab file.
type labSlot struct {
	Key       string   // the {{key}} placeholder in the file's lines
	Label     string   // the blank's name — the input's aria-label
	Check     []string // the check tiers to run: pick, shape, render, relate
	Form      string   // the shape tier's declared form (identifier|int|path|slug|file|quoted|flag)
	Canonical string   // the reveal value and the render comparison's source
	Hint      string   // shown on a wrong first attempt
	RenderTpl string   // the line the value renders into, __SLOT__ where the value goes
	Relate    *labRelate
	Options   []string // when set, the slot is a pick: the candidates the reader drops in
}

// labFile is one editable file of a quest's lab: its lines (with {{key}}
// placeholders in fill mode) and the expected-result strip. Mode is "fill"
// (inline slots — typed or picked) or "order" (the lines scrambled — the
// reader rebuilds the whole change by ordering them, a stronger exercise than
// transcribing).
type labFile struct {
	Name   string
	Mode   string // "fill" (default) | "order"
	Lines  []string
	Slots  []labSlot
	Result string
}

// labSpec is one quest's lab: the editable files and the frame's slot count.
type labSpec struct {
	Files []labFile
}

// labSlots counts the quest's practice units across its files — the ledger's
// cap, the chip's total and the frame's data-lab-slots. A fill file counts
// its slots; an order file is one unit (the whole change, +5 once).
func labSlots(s labSpec) int {
	n := 0
	for _, f := range s.Files {
		if f.Mode == "order" {
			n++
		} else {
			n += len(f.Slots)
		}
	}
	return n
}

// labSpecs maps quest ids to their step lab. Only quests with a lab get an
// entry; quests without one render no frame (06-step-lab.md's slot table,
// adapted). The exercises test the quest's logic: order files rebuild the
// change, pick slots offer the plausible candidates, and only a value the
// reader must compute stays a typed input.
var labSpecs = map[string]labSpec{
	"blog-1": {
		Files: []labFile{{
			Name: "terminal",
			Mode: "order",
			Lines: []string{
				`mkdir myblog && cd myblog`,
				`go mod init example.com/myblog`,
				`go get github.com/LukasDerBaum42/astwerk github.com/a-h/templ`,
			},
			Result: "the setup in order: create the project folder, init the module, then add the two dependencies",
		}},
	},
	"blog-2": {
		Files: []labFile{
			{
				Name: "main.go",
				Lines: []string{
					`func main() {`,
					`	{{tree}}`,
					`	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {`,
					`		log.Fatal(err)`,
					`	}`,
					`}`,
				},
				Slots: []labSlot{
					pickSlot("tree", "the site root", []string{"render"}, `root := ssg.Node{ Title: "My Blog", Page: ssg.Templ(pages.Home) }`,
						"the root node is the site — Page takes the templ component, and the walker fills in the page's title, path and prefix",
						"__SLOT__",
						`root := ssg.Node{ Title: "My Blog", Page: ssg.Templ(pages.Home) }`,
						`root := ssg.Node{ Title: "My Blog", Page: pages.Home }`,
						`root := pages.Home`,
					),
				},
				Result: "the site root is one ssg.Node holding the site title and rendering the home page",
			},
			{
				Name: "pages/home.templ",
				Lines: []string{
					`templ Home(title string, path string, prefix string) {`,
					`	<h1>{{h1}}</h1>`,
					`	<p>Posts about Go, templ and static sites.</p>`,
					`}`,
				},
				Slots: []labSlot{
					pickSlot("h1", "the home page's h1", []string{"render"}, "My Blog",
						"the h1 is the page's heading — a one-page site wears the site title",
						`<h1>__SLOT__</h1>`,
						"My Blog", "Posts", "Hello, blog",
					),
				},
				Result: "the home page's h1 reads My Blog",
			},
		},
	},
	"blog-3": {
		Files: []labFile{{
			Name: "pages/layout.templ",
			Mode: "order",
			Lines: []string{
				`<nav>`,
				`	<a href={ prefix + "/" }>My Blog</a>`,
				`	<a href={ prefix + "/posts/" }>Posts</a>`,
				`</nav>`,
				`<link rel="stylesheet" href={ prefix + "/style/style.css" }/>`,
				`<main>`,
				`	{ children... }`,
				`</main>`,
				`<footer>Built with astwerk</footer>`,
			},
			Result: "the shell in order: the nav, the stylesheet, the page's own content, then the footer",
		}},
	},
	"blog-4": {
		Files: []labFile{{
			Name: "content/posts/hello.md",
			Lines: []string{
				`+++`,
				`{{title}}`,
				`Description = "The first post."`,
				`Date = "2026-08-28"`,
				`Tags = ["{{tag}}", "astwerk"]`,
				`+++`,
			},
			Slots: []labSlot{
				writeSlot("title", "the front-matter title line", `Title = "Hello, blog"`,
					"the field that becomes the post page's h1 — a post's Title, not its Description",
				),
				pickSlot("tag", "a tag for the post", []string{"render"}, "meta",
					"the Tags list already carries astwerk — the new tag must be a different one",
					`Tags = ["__SLOT__", "astwerk"]`,
					"meta", "astwerk", "hello",
				),
			},
			Result: "the post page's h1 reads Hello, blog and the post carries the meta tag",
		}},
	},
	"blog-5": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`collection, err := content.LoadCollection[FrontMatter]("{{dir}}")`,
				`if err != nil {`,
				`	log.Fatal(err)`,
				`}`,
				`return ssg.Node{`,
				`	Title: "{{h1}}",`,
				`	Page: func(c ssg.Ctx) templ.Component {`,
				`		return pages.Posts(c.Title, c.Path, c.Prefix, tiles)`,
				`	},`,
			},
			Slots: []labSlot{
				pickSlot("dir", "the posts directory", []string{"render"}, "content/posts",
					"the directory the markdown lives in, relative to the project root",
					`collection, err := content.LoadCollection[FrontMatter]("__SLOT__")`,
					"content/posts", "posts", "content/",
				),
				pickSlot("h1", "the index's h1", []string{"render"}, "Posts",
					"the index is the collection's own page — it wears the collection's heading",
					`	Title: "__SLOT__",`,
					"Posts", "My Blog", "Hello, blog",
				),
			},
			Result: "the collection loads content/posts and the index page's h1 reads Posts",
		}},
	},
	"blog-6": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`tagged := posts.Filter(func(it content.Item[FrontMatter]) bool {`,
				`	for _, t := range it.Meta.Tags {`,
				`		if t == "{{tag}}" {`,
				`			return true`,
				`		}`,
				`	}`,
				`	return false`,
				`})`,
				`{{size}}`,
			},
			Slots: []labSlot{
				pickSlot("tag", "the tag the page filters by", []string{"render"}, "meta",
					"the tag page filters by one of the tags the posts carry — meta is the one from step 4",
					`		if t == "__SLOT__" {`,
					"meta", "astwerk", "posts",
				),
				writeSlot("size", "the pagination line", `for i, group := range posts.Chunk(5) {`,
					"the Chunk call that splits ten posts into two pages of five — write the whole line",
				),
			},
			Result: "the tag page shows only meta posts and ten posts split into two pages of five",
		}},
	},
	"blog-7": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`Files: map[string]ssg.PageFunc{`,
				`	{{feed}}`,
				`	{{sitemap}}`,
				`},`,
			},
			Slots: []labSlot{
				writeSlot("feed", "the feed entry", `"rss.xml": rssFeed(posts),`,
					"the feed file every reader expects at the site root, written by Node.Files — key, colon, call, comma",
				),
				writeSlot("sitemap", "the sitemap entry", `"sitemap.xml": sitemap(allPages),`,
					"the sitemap file every crawler expects at the site root, written by Node.Files — key, colon, call, comma",
				),
			},
			Result: "the build writes rss.xml and sitemap.xml at the site root",
		}},
	},
	"blog-8": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`root = ssg.BuildLocales(root, []ssg.Locale{{`,
				`	Code: "{{code}}",`,
				`	Override: map[string]func(ssg.Node) ssg.Node{`,
				`		"": func(n ssg.Node) ssg.Node {`,
				`			n.Title = "{{home}}"`,
				`			n.Page = func(c ssg.Ctx) templ.Component {`,
				`				return pages.HomeDE(c.Title, c.Path, c.Prefix)`,
				`			}`,
				`			return n`,
				`		},`,
				`	},`,
				`}})`,
			},
			Slots: []labSlot{
				pickSlot("code", "the locale's code", []string{"render"}, "de",
					"the two-letter language code — the URL prefix becomes /de/",
					`	Code: "__SLOT__",`,
					"de", "deu", "DE",
				),
				pickRelateSlot("home", "the overridden home page's title", []string{"relate", "render"}, "Mein Blog",
					"the override must genuinely differ from the base home's title",
					`			n.Title = "__SLOT__"`, "neq", "My Blog", "the override must differ from the base — you are overriding the page, so it must genuinely differ",
					"Mein Blog", "My Blog", "Der Blog",
				),
			},
			Result: "a German mirror appears under /de/ with the home page titled Mein Blog",
		}},
	},
	"blog-9": {
		Files: []labFile{{
			Name: "pages/home.templ",
			Lines: []string{
				`{{signal}}`,
				``,
				`templ Home(title string, path string, prefix string) {`,
				`	@Layout(title, path, prefix) {`,
				`		<h1>My Blog</h1>`,
				`		@x.Document(counter())`,
				`	}`,
				`}`,
			},
			Slots: []labSlot{
				writeSlot("signal", "the signal declaration", `var count = x.NamedNumSignal("count", 0);`,
					"the page's state — a lowercase name, created with NamedNumSignal, starting at 0; scripts reach it by that name",
				),
			},
			Result: "the page declares the count signal and the counter starts at 0",
		}},
	},
	"blog-10": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`func main() {`,
				`	if len(os.Args) > 1 && os.Args[1] == "{{flag}}" {`,
				`		devserver.Run(context.Background(), devserver.Config{`,
				`			Build: devserver.Steps(`,
				`				devserver.Command("templ", "generate"),`,
				`				devserver.Command("go", "run", ".", "--dev"),`,
				`			),`,
				`		})`,
				`		return`,
				`	}`,
				`	if err := ssg.Build(tree(), ssg.BuildOptions{Dev: dev}); err != nil {`,
				`		log.Fatal(err)`,
				`	}`,
				`}`,
			},
			Slots: []labSlot{
				pickSlot("flag", "the serve flag", []string{"pick"}, "--serve",
					"the flag that starts the dev loop — two dashes, as the overview's command shows",
					"",
					"--serve", "serve", "-serve",
				),
			},
			Result: "go run . --serve starts the dev loop instead of a one-shot build",
		}},
	},
	"blog-11": {
		Files: []labFile{{
			Name: ".github/workflows/deploy.yml",
			Lines: []string{
				`name: deploy`,
				`on:`,
				`  push:`,
				`    branches: [{{branch}}]`,
				`permissions:`,
				`  contents: read`,
				`  pages: write`,
				`  id-token: write`,
				`jobs:`,
				`  build:`,
				`    runs-on: ubuntu-latest`,
				`    steps:`,
				`      - uses: actions/checkout@v4`,
				`      - uses: actions/setup-go@v5`,
				`      - run: templ generate`,
				`      - run: go run .`,
				`      - uses: actions/configure-pages@v5`,
				`      - uses: actions/upload-pages-artifact@v3`,
				`        with:`,
				`          path: {{output}}`,
				`      - uses: actions/deploy-pages@v4`,
			},
			Slots: []labSlot{
				pickSlot("branch", "the branch the deploy runs on", []string{"render"}, "main",
					"the branch you push to — where the workflow's push trigger listens",
					`    branches: [__SLOT__]`,
					"main", "master", "dev",
				),
				pickSlot("output", "the directory to upload", []string{"render"}, "build",
					"what go run . writes — the static site directory",
					`          path: __SLOT__`,
					"build", "dist", "out",
				),
			},
			Result: "every push to main builds and deploys the build/ directory",
		}},
	},
	"clicker-1": {
		Files: []labFile{
			{
				Name: "pages/home.templ",
				Lines: []string{
					`var cookies = x.NamedNumSignal("cookies", 0);`,
					``,
					`templ oven() {`,
					`	<div class="demo">`,
					`		{{button}}`,
					`		<span class="counter-value">@x.Text(cookies)</span>`,
					`	</div>`,
					`}`,
				},
				Slots: []labSlot{
					pickSlot("button", "the clickable button", []string{"render"}, `@x.El("button", x.On("click", cookies.Inc(1))) { 🍪 }`,
						"the marker set: x.El makes the element and x.On binds the click — one component, one action",
						"__SLOT__",
						`@x.El("button", x.On("click", cookies.Inc(1))) { 🍪 }`,
						`<button @click="cookies.Inc(1)">🍪</button>`,
						`@x.On("button", "click", cookies.Inc(1))`,
					),
				},
				Result: "the button is an x.El with the click bound to cookies.Inc(1)",
			},
			{
				Name: "the counter",
				Lines: []string{
					`var cookies = x.NamedNumSignal("cookies", {{start}});`,
				},
				Slots: []labSlot{
					pickSlot("start", "the counter's start value", []string{"pick"}, "0",
						"the number the counter begins at — the value rendered before any script runs",
						"",
						"0", "1", "5",
					),
				},
				Result: "the counter starts at 0 — the static value rendered before any script runs",
			},
		},
	},
	"clicker-2": {
		Files: []labFile{{
			Name: "the script",
			Lines: []string{
				`// The escape hatch: a script reading the named signal.`,
				`setInterval(() => {`,
				`	{{signal}}.set({{signal}}.get() + 1);`,
				`}, {{interval}});`,
			},
			Slots: []labSlot{
				pickSlot("signal", "the signal the loop reads", []string{"pick"}, "cookies",
					"the name the page declared in step 1 — scripts reach signals by name",
					"",
					"cookies", "count", "level",
				),
				pickSlot("interval", "the interval in milliseconds", []string{"render"}, "1000",
					"one cookie per second — the milliseconds between ticks",
					`}, __SLOT__);`,
					"1000", "500", "100",
				),
			},
			Result: "one cookie lands per second, written through the named signal",
		}},
	},
	"clicker-3": {
		Files: []labFile{
			{
				Name: "the state",
				Lines: []string{
					`var level = x.NamedNumSignal("level", 0)`,
					`var next = x.Computed(x.Add(level, 1))`,
					`var price = x.Computed(x.Mul(next, {{mult}}))`,
				},
				Slots: []labSlot{
					pickSlot("mult", "the price multiplier", []string{"render"}, "10",
						"each level multiplies the price by this — the next upgrade costs (level + 1) × this",
						`var price = x.Computed(x.Mul(next, __SLOT__))`,
						"10", "5", "100",
					),
				},
				Result: "the next upgrade's price is (level + 1) × 10",
			},
			{
				Name: "the buy action",
				Lines: []string{
					`@x.El("button", x.On("click", cookies.Inc(x.Add(level, {{gain}})))) {`,
					`	🍪`,
					`}`,
				},
				Slots: []labSlot{
					pickSlot("gain", "the base click gain", []string{"render"}, "1",
						"every click earns level + this — the base gain before any upgrade",
						`@x.El("button", x.On("click", cookies.Inc(x.Add(level, __SLOT__)))) {`,
						"1", "5", "0",
					),
				},
				Result: "every click earns 1 + level and the price is (level + 1) × 10",
			},
		},
	},
	"clicker-4": {
		Files: []labFile{{
			Name: "the formatter",
			Lines: []string{
				`const fmt = (n) => n >= 1e6 ? (n / 1e6).toFixed(1) + "M"`,
				`	: n >= 1e3 ? (n / 1e3).toFixed(1) + "{{suffix}}"`,
				`	: String(Math.floor(n));`,
			},
			Slots: []labSlot{
				pickSlot("suffix", "the thousands suffix", []string{"render"}, "k",
					"the thousands marker — 1.2k reads better than 1234",
					`	: n >= 1e3 ? (n / 1e3).toFixed(1) + "__SLOT__"`,
					"k", "K", "m",
				),
			},
			Result: "counts above a thousand display as 1.2k",
		}},
	},
	"clicker-5": {
		Files: []labFile{{
			Name: "the script",
			Mode: "order",
			Lines: []string{
				`const KEY = "cookie-clicker-save";`,
				`cookies.set(Number(localStorage.getItem(KEY) || "0"));`,
				`setInterval(() => {`,
				`	cookies.set(cookies.get() + 1);`,
				`	localStorage.setItem(KEY, String(cookies.get()));`,
				`}, 1000);`,
			},
			Result: "the save in order: the key first, the load before the first write, then the interval that saves after every change",
		}},
	},
	"clicker-6": {
		Files: []labFile{{
			Name: "main.go",
			Lines: []string{
				`//go:build js && wasm`,
				``,
				`func main() {`,
				`	reactive.Hydrate("#clicker", func(root wasmwrap.Element) {`,
				`		cookies := reactive.Num(0.0)`,
				`		reactive.BindText(root.Find("[data-cookies]"), reactive.Computed(func() string {`,
				`			return fmtCookie(cookies.Get())`,
				`		}))`,
				`		reactive.On(root.Find("[data-bake]"), "click", func(wasmwrap.Event) {`,
				`			cookies.Inc(1)`,
				`		})`,
				`		wasmwrap.SetInterval(func() { cookies.Inc(1) }, {{interval}})`,
				`	})`,
				``,
				`	select {}`,
				`}`,
			},
			Slots: []labSlot{
				pickSlot("interval", "the auto-click interval", []string{"render"}, "1000",
					"one cookie per second — the milliseconds between ticks",
					`		wasmwrap.SetInterval(func() { cookies.Inc(1) }, __SLOT__)`,
					"1000", "500", "100",
				),
			},
			Result: "the wasm loop adds one cookie per second",
		}},
	},
	"clicker-7": {
		Files: []labFile{{
			Name: ".github/workflows/deploy.yml",
			Lines: []string{
				`name: deploy`,
				`on:`,
				`  push:`,
				`    branches: [{{branch}}]`,
				`permissions:`,
				`  contents: read`,
				`  pages: write`,
				`  id-token: write`,
				`jobs:`,
				`  build:`,
				`    runs-on: ubuntu-latest`,
				`    steps:`,
				`      - uses: actions/checkout@v4`,
				`      - uses: actions/setup-go@v5`,
				`      - run: templ generate`,
				`      - run: go run .`,
				`      - uses: actions/configure-pages@v5`,
				`      - uses: actions/upload-pages-artifact@v3`,
				`        with:`,
				`          path: {{output}}`,
				`      - uses: actions/deploy-pages@v4`,
			},
			Slots: []labSlot{
				pickSlot("branch", "the branch the deploy runs on", []string{"render"}, "main",
					"the branch you push to — where the workflow's push trigger listens",
					`    branches: [__SLOT__]`,
					"main", "master", "dev",
				),
				pickSlot("output", "the directory to upload", []string{"render"}, "build",
					"what go run . writes — the static site directory",
					`          path: __SLOT__`,
					"build", "dist", "out",
				),
			},
			Result: "every push to main builds and deploys the build/ directory",
		}},
	},
}

// slot builds a labSlot with no render template.
func slot(key, label, form string, checks []string, canonical, hint string) labSlot {
	return labSlot{Key: key, Label: label, Form: form, Check: checks, Canonical: canonical, Hint: hint}
}

// renderSlot builds a labSlot whose render check compares the value
// substituted into renderTpl against the same template rendered from the
// canonical answer — computed here, at build time, so spec and verdict can
// never disagree.
func renderSlot(key, label, form string, checks []string, canonical, hint, renderTpl string) labSlot {
	s := slot(key, label, form, checks, canonical, hint)
	s.RenderTpl = renderTpl
	return s
}

// relateSlot builds a labSlot with an additional relate check.
func relateSlot(key, label, form string, checks []string, canonical, hint, renderTpl, kind, with, relateLabel string) labSlot {
	s := renderSlot(key, label, form, checks, canonical, hint, renderTpl)
	s.Relate = &labRelate{Kind: kind, With: with, Label: relateLabel}
	return s
}

// writeSlot builds a whole-line slot: the reader writes the mechanism line
// (the placeholder occupies the whole line), checked by tokens — spacing,
// quotes and formatting never need to match exactly, the mechanism does.
func writeSlot(key, label, canonical, hint string) labSlot {
	return labSlot{Key: key, Label: label, Check: []string{"line"}, Canonical: canonical, Hint: hint}
}

// pickSlot builds a pick slot: the Options are the candidates the reader
// drops in — recognition and reasoning instead of recall. The chosen option
// renders through the slot's template line like any typed value; a pick with
// no render template checks the choice itself ("pick").
func pickSlot(key, label string, checks []string, canonical, hint, renderTpl string, options ...string) labSlot {
	s := renderSlot(key, label, "", checks, canonical, hint, renderTpl)
	s.Options = options
	return s
}

// pickRelateSlot builds a pick slot with an additional relate check.
func pickRelateSlot(key, label string, checks []string, canonical, hint, renderTpl, kind, with, relateLabel string, options ...string) labSlot {
	s := pickSlot(key, label, checks, canonical, hint, renderTpl, options...)
	s.Relate = &labRelate{Kind: kind, With: with, Label: relateLabel}
	return s
}

// labFileHTML renders one lab file for the frame: fill mode as numbered lines
// with inline slots (typed inputs or pick chips), order mode as the whole
// change scrambled into movable chips (no numbers — the order is the answer).
// The frame's expected-result strip is rendered from the same canonical
// content, so what the checker compares against is always in view.
func labFileHTML(f labFile) string {
	if f.Mode == "order" {
		return labOrderHTML(f)
	}
	var b strings.Builder
	b.WriteString(`<pre class="lab-code"><code>`)
	for i, line := range f.Lines {
		b.WriteString(`<span class="lab-line"><span class="lab-n">`)
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(`</span>`)
		b.WriteString(renderLabLine(line, f.Slots))
		b.WriteString(`</span>`)
	}
	b.WriteString(`</code></pre>`)
	return b.String()
}

// labOrderHTML renders the order-mode frame: the canonical lines in a
// deterministic (but scrambled) order, each a chip the reader moves by
// dragging the handle or with its up/down buttons — both on the same side.
// data-order-pos carries the line's canonical index — the checker compares
// positions, never text.
func labOrderHTML(f labFile) string {
	order := shuffledOrder(len(f.Lines), f.Name)
	var b strings.Builder
	b.WriteString(`<div class="lab-order-list" role="list">`)
	for _, pos := range order {
		b.WriteString(`<div class="lab-order-line" role="listitem" data-order-pos="`)
		b.WriteString(strconv.Itoa(pos))
		b.WriteString(`"><span class="lab-order-handle" data-order-handle title="Drag to reorder" aria-hidden="true">&#8801;</span><button type="button" class="lab-order-btn" data-order-up aria-label="Move up">&uarr;</button><button type="button" class="lab-order-btn" data-order-down aria-label="Move down">&darr;</button><code>`)
		b.WriteString(templ.EscapeString(f.Lines[pos]))
		b.WriteString(`</code></div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// shuffledOrder returns the indices 0..n-1 in a deterministic scrambled order
// for the given seed, never the identity (the canonical order would defeat
// the exercise).
func shuffledOrder(n int, seed string) []int {
	if n <= 1 {
		idx := make([]int, n)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	h := fnv.New32a()
	h.Write([]byte(seed))
	rng := rand.New(rand.NewSource(int64(h.Sum32())))
	idx := rng.Perm(n)
	identity := true
	for i, v := range idx {
		if v != i {
			identity = false
			break
		}
	}
	if identity {
		idx[0], idx[1] = idx[1], idx[0]
	}
	return idx
}

// renderLabLine escapes one code line, then swaps each {{key}} placeholder for
// its slot input. The swap runs in two passes over a sentinel, so a slot
// whose template mentions another slot's {{key}} can never nest markup inside
// an attribute. The line is escaped before the swap, so the input markup
// itself is never double-escaped.
func renderLabLine(line string, slots []labSlot) string {
	line = templ.EscapeString(line)
	for _, s := range slots {
		line = strings.ReplaceAll(line, "{{"+s.Key+"}}", "\x00lab-"+s.Key+"\x00")
	}
	for _, s := range slots {
		line = strings.ReplaceAll(line, "\x00lab-"+s.Key+"\x00", labInputHTML(s))
	}
	return line
}

// labInputHTML is one slot's inline widget plus its gutter mark: a text input
// for typed values, a full-width input for a whole-line write, a chip row for
// picks, and a live "what your code would make" preview wherever the slot has
// a render template.
func labInputHTML(s labSlot) string {
	if len(s.Options) > 0 {
		return labPickHTML(s)
	}
	if lineSlot(s) {
		return labLineHTML(s)
	}
	attrs := fmt.Sprintf(
		`class="lab-slot" type="text" data-lab-slot="%s" data-check="%s" data-form="%s" data-canonical="%s" data-hint="%s" aria-label="%s" placeholder="…" spellcheck="false" autocomplete="off"`,
		attr(s.Key), attr(strings.Join(s.Check, ",")), attr(s.Form), attr(s.Canonical), attr(s.Hint), attr(s.Label),
	)
	attrs += labCheckAttrs(s)
	return `<input ` + attrs + `><span class="lab-mark" data-lab-mark="` + attr(s.Key) + `" aria-hidden="true"></span>` + labLiveHTML(s)
}

// lineSlot reports whether the slot is a whole-line write.
func lineSlot(s labSlot) bool {
	for _, c := range s.Check {
		if c == "line" {
			return true
		}
	}
	return false
}

// labLineHTML is one whole-line slot: the mechanism line as a full-width
// input, token-checked against the canonical — the reader writes the line,
// not a remembered value.
func labLineHTML(s labSlot) string {
	return `<input class="lab-line-slot" type="text" data-lab-slot="` + attr(s.Key) + `" data-check="line" data-canonical="` + attr(s.Canonical) + `" data-hint="` + attr(s.Hint) + `" aria-label="` + attr(s.Label) + `" placeholder="write the line…" spellcheck="false" autocomplete="off">`
}

// labPickHTML renders one pick slot: the candidate values as chips the reader
// drops in. The chosen chip becomes the slot's value and the check chain runs
// on it like any typed value.
func labPickHTML(s labSlot) string {
	var b strings.Builder
	b.WriteString(`<span class="lab-pick" role="group" aria-label="` + attr(s.Label) + `" data-lab-slot="` + attr(s.Key) + `" data-check="` + attr(strings.Join(s.Check, ",")) + `" data-canonical="` + attr(s.Canonical) + `" data-hint="` + attr(s.Hint) + `">`)
	for _, o := range s.Options {
		b.WriteString(`<button type="button" class="lab-pick-opt" data-lab-option="` + attr(o) + `">`)
		b.WriteString(templ.EscapeString(o))
		b.WriteString(`</button>`)
	}
	b.WriteString(`</span>`)
	b.WriteString(labLiveHTML(s))
	return b.String()
}

// labCheckAttrs is the slot's check-chain attributes — the render template
// and the relate relationship — shared by the input and the pick markup.
func labCheckAttrs(s labSlot) string {
	attrs := ""
	if s.RenderTpl != "" {
		attrs += fmt.Sprintf(
			` data-render-tpl="%s" data-render-expected="%s"`,
			attr(s.RenderTpl), attr(strings.ReplaceAll(s.RenderTpl, "__SLOT__", s.Canonical)),
		)
	}
	if s.Relate != nil {
		attrs += fmt.Sprintf(
			` data-relate-kind="%s" data-relate-with="%s" data-relate-label="%s"`,
			attr(s.Relate.Kind), attr(s.Relate.With), attr(s.Relate.Label),
		)
	}
	return attrs
}

// labLiveHTML is the slot's live preview: the template line substituted as
// the reader types or picks — the frame's honest reaction, "what your code
// would make", so the lab visibly affects the page's content.
func labLiveHTML(s labSlot) string {
	if s.RenderTpl == "" {
		return ""
	}
	return `<span class="lab-live" data-lab-live="` + attr(s.Key) + `" aria-hidden="true"></span>`
}

// attr escapes one attribute value.
func attr(s string) string { return templ.EscapeString(s) }
