# Compatibility and verification

The binary is standalone Go. Input supports version-1 and version-2 source contracts; arbitrary component/interface/factory names are accepted within the documented grammar. The runtime conversion uses no component catalog or producer executable. Integration tests build the sibling producer, compile all 37 default v2 contracts, and use a captured accordion v2 fixture for browser behavior.

## Toolchain

- Go module baseline: 1.24.0; standard library only.
- Vue runtime and SFC/compiler packages: 3.6.0-rc.9.
- TypeScript native compiler: 7.0.2.
- Vue template/type checker: vue-tsc 3.3.11 with @typescript/typescript6 6.0.2 for its JavaScript API.
- Playwright: 1.63.0, Chromium, Firefox, and WebKit.

Vue's checker still consumes TypeScript's JavaScript API, as described in its [official compatibility change](https://github.com/vuejs/language-tools/pull/6123). The primary TypeScript dependency remains 7.0.2; it independently checks generated setup code. This compatibility package is development-only and does not affect the Go binary or generated component runtime.

Generated native bindings use Vapor function directives with watchEffect. This retains ordered DOM property/attribute writes without forcing defaultChecked/defaultValue into Vue's HTML attribute type definitions. Reactive effects and event listeners are released with the component. See the [Vapor API overview](https://github.com/vuejs/core/releases/tag/v3.6.0-rc.1) and [target release](https://github.com/vuejs/core/releases/tag/v3.6.0-rc.9).

## Checks

Run `make check` from the converter directory. The checks combine hand-written source strings, synthetic consumers, and real producer output:

- Go parsing/generation/CLI tests: located rejection, metadata consistency, arbitrary names, scope aliases, graph validation, escaping, native state inference, and stream error behavior.
- Actual Vue Vapor compilation, including renamed inputs, escaped defaults, and all 37 v2 producer contracts.
- TypeScript 7 setup checks and strict Vue consumer checks for props, events, scoped slots, named models, all generated v2 SFCs, and exported part/state override types. Required slot signatures are checked through the public slot type; the current Vue checker does not diagnose every omitted required slot in template syntax.
- Browser cases in all three engines: native disclosure activation and grouping; live scoped slots/models; unrelated parent updates; mapped prop changes; original event forwarding; input/checkbox reset baselines; canceled reset; empty numeric NaN; range sanitization; changed form association; disposal of queued reset work; descendant event origins; omitted optional attributes.
- Pure Vapor and VDOM interop disclosure cases.
- A fresh Go binary accepts stdin with `PATH=/nonexistent`, requiring no Node, Bun, TypeScript, or producer executable.

Final review and full-check results are recorded in [the implementation ledger](implementation-progress.md).

## Boundaries

The parser supports the constrained source grammar, not arbitrary TypeScript programs. Unknown calls, imports, statements, expressions, versions, node/slot/state references, and incompatible native types fail with diagnostics. Rooted native trees are limited to 256 levels to reject cyclic/excessively deep structures safely.

Vue-reserved public props (`key`, `ref`, `ref_key`, `ref_for`, `__proto__`), directive/reserved attributes, native `slot` elements, children of void elements, and `textContent` combined with children receive explicit diagnostics. HTML attribute collision checks ignore case. Scoped fields retain their names through object bindings, including a field named `name`.

Scoped field `__proto__` is rejected to prevent object prototype semantics.

Attribute assignments preserve native `setAttribute` stringification: `String(false)` is the string `"false"`, and a native boolean attribute remains present. Boolean DOM property assignments retain native boolean behavior.

Adapter-required input can encode markup, events, and native state, but types cannot specify missing focus/keyboard/dismissal algorithms. Such output carries its limitation in an SFC comment and stderr diagnostic. No extra interaction algorithm is invented.

Initial form reset baselines stay fixed while later props control live state. Reset subscriptions refresh when generated native binding effects update form association, and are cleaned up on unmount. Arbitrary external DOM relocation without a relevant reactive update is outside the synchronization guarantee.

The compatibility claim applies to the pinned toolchain and exercised contract features. No SSR/hydration compatibility claim is made by these browser checks.

## Version 2 verification, 2026-09-28

Go race tests and vet passed. The full suite passed 46 compiler cases, four type-check cases, and 45 browser cases across Chromium, Firefox, and WebKit. The standalone binary check also passed. Disclosure browser cases use the v2 fixture and assert retained markers in both Vapor and VDOM interop modes.

The generated regular script includes precise Vue module augmentations for `data-ui`, `data-ui-part`, `popover`, and `command` when present, because the pinned Vue HTML types lack these attributes. The declarations add no runtime behavior and do not disable strict template checking. Other Vue versions may require revisiting this compatibility shim as upstream native attribute types evolve.
