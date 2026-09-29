package generate

import (
	"fmt"
	"strings"
)

func (r *renderer) writeScript() {
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	if r.hasBindings() {
		p("import { ref as _htmlUiRef, watch as _htmlUiWatch, onMounted as _htmlUiOnMounted, onUpdated as _htmlUiOnUpdated, onBeforeUnmount as _htmlUiOnBeforeUnmount, watchEffect as _htmlUiWatchEffect } from 'vue'\n\n")
	}
	p("interface Props {\n")
	for _, prop := range r.c.Props {
		optional := ""
		if prop.Optional {
			optional = "?"
		}
		p("  %s%s: %s\n", js(prop.Name), optional, prop.Type)
	}
	p("}\nconst _htmlUiProps = withDefaults(defineProps<Props>(), {\n")
	for _, prop := range r.c.Props {
		if prop.Default != nil {
			p("  %s: %s,\n", js(prop.Name), normalizedLiteral(prop.Default))
		}
	}
	p("})\ndefineSlots<{\n")
	for _, slot := range r.c.Slots {
		optional := ""
		if slot.Optional {
			optional = "?"
		}
		p("  %s%s: (", js(slot.Name), optional)
		if len(slot.Scope) > 0 {
			p("scope: { ")
			for _, field := range slot.Scope {
				p("%s: %s; ", js(field.Name), field.Type)
			}
			p("}")
		}
		p(") => unknown\n")
	}
	p("}>()\n")
	r.writeEmits()
	r.writeState()
	r.writeNativeBindings()
	r.writeEvents()
	r.writeReset()
}
func (r *renderer) expression(prop string) string {
	if index := r.stateIndex(prop); index >= 0 {
		return fmt.Sprintf("_htmlUiState%d", index)
	}
	return "_htmlUiProps." + prop
}

func (r *renderer) nativeAttributeTypes() string {
	var b strings.Builder
	p := func(f string, args ...any) { fmt.Fprintf(&b, f, args...) }
	attributes := map[string]bool{}
	for _, node := range r.c.Nodes {
		for _, attr := range node.Attributes {
			attributes[attr.Name] = true
		}
	}
	if attributes["data-ui"] || attributes["popover"] || attributes["command"] {
		p("// Native attributes absent from the pinned Vue HTML types.\n")
		p("declare module 'vue' {\n")
		if attributes["data-ui"] || attributes["popover"] {
			p("  interface HTMLAttributes {\n")
			if attributes["data-ui"] {
				p("    'data-ui'?: string;\n    'data-ui-part'?: string;\n")
			}
			if attributes["popover"] {
				p("    popover?: '' | 'auto' | 'manual' | 'hint';\n")
			}
			p("  }\n")
		}
		if attributes["command"] {
			p("  interface ButtonHTMLAttributes { command?: string; }\n")
		}
		p("}\n\n")
	}

	return b.String()
}
