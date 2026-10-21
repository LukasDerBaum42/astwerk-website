+++
Title = "Scripts & WebAssembly"
Description = "Compiling Go to wasm as part of the build, with no JS toolchain."
Order = 75
Group = "Build a site"
+++

A node with `CompileFrom` compiles a scripts directory into the output. No
bundler, no `package.json` — a `.go` file becomes a `.wasm` module, a `.ts` file
is handed to `tsc`.

## Documentation

### CompileFrom

```go
"scripts": {CompileFrom: "scripts"},
```

The field wires a directory into the build. It runs `ssg.CompileScripts(srcDir,
outDir)`, which looks at each top-level entry:

| Entry | Output |
| --- | --- |
| a `.go` file with `//go:build js && wasm` | `<name>.wasm` |
| a subdirectory containing at least one such file | `<dirname>.wasm` |
| `.ts` files | handed to `tsc` in one batch |

Anything else is ignored, so helper packages and non-wasm Go files can live in
the same directory.

<details>
<summary>How the build runs</summary>

Compilation runs from the entry's parent directory, so a script can import
your project's own packages. When at least one `.wasm` is produced, Go's own
`wasm_exec.js` glue is copied in beside it. Each `.wasm` is built by its own
`go build` process in parallel, and `.ts` files are handed to `tsc` in one
batch.

</details>

### The build constraint

```go
//go:build js && wasm

package main

func main() {
	// …
	select {} // keep the module alive
}
```

The constraint is required — without it, `go build ./...` at your project root
would also try to compile the script for the host platform, where `syscall/js`
doesn't exist.

### Loading a module

```templ
templ WasmScript(c ssg.Ctx, name string) {
	<script src={ c.Asset("scripts/wasm_exec.js") }></script>
	<script data-wasm={ c.Asset("scripts/" + name + ".wasm") }>
		(() => {
			const src = document.currentScript.dataset.wasm;
			const go = new Go();
			WebAssembly.instantiateStreaming(fetch(src), go.importObject)
				.then(res => go.run(res.instance));
		})();
	</script>
}
```

Two details in there are load-bearing: the path travels in a `data-` attribute,
because templ treats `<script>` contents as raw text and would emit an
interpolated expression literally; and `document.currentScript` must be read
synchronously, because it's `null` once a promise resolves. `wasm_exec.js`
defines `globalThis.Go`, so it has to load first — a plain script tag, not a
module import.

`starter/wasm/wasm_script.templ` ships this component ready to copy.

<details>
<summary>Size</summary>

A Go wasm binary starts around 2 MB. That's the Go runtime, and it compresses
to roughly a quarter of that over the wire. If a page only needs a nav toggle,
ten lines of hand-written JavaScript is genuinely the better engineering
choice. Reach for wasm when the logic is substantial enough that writing it in
Go pays for the runtime.

</details>

## Example

```go
//go:build js && wasm

package main

import "syscall/js"

func main() {
	doc := js.Global().Get("document")
	doc.Call("querySelector", "#toggle").Call("addEventListener", "click",
		js.FuncOf(func(js.Value, []js.Value) any {
			doc.Call("querySelector", "#nav").Get("classList").Call("toggle", "open")
			return nil
		}))
	select {} // keep the module alive for the handler
}
```

Dropped in `scripts/` and wired with `CompileFrom`, this builds to
`scripts/nav.wasm` and is loaded by the component above.

Client-side helpers that make this less verbose — `x`, `reactive`, `wasmwrap`,
`jsdom` — live in the optional
[astwerk-live](https://github.com/LukasDerBaum42/astwerk-live) extension. They
are not part of core astwerk and are still in heavy redesign.
