package contract

import (
	"errors"
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestParseSyntheticInputsAndArbitraryNames(t *testing.T) {
	renamed := strings.ReplaceAll(testinput.Disclosure, "Disclosure", "CustomSection")
	renamed = strings.Replace(renamed, "function disclose", "function customSection", 1)
	for _, source := range []string{testinput.Disclosure, testinput.Checkbox, testinput.Static, renamed} {
		c, err := Parse(source)
		if err != nil {
			t.Fatal(err)
		}
		if c.Root != "root" || len(c.Nodes) == 0 {
			t.Fatalf("%#v", c)
		}
	}
}
func TestParseRejectsInvalidFactory(t *testing.T) {
	for _, pair := range [][2]string{
		{"  return root;", "  arbitraryCall();\n  return root;"},
		{"  return root;", ""},
		{"root.open = open;", "root.open = unknown;"},
		{"if (name !== undefined) ", ""},
		{"root.append(summary);", "root.append(summary); root.append(summary);"},
		{"root.append(summary);", "summary.append(root); root.append(summary);"},
		{"slots.content({ open })", "slots.content({ wrong: open })"},
		{"open = defaults.open", "open"},
		{"target: \"root\"", "target: \"absent\""},
		{"state: { open: \"open\" }", "state: { open: \"checked\" }"},
		{"document.createElement(\"details\")", "arbitraryCall()"},
		{"root.open = open;", "root.open = open; root.innerHTML = open;"},
	} {
		source := strings.Replace(testinput.Disclosure, pair[0], pair[1], 1)
		_, err := Parse(source)
		var d *Diagnostic
		if !errors.As(err, &d) {
			t.Fatalf("accepted or unlocated rejection for %q: %v", pair[1], err)
		}
	}
	for _, source := range []string{testinput.Disclosure + "arbitraryCall();", strings.Split(testinput.Disclosure, "export function")[0]} {
		if _, err := Parse(source); err == nil {
			t.Fatal("accepted trailing/incomplete input")
		}
	}
}
func TestParseSlotAliases(t *testing.T) {
	source := strings.Replace(testinput.Disclosure, "scope: { open: boolean }", "scope: { isOpen: boolean }", 1)
	source = strings.Replace(source, "slots.content({ open })", "slots.content({ isOpen: open })", 1)
	c, err := Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	child := c.Nodes[0].Children[1]
	if child.Scope[0].Name != "isOpen" || child.Scope[0].Prop != "open" {
		t.Fatalf("%#v", child)
	}
}
func TestParseRejectsVueSyntaxInNativeAttributes(t *testing.T) {
	for _, name := range []string{"v-html", "v-on:click", "ref", "key"} {
		source := strings.Replace(testinput.Static, "  return root;", "  root.setAttribute(\""+name+"\", \"arbitraryCall()\");\n  return root;", 1)
		if _, err := Parse(source); err == nil {
			t.Fatalf("accepted Vue-reserved native attribute %s", name)
		}
	}
}
func TestParseRejectsChildrenOfVoidElement(t *testing.T) {
	source := strings.Replace(testinput.Static, "createElement(\"div\")", "createElement(\"input\")", 1)
	source = strings.Replace(source, ": HTMLDivElement", ": HTMLInputElement", 1)
	source = strings.Replace(source, "  return root;", "  const child = document.createElement(\"span\");\n  root.append(child);\n  return root;", 1)
	if _, err := Parse(source); err == nil {
		t.Fatal("accepted an input with children")
	}
}

func TestParseRejectsVueReservedProps(t *testing.T) {
	for _, name := range []string{"key", "ref", "ref_for", "ref_key", "__proto__"} {
		source := strings.Replace(testinput.Static, "ExampleProps {}", "ExampleProps { "+name+"?: string; }", 1)
		source = strings.Replace(source, "  const root", "  const { "+name+" } = props;\n  const root", 1)
		if _, err := Parse(source); err == nil {
			t.Fatalf("accepted reserved prop %s", name)
		}
	}
}

func TestParseRejectsNativeSlotTag(t *testing.T) {
	source := strings.Replace(testinput.Static, "createElement(\"div\")", "createElement(\"slot\")", 1)
	source = strings.Replace(source, ": HTMLDivElement", ": HTMLElement", 1)
	if _, err := Parse(source); err == nil {
		t.Fatal("accepted native slot tag")
	}
}

func TestParseRejectsTextContentWithChildren(t *testing.T) {
	source := strings.Replace(testinput.Static, "ExampleProps {}", "ExampleProps { text: string; }", 1)
	source = strings.Replace(source, "  const root", "  const { text } = props;\n  const root", 1)
	source = strings.Replace(source, "  return root;", "  root.textContent = text;\n  const child = document.createElement(\"span\");\n  root.append(child);\n  return root;", 1)
	if _, err := Parse(source); err == nil {
		t.Fatal("accepted textContent with children")
	}
}

func TestParseRejectsCaseInsensitiveAttributeDuplicates(t *testing.T) {
	source := strings.Replace(testinput.Static, "  return root;", "  root.setAttribute(\"title\", \"first\");\n  root.setAttribute(\"TITLE\", \"second\");\n  return root;", 1)
	if _, err := Parse(source); err == nil {
		t.Fatal("accepted duplicate HTML attribute with different casing")
	}
}
