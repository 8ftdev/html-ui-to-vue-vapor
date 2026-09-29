package contract

import (
	"regexp"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[\pL_$][\pL\pN_$]*$`)
var attributePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.:-]*$`)
var domTypes = map[string]string{
	"details": "HTMLDetailsElement", "summary": "HTMLElement", "input": "HTMLInputElement", "textarea": "HTMLTextAreaElement", "select": "HTMLSelectElement", "option": "HTMLOptionElement", "button": "HTMLButtonElement", "label": "HTMLLabelElement", "div": "HTMLDivElement", "span": "HTMLSpanElement", "fieldset": "HTMLFieldSetElement", "legend": "HTMLLegendElement", "form": "HTMLFormElement", "img": "HTMLImageElement", "hr": "HTMLHRElement", "meter": "HTMLMeterElement", "progress": "HTMLProgressElement", "dialog": "HTMLDialogElement", "a": "HTMLAnchorElement", "ul": "HTMLUListElement", "ol": "HTMLOListElement", "li": "HTMLLIElement", "p": "HTMLParagraphElement", "output": "HTMLOutputElement", "datalist": "HTMLDataListElement", "h1": "HTMLHeadingElement", "h2": "HTMLHeadingElement", "h3": "HTMLHeadingElement", "h4": "HTMLHeadingElement", "h5": "HTMLHeadingElement", "h6": "HTMLHeadingElement", "table": "HTMLTableElement", "tbody": "HTMLTableSectionElement", "thead": "HTMLTableSectionElement", "tfoot": "HTMLTableSectionElement", "tr": "HTMLTableRowElement", "td": "HTMLTableCellElement", "th": "HTMLTableCellElement", "caption": "HTMLTableCaptionElement", "col": "HTMLTableColElement", "colgroup": "HTMLTableColElement", "br": "HTMLBRElement", "video": "HTMLVideoElement", "audio": "HTMLAudioElement",
}
var genericTags = strings.Fields("section article aside header footer nav main address figure figcaption search strong em b i small code pre blockquote cite q time abbr mark s u sub sup kbd samp var ruby rt rp wbr dl dt dd optgroup picture source track canvas noscript template slot")

func DOMType(tag string) string {
	if typ := domTypes[tag]; typ != "" {
		return typ
	}
	for _, t := range genericTags {
		if tag == t {
			return "HTMLElement"
		}
	}
	return ""
}

var propertyTypes = map[string]string{"method": "string", "action": "string", "open": "boolean", "checked": "boolean", "defaultChecked": "boolean", "disabled": "boolean", "required": "boolean", "hidden": "boolean", "multiple": "boolean", "readOnly": "boolean", "noValidate": "boolean", "indeterminate": "boolean", "selected": "boolean", "defaultSelected": "boolean", "value": "string-or-number", "defaultValue": "string", "name": "string", "type": "string", "id": "string", "src": "string", "alt": "string", "placeholder": "string", "maxLength": "number", "minLength": "number", "tabIndex": "number", "htmlFor": "string", "textContent": "string", "min": "string-or-number", "max": "string-or-number", "step": "string-or-number", "low": "number", "high": "number", "optimum": "number"}
var readProperties = map[string]map[string]string{
	"HTMLDetailsElement": {"open": "boolean"}, "HTMLDialogElement": {"open": "boolean"},
	"HTMLInputElement":    {"checked": "boolean", "value": "string", "valueAsNumber": "number", "indeterminate": "boolean"},
	"HTMLTextAreaElement": {"value": "string"}, "HTMLSelectElement": {"value": "string", "selectedIndex": "number"}, "HTMLButtonElement": {"value": "string"},
}

func validAttribute(name string) bool {
	lower := strings.ToLower(name)
	if !attributePattern.MatchString(name) || strings.HasPrefix(lower, "on") || strings.HasPrefix(lower, "v-") {
		return false
	}
	switch lower {
	case "is", "ref", "key", "ref_key", "ref_for":
		return false
	}
	return true
}

func reservedProp(name string) bool {
	switch name {
	case "key", "ref", "ref_key", "ref_for", "__proto__":
		return true
	}
	return false
}

func IsVoidTag(tag string) bool {
	return strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+tag+" ")
}
func Validate(c *Component) error {
	bad := func(pos Position, message string) error {
		if pos.Line == 0 {
			pos = Position{Line: 1, Column: 1}
		}
		return diagnostic(pos, message)
	}
	if c == nil {
		return bad(Position{}, "missing component")
	}
	if (c.Version != 1 && c.Version != 2) || !identifierPattern.MatchString(c.Name) || !identifierPattern.MatchString(c.Factory) {
		return bad(Position{}, "invalid contract identity")
	}
	props := map[string]Prop{}
	slots := map[string]Slot{}
	nodes := map[string]Node{}
	for _, prop := range c.Props {
		if _, ok := props[prop.Name]; ok {
			return bad(prop.Pos, "duplicate prop")
		}
		if !identifierPattern.MatchString(prop.Name) {
			return bad(prop.Pos, "invalid prop name")
		}
		if reservedProp(prop.Name) {
			return bad(prop.Pos, "Vue-reserved prop name "+prop.Name)
		}
		if prop.Default != nil && !defaultMatches(prop) {
			return bad(prop.Pos, "invalid default")
		}
		props[prop.Name] = prop
	}
	for _, slot := range c.Slots {
		for _, field := range slot.Scope {
			if !identifierPattern.MatchString(field.Name) || field.Name == "__proto__" {
				return bad(field.Pos, "unsupported scoped field name "+field.Name)
			}
		}
		if _, ok := slots[slot.Name]; ok {
			return bad(slot.Pos, "duplicate slot")
		}
		slots[slot.Name] = slot
	}
	for _, node := range c.Nodes {
		if _, ok := nodes[node.ID]; ok {
			return bad(node.Pos, "duplicate node")
		}
		if DOMType(node.Tag) == "" || node.DOMType != DOMType(node.Tag) || node.Tag == "slot" {
			return bad(node.Pos, "unsupported native tag "+node.Tag)
		}
		if IsVoidTag(node.Tag) && len(node.Children) > 0 {
			return bad(node.Pos, "void element cannot have children")
		}
		nodes[node.ID] = node
	}
	root, ok := nodes[c.Root]
	if !ok {
		return bad(Position{}, "missing root node")
	}
	if c.RootType != root.DOMType && c.RootType != "HTMLElement" {
		return bad(root.Pos, "incorrect root DOM type")
	}
	placed := map[string]int{}
	slotUses := map[string]int{}
	for _, node := range c.Nodes {
		attributes := map[string]bool{}
		bindings := map[string]bool{}
		for _, attr := range node.Attributes {
			name := strings.ToLower(attr.Name)
			if !validAttribute(attr.Name) || attributes[name] {
				return bad(attr.Pos, "invalid or duplicate static attribute")
			}
			attributes[name] = true
		}
		for _, binding := range node.Bindings {
			prop, ok := props[binding.Prop]
			if !ok {
				return bad(binding.Pos, "unknown binding prop")
			}
			key := binding.Kind + ":" + binding.Name
			if binding.Kind == "attribute" {
				key = binding.Kind + ":" + strings.ToLower(binding.Name)
			}
			if bindings[key] || attributes[strings.ToLower(binding.Name)] {
				return bad(binding.Pos, "duplicate native binding")
			}
			bindings[key] = true
			if binding.Guarded != (prop.Optional && prop.Default == nil) {
				return bad(binding.Pos, "incorrect optional binding")
			}
			if binding.Kind == "attribute" {
				if !validAttribute(binding.Name) {
					return bad(binding.Pos, "invalid attribute name")
				}
			} else if binding.Kind == "property" {
				if binding.Name == "textContent" && len(node.Children) > 0 {
					return bad(binding.Pos, "textContent cannot be combined with children")
				}
				typ := propertyTypes[binding.Name]
				if typ == "" || !(typ == prop.Type || typ == "string-or-number" && (prop.Type == "string" || prop.Type == "number") || typ == "string" && strings.HasPrefix(prop.Type, "\"")) {
					return bad(binding.Pos, "unsupported or incompatible DOM property "+binding.Name)
				}
			} else {
				return bad(binding.Pos, "unknown binding kind")
			}
		}
		for _, child := range node.Children {
			if child.Node != "" {
				if child.Slot != "" {
					return bad(child.Pos, "ambiguous child")
				}
				if _, ok := nodes[child.Node]; !ok {
					return bad(child.Pos, "unknown child node")
				}
				placed[child.Node]++
			} else {
				slot, ok := slots[child.Slot]
				if !ok {
					return bad(child.Pos, "unknown child slot")
				}
				slotUses[child.Slot]++
				if child.Optional != slot.Optional || len(child.Scope) != len(slot.Scope) {
					return bad(child.Pos, "inconsistent slot scope")
				}
				seen := map[string]bool{}
				for _, mapping := range child.Scope {
					if seen[mapping.Name] {
						return bad(mapping.Pos, "duplicate scope field")
					}
					seen[mapping.Name] = true
					prop, ok := props[mapping.Prop]
					if !ok {
						return bad(mapping.Pos, "unknown scope prop")
					}
					found := false
					for _, field := range slot.Scope {
						if field.Name == mapping.Name && field.Type == prop.Type {
							found = true
						}
					}
					if !found {
						return bad(mapping.Pos, "incompatible slot scope")
					}
				}
			}
		}
	}
	if placed[c.Root] != 0 {
		return bad(root.Pos, "root inserted as a child")
	}
	for id, node := range nodes {
		if id != c.Root && placed[id] != 1 {
			return bad(node.Pos, "node must be placed exactly once")
		}
	}
	for name, slot := range slots {
		if slotUses[name] != 1 {
			return bad(slot.Pos, "slot must be placed exactly once")
		}
	}
	visited := map[string]bool{}
	var walk func(string, int) error
	walk = func(id string, depth int) error {
		n := nodes[id]
		if visited[id] || depth > 256 {
			return bad(n.Pos, "cyclic or excessively deep DOM tree")
		}
		visited[id] = true
		for _, child := range n.Children {
			if child.Node != "" {
				if err := walk(child.Node, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(c.Root, 0); err != nil {
		return err
	}
	if len(visited) != len(nodes) {
		return bad(root.Pos, "unreachable native node")
	}
	events := map[string]bool{}
	stateSources := map[string]string{}
	for _, event := range c.Events {
		node, ok := nodes[event.Target]
		if !ok {
			return bad(event.Pos, "unknown event target")
		}
		if !eventTypes[event.Type] || !identifierPattern.MatchString(event.Name) || events[event.Name] {
			return bad(event.Pos, "invalid or duplicate event")
		}
		events[event.Name] = true
		for _, read := range event.State {
			prop, ok := props[read.Prop]
			if !ok || readProperties[node.DOMType][read.Property] != prop.Type {
				return bad(read.Pos, "invalid native state read")
			}
			source := event.Target + "." + read.Property
			if old := stateSources[read.Prop]; old != "" && old != source {
				return bad(read.Pos, "ambiguous native state source")
			}
			stateSources[read.Prop] = source
			bound := false
			for _, b := range node.Bindings {
				if b.Prop == read.Prop {
					bound = true
				}
			}
			if !bound {
				return bad(read.Pos, "state field has no native binding")
			}
		}
	}
	return validateUI(c)
}
