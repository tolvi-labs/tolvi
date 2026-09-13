// Package vault provides vault discovery, doc loading, and workspace
// metadata read/write.
package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SupportedSchemaVersion is the only vault format version this CLI
// understands. A vault tagged with a different version is rejected at read
// time with a clear "this CLI is too old/new" message.
//
// v2 replaced the two-root routing fields with identity: a vault says who it
// is, and ~/.config/tolvi/roots.json says where the roots are.
const SupportedSchemaVersion = 2

// Meta is the marshalled form of <vault>/.vault-meta.json.
//
// Fields are emitted in this struct order; ReadMeta validates required
// fields are present and that SchemaVersion matches SupportedSchemaVersion.
type Meta struct {
	// Workspace is the container this vault belongs to, and the unit the
	// server keys multi-tenant isolation on.
	Workspace string `json:"workspace"`
	// Repo is the member within that workspace. Optional: a container vault
	// serves a whole workspace and names no repo.
	Repo string `json:"repo,omitempty"`
	// Product names a group of repos that genuinely share a brain. Optional,
	// and declared only where such a group exists.
	Product        string `json:"product,omitempty"`
	EmbeddingModel string `json:"embedding_model"`
	SchemaVersion  int    `json:"schema_version"`
}

// Identity returns the parts of a meta that resolve a chain of roots.
func (m Meta) Identity() Identity {
	return Identity{Workspace: m.Workspace, Repo: m.Repo, Product: m.Product}
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
			"%s: schema_version is %d but this CLI only supports %d — upgrade the CLI, or migrate the vault by replacing visibility/private_vault with repo and declaring the roots in %s",
			path, m.SchemaVersion, SupportedSchemaVersion, RootsConfigFileName,
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
