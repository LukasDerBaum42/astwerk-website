package examples

// Source snippets for the compiled-markup example pages. They live in a plain
// .go file rather than a .templ one because templ rewrites raw strings that
// contain templ syntax into their generated Go — a snippet is meant to show
// the source, not the output of the compiler.

const xCounterStateSrc = `var XCount  = x.NamedNumSignal("count", 0)
var XDouble = x.Computed(x.Mul(XCount, 2))`

const xCounterButtonsSrc = `<div class="counter-row">
	@x.El("button",
		x.On("click", x.If(x.Computed(x.Gt(XCount, 0)), XCount.Dec(1))),
		x.Disabled(x.Computed(x.Lte(XCount, 0))),
		x.HtmlAttr("aria-label", "Decrement"),
	) {
		&minus;
	}
	@x.El("span", x.HtmlAttr("class", "counter-value")) {
		@x.Text(XCount)
	}
	@x.El("button", x.On("click", XCount.Inc(1))) {
		+
	}
	@x.El("button", x.On("click", XCount.SetTo(0))) {
		reset
	}
</div>`

const xCounterDerivedSrc = `<p class="demo-note">Double: <span>@x.Text(XDouble)</span></p>`

const xCounterSrc = `var XCount  = x.NamedNumSignal("count", 0)
var XDouble = x.Computed(x.Mul(XCount, 2))

templ XCounterDemo() {
	<div class="demo">
		<div class="counter-row">
			@x.El("button",
				x.On("click", x.If(x.Computed(x.Gt(XCount, 0)), XCount.Dec(1))),
				x.Disabled(x.Computed(x.Lte(XCount, 0))),
				x.HtmlAttr("aria-label", "Decrement"),
			) {
				&minus;
			}
			@x.El("span", x.HtmlAttr("class", "counter-value")) {
				@x.Text(XCount)
			}
			@x.El("button", x.On("click", XCount.Inc(1))) {
				+
			}
			@x.El("button", x.On("click", XCount.SetTo(0))) {
				reset
			}
		</div>
		<p class="demo-note">Double: @x.Text(XDouble)</p>
	</div>
}

// On the page: @x.Document(XCounterDemo())`

const xFormsStateSrc = `var XPrice = x.NewNumSignal(12.0)
var XQty   = x.NewNumSignal(2.0)
var XTotal = x.Computed(x.Mul(XPrice, XQty))
var XTax   = x.NewBoolSignal(false)
var XTaxed = x.Computed(x.Mul(XTotal, 1.19))`

const xFormsInputsSrc = `<p><label>Price</label> @x.Number(XPrice)</p>
<p><label>Quantity</label> @x.Number(XQty)</p>`

const xFormsCheckboxSrc = `<p>@x.Checkbox(XTax) <label>include 19% tax</label></p>
<p class="demo-note">
	Total:
	@x.El("span", x.Show(x.Computed(x.Not(XTax)))) {
		@x.Text(XTotal)
	}
	@x.El("span", x.Show(XTax)) {
		@x.Text(XTaxed)
	}
</p>

// Show toggles display, keeping both spans live — exactly one is visible.
// x.Not inside a memo is the inverse; the two can never both show.`

const xFormsSrc = `var XPrice = x.NewNumSignal(12.0)
var XQty   = x.NewNumSignal(2.0)
var XTotal = x.Computed(x.Mul(XPrice, XQty))
var XTax   = x.NewBoolSignal(false)
var XTaxed = x.Computed(x.Mul(XTotal, 1.19))

templ XFormsDemo() {
	<div class="demo">
		<p><label>Price</label> @x.Number(XPrice)</p>
		<p><label>Quantity</label> @x.Number(XQty)</p>
		<p class="demo-note">
			Total:
			@x.El("span", x.Show(x.Computed(x.Not(XTax)))) {
				@x.Text(XTotal)
			}
			@x.El("span", x.Show(XTax)) {
				@x.Text(XTaxed)
			}
		</p>
		<p>@x.Checkbox(XTax) <label>include 19% tax</label></p>
	</div>
}

// On the page: @x.Document(XFormsDemo())`

const xTodoStateSrc = `type XTodo struct {
	ID   string ` + "`json:\"id\"`" + `
	Text string ` + "`json:\"text\"`" + `
}

var XTodos = x.NamedListSignal("todos", []XTodo{
	{ID: "1", Text: "Read the docs"},
	{ID: "2", Text: "Build something"},
})
var XDraft = x.NamedStrSignal("draft", "")`

const xTodoListSrc = `<div data-x-list class="todo-list">
	@x.List(XTodos, "id", xTodoRow, "todoRow")
</div>`

const xTodoCountSrc = `<p class="demo-note"><span data-x-remaining>2 remaining</span></p>

@x.Effect() {
	<script>
		document.querySelector("[data-x-remaining]").textContent = todos.get().length + " remaining";
	</script>
}`

const xTodoScriptSrc = `<script>
	function todoRow(item) {
		return "<li>" + item.text + "</li>";
	}
	document.querySelector("[data-x-form]").addEventListener("submit", (e) => {
		e.preventDefault();
		const text = draft.get();
		if (!text) return;
		todos.set([...todos.get(), { id: String(Date.now()), text }]);
		draft.set("");
	});
</script>`

const xTodoSrc = `type XTodo struct {
	ID   string ` + "`json:\"id\"`" + `
	Text string ` + "`json:\"text\"`" + `
}

var XTodos = x.NamedListSignal("todos", []XTodo{
	{ID: "1", Text: "Read the docs"},
	{ID: "2", Text: "Build something"},
})
var XDraft = x.NamedStrSignal("draft", "")

templ XTodoDemo() {
	<div class="demo">
		<form class="todo-form" data-x-form>
			@x.Input(XDraft)
			<button type="submit">Add</button>
		</form>
		<div data-x-list class="todo-list">
			@x.List(XTodos, "id", xTodoRow, "todoRow")
		</div>
		<p class="demo-note"><span data-x-remaining>2 remaining</span></p>
		@xTodoScript()
		@x.Effect() {
			<script>
				document.querySelector("[data-x-remaining]").textContent = todos.get().length + " remaining";
			</script>
		}
	</div>
}

templ xTodoRow(item XTodo) {
	<li>{ item.Text } <button data-remove={ item.ID }>remove</button></li>
}

templ xTodoScript() {
	<script>
		function todoRow(item) {
			return "<li>" + item.text + "</li>";
		}
		document.querySelector("[data-x-form]").addEventListener("submit", (e) => {
			e.preventDefault();
			const text = draft.get();
			if (!text) return;
			todos.set([...todos.get(), { id: String(Date.now()), text }]);
			draft.set("");
		});
	</script>
}

// On the page: @x.Document(XTodoDemo())`

const xAsyncStateSrc = `var XPosts = x.NamedListSignal("posts", []XPost{
	{ID: "0", Title: "Rendered at build time"},
})
var XLoading = x.NamedBoolSignal("loading", false)`

const xAsyncListSrc = `<div data-x-list class="todo-list">
	@x.List(XPosts, "id", xPostRow, "postRow")
</div>`

const xAsyncScriptSrc = `templ xAsyncScript() {
	<script>
		function postRow(p) {
			return "<li>" + p.title + "</li>";
		}
		const status = document.querySelector("[data-x-status]");
		function load() {
			loading.set(true);
			status.textContent = "Loading…";
			setTimeout(() => {
				posts.set([
					{ id: "1", title: "Signals track their own dependencies" },
					{ id: "2", title: "Keyed lists move nodes instead of rebuilding them" },
					{ id: "3", title: "Scripts are the escape hatch — this is one" },
				]);
				loading.set(false);
				status.textContent = posts.get().length + " posts loaded";
			}, 500);
		}
		document.querySelector("[data-x-reload]").addEventListener("click", load);
		load();
	</script>
}`

const xAsyncSrc = `type XPost struct {
	ID    string ` + "`json:\"id\"`" + `
	Title string ` + "`json:\"title\"`" + `
}

var XPosts = x.NamedListSignal("posts", []XPost{
	{ID: "0", Title: "Rendered at build time"},
})
var XLoading = x.NamedBoolSignal("loading", false)

templ XAsyncDemo() {
	<div class="demo">
		<div class="counter-row">
			<button type="button" data-x-reload>Reload</button>
			@x.El("span", x.Show(XLoading)) {
				Loading…
			}
		</div>
		<p class="demo-note" data-x-status>1 post rendered at build time</p>
		<div data-x-list class="todo-list">
			@x.List(XPosts, "id", xPostRow, "postRow")
		</div>
		@xAsyncScript()
	</div>
}

templ xAsyncScript() {
	<script>
		function postRow(p) {
			return "<li>" + p.title + "</li>";
		}
		const status = document.querySelector("[data-x-status]");
		function load() {
			loading.set(true);
			status.textContent = "Loading…";
			setTimeout(() => {
				posts.set([
					{ id: "1", title: "Signals track their own dependencies" },
					{ id: "2", title: "Keyed lists move nodes instead of rebuilding them" },
					{ id: "3", title: "Scripts are the escape hatch — this is one" },
				]);
				loading.set(false);
				status.textContent = posts.get().length + " posts loaded";
			}, 500);
		}
		document.querySelector("[data-x-reload]").addEventListener("click", load);
		load();
	</script>
}

// On the page: @x.Document(XAsyncDemo())`

const xFaqStateSrc = `var XFaq1 = x.NamedBoolSignal("faq1", true)
var XFaq2 = x.NamedBoolSignal("faq2", false)
var XFaq3 = x.NamedBoolSignal("faq3", false)`

const xFaqToggleSrc = `@x.El("button", x.On("click", XFaq1.Toggle()), x.Class("on", XFaq1)) {
			What is compiled markup?
		}`

const xFaqWhenSrc = `@x.When(XFaq1, xFaqAnswer("The markers render at build time and subscribe in the browser."))

// xFaqAnswer is the subtree When mounts and unmounts.
templ xFaqAnswer(text string) {
	<p class="demo-note">{ text }</p>
}`

const xFaqSrc = `var XFaq1 = x.NamedBoolSignal("faq1", true)
var XFaq2 = x.NamedBoolSignal("faq2", false)
var XFaq3 = x.NamedBoolSignal("faq3", false)

templ XFaqDemo() {
	<div class="demo">
		@x.El("button", x.On("click", XFaq1.Toggle()), x.Class("on", XFaq1)) {
			What is compiled markup?
		}
		@x.When(XFaq1, xFaqAnswer("The markers render at build time and subscribe in the browser — the page reads correctly with JavaScript disabled."))
		@x.El("button", x.On("click", XFaq2.Toggle()), x.Class("on", XFaq2)) {
			When should I reach for WebAssembly instead?
		}
		@x.When(XFaq2, xFaqAnswer("When the behaviour is a program — canvas, fetching, timers. Markers cover state, bindings and structure; everything else is Go in the browser."))
		@x.El("button", x.On("click", XFaq3.Toggle()), x.Class("on", XFaq3)) {
			What happens without JavaScript?
		}
		@x.When(XFaq3, xFaqAnswer("The page keeps its static form — the first answer is already open. Nothing needs a script to be readable."))
	</div>
}

templ xFaqAnswer(text string) {
	<p class="demo-note">{ text }</p>
}

// On the page: @x.Document(XFaqDemo())`
