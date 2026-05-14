package format

import (
	"strings"
	"testing"
)

func TestParseFrontmatter_Decision(t *testing.T) {
	src := `---
tags: [decision]
date: 2026-04-12
repo: tolvi
status: active
---

# Why postgres

Body content here.
`
	fm, body, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	if got, want := fm["status"], "active"; got != want {
		t.Errorf("status = %q, want %q", got, want)
	}
	if got, want := fm["repo"], "tolvi"; got != want {
		t.Errorf("repo = %q, want %q", got, want)
	}
	if got, want := fm["date"], "2026-04-12"; got != want {
		t.Errorf("date = %q, want %q", got, want)
	}
	if !strings.Contains(string(body), "Why postgres") {
		t.Errorf("body missing heading: %q", body)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	src := `# Just a heading

No frontmatter at all.
`
	_, _, err := ParseFrontmatter([]byte(src))
	if err == nil {
		t.Fatal("expected error for missing frontmatter delimiter")
	}
}

func TestParseFrontmatter_UnterminatedFrontmatter(t *testing.T) {
	src := `---
status: active
# (no closing delimiter)
`
	_, _, err := ParseFrontmatter([]byte(src))
	if err == nil {
		t.Fatal("expected error for unterminated frontmatter")
	}
}

func TestParseFrontmatter_InvalidYAML(t *testing.T) {
	src := `---
status: : :: nonsense
---

body
`
	_, _, err := ParseFrontmatter([]byte(src))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseFrontmatter_Tags(t *testing.T) {
	src := `---
tags: [pattern, golang, testing]
status: active
---

body
`
	fm, _, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}
	tags, ok := fm.Tags()
	if !ok {
		t.Fatal("Tags() returned not-ok")
	}
	if len(tags) != 3 {
		t.Fatalf("tags len = %d, want 3", len(tags))
	}
	if tags[0] != "pattern" {
		t.Errorf("tags[0] = %q, want pattern", tags[0])
	}
}

func TestRenderFrontmatter_RoundTrip(t *testing.T) {
	src := `---
tags: [decision]
date: 2026-04-12
repo: tolvi
status: active
---

# Body

Some body content.
`

	fm, body, err := ParseFrontmatter([]byte(src))
	if err != nil {
		t.Fatalf("ParseFrontmatter: %v", err)
	}

	out, err := RenderDocument(fm, body)
	if err != nil {
		t.Fatalf("RenderDocument: %v", err)
	}

	// Round-trip: parsing the output produces the same frontmatter map.
	fm2, body2, err := ParseFrontmatter(out)
	if err != nil {
		t.Fatalf("ParseFrontmatter(rendered): %v", err)
	}
	if fm.String("status") != fm2.String("status") {
		t.Errorf("status drift: %q vs %q", fm.String("status"), fm2.String("status"))
	}
	if fm.String("repo") != fm2.String("repo") {
		t.Errorf("repo drift")
	}
	if fm.String("date") != fm2.String("date") {
		t.Errorf("date drift")
	}
	if string(body) != string(body2) {
		t.Errorf("body drift: %q vs %q", body, body2)
	}
}

func TestRenderDocument_StableKeyOrder(t *testing.T) {
	fm := Frontmatter{
		"tags":   []any{"decision"},
		"date":   "2026-04-12",
		"repo":   "tolvi",
		"status": "active",
	}
	body := []byte("# Heading\n")

	out1, err := RenderDocument(fm, body)
	if err != nil {
		t.Fatalf("RenderDocument: %v", err)
	}
	out2, err := RenderDocument(fm, body)
	if err != nil {
		t.Fatalf("RenderDocument (second call): %v", err)
	}
	if string(out1) != string(out2) {
		t.Errorf("render not deterministic:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}
}
