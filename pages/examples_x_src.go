package pages

// Source snippets for the compiled-markup example pages. They live in a plain
// .go file rather than a .templ one because templ rewrites raw strings that
// contain templ syntax into their generated Go — a snippet is meant to show
// the source, not the output of the compiler.

const xCounterSrc = `var XCount = x.Named("count", 0)
var XOpen  = x.NewSignal(false)

templ XCounterDemo() {
	<div class="demo">
		<div class="counter-row">
			@x.El("button", x.On("click", XCount.Set(XCount.Add(1)))) {
				@x.Text(XCount) +
			}
			@x.El("button", x.On("click", XCount.Set(x.Lit(0)))) {
				reset
			}
		</div>
		<p>
			@x.El("button", x.On("click", XOpen.Set(x.Not(XOpen)))) {
				toggle secret
			}
			@x.When(XOpen, xSecret())
		</p>
		@x.Effect() {
			<script>
				document.querySelector("[data-x-note]").textContent = "Count: " + count.get();
			</script>
		}
	</div>
}

// On the page: @x.Document(XCounterDemo())`

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
		@xTodoScript()
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
