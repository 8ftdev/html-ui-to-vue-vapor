package contract

import (
	"os"
	"strings"
	"testing"
)

func TestRejectsInvalidUIContracts(t *testing.T) {
	data, err := os.ReadFile("../testinput/accordion-v2.ts")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, pair := range [][2]string{
		{`"present": true`, `"present": true, "value": null`},
		{`"requirements": [`, `"requirements": [null,`},
		{`"node": "summary"`, `"node": "missing"`},
		{`"pseudo": "hover"`, `"pseudo": "made-up"`},
		{`"attribute": "open"`, `"attribute": "open", "value": "true"`},
		{`"component": "accordion"`, `"component": "accordion", "unknown": true`},
		{`"component": "accordion"`, `"component": "accordion", "component": "other"`},
		{`"kind": "native"`, `"kind": "adapter-required"`},
		{`"component": "accordion"`, `"component": arbitraryCall()`},
		{`trigger?: {`, `invented?: {`},
		{`hover?: AccordionStyle<Style>`, `disabled?: AccordionStyle<Style>`},
		{`summary.setAttribute("data-ui-part", "trigger");`, `summary.setAttribute("data-ui-part", "wrong");`},
	} {
		altered := strings.Replace(source, pair[0], pair[1], 1)
		if altered == source {
			t.Fatalf("test mutation missing: %s", pair[0])
		}
		if _, err := Parse(altered); err == nil {
			t.Errorf("accepted %s", pair[1])
		}
	}
}
