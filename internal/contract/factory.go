package contract

func Parse(source string) (c *Component, err error) {
	defer recoverDiagnostic(&err)
	p, err := newParser(source)
	if err != nil {
		return nil, err
	}
	c, err = parseDeclarations(p)
	if err != nil {
		return nil, err
	}
	p.want("export", "function")
	c.Factory = p.identifier()
	p.want("(", "props", ":", c.Name+"Props", ",", "slots", ":", c.Name+"Slots", "<", "HTMLElement", ">", ")", ":")
	c.RootType = p.identifier()
	p.want("{")
	bound := map[string]bool{}
	returned := false
	for !p.accept("}") {
		if returned {
			p.fail("statement after return")
		}
		if p.accept("const") {
			if p.accept("{") {
				for !p.accept("}") {
					name := p.identifier()
					prop := findProp(c, name)
					if prop == nil || bound[name] {
						p.fail("unknown or duplicate prop binding " + name)
					}
					bound[name] = true
					if prop.Default != nil {
						p.want("=", "defaults", ".", name)
					} else if p.peek().Text == "=" {
						p.fail("unsupported prop initializer")
					}
					if !p.accept(",") && p.peek().Text != "}" {
						p.fail("expected comma")
					}
				}
				p.want("=", "props", ";")
				continue
			}
			pos := p.peek().Pos
			id := p.identifier()
			if findNode(c, id) != nil {
				p.fail("duplicate node " + id)
			}
			p.want("=", "document", ".", "createElement", "(")
			tag := p.stringLiteral()
			p.want(")", ";")
			c.Nodes = append(c.Nodes, Node{ID: id, Tag: tag, DOMType: DOMType(tag), Pos: pos})
			continue
		}
		if p.accept("return") {
			c.Root = p.identifier()
			p.want(";")
			returned = true
			continue
		}
		guard := ""
		if p.accept("if") {
			p.want("(")
			guard = p.identifier()
			if guard == "slots" {
				p.want(".")
				guard += "." + p.identifier()
			}
			p.want("!==", "undefined", ")")
		}
		p.operation(c, bound, guard)
	}
	if !returned {
		p.fail("missing factory return")
	}
	if len(bound) != len(c.Props) {
		p.fail("incomplete prop destructuring")
	}
	if p.peek().Kind != "eof" {
		p.fail("unexpected trailing source")
	}
	if err = Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}
func (p *parser) operation(c *Component, bound map[string]bool, guard string) {
	pos := p.peek().Pos
	id := p.identifier()
	node := findNode(c, id)
	if node == nil {
		p.fail("unknown node " + id)
	}
	p.want(".")
	method := p.identifier()
	if p.accept("=") {
		prop := p.identifier()
		if !bound[prop] {
			p.fail("unknown prop binding")
		}
		p.want(";")
		node.Bindings = append(node.Bindings, Binding{Kind: "property", Name: method, Prop: prop, Guarded: guard != "", Pos: pos})
		checkGuard(p, c, prop, guard)
		return
	}
	p.want("(")
	switch method {
	case "setAttribute":
		name := p.stringLiteral()
		p.want(",")
		if p.peek().Kind == "string" {
			if guard != "" {
				p.fail("static attribute cannot be guarded")
			}
			if len(node.Bindings) > 0 {
				p.fail("static attributes must precede reactive bindings")
			}
			value := p.stringLiteral()
			node.Attributes = append(node.Attributes, Attribute{Name: name, Value: value, Pos: pos})
		} else {
			stringify := p.accept("String")
			if stringify {
				p.want("(")
			}
			prop := p.identifier()
			if !bound[prop] {
				p.fail("unknown prop binding")
			}
			if stringify {
				p.want(")")
			}
			checkGuard(p, c, prop, guard)
			node.Bindings = append(node.Bindings, Binding{Kind: "attribute", Name: name, Prop: prop, Guarded: guard != "", Stringify: stringify, Pos: pos})
		}
	case "append":
		child := Child{Pos: pos}
		target := p.identifier()
		if target == "slots" {
			p.want(".")
			child.Slot = p.identifier()
			slot := findSlot(c, child.Slot)
			if slot == nil {
				p.fail("unknown slot")
			}
			child.Optional = slot.Optional
			wanted := ""
			if slot.Optional {
				wanted = "slots." + slot.Name
			}
			if guard != wanted {
				p.fail("incorrect optional slot guard")
			}
			p.want("(")
			if p.accept("{") {
				seen := map[string]bool{}
				for !p.accept("}") {
					field := ScopeBinding{Pos: p.peek().Pos, Name: p.identifier()}
					if seen[field.Name] {
						p.fail("duplicate scope field")
					}
					seen[field.Name] = true
					field.Prop = field.Name
					if p.accept(":") {
						field.Prop = p.identifier()
					}
					if !bound[field.Prop] {
						p.fail("unknown scope prop")
					}
					child.Scope = append(child.Scope, field)
					if !p.accept(",") && p.peek().Text != "}" {
						p.fail("expected comma")
					}
				}
			}
			p.want(")")
		} else {
			if guard != "" {
				p.fail("native child cannot be guarded")
			}
			child.Node = target
		}
		node.Children = append(node.Children, child)
	default:
		p.fail("unsupported native call " + method)
	}
	p.want(")", ";")
}
func checkGuard(p *parser, c *Component, prop, guard string) {
	field := findProp(c, prop)
	wanted := ""
	if field.Optional && field.Default == nil {
		wanted = prop
	}
	if guard != wanted {
		p.fail("incorrect optional prop guard")
	}
}
func findProp(c *Component, name string) *Prop {
	for i := range c.Props {
		if c.Props[i].Name == name {
			return &c.Props[i]
		}
	}
	return nil
}
func findSlot(c *Component, name string) *Slot {
	for i := range c.Slots {
		if c.Slots[i].Name == name {
			return &c.Slots[i]
		}
	}
	return nil
}
func findNode(c *Component, id string) *Node {
	for i := range c.Nodes {
		if c.Nodes[i].ID == id {
			return &c.Nodes[i]
		}
	}
	return nil
}
