package cli

import (
	"bytes"
	"errors"
	"html-ui-to-vue-vapor/internal/testinput"
	"io"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestCLIStream(t *testing.T) {
	var out, errout bytes.Buffer
	if code := Run(nil, strings.NewReader(testinput.Disclosure), &out, &errout); code != 0 || !strings.Contains(out.String(), "<template>") || errout.Len() != 0 {
		t.Fatalf("%d %s %s", code, out.String(), errout.String())
	}
}
func TestInvalidInputDoesNotWriteStdout(t *testing.T) {
	for _, source := range []string{"", testinput.Disclosure + "call();", strings.Replace(testinput.Disclosure, "contractVersion = 1", "contractVersion = 99", 1)} {
		var out, errout bytes.Buffer
		code := Run(nil, strings.NewReader(source), &out, &errout)
		if code != 2 || out.Len() != 0 || errout.Len() == 0 {
			t.Fatalf("%d %q %q", code, out.String(), errout.String())
		}
	}
}
func TestHelpAndVersionDoNotReadInput(t *testing.T) {
	for _, arg := range []string{"--help", "-h", "--version"} {
		var out bytes.Buffer
		if code := Run([]string{arg}, brokenReader{}, &out, io.Discard); code != 0 || out.Len() == 0 {
			t.Fatal(code)
		}
	}
}
func TestIOAndArgumentErrors(t *testing.T) {
	if Run(nil, brokenReader{}, io.Discard, io.Discard) != 1 {
		t.Fatal("read failure code")
	}
	if Run(nil, strings.NewReader(testinput.Disclosure), shortWriter{}, io.Discard) != 1 {
		t.Fatal("short write code")
	}
	for _, args := range [][]string{{"--unknown"}, {"example"}, {"--help", "--version"}} {
		if Run(args, brokenReader{}, io.Discard, io.Discard) != 2 {
			t.Fatal(args)
		}
	}
}
func TestAdapterWarningStaysOnStderr(t *testing.T) {
	source := strings.Replace(testinput.Static, "Behavior: native", "Behavior: adapter-required", 1)
	var out, errout bytes.Buffer
	if code := Run(nil, strings.NewReader(source), &out, &errout); code != 0 || !strings.Contains(errout.String(), "adapter-required") || !strings.HasPrefix(out.String(), "<!--") {
		t.Fatalf("%d %s %s", code, out.String(), errout.String())
	}
}
