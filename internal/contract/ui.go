package contract

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var regexpName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type UIContract struct {
	Component string `json:"component"`
	Behavior  struct {
		Kind         string   `json:"kind"`
		Requirements []string `json:"requirements"`
	} `json:"behavior"`
	Parts map[string]UIPart `json:"parts"`
}
type UIPart struct {
	Node      string             `json:"node"`
	StyleRole string             `json:"styleRole"`
	State     map[string]UIState `json:"state,omitempty"`
}
type UIState struct {
	Source UISource `json:"source"`
}
type UISource struct {
	Node      string  `json:"node"`
	Pseudo    string  `json:"pseudo,omitempty"`
	Attribute string  `json:"attribute,omitempty"`
	Present   *bool   `json:"present,omitempty"`
	Value     *string `json:"value,omitempty"`
}

// UI literals are data only: expressions, duplicate keys and unknown fields fail closed.
func (p *parser) uiLiteral() json.RawMessage {
	if p.accept("{") {
		fields := map[string]json.RawMessage{}
		for !p.accept("}") {
			key := p.stringLiteral()
			if _, exists := fields[key]; exists {
				p.fail("duplicate UI field " + key)
			}
			p.want(":")
			fields[key] = p.uiLiteral()
			if !p.accept(",") && p.peek().Text != "}" {
				p.fail("expected comma")
			}
		}
		b, _ := json.Marshal(fields)
		return b
	}
	if p.accept("[") {
		values := []json.RawMessage{}
		for !p.accept("]") {
			values = append(values, p.uiLiteral())
			if !p.accept(",") && p.peek().Text != "]" {
				p.fail("expected comma")
			}
		}
		b, _ := json.Marshal(values)
		return b
	}
	return p.literal()
}
func (p *parser) parseUI(c *Component) {
	p.want("export", "const", "ui", "=")
	raw := p.uiLiteral()
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	var ui UIContract
	if err := decoder.Decode(&ui); err != nil {
		p.fail("invalid UI metadata: " + err.Error())
	}
	c.UI = &ui
	p.want("as", "const", ";")
	// Types must agree with metadata; arbitrary TypeScript is never copied through.
	expected, err := newParser(UIStyleTypes(c))
	if err != nil {
		p.fail("invalid UI type names")
	}
	for expected.peek().Kind != "eof" {
		want := expected.take()
		if p.peek().Kind != want.Kind || p.peek().Text != want.Text {
			p.fail("styling types do not match UI metadata")
		}
		p.take()
	}
}
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// UIStyleTypes emits the portable styling API independently of Vue props.
func UIStyleTypes(c *Component) string {
	var b strings.Builder
	p := func(f string, a ...any) { fmt.Fprintf(&b, f, a...) }
	p("export type %sStyle<Style = string> = Style | { mode: \"replace\"; value: Style } | { mode: \"omit\" };\n\n", c.Name)
	p("export interface %sClasses<Style = string> {\n", c.Name)
	for _, key := range sortedKeys(c.UI.Parts) {
		part := c.UI.Parts[key]
		p("  %s?: {\n    base?: %sStyle<Style>;\n    unstyled?: boolean;\n", key, c.Name)
		if len(part.State) > 0 {
			p("    state?: {\n")
			for _, state := range sortedKeys(part.State) {
				p("      %s?: %sStyle<Style>;\n", state, c.Name)
			}
			p("    };\n")
		}
		p("  };\n")
	}
	p("}\n")
	return b.String()
}

// UIExports uses JSON escaping to prevent metadata from closing an SFC script.
func UIExports(c *Component) string {
	data, _ := json.MarshalIndent(c.UI, "", "  ")
	return fmt.Sprintf("export const contractVersion = 2 as const;\nexport const ui = %s as const;\n\n%s", data, UIStyleTypes(c))
}
func validateUI(c *Component) error {
	bad := func(message string) error { return diagnostic(Position{Line: 1, Column: 1}, "UI contract: "+message) }
	if c.Version == 1 {
		if c.UI != nil {
			return bad("metadata requires version 2")
		}
		return nil
	}
	if c.UI == nil {
		return bad("missing metadata")
	}
	ui := c.UI
	if !regexpName.MatchString(ui.Component) || (ui.Behavior.Kind != "native" && ui.Behavior.Kind != "adapter-required") {
		return bad("invalid identity or behavior")
	}
	if c.Behavior != "unknown" && c.Behavior != ui.Behavior.Kind {
		return bad("behavior conflicts with documentation")
	}
	nodes := map[string]Node{}
	for _, n := range c.Nodes {
		nodes[n.ID] = n
	}
	seen := map[string]bool{}
	for key, part := range ui.Parts {
		n, exists := nodes[part.Node]
		if !identifierPattern.MatchString(key) || key == "__proto__" || !regexpName.MatchString(part.StyleRole) || !exists || seen[part.Node] {
			return bad("invalid or duplicate part " + key)
		}
		seen[part.Node] = true
		attrs := map[string]string{}
		for _, attr := range n.Attributes {
			attrs[attr.Name] = attr.Value
		}
		if attrs["data-ui"] != ui.Component || attrs["data-ui-part"] != key {
			return bad("part markers do not match " + key)
		}
		for _, binding := range n.Bindings {
			if binding.Name == "data-ui" || binding.Name == "data-ui-part" {
				return bad("part markers must be static")
			}
		}
		for state, binding := range part.State {
			source := binding.Source
			node, exists := nodes[source.Node]
			if !identifierPattern.MatchString(state) || state == "__proto__" || !exists {
				return bad("invalid state " + state)
			}
			if source.Pseudo != "" {
				if source.Attribute != "" || source.Present != nil || source.Value != nil || !validUIPseudo(node, source.Pseudo) {
					return bad("invalid pseudo source " + state)
				}
			} else {
				if !validAttribute(source.Attribute) || (source.Present == nil) == (source.Value == nil) {
					return bad("invalid attribute source " + state)
				}
				if source.Attribute == "open" && node.Tag != "details" && node.Tag != "dialog" {
					return bad("open requires details or dialog")
				}
			}
			if state == "disabled" && source.Pseudo != "disabled" && !(source.Attribute == "aria-disabled" && source.Value != nil && *source.Value == "true") {
				return bad("disabled requires native or ARIA disabled source")
			}
		}
	}
	if len(seen) != len(nodes) || ui.Parts["root"].Node != c.Root {
		return bad("parts must cover owned nodes and identify root")
	}
	return nil
}
func validUIPseudo(node Node, pseudo string) bool {
	typ := ""
	for _, attr := range node.Attributes {
		if attr.Name == "type" {
			typ = attr.Value
		}
	}
	switch pseudo {
	case "hover", "active", "focus-visible", "focus-within":
		return true
	case "disabled":
		return strings.Contains(" button input select textarea fieldset optgroup option ", " "+node.Tag+" ")
	case "checked":
		return node.Tag == "option" || node.Tag == "input" && (typ == "checkbox" || typ == "radio")
	case "indeterminate":
		return node.Tag == "progress" || node.Tag == "input" && (typ == "checkbox" || typ == "radio")
	case "required", "invalid":
		return node.Tag == "select" || node.Tag == "textarea" || node.Tag == "input" && !strings.Contains(" hidden range color button submit reset image ", " "+typ+" ")
	case "popover-open":
		for _, a := range node.Attributes {
			if a.Name == "popover" {
				return true
			}
		}
		for _, b := range node.Bindings {
			if b.Name == "popover" {
				return true
			}
		}
	}
	return false
}
