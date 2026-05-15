package citations

import (
	"reflect"
	"testing"
)

func TestExtract_UniqueSlugs(t *testing.T) {
	text := "See [[idempotent-migrations]] and [[postgres-locks]] for details."
	got := Extract(text)
	want := []string{"idempotent-migrations", "postgres-locks"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Extract = %v, want %v", got, want)
	}
}

func TestExtract_Dedupe(t *testing.T) {
	got := Extract("[[a]] then [[b]] then [[a]] again")
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Extract = %v, want %v", got, want)
	}
}

func TestExtract_IgnoresNonWikiLinkBrackets(t *testing.T) {
	got := Extract("Single [brackets] and [[but-not-this stop early.")
	if len(got) != 0 {
		t.Errorf("Extract returned: %v, want empty", got)
	}
}

func TestExtract_SlugShape(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"[[foo bar]]", nil},       // space disallowed
		{"[[FOO]]", nil},           // uppercase disallowed
		{"[[foo-]]", nil},          // trailing hyphen disallowed
		{"[[foo-bar]]", []string{"foo-bar"}},
		{"[[a]]", []string{"a"}},   // single char allowed
		{"[[task-42]]", []string{"task-42"}}, // digits in middle allowed
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Extract(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("Extract(%q) = %v, want %v", tt.in, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Extract(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestVerify_PartitionsMatchAndMismatch(t *testing.T) {
	known := map[string]bool{"real-slug": true}
	matched, unmatched := Verify([]string{"real-slug", "fake-slug"}, known)
	if !reflect.DeepEqual(matched, []string{"real-slug"}) {
		t.Errorf("matched = %v", matched)
	}
	if !reflect.DeepEqual(unmatched, []string{"fake-slug"}) {
		t.Errorf("unmatched = %v", unmatched)
	}
}
