+++
Title = "Dev server"
Description = "Watch, rebuild, reload — the loop you leave running while you work."
Order = 70
Group = "Build a site"
+++

`go run .` builds the site once. While you work you want that on a loop — edit a
file, see the page. That's `devserver`: it serves the build directory, watches
your sources, rebuilds on change, and reloads the browser. Stdlib Go only.

## Documentation

### Run

```go
import "github.com/LukasDerBaum42/astwerk/devserver"

devserver.Run(context.Background(), devserver.Config{
	Build: devserver.Steps(
		devserver.Command("templ", "generate"),
		devserver.Command("go", "run", ".", "--dev"),
	),
})
```

`Run` builds once at startup, serves `Dir`, and rebuilds whenever a watched
file changes. Only `Config.Build` is required.

### Build must be a subprocess

The build is two commands, not a Go function — and that is the most important
thing on this page. A running Go program cannot load new code, so an in-process
build would keep re-rendering the components compiled into the binary that
started the server: saving a `.templ` would reload the browser and change
nothing, with no error. A subprocess recompiles.

`Command` treats a non-zero exit as a failed build and puts the command's output
in the error, so a compiler message lands in the browser overlay, file and line
included. `Steps` runs its steps in order and stops at the first failure.

Pass `--dev` to the inner build: a normal build deletes the output directory
first, so a browser fetching an asset mid-rebuild gets a 404 for a file that
exists a moment later.

<details>
<summary>When an in-process build is fine</summary>

A closure is right only when the build reads nothing but data — markdown, CSS,
images. The rule: anything that involves compiled Go code must shell out.

</details>

### Config

Only `Build` is required.

| Field | Default | Meaning |
| --- | --- | --- |
| `Build` | — | rebuilds the site; called once at startup and on every change |
| `Dir` | `"build"` | the directory to serve |
| `Addr` | `"localhost:8080"` | listen address; `"localhost:0"` picks a free port |
| `Watch` | `["."]` | directories watched for changes |
| `Exts` | `devserver.DefaultExts` | extensions that trigger a rebuild |
| `Ignore` | `devserver.DefaultIgnore` | directory names never descended into |
| `Interval` | `300ms` | how often the watcher polls |
| `Log` | stderr | where build and reload messages go |
| `NoColor` | `false` | turn off ANSI colour (automatic off a terminal, and on `NO_COLOR`) |
| `Ready` | — | called with the bound address once listening; mainly for tests |

`Dir` is added to the ignore list automatically, or the build's own output would
trigger the next build forever. `DefaultExts` covers everything that can change
the output (`.go`, `.templ`, `.md`, `.html`, `.css`, `.js`, `.ts`, `.toml`,
`.json`, `.yaml`, `.yml`, `.svg`); `DefaultIgnore` is `.git`, `node_modules`,
`.cache`, `vendor`.

### Watching

The watcher polls file size and modification time every `Interval` rather than
using OS file events. That keeps astwerk dependency-free and immune to editors
that save by renaming a temp file. A burst of changes is followed by one quiet
tick, so saving five files at once is one rebuild, not five.

<details>
<summary>What it does that a plain file server doesn't</summary>

- Serves the site the way a static host will: directory URLs get a trailing
  slash redirect (otherwise relative asset URLs resolve one level too high),
  and a missing path gets your own `404.html` with a 404 status.
- Injects a reload script into every HTML response — nothing is written into
  your build output, so `build/` stays exactly what you'd deploy.
- Shows a failed build as a full-screen overlay instead of a stale page, and
  clears it on the next successful build.

`devserver` imports nothing from `ssg`. You hand it a `func() error`; it calls
it. A project whose build does more than call `ssg.Build` — generating a CSS
file, fetching data, running `templ generate` — gets exactly the same loop.

</details>

## Example

```go
func main() {
	if !*serve {
		if err := ssg.Build(tree(), ssg.BuildOptions{Dev: *dev}); err != nil {
			log.Fatal(err)
		}
		return
	}

	devserver.Run(context.Background(), devserver.Config{
		Build: devserver.Steps(
			devserver.Command("templ", "generate"),
			devserver.Command("go", "run", ".", "--dev"),
		),
	})
}
```

```sh
go run . --serve
```
