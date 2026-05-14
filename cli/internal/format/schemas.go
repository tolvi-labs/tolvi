// Package format provides frontmatter parsing/rendering and JSON Schema
// validation against the four tolvi-format-v1 schemas, embedded at build
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
