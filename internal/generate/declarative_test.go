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
			expected := `if (Reflect.get(node, "value") !== String(next)) Reflect.set(node, "value", next)`
			if tag == "select" {
				expected = `if (value !== undefined && Reflect.get(node, "value") !== String(value)) Reflect.set(node, "value", value)`
			}
			if !strings.Contains(output, expected) {
				t.Fatal("string-valued control must preserve native user edits by skipping identical writes")
			}
		})
	}
}

func TestNativeBindingsTrackIndependentDependencies(t *testing.T) {
	c := parsed(t, testinput.Checkbox)
	c.Props = append(c.Props, contract.Prop{Field: contract.Field{Name: "indeterminate", Type: "boolean", Optional: true}})
	c.Nodes[1].Bindings = append(c.Nodes[1].Bindings, contract.Binding{Kind: "property", Name: "indeterminate", Prop: "indeterminate", Guarded: true})
	source := generated(t, c).Source
	start := strings.Index(source, "const vHtmlUiBind1")
	end := strings.Index(source[start:], "function _htmlUi")
	if end < 0 {
		t.Fatal("missing native event helpers")
	}
	bindings := source[start : start+end]
	if strings.Count(bindings, "_htmlUiWatchEffect(() => {") != 3 {
		t.Fatal("checked and indeterminate must have independent effects: native activation clears indeterminate without changing its prop")
	}
}

func TestOmittedSelectValuePreservesSelectedOption(t *testing.T) {
	c := parsed(t, testinput.Static)
	c.Nodes[0].Tag = "select"
	c.Nodes[0].DOMType = "HTMLSelectElement"
	c.RootType = "HTMLSelectElement"
	c.Props = append(c.Props, contract.Prop{Field: contract.Field{Name: "value", Type: "string", Optional: true}})
	c.Nodes[0].Bindings = append(c.Nodes[0].Bindings, contract.Binding{Kind: "property", Name: "value", Prop: "value", Guarded: true})
	source := generated(t, c).Source
	if !strings.Contains(source, `if (value !== undefined && Reflect.get(node, "value") !== String(value)) Reflect.set(node, "value", value)`) {
		t.Fatal("an omitted select value must preserve the selected option, rather than assigning an empty string")
	}
}
