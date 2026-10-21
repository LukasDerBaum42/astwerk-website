## Why This Is Bad Documentation

The following looks reasonable at first glance:

```md
# Getting started

Install astwerk, write a tree, generate a site.

astwerk generates a static site from a Go program. You write a tree describing
the output directory; it walks the tree and writes files.
```

But it has several problems.

### 1. The important information is mixed together

The introduction, installation instructions, architecture explanation, tutorial,
and API details are all mixed into one page.

For example:

```md
`ssg.Templ` adapts the `(title, path, prefix)` signature — the walker knows each
page's path and locale, so it fills those in and you supply only the title.
```

This is important information, but it appears **after a large example** and is
only explained after the reader has already had to use the API.

It should either be documented next to `ssg.Templ` or linked to from the example.

---

### 2. The example is too large for the documentation

The example contains a complete template and a complete `main.go`.

That makes the reader scan through a lot of code before seeing what actually
matters.

For example, this is probably enough to demonstrate the basic idea:

```go
root := ssg.Node{
    Title: "Home",
    Page:  ssg.Templ(views.Page),
    Children: map[string]ssg.Node{
        "about": {Title: "About", Page: ssg.Templ(views.Page)},
    },
}

ssg.Build(root, ssg.BuildOptions{})
```

The full HTML template can be shown separately if necessary.

**Examples should demonstrate the feature, not reproduce an entire project.**

---

### 3. It explains too much in the main text

This paragraph:

```md
`Build` walked the tree: made each directory, rendered `Page` into `index.html`,
recursed into `Children` using the map keys as directory names, and copied
`style/` in as-is. `ssg.Templ` adapts the `(title, path, prefix)` signature —
the walker knows each page's path and locale, so it fills those in and you
supply only the title.
```

contains several separate concepts:

- How `Build` works
- How `Children` works
- How copying works
- How `Templ` works
- How paths work
- How locales work

That's too much information immediately after the example.

The basic documentation should say something like:

```md
`Build` walks the node tree and generates the corresponding files.

See [The Node tree] for details about nodes and `Children`.
```

The deeper explanation belongs in the relevant documentation page.

---

### 4. It documents implementation instead of usage

This:

```md
the walker knows each page's path and locale, so it fills those in
```

describes how the implementation works.

The user mostly needs to know:

```md
`ssg.Templ` turns a templ component into a page that can be used by a node.
```

If understanding the implementation is useful, put it behind an optional
"How it works" section.

---

### 5. There is an unexplained placeholder

```md
## Dev mode

go run . --dev

Why --dev exists
```

This is neither documentation nor a finished explanation.

If `--dev` is important, explain it briefly:

```md
## Dev mode

Use `--dev` to enable development-specific behavior.

```sh
go run . --dev
```

<details>
<summary>What does dev mode do?</summary>

Detailed explanation...

</details>
```

If it isn't ready to be documented, **don't pretend it is documented**.

---

### 6. It uses unnecessary tutorial language

```md
Prefer to build something?
```

and

```md
The one page worth reading in full
```

aren't inherently wrong, but they don't help explain the API.

Documentation should prioritize useful information over conversational filler.

---

### 7. Installation is incomplete

It says:

```sh
go get github.com/LukasDerBaum42/astwerk
go install github.com/a-h/templ/cmd/templ@latest
```

but doesn't explain:

- Which Go version is required.
- Whether `go get` is actually required.
- Whether the project already has a `go.mod`.
- What `templ` is needed for.
- What the user should do if `templ` isn't in `$PATH`.

Installation documentation should contain the **minimum information needed to get running**, not merely a list of commands.

---

### 8. The generated output is useful — but underused

This part is good:

```text
build/
  index.html
  about/index.html
  style/…
```

It immediately shows what Astwerk produces.

This is the kind of information documentation should favor: **concrete results**.

A good example often follows:

```text
Input → Command → Output
```

rather than several paragraphs explaining what happened.

---

### 9. The page tries to do too many jobs

This single page is simultaneously:

- An installation guide
- A quick-start tutorial
- An API explanation
- An architecture explanation
- A dev-mode reference
- A local-server guide
- A documentation index

Those should be separate concerns.

A better structure is:

```text
Getting Started
    → minimal working example

Node Tree
    → Node, Children, paths, copying

Templ
    → templ integration and arguments

Dev Mode
    → --dev behavior

Serving
    → local development server
```

The Getting Started page should get someone from **zero → working site** as quickly as possible.

---

## What the page should feel like

A good Getting Started page should be closer to:

```md
# Getting Started

Astwerk generates static sites from a Go program.

## Install

```sh
go get github.com/LukasDerBaum42/astwerk
go install github.com/a-h/templ/cmd/templ@latest
```

## Example

Create a templ component:

```go
templ Page(title string) {
    <html>
        <head><title>{ title }</title></head>
        <body><h1>{ title }</h1></body>
    </html>
}
```

Create the site tree:

```go
root := ssg.Node{
    Title: "Home",
    Page: ssg.Templ(views.Page),
    Children: map[string]ssg.Node{
        "about": {
            Title: "About",
            Page: ssg.Templ(views.Page),
        },
    },
}
```

Build it:

```sh
templ generate
go run .
```

This produces:

```text
build/
├── index.html
└── about/
    └── index.html
```

`ssg.Build` walks the tree and generates the corresponding files.

## Next

- [The Node tree] — learn how site trees work.
- [Content and collections] — add content.
- [Locales] — add translations.
```

That's enough to get started.

Everything else can live in the appropriate documentation page.

---

## The Rule

When reviewing documentation, ask:

> **"Could I delete this paragraph and still understand how to use the thing?"**

If yes, either remove it or move it into an optional section or a more appropriate document.

The main documentation should be the **shortest path from "What is this?" to "I can use it."**
