package generate

import (
	"encoding/json"
	"fmt"
	"html-ui-to-vue-vapor/internal/contract"
)

type stateRead struct{ Target, DOMType, Property string }
type stateField struct {
	Prop, Type string
	Default    json.RawMessage
	Reads      []stateRead
}

func inferState(c *contract.Component) []stateField {
	fields := []stateField{}
	for _, prop := range c.Props {
		field := stateField{Prop: prop.Name, Type: prop.Type, Default: prop.Default}
		if prop.Optional && prop.Default == nil {
			field.Type += " | undefined"
		}
		for _, event := range c.Events {
			for _, read := range event.State {
				if read.Prop != prop.Name {
					continue
				}
				duplicate := false
				for _, old := range field.Reads {
					if old.Target == event.Target && old.Property == read.Property {
						duplicate = true
					}
				}
				if !duplicate {
					for _, node := range c.Nodes {
						if node.ID == event.Target {
							field.Reads = append(field.Reads, stateRead{event.Target, node.DOMType, read.Property})
						}
					}
				}
			}
		}
		if len(field.Reads) > 0 {
			fields = append(fields, field)
		}
	}
	return fields
}
func (r *renderer) stateIndex(prop string) int {
	for i, f := range r.fields {
		if f.Prop == prop {
			return i
		}
	}
	return -1
}
func (r *renderer) nodeRef(id string) int {
	for i, node := range r.c.Nodes {
		if node.ID == id {
			for _, field := range r.fields {
				for _, read := range field.Reads {
					if read.Target == id {
						return i
					}
				}
			}
		}
	}
	return -1
}
func (r *renderer) writeState() {
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	for i, field := range r.fields {
		p("const _htmlUiState%d = _htmlUiRef<%s>(_htmlUiProps[%s])\n", i, field.Type, js(field.Prop))
		p("_htmlUiWatch(() => _htmlUiProps[%s], value => { _htmlUiState%d.value = value })\n", js(field.Prop), i)
	}
	for i, node := range r.c.Nodes {
		if r.nodeRef(node.ID) >= 0 {
			p("const _htmlUiNode%d = _htmlUiRef<%s | null>(null)\n", i, node.DOMType)
		}
	}
	for i, node := range r.c.Nodes {
		for j, b := range node.Bindings {
			if b.Name == "defaultValue" || b.Name == "defaultChecked" || b.Kind == "attribute" && b.Name == "value" && r.stateIndex(b.Prop) >= 0 {
				p("const _htmlUiBaseline%d_%d = _htmlUiProps[%s]\n", i, j, js(b.Prop))
			}
		}
	}
}
func (r *renderer) bindingExpression(nodeIndex, bindingIndex int, b contract.Binding) string {
	if b.Name == "defaultValue" || b.Name == "defaultChecked" || b.Kind == "attribute" && b.Name == "value" && r.stateIndex(b.Prop) >= 0 {
		return fmt.Sprintf("_htmlUiBaseline%d_%d", nodeIndex, bindingIndex)
	}
	return r.expression(b.Prop)
}
