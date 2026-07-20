package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
)

func setTempCache(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	return filepath.Join(dir, "pattern-drill")
}

func writeRawCacheFile(t *testing.T, cacheDir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, name), []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
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

func TestLoadProblemsMalformedJSON(t *testing.T) {
	dir := setTempCache(t)
	writeRawCacheFile(t, dir, "problems.json", "not valid json")

	_, err := cache.LoadProblems()
	if err == nil {
		t.Fatal("LoadProblems() expected error on malformed JSON, got nil")
	}
}

func TestLoadProblemsExpiredTTL(t *testing.T) {
	dir := setTempCache(t)
	stale := `{"fetchedAt":"` + time.Now().Add(-48*time.Hour).Format(time.RFC3339) + `","problems":[{"questionId":"1"}]}`
	writeRawCacheFile(t, dir, "problems.json", stale)

	got, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems() error: %v", err)
	}
	if got != nil {
		t.Errorf("LoadProblems() = %v, want nil for TTL-expired cache", got)
	}
}

// ── cacheDir/ensureCacheDir error propagation ───────────────────────────────

func withNoHome(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")
}

func TestLoadProblemsErrorWhenNoHome(t *testing.T) {
	withNoHome(t)
	_, err := cache.LoadProblems()
	if err == nil {
		t.Fatal("LoadProblems() expected error when neither XDG_CACHE_HOME nor HOME is set")
	}
}

func TestSaveProblemsErrorWhenNoHome(t *testing.T) {
	withNoHome(t)
	if err := cache.SaveProblems(sampleProblems); err == nil {
		t.Fatal("SaveProblems() expected error when neither XDG_CACHE_HOME nor HOME is set")
	}
}

func TestLoadContentErrorWhenNoHome(t *testing.T) {
	withNoHome(t)
	got, err := cache.LoadContent()
	if err == nil {
		t.Fatal("LoadContent() expected error when neither XDG_CACHE_HOME nor HOME is set")
	}
	if got == nil || len(got) != 0 {
		t.Errorf("LoadContent() = %v, want empty (non-nil) map alongside the error", got)
	}
}

func TestSaveContentErrorWhenNoHome(t *testing.T) {
	withNoHome(t)
	if err := cache.SaveContent(map[string]string{"a": "b"}); err == nil {
		t.Fatal("SaveContent() expected error when neither XDG_CACHE_HOME nor HOME is set")
	}
}

func TestClearErrorWhenNoHome(t *testing.T) {
	withNoHome(t)
	if err := cache.Clear(); err == nil {
		t.Fatal("Clear() expected error when neither XDG_CACHE_HOME nor HOME is set")
	}
}

func TestLoadContentMalformedJSON(t *testing.T) {
	dir := setTempCache(t)
	writeRawCacheFile(t, dir, "content.json", "not valid json")

	got, err := cache.LoadContent()
	if err == nil {
		t.Fatal("LoadContent() expected error on malformed JSON, got nil")
	}
	if got == nil || len(got) != 0 {
		t.Errorf("LoadContent() = %v, want empty (non-nil) map alongside the error", got)
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
