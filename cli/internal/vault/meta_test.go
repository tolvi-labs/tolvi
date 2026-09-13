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
		SchemaVersion:  SupportedSchemaVersion,
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
		SchemaVersion:  SupportedSchemaVersion,
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
	if out.SchemaVersion != SupportedSchemaVersion {
		t.Errorf("schema_version = %d, want %d", out.SchemaVersion, SupportedSchemaVersion)
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

// --- meta v2 identity ------------------------------------------------------
//
// v2 keeps workspace as the container and repo as the member, matching the
// meanings the API and the published SDK already key uniqueness on, and adds
// an optional product. visibility and private_vault are gone: a vault no
// longer describes where its private content goes, roots.json does.

func TestSupportedSchemaVersionIsTwo(t *testing.T) {
	if SupportedSchemaVersion != 2 {
		t.Fatalf("SupportedSchemaVersion = %d, want 2", SupportedSchemaVersion)
	}
}

func TestReadMeta_RoundTripsIdentity(t *testing.T) {
	dir := t.TempDir()
	in := Meta{
		Workspace:      "tolvi-labs",
		Repo:           "tolvi",
		Product:        "stack",
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  SupportedSchemaVersion,
	}
	if err := WriteMeta(dir, in); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	out, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if out.Repo != "tolvi" {
		t.Errorf("repo drift: %q", out.Repo)
	}
	if out.Product != "stack" {
		t.Errorf("product drift: %q", out.Product)
	}
	if out.Workspace != "tolvi-labs" {
		t.Errorf("workspace drift: %q", out.Workspace)
	}
}

func TestReadMeta_RepoAndProductAreOptional(t *testing.T) {
	// A container vault holds docs for a whole workspace and names no repo.
	dir := t.TempDir()
	body := `{"workspace":"tolvi-labs","embedding_model":"nomic-embed-text","schema_version":2}`
	if err := os.WriteFile(filepath.Join(dir, ".vault-meta.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if out.Repo != "" || out.Product != "" {
		t.Errorf("expected empty repo/product, got %q/%q", out.Repo, out.Product)
	}
}

func TestWriteMeta_OmitsEmptyIdentityFields(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMeta(dir, Meta{Workspace: "x", EmbeddingModel: "nomic-embed-text"}); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".vault-meta.json"))
	for _, gone := range []string{`"repo"`, `"product"`, "visibility", "private_vault"} {
		if contains(string(data), gone) {
			t.Errorf("empty field %s should be omitted: %s", gone, data)
		}
	}
}

func TestReadMeta_RejectsSchemaVersionOne(t *testing.T) {
	// v0.2.0 is a breaking release with no shipped migrator, so a v1 vault must
	// fail with a message that says what to do rather than a bare mismatch.
	dir := t.TempDir()
	body := `{"workspace":"x","embedding_model":"nomic-embed-text","schema_version":1}`
	if err := os.WriteFile(filepath.Join(dir, ".vault-meta.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadMeta(dir)
	if err == nil {
		t.Fatal("expected a v1 meta to be rejected")
	}
	if !contains(err.Error(), "schema_version") {
		t.Errorf("error should name the field that is wrong: %v", err)
	}
}
