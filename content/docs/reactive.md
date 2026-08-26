+++
Title = "reactive — state"
Description = "Signals, effects, computed values and ownership — the state model behind the bindings."
Order = 60
Group = "Client-side"
+++

`reactive` is opt-in state management on top of `wasmwrap`, fine-grained rather
than virtual-DOM based: a binding updates exactly the property it owns, and a
keyed list moves nodes rather than rebuilding them. This page is the state
model — signals, effects and ownership. The rest of the package is split across
[DOM bindings](../reactive-bindings/), [routing & async](../reactive-router/)
and [templ in the browser](../reactive-templ/).

The lightweight path is [`x`](../x/): the same reactive model, compiled to plain
JavaScript at build time, with no WASM and no framework. See both running on the
[Examples](../../examples/) page.

## Documentation

### Signal

```go
count := reactive.NewSignal(0)

count.Get()    // read — subscribes if inside an Effect
count.Peek()   // read without subscribing
count.Set(5)
count.Update(func(v int) int { return v + 1 })
```

`Set` compares against the current value and does nothing if they're equal, so a
redundant write wakes nobody.

### Subscribe

```go
unsub := count.Subscribe(func(v int) {
	fmt.Println(v)
})
unsub()
```

A manual subscription: fires on every change, runs outside effect tracking, and
returns an unsubscribe function. Effects use the same mechanism internally.

### Effect

```go
reactive.Effect(func() {
	if useA.Get() {
		fmt.Println(a.Get()) // only a is a dependency
	} else {
		fmt.Println(b.Get()) // only b is
	}
})
```

Runs immediately and again whenever a signal it read changes. Dependencies are
tracked, not declared — the set is re-collected on every run, so a branch not
taken creates no dependency.

<details>
<summary>Why an effect runs immediately</summary>

The first run is how the effect discovers what it depends on: every signal read
while it runs registers as a dependency. Nothing is declared up front, so there
is no list to keep in sync with the code.

</details>

<details>
<summary>Why dependencies are re-collected every run</summary>

Each re-run unlinks the effect from everything it read last time and collects
afresh as it goes. That is what makes a branch not taken free: flip `useA` and
the effect re-runs, drops `a`, and only `b` keeps it alive.

</details>

### Batch

```go
reactive.Batch(func() {
	first.Set("Ada")
	last.Set("Lovelace")
})
```

Collapses many writes into one update per affected effect; batches nest.

### Untracked

```go
v := reactive.Untracked(func() string { return count.Get() })
```

Runs `fn` without recording any dependencies — reads inside don't subscribe the
enclosing effect. `Peek` is the single-signal version of this.

### Computed and Memo

```go
full := reactive.Computed(func() string {
	return first.Get() + " " + last.Get()
})

full.Get() // "Ada Lovelace"
```

`Computed` returns a `*Memo[T]` — a signal whose value is derived from other
signals. It recomputes when its inputs change and is itself trackable, so
derived values chain. If the recomputed value equals the previous one,
downstream effects don't run.

```go
m := reactive.Computed(fn)
m.Get()        // like any signal
m.Peek()       // without subscribing
m.Subscribe(f) // manual subscription
m.Dispose()    // detach from inputs, stop recomputing
```

### Scope and OnCleanup

```go
dispose := reactive.Scope(func() {
	reactive.BindText(a, …)
	reactive.BindText(b, …)
})
dispose() // tears down both
```

Effects nest: one created inside another is disposed and rebuilt when the outer
re-runs, so bindings inside a rebuilt view don't accumulate. `OnCleanup` hooks
into the same lifecycle. Every `Bind*` returns a dispose func; inside an effect
it's owned for you.

<details>
<summary>Threading</summary>

Go on js/wasm runs one goroutine at a time, so signals need no locking. A
goroutine may still write one — that's how `Resource` reports its result — and
the effects it triggers run on that goroutine.

</details>

## Example

```go
count := reactive.NewSignal(0)
doubled := reactive.Computed(func() int { return count.Get() * 2 })

reactive.Effect(func() {
	fmt.Println(doubled.Get()) // 0, then 6 after the Set below
})

count.Set(3)
```

A signal, a derived value, and an effect that tracks both — `Set(3)` recomputes
the memo and re-runs the effect exactly once.
