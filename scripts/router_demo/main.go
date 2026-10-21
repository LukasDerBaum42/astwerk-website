//go:build js && wasm

// Command router-demo is the live client-side routing demo: a two-view SPA (a
// list of posts, a detail view) that owns its own iframe on the Examples page.
// A path-based router cannot live on a page whose URL belongs to the site's
// navigation — that is the honest reason it was "in code" — so the site builds
// this page as a standalone document and embeds it. The frame's URL is the
// demo's to drive; nothing here touches the site's own address bar.
//
// The views are built with the wasmwrap widgets, and the current path is a
// signal: Path() inside a binding re-runs it on every navigation, the browser's
// back and forward buttons included.
package main

import (
	"strings"

	"github.com/LukasDerBaum42/astwerk-live/reactive"
	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

type post struct {
	ID    string
	Title string
	Body  string
}

var posts = []post{
	{ID: "1", Title: "Signals track their own dependencies", Body: "Reading a signal inside an effect registers the dependency; the effect re-runs when the value changes."},
	{ID: "2", Title: "Keyed lists move nodes instead of rebuilding them", Body: "BindList reconciles by key, so rows keep their focus, scroll and transitions."},
	{ID: "3", Title: "The escape hatch is plain Go", Body: "Anything the framework can't express is a function call away — the whole stdlib ships with the module."},
}

// appRoot returns this page's own path — written into a data attribute by the
// site at build time — so the routes below match relative to it. The demo then
// works at any base: from disk, at a localhost root, or deployed under
// /astwerk-website/.
func appRoot(root wasmwrap.Element) string {
	return strings.TrimSuffix(root.Dataset().Get("app-root"), "/")
}

func main() {
	root := wasmwrap.Query("#app")
	base := appRoot(root)

	reactive.Router(root, []reactive.Route{
		{Path: base + "/", View: listView(base)},
		{Path: base + "/posts/:slug", View: detailView(base)},
		{Path: base + "/*", View: notFoundView(base)},
	})
	reactive.InterceptLinks(root)

	// The address bar, live: reactive.Path is a signal, so this binding re-runs
	// on every navigation — including the browser's back and forward buttons,
	// which Navigate never sees.
	reactive.BindText(wasmwrap.Query("[data-path]"), reactive.Computed(func() string {
		p := strings.TrimPrefix(reactive.Path(), base)
		if p == "" {
			p = "/"
		}
		return p
	}))

	// Keep the module alive: main returning would tear down every handler.
	select {}
}

// listView is the app's index: one link per post. The links are full URLs the
// router matches — InterceptLinks turns their clicks into Navigate calls.
func listView(base string) func(reactive.Params) wasmwrap.Element {
	return func(reactive.Params) wasmwrap.Element {
		view := wasmwrap.Panel()
		view.Class().Add("router-view")
		view.Append(wasmwrap.Create("h3").SetText("Posts"))
		for _, p := range posts {
			link := wasmwrap.Link(base+"/posts/"+p.ID, p.Title)
			link.Class().Add("router-link")
			view.Append(link)
		}
		note := wasmwrap.Create("p")
		note.Class().Add("demo-note")
		note.SetText("Click a post, then use the browser's back button — each view gets its own scope, disposed on navigation.")
		view.Append(note)
		return view
	}
}

// detailView renders one post by its :slug capture, with a link back to the
// list. The catch-all route above means an unknown slug lands on a view too.
func detailView(base string) func(reactive.Params) wasmwrap.Element {
	return func(p reactive.Params) wasmwrap.Element {
		slug := p.Get("slug")
		view := wasmwrap.Panel()
		view.Class().Add("router-view")

		for _, post := range posts {
			if post.ID == slug {
				view.Append(wasmwrap.Create("h3").SetText(post.Title))
				view.Append(wasmwrap.Create("p").SetText(post.Body))
				view.Append(wasmwrap.Link(base+"/", "← Back to the list"))
				return view
			}
		}
		view.Append(wasmwrap.Create("h3").SetText("No such post"))
		view.Append(wasmwrap.Create("p").SetText("The slug " + slug + " matches nothing in the demo."))
		view.Append(wasmwrap.Link(base+"/", "← Back to the list"))
		return view
	}
}

// notFoundView catches every unmatched path — the "/*" route.
func notFoundView(base string) func(reactive.Params) wasmwrap.Element {
	return func(reactive.Params) wasmwrap.Element {
		view := wasmwrap.Panel()
		view.Class().Add("router-view")
		view.Append(wasmwrap.Create("h3").SetText("404"))
		view.Append(wasmwrap.Create("p").SetText("No route matches this path."))
		view.Append(wasmwrap.Link(base+"/", "← Back to the list"))
		return view
	}
}
