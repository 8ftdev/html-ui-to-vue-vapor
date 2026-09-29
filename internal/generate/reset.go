package generate

import "fmt"

func (r *renderer) hasFormControls() bool {
	for _, n := range r.c.Nodes {
		if r.nodeRef(n.ID) >= 0 && (n.DOMType == "HTMLInputElement" || n.DOMType == "HTMLTextAreaElement" || n.DOMType == "HTMLSelectElement") {
			return true
		}
	}
	return false
}
func (r *renderer) writeReset() {
	if len(r.fields) == 0 {
		return
	}
	p := func(f string, args ...any) { fmt.Fprintf(&r.script, f, args...) }
	p("let _htmlUiDisposed = false\n")
	if r.hasFormControls() {
		p("const _htmlUiForms = new Set<HTMLFormElement>()\n")
	}
	p("function _htmlUiReadNativeState() {\n")
	for i, field := range r.fields {
		read := field.Reads[0]
		ref := r.nodeRef(read.Target)
		p("  if (_htmlUiNode%d.value) {\n    const next = _htmlUiNode%d.value[%s]\n    if (!Object.is(_htmlUiState%d.value, next)) {\n      _htmlUiState%d.value = next\n      _htmlUiEmit(%s, next)\n    }\n  }\n", ref, ref, js(read.Property), i, i, js("update:"+field.Prop))
	}
	p("}\n")
	p("let _htmlUiReadQueued = false\nfunction _htmlUiScheduleNativeRead() {\n  if (_htmlUiReadQueued || _htmlUiDisposed) return\n  _htmlUiReadQueued = true\n  queueMicrotask(() => {\n    _htmlUiReadQueued = false\n    if (!_htmlUiDisposed) _htmlUiReadNativeState()\n  })\n}\n")
	if !r.hasFormControls() {
		p("_htmlUiOnMounted(_htmlUiReadNativeState)\n_htmlUiOnUpdated(_htmlUiReadNativeState)\n_htmlUiOnBeforeUnmount(() => { _htmlUiDisposed = true })\n")
		return
	}
	p("function _htmlUiOnReset(event: Event) {\n  queueMicrotask(() => {\n    if (_htmlUiDisposed || event.defaultPrevented) return\n    _htmlUiReadNativeState()\n  })\n}\n")
	p("function _htmlUiDetachResetListeners() {\n  for (const form of _htmlUiForms) form.removeEventListener('reset', _htmlUiOnReset)\n  _htmlUiForms.clear()\n}\n")
	p("function _htmlUiSyncForms() {\n  const next = new Set<HTMLFormElement>()\n")
	for i, node := range r.c.Nodes {
		if r.nodeRef(node.ID) < 0 {
			continue
		}
		if node.DOMType == "HTMLInputElement" || node.DOMType == "HTMLTextAreaElement" || node.DOMType == "HTMLSelectElement" {
			p("  if (_htmlUiNode%d.value?.form) next.add(_htmlUiNode%d.value.form)\n", i, i)
		}
	}
	p("  for (const form of _htmlUiForms) if (!next.has(form)) form.removeEventListener('reset', _htmlUiOnReset)\n  for (const form of next) if (!_htmlUiForms.has(form)) form.addEventListener('reset', _htmlUiOnReset)\n  _htmlUiForms.clear()\n  for (const form of next) _htmlUiForms.add(form)\n}\n")
	p("_htmlUiOnMounted(() => { _htmlUiSyncForms(); _htmlUiReadNativeState() })\n_htmlUiOnUpdated(() => { _htmlUiSyncForms(); _htmlUiReadNativeState() })\n_htmlUiOnBeforeUnmount(() => { _htmlUiDisposed = true; _htmlUiDetachResetListeners() })\n")
}
