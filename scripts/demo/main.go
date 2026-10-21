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
//
// The reactive v2 split types are used throughout: Num/Bool/Str/List for state
// (each kind carries only the methods that make sense for it), the generic
// Signal only where T is genuinely arbitrary — a Resource's Data, say.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
	"github.com/a-h/templ"
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
	count := reactive.Num(0)
	doubled := reactive.Computed(func() int { return count.Get() * 2 })

	reactive.Hydrate("#counter", func(root wasmwrap.Element) {
		reactive.BindText(root.Find("[data-count]"), count)
		reactive.BindText(root.Find("[data-doubled]"), doubled)

		reactive.On(root.Find("[data-inc]"), "click", func(wasmwrap.Event) {
			count.Inc(1)
		})
		reactive.On(root.Find("[data-dec]"), "click", func(wasmwrap.Event) {
			count.Dec(1)
		})
		reactive.On(root.Find("[data-reset]"), "click", func(wasmwrap.Event) {
			count.Set(0)
		})
	})
}

type item struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// todoStorage persists the todo items under a fixed localStorage key — one
// call per mutation, the WASM advantage the x twin notes it keeps a script
// for.
const todoKey = "astwerk-todo"

func loadTodos() []item {
	raw := wasmwrap.LocalStorage().Get(todoKey)
	if raw == "" {
		return nil
	}
	var items []item
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func saveTodos(items []item) {
	if b, err := json.Marshal(items); err == nil {
		wasmwrap.LocalStorage().Set(todoKey, string(b))
	}
}

func todoList() {
	items := reactive.List([]item{
		{ID: "1", Text: "Read the docs"},
		{ID: "2", Text: "Build something"},
	})
	if saved := loadTodos(); saved != nil {
		items.Set(saved)
	}

	reactive.Hydrate("#todo", func(root wasmwrap.Element) {
		field := root.Find("[data-new]")
		draft := reactive.Str("")
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
			saveTodos(items.Get())
		})

		reactive.BindText(root.Find("[data-remaining]"), reactive.Computed(func() string {
			return fmt.Sprintf("%d remaining", len(items.Get()))
		}))

		reactive.BindList(root.Find("[data-list]"), items,
			func(it item) string { return it.ID },
			func(it item, row *reactive.Row) wasmwrap.Element {
				// The same row shape as the x twin: text and a remove button,
				// keyed by id.
				el := wasmwrap.Create("li").SetText(it.Text)
				remove := wasmwrap.Create("button").SetText("remove")

				reactive.On(remove, "click", func(wasmwrap.Event) {
					items.Update(func(all []item) []item {
						return slices.DeleteFunc(slices.Clone(all), func(o item) bool {
							return o.ID == it.ID
						})
					})
					saveTodos(items.Get())
				})
				return el.Append(remove)
			})
	})
}

// formDemo two-way binds two number fields, derives a total from them, and
// lets a checkbox switch the total to include 19% tax — the same features as
// the x forms demo, with wasm's real if where the x twin uses an x.Not memo.
func formDemo() {
	reactive.Hydrate("#form-demo", func(root wasmwrap.Element) {
		price := reactive.Num(12.0)
		qty := reactive.Num(2.0)
		total := reactive.Computed(func() float64 { return price.Get() * qty.Get() })
		tax := reactive.Bool(false)

		reactive.BindNumber(root.Find("[data-price]"), price)
		reactive.BindNumber(root.Find("[data-qty]"), qty)
		reactive.BindText(root.Find("[data-total]"), reactive.Computed(func() string {
			return fmt.Sprintf("%.2f", total.Get())
		}))

		// Exactly one total is visible: the plain one while the box is off, the
		// taxed one while it's on — both stay live, nothing is rebuilt.
		reactive.BindShow(root.Find("[data-total]"), reactive.Computed(func() bool {
			return !tax.Get()
		}))
		reactive.BindText(root.Find("[data-taxed]"), reactive.Computed(func() string {
			if tax.Get() {
				return fmt.Sprintf("%.2f", total.Get()*1.19)
			}
			return ""
		}))
		reactive.BindShow(root.Find("[data-taxed]"), tax)
		reactive.BindChecked(root.Find("[data-tax]"), tax)
	})
}

// bindingsDemo drives four one-property bindings — href, class, disabled,
// style — from four signals, each owned by a control button. The WASM twin
// of the x attribute-bindings demo.
func bindingsDemo() {
	reactive.Hydrate("#bindings-demo", func(root wasmwrap.Element) {
		url := reactive.Str("https://templ.guide")
		active := reactive.Bool(true)
		busy := reactive.Bool(false)
		colour := reactive.Str("#4f7cff")

		reactive.BindAttr(root.Find("[data-bind-href]"), "href", url)
		reactive.BindClass(root.Find("[data-bind-class]"), "on", active)
		reactive.BindAttr(root.Find("[data-bind-disabled]"), "disabled", reactive.Computed(func() string {
			if busy.Get() {
				return "disabled"
			}
			return "" // empty removes the attribute
		}))
		reactive.BindStyle(root.Find("[data-bind-style]"), "color", colour)

		reactive.On(root.Find("[data-toggle-class]"), "click", func(wasmwrap.Event) {
			active.Toggle()
		})
		reactive.On(root.Find("[data-toggle-disabled]"), "click", func(wasmwrap.Event) {
			busy.Toggle()
		})

		// The disabled state is visible, not just functional: a status line,
		// and the toggle's own aria-pressed and .on class — the same signal
		// driving all three.
		reactive.BindText(root.Find("[data-bind-status]"), reactive.Computed(func() string {
			if busy.Get() {
				return "the save button is disabled"
			}
			return "the save button is enabled"
		}))
		reactive.BindAttr(root.Find("[data-toggle-disabled]"), "aria-pressed", reactive.Computed(func() string {
			if busy.Get() {
				return "true"
			}
			return "false"
		}))
		reactive.BindClass(root.Find("[data-toggle-disabled]"), "on", busy)

		reactive.On(root.Find("[data-recolour]"), "click", func(wasmwrap.Event) {
			colour.Set("#ff4f7c")
		})
		reactive.On(root.Find("[data-retarget]"), "click", func(wasmwrap.Event) {
			url.Set("https://pkg.go.dev/github.com/LukasDerBaum42/astwerk")
		})
	})
}

type post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// asyncDemo drives a Resource from a real fetch of a file the site itself
// generated (examples/data/posts.json — a Files entry on the examples node).
// Data renders through BindTempl: a Resource's Data is a plain Signal[[]post]
// and BindList needs a ListSignal, so the whole list re-renders on load — the
// honest trade-off for a display-only chunk. Feed a ListSignal from the loader
// when rows must be keyed.
func asyncDemo() {
	reactive.Hydrate("#async-demo", func(root wasmwrap.Element) {
		posts := reactive.NewResource(func() ([]post, error) {
			return reactive.FetchJSON[[]post]("../data/posts.json")
		})

		reactive.BindText(root.Find("[data-status]"), reactive.Computed(func() string {
			if posts.Loading.Get() {
				return "Loading…"
			}
			if err := posts.Err.Get(); err != nil {
				return "Error: " + err.Error()
			}
			return strconv.Itoa(len(posts.Data.Get())) + " posts loaded"
		}))

		reactive.BindTempl(root.Find("[data-posts]"), func() templ.Component {
			if err := posts.Err.Get(); err != nil {
				return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
					_, err := io.WriteString(w, `<li class="warn">Couldn't load posts.</li>`)
					return err
				})
			}
			return postList(posts.Data.Get())
		})

		reactive.On(root.Find("[data-reload]"), "click", func(wasmwrap.Event) {
			posts.Reload()
		})
	})
}

// postList renders the loaded posts as a fragment of <li> rows — the
// display-only chunk BindTempl re-renders on load.
func postList(posts []post) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		for _, p := range posts {
			if _, err := io.WriteString(w, "<li>"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, templ.EscapeString(p.Title)); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "</li>"); err != nil {
				return err
			}
		}
		return nil
	})
}

func canvasDemo() {
	reactive.Hydrate("#canvas-demo", func(root wasmwrap.Element) {
		canvas := root.Find("[data-canvas]").AsCanvas()
		ctx2d := canvas.Context2D()
		w, h := canvas.FitToDisplay(ctx2d)

		pointer := reactive.Num(w / 2)
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
		elapsed := reactive.Num(0.0)
		var cancel func()
		running := false

		start := root.Find("[data-start]")
		reactive.BindText(root.Find("[data-time]"), reactive.Computed(func() string {
			v := elapsed.Get()
			ms := int(v*1000) % 1000 / 100
			sec := int(v) % 60
			min := int(v) / 60
			return fmt.Sprintf("%d:%02d.%d", min, sec, ms)
		}))

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
// signal drives an effect that computes the visible slice into a ListSignal,
// and BindList moves the keyed rows instead of rebuilding them. x has no
// list-filter combinator either — filtering is plain Go (or plain JS), the
// documented escape hatch; the DOM diffing is the framework's.
func filterDemo() {
	reactive.Hydrate("#filter-demo", func(root wasmwrap.Element) {
		posts := []post{
			{ID: "1", Title: "Signals track their own dependencies"},
			{ID: "2", Title: "Keyed lists move nodes instead of rebuilding them"},
			{ID: "3", Title: "A binding owns exactly the property it writes"},
			{ID: "4", Title: "Computed values chain and gate"},
			{ID: "5", Title: "The escape hatch is plain Go"},
		}
		query := reactive.Str("")
		reactive.BindValue(root.Find("[data-query]"), query)

		matches := reactive.List(posts)
		reactive.Effect(func() {
			q := strings.ToLower(strings.TrimSpace(query.Get()))
			if q == "" {
				matches.Set(posts)
				return
			}
			var out []post
			for _, p := range posts {
				if strings.Contains(strings.ToLower(p.Title), q) {
					out = append(out, p)
				}
			}
			matches.Set(out)
		})

		reactive.BindText(root.Find("[data-count]"), reactive.Computed(func() int {
			return len(matches.Get())
		}))
		reactive.BindList(root.Find("[data-results]"), matches,
			func(p post) string { return p.ID },
			func(p post, row *reactive.Row) wasmwrap.Element {
				return wasmwrap.Create("li").SetText(p.Title)
			})
	})
}
