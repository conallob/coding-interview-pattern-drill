package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/leetcode"
	"github.com/conallob/coding-interview-pattern-drill/patterns"
	"github.com/conallob/coding-interview-pattern-drill/quiz"
)

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
