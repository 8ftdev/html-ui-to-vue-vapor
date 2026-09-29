package contract

import (
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestNativeTableFooterContract(t *testing.T) {
	source := strings.ReplaceAll(testinput.Static, "HTMLDivElement", "HTMLTableSectionElement")
	source = strings.ReplaceAll(source, `createElement("div")`, `createElement("tfoot")`)
	component, err := Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	if component.Nodes[0].DOMType != "HTMLTableSectionElement" {
		t.Fatal("incorrect footer DOM type")
	}
}
