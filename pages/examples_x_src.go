package pages

// Source snippets for the compiled-markup example pages. They live in a plain
// .go file rather than a .templ one because templ rewrites raw strings that
// contain templ syntax into their generated Go — a snippet is meant to show
// the source, not the output of the compiler.

const xCounterStateSrc = `var XCount  = x.Named("count", 0)
var XDouble = x.Computed(XCount.Mul(2))`

const xCounterButtonsSrc = `<div class="counter-row">
	@x.El("button", x.On("click", XCount.Set(XCount.Sub(1)))) {
		&minus;
	}
	@x.Text(XCount)
	@x.El("button", x.On("click", XCount.Set(XCount.Add(1)))) {
		+
	}
	@x.El("button", x.On("click", XCount.Set(x.Lit(0)))) {
		reset
	}
</div>`

const xCounterDerivedSrc = `<p class="demo-note">Double: <span>@x.Text(XDouble)</span></p>`

const xCounterSrc = `var XCount  = x.Named("count", 0)
var XDouble = x.Computed(XCount.Mul(2))

templ XCounterDemo() {
	<div class="demo">
		<div class="counter-row">
			@x.El("button", x.On("click", XCount.Set(XCount.Sub(1)))) {
				&minus;
			}
			@x.Text(XCount)
			@x.El("button", x.On("click", XCount.Set(XCount.Add(1)))) {
				+
			}
			@x.El("button", x.On("click", XCount.Set(x.Lit(0)))) {
				reset
			}
		</div>
		<p class="demo-note">Double: @x.Text(XDouble)</p>
	</div>
}

// On the page: @x.Document(XCounterDemo())`

const xFormsStateSrc = `var XPrice = x.NewSignal(10.0)
var XQty   = x.NewSignal(2.0)
var XTotal = x.Computed(XPrice.Mul(XQty))
var XTax   = x.NewSignal(false)`

const xFormsInputsSrc = `<p><label>Price</label> @x.Number(XPrice)</p>
<p><label>Quantity</label> @x.Number(XQty)</p>
<p class="demo-note"><span>Total: </span>@x.Text(XTotal)</p>`

const xFormsCheckboxSrc = `<p>@x.Checkbox(XTax) <label>include tax</label></p>
@x.El("p", x.Show(XTax)) {
	A checked box can drive visibility, too.
}`

const xFormsSrc = `var XPrice = x.NewSignal(10.0)
var XQty   = x.NewSignal(2.0)
var XTotal = x.Computed(XPrice.Mul(XQty))
var XTax   = x.NewSignal(false)

templ XFormsDemo() {
	<div class="demo">
		<p><label>Price</label> @x.Number(XPrice)</p>
		<p><label>Quantity</label> @x.Number(XQty)</p>
		<p class="demo-note"><span>Total: </span>@x.Text(XTotal)</p>
		<p>@x.Checkbox(XTax) <label>include tax</label></p>
		@x.El("p", x.Show(XTax)) {
			A checked box can drive visibility, too.
		}
	</div>
}

// On the page: @x.Document(XFormsDemo())`

const xTodoStateSrc = `type XTodo struct {
	ID   string ` + "`json:\"id\"`" + `
	Text string ` + "`json:\"text\"`" + `
}

var XTodos = x.Named("todos", []XTodo{
	{ID: "1", Text: "Read the docs"},
	{ID: "2", Text: "Build something"},
})
var XDraft = x.Named("draft", "")`

const xTodoListSrc = `<div data-x-list>
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

var XTodos = x.Named("todos", []XTodo{
	{ID: "1", Text: "Read the docs"},
	{ID: "2", Text: "Build something"},
})
var XDraft = x.Named("draft", "")

templ XTodoDemo() {
	<div class="demo">
		<form class="todo-form" data-x-form>
			@x.Input(XDraft)
			<button type="submit">Add</button>
		</form>
		<div data-x-list>
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

const xAsyncStateSrc = `var XPosts = x.Named("posts", []XPost{
	{ID: "0", Title: "Rendered at build time"},
})
var XLoading = x.Named("loading", false)`

const xAsyncListSrc = `<div data-x-list>
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

var XPosts = x.Named("posts", []XPost{
	{ID: "0", Title: "Rendered at build time"},
})
var XLoading = x.Named("loading", false)

templ XAsyncDemo() {
	<div class="demo">
		<div class="counter-row">
			<button type="button" data-x-reload>Reload</button>
			@x.El("span", x.Show(XLoading)) {
				Loading…
			}
		</div>
		<p class="demo-note" data-x-status>1 post rendered at build time</p>
		<div data-x-list>
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

const xCanvasMarkupSrc = `<canvas data-x-canvas width="600" height="200"></canvas>
<p class="demo-note">Pointer x: <span>@x.Text(XPointerX)</span> — move the pointer over the canvas.</p>`

const xCanvasScriptSrc = `templ xCanvasScript() {
	<script>
		const canvas = document.querySelector("[data-x-canvas]");
		const ctx = canvas.getContext("2d");
		const r = canvas.getBoundingClientRect();
		canvas.width = Math.round(r.width);
		canvas.height = Math.round(r.height);
		const w = canvas.width;
		const h = canvas.height;
		let elapsed = 0;
		let last = null;
		let px = -9999;
		canvas.addEventListener("pointermove", (e) => {
			const rect = canvas.getBoundingClientRect();
			px = e.clientX - rect.left;
			pointerX.set(Math.round(px));
		});
		function frame(now) {
			if (last !== null) elapsed += (now - last) / 1000;
			last = now;
			ctx.clearRect(0, 0, w, h);
			for (let i = 0; i < w; i += 6) {
				const wave = Math.sin(i / 40 + elapsed * 2) * 40;
				const lift = Math.exp(-Math.pow(i - px, 2) / 4000) * 30;
				ctx.fillStyle = "#4f7cff";
				ctx.fillRect(i, h / 2 + wave - lift, 3, 3);
			}
			requestAnimationFrame(frame);
		}
		requestAnimationFrame(frame);
	</script>
}`

const xCanvasSrc = `templ XCanvasDemo() {
	<div class="demo">
		<canvas data-x-canvas width="600" height="200"></canvas>
		<p class="demo-note">Pointer x: <span>@x.Text(XPointerX)</span> — move the pointer over the canvas.</p>
		@xCanvasScript()
	</div>
}

templ xCanvasScript() {
	<script>
		const canvas = document.querySelector("[data-x-canvas]");
		const ctx = canvas.getContext("2d");
		const r = canvas.getBoundingClientRect();
		canvas.width = Math.round(r.width);
		canvas.height = Math.round(r.height);
		const w = canvas.width;
		const h = canvas.height;
		let elapsed = 0;
		let last = null;
		let px = -9999;
		canvas.addEventListener("pointermove", (e) => {
			const rect = canvas.getBoundingClientRect();
			px = e.clientX - rect.left;
			pointerX.set(Math.round(px));
		});
		function frame(now) {
			if (last !== null) elapsed += (now - last) / 1000;
			last = now;
			ctx.clearRect(0, 0, w, h);
			for (let i = 0; i < w; i += 6) {
				const wave = Math.sin(i / 40 + elapsed * 2) * 40;
				const lift = Math.exp(-Math.pow(i - px, 2) / 4000) * 30;
				ctx.fillStyle = "#4f7cff";
				ctx.fillRect(i, h / 2 + wave - lift, 3, 3);
			}
			requestAnimationFrame(frame);
			}
		</script>
		}

		// On the page: @x.Document(XCanvasDemo())`

const xFaqStateSrc = `var XFaq1 = x.Named("faq1", true)
		var XFaq2 = x.Named("faq2", false)
		var XFaq3 = x.Named("faq3", false)`

const xFaqToggleSrc = `@x.El("button", x.On("click", XFaq1.Set(x.Not(XFaq1))), x.Class("on", XFaq1)) {
			What is compiled markup?
		}`

const xFaqWhenSrc = `@x.When(XFaq1, xFaqAnswer("The markers render at build time and subscribe in the browser."))

// xFaqAnswer is the subtree When mounts and unmounts.
templ xFaqAnswer(text string) {
	<p class="demo-note">{ text }</p>
}`

const xFaqSrc = `var XFaq1 = x.Named("faq1", true)
var XFaq2 = x.Named("faq2", false)
var XFaq3 = x.Named("faq3", false)

templ XFaqDemo() {
	<div class="demo">
		@x.El("button", x.On("click", XFaq1.Set(x.Not(XFaq1))), x.Class("on", XFaq1)) {
			What is compiled markup?
		}
		@x.When(XFaq1, xFaqAnswer("The markers render at build time and subscribe in the browser — the page reads correctly with JavaScript disabled."))
		@x.El("button", x.On("click", XFaq2.Set(x.Not(XFaq2))), x.Class("on", XFaq2)) {
			When should I reach for WebAssembly instead?
		}
		@x.When(XFaq2, xFaqAnswer("When the behaviour is a program — canvas, fetching, timers. Markers cover state, bindings and structure; everything else is Go in the browser."))
		@x.El("button", x.On("click", XFaq3.Set(x.Not(XFaq3))), x.Class("on", XFaq3)) {
			What happens without JavaScript?
		}
		@x.When(XFaq3, xFaqAnswer("The page keeps its static form — the first answer is already open. Nothing needs a script to be readable."))
	</div>
}

templ xFaqAnswer(text string) {
	<p class="demo-note">{ text }</p>
}

// On the page: @x.Document(XFaqDemo())`
