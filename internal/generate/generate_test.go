package generate

import (
	"html-ui-to-vue-vapor/internal/contract"
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func parsed(t *testing.T, source string) *contract.Component {
	t.Helper()
	c, err := contract.Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func generated(t *testing.T, c *contract.Component) Result {
	t.Helper()
	r, err := Generate(c)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestGenerateStructureAndContracts(t *testing.T) {
	r := generated(t, parsed(t, testinput.Disclosure))
	for _, fragment := range []string{`<script setup lang="ts" vapor>`, `"open"?: boolean`, `summary`, `<summary>`, `<slot name="content"`, `v-html-ui-bind0`} {
		if !strings.Contains(r.Source, fragment) {
			t.Fatalf("missing %s:\n%s", fragment, r.Source)
		}
	}
	if strings.Index(r.Source, "<summary>") > strings.Index(r.Source, `<slot name="content"`) {
		t.Fatal("child order changed")
	}
	if generated(t, parsed(t, testinput.Disclosure)).Source != r.Source {
		t.Fatal("nondeterministic output")
	}
}
func TestGenerateEscapesContexts(t *testing.T) {
	c := parsed(t, testinput.Disclosure)
	c.Nodes[0].Attributes = append(c.Nodes[0].Attributes, contract.Attribute{Name: "data-label", Value: `a"&é</script>`})
	r := generated(t, c)
	if strings.Count(r.Source, "</script>") != 1 || !strings.Contains(r.Source, "&amp;") || !strings.Contains(r.Source, "&#34;") {
		t.Fatal(r.Source)
	}
	c = parsed(t, testinput.Disclosure)
	c.Props[0].Default = []byte(`"</script>&"`)
	// A prop with a default no longer requires its optional guard.
	c.Nodes[0].Bindings[0].Guarded = false
	r = generated(t, c)
	if strings.Count(r.Source, "</script>") != 1 || !strings.Contains(r.Source, `\u003c/script\u003e`) {
		t.Fatal(r.Source)
	}
}
func TestGenerateVoidElements(t *testing.T) {
	r := generated(t, parsed(t, testinput.Checkbox))
	if strings.Contains(r.Source, "</input>") {
		t.Fatal(r.Source)
	}
}
