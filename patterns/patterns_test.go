package patterns_test

import (
	"testing"

	"github.com/conallob/coding-interview-pop-quiz/patterns"
)

func TestAllCount(t *testing.T) {
	if len(patterns.All) != 18 {
		t.Errorf("expected 18 patterns, got %d", len(patterns.All))
	}
}

func TestAllHaveNonEmptyFields(t *testing.T) {
	for _, p := range patterns.All {
		if p.Name == "" {
			t.Errorf("pattern with slug %q has empty Name", p.Slug)
		}
		if p.Slug == "" {
			t.Errorf("pattern with name %q has empty Slug", p.Name)
		}
	}
}

func TestBySlugCoversAll(t *testing.T) {
	for _, p := range patterns.All {
		got, ok := patterns.BySlug[p.Slug]
		if !ok {
			t.Errorf("BySlug missing slug %q", p.Slug)
			continue
		}
		if got.Name != p.Name {
			t.Errorf("BySlug[%q].Name = %q, want %q", p.Slug, got.Name, p.Name)
		}
	}
}

func TestBySlugSize(t *testing.T) {
	if len(patterns.BySlug) != len(patterns.All) {
		t.Errorf("BySlug has %d entries, want %d", len(patterns.BySlug), len(patterns.All))
	}
}

func TestIsPatternSlugKnown(t *testing.T) {
	for _, p := range patterns.All {
		if !patterns.IsPatternSlug(p.Slug) {
			t.Errorf("IsPatternSlug(%q) = false, want true", p.Slug)
		}
	}
}

func TestIsPatternSlugUnknown(t *testing.T) {
	unknown := []string{"", "array", "string", "not-a-pattern", "two_pointers"}
	for _, s := range unknown {
		if patterns.IsPatternSlug(s) {
			t.Errorf("IsPatternSlug(%q) = true, want false", s)
		}
	}
}

func TestDistractorsCount(t *testing.T) {
	for _, p := range patterns.All {
		if len(p.Distractors) != 3 {
			t.Errorf("pattern %q has %d distractors, want exactly 3", p.Slug, len(p.Distractors))
		}
	}
}

func TestDistractorsAreValidSlugs(t *testing.T) {
	for _, p := range patterns.All {
		for _, d := range p.Distractors {
			if !patterns.IsPatternSlug(d) {
				t.Errorf("pattern %q distractor %q is not a valid pattern slug", p.Slug, d)
			}
		}
	}
}

func TestDistractorsNoSelf(t *testing.T) {
	for _, p := range patterns.All {
		for _, d := range p.Distractors {
			if d == p.Slug {
				t.Errorf("pattern %q lists itself as a distractor", p.Slug)
			}
		}
	}
}

func TestDistractorsUnique(t *testing.T) {
	for _, p := range patterns.All {
		seen := map[string]bool{}
		for _, d := range p.Distractors {
			if seen[d] {
				t.Errorf("pattern %q has duplicate distractor %q", p.Slug, d)
			}
			seen[d] = true
		}
	}
}

func TestSlugsGloballyUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range patterns.All {
		if seen[p.Slug] {
			t.Errorf("duplicate slug %q in All", p.Slug)
		}
		seen[p.Slug] = true
	}
}
