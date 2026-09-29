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
