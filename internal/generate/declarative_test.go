package generate

import (
	"html-ui-to-vue-vapor/internal/contract"
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestDeclarativeAttributeAvoidsWatchEffect(t *testing.T) {
	c := parsed(t, testinput.Static)
	c.Props = append(c.Props, contract.Prop{Field: contract.Field{Name: "id", Type: "string", Optional: true}})
	c.Nodes[0].Bindings = append(c.Nodes[0].Bindings, contract.Binding{Kind: "attribute", Name: "id", Prop: "id", Guarded: true, Stringify: true})
	r := generated(t, c).Source
	if strings.Contains(r, "watchEffect") {
		t.Fatal("simple id binding emits an effect")
	}
	if !strings.Contains(r, `:id="_htmlUiProps.id"`) {
		t.Fatal("missing declarative id")
	}
}

func TestOmittedProgressValueRemainsIndeterminate(t *testing.T) {
	c := parsed(t, testinput.Static)
	c.Nodes[0].Tag = "progress"
	c.Nodes[0].DOMType = "HTMLProgressElement"
	c.RootType = "HTMLProgressElement"
	c.Props = append(c.Props, contract.Prop{Field: contract.Field{Name: "value", Type: "number", Optional: true}})
	c.Nodes[0].Bindings = append(c.Nodes[0].Bindings, contract.Binding{Kind: "property", Name: "value", Prop: "value", Guarded: true})
	output := generated(t, c).Source
	if !strings.Contains(output, `if (value === undefined) node.removeAttribute("value")`) {
		t.Fatal("omitted progress value must remove its value attribute")
	}
}

func TestStringControlAvoidsRedundantValueWrites(t *testing.T) {
	for _, tag := range []string{"input", "textarea", "select"} {
		t.Run(tag, func(t *testing.T) {
			c := parsed(t, testinput.Static)
			c.Nodes[0].Tag = tag
			c.Nodes[0].DOMType = contract.DOMType(tag)
			c.RootType = c.Nodes[0].DOMType
			c.Props = append(c.Props, contract.Prop{Field: contract.Field{Name: "value", Type: "string", Optional: true}})
			c.Nodes[0].Bindings = append(c.Nodes[0].Bindings, contract.Binding{Kind: "property", Name: "value", Prop: "value", Guarded: true})
			output := generated(t, c).Source
			if !strings.Contains(output, `if (Reflect.get(node, "value") !== String(next)) Reflect.set(node, "value", next)`) {
				t.Fatal("string-valued control must preserve native user edits by skipping identical writes")
			}
		})
	}
}
