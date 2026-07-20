package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
	"github.com/conallob/coding-interview-pattern-drill/patterns"
	"github.com/conallob/coding-interview-pattern-drill/quiz"
)

// withStdin replaces os.Stdin with a pipe pre-loaded with input, restoring
// the original after the test.
func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	w.Close()

	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
}

func TestSplitCSVEmpty(t *testing.T) {
	if got := splitCSV(""); got != nil {
		t.Errorf("splitCSV(\"\") = %v, want nil", got)
	}
}

func TestSplitCSVSingle(t *testing.T) {
	got := splitCSV("Easy")
	want := []string{"easy"}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("splitCSV(\"Easy\") = %v, want %v", got, want)
	}
}

func TestSplitCSVMultipleTrimmedLowercased(t *testing.T) {
	got := splitCSV(" Easy , MEDIUM ,hard")
	want := []string{"easy", "medium", "hard"}
	if len(got) != len(want) {
		t.Fatalf("splitCSV() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("splitCSV()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitCSVDropsEmptyEntries(t *testing.T) {
	got := splitCSV("easy,,hard,")
	want := []string{"easy", "hard"}
	if len(got) != len(want) {
		t.Fatalf("splitCSV() = %v, want %v", got, want)
	}
}

func TestDifficultyColor(t *testing.T) {
	cases := map[string]string{
		"easy":   colorGreen,
		"Easy":   colorGreen,
		"medium": colorYellow,
		"MEDIUM": colorYellow,
		"hard":   colorRed,
		"Hard":   colorRed,
		"":       "",
		"weird":  "",
	}
	for input, want := range cases {
		if got := difficultyColor(input); got != want {
			t.Errorf("difficultyColor(%q) = %q, want %q", input, got, want)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintPatternTable(t *testing.T) {
	out := captureStdout(t, printPatternTable)
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Slug") {
		t.Errorf("expected header row, got %q", out)
	}
	if len(patterns.All) == 0 {
		t.Fatal("patterns.All is empty, cannot verify table contents")
	}
	first := patterns.All[0]
	if !strings.Contains(out, first.Name) || !strings.Contains(out, first.Slug) {
		t.Errorf("expected %q / %q in output, got %q", first.Name, first.Slug, out)
	}
}

func TestDisplayQuestion(t *testing.T) {
	problems := []leetcode.Problem{
		{TitleSlug: "two-sum", Title: "Two Sum", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "hash-table"}}},
	}
	session := quiz.NewSession(problems, 1)
	q := session.Current()

	contentCache := map[string]string{"two-sum": "<p>Given an array of integers</p>"}

	out := captureStdout(t, func() {
		displayQuestion(session, q, len(session.Questions), contentCache)
	})

	if !strings.Contains(out, "Two Sum") {
		t.Errorf("expected title in output, got %q", out)
	}
	if !strings.Contains(out, "Given an array of integers") {
		t.Errorf("expected description in output, got %q", out)
	}
	for _, opt := range q.Options {
		if !strings.Contains(out, opt.Name) {
			t.Errorf("expected option %q in output, got %q", opt.Name, out)
		}
	}
}

func TestDisplayResultCorrect(t *testing.T) {
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "A", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "greedy"}}},
	}
	session := quiz.NewSession(problems, 1)
	q := session.Current()

	out := captureStdout(t, func() {
		displayResult(true, q.Answer, q)
	})
	if !strings.Contains(out, "Correct") {
		t.Errorf("expected 'Correct' in output, got %q", out)
	}
	if !strings.Contains(out, q.Primary.Name) {
		t.Errorf("expected primary pattern name in output, got %q", out)
	}
}

func TestDisplayResultIncorrect(t *testing.T) {
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "A", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "greedy"}}},
	}
	session := quiz.NewSession(problems, 1)
	q := session.Current()
	wrongChoice := (q.Answer + 1) % 4

	out := captureStdout(t, func() {
		displayResult(false, wrongChoice, q)
	})
	if !strings.Contains(out, "Incorrect") {
		t.Errorf("expected 'Incorrect' in output, got %q", out)
	}
	if !strings.Contains(out, "You chose") {
		t.Errorf("expected chosen-answer line in output, got %q", out)
	}
}

func TestRunListTags(t *testing.T) {
	out := captureStdout(t, func() {
		Run([]string{"--list-tags"})
	})
	if !strings.Contains(out, "Slug") {
		t.Errorf("expected header row in --list-tags output, got %q", out)
	}
	if len(patterns.All) == 0 {
		t.Fatal("patterns.All is empty, cannot verify table contents")
	}
	if !strings.Contains(out, patterns.All[0].Slug) {
		t.Errorf("expected first pattern slug in output, got %q", out)
	}
}

func TestRunQuizQuitImmediately(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	problem := leetcode.Problem{
		QuestionID: "1", Title: "Two Sum", TitleSlug: "two-sum", Difficulty: "Easy",
		TopicTags: []leetcode.Tag{{Slug: "hash-table"}},
	}
	if err := cache.SaveProblems([]leetcode.Problem{problem}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}
	if err := cache.SaveContent(map[string]string{"two-sum": "<p>desc</p>"}); err != nil {
		t.Fatalf("SaveContent: %v", err)
	}

	withStdin(t, "q\n")

	out := captureStdout(t, func() {
		Run([]string{"--count", "1"})
	})

	if !strings.Contains(out, "Quitting") {
		t.Errorf("expected quit message in output, got %q", out)
	}
}

func TestRunQuizAnswerThenComplete(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	problem := leetcode.Problem{
		QuestionID: "1", Title: "Two Sum", TitleSlug: "two-sum", Difficulty: "Easy",
		TopicTags: []leetcode.Tag{{Slug: "hash-table"}},
	}
	if err := cache.SaveProblems([]leetcode.Problem{problem}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}
	if err := cache.SaveContent(map[string]string{"two-sum": "<p>desc</p>"}); err != nil {
		t.Fatalf("SaveContent: %v", err)
	}

	// A single-problem, single-question session ends the loop after one
	// answer, regardless of which letter is picked, so "a" is safe here
	// even though the correct answer letter is randomised per run.
	withStdin(t, "a\n")

	out := captureStdout(t, func() {
		Run([]string{"--count", "1"})
	})

	if !strings.Contains(out, "Quiz Complete") {
		t.Errorf("expected completion banner in output, got %q", out)
	}
}

func TestRunQuizInvalidThenQuit(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	problem := leetcode.Problem{
		QuestionID: "1", Title: "Two Sum", TitleSlug: "two-sum", Difficulty: "Easy",
		TopicTags: []leetcode.Tag{{Slug: "hash-table"}},
	}
	if err := cache.SaveProblems([]leetcode.Problem{problem}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}
	if err := cache.SaveContent(map[string]string{"two-sum": "<p>desc</p>"}); err != nil {
		t.Fatalf("SaveContent: %v", err)
	}

	withStdin(t, "z\nq\n")

	out := captureStdout(t, func() {
		Run([]string{"--count", "1"})
	})

	if !strings.Contains(out, "Please enter a, b, c, or d.") {
		t.Errorf("expected invalid-choice prompt in output, got %q", out)
	}
	if !strings.Contains(out, "Quitting") {
		t.Errorf("expected quit message in output, got %q", out)
	}
}
