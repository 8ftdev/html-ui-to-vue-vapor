package contract

import (
	"regexp"
	"strings"
)

type parser struct {
	tokens   []token
	index    int
	behavior string
}

var behaviorPattern = regexp.MustCompile(`\bBehavior:\s*(native|adapter-required)\b`)

func newParser(source string) (*parser, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}
	p := &parser{behavior: "unknown"}
	for _, tok := range tokens {
		if tok.Kind == "doc" {
			if len(p.tokens) == 0 {
				if m := behaviorPattern.FindStringSubmatch(tok.Text); m != nil {
					p.behavior = m[1]
				}
			}
			continue
		}
		p.tokens = append(p.tokens, tok)
	}
	return p, nil
}
func (p *parser) peek() token { return p.tokens[p.index] }
func (p *parser) take() token {
	t := p.peek()
	if t.Kind != "eof" {
		p.index++
	}
	return t
}
func (p *parser) fail(message string) { panic(diagnostic(p.peek().Pos, message)) }
func (p *parser) want(values ...string) {
	for _, value := range values {
		if p.peek().Text != value || p.peek().Kind == "eof" {
			p.fail("expected " + value)
		}
		p.take()
	}
}
func (p *parser) accept(value string) bool {
	if p.peek().Text == value && p.peek().Kind != "eof" {
		p.take()
		return true
	}
	return false
}
func (p *parser) identifier() string {
	if p.peek().Kind != "identifier" {
		p.fail("expected identifier")
	}
	return p.take().Text
}
func (p *parser) stringLiteral() string {
	if p.peek().Kind != "string" {
		p.fail("expected string literal")
	}
	return p.take().Text
}
func recoverDiagnostic(err *error) {
	if caught := recover(); caught != nil {
		if d, ok := caught.(*Diagnostic); ok {
			*err = d
			return
		}
		panic(caught)
	}
}
func (p *parser) fields() []Field {
	p.want("{")
	var fields []Field
	seen := map[string]bool{}
	for !p.accept("}") {
		f := Field{Pos: p.peek().Pos, Name: p.identifier()}
		if seen[f.Name] {
			p.fail("duplicate field " + f.Name)
		}
		seen[f.Name] = true
		f.Optional = p.accept("?")
		p.want(":")
		f.Type = p.propType()
		fields = append(fields, f)
		if !p.accept(";") && !p.accept(",") && p.peek().Text != "}" {
			p.fail("expected field separator")
		}
	}
	return fields
}
func (p *parser) propType() string {
	if p.peek().Kind == "string" {
		var values []string
		for {
			values = append(values, quote(p.stringLiteral()))
			if !p.accept("|") {
				break
			}
			if p.peek().Kind != "string" {
				p.fail("expected literal union member")
			}
		}
		return strings.Join(values, " | ")
	}
	typ := p.identifier()
	if typ != "string" && typ != "boolean" && typ != "number" {
		p.fail("unsupported prop type " + typ)
	}
	return typ
}
