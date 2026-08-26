+++
Title = "Dev server"
Description = "Watch, rebuild, reload — the loop you leave running while you work."
Order = 25
Group = "Start here"
+++

`go run .` builds the site once. While you're actually working on it you want
that on a loop: edit a file, see the page. That's `devserver`.

It serves the build directory, watches your sources, rebuilds when they change,
and reloads the browser. No Node, no extra tool, no dependency — it's stdlib Go
and a few dozen lines of vendored JavaScript.

## Wiring it up

```go
import "github.com/LukasDerBaum42/astwerk/devserver"

func main() {
	serve := flag.Bool("serve", false, "serve and rebuild on change")
	dev := flag.Bool("dev", false, "overwrite the output in place")
	flag.Parse()

	if !*serve {
		// The one-shot build — and what --serve shells out to.
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

```
go run . --serve
2026/07/30 23:28:03 built build/
2026/07/30 23:28:03 serving build/ on http://127.0.0.1:8080
2026/07/30 23:28:12 changed: templates/layout.templ, rebuilding
2026/07/30 23:28:15 rebuilt in 3.701s, reloading
```

## Build must run out of process

Look at what `Build` is above: two commands, not a Go function that renders
your site. That is the most important thing on this page.

**A running Go program cannot load new code.** Your pages are compiled templ
components. If `Build` calls your render functions directly, it re-renders the
components that were compiled into the binary that started the dev server —
the ones from before your edit. So you save a `.templ` file, the watcher fires,
the build "succeeds", the browser reloads, and the page is identical.

That is worse than nothing happening, because everything about it looks like it
worked. There is no error, and the log cheerfully says `rebuilt`.

`devserver.Command` and `devserver.Steps` are the fix. They run the build as a
subprocess, which recompiles, which is what makes the change real:

```go
func devserver.Command(name string, args ...string) func() error
func devserver.Steps(steps ...func() error) func() error
```

`Command` treats a non-zero exit as a failed build and puts the command's
combined output in the error — so a compiler message lands in the browser
overlay, file and line included, instead of only in your terminal.

An in-process closure is only correct if everything the build reads is *data* —
markdown, CSS, images — and never compiled Go. It's faster, and it becomes a
trap the moment a `.templ` file is involved.

Two more things worth doing deliberately:

**Pass `--dev` to the inner build.** A normal build removes the output
directory first, so a browser fetching an asset mid-rebuild gets a 404 for a
file that exists a moment later. Dev mode overwrites in place instead.

**Read your content inside the build, not at startup.** The inner build is a
fresh process each time, so this comes free — but it's the same reason an
in-process build has to load its content per call rather than once.

## Config

Only `Build` is required.

| Field | Default | Meaning |
| --- | --- | --- |
| `Build` | — | rebuilds the site; called once at startup and on every change. Use `Command`/`Steps` |
| `Dir` | `"build"` | the directory to serve |
| `Addr` | `"localhost:8080"` | listen address; `"localhost:0"` picks a free port |
| `Watch` | `["."]` | directories watched for changes |
| `Exts` | `devserver.DefaultExts` | extensions that trigger a rebuild |
| `Ignore` | `devserver.DefaultIgnore` | directory names never descended into |
| `Interval` | `300ms` | how often the watcher polls |
| `Log` | `log.Default()` | where build and reload messages go |

`Dir` is added to the ignore list automatically. Without that the build's own
output would trigger the next build, forever.

## What it does that a plain file server doesn't

**Serves the site the way a static host will.** A directory means its
`index.html`; a directory URL without a trailing slash redirects (otherwise the
page's relative asset URLs resolve one level too high); a missing path gets your
own `404.html` with a 404 status, not a stock Go error page. So the 404 page you
wrote is the 404 page you see.

**Injects the reload script into every HTML response** — before `</body>`, or at
the end if there isn't one. Nothing is written into your build output; the
injection happens on the way out, so `build/` stays exactly what you'd deploy.

**Shows you a failed build instead of a stale page.** When `Build` returns an
error the server keeps running — losing your reload connections over a typo
would be a bad trade — logs it, and pushes it to the browser as a full-screen
overlay. The next successful build clears it and reloads. A browser that
connects while the build is broken is told immediately, rather than sitting on
an old page with no explanation.

## Why polling instead of file-system events

The watcher compares a snapshot of each file's size and modification time every
`Interval`. That's a deliberate trade, twice over.

`fsnotify` would be instant, and it would be astwerk's first dependency outside
templ, goldmark and toml — in a library whose whole pitch is that a Go toolchain
is all you need. It also needs a debounce and a polling fallback anyway, because
plenty of editors save by writing a temp file and renaming it.

And within polling: stamps rather than content hashes. A false positive costs
one wasted rebuild; reading every watched file on every tick to avoid it is much
the worse deal.

The cost is a sub-second delay before a rebuild starts. Turn `Interval` down if
you want it tighter.

A change is followed by one quiet tick before the build runs, so an editor
saving five files at once is one rebuild, not five.

## It doesn't know about `ssg`

`devserver` imports nothing from `ssg`. You hand it a `func() error`; it calls
it. That keeps it a general watch-rebuild-reload loop rather than a second entry
point into the walker — and it means a project whose build does more than call
`ssg.Build` (generating a CSS file, fetching data, running `templ generate`)
gets exactly the same loop with no extra plumbing.
