package contract

import (
	"encoding/json"
	"strconv"
	"strings"
)

func quote(value string) string { b, _ := json.Marshal(value); return string(b) }
func (p *parser) literal() json.RawMessage {
	switch p.peek().Kind {
	case "string":
		return json.RawMessage(quote(p.stringLiteral()))
	case "number":
		tok := p.take()
		if _, err := strconv.ParseFloat(tok.Text, 64); err != nil {
			panic(diagnostic(tok.Pos, "invalid number"))
		}
		if !json.Valid([]byte(tok.Text)) {
			panic(diagnostic(tok.Pos, "invalid number"))
		}
		return json.RawMessage(tok.Text)
	}
	if p.accept("-") {
		v := p.literal()
		if len(v) == 0 || v[0] < '0' || v[0] > '9' {
			p.fail("expected numeric literal")
		}
		return append(json.RawMessage{'-'}, v...)
	}
	if p.peek().Text == "true" || p.peek().Text == "false" {
		return json.RawMessage(p.take().Text)
	}
	p.fail("expected primitive literal")
	return nil
}
func parseDeclarations(p *parser) (c *Component, err error) {
	defer recoverDiagnostic(&err)
	c = &Component{Behavior: p.behavior}
	p.want("export", "const", "contractVersion", "=")
	version := p.literal()
	if string(version) != "1" && string(version) != "2" {
		p.fail("unsupported contract version")
	}
	c.Version, _ = strconv.Atoi(string(version))
	p.want("as", "const", ";")
	p.want("export", "interface")
	name := p.identifier()
	if !strings.HasSuffix(name, "Props") || len(name) == 5 {
		p.fail("expected named Props interface")
	}
	c.Name = strings.TrimSuffix(name, "Props")
	for _, field := range p.fields() {
		c.Props = append(c.Props, Prop{Field: field})
	}
	p.want("export", "const", "defaults", "=", "{")
	defaults := map[string]json.RawMessage{}
	for !p.accept("}") {
		key := p.identifier()
		if _, ok := defaults[key]; ok {
			p.fail("duplicate default " + key)
		}
		p.want(":")
		defaults[key] = p.literal()
		if !p.accept(",") && p.peek().Text != "}" {
			p.fail("expected comma")
		}
	}
	p.want("as", "const", "satisfies", "Partial", "<", name, ">", ";")
	for key, value := range defaults {
		found := false
		for i := range c.Props {
			if c.Props[i].Name == key {
				found = true
				c.Props[i].Default = value
				if !defaultMatches(c.Props[i]) {
					p.fail("default does not match " + key)
				}
			}
		}
		if !found {
			p.fail("unknown default prop " + key)
		}
	}
	p.want("export", "interface", c.Name+"Slots", "<", "Content", "=", "HTMLElement", ">", "{")
	seen := map[string]bool{}
	for !p.accept("}") {
		slot := Slot{Pos: p.peek().Pos, Name: p.identifier()}
		if seen[slot.Name] {
			p.fail("duplicate slot")
		}
		seen[slot.Name] = true
		slot.Optional = p.accept("?")
		p.want(":", "(")
		if !p.accept(")") {
			p.want("scope", ":")
			slot.Scope = p.fields()
			for _, f := range slot.Scope {
				if f.Optional {
					p.fail("optional scope fields are unsupported")
				}
			}
			p.want(")")
		}
		p.want("=>", "Content")
		if !p.accept(";") && p.peek().Text != "}" {
			p.fail("expected semicolon")
		}
		c.Slots = append(c.Slots, slot)
	}
	p.want("export", "interface", c.Name+"Events", "{")
	seen = map[string]bool{}
	for !p.accept("}") {
		event := Event{Pos: p.peek().Pos, Name: p.identifier()}
		if seen[event.Name] {
			p.fail("duplicate event")
		}
		seen[event.Name] = true
		p.want(":")
		event.Type = p.identifier()
		if !eventTypes[event.Type] {
			p.fail("unsupported DOM event type " + event.Type)
		}
		if !p.accept(";") && p.peek().Text != "}" {
			p.fail("expected semicolon")
		}
		c.Events = append(c.Events, event)
	}
	p.want("export", "const", "nativeEvents", "=", "{")
	mapped := map[string]bool{}
	for !p.accept("}") {
		eventName := p.identifier()
		if mapped[eventName] {
			p.fail("duplicate native event")
		}
		mapped[eventName] = true
		index := -1
		for i, e := range c.Events {
			if e.Name == eventName {
				index = i
				break
			}
		}
		if index < 0 {
			p.fail("native event missing from Events interface")
		}
		p.want(":", "{", "target", ":")
		c.Events[index].Target = p.stringLiteral()
		p.want(",", "state", ":", "{")
		fields := map[string]bool{}
		for !p.accept("}") {
			read := StateRead{Pos: p.peek().Pos, Prop: p.identifier()}
			if fields[read.Prop] {
				p.fail("duplicate state field")
			}
			fields[read.Prop] = true
			p.want(":")
			read.Property = p.stringLiteral()
			c.Events[index].State = append(c.Events[index].State, read)
			if !p.accept(",") && p.peek().Text != "}" {
				p.fail("expected comma")
			}
		}
		p.accept(",")
		p.want("}")
		if !p.accept(",") && p.peek().Text != "}" {
			p.fail("expected comma")
		}
	}
	p.want("as", "const", ";")
	for _, e := range c.Events {
		if !mapped[e.Name] {
			p.fail("event missing from nativeEvents")
		}
	}
	if c.Version == 2 {
		p.parseUI(c)
		if c.Behavior == "unknown" {
			c.Behavior = c.UI.Behavior.Kind
		}
	}
	return c, nil
}

var eventTypes = map[string]bool{"Event": true, "KeyboardEvent": true, "MouseEvent": true, "SubmitEvent": true, "ToggleEvent": true, "InputEvent": true, "FocusEvent": true, "PointerEvent": true, "WheelEvent": true, "DragEvent": true, "AnimationEvent": true, "TransitionEvent": true}

func defaultMatches(prop Prop) bool {
	var value any
	if json.Unmarshal(prop.Default, &value) != nil {
		return false
	}
	switch prop.Type {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	}
	for _, member := range strings.Split(prop.Type, " | ") {
		if member == string(prop.Default) {
			return true
		}
	}
	return false
}
