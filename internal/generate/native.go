package generate

import (
	"fmt"
	"html-ui-to-vue-vapor/internal/contract"
)

func (r *renderer) declarative(b contract.Binding) bool {
	if r.stateIndex(b.Prop) >= 0 {
		return false
	}
	switch b.Name {
	case "id", "name", "type", "placeholder", "autocomplete", "aria-label":
		return true
	case "disabled", "required":
		return b.Kind == "property"
	}
	return false
}
func (r *renderer) imperative(node contract.Node) bool {
	for _, b := range node.Bindings {
		if !r.declarative(b) {
			return true
		}
	}
	return false
}
func (r *renderer) hasBindings() bool {
	for _, node := range r.c.Nodes {
		if r.imperative(node) {
			return true
		}
	}
	return false
}
func (r *renderer) scriptExpression(prop string) string {
	if index := r.stateIndex(prop); index >= 0 {
		return fmt.Sprintf("_htmlUiState%d.value", index)
	}
	return "_htmlUiProps." + prop
}
func (r *renderer) writeNativeBindings() {
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	for i, node := range r.c.Nodes {
		if !r.imperative(node) {
			continue
		}
		p("const vHtmlUiBind%d = (node: Element) => {\n  _htmlUiWatchEffect(() => {\n", i)
		for j, b := range node.Bindings {
			if r.declarative(b) {
				continue
			}
			expr := r.scriptExpression(b.Prop)
			if b.Name == "defaultValue" || b.Name == "defaultChecked" || b.Kind == "attribute" && b.Name == "value" && r.stateIndex(b.Prop) >= 0 {
				expr = fmt.Sprintf("_htmlUiBaseline%d_%d", i, j)
			}
			p("    {\n      const value = %s\n", expr)
			if b.Kind == "attribute" {
				p("      if (value === undefined) node.removeAttribute(%s)\n      else node.setAttribute(%s, String(value))\n", js(b.Name), js(b.Name))
				if b.Name == "value" && r.stateIndex(b.Prop) >= 0 {
					p("      const next = %s ?? ''\n      if (Reflect.get(node, \"value\") !== String(next)) Reflect.set(node, \"value\", next)\n", r.scriptExpression(b.Prop))
				}
			} else {
				if b.Name == "value" && node.Tag == "progress" {
					p("      if (value === undefined) node.removeAttribute(\"value\")\n      else Reflect.set(node, \"value\", value)\n")
				} else if b.Name == "value" && (node.Tag == "input" || node.Tag == "textarea" || node.Tag == "select") {
					p("      const next = value ?? ''\n      if (Reflect.get(node, \"value\") !== String(next)) Reflect.set(node, \"value\", next)\n")
				} else if b.Name == "value" {
					p("      Reflect.set(node, %s, value ?? '')\n", js(b.Name))
				} else {
					p("      if (value !== undefined) Reflect.set(node, %s, value)\n", js(b.Name))
				}
			}
			p("    }\n")
		}
		if len(r.fields) > 0 {
			if r.hasFormControls() {
				p("    _htmlUiSyncForms()\n")
			}
			p("    _htmlUiScheduleNativeRead()\n")
		}
		p("  })\n}\n")
	}
}
