# html-ui-to-vue-vapor design

Status: approved and implemented; verification and review recorded in docs/implementation-progress.md.

## Goal and scope

Build a Go CLI in `html-ui-to-vue-vapor` that converts the existing
`html-ui` TypeScript source contract into a standalone Vue Vapor single-file
component. All implementation, dependencies, tests, and documentation belong
to this directory. Do not modify `html-ui-cli` or import its internal catalog.

The existing producer is only a reference for the input syntax. Implementation and tests must not invoke it, capture its outputs, enumerate its catalog, or depend on specific primitive names. Its default output
and `--ts` output are TypeScript, including a constrained DOM factory. The
converter must work from the stream without knowing the primitive name in
advance or invoking the producer internally.

```sh
html-ui accordion | html-ui-to-vue-vapor > accordion.vue
html-ui accordion --ts | html-ui-to-vue-vapor > accordion.vue
```

## Input decision

Consume source directly. A declaration-only input retains types and literal
metadata but loses the factory body, so it cannot independently recover node
structure, binding destinations, or slot placement. Do not require a separate
declaration file or a TypeScript compiler at CLI runtime.

Three approaches were considered:

1. A Go parser for the producer's constrained source contract: recommended.
   It supports a standalone binary and explicit diagnostics, but requires
   maintaining a clearly limited grammar.
2. A Go CLI with a TypeScript compiler helper: provides a mature TypeScript
   parser, but adds a JavaScript runtime and distribution dependencies.
3. HTML plus declarations: needs additional markup conventions and producer
   changes, outside this task's scope.

## Architecture

Separate command handling, contract parsing/validation, and Vue generation.

- `cmd/html-ui-to-vue-vapor`: process entry point.
- `internal/cli`: stdin, flags, stdout, stderr, and exit codes.
- `internal/contract`: lexical scanning, constrained parsing, and component
  model validation.
- `internal/generate`: deterministic SFC rendering from the validated model.
- `tests`: Vue compiler, type, and browser integration checks.

The component model contains props and defaults, slot signatures and scope
sources, event types and native event mappings, native nodes, ordered bindings,
and ordered child insertions. Parsing must preserve binding order as well as
child order; range constraints before value initialization matter.

## Supported source

Accept contract version 1 and the forms emitted by the existing generator:
exported interfaces, literal constants with `as const` and `satisfies`, prop
destructuring with defaults, literal element creation, static attributes,
property bindings, attribute bindings including `String(prop)`, guarded
optional bindings, child insertion, slot invocation with explicit or shorthand
scope fields, and the factory return.

This is a parser for the generated contract, not a general TypeScript compiler.
Reject unknown statements, arbitrary expressions, unsupported versions,
duplicate declarations, invalid references, cyclic or multiply inserted nodes,
inconsistent slot scopes, and incomplete factories. Never evaluate input or
resolve its imports. Report line and column locations for parse failures.

## Vue output

Emit `<script setup lang="ts" vapor>` and a template. Preserve prop optionality,
literal union types, defaults, native event types, required and optional slots,
and scoped slot fields. Use template slot outlets rather than invoking slot
functions in setup. Slot content types must describe Vue rendering instead of
requiring native `HTMLElement` values.

Static attributes remain static. Attribute and DOM property bindings retain
their distinction; property bindings must explicitly target properties where
necessary. Omit absent optional attributes. Preserve source `setAttribute`
stringification, including boolean strings; native boolean property assignments
control presence through the DOM. ARIA boolean values remain strings. Preserve native root tags,
child ordering, void elements, and accessibility attributes. Emit no styles.

## State inference and public behavior

A default supplies an initial value; it does not establish mutability. Infer
local mutable state only from fields in `nativeEvents[event].state`.

For accordion, infer `open` from the toggle mapping, initialize its ref from
the supplied prop or `false`, bind it to the details element, read
`event.currentTarget.open` on toggle, and pass its current value to the content
slot. Keep `name` as a prop binding.

For each mapped field:

- Initialize local state from the prop and its contract default.
- Synchronize later parent prop changes into local state without resetting it
  on unrelated parent updates.
- Update state from the mapped native DOM property on the native event.
- Forward the original typed event and expose a typed `update:<prop>` event
  so consumers can use named `v-model` bindings.
- Generate no independent local ref for props without a native state mapping.

Preserve initial form reset baselines for `defaultValue` and `defaultChecked`.
After native form reset, synchronize local state and slot scopes from the DOM;
release any added listeners on unmount. Native sanitization and numeric
properties such as `valueAsNumber` must be respected rather than approximated
with string parsing.

## Compatibility boundary

Target a pinned Vue version that supports Vapor and verify against its actual
compiler and runtime. Document pure Vapor mounting and VDOM interop setup.

An input may be marked `adapter-required` in its metadata. Such source
does not encode complete keyboard, focus, selection, or dismissal algorithms.
Conversion preserves their declared structure and native event behavior, but
must identify the missing adapter behavior in generated documentation and a
stderr diagnostic. Do not advertise such output as a complete interaction
implementation. Implementing those algorithms is a separate, explicitly
specified scope; types and defaults cannot reliably infer them.

## CLI and errors

Read source from stdin and write only the complete generated SFC to stdout.
Support `--help` and `--version`. Buffer conversion before writing output so
invalid input cannot produce a partially generated component. Use stderr for
diagnostics. Return 0 for successful conversion, 2 for invalid arguments or
unsupported/invalid input, and 1 for input/output failures.

## Verification and acceptance

1. Go tests exercise tokenization, parsing, validation, inference, escaping,
   deterministic rendering, and command error/output behavior.
2. Use small hand-written input strings to exercise syntax and behavior: DOM
   properties versus attributes, optional bindings, nested event targets,
   scoped slots, and state mappings. No captured producer output is required.
3. Compile components generated from these synthetic inputs with the pinned
   Vue Vapor toolchain. Component names and supported combinations must not
   be tied to a primitive list or fixed catalog size.
4. Type checks verify props, defaults, named model updates, event payloads,
   required slots, and scoped slot types, including rejected invalid usage.
5. Browser checks verify disclosure toggle and grouping, live scoped slots,
   parent prop updates, input/checkbox interaction, numeric sanitization, form
   reset, event forwarding, and cleanup after unmount.
6. Verify a freshly built standalone Go binary can convert a hand-written input stream
   without Node, TypeScript, or the producer installed.

Success means the pipe produces a valid Vapor SFC whose declared native
structure, props, slots, and event/state behavior survive conversion. Any
adapter-required behavior remains explicitly identified and is never silently
invented or represented as implemented.
