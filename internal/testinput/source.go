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
