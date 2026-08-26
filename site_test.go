package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// buildOnce generates the site into a temp directory and returns its path.
func buildOnce(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "build")
	if _, err := buildSite(out, "/astwerk-website", false); err != nil {
		t.Fatal(err)
	}
	return out
}

func htmlFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".html") {
			out = append(out, p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

var urlAttr = regexp.MustCompile(`(?:href|src)="([^"]+)"`)

// Every internal link must resolve to a real file from the directory of the page
// that emitted it. This is the property that makes one build work at a localhost
// root, under /astwerk-website/, and opened straight off disk.
func TestInternalLinksResolveFromEachPage(t *testing.T) {
	out := buildOnce(t)
	pages := htmlFiles(t, out)
	if len(pages) < 10 {
		t.Fatalf("only %d pages built; the site is probably not being generated", len(pages))
	}

	for _, page := range pages {
		body, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Dir(page)
		rel, _ := filepath.Rel(out, page)

		for _, m := range urlAttr.FindAllStringSubmatch(string(body), -1) {
			u := m[1]
			if strings.HasPrefix(u, "http") || strings.HasPrefix(u, "#") ||
				strings.HasPrefix(u, "mailto:") || strings.HasPrefix(u, "data:") {
				continue
			}
			if strings.HasPrefix(u, "/") {
				t.Errorf("%s: %q is absolute — use Ctx.Asset or Ctx.Link", rel, u)
				continue
			}

			// A path can carry its own "#anchor" (a sidebar subsection link,
			// say); only the path half needs to resolve to a file.
			path, _, _ := strings.Cut(u, "#")
			if path == "" {
				continue
			}

			target := filepath.Clean(filepath.Join(dir, path))
			if !strings.HasPrefix(target, out) {
				t.Errorf("%s: %q escapes the output directory", rel, u)
				continue
			}
			info, errFile := os.Stat(target)
			indexPath := filepath.Join(target, "index.html")
			_, errIndex := os.Stat(indexPath)
			if errFile != nil && errIndex != nil {
				t.Errorf("%s: %q does not resolve to anything", rel, u)
				continue
			}

			if _, anchor, ok := strings.Cut(u, "#"); ok && anchor != "" {
				dest := target
				if errFile != nil || info.IsDir() {
					dest = indexPath
				}
				page, err := os.ReadFile(dest)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(page), `id="`+anchor+`"`) {
					t.Errorf("%s: %q points at an anchor that doesn't exist in the target page", rel, u)
				}
			}
		}
	}
}

// Canonical URLs are the one place an absolute URL is correct, and they must
// carry the base path or they point at the wrong site.
func TestCanonicalURLsAreAbsoluteAndBased(t *testing.T) {
	out := buildOnce(t)
	canonical := regexp.MustCompile(`<link rel="canonical" href="([^"]+)"`)

	for _, page := range htmlFiles(t, out) {
		body, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(out, page)
		m := canonical.FindSubmatch(body)
		if m == nil {
			t.Errorf("%s: no canonical link", rel)
			continue
		}
		got := string(m[1])
		if !strings.HasPrefix(got, "https://lukasderbaum42.github.io/astwerk-website/") {
			t.Errorf("%s: canonical %q is missing the origin or base path", rel, got)
		}
	}
}

func TestExpectedPagesExist(t *testing.T) {
	out := buildOnce(t)
	for _, want := range []string{
		"index.html",
		"404.html",
		".nojekyll",
		"docs/index.html",
		"docs/getting-started/index.html",
		"docs/the-node-tree/index.html",
		"docs/x/index.html",
		"tutorial/index.html",
		"tutorial/1-setup/index.html",
		"tutorial/2-the-tree/index.html",
		"tutorial/3-a-layout/index.html",
		"tutorial/4-more-pages/index.html",
		"tutorial/5-a-collection/index.html",
		"tutorial/6-styles/index.html",
		"tutorial/7-a-locale/index.html",
		"tutorial/8-interactivity/index.html",
		"tutorial/9-the-dev-loop/index.html",
		"tutorial/10-ship-it/index.html",
		"examples/index.html",
		"examples/x-counter/index.html",
		"examples/x-forms/index.html",
		"examples/x-bindings/index.html",
		"examples/x-todo/index.html",
		"examples/x-under-the-hood/index.html",
		"examples/counter/index.html",
		"examples/list/index.html",
		"examples/forms/index.html",
		"examples/bindings/index.html",
		"examples/async/index.html",
		"examples/canvas/index.html",
		"examples/templ-basics/index.html",
		"examples/tree-and-content/index.html",
		"examples/scripts-and-routing/index.html",
		"examples/wasmwrap-and-templ/index.html",
		"examples/x-async/index.html",
		"examples/x-canvas/index.html",
		"examples/x-faq/index.html",
		"examples/stopwatch/index.html",
		"examples/filter/index.html",
		"style/style.css",
		"style/code.css",
		"public/favicon.svg",
		"scripts/demo.wasm",
		"scripts/wasm_exec.js",
	} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}
}

// The compiled-markup example pages are self-contained: their JavaScript is
// generated at build time and embedded in the page, and they must not load the
// WASM demo module. Each must carry the runtime (aw), at least one stamped
// node, and at least one binding — that is the contract that makes them work
// with JavaScript on and read fine with it off.
func TestXExamplesAreSelfContained(t *testing.T) {
	out := buildOnce(t)

	for _, page := range []string{
		"examples/x-counter/index.html",
		"examples/x-forms/index.html",
		"examples/x-bindings/index.html",
		"examples/x-todo/index.html",
		"examples/x-async/index.html",
		"examples/x-canvas/index.html",
		"examples/x-faq/index.html",
		"examples/x-under-the-hood/index.html",
		"tutorial/8-interactivity/index.html",
	} {
		body, err := os.ReadFile(filepath.Join(out, page))
		if err != nil {
			t.Fatal(err)
		}
		html := string(body)

		for _, want := range []string{
			"globalThis.aw =", // the generated runtime
			`data-x="0"`,      // the first stamped node
			"aw.",             // at least one binding
		} {
			if !strings.Contains(html, want) {
				t.Errorf("%s is missing %q — the compiled markup did not generate", page, want)
			}
		}
		if strings.Contains(html, "wasm_exec.js") || strings.Contains(html, "data-wasm=") {
			t.Errorf("%s loads the WASM module; it should be self-contained", page)
		}
	}
}

// Each live example page is hydrated by scripts/demo.go, which finds its
// markup by selector. If a data attribute is renamed in one place and not the
// other the demo silently does nothing, so pin the contract per page.
func TestExamplesMarkupMatchesDemoSelectors(t *testing.T) {
	out := buildOnce(t)

	for page, selectors := range map[string][]string{
		"counter":   {`id="counter"`, "data-count", "data-inc", "data-dec", "data-reset", "data-doubled"},
		"list":      {`id="todo"`, "data-form", "data-new", "data-list", "data-remaining"},
		"forms":     {`id="form-demo"`, "data-price", "data-qty", "data-total", "data-tax", "data-tax-note"},
		"bindings":  {`id="bindings-demo"`, "data-bind-href", "data-bind-class", "data-bind-disabled", "data-bind-style", "data-toggle-class", "data-toggle-disabled", "data-recolour", "data-retarget"},
		"async":     {`id="async-demo"`, "data-reload", "data-status", "data-posts"},
		"canvas":    {`id="canvas-demo"`, "data-canvas", "data-toggle"},
		"stopwatch": {`id="stopwatch-demo"`, "data-time", "data-start", "data-reset"},
		"filter":    {`id="filter-demo"`, "data-query", "data-count", "data-results"},
	} {
		body, err := os.ReadFile(filepath.Join(out, "examples", page, "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		html := string(body)

		for _, sel := range selectors {
			if !strings.Contains(html, sel) {
				t.Errorf("examples/%s is missing %s, which scripts/demo.go hydrates", page, sel)
			}
		}
		if !strings.Contains(html, `data-wasm=`) {
			t.Errorf("examples/%s doesn't load the demo WASM module", page)
		}
	}
}

// Syntax highlighting runs at build time; no JS highlighter should be shipped.
func TestCodeIsHighlightedAtBuildTime(t *testing.T) {
	out := buildOnce(t)
	body, err := os.ReadFile(filepath.Join(out, "docs", "the-node-tree", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)

	if !strings.Contains(html, `class="chroma"`) {
		t.Error("no chroma output — code blocks are not highlighted")
	}
	if n := strings.Count(html, `class="chroma"`); n < 5 {
		t.Errorf("only %d highlighted blocks on the Node tree page", n)
	}
}

// Every docs page needs a Title and Description: the sidebar, index and pager
// are all built from them.
func TestDocsFrontMatterIsComplete(t *testing.T) {
	docs, err := loadDocs("content/docs")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) < 8 {
		t.Fatalf("only %d docs pages loaded", len(docs))
	}

	seen := map[int]string{}
	for _, d := range docs {
		if d.FM.Description == "" {
			t.Errorf("%s.md has no Description", d.Slug)
		}
		if d.FM.Order == 0 {
			t.Errorf("%s.md has no Order, so its sidebar position is arbitrary", d.Slug)
		}
		if other, dup := seen[d.FM.Order]; dup {
			t.Errorf("%s.md and %s.md share Order %d", d.Slug, other, d.FM.Order)
		}
		seen[d.FM.Order] = d.Slug
	}
}

// loadDocs must return pages in Order, since the sidebar and pager rely on it.
func TestDocsAreOrdered(t *testing.T) {
	docs, err := loadDocs("content/docs")
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(docs); i++ {
		if docs[i].FM.Order < docs[i-1].FM.Order {
			t.Fatalf("docs out of order: %s (%d) after %s (%d)",
				docs[i].Slug, docs[i].FM.Order, docs[i-1].Slug, docs[i-1].FM.Order)
		}
	}
}

// The site's theme is a manual [data-theme] toggle (dark by default), not
// prefers-color-scheme — so code.css has to switch on the same signal or a
// visitor who toggles to light against a dark OS preference gets dark code
// blocks on a light page. Every light rule must be scoped to
// [data-theme="light"], and the unscoped default (what renders with no
// attribute at all) must be the dark palette.
func TestCodeCSSIsScopedByThemeAttribute(t *testing.T) {
	css, err := os.ReadFile(filepath.Join("style", "code.css"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(css)

	if strings.Contains(body, "@media") {
		t.Error("code.css still has a media query; the site's theme is a manual [data-theme] toggle, not prefers-color-scheme")
	}
	if !strings.Contains(body, `[data-theme="light"] .chroma {`) {
		t.Error(`code.css has no [data-theme="light"] .chroma rule — light mode would show dark code blocks`)
	}

	// The first .chroma rule (there are many, one per token class) is the
	// unscoped default. It must not itself be scoped to a data-theme
	// attribute, since the whole point is that dark renders with none.
	i := strings.Index(body, ".chroma {")
	if i < 0 {
		t.Fatal("no .chroma rule in code.css")
	}
	line := body[:i]
	if start := strings.LastIndex(line, "\n"); start >= 0 {
		line = line[start+1:]
	}
	if strings.Contains(line, "data-theme") {
		t.Error("the first .chroma rule is scoped to a data-theme attribute; the unscoped default should be dark")
	}
}

// The default (unscoped) block is dark and must set a base colour on
// .chroma, since unstyled tokens inherit it rather than having a rule of
// their own.
func TestDarkCodeHasReadableBaseColor(t *testing.T) {
	css, err := os.ReadFile(filepath.Join("style", "code.css"))
	if err != nil {
		t.Fatal(err)
	}

	m := regexp.MustCompile(`(?m)^\s*(?:/\*.*?\*/\s*)?\.chroma \{([^}]*)\}`).FindStringSubmatch(string(css))
	if m == nil {
		t.Fatal("no unscoped default .chroma rule")
	}

	// Split into declarations rather than regexping for "color:", which also
	// matches inside "background-color:".
	var base string
	for _, decl := range strings.Split(m[1], ";") {
		prop, value, ok := strings.Cut(decl, ":")
		if ok && strings.TrimSpace(prop) == "color" {
			base = strings.TrimSpace(value)
		}
	}
	if base == "" {
		t.Fatal("dark .chroma sets no base colour; unstyled tokens would inherit the page colour")
	}

	var r, g, b int
	if _, err := fmt.Sscanf(base, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("unparseable base colour %q: %v", base, err)
	}
	if r+g+b < 3*128 {
		t.Errorf("dark base code colour %s is dark; unstyled tokens like .nx would "+
			"be unreadable on the dark background", base)
	}
}
