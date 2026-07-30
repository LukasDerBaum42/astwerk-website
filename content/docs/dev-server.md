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
	flag.Parse()

	if *serve {
		devserver.Run(context.Background(), devserver.Config{
			Build: func() error { return build(true) },
		})
		return
	}
	if err := build(false); err != nil {
		log.Fatal(err)
	}
}

func build(dev bool) error {
	return ssg.Build(tree(), ssg.BuildOptions{Dev: dev})
}
```

```
go run . --serve
2026/07/30 21:19:52 built build/
2026/07/30 21:19:52 serving build/ on http://127.0.0.1:8080
2026/07/30 21:20:24 changed: content/docs/getting-started.md, rebuilding
2026/07/30 21:20:27 rebuilt in 3.009s, reloading
```

Two things about that `Build` function are worth doing deliberately.

**Pass `Dev: true`.** A normal build removes the output directory first. A
browser fetching an asset mid-rebuild then gets a 404 for a file that exists a
moment later. Dev mode overwrites in place instead.

**Build the tree inside the function, not outside it.** `Build` is called again
on every change, so anything read at startup — a content directory, a config
file — has to be read inside it or your edits won't show up. If your build
already looks like `buildSite()`, you're fine.

## Config

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
