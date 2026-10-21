package examples

// Source snippets for the remaining example pages. Like examples_x_src.go,
// these live in a plain .go file: templ rewrites raw strings that contain
// templ syntax into their generated Go, and a snippet is meant to show the
// source.

const xBindingsStateSrc = `var BURL     = x.NewStrSignal("https://templ.guide")
var BActive  = x.NewBoolSignal(true)
var BBusy    = x.NewBoolSignal(false)
var BPressed = x.NamedStrSignal("pressed", "false")
var BColour  = x.NewStrSignal("#4f7cff")`

const xBindingsElementsSrc = `<p>@x.El("a", x.Attr("href", BURL)) { templ.guide }</p>
<p>@x.El("span", x.Class("on", BActive)) { a toggled class }</p>
<p>@x.El("button", x.Disabled(BBusy)) { save }</p>
<p>@x.El("span", x.Style("color", BColour)) { styled text }</p>`

const xBindingsButtonsSrc = `<div class="counter-row">
	@x.El("button", x.On("click", BActive.Toggle())) {
		toggle class
	}
	@x.El("button",
		x.On("click", x.Seq(BBusy.Toggle(), x.If(BBusy, BPressed.SetTo("true"), BPressed.SetTo("false")))),
		x.Class("on", BBusy),
		x.Attr("aria-pressed", BPressed),
	) {
		toggle disabled
	}
	@x.El("button", x.On("click", BColour.SetTo("#ff4f7c"))) {
		recolour
	}
	@x.El("button", x.On("click", BURL.SetTo("https://pkg.go.dev/github.com/LukasDerBaum42/astwerk"))) {
		retarget link
	}
</div>

// The save button's state is visible twice: a Show/Not pair of spans under
// it, and the toggle button's own aria-pressed + .on — the pressed value is a
// named string the click's x.If keeps in step with the busy bool.`

const xBindingsSrc = `var BURL     = x.NewStrSignal("https://templ.guide")
var BActive  = x.NewBoolSignal(true)
var BBusy    = x.NewBoolSignal(false)
var BPressed = x.NamedStrSignal("pressed", "false")
var BColour  = x.NewStrSignal("#4f7cff")

templ XBindingsDemo() {
	<div class="demo">
		<p>@x.El("a", x.Attr("href", BURL)) { templ.guide }</p>
		<p>@x.El("span", x.Class("on", BActive)) { a toggled class }</p>
		<p>@x.El("button", x.Disabled(BBusy)) { save }</p>
		<p class="demo-note">
			@x.El("span", x.Show(x.Computed(x.Not(BBusy)))) {
				the save button is enabled
			}
			@x.El("span", x.Show(BBusy)) {
				the save button is disabled
			}
		</p>
		<p>@x.El("span", x.Style("color", BColour)) { styled text }</p>
		<div class="counter-row">
			@x.El("button", x.On("click", BActive.Toggle())) {
				toggle class
			}
			@x.El("button",
				x.On("click", x.Seq(BBusy.Toggle(), x.If(BBusy, BPressed.SetTo("true"), BPressed.SetTo("false")))),
				x.Class("on", BBusy),
				x.Attr("aria-pressed", BPressed),
			) {
				toggle disabled
			}
			@x.El("button", x.On("click", BColour.SetTo("#ff4f7c"))) {
				recolour
			}
			@x.El("button", x.On("click", BURL.SetTo("https://pkg.go.dev/github.com/LukasDerBaum42/astwerk"))) {
				retarget link
			}
		</div>
	</div>
}

// On the page: @x.Document(XBindingsDemo())`

const hoodStateSrc = `var HoodCount = x.NamedNumSignal("hood", 0)`

const hoodWidgetSrc = `<div class="demo">
	@x.El("button", x.On("click", HoodCount.Inc(1))) {
		@x.Text(HoodCount) +
	}
</div>`

const hoodSrc = `var HoodCount = x.NamedNumSignal("hood", 0)

templ HoodDemo() {
	<div class="demo">
		@x.El("button", x.On("click", HoodCount.Inc(1))) {
			@x.Text(HoodCount) +
		}
	</div>
}

// On the page: @x.Document(HoodDemo())`

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

const templBasicsSrc = `// A component is a function; parameters are plain Go.
templ Greeting(name string) {
	<p>Hello, { name }!</p>
}

// Expressions are HTML-escaped: & < > " ' all come out escaped,
// so a user-provided name cannot inject markup.
templ Profile(user User) {
	<h1>{ user.Name }</h1>
	<p>{ user.Bio }</p>
}

// Components nest. A component that accepts children declares them
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
}

// Conditionals and loops are Go, inline in the markup.
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
}

// Attributes: a conditional attribute, a URL helper, and a CSS class
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
