//go:build js && wasm

// Command demo powers the live examples on the Examples page.
//
// It hydrates markup the site already generated rather than building any DOM
// from scratch, which is the pattern the docs recommend: one source of markup,
// surgical updates, and a page that still reads with WebAssembly disabled.
//
// Every demo here has a compiled-markup twin on the x side of the Examples
// page — same behaviour, different implementation — so the two can be compared
// feature for feature.
package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LukasDerBaum42/astwerk/reactive"
	"github.com/LukasDerBaum42/astwerk/wasmwrap"
)

func main() {
	counter()
	todoList()
	formDemo()
	bindingsDemo()
	asyncDemo()
	canvasDemo()
	stopwatchDemo()
	filterDemo()

	// Keep the module alive: main returning would tear down every handler.
	select {}
}

func counter() {
	count := reactive.NewSignal(0)
	doubled := reactive.Computed(func() int { return count.Get() * 2 })

	reactive.Hydrate("#counter", func(root wasmwrap.Element) {
		reactive.BindText(root.Find("[data-count]"), func() string {
			return strconv.Itoa(count.Get())
		})
		reactive.BindText(root.Find("[data-doubled]"), func() string {
			return strconv.Itoa(doubled.Get())
		})

		reactive.On(root.Find("[data-inc]"), "click", func(wasmwrap.Event) {
			count.Update(func(v int) int { return v + 1 })
		})
		reactive.On(root.Find("[data-dec]"), "click", func(wasmwrap.Event) {
			count.Update(func(v int) int { return v - 1 })
		})
		reactive.On(root.Find("[data-reset]"), "click", func(wasmwrap.Event) {
			count.Set(0)
		})
	})
}

type item struct {
	ID   string
	Text string
}

func todoList() {
	items := reactive.NewSignal([]item{
		{ID: "1", Text: "Read the docs"},
		{ID: "2", Text: "Build something"},
	})

	reactive.Hydrate("#todo", func(root wasmwrap.Element) {
		field := root.Find("[data-new]")
		draft := reactive.NewSignal("")
		reactive.BindValue(field, draft)

		// Submit rather than click, so the Enter key works too.
		reactive.On(root.Find("[data-form]"), "submit", func(e wasmwrap.Event) {
			e.PreventDefault()
			text := strings.TrimSpace(draft.Get())
			if text == "" {
				return
			}
			reactive.Batch(func() {
				items.Update(func(all []item) []item {
					return append(all, item{ID: fmt.Sprintf("%d", len(all)+1), Text: text})
				})
				draft.Set("")
			})
		})

		reactive.BindText(root.Find("[data-remaining]"), func() string {
			return strconv.Itoa(len(items.Get()))
		})

		reactive.BindList(root.Find("[data-list]"), items.Get,
			func(it item) string { return it.ID },
			func(it item) wasmwrap.Element {
				row := wasmwrap.Create("li")
				remove := wasmwrap.Create("button").SetText("remove")
				remove.Class().Add("remove")

				reactive.On(remove, "click", func(wasmwrap.Event) {
					items.Update(func(all []item) []item {
						return slices.DeleteFunc(slices.Clone(all), func(o item) bool {
							return o.ID == it.ID
						})
					})
				})
				return row.Append(wasmwrap.Create("span").SetText(it.Text), remove)
			})
	})
}

// formDemo two-way binds two number fields, derives a total from them, and
// lets a checkbox reveal a panel — the same features as the x forms demo.
func formDemo() {
	reactive.Hydrate("#form-demo", func(root wasmwrap.Element) {
		price := reactive.NewSignal(10.0)
		qty := reactive.NewSignal(2.0)
		total := reactive.Computed(func() float64 { return price.Get() * qty.Get() })
		tax := reactive.NewSignal(false)

		reactive.BindNumber(root.Find("[data-price]"), price)
		reactive.BindNumber(root.Find("[data-qty]"), qty)
		reactive.BindText(root.Find("[data-total]"), func() string {
			return fmt.Sprintf("%.2f", total.Get())
		})
		reactive.BindChecked(root.Find("[data-tax]"), tax)

		// BindShow, not BindWhen: the note keeps no state, so toggling
		// visibility is all this needs.
		reactive.BindShow(root.Find("[data-tax-note]"), tax.Get)
	})
}

// bindingsDemo drives four one-property bindings — href, class, disabled,
// style — from four signals, each owned by a control button. The WASM twin
// of the x attribute-bindings demo.
func bindingsDemo() {
	reactive.Hydrate("#bindings-demo", func(root wasmwrap.Element) {
		url := reactive.NewSignal("https://templ.guide")
		active := reactive.NewSignal(true)
		busy := reactive.NewSignal(false)
		colour := reactive.NewSignal("#4f7cff")

		reactive.BindAttr(root.Find("[data-bind-href]"), "href", url.Get)
		reactive.BindClass(root.Find("[data-bind-class]"), "on", active.Get)
		reactive.BindAttr(root.Find("[data-bind-disabled]"), "disabled", func() string {
			if busy.Get() {
				return "disabled"
			}
			return "" // empty removes the attribute
		})
		reactive.BindStyle(root.Find("[data-bind-style]"), "color", colour.Get)

		toggle := func(s *reactive.Signal[bool]) func(wasmwrap.Event) {
			return func(wasmwrap.Event) {
				s.Update(func(v bool) bool { return !v })
			}
		}
		reactive.On(root.Find("[data-toggle-class]"), "click", toggle(active))
		reactive.On(root.Find("[data-toggle-disabled]"), "click", toggle(busy))
		reactive.On(root.Find("[data-recolour]"), "click", func(wasmwrap.Event) {
			colour.Set("#ff4f7c")
		})
		reactive.On(root.Find("[data-retarget]"), "click", func(wasmwrap.Event) {
			url.Set("https://pkg.go.dev/github.com/LukasDerBaum42/astwerk")
		})
	})
}

type post struct {
	ID    string
	Title string
}

// asyncDemo drives a Resource from a loader that stands in for a network call,
// binding its Loading/Err/Data signals and reloading on demand.
func asyncDemo() {
	reactive.Hydrate("#async-demo", func(root wasmwrap.Element) {
		posts := reactive.NewResource(func() ([]post, error) {
			time.Sleep(500 * time.Millisecond) // stands in for wasmwrap.Fetch
			return []post{
				{ID: "1", Title: "Signals track their own dependencies"},
				{ID: "2", Title: "Keyed lists move nodes instead of rebuilding them"},
				{ID: "3", Title: "A binding owns exactly the property it writes"},
			}, nil
		})

		reactive.BindText(root.Find("[data-status]"), func() string {
			if posts.Loading.Get() {
				return "Loading…"
			}
			if err := posts.Err.Get(); err != nil {
				return "Error: " + err.Error()
			}
			return strconv.Itoa(len(posts.Data.Get())) + " posts loaded"
		})

		reactive.BindList(root.Find("[data-posts]"), posts.Data.Get,
			func(p post) string { return p.ID },
			func(p post) wasmwrap.Element {
				return wasmwrap.Create("li").SetText(p.Title)
			})

		reactive.On(root.Find("[data-reload]"), "click", func(wasmwrap.Event) {
			posts.Reload()
		})
	})
}

func canvasDemo() {
	reactive.Hydrate("#canvas-demo", func(root wasmwrap.Element) {
		canvas := root.Find("[data-canvas]").AsCanvas()
		ctx2d := canvas.Context2D()
		w, h := canvas.FitToDisplay(ctx2d)

		pointer := reactive.NewSignal(w / 2)
		reactive.On(canvas.Element(), "pointermove", func(e wasmwrap.Event) {
			x, _ := e.PosIn(canvas.Element())
			pointer.Set(float64(x))
		})

		elapsed := 0.0
		draw := func(dt float64) {
			elapsed += dt
			ctx2d.Clear(0, 0, w, h)
			for i := 0.0; i < w; i += 6 {
				wave := math.Sin(i/40+elapsed*2) * 40
				// Peek, not Get: this runs outside an effect, and a dependency
				// here would be meaningless — the loop already redraws every frame.
				lift := math.Exp(-math.Pow(i-pointer.Peek(), 2)/4000) * 30
				ctx2d.FillStyle("#4f7cff").FillRect(i, h/2+wave-lift, 3, 3)
			}
		}

		cancel := ctx2d.RequestAnimationFrame(draw)
		running := true

		toggle := root.Find("[data-toggle]")
		reactive.On(toggle, "click", func(wasmwrap.Event) {
			if running {
				cancel()
				toggle.SetText("Resume")
			} else {
				cancel = ctx2d.RequestAnimationFrame(draw)
				toggle.SetText("Pause")
			}
			running = !running
		})
	})
}

// stopwatchDemo is a running stopwatch: wasmwrap.SetInterval ticks a signal,
// and bindings render it like anything else.
func stopwatchDemo() {
	reactive.Hydrate("#stopwatch-demo", func(root wasmwrap.Element) {
		elapsed := reactive.NewSignal(0.0)
		var cancel func()
		running := false

		start := root.Find("[data-start]")
		reactive.BindText(root.Find("[data-time]"), func() string {
			v := elapsed.Get()
			ms := int(v*1000) % 1000 / 100
			sec := int(v) % 60
			min := int(v) / 60
			return fmt.Sprintf("%d:%02d.%d", min, sec, ms)
		})

		reactive.On(start, "click", func(wasmwrap.Event) {
			if running {
				cancel()
				start.SetText("Resume")
			} else {
				cancel = wasmwrap.SetInterval(func() {
					elapsed.Update(func(v float64) float64 { return v + 0.1 })
				}, 100)
				start.SetText("Pause")
			}
			running = !running
		})
		reactive.On(root.Find("[data-reset]"), "click", func(wasmwrap.Event) {
			elapsed.Set(0)
		})
	})
}

// filterDemo filters a list through a two-way bound search box: the query
// signal drives a Computed slice, and BindList renders whatever it contains.
func filterDemo() {
	reactive.Hydrate("#filter-demo", func(root wasmwrap.Element) {
		posts := []post{
			{ID: "1", Title: "Signals track their own dependencies"},
			{ID: "2", Title: "Keyed lists move nodes instead of rebuilding them"},
			{ID: "3", Title: "A binding owns exactly the property it writes"},
			{ID: "4", Title: "Computed values chain and gate"},
			{ID: "5", Title: "The escape hatch is plain Go"},
		}
		query := reactive.NewSignal("")
		reactive.BindValue(root.Find("[data-query]"), query)

		matches := reactive.Computed(func() []post {
			q := strings.ToLower(strings.TrimSpace(query.Get()))
			if q == "" {
				return posts
			}
			var out []post
			for _, p := range posts {
				if strings.Contains(strings.ToLower(p.Title), q) {
					out = append(out, p)
				}
			}
			return out
		})

		reactive.BindText(root.Find("[data-count]"), func() string {
			return strconv.Itoa(len(matches.Get()))
		})
		reactive.BindList(root.Find("[data-results]"), matches.Get,
			func(p post) string { return p.ID },
			func(p post) wasmwrap.Element {
				return wasmwrap.Create("li").SetText(p.Title)
			})
	})
}
