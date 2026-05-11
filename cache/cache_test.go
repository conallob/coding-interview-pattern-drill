package cache_test

import (
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
)

func setTempCache(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
}

var sampleProblems = []leetcode.Problem{
	{
		QuestionID: "1",
		Title:      "Two Sum",
		TitleSlug:  "two-sum",
		Difficulty: "Easy",
		TopicTags:  []leetcode.Tag{{Name: "Hash Table", Slug: "hash-table"}},
	},
	{
		QuestionID: "2",
		Title:      "Add Two Numbers",
		TitleSlug:  "add-two-numbers",
		Difficulty: "Medium",
		TopicTags:  []leetcode.Tag{{Name: "Linked List", Slug: "linked-list"}},
	},
}

// ── Problems cache ──────────────────────────────────────────────────────────

func TestLoadProblemsAbsent(t *testing.T) {
	setTempCache(t)
	got, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems() error: %v", err)
	}
	if got != nil {
		t.Errorf("LoadProblems() = %v, want nil when cache absent", got)
	}
}

func TestSaveAndLoadProblems(t *testing.T) {
	setTempCache(t)

	if err := cache.SaveProblems(sampleProblems); err != nil {
		t.Fatalf("SaveProblems() error: %v", err)
	}

	got, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems() error: %v", err)
	}
	if len(got) != len(sampleProblems) {
		t.Fatalf("LoadProblems() returned %d problems, want %d", len(got), len(sampleProblems))
	}
	if got[0].TitleSlug != "two-sum" {
		t.Errorf("first problem TitleSlug = %q, want %q", got[0].TitleSlug, "two-sum")
	}
	if got[1].Difficulty != "Medium" {
		t.Errorf("second problem Difficulty = %q, want Medium", got[1].Difficulty)
	}
}

func TestSaveAndLoadProblemsPreservesTopicTags(t *testing.T) {
	setTempCache(t)

	problems := []leetcode.Problem{{
		QuestionID: "3",
		TitleSlug:  "longest-substring",
		TopicTags: []leetcode.Tag{
			{Name: "Sliding Window", Slug: "sliding-window"},
			{Name: "Hash Table", Slug: "hash-table"},
		},
	}}

	_ = cache.SaveProblems(problems)
	got, _ := cache.LoadProblems()

	if len(got[0].TopicTags) != 2 {
		t.Errorf("TopicTags: got %d, want 2", len(got[0].TopicTags))
	}
	if got[0].TopicTags[0].Slug != "sliding-window" {
		t.Errorf("first tag slug = %q, want sliding-window", got[0].TopicTags[0].Slug)
	}
}

func TestSaveEmptyProblems(t *testing.T) {
	setTempCache(t)
	if err := cache.SaveProblems([]leetcode.Problem{}); err != nil {
		t.Fatalf("SaveProblems(empty) error: %v", err)
	}
	got, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("LoadProblems() = %d problems, want 0", len(got))
	}
}

// ── Content cache ───────────────────────────────────────────────────────────

func TestLoadContentAbsent(t *testing.T) {
	setTempCache(t)
	got, err := cache.LoadContent()
	if err != nil {
		t.Fatalf("LoadContent() error: %v", err)
	}
	if got == nil {
		t.Error("LoadContent() = nil, want empty map when cache absent")
	}
	if len(got) != 0 {
		t.Errorf("LoadContent() = %v, want empty map", got)
	}
}

func TestSaveAndLoadContent(t *testing.T) {
	setTempCache(t)
	contents := map[string]string{
		"two-sum":         "<p>Given an array...</p>",
		"add-two-numbers": "<p>You are given...</p>",
	}

	if err := cache.SaveContent(contents); err != nil {
		t.Fatalf("SaveContent() error: %v", err)
	}

	got, err := cache.LoadContent()
	if err != nil {
		t.Fatalf("LoadContent() error: %v", err)
	}
	for slug, want := range contents {
		if got[slug] != want {
			t.Errorf("content[%q] = %q, want %q", slug, got[slug], want)
		}
	}
}

func TestContentCacheCanBeUpdated(t *testing.T) {
	setTempCache(t)

	_ = cache.SaveContent(map[string]string{"two-sum": "v1"})
	_ = cache.SaveContent(map[string]string{"two-sum": "v2", "new-problem": "content"})

	got, _ := cache.LoadContent()
	if got["two-sum"] != "v2" {
		t.Errorf("two-sum = %q, want v2 after update", got["two-sum"])
	}
	if got["new-problem"] != "content" {
		t.Errorf("new-problem = %q, want content", got["new-problem"])
	}
}

// ── Clear ───────────────────────────────────────────────────────────────────

func TestClearRemovesProblems(t *testing.T) {
	setTempCache(t)
	_ = cache.SaveProblems(sampleProblems)

	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	got, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems() after Clear() error: %v", err)
	}
	if got != nil {
		t.Errorf("LoadProblems() = %v after Clear(), want nil", got)
	}
}

func TestClearRemovesContent(t *testing.T) {
	setTempCache(t)
	_ = cache.SaveContent(map[string]string{"two-sum": "content"})

	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	got, err := cache.LoadContent()
	if err != nil {
		t.Fatalf("LoadContent() after Clear() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("LoadContent() = %v after Clear(), want empty map", got)
	}
}

func TestClearIdempotent(t *testing.T) {
	setTempCache(t)
	// Clear on an already-empty cache should not error.
	if err := cache.Clear(); err != nil {
		t.Errorf("Clear() on empty cache error: %v", err)
	}
}
