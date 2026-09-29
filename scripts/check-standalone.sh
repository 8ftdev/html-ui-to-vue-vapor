#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p .test-output
converter_root="$(pwd)"
GOCACHE="$converter_root/.test-output/go-cache" go build -o "$converter_root/.test-output/html-ui-to-vue-vapor" ./cmd/html-ui-to-vue-vapor
PATH=/nonexistent "$converter_root/.test-output/html-ui-to-vue-vapor" > "$converter_root/.test-output/standalone.vue" <<'SOURCE'
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
