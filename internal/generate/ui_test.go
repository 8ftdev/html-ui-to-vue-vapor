package generate_test

import (
	"html-ui-to-vue-vapor/internal/contract"
	"html-ui-to-vue-vapor/internal/generate"
	"os"
	"strings"
	"testing"
)

func TestV2PreservesThemeContract(t *testing.T) {
	source, err := os.ReadFile("../testinput/accordion-v2.ts")
	if err != nil {
		t.Fatal(err)
	}
	c, err := contract.Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	output, err := generate.Generate(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<script lang="ts">`, `export const contractVersion = 2 as const;`, `export const ui =`, `export interface AccordionClasses<Style = string>`, `data-ui="accordion"`, `data-ui-part="trigger"`, `"node": "summary"`, `"attribute": "open"`} {
		if !strings.Contains(output.Source, want) {
			t.Fatalf("missing %s", want)
		}
	}
}
