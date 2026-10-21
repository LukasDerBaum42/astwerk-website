package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LukasDerBaum42/astwerk-website/scripts/game"
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
		"docs/which-tool/index.html",
		"docs/common-mistakes/index.html",
		"docs/deploy/index.html",
		"docs/getting-started/index.html",
		"docs/the-node-tree/index.html",
		"docs/x/index.html",
		"tutorial/index.html",
		"tutorial/1-setup/index.html",
		"tutorial/2-the-tree/index.html",
		"tutorial/3-a-layout/index.html",
		"tutorial/4-markdown-posts/index.html",
		"tutorial/5-a-collection/index.html",
		"tutorial/6-tags-and-pagination/index.html",
		"tutorial/7-rss-and-sitemap/index.html",
		"tutorial/8-a-locale/index.html",
		"tutorial/9-interactivity/index.html",
		"tutorial/10-the-dev-loop/index.html",
		"tutorial/11-ship-it/index.html",
		"tutorial/clicker/1-the-button/index.html",
		"tutorial/clicker/2-the-clicker/index.html",
		"tutorial/clicker/3-the-upgrades/index.html",
		"tutorial/clicker/4-the-numbers/index.html",
		"tutorial/clicker/5-the-save/index.html",
		"tutorial/clicker/6-the-twin/index.html",
		"tutorial/clicker/7-the-launch/index.html",
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
		"examples/x-faq/index.html",
		"examples/stopwatch/index.html",
		"examples/filter/index.html",
		"examples/router-demo/index.html",
		"examples/bridge/index.html",
		"examples/router/index.html",
		"examples/data/posts.json",
		"style/style.css",
		"style/code.css",
		"public/favicon.svg",
		"public/og.svg",
		"scripts/demo.wasm",
		"scripts/router_demo.wasm",
		"scripts/bridge_demo.wasm",
		"scripts/clicker_demo.wasm",
		"scripts/tutorial_game.wasm",
		"scripts/wasm_exec.js",
	} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("missing %s", want)
		}
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

	for _, path := range []string{"", "docs/x/", "docs/which-tool/", "tutorial/", "tutorial/1-setup/", "examples/", "examples/x-counter/"} {
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
		"/astwerk-website/tutorial/1-setup/",
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
		"de/tutorial/index.html",
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
	// includes the de/ segment, so the match strips it (07-1d) — every docs
	// page renders exactly one active sidebar link, en and de alike.
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
	tutorial, _ := os.ReadFile(filepath.Join(out, "de", "tutorial", "3-a-layout", "index.html"))
	// Tutorial pages swap the docs sidebar for the quest hub: the current quest
	// is the hub chain's aria-current="page" row — exactly one, de and en alike.
	if n := len(regexp.MustCompile(`aria-current="page"`).FindAll(tutorial, -1)); n != 1 {
		t.Errorf("de tutorial page: %d aria-current rows, want exactly 1 (the hub's current quest)", n)
	}
	if !strings.Contains(string(tutorial), "data-quest-hub") {
		t.Error("de tutorial page is missing the quest hub")
	}
	if strings.Contains(string(tutorial), `<aside class="sidebar"`) {
		t.Error("tutorial pages must not render the docs sidebar — the quest hub replaces it")
	}

	home, _ := os.ReadFile(filepath.Join(out, "index.html"))
	for _, want := range []string{`hreflang="en"`, `hreflang="de"`, `id="lang-select"`, `<option value="de/"`, `selected>`} {
		if !strings.Contains(string(home), want) {
			t.Errorf("home page is missing %q (language dropdown or hreflang)", want)
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
		"examples/x-faq/index.html",
		"examples/x-under-the-hood/index.html",
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
		"forms":     {`id="form-demo"`, "data-price", "data-qty", "data-total", "data-tax", "data-taxed"},
		"bindings":  {`id="bindings-demo"`, "data-bind-href", "data-bind-class", "data-bind-disabled", "data-bind-style", "data-toggle-class", "data-toggle-disabled", "data-recolour", "data-retarget", "data-bind-status"},
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

// The tutorial game shell (tutorial-game/): every step page carries the HUD
// skeleton with its quest's data — the shell hydrates it — and the overview
// carries the level map. The quiz event hook ships in the shared layout.
func TestTutorialGameShell(t *testing.T) {
	out := buildOnce(t)

	// blogSteps maps each blog quest to its page's expected data; the clicker
	// line gets the same treatment below. slots is the quest's step-lab slot
	// count (06-step-lab.md's table, with 6 and 8 still living in the line).
	blogSteps := map[int]struct {
		objectives int
		quizzes    int
		slots      int
	}{
		1: {3, 2, 1}, 2: {3, 2, 2}, 3: {3, 2, 1}, 4: {2, 2, 2}, 5: {2, 2, 2},
		6: {2, 2, 2}, 7: {2, 2, 2}, 8: {2, 2, 2}, 9: {1, 3, 1}, 10: {1, 2, 1},
		11: {2, 2, 2},
	}
	for step, spec := range blogSteps {
		body, err := os.ReadFile(filepath.Join(out, "tutorial", fmt.Sprintf("%d-%s", step, stepSlug(step)), "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		html := string(body)
		assertQuestPage(t, html, fmt.Sprintf("blog-%d", step), spec.objectives, spec.quizzes, spec.slots)
	}

	for step, spec := range map[int]struct {
		objectives int
		quizzes    int
		slots      int
	}{
		1: {2, 2, 2}, 2: {2, 2, 2}, 3: {2, 2, 2}, 4: {2, 2, 1}, 5: {2, 2, 1},
		6: {3, 2, 1}, 7: {2, 2, 2},
	} {
		body, err := os.ReadFile(filepath.Join(out, "tutorial", "clicker", fmt.Sprintf("%d-%s", step, clickerSlug(step)), "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		html := string(body)
		assertQuestPage(t, html, fmt.Sprintf("clicker-%d", step), spec.objectives, spec.quizzes, spec.slots)
	}

	overview, err := os.ReadFile(filepath.Join(out, "tutorial", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`data-game-map`, `data-map-node="blog-1"`, `data-map-node="blog-11"`,
		`data-map-node="clicker-1"`, `data-map-node="clicker-7"`, `data-map-path`,
		`data-map-dep="blog-4"`, `game-map-lock`,
		`data-game-line-tab="blog"`, `data-game-line-tab="clicker"`, `data-game-summary`,
		`data-badge="first-quest"`, `data-badge="completionist"`, `data-badge="cookie-monster"`,
		`data-quest-hub`, `data-hub-line-tab="blog"`, `data-hub-chain-line="clicker"`,
		`data-hub-row="blog-1"`, `data-hub-row="clicker-7"`,
		`data-map-toggle`, `data-hub-toggle`,
		`data-stage="go"`, `data-stage="static"`, `data-boss-callout`,
	} {
		if !strings.Contains(string(overview), want) {
			t.Errorf("tutorial overview is missing %q (the level map + hub)", want)
		}
	}
	if strings.Contains(string(overview), `data-header-state`) {
		t.Error("the overview is not a quest page and must not carry the header state chip")
	}

	home, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(home), `astwerk:quiz`) {
		t.Error("the quiz script's game event hook is missing from the shared layout")
	}
	if !strings.Contains(string(home), `astwerk:lab`) {
		t.Error("the lab script's game event hook is missing from the shared layout")
	}

	// The emitted totals must unmarshal into the ledger's Totals — the badge
	// rules, the map's per-line progress and the dependency graph read them,
	// so a shape drift would silently disable every badge.
	raw, err := os.ReadFile(filepath.Join(out, "tutorial", "3-a-layout", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`data-game-totals="([^"]*)"`).FindSubmatch(raw)
	if m == nil {
		t.Fatal("no data-game-totals on the step page")
	}
	var tot game.Totals
	if err := json.Unmarshal([]byte(html.UnescapeString(string(m[1]))), &tot); err != nil {
		t.Fatalf("data-game-totals is not the ledger's shape: %v", err)
	}
	if tot.Lines["blog"].Quests != 11 || tot.Lines["clicker"].Quests != 7 {
		t.Errorf("data-game-totals round-trip: blog=%d clicker=%d quests, want 11 and 7",
			tot.Lines["blog"].Quests, tot.Lines["clicker"].Quests)
	}
	if len(tot.Lines["blog"].Ids) != 11 || len(tot.Lines["clicker"].Ids) != 7 {
		t.Errorf("data-game-totals round-trip: blog ids=%d clicker ids=%d, want 11 and 7",
			len(tot.Lines["blog"].Ids), len(tot.Lines["clicker"].Ids))
	}
	if len(tot.Deps["blog-11"]) != 3 {
		t.Errorf("data-game-totals round-trip: blog-11 requires %v, want 3 entries", tot.Deps["blog-11"])
	}
	if tot.Names["blog-3"] != "3 · The Shell" {
		t.Errorf("data-game-totals round-trip: blog-3 name = %q, want \"3 · The Shell\"", tot.Names["blog-3"])
	}
	if tot.Quests["blog-3"].Objectives != 3 || tot.Quests["blog-3"].Quizzes != 2 || tot.Quests["blog-3"].Lab != 1 {
		t.Errorf("data-game-totals round-trip: blog-3 meta = %+v, want 3 objectives / 2 quizzes / 1 lab", tot.Quests["blog-3"])
	}
	if tot.Quests["clicker-6"].Objectives != 3 || tot.Quests["clicker-1"].Lab != 2 {
		t.Errorf("data-game-totals round-trip: clicker meta wrong: %+v %+v", tot.Quests["clicker-1"], tot.Quests["clicker-6"])
	}

	// The paths map the shell needs for the modal's onward links round-trips too.
	raw, err = os.ReadFile(filepath.Join(out, "tutorial", "3-a-layout", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	pm := regexp.MustCompile(`data-game-paths="([^"]*)"`).FindSubmatch(raw)
	if pm == nil {
		t.Fatal("no data-game-paths on the step page")
	}
	var paths map[string]string
	if err := json.Unmarshal([]byte(html.UnescapeString(string(pm[1]))), &paths); err != nil {
		t.Fatalf("data-game-paths is not a string map: %v", err)
	}
	if paths["blog-4"] == "" || paths["tutorial"] == "" {
		t.Errorf("data-game-paths missing entries: %v", paths)
	}
}

// assertQuestPage checks one quest page's game contract: the HUD's data, the
// objective checkboxes (in the step body) and their hub mirrors, the quiz
// count — which must match the questSpecs entry the HUD data came from — and
// the step-lab frame's slot count, so the totals the badges rely on can never
// drift from the pages (07-5d).
func assertQuestPage(t *testing.T, html, id string, objectives, quizzes, slots int) {
	t.Helper()
	if !strings.Contains(html, fmt.Sprintf(`data-quest="%s"`, id)) {
		t.Errorf("%s: page has no data-quest=%q", id, id)
	}
	if !strings.Contains(html, fmt.Sprintf(`data-objectives="%d"`, objectives)) {
		t.Errorf("%s: HUD data-objectives != %d", id, objectives)
	}
	if !strings.Contains(html, fmt.Sprintf(`data-quizzes="%d"`, quizzes)) {
		t.Errorf("%s: HUD data-quizzes != %d", id, quizzes)
	}
	if n := len(objectiveRe.FindAllString(html, -1)); n != objectives {
		t.Errorf("%s: page emits %d objective checkboxes, want %d", id, n, objectives)
	}
	if n := len(mirrorRe.FindAllString(html, -1)); n != objectives*2 {
		t.Errorf("%s: quest hub mirrors + section toggles = %d, want %d (one hub mirror and one section toggle per objective)", id, n, objectives*2)
	}
	if n := len(sectionRe.FindAllString(html, -1)); n != objectives {
		t.Errorf("%s: page emits %d content sections, want %d (one per objective)", id, n, objectives)
	}
	if n := strings.Count(html, `class="hub-row-meta"`); n != 18 {
		t.Errorf("%s: hub chain emits %d row-meta spans, want 18 (both lines' rows)", id, n)
	}
	if n := strings.Count(html, `class="quiz"`); n != quizzes {
		t.Errorf("%s: page has %d quizzes, questSpecs says %d", id, n, quizzes)
	}
	if !strings.Contains(html, "data-game-modal") || !strings.Contains(html, "data-game-toasts") {
		t.Errorf("%s: page is missing the modal or toast containers", id)
	}
	if !strings.Contains(html, `data-header-state`) {
		t.Errorf("%s: page is missing the header state chip (08-reactive-pages)", id)
	}
	if !strings.Contains(html, `data-quest-hub`) {
		t.Errorf("%s: page is missing the quest hub", id)
	}
	if !strings.Contains(html, `data-hub-row="blog-1"`) || !strings.Contains(html, `data-hub-row="clicker-1"`) {
		t.Errorf("%s: the hub's chain must carry both lines' rows", id)
	}
	if !strings.Contains(html, `data-rail`) != (objectives < 2) {
		t.Errorf("%s: step rail present=%v, want it only on quests with >= 2 objectives",
			id, strings.Contains(html, `data-rail`))
	}
	if !strings.Contains(html, fmt.Sprintf(`data-lab-slots="%d"`, slots)) {
		t.Errorf("%s: lab frame data-lab-slots != %d", id, slots)
	}
	// A slot may appear twice in its file (the signal name is both the
	// variable and the string) — count the unique fill, line and pick keys,
	// plus the order-mode files (each one practice unit), against the
	// ledger's cap.
	unique := map[string]bool{}
	for _, m := range labSlotRe.FindAllStringSubmatch(html, -1) {
		unique[m[1]] = true
	}
	for _, m := range labLineRe.FindAllStringSubmatch(html, -1) {
		unique[m[1]] = true
	}
	for _, m := range labPickRe.FindAllStringSubmatch(html, -1) {
		unique[m[1]] = true
	}
	n := len(unique) + len(labOrderRe.FindAllString(html, -1))
	if n != slots {
		t.Errorf("%s: lab frame emits %d unique fill+line+pick keys + %d order files, want %d total",
			id, len(unique), len(labOrderRe.FindAllString(html, -1)), slots)
	}
	if !strings.Contains(html, `data-complete-panel`) {
		t.Errorf("%s: page is missing the completed-quest footer panel", id)
	}
}

// objectiveRe matches the objective checkboxes in the step body — the exact
// input markup, so the selectors inside the embedded scripts can't pollute
// the count (and neither can the hub's mirrors). The boundary keeps
// data-objective-mirror out.
var objectiveRe = regexp.MustCompile(`<input type="checkbox" data-objective[\s>]`)

// mirrorRe matches the quest hub's templ-rendered mirrors, which carry the
// objective's index as data-mirror-of.
var mirrorRe = regexp.MustCompile(`<input type="checkbox" data-objective-mirror data-mirror-of="\d+"[\s>]`)

// sectionRe matches the content sections the pages are split into — one per
// objective, each the rail's scroll target. Anchored on the section tag so the
// scripts' querySelector strings can't pollute the count.
var sectionRe = regexp.MustCompile(`<section class="quest-section"`)

// labSlotRe extracts each fill slot's key from its input's data-lab-slot —
// anchored on the input's class so the shared lab script's querySelector
// strings can't pollute the count.
var labSlotRe = regexp.MustCompile(`class="lab-slot"[^>]*data-lab-slot="([^"]+)"`)

// labLineRe extracts each whole-line slot's key — the mechanism lines the
// reader writes in full.
var labLineRe = regexp.MustCompile(`class="lab-line-slot"[^>]*data-lab-slot="([^"]+)"`)

// labPickRe extracts each pick slot's key from its container — the pick
// chips' buttons carry data-lab-option, not data-lab-slot, so the container's
// data-lab-slot is the anchor.
var labPickRe = regexp.MustCompile(`class="lab-pick"[^>]*data-lab-slot="([^"]+)"`)

// labOrderRe matches an order-mode lab file's frame — anchored on the file
// div so the shared lab script's selector strings can't pollute the count.
var labOrderRe = regexp.MustCompile(`<div class="lab-file"[^>]*data-lab-order-file`)

// stepSlug maps a blog step number to its tutorial path segment, mirroring
// questLines' slugs.
func stepSlug(step int) string {
	slugs := map[int]string{
		1: "setup", 2: "the-tree", 3: "a-layout", 4: "markdown-posts",
		5: "a-collection", 6: "tags-and-pagination", 7: "rss-and-sitemap",
		8: "a-locale", 9: "interactivity", 10: "the-dev-loop", 11: "ship-it",
	}
	return slugs[step]
}

// clickerSlug maps a clicker step number to its tutorial path segment.
func clickerSlug(step int) string {
	slugs := map[int]string{
		1: "the-button", 2: "the-clicker", 3: "the-upgrades", 4: "the-numbers",
		5: "the-save", 6: "the-twin", 7: "the-launch",
	}
	return slugs[step]
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
