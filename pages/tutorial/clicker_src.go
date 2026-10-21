package tutorial

// Source snippets for the cookie clicker quest line (tutorial-game/
// 04-quest-lines.md). Like tutorial_src.go, these live in a plain .go file:
// templ rewrites raw strings that contain templ syntax into their generated
// Go, and a snippet is meant to show the source.

// clickerStep1Home is quest 1's whole change: a page with one button and one
// counter — the smallest thing a game can be.
const clickerStep1Home = `package pages

import x "github.com/LukasDerBaum42/astwerk-live/x"

// cookies is the whole game state. Named, so scripts can read and write it
// from the escape hatch later.
var cookies = x.NamedNumSignal("cookies", 0)

templ Home(title string, path string, prefix string) {
	@Layout(title, path, prefix) {
		<h1>Cookie Clicker</h1>
		@x.Document(oven())
	}
}

// oven is the game so far: one button, one counter.
templ oven() {
	<div class="demo">
		@x.El("button", x.On("click", cookies.Inc(1))) {
			🍪 Bake a cookie
		}
		<span class="counter-value">@x.Text(cookies)</span>
	</div>
}`

// clickerStep2Script is quest 2's whole change: a script that clicks for you.
// Timers are a program, not a marker — this is the escape hatch, reading the
// named signal the page already declared.
const clickerStep2Script = `// The escape hatch: a script reading the named signal. One cookie per
// second, forever — the game plays itself now.
setInterval(() => {
	cookies.set(cookies.get() + 1);
}, 1000);

// The script sits next to the markup it drives. View source: the counter is
// still static HTML; this script only moves the number.`

// clickerStep3State is quest 3's state: the level and the price derived from
// it — a memo chain the browser recomputes. Each combinator takes a value, so
// the intermediate (level + 1) is a memo feeding the price memo.
const clickerStep3State = `var level = x.NamedNumSignal("level", 0)
// The next level — the memo chain's intermediate step.
var next = x.Computed(x.Add(level, 1))
// The next upgrade's price: (level + 1) * 10, derived — not stored.
var price = x.Computed(x.Mul(next, 10))`

// clickerStep3Buy is quest 3's buy action: the click binding's gain is
// derived from the level, and the payment reads the price before it changes —
// the whole action set in two bindings.
const clickerStep3Buy = `// Every click earns 1 + level — the gain is derived, so a purchase is felt
// on the very next click.
@x.El("button", x.On("click", cookies.Inc(x.Add(level, 1)))) {
	🍪
}

@x.El("button",
	x.On("click", x.If(
		x.Computed(x.Gte(cookies, price)), // afford it?
		// Pay with the old price first, then level up: Dec(price) reads the
		// memo before it changes, so the wallet never dips negative.
		x.Seq(cookies.Dec(price), level.Inc(1)),
	)),
	x.Disabled(x.Computed(x.Lt(cookies, price))),
) {
	Buy an upgrade —
	@x.Text(price) 🍪
}`

// clickerStep4Script is quest 4's formatter: k/M suffixes are a real program,
// so the escape hatch writes the display signal and the marker renders it.
const clickerStep4Script = `// The display is a named string; the formatting is a program, so it is a
// script: k/M suffixes, rounding — the works.
const fmt = (n) => n >= 1e6 ? (n / 1e6).toFixed(1) + "M"
	: n >= 1e3 ? (n / 1e3).toFixed(1) + "k"
	: String(Math.floor(n));

setInterval(() => {
	cookies.set(cookies.get() + 1);
	display.set(fmt(cookies.get()));
}, 1000);

// In the markup: <span>@x.Text(display)</span> — the marker renders the
// current value and subscribes, exactly like the counter.`

// clickerStep4Concat is the no-script half of the story: string composition
// through x.Concat — a name signal joined with a literal.
const clickerStep4Concat = `// The simple cases need no script — x.Concat composes strings:
// <p class="demo-note">@x.Text(x.Computed(x.Concat(x.Lit("Baked: "), name)))</p>

// Combinators stop where a real program begins: rounding and suffixes are
// that boundary, and the script above is on the other side of it.`

// clickerStep5Script is quest 5's persistence: localStorage through the
// escape hatch — load on boot, save on every change.
const clickerStep5Script = `// localStorage through the escape hatch: load on boot, save on every
// change. Reload the page — the game remembers.
const KEY = "cookie-clicker-save";
cookies.set(Number(localStorage.getItem(KEY) || "0"));

setInterval(() => {
	cookies.set(cookies.get() + 1);
	localStorage.setItem(KEY, String(cookies.get()));
}, 1000);`

// clickerStep6Main is quest 6's whole change: the same game on the WASM
// path — reactive state, a wasmwrap timer, the same markup contract.
const clickerStep6Main = `//go:build js && wasm

package main

import (
	"fmt"

	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func main() {
	reactive.Hydrate("#clicker", func(root wasmwrap.Element) {
		cookies := reactive.Num(0.0)

		reactive.BindText(root.Find("[data-cookies]"), reactive.Computed(func() string {
			return fmtCookie(cookies.Get())
		}))
		reactive.On(root.Find("[data-bake]"), "click", func(wasmwrap.Event) {
			cookies.Inc(1)
		})

		// The same timer the x line kept as a script — here it is a
		// wasmwrap.SetInterval in Go.
		wasmwrap.SetInterval(func() { cookies.Inc(1) }, 1000)
	})

	select {}
}

// fmtCookie formats the counter — Go's strconv, the whole stdlib on tap.
func fmtCookie(n float64) string {
	if n >= 1e6 {
		return fmt.Sprintf("%.1fM", n/1e6)
	}
	if n >= 1e3 {
		return fmt.Sprintf("%.1fk", n/1e3)
	}
	return fmt.Sprintf("%.0f", n)
}`

// clickerStep7Workflow is quest 7's whole change: the deploy workflow — the
// blog line's last step, made concrete. The build directory is the site; this
// is how it gets from the repo to the pages.
const clickerStep7Workflow = `# .github/workflows/deploy.yml — the build directory is the site.
name: deploy
on:
  push:
    branches: [main]
permissions:
  contents: read
  pages: write
  id-token: write
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
      - run: go install github.com/a-h/templ/cmd/templ@latest
      - run: templ generate
      - run: go run .
      - uses: actions/configure-pages@v5
      - uses: actions/upload-pages-artifact@v3
        with:
          path: build
      - uses: actions/deploy-pages@v4`
