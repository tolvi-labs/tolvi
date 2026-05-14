package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteMeta_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	meta := Meta{
		Workspace:      "my-project",
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  1,
	}
	if err := WriteMeta(dir, meta); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".vault-meta.json"))
	if err != nil {
		t.Fatalf("readback: %v", err)
	}
	if got := string(data); !contains(got, `"workspace": "my-project"`) {
		t.Errorf("written meta missing workspace: %s", got)
	}
}

func TestReadMeta_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := Meta{
		Workspace:      "round-trip",
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  1,
	}
	if err := WriteMeta(dir, in); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	out, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if out.Workspace != in.Workspace {
		t.Errorf("workspace drift: %q vs %q", out.Workspace, in.Workspace)
	}
	if out.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", out.SchemaVersion)
	}
}

func TestReadMeta_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadMeta(dir)
	if err == nil {
		t.Fatal("expected error on missing .vault-meta.json")
	}
}

func TestReadMeta_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".vault-meta.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadMeta(dir)
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}

func TestReadMeta_SchemaVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".vault-meta.json"),
		[]byte(`{"workspace":"x","embedding_model":"nomic-embed-text","schema_version":99}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadMeta(dir)
	if err == nil {
		t.Fatal("expected error on schema_version=99")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
