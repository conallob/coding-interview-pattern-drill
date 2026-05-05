package quiz_test

import (
	"fmt"
	"testing"

	"github.com/conallob/coding-interview-pop-quiz/internal/leetcode"
	"github.com/conallob/coding-interview-pop-quiz/internal/patterns"
	"github.com/conallob/coding-interview-pop-quiz/internal/quiz"
)

// makeProblem is a test helper that builds a Problem with the given tag slugs.
func makeProblem(slug, difficulty string, tagSlugs ...string) leetcode.Problem {
	tags := make([]leetcode.Tag, len(tagSlugs))
	for i, s := range tagSlugs {
		tags[i] = leetcode.Tag{Name: s, Slug: s}
	}
	return leetcode.Problem{
		QuestionID: slug,
		Title:      slug,
		TitleSlug:  slug,
		Difficulty: difficulty,
		TopicTags:  tags,
	}
}

func slugsOf(problems []leetcode.Problem) []string {
	out := make([]string, len(problems))
	for i, p := range problems {
		out[i] = p.TitleSlug
	}
	return out
}

// ── FilterProblems ──────────────────────────────────────────────────────────

func TestFilterExcludesNoPatternProblems(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("no-pattern", "Easy", "array", "string"), // non-pattern tags only
		makeProblem("has-pattern", "Easy", "two-pointers"),
	}
	got := quiz.FilterProblems(problems, "", "")
	if len(got) != 1 || got[0].TitleSlug != "has-pattern" {
		t.Errorf("got %v, want [has-pattern]", slugsOf(got))
	}
}

func TestFilterByDifficultyEasy(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
		makeProblem("b", "Medium", "sliding-window"),
		makeProblem("c", "Hard", "dynamic-programming"),
	}
	got := quiz.FilterProblems(problems, "easy", "")
	if len(got) != 1 || got[0].TitleSlug != "a" {
		t.Errorf("got %v, want [a]", slugsOf(got))
	}
}

func TestFilterByDifficultyCaseInsensitive(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Medium", "binary-search"),
		makeProblem("b", "Easy", "binary-search"),
	}
	for _, d := range []string{"Medium", "medium", "MEDIUM"} {
		got := quiz.FilterProblems(problems, d, "")
		if len(got) != 1 || got[0].TitleSlug != "a" {
			t.Errorf("difficulty=%q: got %v, want [a]", d, slugsOf(got))
		}
	}
}

func TestFilterByTagMatchesPrimary(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
		makeProblem("b", "Easy", "sliding-window"),
	}
	got := quiz.FilterProblems(problems, "", "two-pointers")
	if len(got) != 1 || got[0].TitleSlug != "a" {
		t.Errorf("got %v, want [a]", slugsOf(got))
	}
}

func TestFilterByTagMatchesSecondary(t *testing.T) {
	// Problem "a" has two-pointers as primary and sliding-window as secondary.
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers", "sliding-window"),
		makeProblem("b", "Easy", "dynamic-programming"),
	}
	got := quiz.FilterProblems(problems, "", "sliding-window")
	if len(got) != 1 || got[0].TitleSlug != "a" {
		t.Errorf("got %v, want [a]", slugsOf(got))
	}
}

func TestFilterByDifficultyAndTag(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("easy-match", "Easy", "binary-search"),
		makeProblem("medium-match", "Medium", "binary-search"),
		makeProblem("easy-no-tag", "Easy", "greedy"),
	}
	got := quiz.FilterProblems(problems, "easy", "binary-search")
	if len(got) != 1 || got[0].TitleSlug != "easy-match" {
		t.Errorf("got %v, want [easy-match]", slugsOf(got))
	}
}

func TestFilterNoMatch(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
	}
	got := quiz.FilterProblems(problems, "hard", "")
	if len(got) != 0 {
		t.Errorf("got %v, want empty slice", slugsOf(got))
	}
}

func TestFilterEmpty(t *testing.T) {
	got := quiz.FilterProblems(nil, "", "")
	if len(got) != 0 {
		t.Errorf("got %v, want empty slice for nil input", got)
	}
}

// ── NewSession ──────────────────────────────────────────────────────────────

func TestNewSessionCount(t *testing.T) {
	problems := make([]leetcode.Problem, 20)
	for i := range problems {
		problems[i] = makeProblem(fmt.Sprintf("p%d", i), "Easy", "two-pointers")
	}
	s := quiz.NewSession(problems, 7)
	if len(s.Questions) != 7 {
		t.Errorf("NewSession(20 problems, count=7): got %d questions, want 7", len(s.Questions))
	}
}

func TestNewSessionCappedAtAvailable(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
		makeProblem("b", "Easy", "sliding-window"),
	}
	s := quiz.NewSession(problems, 100)
	if len(s.Questions) != 2 {
		t.Errorf("NewSession(2 problems, count=100): got %d questions, want 2 (capped)", len(s.Questions))
	}
}

func TestNewSessionSkipsNoPatternProblems(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("no-pattern", "Easy", "array"),
		makeProblem("has-pattern", "Easy", "dynamic-programming"),
	}
	s := quiz.NewSession(problems, 10)
	if len(s.Questions) != 1 {
		t.Errorf("got %d questions, want 1 (only the pattern problem)", len(s.Questions))
	}
	if s.Questions[0].Primary.Slug != "dynamic-programming" {
		t.Errorf("primary pattern = %q, want dynamic-programming", s.Questions[0].Primary.Slug)
	}
}

// ── Session behaviour ───────────────────────────────────────────────────────

func TestSessionCurrentNilWhenEmpty(t *testing.T) {
	s := quiz.NewSession(nil, 5)
	if s.Current() != nil {
		t.Error("Current() should be nil for empty session")
	}
}

func TestSessionDoneImmediatelyWhenEmpty(t *testing.T) {
	s := quiz.NewSession(nil, 5)
	if !s.Done() {
		t.Error("Done() should be true for empty session")
	}
}

func TestSessionSubmitCorrect(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
	}, 1)

	q := s.Current()
	if q == nil {
		t.Fatal("Current() = nil, expected a question")
	}
	correct := s.Submit(q.Answer)
	if !correct {
		t.Error("Submit(correct answer) = false, want true")
	}
	if s.Score != 1 {
		t.Errorf("Score = %d, want 1 after correct answer", s.Score)
	}
}

func TestSessionSubmitWrong(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
	}, 1)

	q := s.Current()
	wrongChoice := (q.Answer + 1) % 4
	correct := s.Submit(wrongChoice)
	if correct {
		t.Error("Submit(wrong answer) = true, want false")
	}
	if s.Score != 0 {
		t.Errorf("Score = %d, want 0 after wrong answer", s.Score)
	}
}

func TestSessionAdvancesAfterSubmit(t *testing.T) {
	problems := []leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
		makeProblem("b", "Easy", "sliding-window"),
	}
	s := quiz.NewSession(problems, 2)

	firstSlug := s.Current().Problem.TitleSlug
	s.Submit(0) // answer doesn't matter for advancement

	secondQ := s.Current()
	if secondQ == nil {
		t.Fatal("Current() = nil after first submit, expected second question")
	}
	if secondQ.Problem.TitleSlug == firstSlug {
		t.Error("Current() returned same problem after Submit(); expected advancement")
	}
}

func TestSessionDoneAfterAllQuestions(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "two-pointers"),
	}, 1)

	if s.Done() {
		t.Error("Done() = true before any submissions")
	}
	s.Submit(0)
	if !s.Done() {
		t.Error("Done() = false after all questions answered")
	}
	if s.Current() != nil {
		t.Error("Current() should be nil when Done()")
	}
}

func TestSessionSubmitOnDoneReturnsFalse(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "greedy"),
	}, 1)
	s.Submit(0) // answer the only question
	// Submitting again on a done session should not panic and return false.
	if s.Submit(0) {
		t.Error("Submit() on done session returned true, want false")
	}
}

// ── Question structure ──────────────────────────────────────────────────────

func TestQuestionOptionLabels(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "hash-table"),
	}, 1)
	q := s.Current()

	want := map[string]bool{"A": true, "B": true, "C": true, "D": true}
	for _, opt := range q.Options {
		delete(want, opt.Label)
	}
	if len(want) != 0 {
		t.Errorf("missing labels: %v", want)
	}
}

func TestQuestionOptionSlugsAreValid(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "dynamic-programming"),
	}, 1)
	q := s.Current()

	for _, opt := range q.Options {
		if !patterns.IsPatternSlug(opt.Slug) {
			t.Errorf("option %q has invalid pattern slug %q", opt.Label, opt.Slug)
		}
	}
}

func TestQuestionOptionSlugsAreUnique(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "backtracking"),
	}, 1)
	q := s.Current()

	seen := map[string]bool{}
	for _, opt := range q.Options {
		if seen[opt.Slug] {
			t.Errorf("duplicate option slug %q", opt.Slug)
		}
		seen[opt.Slug] = true
	}
}

func TestQuestionAnswerIndexPointsToPrimary(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "sliding-window"),
	}, 1)
	q := s.Current()

	if q.Options[q.Answer].Slug != q.Primary.Slug {
		t.Errorf("Options[Answer].Slug = %q, want primary slug %q",
			q.Options[q.Answer].Slug, q.Primary.Slug)
	}
}

func TestQuestionPrimaryMatchesFirstPatternTag(t *testing.T) {
	// The first pattern tag should be nominated as primary.
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "binary-search", "two-pointers"),
	}, 1)
	q := s.Current()

	if q.Primary.Slug != "binary-search" {
		t.Errorf("Primary.Slug = %q, want binary-search (first pattern tag)", q.Primary.Slug)
	}
}

func TestQuestionSecondaryPatterns(t *testing.T) {
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Medium", "depth-first-search", "backtracking", "dynamic-programming"),
	}, 1)
	q := s.Current()

	if len(q.Secondary) != 2 {
		t.Errorf("Secondary has %d patterns, want 2", len(q.Secondary))
	}
}

func TestQuestionSecondaryExcludesNonPatternTags(t *testing.T) {
	// "array" is not a pattern tag; only "greedy" and "sliding-window" should appear.
	s := quiz.NewSession([]leetcode.Problem{
		makeProblem("a", "Easy", "greedy", "array", "sliding-window"),
	}, 1)
	q := s.Current()

	if q.Primary.Slug != "greedy" {
		t.Errorf("Primary.Slug = %q, want greedy", q.Primary.Slug)
	}
	if len(q.Secondary) != 1 || q.Secondary[0].Slug != "sliding-window" {
		t.Errorf("Secondary = %v, want [sliding-window]", q.Secondary)
	}
}
