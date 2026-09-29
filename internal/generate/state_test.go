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
