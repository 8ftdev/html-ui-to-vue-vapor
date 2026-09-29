# html-ui-to-vue-vapor

A standalone Go CLI that converts an `html-ui` TypeScript contract on stdin into a typed Vue Vapor component on stdout.

```sh
go install ./cmd/html-ui-to-vue-vapor
html-ui accordion | html-ui-to-vue-vapor > Accordion.vue
```

For a local binary:

```sh
make build
./bin/html-ui-to-vue-vapor < primitive.ts > Primitive.vue
```

The converter derives everything from its input. It has no primitive catalog, does not invoke `html-ui`, and never executes the incoming TypeScript. Node, Bun, and TypeScript are not needed to run the binary.

The input is the complete source contract, including its DOM factory. A declaration file alone does not preserve the factory's structure or slot placement. `html-ui accordion --ts` works as input too.

## Generated components

Output uses `<script setup lang="ts" vapor>`, typed props and defaults, named/scoped slots, and original typed native events. DOM bindings use self-contained Vapor directives to preserve ordered attribute/property writes and native form reset baselines. No shared converter runtime or CSS is required.

A field becomes local reactive state when `nativeEvents` maps an event to a DOM property for that field. A default alone does not imply local mutable state. Native events update that state and emit `update:<field>`, enabling named models:

```vue
<script setup lang="ts" vapor>
import { ref } from 'vue'
import Accordion from './Accordion.vue'
const open = ref(false)
</script>

<template>
  <Accordion v-model:open="open">
    <template #summary>Details</template>
    <template #content="scope">
      <p>{{ scope.open ? 'Expanded' : 'Collapsed' }}</p>
    </template>
  </Accordion>
</template>
```

Later changes to a mapped parent prop update local state. Unrelated parent updates preserve local edits. Form reset restores the initial native baseline and synchronizes the model; later live prop changes do not replace that baseline. Numeric controls read `valueAsNumber`, including `NaN` for an empty number input.

Mount pure Vapor applications with `createVaporApp(App).mount('#app')`. To use Vapor components in a VDOM application, mount with `createApp(App).use(vaporInteropPlugin).mount('#app')`; both helpers come from `vue`.

## Input and diagnostics

The parser accepts contract versions 1 and 2: Props, literal defaults, Slots, Events, `nativeEvents`, and the constrained native factory syntax. It supports literal element creation, static attributes, prop attribute/property bindings, optional guards, ordered child insertion, scoped slot mappings, and a returned root. It accepts arbitrary interface and factory names within that syntax. It is not a general TypeScript compiler.

Invalid input produces a line/column diagnostic on stderr and no component output. Exit codes are 0 for conversion, 2 for invalid arguments/input, and 1 for stream failures. `--help`, `-h`, and `--version` are available.

Inputs marked `adapter-required` generate their declared native structure and event/state behavior, along with a warning on stderr and a comment in the SFC. Focus management, keyboard selection, dismissal, and other behavior absent from the contract cannot be inferred. Unclassified input also carries a diagnostic rather than claiming complete interactions.

## Development

```sh
bun install --frozen-lockfile
bunx --no-install playwright install chromium firefox webkit
make check
```

Checks combine small hand-written input strings with all 37 real default producer contracts and a v2 accordion fixture. The converter binary itself has no producer dependency. Development dependencies belong to this converter directory.

The pinned runtime/compiler target is Vue `3.6.0-rc.9`. The primary compiler is TypeScript `7.0.2`. Vue-specific type checks use `vue-tsc` with the official `@typescript/typescript6` JavaScript API compatibility package; TypeScript 7 separately checks generated setup code. See [compatibility and verification](docs/compatibility.md) for coverage and limits.

## Portable UI contract (default producer output)

```sh
html-ui accordion | html-ui-to-vue-vapor > Accordion.vue
```

Version 2 preserves `contractVersion`, `ui`, and generated `<Name>Style`/`<Name>Classes` types as named exports in a regular `<script lang="ts">` block. The component implementation remains in `<script setup lang="ts" vapor>`. Native nodes retain `data-ui` and `data-ui-part` markers, including public part names that differ from factory node identifiers (accordion `trigger` targets `summary`). Metadata state sources refer to the original factory node IDs; a subsequent theme transform maps those IDs through `ui.parts` to the corresponding template markers, scoped to the component instance.

The parser validates owned-node coverage, static markers, state sources, behavior, and matching styling types. It never evaluates input or passes through arbitrary TypeScript. Metadata is JSON-escaped when emitted into the SFC.

This stage generates framework code only. Styling types are preserved contracts for theme tools; there is no `classes` prop, token selection, CVA runtime, or recipe resolver yet. Adapter-required primitives retain that limitation and emit a warning. No collection behavior is invented.

The compiler suite builds the sibling `../html-ui-cli` producer and converts all 37 default contracts. Set `HTML_UI_SOURCE` to another producer source directory if needed. Run `bun run test:compiler` before the type suite, which checks those generated components and their named styling-type exports. Browser disclosure cases use v2 and verify markers, native state, models and events.

### UI plugin integration

Generated v2 SFCs can feed `html-ui-shadcn --framework vue --plugin shadcn-ui`. Safe simple attributes and disabled/required properties use declarative bindings. Native model/default-value/reset synchronization remains explicit; components without form controls omit form-reset subscriptions. The downstream library writer can separate recipe and contract metadata companions for editable application source.

Text-value directives skip writes when the native control already holds the model value. This preserves browser user-edit validation (such as minlength), selection and native editing while keeping external prop changes synchronized. Native table sections include thead, tbody and tfoot.
