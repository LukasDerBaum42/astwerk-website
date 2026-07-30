# astwerk-website

The [astwerk](https://github.com/LukasDerBaum42/astwerk) documentation site,
built with astwerk. Deployed to
<https://lukasderbaum42.github.io/astwerk-website/>.

It lives in its own repository so the library stays a library: the site's
dependencies — chroma, the goldmark extensions — never reach anyone who
`go get`s astwerk, and the library's module has nothing in it but the library.

It depends on a **published tag** of astwerk, like any other consumer would.
That's the point: if the site builds, the released API is usable as released.

## Building

```bash
templ generate          # or: go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
go run .                # into build/, base path /astwerk-website
go run . -serve         # build, serve on :8080, rebuild and reload on change
go run . -dev           # keep existing output, overwrite in place
go run . -base ""       # for a domain root instead of /astwerk-website
```

`-serve` is the one to leave running while writing docs — it watches the
sources, rebuilds, and reloads the browser.

## Building against unreleased astwerk

To try a library change before tagging it, use a Go workspace rather than
editing `go.mod` — `go.work` is gitignored, so it can't be committed by
accident:

```bash
go work init . ../astwerk
go run .
rm go.work            # back to the published tag
```

A workspace overrides where the module is *read from*, but not whether its
required version exists: Go still resolves the `require` line against the
registry. So the version in `go.mod` has to be a tag that has been pushed, even
when a workspace is redirecting the code to a sibling checkout.

Bumping to a new release is the ordinary thing:

```bash
go get github.com/LukasDerBaum42/astwerk@v0.6.0
go mod tidy
```

## Previewing

`go run . -serve` is the normal way. Every in-page URL is relative
(`RelativeURLs: true`), so the output is portable and anything else works too:

```bash
cd build && python3 -m http.server 8000   # http://localhost:8000/
```

Opening `build/index.html` straight off disk works. So does serving it under
`/astwerk-website/`. Same build, no flags.

`BaseURL` is still set, but only feeds `Ctx.URL` for canonical tags and Open
Graph, which have to be absolute.

## Layout

Following the same split as the original site: `templates/` for the shared shell,
and page bodies in their own package.

| | |
| --- | --- |
| `main.go` | the Node tree, markdown config, docs loading |
| `templates/` | the shell — `Layout`, `CodeBlock`, `WasmScript`, `NavItem` |
| `pages/` | page bodies — `Home`, `DocPage`, `DocsIndex`, `Examples`, `NotFound` |
| `content/docs/*.md` | the documentation, ordered by `Order` in front matter |
| `scripts/demo.go` | the live demos, compiled to wasm by `CompileFrom` |
| `style/style.css` | hand-written, no preprocessor |
| `style/code.css` | **generated** — see below |
| `tools/gencss` | regenerates `style/code.css` from chroma |
| `site_test.go` | builds the site and validates every link resolves |

`pages` imports `templates`; nothing goes the other way.

## Regenerating the syntax highlighting CSS

```bash
go run tools/gencss/main.go > style/code.css
```

The highlighter emits classes rather than inline styles, so one stylesheet covers
both markdown code blocks and templ-embedded snippets. Generating it means the
colours come from the same chroma styles the formatter uses instead of being
copied by hand.

## Two things that will bite

**Don't name a templ parameter `ctx`.** templ's generated code declares its own
`ctx context.Context`; a parameter with that name collides and the generated file
won't compile. Every component here takes `c ssg.Ctx`.

**Don't write absolute URLs as literals.** Use `c.Asset` for assets and `c.Link`
for pages. `site_test.go` walks every generated page and resolves each link
against the directory it came from, so an absolute or broken URL fails the build.

## Adding a page

Drop a markdown file in `content/docs/` with front matter:

```markdown
+++
Title = "My page"
Description = "One line for the index and the sidebar."
Order = 45
+++
```

`Order` decides the sidebar and pager sequence. Nothing else to wire up — the
tree generates the node, the sidebar, the index entry and the prev/next links.
