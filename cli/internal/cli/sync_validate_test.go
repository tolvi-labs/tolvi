package cli

import (
	"testing"
	"time"
)

func TestValidateFrontmatter_Decision_Valid(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "decision",
		Title:     "Why postgres",
		Slug:      "why-postgres",
		Workspace: "tolvi",
		Date:      time.Now(),
		Status:    "active",
	})
	if err := ValidateFrontmatter("decision", fm); err != nil {
		t.Errorf("ValidateFrontmatter: %v", err)
	}
}

func TestValidateFrontmatter_Pattern_Valid(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "pattern",
		Title:     "Idempotent migrations",
		Slug:      "idempotent-migrations",
		Workspace: "tolvi",
		Date:      time.Now(),
		Status:    "active",
	})
	if err := ValidateFrontmatter("pattern", fm); err != nil {
		t.Errorf("ValidateFrontmatter: %v", err)
	}
}

func TestValidateFrontmatter_InvalidStatus(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "decision",
		Title:     "x",
		Slug:      "x",
		Workspace: "t",
		Date:      time.Now(),
		Status:    "active",
	})
	fm["status"] = "not-a-real-status"
	if err := ValidateFrontmatter("decision", fm); err == nil {
		t.Fatal("expected validation error for invalid status enum")
	}
}

func TestValidateFrontmatter_MissingRequiredField(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "decision",
		Title:     "x",
		Slug:      "x",
		Workspace: "t",
		Date:      time.Now(),
		Status:    "active",
	})
	delete(fm, "repo") // required for decisions
	if err := ValidateFrontmatter("decision", fm); err == nil {
		t.Fatal("expected validation error for missing required field")
	}
}
