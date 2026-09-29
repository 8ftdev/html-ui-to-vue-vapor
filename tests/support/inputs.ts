// Hand-written contract examples shared with Go tests; no producer dependency.
export const disclosure = "/** Behavior: native. */\nexport const contractVersion = 1 as const;\nexport interface DisclosureProps { name?: string; open?: boolean; }\nexport const defaults = { open: false } as const satisfies Partial<DisclosureProps>;\nexport interface DisclosureSlots<Content = HTMLElement> {\n  summary: () => Content;\n  content?: (scope: { open: boolean }) => Content;\n}\nexport interface DisclosureEvents { toggle: ToggleEvent; }\nexport const nativeEvents = { toggle: { target: \"root\", state: { open: \"open\" } } } as const;\nexport function disclose(props: DisclosureProps, slots: DisclosureSlots<HTMLElement>): HTMLDetailsElement {\n  const { name, open = defaults.open } = props;\n  const root = document.createElement(\"details\");\n  const summary = document.createElement(\"summary\");\n  if (name !== undefined) root.setAttribute(\"name\", name);\n  root.open = open;\n  summary.append(slots.summary());\n  root.append(summary);\n  if (slots.content !== undefined) root.append(slots.content({ open }));\n  return root;\n}";
export const checkbox = "/** Behavior: native. */\nexport const contractVersion = 1 as const;\nexport interface BinaryProps { checked?: boolean; disabled?: boolean; }\nexport const defaults = { checked: false, disabled: false } as const satisfies Partial<BinaryProps>;\nexport interface BinarySlots<Content = HTMLElement> { label: () => Content; }\nexport interface BinaryEvents { input: Event; change: Event; }\nexport const nativeEvents = {\n  input: { target: \"control\", state: { checked: \"checked\" } },\n  change: { target: \"control\", state: { checked: \"checked\" } }\n} as const;\nexport function binary(props: BinaryProps, slots: BinarySlots<HTMLElement>): HTMLLabelElement {\n  const { checked = defaults.checked, disabled = defaults.disabled } = props;\n  const root = document.createElement(\"label\");\n  const control = document.createElement(\"input\");\n  control.setAttribute(\"type\", \"checkbox\");\n  control.disabled = disabled;\n  control.defaultChecked = checked;\n  control.checked = checked;\n  root.append(slots.label());\n  root.append(control);\n  return root;\n}";
export const staticInput = "/** Behavior: native. */\nexport const contractVersion = 1 as const;\nexport interface ExampleProps {}\nexport const defaults = {} as const satisfies Partial<ExampleProps>;\nexport interface ExampleSlots<Content = HTMLElement> {}\nexport interface ExampleEvents {}\nexport const nativeEvents = {} as const;\nexport function example(props: ExampleProps, slots: ExampleSlots<HTMLElement>): HTMLDivElement {\n  const root = document.createElement(\"div\");\n  return root;\n}";

export function controlInput(kind: 'text' | 'number' | 'range'): string {
 const numeric=kind!=='text'
 const type=numeric?'number':'string'
 const defaults=numeric?'disabled: false':'disabled: false, value: "initial"'
 const baseline=numeric?'':'  control.defaultValue = value;'
 const read=numeric?'valueAsNumber':'value'
 return `/** Behavior: native. */
export const contractVersion = 1 as const;
export interface ControlProps { disabled?: boolean; value${numeric?'':'?'}: ${type}; min?: number; max?: number; step?: number; name?: string; form?: string; }
export const defaults = { ${defaults} } as const satisfies Partial<ControlProps>;
export interface ControlSlots<Content = HTMLElement> { label: () => Content; content?: (scope: { value: ${type} }) => Content; }
export interface ControlEvents { input: Event; change: Event; }
export const nativeEvents = {
 input: { target: "control", state: { value: "${read}" } },
 change: { target: "control", state: { value: "${read}" } }
} as const;
export function controlExample(props: ControlProps, slots: ControlSlots<HTMLElement>): HTMLLabelElement {
 const { disabled = defaults.disabled, value${numeric?'':' = defaults.value'}, min, max, step, name, form } = props;
 const root = document.createElement("label");
 const control = document.createElement("input");
 control.setAttribute("type", "${kind}");
 control.disabled = disabled;
 if (min !== undefined) control.setAttribute("min", String(min));
 if (max !== undefined) control.setAttribute("max", String(max));
 if (step !== undefined) control.setAttribute("step", String(step));
 if (name !== undefined) control.setAttribute("name", name);
 if (form !== undefined) control.setAttribute("form", form);
${baseline}
 control.${numeric?'setAttribute("value", String(value))':'value = value'};
 root.append(slots.label());
 root.append(control);
 if (slots.content !== undefined) root.append(slots.content({ value }));
 return root;
}`
}
