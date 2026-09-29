package generate

import "fmt"

func (r *renderer) writeEmits() {
	if len(r.c.Events) == 0 && len(r.fields) == 0 {
		return
	}
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	p("const _htmlUiEmit = defineEmits<{\n")
	for _, event := range r.c.Events {
		p("  %s: [event: %s]\n", js(event.Name), event.Type)
	}
	for _, field := range r.fields {
		p("  %s: [value: %s]\n", js("update:"+field.Prop), field.Type)
	}
	p("}>()\n")
}
func (r *renderer) writeEvents() {
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	for i, event := range r.c.Events {
		p("function _htmlUiEvent%d(event: %s) {\n", i, event.Type)
		if len(event.State) > 0 {
			p("  const node = event.currentTarget as %s\n", r.node(event.Target).DOMType)
		}
		for _, read := range event.State {
			index := r.stateIndex(read.Prop)
			p("  if (!Object.is(_htmlUiState%d.value, node[%s])) {\n", index, js(read.Property))
			p("    _htmlUiState%d.value = node[%s]\n    _htmlUiEmit(%s, _htmlUiState%d.value)\n  }\n", index, js(read.Property), js("update:"+read.Prop), index)
		}
		p("  _htmlUiEmit(%s, event)\n}\n", js(event.Name))
	}
}
