package generate

import (
	"fmt"
	"html-ui-to-vue-vapor/internal/contract"
	"strings"
)

func (r *renderer) writeNode(id string, depth int) {
	node := r.node(id)
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(&r.template, "%s<%s", indent, node.Tag)
	for _, attr := range node.Attributes {
		fmt.Fprintf(&r.template, " %s=\"%s\"", attr.Name, escaped(attr.Value))
	}
	for i, n := range r.c.Nodes {
		if n.ID == id && r.imperative(n) {
			fmt.Fprintf(&r.template, " v-html-ui-bind%d", i)
		}
	}

	for _, binding := range node.Bindings {
		if r.declarative(binding) {
			fmt.Fprintf(&r.template, " :%s=\"%s\"", binding.Name, escaped(r.expression(binding.Prop)))
		}
	}

	for i, e := range r.c.Events {
		if e.Target == id {
			fmt.Fprintf(&r.template, " @%s=\"_htmlUiEvent%d\"", e.Name, i)
		}
	}
	if index := r.nodeRef(id); index >= 0 {
		fmt.Fprintf(&r.template, " ref=\"_htmlUiNode%d\"", index)
	}
	if contract.IsVoidTag(node.Tag) {
		r.template.WriteString(" />\n")
		return
	}
	if len(node.Children) == 0 {
		fmt.Fprintf(&r.template, "></%s>\n", node.Tag)
		return
	}
	r.template.WriteString(">\n")
	for _, child := range node.Children {
		if child.Node != "" {
			r.writeNode(child.Node, depth+1)
		} else {
			r.writeSlot(child, depth+1)
		}
	}
	fmt.Fprintf(&r.template, "%s</%s>\n", indent, node.Tag)
}
func (r *renderer) writeSlot(child contract.Child, depth int) {
	fmt.Fprintf(&r.template, "%s<slot name=\"%s\"", strings.Repeat("  ", depth), escaped(child.Slot))
	if len(child.Scope) > 0 {
		var fields []string
		for _, scope := range child.Scope {
			fields = append(fields, scope.Name+": "+r.expression(scope.Prop))
		}
		fmt.Fprintf(&r.template, " v-bind=\"%s\"", escaped("{ "+strings.Join(fields, ", ")+" }"))
	}
	r.template.WriteString(" />\n")
}
