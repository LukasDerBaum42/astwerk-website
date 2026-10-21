+++
Title = "Ctx & URLs"
Description = "What the walker tells a page — paths, locale, base — and how to build URLs from it."
Order = 60
Group = "Build a site"
+++

Every page is a function of a `Ctx` describing where the page lives. URL
building goes through `Ctx` too: `Asset` and `Link` are the only correct way to
reference the site's own files. The tree these pages hang off is on
[The Node tree](../the-node-tree/).

## Documentation

### Ctx

```go
type Ctx struct {
	Title    string // the node's Title
	Path     string // "about/", "de/projects/thing/"
	Prefix   string // "" or "/de" — the locale's URL prefix
	Locale   string // "" or "de"
	Base     string // "" or "/astwerk" — the site's base path
	Relative bool   // RelativeURLs is on — Asset/Link emit relative URLs
}

func (c Ctx) URL() string               // this page's absolute URL
func (c Ctx) Asset(path string) string  // base-aware asset URL
func (c Ctx) Link(path string) string   // base- and locale-aware link
func (c Ctx) Rel() string               // Path without the locale segment
func (c Ctx) InLocale(prefix string) string // this page's URL in another locale
```

The walker already knows where a page lives, so it tells the page instead of
making you retype it. `Title` comes from the node; `Path`, `Prefix` and
`Locale` are filled in by the walker — see [Locales & i18n](../i18n/) for what
the locale fields mean.

### Asset and Link

```templ
templ Layout(c ssg.Ctx) {
	<link rel="stylesheet" href={ c.Asset("style/style.css") }/>
	<a href={ c.Link("docs/") }>Docs</a>
}
```

`Asset` resolves a path relative to the site root; `Link` does the same and
applies the locale prefix. Never write a URL as a literal: `/style/style.css`
breaks the moment the site is served from a subdirectory.

Name the parameter `c`, not `ctx`: templ's generated code declares its own
`ctx`, and a parameter with that name won't compile.

<details>
<summary>Why a literal breaks under a subdirectory</summary>

Every GitHub project Pages site is served from `user.github.io/repo/` — a
subdirectory of the domain. A literal `/style/style.css` resolves against the
domain root, which is a *different site*, and there is no URL rewriting to save
you. `Asset` and `Link` embed the base path (or emit relative URLs, below) and
keep working wherever the output is served.

</details>

### Relative URLs

For portable output, build with relative URLs:

```go
ssg.Build(root, ssg.BuildOptions{
	RelativeURLs: true,
	BaseURL:      "/astwerk", // only feeds Ctx.URL, for canonical tags
})
```

One build then works opened from disk, at a localhost root, and deployed under
`/astwerk/` — no rebuild, no flag.

<details>
<summary>Why Ctx.URL stays absolute</summary>

`Ctx.URL` is unaffected by `RelativeURLs` and always carries the base path,
because canonical links and Open Graph tags have to be absolute. In-page links
and assets are the ones that get the relative treatment.

</details>

## Example

```templ
templ Layout(c ssg.Ctx) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<link rel="canonical" href={ c.URL() }/>
			<link rel="stylesheet" href={ c.Asset("style/style.css") }/>
		</head>
		<body>
			{ children... }
			<a href={ c.Link("about/") }>About</a>
		</body>
	</html>
}
```

The three URL builders cover every case: `URL` for anything that must be
absolute, `Asset` for files, `Link` for pages.
