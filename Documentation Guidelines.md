# Documentation Guidelines

This guide defines what good project documentation should look like.

The goal is **not** to explain everything. The goal is to let a programmer quickly understand and use something without having to read the source code.

Assume the reader knows how programming works, but does not know your project.

---

## 1. Keep the actual documentation short

The main documentation should answer:

- What is this?
- What does it do?
- What do I need to know to use it?

Avoid explaining things the reader can already infer from the code.

Prefer:

```md
A `Signal` stores a reactive value. Effects that read the
signal automatically run again when its value changes.
```

Over:

```md
A Signal is a special object that contains a value. A value
is some piece of information that can be represented in
memory. When the value changes, the program...
```

**Rule:** If a programmer already understands it, don't explain it.

---

## 2. Use a consistent structure

A documented feature should generally look like:

```md
# Feature

Short description.

## Documentation

The important information about the feature.

<details>
<summary>More information</summary>

Optional detailed information, internals, edge cases, etc.

</details>

## Example

```language
// Small useful example
```

Short explanation if necessary.
```

Not every feature needs every part, but **Documentation** and **Example** should be the default.

---

## 3. Documentation section

The `Documentation` section describes how the feature works from the user's perspective.

Include:

- What it does.
- Important parameters/options.
- Return values or effects.
- Important limitations.
- Anything that is surprising or easy to get wrong.

Don't include implementation details unless they are useful to the user.

### Good

```md
## Documentation

`computed()` creates a value derived from other reactive values.
It automatically updates when one of its dependencies changes.
```

### Bad

```md
## Documentation

`computed()` creates a `ReactiveNode`, pushes it onto the global
tracking stack, executes the callback, collects dependencies,
then subscribes the node to each dependency...
```

The second one belongs in an **implementation/architecture** document, not the normal API documentation.

---

## 4. Optional information

Detailed information should not clutter the main documentation.

Use collapsible sections when supported:

```md
<details>
<summary>How it works internally</summary>

Detailed explanation...

</details>
```

Good candidates for optional information:

- Implementation details
- Advanced behavior
- Edge cases
- Performance considerations
- Design decisions
- Internals

The reader should be able to understand and use the feature **without opening these sections**.

---

## 5. Example section

Every non-trivial feature should have a small example.

The example should show the **normal way to use it**, not every possible feature.

### Good

```js
const count = signal(0);

effect(() => {
    console.log(count.value);
});

count.value++;
```

Then, if necessary:

```md
The effect runs again when `count` changes.
```

Avoid huge examples that demonstrate ten unrelated features.

The example should be:

- [ ] Small
- [ ] Valid
- [ ] Copyable
- [ ] Representative
- [ ] Easy to understand

---

## 6. Explain concepts, not syntax

Don't explain basic programming concepts.

Do explain project-specific concepts.

For example, don't explain what a function parameter is.

Do explain what `Signal` means **in your project**, especially if it behaves differently from similar concepts elsewhere.

---

## 7. Explain the "why" when necessary

Usually documentation only needs to explain **what** something does.

Explain **why** when the behavior would otherwise seem strange.

Example:

```md
Components are destroyed automatically when their parent is
destroyed. This prevents reactive effects from continuing to
run after the component has been removed.
```

This explains a behavior that might otherwise be confusing.

---

## 8. Keep information close to where it belongs

Use documentation in the code itself for small API-level information:

```go
// Signal stores a reactive value.
//
// Reading the signal inside an effect creates a dependency.
type Signal[T any] struct {
    ...
}
```

Use external documentation for larger explanations:

```text
API comments
    ↓
What does this thing do?

Documentation
    ↓
How do I use it?

Architecture docs
    ↓
How does the system work internally?
```

This is similar to a good `.h` file: the important information should be close to the thing being used.

---

## 9. Make documentation searchable

A programmer should be able to find something by searching for:

- The function/type name.
- An option name.
- An error message.
- A relevant concept.

Use actual project terminology in headings.

Prefer:

```md
# Signals
## signal()
## Signal.value
```

Over:

```md
# Reactive Things
## Creating Stuff
## Changing Stuff
```

---

## 10. Don't duplicate documentation

There should be one **canonical explanation** for a concept.

Other places should link to it rather than copying it.

This prevents:

```text
README → outdated explanation
Wiki → different explanation
API docs → another explanation
Source → fourth explanation
```

---

# Documentation Checklist

## General

- [ ] Does it immediately explain what the thing is?
- [ ] Is the main explanation short?
- [ ] Does it assume programming knowledge?
- [ ] Does it avoid unnecessary detail?
- [ ] Is the terminology consistent?

## Documentation section

- [ ] Is the behavior clearly explained?
- [ ] Are important arguments/options documented?
- [ ] Are important limitations mentioned?
- [ ] Are surprising behaviors explained?
- [ ] Is unnecessary implementation detail excluded?
- [ ] Is advanced information optional?

## Example section

- [ ] Is there a minimal useful example?
- [ ] Does the example actually work?
- [ ] Does it show normal usage?
- [ ] Is it easy to copy and modify?
- [ ] Is unnecessary code removed?

## Quality

- [ ] Can a programmer understand it without reading the source?
- [ ] Can they start using it after reading it?
- [ ] Can they find it through search?
- [ ] Does it match the current implementation?
- [ ] Are details that aren't needed kept out of the main text?

---

# The Core Principle

> **Explain enough to understand.  
> Show enough to use.  
> Hide everything else until it is needed.**

Good documentation should feel more like a **good `.h` file + Arch Wiki** than a textbook.