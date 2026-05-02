package parse

import (
	"strings"
	"testing"
)

func TestParseRPSLContinuation(t *testing.T) {
	input := `
organisation: ORG-EX1-RIPE
org-name: Example Org
country: RU
remarks: first line
 second line

`

	var got Object
	err := ParseRPSL(strings.NewReader(input), map[string]struct{}{"organisation": {}}, func(obj Object) error {
		got = obj
		return nil
	})
	if err != nil {
		t.Fatalf("ParseRPSL() error = %v", err)
	}
	if got.Type != "organisation" {
		t.Fatalf("Type = %q", got.Type)
	}
	if got.Attributes["remarks"][0] != "first line second line" {
		t.Fatalf("remarks = %q", got.Attributes["remarks"][0])
	}
}
