package main

import (
	"encoding/json"
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
		"search/index.json",
		"sitemap.xml",
		"robots.txt",
		"docs/index.html",
		"docs/how-astwerk-thinks/index.html",
		"docs/escape-hatches/index.html",
		"docs/comparison/index.html",
		"docs/getting-started/index.html",
		"docs/the-node-tree/index.html",
		"docs/ctx/index.html",
		"docs/dev-server/index.html",
		"docs/scripts-and-wasm/index.html",
		"docs/content-and-collections/index.html",
		"docs/i18n/index.html",
		"docs/quick-reference/index.html",
		"docs/starter-templates/index.html",
		"docs/how-it-works/index.html",
		"docs/common-mistakes/index.html",
		"docs/deploy/index.html",
		"docs/which-tool/index.html",
		"docs/x/index.html",
		"docs/reactive/index.html",
		"docs/wasmwrap/index.html",
		"docs/testing/index.html",
		"examples/index.html",
		"examples/templ-basics/index.html",
		"examples/tree-and-content/index.html",
		"examples/scripts/index.html",
		"style/style.css",
		"style/code.css",
		"public/favicon.svg",
		"public/og.svg",
	} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}

	// The tutorial is archived, not built.
	if _, err := os.Stat(filepath.Join(out, "tutorial", "index.html")); err == nil {
		t.Error("tutorial/ was built; it should stay in _archive/")
	}
}

// The build-time site search, sitemap and robots file must exist and agree:
// every generated page is in the index and the sitemap, and the search box's
// fetch target resolves from a page two directories deep.
func TestSearchIndexSitemapRobots(t *testing.T) {
	out := buildOnce(t)

	index := map[string]searchEntry{}
	data, err := os.ReadFile(filepath.Join(out, "search", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []searchEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("search/index.json is not valid JSON: %v", err)
	}
	for _, e := range entries {
		index[e.Path] = e
	}

	for _, path := range []string{"", "docs/getting-started/", "docs/x/", "docs/which-tool/", "examples/", "examples/templ-basics/"} {
		e, ok := index[path]
		if !ok {
			t.Errorf("search index is missing %q", path)
			continue
		}
		if e.Title == "" {
			t.Errorf("search entry %q has no title", path)
		}
	}

	sm, err := os.ReadFile(filepath.Join(out, "sitemap.xml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, loc := range []string{
		"/astwerk-website/",
		"/astwerk-website/docs/getting-started/",
		"/astwerk-website/de/docs/getting-started/",
		"/astwerk-website/examples/templ-basics/",
	} {
		if !strings.Contains(string(sm), loc) {
			t.Errorf("sitemap is missing %q", loc)
		}
	}

	// The search box on a docs page fetches the root index via a relative path:
	// from docs/x/index.html that is ../../search/index.json.
	doc, err := os.ReadFile(filepath.Join(out, "docs", "x", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), `data-index="../../search/index.json"`) {
		t.Error("search input's data-index does not resolve to the root index from a nested page")
	}
}

// The German locale is the site's own i18n demo: a BuildLocales subtree with
// the beginner path translated and everything else falling back to English
// behind a banner. Assets must stay at the site root — BuildLocales drops
// CopyFrom/CompileFrom per locale by design.
func TestGermanLocale(t *testing.T) {
	out := buildOnce(t)

	for _, want := range []string{
		"de/index.html",
		"de/docs/index.html",
		"de/docs/getting-started/index.html",
		"de/docs/how-astwerk-thinks/index.html",
		"de/docs/escape-hatches/index.html",
	} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("missing %s", want)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "de", "style", "style.css")); err == nil {
		t.Error("de/style/style.css exists — assets must stay at the site root")
	}

	banner := `noch nicht übersetzt`
	tr, _ := os.ReadFile(filepath.Join(out, "de", "docs", "getting-started", "index.html"))
	if strings.Contains(string(tr), banner) {
		t.Error("translated page shows the not-translated banner")
	}
	un, _ := os.ReadFile(filepath.Join(out, "de", "docs", "x", "index.html"))
	if !strings.Contains(string(un), banner) {
		t.Error("untranslated de page is missing the not-translated banner")
	}

	// The sidebar's active entry must work under the locale too: c.Path
	// includes the de/ segment, so the match strips it — every docs page renders
	// exactly one active sidebar link, en and de alike.
	sidebarActive := regexp.MustCompile(`class="active" aria-current="page"`)
	for _, page := range []string{
		filepath.Join("docs", "getting-started", "index.html"),
		filepath.Join("de", "docs", "getting-started", "index.html"),
		filepath.Join("de", "docs", "x", "index.html"),
	} {
		body, err := os.ReadFile(filepath.Join(out, page))
		if err != nil {
			t.Fatal(err)
		}
		if n := len(sidebarActive.FindAll(body, -1)); n != 1 {
			t.Errorf("%s: %d active sidebar entries, want exactly 1", page, n)
		}
	}

	home, _ := os.ReadFile(filepath.Join(out, "index.html"))
	for _, want := range []string{`hreflang="en"`, `hreflang="de"`, `id="lang-select"`, `<option value="de/"`, `selected>`} {
		if !strings.Contains(string(home), want) {
			t.Errorf("home page is missing %q (language dropdown or hreflang)", want)
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
