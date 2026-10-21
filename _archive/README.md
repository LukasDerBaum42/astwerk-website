# Archive

Material kept for later, not part of the current build.

The directory starts with `_`, so the Go tool ignores it and `go build ./...`
does not compile it. `templ generate` is also told to skip it in the Makefile
and CI; if you move something back into the site, move it out of `_archive/`
first.

## tutorial/

The former tutorial (twelve blog steps plus the seven-part cookie clicker) and
its wasm scripts. It is archived while the site is rewritten around core
astwerk (`ssg`, `content`, `assets`, `devserver`, `starter`). Most of it depends
on `astwerk-live` (`x`, `reactive`, `wasmwrap`), so it is not part of the
core-only site build.

Nothing here is deleted. Restore it by moving the files back and re-adding the
site wiring in `main.go` once the tutorial is redesigned.
