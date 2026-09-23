// Package format provides frontmatter parsing/rendering and JSON Schema
// validation against the four tolvi-format-v2 schemas, embedded at build
// time so the CLI works offline.
package format

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Embedded schemas. Paths are relative to this file's directory at
// compile time. The schemas live as byte-identical copies under
// schemas/ — CI verifies they match the canonical sources at
// <repo-root>/spec/schemas/.
//
// Why duplicated rather than embedded from ../../../spec/schemas/:
// Go's //go:embed cannot traverse upward from the source file's
// directory.

//go:embed schemas/vault-meta.json
var VaultMetaSchema []byte

//go:embed schemas/decision.json
var DecisionSchema []byte

//go:embed schemas/session.json
var SessionSchema []byte

//go:embed schemas/pattern.json
var PatternSchema []byte

// Output-contract schemas. These describe what commands print with --json
// rather than what a vault document contains, so they are versioned with the
// CLI rather than with tolvi-format. They are embedded so the CLI's own tests
// can validate real output against the published shape.

//go:embed schemas/doctor.json
var DoctorSchema []byte

//go:embed schemas/vault-health.json
var VaultHealthSchema []byte

//go:embed schemas/repos-list.json
var ReposListSchema []byte

//go:embed schemas/packs-list.json
var PacksListSchema []byte

// ValidatorForDocType returns a compiled JSON Schema validator for one
// of "decision" | "session" | "pattern". Unknown types return an error.
func ValidatorForDocType(docType string) (*jsonschema.Schema, error) {
	var raw []byte
	var name string
	switch docType {
	case "decision":
		raw, name = DecisionSchema, "decision.json"
	case "session":
		raw, name = SessionSchema, "session.json"
	case "pattern":
		raw, name = PatternSchema, "pattern.json"
	default:
		return nil, fmt.Errorf("unknown doc type %q (want decision|session|pattern)", docType)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(name, bytes.NewReader(raw)); err != nil {
		return nil, fmt.Errorf("add schema resource: %w", err)
	}
	return compiler.Compile(name)
}
