+++
Title = "Getting started"
Description = "Install astwerk, write a tree, generate a site."
Order = 10
Group = "Start here"
+++

astwerk generates a static site from a Go program. You describe the output
directory as a tree of `Node`s; `Build` walks the tree and writes files.
Templates are [templ](https://templ.guide) components — real Go, type-checked
at compile time.

For a step-by-step walkthrough, see the [Build a site](../../tutorial/)
tutorial.

## Install

Requires Go 1.25 or newer.

```sh
go get github.com/LukasDerBaum42/astwerk
go install github.com/a-h/templ/cmd/templ@latest
```

`go get` adds astwerk to your module's `go.mod`. The templ CLI compiles
`.templ` files to Go — that's the whole toolchain, no Node.js, no bundler, no
plugin system. If `templ` isn't on your `PATH`, the `go run` form works without
installing anything:

```sh
go run github.com/a-h/templ/cmd/templ@latest generate
```

## Example

A template, `views/page.templ`:

```templ
package views

templ Page(title string, path string, prefix string) {
	<html>
		<head><title>{ title }</title></head>
		<body><h1>{ title }</h1></body>
	</html>
}
```

`main.go` — the tree:

```go
package main

import (
	"log"

	"github.com/LukasDerBaum42/astwerk/ssg"
	"example.com/site/views"
)

func main() {
	root := ssg.Node{
		Title: "Home",
		Page:  ssg.Templ(views.Page),
		Children: map[string]ssg.Node{
			"about": {Title: "About", Page: ssg.Templ(views.Page)},
		},
	}

	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
		log.Fatal(err)
	}
}
```

Build:

```sh
templ generate
go run .
```

That writes:

```text
build/
├── index.html
└── about/
    └── index.html
```

`ssg.Build` walks the tree and generates the corresponding files — each
`Children` key becomes a directory, each `Page` its `index.html`. How nodes,
children and copying work is on [The Node tree](../the-node-tree/).

<details>
<summary>What exactly does Build do?</summary>

`Build` walks the tree depth-first: for each node it makes the directory,
renders `Page` into `index.html`, renders each `Files` entry, copies
`CopyFrom`, compiles `CompileFrom`, merges `Generate` into `Children`, and
recurses. That is the entire mechanism — nothing else happens between your
tree and the output directory.

</details>

Open `build/index.html` in a browser, or serve the folder with any static file
server.

## Next

- [The Node tree](../the-node-tree/) — nodes, children, copying, `Build`.
- [Ctx & URLs](../ctx/) — what pages receive, and how to build URLs.
- [Content & collections](../content-and-collections/) — markdown with front matter.
- [Locales & i18n](../i18n/) — one tree, many languages.
- [Dev server](../dev-server/) — watch, rebuild, reload while you work.
