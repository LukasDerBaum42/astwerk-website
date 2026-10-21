+++
Title = "wasmwrap — canvas"
Description = "Typed 2D drawing — Canvas, Ctx2D, animation loops."
Order = 150
Group = "Add interactivity"
+++

The one place `wasmwrap` adds a real API surface: 2D drawing with typed calls
instead of raw `getContext("2d")` juggling. The rest of the wrapper is on
[wasmwrap](../wasmwrap/).

## Documentation

### Canvas

```go
canvas := wasmwrap.Query("#c").AsCanvas()
ctx := canvas.Context2D()
```

`AsCanvas` wraps a `<canvas>` element. `Canvas` then offers:

| Method | Does |
| --- | --- |
| `Context2D()` | the drawing context everything below lives on |
| `Size()` / `SetSize(w, h)` | reads and sets the drawing buffer |
| `Element()` | the element back, for styling or events |
| `FitToDisplay(ctx)` | sizes the buffer to the element's CSS size and returns the logical size |

### FitToDisplay

```go
canvas := wasmwrap.Query("#c").AsCanvas()
ctx := canvas.Context2D()
w, h := canvas.FitToDisplay(ctx)
```

Sizes the drawing buffer to the element's CSS size times the device pixel
ratio and returns the logical size to draw against.

<details>
<summary>Why skipping it makes drawings blurry</summary>

A canvas has two sizes that are easy to confuse: the `width`/`height`
attributes are the pixel buffer, and the CSS size is the display size. If they
differ — which they do on any high-DPI screen — the browser stretches the
buffer to fit, and every line gets smeared across pixels. `FitToDisplay` aligns
the buffer with the display so each buffer pixel is one screen pixel.

</details>

### Drawing

```go
ctx.FillStyle("#4f7cff").
	BeginPath().
	Arc(w/2, h/2, 40, 0, 2*math.Pi).
	Fill()
```

Everything chains, and state setters (`FillStyle`, `StrokeStyle`, `LineWidth`,
`Font`) apply until changed. Path building is the usual `BeginPath` /
`MoveTo` / `LineTo` / `Arc` / `ClosePath`, finished by `Fill` or `Stroke`.
`FillText` draws text and `MeasureText` returns its width.

<details>
<summary>The full Ctx2D surface</summary>

```go
ctx.StrokeStyle(color).LineWidth(w).Font(f)
ctx.Save().Restore()
ctx.FillRect(x, y, w, h).StrokeRect(x, y, w, h)
ctx.BeginPath().MoveTo(x, y).LineTo(x, y).Arc(x, y, r, a, b).ClosePath()
ctx.Fill().Stroke()
ctx.FillText(text, x, y)
w := ctx.MeasureText(text)
```

Everything chains and everything that can't be expressed drops to
`ctx.Value().Call(...)`.

</details>

### Animation

```go
cancel := ctx.RequestAnimationFrame(func(dt float64) {
	elapsed += dt
	ctx.Clear(0, 0, w, h)
	// draw
})
```

`RequestAnimationFrame` drives a loop rather than a single frame — it
reschedules itself and hands you the delta in seconds. The `cancel` it returns
stops the loop.

A draw loop runs its own animation, outside the stylesheet, so the site-wide
`prefers-reduced-motion` rule cannot collapse it. Canvas demos must check the
media query at boot themselves and draw a single static frame instead of
starting the loop when it matches — the same rule the site's own canvas demos
follow.

## Example

```go
//go:build js && wasm

package main

import (
	"math"

	"github.com/LukasDerBaum42/astwerk-live/wasmwrap"
)

func main() {
	canvas := wasmwrap.Query("#c").AsCanvas()
	ctx := canvas.Context2D()
	w, h := canvas.FitToDisplay(ctx)

	angle := 0.0
	ctx.RequestAnimationFrame(func(dt float64) {
		angle += dt
		ctx.Clear(0, 0, w, h)
		ctx.FillStyle("#4f7cff").
			BeginPath().
			Arc(w/2+math.Cos(angle)*50, h/2+math.Sin(angle)*50, 30, 0, 2*math.Pi).
			Fill()
	})
	select {}
}
```

A dot orbiting the canvas centre, driven by the animation loop. The same
pattern — clear, draw, repeat — is the skeleton of any canvas demo.
