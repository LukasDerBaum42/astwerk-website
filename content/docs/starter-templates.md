+++
Title = "Starter templates"
Description = "Seventeen copyable .templ files. Copied, never imported."
Order = 80
Group = "Reference"
+++

`starter/` is boilerplate to copy, not a library to import. Nothing in `ssg` or
`content` references it. Once the files are in your project they are your code —
no base class, nothing to override, nothing that breaks when astwerk updates.

<details>
<summary>Why copy, not import</summary>

A template library would be a dependency with opinions: it would pin component
signatures, force updates whenever astwerk changes, and hand you a base class
to override. Copied files are your code — change anything, and an astwerk
update can't break them.

</details>

## Documentation

### The x widgets

`widgets/` uses [`x`](../x/) — a real import, so unlike the placeholder paths it
needs no rewriting. Three rules:

- **Wrap the page content in `x.Document`.** That is what emits the generated
  script; outside it the widgets still render, just statically.
- **Name anything a script reads** with `x.Named`, so the script can call
  `count.get()`.
- **Treat a `<script>` as the escape hatch** for anything a combinator can't
  say — see `widgets/todo.templ`.

### The files

| File | Purpose |
| --- | --- |
| `layout/base_layout.templ` | `<html>`/`<head>`/`<body>`, locale-aware `<html lang>`, content via `{ children... }` |
| `layout/nav.templ` | nav driven by a `[]NavLink` slice, marks the active link |
| `layout/footer.templ` | copyright, locale switcher, social links |
| `content/collection_index.templ` | grid of tiles — the list half of list + detail |
| `content/collection_item.templ` | `Tile` for the grid, `CollectionItem` for the detail page |
| `content/tag_list.templ` | tags from front matter |
| `pages/home.templ` | front page, near-empty on purpose |
| `pages/about.templ` | a plain one-off page |
| `pages/not_found.templ` | 404 page, with wiring instructions |
| `pages/reactive.templ` | a page of `widgets/`, wrapped in `x.Document` |
| `widgets/counter.templ` | counter with increment/reset, a toggle-`When` section and a title effect |
| `widgets/todo.templ` | two-way input + keyed list, with the script escape hatch for add |
| `widgets/price.templ` | computed value over two two-way number inputs |
| `wasm/wasm_script.templ` | loads a `.wasm` build plus `wasm_exec.js` |
| `wasm/canvas_host.templ` | sized `<canvas>` paired with its script |
| `meta/seo_head.templ` | description, Open Graph and Twitter tags |
| `meta/rss_feed.templ` | RSS 2.0 feed over a `[]FeedItem` |

<details>
<summary>Why they're shaped this way</summary>

**Pages take `(title, path, prefix)`**, which is exactly what `ssg.Templ` fills
in, so a starter page drops into a tree with no adapter:

```go
"about": {Title: "About", Page: ssg.Templ(pages.About)},
```

**Tile links are bare slugs** — `href="thing/"` — so they resolve relative to
whichever index page they're on, and work unchanged inside a locale subtree.

**`rss_feed.templ` marshals with `encoding/xml`** rather than writing markup.
`<link>` is a void element in HTML and templ's parser is HTML-aware, so written
as markup it self-closes and you get a feed whose items have no links at all.

**`wasm_script.templ` passes its path via `data-wasm`.** templ treats `<script>`
contents as raw text, so an interpolated expression in the body would be emitted
literally.

</details>

## Example

Copy the folder, fix the placeholder import path, generate:

```sh
cp -r starter/* .   # or download the folder from GitHub
grep -rl example.com/yoursite . | xargs sed -i 's|example.com/yoursite|YOUR/MODULE/PATH|g'
templ generate
```

The directories are separate Go packages; flatten them if you prefer. Copy the
files out before running `templ generate`: generating inside an astwerk checkout
produces `_templ.go` files importing the placeholder path, which breaks
`go build ./...` there.
