+++
Title = "Comparison"
Description = "How astwerk differs from Hugo, Zola, Eleventy and Astro — and when one of those is the better choice."
Order = 30
Group = "Introduction"
+++

This is a static-site generator among many. The honest comparison is against
the ones people reach for first, on the axes that actually change a build: how
content is described, what the template language is, and how much the tool
wants to own your project.

## The short version

Choose astwerk when the team already writes Go, the templates should be
type-checked at compile time, and the site's structure should live in code
rather than a directory convention. Choose Hugo, Zola, Eleventy or Astro when
you want a theme ecosystem, a component framework for islands, or builds at a
scale where incremental compilation matters.

## At a glance

| | astwerk | Hugo | Zola | Eleventy | Astro |
| --- | --- | --- | --- | --- | --- |
| Language | Go | Go | Rust | JS/Node | JS/Node |
| Templates | templ (compiled Go) | Go `html/template` | Tera | Nunjucks/Liquid/JS | `.astro` (JSX-like) |
| Site structure | a `Node` tree you write | convention: `content/` + `layouts/` | convention: `content/` + `templates/` | convention + config | file-based routing in `src/pages` |
| Content discovery | explicit, via `Generate` | directory scan | directory scan | directory scan | directory scan |
| Client JS | none built in; scripts are opt-in | none built in | none built in | none built in | islands, any framework |
| Config format | none — Go struct literals | TOML/YAML/JSON | TOML | JS/JSON | JS/TS |
| Toolchain | `go` + `templ` CLI | single binary | single binary | Node + npm | Node + npm |

## Where astwerk is a good fit

**You already write Go, and want the type checker on your templates.** templ
components are Go functions — a typo in a variable name is a compile error, not
a blank spot in production. Hugo's `html/template` is Go too, but its templates
are strings interpreted at runtime; a bad field access there fails at build
time at best, in the browser at worst.

**You want the site's structure to be code, not a directory convention the
generator infers.** Hugo, Zola and Eleventy all work by scanning `content/`
and matching it against a layout by naming convention. That's fast to start
with and awkward to deviate from — an unusual routing need means fighting the
convention. A `Node` tree has no convention to fight; it's a Go value, so
anything expressible in Go is expressible in the tree, including a `Generate`
that computes children from an API response or a database query at build time.

**You want static output without a client runtime attached.** Core astwerk
ships no JavaScript. A page with no script is HTML and CSS, the same as a
hand-written site. If a page needs behaviour, add a normal `<script>` tag, a
TypeScript file that `CompileScripts` hands to `tsc`, or a wasm module built by
`CompileFrom`. Nothing is injected into pages that didn't ask for it.

### astwerk vs Astro

Astro is the closest philosophical peer — HTML-first, zero-JS by default — so
the contrast is the sharpest:

| | Astro | astwerk |
| --- | --- | --- |
| Default output | HTML-first | HTML-first |
| Islands | JS islands (React, Vue, Svelte…) | none in core; scripts and wasm are opt-in build inputs |
| Toolchain | Node ecosystem | Go ecosystem |
| Templates | `.astro` (JSX-like) | templ (compiled Go, type-checked) |
| Site structure | file-based routing in `src/pages` | a `Node` tree you write |

Choose Astro when you want component-level islands across multiple UI
frameworks in one project. Choose astwerk when the team already writes Go and
the build should stay one language and one toolchain.

Astro's island model is genuinely better when a site is mostly interactive and
the interactive parts are React/Vue/Svelte components you already have. Core
astwerk does not try to be that. Its optional client-side extensions
([astwerk-live](https://github.com/LukasDerBaum42/astwerk-live)) are separate
packages, still in heavy redesign, and not part of the core promise.

## The deliberate absences

astwerk doesn't have these things, and that's a decision, not a gap:

- **No config file** — Go struct literals are the configuration. There is no
  format to learn and no schema to look up.
- **No conventions** — nothing is discovered. You write the tree, so there is
  no framework assumption to fight when a site deviates.
- **No plugins** — extensibility is Go. A custom generator is a function, not
  an interface to implement and a registry to register in.
- **No themes** — there is a folder of files you copy and edit. No marketplace,
  no inheritance, no update breaking your customizations.
- **No client runtime** — core outputs static HTML. Interactivity is a script
  or a wasm module you choose to add, not a framework the generator injects.
- **No Node.js** — the toolchain is `go` plus the templ CLI, and that's the
  whole toolchain.

## Where something else is the better choice

**You want a plugin ecosystem and a theme marketplace.** Hugo and Eleventy both
have large ecosystems of drop-in themes and shortcodes. astwerk's `starter/` is
thirteen files you copy and edit — there is no marketplace, and there is not
meant to be one. If "install a theme, change the logo" is the whole
requirement, Hugo will get there faster.

**You want component-level islands across multiple UI frameworks in one
project.** That's Astro's specific strength — mixing React, Svelte and Vue
islands on one page — and it isn't something core astwerk tries to do. The
optional astwerk-live packages are a small, separate client story, not a place
to plug in an existing component framework — and they are still being
redesigned.

**You need incremental/parallel builds for a very large site (10k+ pages).**
Hugo in particular is tuned hard for this. astwerk's walker is a straightforward
depth-first recursion with no incremental cache — every build is a full build.
For most project sites and docs sites that's unnoticeable; for a large content
site it's a real gap (tracked as a possible improvement, not solved today).
