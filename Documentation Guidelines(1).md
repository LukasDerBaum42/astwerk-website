# Documentation Guidelines

Good documentation lets a programmer understand and use something quickly.

Assume the reader:
- Knows how programming works.
- Can read code.
- Does not necessarily know your project.

The **actual documentation should be short**. This guide is intentionally more detailed so contributors know what "good documentation" means.

---

## 1. Keep the documentation short

The main documentation should answer:

- What is this?
- What does it do?
- What do I need to know to use it?

Don't explain basic programming concepts.

### Good

```md
A `Signal` stores a reactive value. Effects that read the
signal automatically run again when its value changes.
```

### Bad

```md
A Signal is a special object that contains a value. A value
is some piece of information stored by a computer. When a
program wants to access the value, it...
```

The reader already knows what a value is. Don't waste their time.

---

## 2. Use `Documentation` and `Example`

A documented feature should normally have this structure:

```md
# Signal

A reactive value that automatically notifies its dependents.

## Documentation

`Signal` stores a value and tracks which reactive computations
read it.

<details>
<summary>More information</summary>

Signals use dependency tracking internally...

</details>

## Example

```js
const count = signal(0);

effect(() => {
    console.log(count.value);
});

count.value++;
```
```

The important information is visible immediately. Extra information is there if someone actually wants it.

---

## 3. Explain what it does, not how computers work

Explain **project-specific behavior**.

### Bad

```md
`set()` is a function. Functions can receive arguments.
The first argument is the object on which the function operates...
```

### Good

```md
`set(value)` changes the signal's value and triggers all
effects that depend on it.
```

Assume programming knowledge unless the concept is specific to your project.

---

## 4. Give a useful example

Examples should demonstrate the **normal use case**.

### Bad

```js
const x = signal(0);
const y = signal(1);
const z = signal(2);

console.log(x, y, z);
```

This technically uses the API, but doesn't show why it exists.

### Good

```js
const count = signal(0);

effect(() => {
    document.title = `Count: ${count.value}`;
});

count.value++;
```

The reader can immediately see what the feature is useful for.

- [ ] Small
- [ ] Valid
- [ ] Copyable
- [ ] Realistic
- [ ] Shows the intended usage

---

## 5. Explain examples briefly

Don't make the example section another tutorial.

### Good

```md
## Example

```js
const name = signal("Lukas");

effect(() => {
    element.textContent = name.value;
});

name.value = "Bob";
```

Changing `name` automatically updates the element.
```

That's enough.

If understanding the example requires a long explanation, consider improving the example instead.

---

## 6. Put advanced information behind a collapse

The main documentation should not be buried under implementation details.

### Good

```md
## Documentation

`computed()` creates a value derived from other reactive values.

<details>
<summary>How it works internally</summary>

When the computation runs, Astwerk tracks every signal it reads.
Those signals become dependencies of the computed value.

</details>
```

Useful things to put here:

- Implementation details
- Internal architecture
- Edge cases
- Performance details
- Design decisions
- Advanced behavior

The reader should **not need to open it** for normal usage.

---

## 7. Explain surprising behavior

If something behaves in a way that a programmer might not expect, mention it.

### Example

```md
## Documentation

`effect()` runs immediately when created and again whenever one
of its dependencies changes.

<details>
<summary>Why does it run immediately?</summary>

The first execution is required to discover which signals the
effect depends on.

</details>
```

The important behavior is in the main text. The explanation of *why* is optional.

---

## 8. Explain the "why" when it matters

Don't explain every design decision.

Do explain decisions that affect how the programmer uses the system.

### Example

```md
Components automatically dispose their effects when destroyed.

<details>
<summary>Why?</summary>

Otherwise effects could continue running after their component
has been removed.
</details>
```

That's useful because it explains behavior the programmer might otherwise find confusing.

---

## 9. Keep API documentation close to the API

Like a good `.h` file, the basic documentation should live next to the thing it describes.

### Example

```go
// Signal stores a reactive value.
//
// Reading a signal inside an effect creates a dependency on it.
type Signal[T any] struct {
    ...
}
```

Use external `.md` documentation for larger explanations.

Think:

```text
Code comment
    ↓
What does this API do?

Documentation
    ↓
How do I use it?

Architecture documentation
    ↓
How does the system work internally?
```

---

## 10. Use consistent terminology

If something is called a `Signal`, keep calling it a `Signal`.

Don't randomly switch between:

```text
Signal
Reactive value
Observable
State object
Reactive variable
```

unless these are actually different things.

### Good

```md
A Signal stores a value.

An Effect executes code based on Signals.

A Computed derives a value from Signals.
```

The reader should always know whether two terms refer to the same thing.

---

## 11. Make documentation searchable

Use the names programmers will actually search for.

### Good

```md
# Signals

## `signal()`

## `Signal.value`

## Updating a Signal
```

### Bad

```md
# Reactive Stuff

## Creating Things

## Changing Things
```

Someone searching for `signal()` shouldn't have to play archaeology with your documentation.

---

## 12. Don't duplicate information

Have one canonical explanation.

### Bad

```text
README:
  Signals automatically update effects.

Wiki:
  Signals automatically update effects.

API docs:
  Signals automatically update effects.

Architecture docs:
  Signals automatically update effects.
```

Now you have four places to keep synchronized.

Instead, write the explanation once and link to it.

---

## 13. Keep examples up to date

A broken example is worse than no example because it actively wastes someone's time.

- [ ] Examples compile/run.
- [ ] APIs used by examples still exist.
- [ ] Output matches the documented behavior.
- [ ] Examples are updated when the API changes.

Prefer examples that can actually be tested automatically.

---

# Documentation Checklist

## Main documentation

- [ ] Immediately explains what the thing is.
- [ ] Explains what it does.
- [ ] Is concise.
- [ ] Assumes programming knowledge.
- [ ] Uses simple language.
- [ ] Uses consistent terminology.
- [ ] Doesn't explain unnecessary implementation details.

## `Documentation`

- [ ] Important behavior is documented.
- [ ] Important arguments/options are documented.
- [ ] Important limitations are mentioned.
- [ ] Surprising behavior is explained.
- [ ] Advanced details are optional/collapsible.

## `Example`

- [ ] Has a small example.
- [ ] Example actually works.
- [ ] Shows normal usage.
- [ ] Can be understood without reading the implementation.
- [ ] Doesn't contain unnecessary code.

## Quality

- [ ] A programmer can use the feature after reading the docs.
- [ ] Important information is easy to find.
- [ ] API names and errors are searchable.
- [ ] Documentation matches the current implementation.
- [ ] No information is unnecessarily duplicated.

---

# The Core Rule

> **Explain enough to understand.  
> Show enough to use.  
> Hide everything else until it is needed.**

The best documentation should feel like a **good `.h` file combined with the Arch Wiki**: concise at first glance, but with deeper information available when you need it.