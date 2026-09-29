package contract

import (
	"errors"
	"testing"
)

func TestLexerPreservesLiteralAndCRLFPosition(t *testing.T) {
	ts, err := lex("// heading\r\nexport const x = \"a\\\"&é\";")
	if err != nil {
		t.Fatal(err)
	}
	if ts[0].Text != "export" || ts[0].Pos.Line != 2 || ts[0].Pos.Column != 1 {
		t.Fatalf("%#v", ts[0])
	}
	found := false
	for _, tok := range ts {
		if tok.Kind == "string" && tok.Text == "a\"&é" {
			found = true
		}
	}
	if !found {
		t.Fatal("literal missing")
	}
}
func TestLexerRejectsTruncation(t *testing.T) {
	for _, source := range []string{`export const x = "unfinished`, "/* unfinished", "`template`", "\xff"} {
		_, err := lex(source)
		var d *Diagnostic
		if !errors.As(err, &d) || d.Pos.Line < 1 || d.Pos.Column < 1 {
			t.Fatalf("%q: %v", source, err)
		}
	}
}
func TestLexerEscapesAndDocumentation(t *testing.T) {
	ts, err := lex("/** Behavior: native. */\n'\\u0061\\n\\\\' -2.5 as const")
	if err != nil {
		t.Fatal(err)
	}
	if ts[0].Kind != "doc" || ts[1].Text != "a\n\\" || ts[2].Text != "-" || ts[3].Text != "2.5" {
		t.Fatalf("%#v", ts)
	}
}
