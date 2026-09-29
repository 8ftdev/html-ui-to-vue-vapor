package generate

import (
	"encoding/json"
	"html"
)

func js(value string) string      { b, _ := json.Marshal(value); return string(b) }
func escaped(value string) string { return html.EscapeString(value) }
func normalizedLiteral(raw json.RawMessage) string {
	var value any
	_ = json.Unmarshal(raw, &value)
	b, _ := json.Marshal(value)
	return string(b)
}
