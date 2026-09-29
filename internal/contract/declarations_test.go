package contract

import (
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestDeclarationsPreserveContract(t *testing.T) {
	p, err := newParser(testinput.Disclosure)
	if err != nil {
		t.Fatal(err)
	}
	c, err := parseDeclarations(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != 1 || c.Name != "Disclosure" || c.Behavior != "native" {
		t.Fatalf("%#v", c)
	}
	if len(c.Props) != 2 || c.Props[1].Name != "open" || string(c.Props[1].Default) != "false" {
		t.Fatalf("%#v", c.Props)
	}
	if c.Slots[0].Optional || !c.Slots[1].Optional || c.Slots[1].Scope[0].Type != "boolean" {
		t.Fatalf("%#v", c.Slots)
	}
	if c.Events[0].Target != "root" || c.Events[0].State[0].Property != "open" {
		t.Fatalf("%#v", c.Events)
	}
}
func TestDeclarationsRejectInconsistentMetadata(t *testing.T) {
	for _, pair := range [][2]string{
		{"contractVersion = 1", "contractVersion = 999"},
		{"open: false", "missing: false"},
		{"open: false", "open: \"no\""},
		{"open?: boolean;", "open?: boolean; open?: boolean;"},
		{"target: \"root\"", "target: \"root\", target: \"other\""},
		{"toggle: ToggleEvent;", "other: ToggleEvent;"},
		{"DisclosureEvents", "DifferentEvents"},
	} {
		p, err := newParser(strings.Replace(testinput.Disclosure, pair[0], pair[1], 1))
		if err != nil {
			continue
		}
		if _, err = parseDeclarations(p); err == nil {
			t.Fatalf("accepted %q", pair[1])
		}
	}
}
