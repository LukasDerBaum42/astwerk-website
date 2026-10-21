+++
Title = "How astwerk thinks"
Description = "One Go program describes the whole output directory as a tree; Build writes it. Interactivity is a choice: tiny compiled JS, or full Go in the browser."
Order = 10
Group = "Introduction"
+++

astwerk generates a static site from a Go program. There is no config file, no
convention scanner, no plugin pipeline — one program, and everything else is a
consequence of that program being ordinary Go.

## The whole pipeline

```
Your Go program
      │
      ▼
   Node tree        ← you write this: the output directory, as data
      │
      ▼
   Build()
      ├── HTML          pages and layouts (templ components)
      ├── assets        copied directories
      ├── scripts       compiled to WebAssembly or TypeScript
      └── generated     pages computed at build time (content collections)
```

You describe the output directory as a tree of `Node`s and call `Build`. The
tree is the site — nothing is discovered, nothing is inferred. What you write
is exactly what gets written.

## Then the client side

Static HTML is the default, and it stays static until you opt in:

```
Static HTML
    │
    ├── x        → tiny generated JavaScript: a little interactivity
    │
    └── reactive → full Go in the browser (via WASM)
```

- **`x`** compiles a restricted reactive model into a small script at build
  time. A counter, a toggle, a form, a keyed list — the common 90% — costs a
  few kilobytes of generated JS and no WASM download.
- **`reactive`** is the same model running as real Go. When the interactive
  half of a page *is* a program — goroutines, the standard library, a router —
  that's the WASM path, with `wasmwrap` underneath as the raw DOM layer.

The two are [two implementations of one reactive model](../x/#documentation):
same concepts, different runtimes. [Which tool do I use?](../which-tool/) is
the decision table.

## The three mechanisms

Everything the framework does fits under three headings, each readable in one
sitting (full detail on [How it works](../how-it-works/)):

| Mechanism | What it does |
| --- | --- |
| The tree walker | `Build` recurses the tree and writes files — depth-first, in sorted key order, reproducibly |
| Locale derivation | `BuildLocales` mirrors the tree per language; you override only what genuinely differs |
| Dependency tracking | signals remember what reads them, so bindings update exactly what they own |

## What that means for you

Because the tree is data in a Go program, everything the language can do is
available at build time: compute children from a database, an API, or a
directory of markdown; generate pages in a loop; conditionally include an
entire subtree. And because pages are templ components, the type checker runs
over your templates.

The corollary: nothing happens by convention. If a site needs something the
tree can't express, it falls back to plain Go — [never a wall it can't get
past](../escape-hatches/).
