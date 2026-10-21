+++
Title = "Comparison"
Description = "How astwerk differs from Hugo, Zola, Eleventy, Astro and Svelte — and when one of those is the better choice."
Order = 30
Group = "Introduction"
+++

This is a static-site generator among many. The honest comparison is against
the ones people reach for first, on the axes that actually change a build: how
content is described, what the template language is, and what client-side
interactivity costs.

## The short version

Choose astwerk when the team already writes Go, the templates should be
type-checked at compile time, the site's structure should live in code rather
than a directory convention, and interactivity shouldn't pull in a JavaScript
toolchain. Choose Hugo, Zola, Eleventy or Astro when you want a theme
ecosystem, a component framework for islands, or builds at a scale where
incremental compilation matters. Choose Svelte when you want compile-time
reactivity — the idea astwerk's `x` is built on — but the component layer in
JavaScript/TypeScript.

## At a glance

| | astwerk | Hugo | Zola | Eleventy | Astro | Svelte |
| --- | --- | --- | --- | --- | --- | --- |
| Language | Go | Go | Rust | JS/Node | JS/Node | JS/TS |
| Templates | templ (compiled Go) | Go `html/template` | Tera | Nunjucks/Liquid/JS | `.astro` (JSX-like) | `.svelte` (compiled) |
| Site structure | a `Node` tree you write | convention: `content/` + `layouts/` | convention: `content/` + `templates/` | convention + config | file-based routing in `src/pages` | SvelteKit routes in `src/routes` |
| Content discovery | explicit, via `Generate` | directory scan | directory scan | directory scan | directory scan | SvelteKit load functions |
| Client JS | Go → WASM, opt-in | none built in | none built in | none built in | islands, any framework | compiled at build time, no runtime framework |
| Config format | none — Go struct literals | TOML/YAML/JSON | TOML | JS/JSON | JS/TS | `svelte.config.js` |
| Toolchain | `go` + `templ` CLI | single binary | single binary | Node + npm | Node + npm | Node + npm/Vite |

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

**You want interactivity without adopting a JS framework.** Astro's islands
are the closest comparison — ship JS only where a component needs it — but the
island itself is still React, Vue, Svelte or similar, with that framework's own
build step. astwerk's interactive paths are [`x`](../x/) — compiled markup,
plain JavaScript at build time, no framework at all — and `reactive`, Go
compiled to WASM: one language, one toolchain, no `npm install` divergence
between the static and interactive halves of a page.

### astwerk vs Astro

Astro is the closest philosophical peer — HTML-first, islands, zero-JS by
default — so the contrast is the sharpest:

| | Astro | astwerk |
| --- | --- | --- |
| Default output | HTML-first | HTML-first |
| Islands | JS islands (React, Vue, Svelte…) | compiled Go/JS islands (`x`), or Go/WASM (`reactive`) |
| Toolchain | Node ecosystem | Go ecosystem |
| Templates | `.astro` (JSX-like) | templ (compiled Go, type-checked) |
| Site structure | file-based routing in `src/pages` | a `Node` tree you write |

Choose Astro when you want component-level islands across multiple UI
frameworks in one project. Choose astwerk when the team already writes Go and
the interactive half of the page should not bring a JS framework's build step
with it.

### astwerk vs Svelte

Svelte is the closest relative of astwerk's [`x`](../x/) package — and its
inspiration. Both decided the same thing: reactivity is a compile-time problem,
so ship the compiler's output instead of a framework. The difference is where
the compiler lives and what it compiles.

| | Svelte | astwerk `x` |
| --- | --- | --- |
| Reactivity | compiled at build time | compiled at build time |
| Runtime | a small runtime, no virtual DOM | per-page fragments, no virtual DOM |
| Components | `.svelte` files, a component syntax | ordinary Go functions in templ |
| Language | JavaScript/TypeScript | Go |
| Toolchain | Node + Vite/SvelteKit | `go` + the templ CLI |

Svelte compiles a component file into vanilla JavaScript; `x` compiles the
markers a page actually uses into a small script — assembled per page from
fragments, and only when a page opts in with `x.Document`. Both are answers to
"why ship a framework when the compiler can do the work?" The difference is
that astwerk keeps the whole stack in one language: no `.svelte` file format,
no second ecosystem, no Node toolchain joining the build. Choose Svelte when
the component layer should live in JavaScript/TypeScript with its own
ecosystem; choose astwerk when it should be Go, inside your existing templ
components.

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
- **No virtual DOM** — bindings patch one property at a time, so nothing is
  ever destroyed and rebuilt behind your back.
- **No Node.js** — the toolchain is `go` plus the templ CLI, and that's the
  whole toolchain.

## Where something else is the better choice

**You want a plugin ecosystem and a theme marketplace.** Hugo and Eleventy both
have large ecosystems of drop-in themes and shortcodes. astwerk's `starter/` is
seventeen files you copy and edit — there is no marketplace, and there is not
meant to be one. If "install a theme, change the logo" is the whole
requirement, Hugo will get there faster.

**You want Go on the client and nothing else will do it as cheaply.** This one
needs a correction, not a caveat: astwerk does not require WASM. A page with no
`CompileFrom` node ships exactly as much client JS as a Hugo or Zola page — zero
— and `CompileScripts` hands `.ts` files to `tsc` in the same build step, so a
script written in plain JavaScript or TypeScript costs nothing beyond what it
would cost on any other generator. The ~2 MB WASM floor is real, but it's the
price of one specific choice — writing that script's logic in Go and running
`reactive` or `wasmwrap` — not a tax astwerk levies on every page. Reach for it
when the interactive logic is substantial enough that sharing a language with
the rest of the site pays for the runtime; reach for a `<script>` tag or a `.ts`
file, same as anywhere else, when it isn't.

**The team is already fluent in JS/TS and wants component-level islands across
multiple UI frameworks in one project.** That's Astro's specific strength —
mixing React, Svelte and Vue islands on one page — and it isn't something
astwerk is trying to do. `reactive` is one small, fine-grained state library,
not a place to plug in an existing component framework.

**You want compile-time reactivity but the component layer in JavaScript.**
That is Svelte's exact pitch — and the idea astwerk's `x` was built on. If the
team is JS/TS-first and wants the component ecosystem around it (SvelteKit,
stores, transitions), Svelte is the better choice; if the components should be
Go functions inside templ, astwerk keeps the whole stack in one language —
see [astwerk vs Svelte](#astwerk-vs-svelte) above.

**You need incremental/parallel builds for a very large site (10k+ pages).**
Hugo in particular is tuned hard for this. astwerk's walker is a straightforward
depth-first recursion with no incremental cache — every build is a full build.
For most project sites and docs sites that's unnoticeable; for a large content
site it's a real gap (tracked as a possible improvement, not solved today).
