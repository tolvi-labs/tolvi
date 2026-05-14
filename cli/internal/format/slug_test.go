package format

import "testing"

func TestSlugFromTitle(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Why we chose Postgres", "why-we-chose-postgres"},
		{"Why we chose Postgres over MySQL", "why-we-chose-postgres-over-mysql"},
		{"  leading/trailing whitespace  ", "leading-trailing-whitespace"},
		{"Snake_case_and-dashes", "snake-case-and-dashes"},
		{"Numbers: 42 and 100", "numbers-42-and-100"},
		{"Émoji 🎉 strip", "emoji-strip"},
		{"---multiple--- dashes", "multiple-dashes"},
		{"!!!special?chars!!!", "special-chars"},
		{"a", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := SlugFromTitle(tt.in); got != tt.want {
				t.Errorf("SlugFromTitle(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSlugFromTitle_MaxLength(t *testing.T) {
	long := "a-very-long-title-that-keeps-going-and-going-and-going-past-the-eighty-character-cap-for-slugs"
	got := SlugFromTitle(long)
	if len(got) > 80 {
		t.Errorf("slug length = %d, want <= 80", len(got))
	}
	// Should not end with a hyphen (slug shape rule).
	if got[len(got)-1] == '-' {
		t.Errorf("slug ends with hyphen: %q", got)
	}
}

func TestSlugFromTitle_LeadingNumberDedup(t *testing.T) {
	// Repeated hyphens get deduped; numbers in the title are preserved.
	if got, want := SlugFromTitle("1-2-3 plan"), "1-2-3-plan"; got != want {
		t.Errorf("SlugFromTitle(%q) = %q, want %q", "1-2-3 plan", got, want)
	}
}
