// Package vault provides vault discovery, doc loading, and workspace
// metadata read/write.
package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SupportedSchemaVersion is the only tolvi-format-v1 schema version this
// CLI understands. A vault tagged with a different version is rejected
// at read time with a clear "this CLI is too old/new" message.
const SupportedSchemaVersion = 1

// Meta is the marshalled form of <vault>/.vault-meta.json.
//
// Fields are emitted in this struct order; ReadMeta validates required
// fields are present and that SchemaVersion matches SupportedSchemaVersion.
type Meta struct {
	Workspace      string `json:"workspace"`
	EmbeddingModel string `json:"embedding_model"`
	SchemaVersion  int    `json:"schema_version"`
}

// ReadMeta parses <vaultPath>/.vault-meta.json.
func ReadMeta(vaultPath string) (Meta, error) {
	path := filepath.Join(vaultPath, ".vault-meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, fmt.Errorf("read %s: %w", path, err)
	}
	var m Meta
	if err := json.Unmarshal(data, &m); err != nil {
		return Meta{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.Workspace == "" {
		return Meta{}, fmt.Errorf("%s: workspace field is required", path)
	}
	if m.EmbeddingModel == "" {
		return Meta{}, fmt.Errorf("%s: embedding_model field is required", path)
	}
	if m.SchemaVersion != SupportedSchemaVersion {
		return Meta{}, fmt.Errorf(
			"%s: schema_version is %d but this CLI only supports %d — upgrade the CLI or migrate the vault",
			path, m.SchemaVersion, SupportedSchemaVersion,
		)
	}
	return m, nil
}

// WriteMeta writes <vaultPath>/.vault-meta.json. Overwrites if it exists;
// callers that need create-only semantics (like tolvi init) check first.
//
// JSON is pretty-printed with 2-space indent + trailing newline to match
// the convention in examples/sample-vault/.vault-meta.json.
func WriteMeta(vaultPath string, m Meta) error {
	if m.SchemaVersion == 0 {
		m.SchemaVersion = SupportedSchemaVersion
	}
	if m.EmbeddingModel == "" {
		m.EmbeddingModel = "nomic-embed-text"
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(vaultPath, ".vault-meta.json")
	return os.WriteFile(path, data, 0o644)
}
