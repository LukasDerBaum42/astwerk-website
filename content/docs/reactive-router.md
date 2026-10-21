+++
Title = "reactive — routing & async"
Description = "Client-side routing and loading async data into signals."
Order = 180
Group = "Add interactivity"
+++

Two things that reach outside the page: routing between views, and loading data
asynchronously. Both land in the same state model as the [bindings](../reactive-bindings/) —
routes are a signal of the current path, and a resource is three signals.

## Documentation

### Router

```go
reactive.Router(root, []reactive.Route{
	{Path: "/", View: home},
	{Path: "/projects/:slug", View: func(p reactive.Params) wasmwrap.Element {
		return project(p.Get("slug"))
	}},
	{Path: "/*", View: notFound},
})
```

`Router` renders the view for the current path into `root`, and swaps it when
the path changes. Routes are tried in order, so put specific ones first. Each
view gets its own scope, so navigating away disposes its effects.

### Route and Params

Patterns match segment by segment:

| Pattern | Matches |
| --- | --- |
| `/` | exactly the root |
| `/projects/:slug` | one segment, captured into `Params` |
| `/*` | whatever remains |

```go
func view(p reactive.Params) wasmwrap.Element {
	slug := p.Get("slug") // "" when the route has no such capture
}
```

### Navigate, Replace, Path

```go
reactive.Navigate("/projects/42") // push a new entry
reactive.Replace("/projects/42")  // replace the current entry
current := reactive.Path()        // the path, as a tracked signal
```

`Path()` is a signal — read it inside an effect or a binding and it re-runs
when the URL changes.

### InterceptLinks

```go
reactive.InterceptLinks(wasmwrap.Body())
```

Makes same-origin anchor clicks route client-side, leaving modified clicks,
external links and `data-native` anchors alone.

### Resource

```go
posts := reactive.NewResource(func() ([]Post, error) {
	return reactive.FetchJSON[[]Post]("/api/posts")
})
```

A resource runs its loader once when created and exposes the result as three
signals:

```go
type Resource[T any] struct {
	Data    *Signal[T]    // the loaded value, or T's zero value before the first success
	Err     *Signal[error] // the last error, cleared when a load succeeds
	Loading *BoolSignal   // true while a load is in flight
}
```

```go
reactive.BindShow(spinner, posts.Loading)
reactive.BindText(errBox, reactive.Computed(func() string {
	if err := posts.Err.Get(); err != nil {
		return err.Error()
	}
	return ""
}))
```

`Data` is a plain `Signal[[]T]` — the generic kind, because `T` here is
whatever the endpoint returns — while `BindList` needs a `*ListSignal[T]`. A
loaded list therefore renders through [`BindTempl`](../reactive-templ/), which
replaces the whole list on load: the honest trade-off for a display-only chunk.
When rows must be keyed, feed a `ListSignal` from the loader instead — the
loader returns the slice and an effect (or the view) copies it into the list
signal, keeping the keyed-row behaviour.

`Reload` re-runs the loader. The loader runs on its own goroutine, so a
blocking `Fetch` belongs there.

<details>
<summary>Why a stale response can't win</summary>

If a reload is in flight and the user reloads again, the older response must
not overwrite the newer one. A resource tracks which load it started, and a
response that isn't from the latest load is dropped.

</details>

## Example

```go
reactive.Router(root, []reactive.Route{
	{Path: "/", View: home},
	{Path: "/projects/:slug", View: func(p reactive.Params) wasmwrap.Element {
		return project(p.Get("slug"))
	}},
	{Path: "/*", View: notFound},
})
reactive.InterceptLinks(wasmwrap.Body())
```

A single-page shell in four lines: two routes, a capture, a catch-all, and
link interception.
