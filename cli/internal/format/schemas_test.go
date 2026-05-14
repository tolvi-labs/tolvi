package format

import (
	"strings"
	"testing"
)

func TestEmbeddedSchemas(t *testing.T) {
	for _, tc := range []struct {
		name      string
		schema    []byte
		wantIDSub string
	}{
		{"vault-meta", VaultMetaSchema, "vault-meta.json"},
		{"decision", DecisionSchema, "decision.json"},
		{"session", SessionSchema, "session.json"},
		{"pattern", PatternSchema, "pattern.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.schema) == 0 {
				t.Fatal("schema bytes are empty")
			}
			if !strings.Contains(string(tc.schema), tc.wantIDSub) {
				t.Fatalf("schema does not contain expected $id substring %q", tc.wantIDSub)
			}
			if !strings.Contains(string(tc.schema), "https://tolvilabs.com/tolvi/spec/schemas/") {
				t.Fatalf("schema $id is not on tolvilabs.com — spec lock broken")
			}
		})
	}
}

func TestValidatorForDocType(t *testing.T) {
	for _, dt := range []string{"decision", "session", "pattern"} {
		t.Run(dt, func(t *testing.T) {
			v, err := ValidatorForDocType(dt)
			if err != nil {
				t.Fatalf("ValidatorForDocType(%q): %v", dt, err)
			}
			if v == nil {
				t.Fatalf("nil validator returned")
			}
		})
	}
}

func TestValidatorForDocType_Unknown(t *testing.T) {
	_, err := ValidatorForDocType("nonsense")
	if err == nil {
		t.Fatal("expected error for unknown doc type")
	}
}
