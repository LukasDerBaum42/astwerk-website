package examples

// Source snippets for the example pages. They live in a plain .go file rather
// than a .templ one because templ rewrites raw strings that contain templ
// syntax into their generated Go — a snippet is meant to show the source, not
// the output of the compiler.

const treeSrc = `root := ssg.Node{
	Title: "My Site",
	Page:  ssg.Templ(views.Home),
	Children: map[string]ssg.Node{
		"about":   {Title: "About", Page: ssg.Templ(views.About)},
		"style":   {CopyFrom: "style"},
		"scripts": {CompileFrom: "scripts"},
	},
}

if err := ssg.Build(root, ssg.BuildOptions{
	RelativeURLs: true, // so the same build works under a subpath
	BaseURL:      "/my-site",
}); err != nil {
	log.Fatal(err)
}`

const collectionSrc = `func projects(dir string) ssg.Node {
	pages, err := content.LoadCollection[FrontMatter](dir)
	if err != nil {
		log.Fatal(err)
	}

	var tiles []templ.Component
	children := map[string]ssg.Node{}

	for _, it := range pages.Items {
		it := it // capture the item for the page closure
		children[it.Slug] = ssg.Node{
			Title: it.Meta.Title,
			Page: func(c ssg.Ctx) templ.Component {
				return views.ProjectPage(c.Title, it.Content(c))
			},
		}
		tiles = append(tiles, views.Tile(it.Meta.Title, it.Meta.Description, it.Slug))
	}

	return ssg.Node{
		Title:    "Projects",
		Page:     func(c ssg.Ctx) templ.Component { return views.ProjectIndex(c.Title, tiles) },
		Children: children,
	}
}

// "projects": projects("content/projects"),`

const localeSrc = `root = ssg.BuildLocales(root, []ssg.Locale{{
	Code:   "de",
	Prefix: "/de", // defaults to "/" + Code
	Override: map[string]func(ssg.Node) ssg.Node{
		"": func(n ssg.Node) ssg.Node {
			n.Title, n.Page = "Meine Seite", ssg.Templ(views_de.Home)
			return n
		},
		"about": func(n ssg.Node) ssg.Node {
			n.Title = "Über mich" // Page and Children survive untouched
			return n
		},
	},
}})

// An override is handed the node BuildLocales would otherwise have derived, so
// changing one field is one line. Everything not named appears under /de/
// rendered by the same components, with Ctx.Prefix set to "/de".`

const scriptsSrc = `//go:build js && wasm

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

// "scripts": {CompileFrom: "scripts"}  →  build/scripts/nav.wasm + wasm_exec.js`

const wasmLoaderSrc = `templ WasmScript(c ssg.Ctx, name string) {
	<script src={ c.Asset("scripts/wasm_exec.js") }></script>
	<script data-wasm={ c.Asset("scripts/" + name + ".wasm") }>
		(() => {
			const src = document.currentScript.dataset.wasm;
			const go = new Go();
			WebAssembly.instantiateStreaming(fetch(src), go.importObject)
				.then(res => go.run(res.instance));
		})();
	</script>
}`

const templBasicsComponentsSrc = `// A component is a function; parameters are plain Go.
templ Greeting(name string) {
	<p>Hello, { name }!</p>
}`

const templBasicsEscapingSrc = `// Expressions are HTML-escaped: & < > " ' all come out escaped,
// so a user-provided name cannot inject markup.
templ Profile(user User) {
	<h1>{ user.Name }</h1>
	<p>{ user.Bio }</p>
}`

const templBasicsChildrenSrc = `// Components nest. A component that accepts children declares them
// with { children... } — callers wrap it instead of passing content in.
templ Layout(title string) {
	<main>
		<h1>{ title }</h1>
		{ children... }
	</main>
}

templ About() {
	@Layout("About") {
		<p>This paragraph is the children.</p>
	}
}`

const templBasicsFlowSrc = `// Conditionals and loops are Go, inline in the markup.
templ List(items []Item) {
	if len(items) == 0 {
		<p>Nothing here.</p>
	} else {
		<ul>
			for _, it := range items {
				<li>{ it.Label }</li>
			}
		</ul>
	}
}`

const templBasicsAttrsSrc = `// Attributes: a conditional attribute, a URL helper, and a CSS class
// helper — templ owns the escaping rules for each context.
templ Menu(current string) {
	<button disabled?={ current == "" }>Submit</button>
	<a href={ templ.URL("/docs/") }>Docs</a>
	<a class={ templ.Class("active", current == "about") }>About</a>
}

// Trusted HTML is opt-in, never the default.
templ Article(htmlBody string) {
	@templ.Raw(htmlBody)
}`
