package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
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
	if err := w.Close(); err != nil {
		t.Fatalf("close stdin pipe: %v", err)
	}

	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
}

// exitSentinel is panicked by the mocked osExit so Run() unwinds immediately
// at the exit point, the same way a real os.Exit() would halt execution,
// without killing the test binary.
type exitSentinel struct{ code int }

// runAndCaptureExit calls Run(args) with osExit replaced by a function that
// records the requested code and unwinds via panic/recover instead of
// terminating the process. Returns the captured code, or -1 if Run returned
// normally without ever calling osExit.
func runAndCaptureExit(t *testing.T, args []string) int {
	t.Helper()
	code := -1

	old := osExit
	osExit = func(c int) {
		code = c
		panic(exitSentinel{c})
	}
	t.Cleanup(func() { osExit = old })

	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(exitSentinel); !ok {
					panic(r) // not ours — a real bug, let it surface
				}
			}
		}()
		Run(args)
	}()

	return code
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

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
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

// ── Run() exit paths ─────────────────────────────────────────────────────────

func TestRunExitsOnBadFlag(t *testing.T) {
	code := runAndCaptureExit(t, []string{"--count", "not-a-number"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunExitsWhenCredentialsErrorLoading(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "") // forces config.Load()'s configDir() to error

	code := runAndCaptureExit(t, nil)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunExitsWhenNoCredentials(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // valid, empty: config.Get() returns nil, nil

	code := runAndCaptureExit(t, nil)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunExitsWhenFetchFails(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("LEETCODE_CSRF", "")
	t.Setenv("XDG_CACHE_HOME", t.TempDir()) // empty cache forces a fetch

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	old := leetcode.GraphQLEndpoint
	leetcode.GraphQLEndpoint = srv.URL
	t.Cleanup(func() { leetcode.GraphQLEndpoint = old })

	code := runAndCaptureExit(t, nil)
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRunExitsWhenNoProblemsMatchFilter(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	problem := leetcode.Problem{
		TitleSlug: "a", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "hash-table"}},
	}
	if err := cache.SaveProblems([]leetcode.Problem{problem}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}

	code := runAndCaptureExit(t, []string{"--difficulty", "hard"})
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}
