package generate

import (
	"html-ui-to-vue-vapor/internal/testinput"
	"strings"
	"testing"
)

func TestDisclosureOmitsFormResetListeners(t *testing.T) {
	s := generated(t, parsed(t, testinput.Disclosure)).Source
	if strings.Contains(s, "addEventListener('reset'") {
		t.Fatal("disclosure emits unused form reset listeners")
	}
}
