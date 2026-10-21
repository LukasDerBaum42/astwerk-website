+++
Title = "How astwerk thinks"
Description = "One Go program describes the whole output directory as a tree; Build writes it. Static HTML, no client runtime in core."
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

## The client side is your choice

Core astwerk outputs static HTML. It ships no client runtime, no framework and
no JavaScript of its own. If a page needs behaviour, that is an ordinary
`<script>` tag, a `.ts` file handed to `tsc`, or a Go file compiled to WASM by
[`CompileFrom`](../scripts-and-wasm/) — all of them just files in the tree.

The optional client-side packages live in the separate
[astwerk-live](https://github.com/LukasDerBaum42/astwerk-live) extension:
compiled markup (`x`), signals and bindings (`reactive`), direct DOM access
(`wasmwrap`) and a browser DOM for tests (`jsdom`). That side is still in heavy
redesign and very likely to change or be removed. The
[client-side tools](../which-tool/) page is the honest entry point.

## The three mechanisms

Everything the framework does fits under three headings, each readable in one
sitting (full detail on [How it works](../how-it-works/)):

| Mechanism | What it does |
| --- | --- |
| The tree walker | `Build` recurses the tree and writes files — depth-first, in sorted key order, reproducibly |
| Locale derivation | `BuildLocales` mirrors the tree per language; you override only what genuinely differs |
| Content loading | `content` turns markdown and front matter into typed items, ready for `Generate` |

## What that means for you

Because the tree is data in a Go program, everything the language can do is
available at build time: compute children from a database, an API, or a
directory of markdown; generate pages in a loop; conditionally include an
entire subtree. And because pages are templ components, the type checker runs
over your templates.

The corollary: nothing happens by convention. If a site needs something the
tree can't express, it falls back to plain Go — [never a wall it can't get
past](../escape-hatches/).
