// Package cli implements the stream-only converter command.
package cli

import (
	"fmt"
	"html-ui-to-vue-vapor/internal/contract"
	"html-ui-to-vue-vapor/internal/generate"
	"io"
)

const Version = "0.1.0"
const usage = `Usage: html-ui-to-vue-vapor [--help | --version]

Read html-ui TypeScript source from stdin and write a Vue Vapor SFC to stdout.
Contract versions 1 and 2 are supported; v2 preserves UI metadata and styling types.
The source must contain contractVersion, Props, defaults, Slots, Events,
nativeEvents, and a constrained DOM factory. Input is never executed.

Example:
  html-ui accordion | html-ui-to-vue-vapor > Accordion.vue

Requires Vue 3.6 Vapor support in the consuming application.
No files, catalog, producer executable, or JavaScript runtime are needed.
`

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fail := func(code int, err error) int { fmt.Fprintf(stderr, "html-ui-to-vue-vapor: %v\n", err); return code }
	write := func(source string) int {
		n, err := io.WriteString(stdout, source)
		if err == nil && n != len(source) {
			err = io.ErrShortWrite
		}
		if err != nil {
			return fail(1, err)
		}
		return 0
	}
	if len(args) > 0 {
		if len(args) != 1 {
			return fail(2, fmt.Errorf("expected at most one flag; see --help"))
		}
		switch args[0] {
		case "--help", "-h":
			return write(usage)
		case "--version":
			return write("html-ui-to-vue-vapor " + Version + "\n")
		default:
			return fail(2, fmt.Errorf("unknown argument %q; see --help", args[0]))
		}
	}
	input, err := io.ReadAll(stdin)
	if err != nil {
		return fail(1, err)
	}
	c, err := contract.Parse(string(input))
	if err != nil {
		return fail(2, err)
	}
	result, err := generate.Generate(c)
	if err != nil {
		return fail(2, err)
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(stderr, "html-ui-to-vue-vapor: "+warning)
	}
	return write(result.Source)
}
