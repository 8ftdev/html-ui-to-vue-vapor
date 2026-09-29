package generate

import (
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestStateInferenceUsesNativeMappingsOnly(t *testing.T) {
	c := parsed(t, testinput.Checkbox)
	fields := inferState(c)
	if len(fields) != 1 || fields[0].Prop != "checked" {
		t.Fatalf("%#v", fields)
	}
	r := generated(t, c)
	for _, fragment := range []string{`_htmlUiRef<boolean>(`, `"update:checked"`, `event.currentTarget`, `queueMicrotask`, `event.defaultPrevented`, `_htmlUiDisposed`} {
		if !strings.Contains(r.Source, fragment) {
			t.Fatalf("missing %s", fragment)
		}
	}
	if strings.Count(r.Source, `_htmlUiRef<boolean>(`) != 1 {
		t.Fatal("duplicate state")
	}
	if strings.Contains(r.Source, "event.target") {
		t.Fatal("reads event origin instead of listener node")
	}
}
func TestStateBacksScopedSlots(t *testing.T) {
	r := generated(t, parsed(t, testinput.Disclosure))
	if !strings.Contains(r.Source, `open: _htmlUiState0`) {
		t.Fatal(r.Source)
	}
}

func TestOptionalNativeModelEmitsOnlyNativeReadType(t *testing.T) {
	c := parsed(t, testinput.Checkbox)
	for i := range c.Props {
		if c.Props[i].Name == "checked" {
			c.Props[i].Optional = true
			c.Props[i].Default = nil
		}
	}
	for i := range c.Nodes {
		for j := range c.Nodes[i].Bindings {
			if c.Nodes[i].Bindings[j].Prop == "checked" {
				c.Nodes[i].Bindings[j].Guarded = true
			}
		}
	}
	source := generated(t, c).Source
	if !strings.Contains(source, `_htmlUiRef<boolean | undefined>`) {
		t.Fatal("optional incoming prop needs optional local initialization")
	}
	if !strings.Contains(source, `"update:checked": [value: boolean]`) {
		t.Fatal("native checked reads always emit boolean, even when incoming prop is omitted")
	}
}
