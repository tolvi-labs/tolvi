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

func TestReadMeta_RoundTrip_VisibilityFields(t *testing.T) {
	dir := t.TempDir()
	in := Meta{
		Workspace:      "public-repo",
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  1,
		Visibility:     "public",
		PrivateVault:   "../private/vault",
	}
	if err := WriteMeta(dir, in); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	out, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if out.Visibility != "public" {
		t.Errorf("visibility drift: %q", out.Visibility)
	}
	if out.PrivateVault != "../private/vault" {
		t.Errorf("private_vault drift: %q", out.PrivateVault)
	}
}

func TestReadMeta_BackwardCompat_NoVisibilityFields(t *testing.T) {
	dir := t.TempDir()
	// A meta written before the visibility fields existed must still parse.
	legacy := `{"workspace":"x","embedding_model":"nomic-embed-text","schema_version":1}`
	if err := os.WriteFile(filepath.Join(dir, ".vault-meta.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta on legacy meta: %v", err)
	}
	if out.Visibility != "" {
		t.Errorf("expected empty visibility, got %q", out.Visibility)
	}
	if out.PrivateVault != "" {
		t.Errorf("expected empty private_vault, got %q", out.PrivateVault)
	}
}

func TestWriteMeta_OmitsEmptyVisibilityFields(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMeta(dir, Meta{Workspace: "x", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1}); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".vault-meta.json"))
	if contains(string(data), "visibility") || contains(string(data), "private_vault") {
		t.Errorf("empty visibility fields should be omitted: %s", data)
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

// --- local routing overlay -------------------------------------------------
//
// .vault-routing.local.json is the git-ignored, machine-local routing config
// that internal-dev checkouts of a public repo carry. ReadMeta must honor it,
// so the CLI and the /tolvi-sync skill agree on where sessions land.

func writeRoutingConfig(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, RoutingConfigFileName), []byte(body), 0o644); err != nil {
		t.Fatalf("write routing config: %v", err)
	}
}

func baseMeta(t *testing.T, dir string) {
	t.Helper()
	if err := WriteMeta(dir, Meta{Workspace: "tolvi", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1}); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
}

func TestReadMeta_LocalRoutingOverlay(t *testing.T) {
	dir := t.TempDir()
	baseMeta(t, dir)
	writeRoutingConfig(t, dir, `{"private_vault": "/private/vault"}`)

	m, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if m.Visibility != "public" {
		t.Errorf("visibility = %q, want %q (routing config presence marks the vault public)", m.Visibility, "public")
	}
	if m.PrivateVault != "/private/vault" {
		t.Errorf("private_vault = %q, want %q", m.PrivateVault, "/private/vault")
	}
}

func TestReadMeta_NoRoutingConfig_LeavesVisibilityUnset(t *testing.T) {
	dir := t.TempDir()
	baseMeta(t, dir)

	m, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if m.Visibility != "" || m.PrivateVault != "" {
		t.Errorf("absent routing config changed meta: visibility=%q private_vault=%q", m.Visibility, m.PrivateVault)
	}
}

func TestReadMeta_LocalRoutingOverlay_WinsOverMeta(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMeta(dir, Meta{
		Workspace:      "tolvi",
		EmbeddingModel: "nomic-embed-text",
		SchemaVersion:  1,
		Visibility:     "public",
		PrivateVault:   "/from/meta",
	}); err != nil {
		t.Fatalf("WriteMeta: %v", err)
	}
	writeRoutingConfig(t, dir, `{"private_vault": "/from/local"}`)

	m, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	if m.PrivateVault != "/from/local" {
		t.Errorf("private_vault = %q, want the machine-local override %q", m.PrivateVault, "/from/local")
	}
}

func TestReadMeta_MalformedRoutingConfig_Errors(t *testing.T) {
	dir := t.TempDir()
	baseMeta(t, dir)
	writeRoutingConfig(t, dir, `{"private_vault":`)

	if _, err := ReadMeta(dir); err == nil {
		t.Fatal("expected an error for malformed routing config — silently ignoring it writes session notes into a public repo")
	}
}

func TestReadMeta_EmptyRoutingConfig_Errors(t *testing.T) {
	dir := t.TempDir()
	baseMeta(t, dir)
	writeRoutingConfig(t, dir, `{}`)

	if _, err := ReadMeta(dir); err == nil {
		t.Fatal("expected an error when the routing config omits private_vault")
	}
}

func TestReadMeta_LocalRouting_RoutesSessionsPrivate(t *testing.T) {
	dir := t.TempDir()
	baseMeta(t, dir)
	writeRoutingConfig(t, dir, `{"private_vault": "/private/vault"}`)

	m, err := ReadMeta(dir)
	if err != nil {
		t.Fatalf("ReadMeta: %v", err)
	}
	root, routed, err := ResolveDocDestination(dir, m, "session", "")
	if err != nil {
		t.Fatalf("ResolveDocDestination: %v", err)
	}
	if !routed || root != "/private/vault" {
		t.Errorf("session routed=%v root=%q, want true and %q", routed, root, "/private/vault")
	}
}
