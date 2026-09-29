package contract

import (
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	Kind, Text string
	Pos        Position
}

func lex(source string) ([]token, error) {
	tokens := []token{}
	offset, line, column := 0, 1, 1
	advance := func() {
		r, size := utf8.DecodeRuneInString(source[offset:])
		offset += size
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	position := func() Position { return Position{offset, line, column} }
	for offset < len(source) {
		start := position()
		r, size := utf8.DecodeRuneInString(source[offset:])
		if r == utf8.RuneError && size == 1 {
			return nil, diagnostic(start, "invalid UTF-8")
		}
		if unicode.IsSpace(r) {
			advance()
			continue
		}
		if strings.HasPrefix(source[offset:], "//") {
			for offset < len(source) && source[offset] != '\n' {
				advance()
			}
			continue
		}
		if strings.HasPrefix(source[offset:], "/*") {
			doc := strings.HasPrefix(source[offset:], "/**")
			advance()
			advance()
			for offset < len(source) && !strings.HasPrefix(source[offset:], "*/") {
				advance()
			}
			if offset == len(source) {
				return nil, diagnostic(start, "unterminated comment")
			}
			advance()
			advance()
			if doc {
				tokens = append(tokens, token{"doc", source[start.Offset:offset], start})
			}
			continue
		}
		if r == '"' || r == '\'' {
			quote := r
			advance()
			var b strings.Builder
			b.WriteByte('"')
			closed := false
			for offset < len(source) {
				ch := source[offset]
				if rune(ch) == quote {
					advance()
					closed = true
					break
				}
				if ch == '\n' || ch == '\r' {
					return nil, diagnostic(start, "newline in string literal")
				}
				if ch == '\\' {
					advance()
					if offset == len(source) {
						break
					}
					esc := source[offset]
					if quote == '\'' && esc == '\'' {
						b.WriteByte('\'')
						advance()
						continue
					}
					b.WriteByte('\\')
					b.WriteByte(esc)
					advance()
					continue
				}
				if ch == '"' {
					b.WriteString(`\"`)
					advance()
					continue
				}
				old := offset
				advance()
				b.WriteString(source[old:offset])
			}
			if !closed {
				return nil, diagnostic(start, "unterminated string")
			}
			b.WriteByte('"')
			var value string
			if err := json.Unmarshal([]byte(b.String()), &value); err != nil {
				return nil, diagnostic(start, "invalid string escape")
			}
			tokens = append(tokens, token{"string", value, start})
			continue
		}
		if unicode.IsLetter(r) || r == '_' || r == '$' {
			advance()
			for offset < len(source) {
				next, _ := utf8.DecodeRuneInString(source[offset:])
				if !unicode.IsLetter(next) && !unicode.IsDigit(next) && next != '_' && next != '$' {
					break
				}
				advance()
			}
			tokens = append(tokens, token{"identifier", source[start.Offset:offset], start})
			continue
		}
		if r >= '0' && r <= '9' {
			advance()
			for offset < len(source) && strings.ContainsRune("0123456789.eE", rune(source[offset])) {
				advance()
				if offset < len(source) && (source[offset-1] == 'e' || source[offset-1] == 'E') && (source[offset] == '+' || source[offset] == '-') {
					advance()
				}
			}
			tokens = append(tokens, token{"number", source[start.Offset:offset], start})
			continue
		}
		matched := false
		for _, op := range []string{"!==", "=>"} {
			if strings.HasPrefix(source[offset:], op) {
				for range op {
					advance()
				}
				tokens = append(tokens, token{"punct", op, start})
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if strings.ContainsRune("{}()[]:;,?.=<>|-", r) {
			advance()
			tokens = append(tokens, token{"punct", string(r), start})
			continue
		}
		return nil, diagnostic(start, "unsupported character "+string(r))
	}
	tokens = append(tokens, token{"eof", "", position()})
	return tokens, nil
}
