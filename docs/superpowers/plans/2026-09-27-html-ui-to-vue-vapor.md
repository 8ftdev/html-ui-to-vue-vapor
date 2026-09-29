# html-ui-to-vue-vapor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone Go binary that reads the existing `html-ui` TypeScript contract from stdin and writes a typed Vue Vapor component to stdout.

**Architecture:** A constrained lexer and parser produce a validated component model. A separate generator renders Vue props, slots, ordered native bindings, event forwarding, and inferred local state. The CLI performs no evaluation, dependency loading, producer invocation, or runtime TypeScript compilation.

**Tech Stack:** Go 1.24.0 and its standard library for the binary. Development checks use Bun, TypeScript 7.0.2, Vue and `@vue/compiler-sfc` 3.6.0-rc.9, `vue-tsc` 3.3.11, and Playwright 1.63.0; package versions are exact and the lockfile is committed.

**Spec:** [Approved design](../specs/2026-09-27-html-ui-to-vue-vapor-design.md).

## Global Constraints

- All implementation, dependencies, tests, and documentation belong to `html-ui-to-vue-vapor`.
- Do not modify `html-ui-cli` or import its internal catalog.
- Neither implementation nor tests invoke the producer, capture its outputs, enumerate primitives, or require a fixed number of components.
- Verify conversion using small synthetic input strings; identify all props, nodes, slots, and state from the supplied contract.
- Accept contract version 1.
- Do not require a separate declaration file or a TypeScript compiler at CLI runtime.
- Never evaluate input or resolve its imports.
- Emit `<script setup lang="ts" vapor>` and a template.
- Emit no styles.
- Infer local mutable state only from fields in `nativeEvents[event].state`.
- Buffer conversion before writing output so invalid input cannot produce a partially generated component.
- Use stderr for diagnostics.
- Return 0 for successful conversion, 2 for invalid arguments or unsupported/invalid input, and 1 for input/output failures.
- Adapter-required output preserves declared structure and native behavior; document its missing interaction algorithms and emit a stderr diagnostic.
- Keep optionality, defaults, native event types, scoped slot types, binding order, and child order.
- Implement locally in the designated directory. It is currently not a Git repository; do not borrow the sibling producer's repository for commits. If this directory has a repository by execution time, use its normal task-sized commits. Creating a repository is not a prerequisite for coding.

## Review Focus

1. An input string contains quotes, ampersands, Unicode, or `</script>`: conversion preserves the value without changing SFC structure. Task 4 owns this regression.
2. Input is truncated, uses CRLF, or contains duplicate/unknown declarations: reject it with a stable line/column diagnostic and no stdout. Tasks 1–3 and 6 own this regression.
3. Two native events map to the same state field, or a slotted child originates the event: create one ref, read `currentTarget`, and forward each native event exactly once. Tasks 5 and 8 own this regression.
4. A reset is canceled, or the component unmounts before queued reset work runs: leave state unchanged and release listeners. Tasks 5 and 8 own this regression.
5. A user edits local state and the parent rerenders for an unrelated prop: preserve the edit; an actual mapped prop change must synchronize. Tasks 5 and 8 own this regression.

## File and interface map

All paths below are relative to `html-ui-to-vue-vapor`.

| Files | Responsibility |
| --- | --- |
| `go.mod`, `.gitignore` | Module boundary and generated-file exclusions |
| `internal/contract/model.go` | Ordered component, prop, slot, event, and DOM model |
| `internal/contract/lexer.go`, `lexer_test.go` | Tokens, literal decoding, source positions |
| `internal/contract/parser.go`, `declarations.go`, `declarations_test.go` | Parser cursor, interfaces, defaults, event metadata |
| `internal/contract/factory.go`, `validate.go`, `parser_test.go` | DOM construction grammar and cross-reference validation |
| `internal/generate/generate.go`, `template.go`, `script.go`, `escape.go`, `generate_test.go` | Static SFC emission and typed declarations |
| `internal/generate/state.go`, `events.go`, `reset.go`, `state_test.go` | State inference and generated runtime behavior |
| `internal/cli/cli.go`, `cli_test.go`, `cmd/html-ui-to-vue-vapor/main.go` | Stream command and exit codes |
| `internal/testinput/source.go` | Minimal hand-written input constants used only by Go tests |
| `package.json`, `bun.lock`, `tests/compiler.test.ts`, `tests/support/compile.ts` | Actual Vapor compiler integration |
| `tests/typecheck.test.ts`, `tests/types/consumer.vue`, `tests/types/tsconfig.json` | Vue-facing type compatibility |
| `playwright.config.ts`, `tests/browser.pw.ts`, `tests/support/harness.ts`, `tests/serve.ts` | Native interaction tests and browser harness |
| `Makefile`, `README.md`, `docs/compatibility.md`, `scripts/check-standalone.sh` | Build, usage, compatibility limits, portable smoke check |

Shared Go interfaces:

```go
// Package contract; ordered slices preserve source order.
type Position struct { Offset, Line, Column int }
type Field struct { Name, Type string; Optional bool; Pos Position }
type Prop struct { Field; Default json.RawMessage }
type Slot struct { Name string; Optional bool; Scope []Field; Pos Position }
type StateRead struct { Prop, Property string; Pos Position }
type Event struct { Name, Type, Target string; State []StateRead; Pos Position }
type Attribute struct { Name, Value string; Pos Position }
type Binding struct {
    Kind, Name, Prop string // Kind is "attribute" or "property".
    Guarded, Stringify bool
    Pos Position
}
type ScopeBinding struct { Name, Prop string; Pos Position }
type Child struct {
    Node, Slot string // Exactly one is nonempty.
    Optional bool
    Scope []ScopeBinding
    Pos Position
}
type Node struct {
    ID, Tag, DOMType string
    Attributes []Attribute
    Bindings []Binding
    Children []Child
    Pos Position
}
type Component struct {
    Name, Factory, Root, RootType, Behavior string
    Version int
    Props []Prop
    Slots []Slot
    Events []Event
    Nodes []Node
}
type Diagnostic struct { Pos Position; Message string }
func (d *Diagnostic) Error() string
func Parse(source string) (*Component, error)
func Validate(c *Component) error

// Package generate.
type Result struct { Source string; Warnings []string }
func Generate(c *contract.Component) (Result, error)

// Package cli.
const Version = "0.1.0"
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int
```

`Component.Name` is the PascalCase interface stem; do not guess it from a filename or shell arguments. Node DOM types come from a converter-owned mapping for the native tags accepted by the grammar, not the sibling catalog. Tests use synthetic contracts; arbitrary valid interface stems and factory names must work without consulting a component list.

Use `json.RawMessage` only for decoded literal defaults, never for unchecked source fragments. `nil` means no default. Validate types before rendering them into script. Keep generated symbols in a reserved internal naming convention and allocate them through a symbol table; source props named `emit`, `props`, or `ref` must not collide with helpers.

Go tests import `html-ui-to-vue-vapor/internal/testinput`, which contains small hand-written source constants. Production packages never import this test-support package. Shared helpers `mustParse` and `mustGenerate` are test-only functions created in the test file that uses them, with these signatures:

```go
func mustParse(t *testing.T, source string) *contract.Component
func mustGenerate(t *testing.T, c *contract.Component) generate.Result
```

In package `contract` tests, `mustParse` returns `*Component` instead of qualifying its own package. In package `generate` tests, `mustGenerate` returns `Result`. Do not create production helper APIs just to satisfy tests.

---

## Task 1: Establish the ordered contract model and lexer

**Files:** Create `go.mod`, `.gitignore`, `internal/contract/model.go`, `internal/contract/lexer.go`, `internal/contract/lexer_test.go`, `internal/testinput/source.go`.

**Consumes:** The documented source contract syntax; no producer executable or catalog.

**Produces:** The shared model above, `Diagnostic.Error()`, and package-private `lex(source string) ([]token, error)`, where `token` contains `Kind`, decoded `Text`, and `Pos`. The scanner emits identifiers, strings, numbers, punctuation, documentation comments, and EOF; parser tasks share these names.

- [ ] Create `go.mod` with `module html-ui-to-vue-vapor` and `go 1.24.0`. Exclude `.test-output/`, `node_modules/`, `test-results/`, `playwright-report/`, and `bin/`.
- [ ] Add minimal hand-written input constants to `internal/testinput/source.go`. These exercise the language features without depending on named producer primitives:

```go
package testinput

const Disclosure = `/** Behavior: native. */
export const contractVersion = 1 as const;
export interface DisclosureProps { name?: string; open?: boolean; }
export const defaults = { open: false } as const satisfies Partial<DisclosureProps>;
export interface DisclosureSlots<Content = HTMLElement> {
  summary: () => Content;
  content?: (scope: { open: boolean }) => Content;
}
export interface DisclosureEvents { toggle: ToggleEvent; }
export const nativeEvents = { toggle: { target: "root", state: { open: "open" } } } as const;
export function disclose(props: DisclosureProps, slots: DisclosureSlots<HTMLElement>): HTMLDetailsElement {
  const { name, open = defaults.open } = props;
  const root = document.createElement("details");
  const summary = document.createElement("summary");
  if (name !== undefined) root.setAttribute("name", name);
  root.open = open;
  summary.append(slots.summary());
  root.append(summary);
  if (slots.content !== undefined) root.append(slots.content({ open }));
  return root;
}`

const Checkbox = `/** Behavior: native. */
export const contractVersion = 1 as const;
export interface BinaryProps { checked?: boolean; disabled?: boolean; }
export const defaults = { checked: false, disabled: false } as const satisfies Partial<BinaryProps>;
export interface BinarySlots<Content = HTMLElement> { label: () => Content; }
export interface BinaryEvents { input: Event; change: Event; }
export const nativeEvents = {
  input: { target: "control", state: { checked: "checked" } },
  change: { target: "control", state: { checked: "checked" } }
} as const;
export function binary(props: BinaryProps, slots: BinarySlots<HTMLElement>): HTMLLabelElement {
  const { checked = defaults.checked, disabled = defaults.disabled } = props;
  const root = document.createElement("label");
  const control = document.createElement("input");
  control.setAttribute("type", "checkbox");
  control.disabled = disabled;
  control.defaultChecked = checked;
  control.checked = checked;
  root.append(slots.label());
  root.append(control);
  return root;
}`

const Static = `/** Behavior: native. */
export const contractVersion = 1 as const;
export interface ExampleProps {}
export const defaults = {} as const satisfies Partial<ExampleProps>;
export interface ExampleSlots<Content = HTMLElement> {}
export interface ExampleEvents {}
export const nativeEvents = {} as const;
export function example(props: ExampleProps, slots: ExampleSlots<HTMLElement>): HTMLDivElement {
  const root = document.createElement("div");
  return root;
}`
```

Tests construct an adapter-marked example with `strings.Replace(testinput.Static, "Behavior: native", "Behavior: adapter-required", 1)`. Introduce additional tiny inline input strings only to pin a specific parser/runtime regression; no source files are captured or catalog is maintained.

- [ ] Write lexer regressions before implementation:

```go
func TestLexerPreservesLiteralAndCRLFPosition(t *testing.T) {
    tokens, err := lex("// heading\r\nexport const x = \"a\\\"&é\";")
    if err != nil { t.Fatal(err) }
    if tokens[0].Text != "export" || tokens[0].Pos.Line != 2 || tokens[0].Pos.Column != 1 {
        t.Fatalf("bad location: %#v", tokens[0])
    }
    found := false
    for _, tok := range tokens { if tok.Kind == "string" && tok.Text == "a\"&é" { found = true } }
    if !found { t.Fatal("decoded literal missing") }
}
func TestLexerRejectsTruncatedString(t *testing.T) {
    _, err := lex("export const x = \"unterminated")
    var diagnostic *Diagnostic
    if !errors.As(err, &diagnostic) || diagnostic.Pos.Line != 1 {
        t.Fatalf("expected located lexical error, got %v", err)
    }
}
```

Also table-test escaped backslashes, `\u` escapes, line/block comments, unterminated block comments, negative numeric literals as separate tokens, and comment text containing executable-looking strings.

- [ ] Run `go test ./internal/contract -run Lexer -v`; expect undefined lexer/model symbols to fail compilation.
- [ ] Implement UTF-8-aware positions, string decoding for the generated JavaScript/JSON-compatible literal forms, comments, punctuation, numeric tokens, and explicit EOF. Retain leading documentation comments for behavior classification; skip ordinary comments during parsing. Do not use a regex to parse nested structures.
- [ ] Run `go test ./internal/contract -run Lexer -v`; expect all regressions to pass. Format with `gofmt -w internal/contract`.
- [ ] Review the model and lexical boundaries; commit this task only if a converter-owned Git repository exists.

## Task 2: Parse declarations, defaults, and native event metadata

**Files:** Create `internal/contract/parser.go`, `declarations.go`, and `declarations_test.go`; extend `model.go` only if the shared model needs a documented correction.

**Consumes:** `lex`, `token`, and the shared model from Task 1.

**Produces:** Package-private `newParser(source string) (*parser, error)` and `parseDeclarations(p *parser) (*Component, error)`. On success, the cursor points to the factory's `export function` declaration; Task 3 consumes it. `parser` has `tokens []token` and `index int`, plus located `peek`, `take`, and `expect` helpers.

- [ ] Write metadata tests using the hand-written disclosure input:

```go
func TestDeclarationsPreserveContract(t *testing.T) {
    p, err := newParser(testinput.Disclosure)
    if err != nil { t.Fatal(err) }
    c, err := parseDeclarations(p)
    if err != nil { t.Fatal(err) }
    if c.Version != 1 || c.Name != "Disclosure" || c.Behavior != "native" { t.Fatalf("%#v", c) }
    if len(c.Props) != 2 || c.Props[1].Name != "open" || string(c.Props[1].Default) != "false" { t.Fatalf("%#v", c.Props) }
    if c.Slots[0].Optional || !c.Slots[1].Optional || c.Slots[1].Scope[0].Type != "boolean" { t.Fatalf("%#v", c.Slots) }
    if c.Events[0].Target != "root" || c.Events[0].State[0].Property != "open" { t.Fatalf("%#v", c.Events) }
}
```

Add table tests for empty interfaces, literal unions, negative numeric/default string literals, duplicate interfaces/fields/defaults, mismatched interface stems, unknown version, unknown exported constants, and an event present in only one of `Events` or `nativeEvents`.

- [ ] Run `go test ./internal/contract -run Declarations -v`; expect absent parser functions to fail.
- [ ] Implement recursive descent for these exact declaration shapes:

```text
export const contractVersion = 1 as const;
export interface <Stem>Props { <identifier>[?]: <prop-type>; ... }
export const defaults = { <identifier>: <literal>, ... } as const satisfies Partial<<Stem>Props>;
export interface <Stem>Slots<Content = HTMLElement> {
  <identifier>[?]: ([scope: { <identifier>: <prop-type>; ... }]) => Content; ...
}
export interface <Stem>Events { <identifier>: <DOM-event-type>; ... }
export const nativeEvents = {
  <identifier>: { target: <string>, state: { <identifier>: <string>, ... } }, ...
} as const;
```

Accept trailing commas and the current semicolon convention. Prop types are `string`, `number`, `boolean`, or unions of quoted literals. Native event types currently include `Event`, `KeyboardEvent`, `MouseEvent`, `SubmitEvent`, and `ToggleEvent`; validate them as DOM types rather than copying arbitrary type expressions into output. Validate literal defaults against their prop types and literal unions. Store the source header's `Behavior: native` or `Behavior: adapter-required` as metadata; an unclassified header yields an explicit unknown classification, not an assumed native claim.

- [ ] Run `go test ./internal/contract -run Declarations -v`; expect all metadata tests to pass.
- [ ] Review interface/default/event consistency and commit when applicable.

## Task 3: Parse and validate the native factory

**Files:** Create `internal/contract/factory.go`, `validate.go`, and `parser_test.go`; extend `parser.go` with public `Parse`.

**Consumes:** Task 2's parser cursor and component metadata.

**Produces:** `Parse(source string) (*Component, error)` and `Validate(c *Component) error`. `Parse` calls `parseDeclarations`, parses exactly one factory, requires EOF, then calls `Validate`.

- [ ] Write the positive and negative factory tests:

```go
func TestParseSyntheticInputsAndArbitraryNames(t *testing.T) {
    renamed := strings.ReplaceAll(testinput.Disclosure, "Disclosure", "CustomSection")
    renamed = strings.Replace(renamed, "function disclose", "function customSection", 1)
    for name, source := range map[string]string{
        "disclosure": testinput.Disclosure,
        "checkbox": testinput.Checkbox,
        "static": testinput.Static,
        "custom-name": renamed,
    } {
        t.Run(name, func(t *testing.T) {
            c, err := Parse(source); if err != nil { t.Fatal(err) }
            if c.Root != "root" || len(c.Nodes) == 0 { t.Fatalf("%#v", c) }
            if name == "custom-name" && c.Name != "CustomSection" { t.Fatalf("%#v", c) }
        })
    }
}
func TestParseRejectsUnknownCall(t *testing.T) {
    source := strings.Replace(testinput.Disclosure, "  return root;", "  arbitraryCall();\n  return root;", 1)
    _, err := Parse(source)
    var diagnostic *Diagnostic
    if !errors.As(err, &diagnostic) { t.Fatalf("expected located rejection, got %v", err) }
}
```

Add negative cases for missing return, trailing statements, duplicate nodes, unknown props, removed optional guards, changed default destructuring, missing slots, scope type mismatches, graph cycles, repeated child insertion, invalid native event targets/properties, arbitrary factory parameter defaults, and declaration-only input. Add a synthetic slot alias `{ isOpen: open }` and assert the `ScopeBinding` stores both names.

- [ ] Run `go test ./internal/contract -run 'Parse|Validate' -v`; expect missing public parser functions to fail.
- [ ] Implement the factory grammar and emit ordered model entries:

```text
export function <factory>(props: <Stem>Props, slots: <Stem>Slots<HTMLElement>): <RootType> {
  const { <prop>[ = defaults.<prop>], ... } = props;
  const <node> = document.createElement(<literal-tag>);
  <node>.setAttribute(<literal-name>, <literal-value>);
  [if (<prop> !== undefined)] <node>.<property> = <prop>;
  [if (<prop> !== undefined)] <node>.setAttribute(<literal-name>, <prop> | String(<prop>));
  <parent>.append(<node>);
  [if (slots.<slot> !== undefined)] <parent>.append(slots.<slot>([ { <field> | <field>: <prop>, ... } ]));
  return root;
}
```

Guard presence must match optionality/defaults. All nodes and slots must be placed exactly once in a single rooted tree. Validate native tags, attribute/property names, and DOM property read types; reject template directive names and event-handler attributes that would change the generated Vue program. Infer tag DOM types from a converter-owned tag table and verify the declared root return type. Resolve event interface entries to native event metadata without introducing additional runtime state.

- [ ] Run `go test ./internal/contract -v` and `go test ./internal/contract -race`; expect pass for synthetic inputs, arbitrary component names, and negative cases.
- [ ] Review the complete parsing boundary and commit when applicable.

## Task 4: Generate static structure and typed Vue contracts

**Files:** Create `internal/generate/generate.go`, `template.go`, `script.go`, `escape.go`, and `generate_test.go`.

**Consumes:** `contract.Component` and `contract.Validate`.

**Produces:** `Generate(c *contract.Component) (Result, error)` and the `Result` struct defined above. The initial implementation emits typed props/slots and static/prop-bound templates; Task 5 adds native events and mapped state.

- [ ] Write structure, optional attribute, and escaping tests:

```go
func TestGeneratePreservesDisclosureStructure(t *testing.T) {
    c := mustParse(t, testinput.Disclosure)
    result := mustGenerate(t, c)
    if !strings.Contains(result.Source, `<script setup lang="ts" vapor>`) { t.Fatal(result.Source) }
    summary := strings.Index(result.Source, `<summary>`)
    content := strings.Index(result.Source, `<slot name="content"`)
    if summary < 0 || content <= summary { t.Fatal(result.Source) }
    if !strings.Contains(result.Source, `open?: boolean`) { t.Fatal(result.Source) }
}
func TestGenerateEscapesScriptAndHTMLContexts(t *testing.T) {
    c := mustParse(t, testinput.Disclosure)
    c.Nodes[0].Attributes = append(c.Nodes[0].Attributes, contract.Attribute{Name: "data-label", Value: `a"&é</script>`})
    result := mustGenerate(t, c)
    if strings.Count(result.Source, "</script>") != 1 { t.Fatal("literal escaped out of SFC script") }
    if !strings.Contains(result.Source, "&amp;") || !strings.Contains(result.Source, "&quot;") { t.Fatal(result.Source) }
}
```

Add tests for slot names and alias fields, void elements, property `.prop` bindings, explicit attribute `.attr` bindings, omitted undefined optional attributes, escaped string defaults containing `</script>`, and helper name collisions. Use complete representative golden SFCs only where the exact public output is the assertion; do not snapshot every internal helper.

- [ ] Run `go test ./internal/generate -run Generate -v`; expect missing generator symbols to fail.
- [ ] Render the script contract using compiler macros, with defaults embedded as validated literals:

```vue
<script setup lang="ts" vapor>
interface Props { name?: string; open?: boolean }
const __htmlUiProps = withDefaults(defineProps<Props>(), { open: false })
defineSlots<{
  summary: () => unknown
  content?: (scope: { open: boolean }) => unknown
}>()
</script>
```

`unknown` deliberately replaces native `HTMLElement` content; Task 7/8 must verify that this return type works with the pinned Vue compiler and language tools. If language tools require a different renderer return type, use a type-only Vue renderer type with a regression test and record that choice in compatibility documentation. Never call a slot to inspect or render its contents in setup.

Traverse the rooted graph in child order. Preserve each node's binding order. Encode source literals with JSON-compatible escaping and HTML-escape attribute values and template expressions. Map optional numeric/boolean attribute expressions to `value === undefined ? undefined : String(value)`, rather than serializing `undefined`. Distinguish boolean native properties from ARIA strings. Allocate helper symbols so user contract names cannot shadow them.

- [ ] Run `go test ./internal/generate -run Generate -v`; expect structural/escaping regressions to pass.
- [ ] Review emitted markup and commit when applicable. Do not claim runtime behavior until Task 5 and browser checks pass.

## Task 5: Generate local state, native event forwarding, and reset synchronization

**Files:** Create `internal/generate/state.go`, `events.go`, `reset.go`, and `state_test.go`; extend `script.go` and `template.go`.

**Consumes:** Validated `Component.Events`, bindings, and prop/slot metadata.

**Produces:** Complete SFC behavior from `Generate`, plus package-private `inferState(c *contract.Component) []stateField`. `stateField` contains `Prop`, `Type`, `Default`, `Reads []stateRead`; `stateRead` contains `Target`, `DOMType`, `Property`. Deduplicate fields and identical target/property reads in source order.

- [ ] Write inference tests:

```go
func TestStateInferenceUsesNativeMappingsOnly(t *testing.T) {
    c := mustParse(t, testinput.Checkbox)
    fields := inferState(c)
    if len(fields) != 1 || fields[0].Prop != "checked" { t.Fatalf("%#v", fields) }
    result := mustGenerate(t, c)
    if strings.Count(result.Source, "ref<boolean>(") != 1 { t.Fatal(result.Source) }
    if !strings.Contains(result.Source, `"update:checked"`) { t.Fatal(result.Source) }
    if !strings.Contains(result.Source, "event.currentTarget") { t.Fatal(result.Source) }
}
```

Also assert accordion exposes a live state-backed content slot; slider/number-field use `valueAsNumber`; input/change share one state ref; unmapped defaults do not create refs; no generated handler reads `event.target`; and reset helpers include cancellation and disposal guards. Behavioral regressions are exercised by Task 8 instead of relying on helper-text assertions alone.

- [ ] Run `go test ./internal/generate -run 'State|Event|Reset' -v`; expect missing inference/runtime generation to fail.
- [ ] Generate typed emits, refs, per-prop watches, and handlers following this concrete pattern:

```ts
const __htmlUiEmit = defineEmits<{
  toggle: [event: ToggleEvent]
  "update:open": [value: boolean]
}>()
const __htmlUiStateOpen = ref<boolean>(__htmlUiProps.open)
watch(() => __htmlUiProps.open, value => { __htmlUiStateOpen.value = value })
function __htmlUiOnToggle(event: ToggleEvent) {
  const node = event.currentTarget as HTMLDetailsElement
  const next = node.open
  if (!Object.is(__htmlUiStateOpen.value, next)) {
    __htmlUiStateOpen.value = next
    __htmlUiEmit("update:open", next)
  }
  __htmlUiEmit("toggle", event)
}
```

Bind mapped live properties to state refs and scoped slots to current state. Bind other props through the props object. Optional mapped props without defaults retain `undefined` in the initial ref type; actual DOM reads may produce sanitized values or `NaN`. Forward original events once, even when the mapped state did not change; emit model updates only for changed state. Emit events with no mapped fields without inventing local state.

- [ ] Keep reset baselines separate from live refs. Capture initial `defaultValue`/`defaultChecked` bindings once when supplied by the source; parent live prop updates must not overwrite those baselines. For mutable fields originally bound through a `value` attribute, preserve the initial attribute as the reset baseline and use the live DOM property after initialization. Initialize min/max/step before value and verify actual sanitization in Task 8.
- [ ] Generate template refs for mutable native targets, mount-time DOM readback for sanitized initial values, and a reset subscription to each target's associated form. Use a microtask because reset is dispatched before native values are restored. Check `event.defaultPrevented` inside the microtask, check a disposed flag, then read the mapped DOM properties. Remove every installed listener on unmount and invalidate queued callbacks. If form association changes, refresh the subscription on component updates; document that arbitrary external DOM movement without any component update is outside the guaranteed synchronization boundary.

```ts
let __htmlUiDisposed = false
function __htmlUiOnReset(event: Event) {
  queueMicrotask(() => {
    if (__htmlUiDisposed || event.defaultPrevented) return
    __htmlUiReadNativeState()
  })
}
onBeforeUnmount(() => {
  __htmlUiDisposed = true
  __htmlUiDetachResetListeners()
})
```

`__htmlUiReadNativeState` and `__htmlUiDetachResetListeners` are generated per component in `reset.go`; they do not require a shared runtime package. DOM readback during mount does not emit fabricated input/change/toggle events; reset readback emits changed model updates so a named `v-model` remains synchronized.

- [ ] Run `go test ./internal/generate -v` and `go test ./internal/generate -race`; expect inference/generation tests to pass.
- [ ] Review state ownership, event ordering, and reset lifecycle; commit when applicable.

## Task 6: Deliver the stream-only CLI

**Files:** Create `internal/cli/cli.go`, `cli_test.go`, and `cmd/html-ui-to-vue-vapor/main.go`.

**Consumes:** `contract.Parse` and `generate.Generate`.

**Produces:** `cli.Run(args, stdin, stdout, stderr) int`, `cli.Version`, and a buildable binary.

- [ ] Write command tests:

```go
func TestInvalidInputDoesNotWriteStdout(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := Run(nil, strings.NewReader("export const contractVersion = 999 as const;"), &stdout, &stderr)
    if code != 2 || stdout.Len() != 0 || stderr.Len() == 0 { t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String()) }
}
func TestAdapterWarningStaysOnStderr(t *testing.T) {
    source := strings.Replace(testinput.Static, "Behavior: native", "Behavior: adapter-required", 1)
    var stdout, stderr bytes.Buffer
    code := Run(nil, strings.NewReader(source), &stdout, &stderr)
    if code != 0 || !strings.Contains(stdout.String(), "<template>") || !strings.Contains(stderr.String(), "adapter-required") { t.Fatalf("%d %q %q", code, stdout.String(), stderr.String()) }
}
```

Add help/version tests with an input reader that fails if touched; reject unknown flags and positional arguments; test empty/truncated/CRLF input, a reader returning an error, a writer returning an error, and a short writer returning fewer bytes without an error. Input/output failures return 1. Successful conversion writes only one complete SFC and preserves warning order.

- [ ] Run `go test ./internal/cli -v`; expect missing `Run` to fail.
- [ ] Implement argument selection before reading input, `io.ReadAll`, parse, generate, warning emission, and a checked `io.WriteString`. CLI diagnostics start with `html-ui-to-vue-vapor:` and include located parser errors. Generate an SFC comment identifying adapter-required limitations, plus `Result.Warnings` for stderr; unknown behavior classification gets an explicit unclassified diagnostic. The component generator remains independent of shell/process APIs.

```go
func main() {
    os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
```

- [ ] Run `go test ./... -v` and `go build -o bin/html-ui-to-vue-vapor ./cmd/html-ui-to-vue-vapor`.
- [ ] Run the CLI success test with `testinput.Disclosure` on stdin and inspect its SFC output; parse errors must never produce partial stdout. Review and commit when applicable.

## Task 7: Compile synthetic inputs with the pinned Vapor toolchain

**Files:** Create `package.json`, `bun.lock`, `tests/support/compile.ts`, `tests/compiler.test.ts`, and converter-owned generation output under `.test-output/`.

**Consumes:** The Go binary from Task 6 and small hand-written source strings.

**Produces:** `compileSFC(source: string, filename: string): string` in `tests/support/compile.ts`; returns compiled JavaScript/TypeScript module source. Export `convert(source: string): { source: string; stderr: string }` from the same support module, invoking only the converter binary with source on stdin. Scripts: `test:compiler`, `test:types`, `test:browser`, and `test`.

- [ ] Write a real compiler integration check before installing/configuring its dependencies:

```ts
import { expect, test } from 'bun:test'
import { compileSFC, convert } from './support/compile'
const source = `/** Behavior: native. */
export const contractVersion = 1 as const;
export interface CustomProps { open?: boolean; }
export const defaults = { open: false } as const satisfies Partial<CustomProps>;
export interface CustomSlots<Content = HTMLElement> { summary: () => Content; }
export interface CustomEvents { toggle: ToggleEvent; }
export const nativeEvents = { toggle: { target: "root", state: { open: "open" } } } as const;
export function custom(props: CustomProps, slots: CustomSlots<HTMLElement>): HTMLDetailsElement {
  const { open = defaults.open } = props;
  const root = document.createElement("details");
  const summary = document.createElement("summary");
  root.open = open;
  summary.append(slots.summary());
  root.append(summary);
  return root;
}`
test('converts an arbitrary input contract into a Vapor component', () => {
  const output = convert(source)
  const code = compileSFC(output.source, 'Custom.vue')
  expect(code).toContain('defineVaporComponent')
})
```

Add compiler regressions for escaped script defaults, slot aliases, static separator with empty interfaces, native `.prop` bindings, and adapter warnings. Compiling a component must not execute its source factory.

- [ ] Run `bun test tests/compiler.test.ts`; expect missing helper/module errors.
- [ ] Add exact devDependencies and install them inside this directory. Pin Vue/compiler-sfc to `3.6.0-rc.9`, TypeScript to `7.0.2`, vue-tsc to `3.3.11`, and Playwright to `1.63.0`. Verify the pinned Vue type checker supports TypeScript 7; if it does not, use an officially supported Vue type-checking tool/version and record its exact pin. Keep TypeScript at `7.0.2`; report an unavailable compatible Vue checker as an unverified type-checking gate rather than silently downgrading TypeScript. Commit `bun.lock`; do not use the sibling package.json or node_modules as the converter's dependency contract.
- [ ] Implement `compileSFC` using the official SFC parser and `compileScript` with an inline template:

```ts
import { parse, compileScript } from '@vue/compiler-sfc'
export function compileSFC(source: string, filename: string): string {
  const { descriptor, errors } = parse(source, { filename })
  if (errors.length) throw new Error(errors.map(String).join('\n'))
  if (!descriptor.scriptSetup || !Object.hasOwn(descriptor.scriptSetup.attrs, 'vapor')) {
    throw new Error('missing vapor marker')
  }
  return compileScript(descriptor, { id: filename, inlineTemplate: true }).content
}
```

The `vapor` marker may be represented by an empty attribute string; check attribute presence. Pass a stable component id. Use `execFileSync` or `spawnSync` with structured argument arrays and explicit stdin to call the converter; never interpolate source into a shell command. Build the converter to `.test-output/html-ui-to-vue-vapor` before checks.

- [ ] Run `bun run test:compiler`; expect the synthetic input cases to pass with actual Vapor compilation. If compiler behavior differs from the pinned API, verify the tagged upstream compiler source, fix the helper, and keep the positive/negative compiler assertions.
- [ ] Add a renamed-interface/factory compiler case and assert it converts and compiles identically in behavior. Neither the test harness nor the binary may enumerate a component catalog or invoke the producer.
- [ ] Review compiler compatibility and commit when applicable.

## Task 8: Verify consumer types and browser behavior

**Files:** Create `tests/typecheck.test.ts`, `tests/types/consumer.vue`, `tests/types/tsconfig.json`, `playwright.config.ts`, `tests/browser.pw.ts`, `tests/support/harness.ts`, and `tests/serve.ts`. Fix generator files only when these checks expose failures.

**Consumes:** Task 7's `compileSFC`/`convert`, SFCs generated from the synthetic test inputs, and the complete runtime generator.

**Produces:** Strict type-check results and browser verification across Chromium, Firefox, and WebKit in pure Vapor mode; a representative accordion/input test also verifies VDOM interop.

- [ ] Create a consumer that uses accordion with `v-model:open`, typed `toggle`, required summary, optional content, and the boolean scoped `open`. Check both valid usage and expected-invalid usage:

```vue
<script setup lang="ts">
import { ref } from 'vue'
import Accordion from '../../.test-output/types/Accordion.vue'
const open = ref(false)
function onToggle(event: ToggleEvent) { void event.newState }
</script>
<template>
  <Accordion v-model:open="open" @toggle="onToggle">
    <template #summary>Heading</template>
    <template #content="scope">{{ scope.open ? 'open' : 'closed' }}</template>
  </Accordion>
</template>
```

Use a strict tsconfig with target ES2022, ESNext modules, Bundler resolution, ES2022/DOM libs, no emit, and strict Vue template checking. Type-check every generated SFC, not just the consumer. Add separate negative consumer examples with wrong boolean/union props, missing required summary, wrong model update payload, unknown slot field, and wrong native event payload. Assert negative checks fail for the expected diagnostic categories; do not accept any unrelated failure as proof of type coverage. Verify public declarations preserve required and optional slot signatures as well as runtime consumption.

- [ ] Run `bun run test:types`; expect pass for valid usage and intentional, categorized failures for invalid consumer examples. Fix the generated contract types if needed.
- [ ] Implement the browser harness around these exact helper interfaces:

```ts
// tests/support/harness.ts
export type Mode = 'vapor' | 'interop'
export async function buildHarnesses(): Promise<void>
// Writes browser bundles/pages into .test-output/browser.
export async function openCase(page: import('@playwright/test').Page, name: string, mode: Mode = 'vapor'): Promise<void>
// Navigates to http://127.0.0.1:4175/<name>-<mode>.html and awaits window.__htmlUiReady.
```

Generate parent SFC harnesses with typed named slots and refs; compile them and the primitive SFCs through `compileSFC`, then bundle with `Bun.build`. Harness entries use `createVaporApp(App).mount(...)` in pure mode or `createApp(App).use(vaporInteropPlugin).mount(...)` in interop mode. No Vite/plugin dependency is needed. A local Bun static server serves only `.test-output/browser`. Playwright starts it via `webServer`, uses one worker, retains failure traces, and runs all three browser projects.

- [ ] Write the disclosure regression:

```ts
import { expect, test } from '@playwright/test'
import { openCase } from './support/harness'
test('native toggle updates slot state and named model', async ({ page }) => {
  await openCase(page, 'accordion')
  await expect(page.locator('[data-slot-state]')).toHaveText('closed')
  await page.locator('summary').click()
  await expect(page.locator('[data-slot-state]')).toHaveText('open')
  await expect(page.locator('[data-parent-model]')).toHaveText('true')
})
```

Harness output uses `[data-slot-state]` for scoped content and `[data-parent-model]` for the parent's model. Add `data-action` buttons that mutate only an unrelated name prop, mutate the mapped prop, cancel reset, and unmount the component so tests can steer behavior through UI.

- [ ] Add these independently named browser checks and run them first to confirm failures identify behavior rather than harness errors:

  - `accordion-group`: opening the second named details closes the first and updates both scoped slots.
  - `parent-sync`: user-opened details survives a name-only update; an actual open prop change updates the DOM.
  - `checkbox-events`: input/change share one checked ref, each forwarded once, with the original currentTarget.
  - `nested-event`: a descendant-originating event reads the mapped listener node, not the originating child.
  - `input-reset`: an initial text baseline survives user input and later live prop changes; reset updates DOM and named model.
  - `checkbox-reset`: checked/defaultChecked reset correctly.
  - `number-empty`: cleared numeric input yields `NaN`, not zero; use `Number.isNaN` in the harness readout.
  - `slider-sanitization`: min/max/step initialize before out-of-range value and local state reflects the browser's sanitized value.
  - `reset-cancel`: canceled native reset changes neither DOM nor model.
  - `reset-dispose`: reset queued immediately before unmount emits no later model update; remount creates one subscription.
  - `optional-attributes`: absent optional attributes stay absent; numeric attributes and `aria-pressed="false"` use their correct string values.
  - `event-once`: native events are forwarded once through declared emits, including interop mode.
  - `baseline-association`: a target's form association changed by component update detaches from the old form and attaches to the new form.

- [ ] Run `bun run test:browser` across all three engines and the selected interop cases. Fix failures in the responsible generator unit; rerun affected Go/compiler/type tests after each runtime change.
- [ ] Review public typing and native behavior together and commit when applicable.

## Task 9: Document, check standalone execution, and run final verification

**Files:** Create `Makefile`, `README.md`, `docs/compatibility.md`, and `scripts/check-standalone.sh`.

**Consumes:** Passing Go, compiler, type, and browser checks from Tasks 1–8.

**Produces:** Install/build instructions, complete verification command, clear compatibility limits, and a standalone conversion smoke check.

- [ ] Write a standalone test script that builds the Go binary, supplies a minimal hand-written input through stdin, clears inherited tool discovery, and checks output without invoking Node/TypeScript/producer:

```sh
#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p .test-output
converter_root="$(pwd)"
go build -o "$converter_root/.test-output/html-ui-to-vue-vapor" ./cmd/html-ui-to-vue-vapor
PATH=/nonexistent "$converter_root/.test-output/html-ui-to-vue-vapor" \
  > "$converter_root/.test-output/standalone.vue" <<'SOURCE'
/** Behavior: native. */
export const contractVersion = 1 as const;
export interface StandaloneProps {}
export const defaults = {} as const satisfies Partial<StandaloneProps>;
export interface StandaloneSlots<Content = HTMLElement> {}
export interface StandaloneEvents {}
export const nativeEvents = {} as const;
export function standalone(props: StandaloneProps, slots: StandaloneSlots<HTMLElement>): HTMLDivElement {
  const root = document.createElement("div");
  return root;
}
SOURCE
test -s "$converter_root/.test-output/standalone.vue"
```

- [ ] Run `sh scripts/check-standalone.sh`; expect a valid nonempty SFC. Verify the binary has no shell/process execution path with `rg -n 'os/exec|exec\.Command' internal cmd`; expect no production matches.
- [ ] Document these commands and include an accordion consumer with summary/content slots and named `v-model:open`:

```sh
go build -o bin/html-ui-to-vue-vapor ./cmd/html-ui-to-vue-vapor
go install ./cmd/html-ui-to-vue-vapor
html-ui accordion | html-ui-to-vue-vapor > Accordion.vue
html-ui accordion --ts | html-ui-to-vue-vapor > Accordion.vue
```

Explain source versus declarations, state inference versus defaults, later parent updates, form reset baselines, original event forwarding, Vapor application/interop setup, CLI error codes, and unsupported syntax diagnostics. Explain adapter-required input metadata and the interaction responsibilities that cannot be inferred from the supplied contract; do not maintain a list of primitive names. Describe the pinned tested versions without claiming general compatibility with untested Vue releases. Link the official [Vue Vapor release notes](https://github.com/vuejs/core/releases/tag/v3.6.0-rc.9), [script setup macros](https://vuejs.org/api/sfc-script-setup.html), and [Vue language tools releases](https://github.com/vuejs/language-tools/releases).

- [ ] Add `make check` with these commands, running from this directory:

```sh
go test ./...
go test -race ./...
go vet ./...
bun run test:compiler
bun run test:types
bun run test:browser
sh scripts/check-standalone.sh
```

`bun install --frozen-lockfile` and browser installation are documented prerequisites, not silent steps inside the Go CLI. Build or refresh `.test-output` artifacts from checks; never depend on a developer's old generated files.

- [ ] Run `make check`. Record the actual package versions, browser coverage, synthetic input features exercised, and exit results in `docs/compatibility.md`. If a required check cannot run because dependencies/network/browser binaries are unavailable, report it as unverified instead of claiming the compatibility requirement passed.
- [ ] Confirm the producer remains unchanged and every new/modified file belongs to this converter. Review the complete diff, update documentation to match the final implementation, and commit only in the converter's own repository when present.

## Execution order and review gates

Execute Tasks 1–9 in order. Tasks 2/3 share parser interfaces; Tasks 4/5 share generator interfaces; integration checks depend on the real CLI. This plan favors native execution in the current session because these boundaries are sequential and splitting each task into new agent contexts adds coordination overhead. Use subagents only if the user selects that execution approach or an applicable execution skill requires a reviewer.

During implementation, follow the selected execution skill's required review workflow. Do not mark a task complete merely because it produces the expected text: compiler/type/browser tasks are mandatory acceptance gates. Run architectural and documentation consistency review after code changes as required by the selected execution workflow.

## Plan self-review

- Scope: every created/modified file is inside the converter; the producer is outside implementation and test dependencies.
- Spec coverage: lexical/contract validation (1–3), Vue structure/types/escaping (4), event state and native reset (5), stream command (6), synthetic input/Vapor compiler checks (7), strict consumer typing and browser behavior (8), packaging/docs/final checks (9).
- Interface consistency: `Parse`, `Validate`, `Generate`, `Result`, `Run`, `compileSFC`, `convert`, and browser harness names/signatures are defined above and used consistently.
- Review focus: all five listed regressions have owning tests and behavioral checks.
- Implementation has not begun; this document is the reviewable execution plan.
