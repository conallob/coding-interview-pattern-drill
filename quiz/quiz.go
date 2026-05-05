package quiz

import (
	"math/rand"
	"strings"

	"github.com/conallob/coding-interview-pop-quiz/leetcode"
	"github.com/conallob/coding-interview-pop-quiz/patterns"
)

// Option represents one answer choice in a quiz question.
type Option struct {
	Label string // "A".."D"
	Name  string
	Slug  string
}

// Question represents a single quiz question.
type Question struct {
	Problem   leetcode.Problem
	Primary   *patterns.Pattern
	Secondary []*patterns.Pattern
	Options   [4]Option
	Answer    int // correct index 0-3
}

// Session represents an active quiz session.
type Session struct {
	Questions []Question
	Index     int
	Score     int
}

// Current returns the current question, or nil if the session is done.
func (s *Session) Current() *Question {
	if s.Done() {
		return nil
	}
	return &s.Questions[s.Index]
}

// Submit records the user's answer choice and advances the index.
// Returns true if the choice was correct.
func (s *Session) Submit(choice int) bool {
	q := s.Current()
	if q == nil {
		return false
	}
	correct := choice == q.Answer
	if correct {
		s.Score++
	}
	s.Index++
	return correct
}

// Done returns true if all questions have been answered.
func (s *Session) Done() bool {
	return s.Index >= len(s.Questions)
}

// labels for answer options
var optionLabels = [4]string{"A", "B", "C", "D"}

// buildQuestion builds a quiz question for a given problem.
// Returns nil if no primary pattern can be determined.
func buildQuestion(p leetcode.Problem) *Question {
	// Find primary pattern: first tag whose slug is in patterns.BySlug
	var primary *patterns.Pattern
	var secondary []*patterns.Pattern

	for _, tag := range p.TopicTags {
		if pat, ok := patterns.BySlug[tag.Slug]; ok {
			if primary == nil {
				primary = pat
			} else {
				secondary = append(secondary, pat)
			}
		}
	}

	if primary == nil {
		return nil
	}

	// Collect distractors from primary's confusion map
	picked := map[string]bool{primary.Slug: true}
	var distractorPatterns []*patterns.Pattern

	for _, slug := range primary.Distractors {
		if pat, ok := patterns.BySlug[slug]; ok {
			if !picked[slug] {
				distractorPatterns = append(distractorPatterns, pat)
				picked[slug] = true
			}
		}
	}

	// Fill remaining spots with random patterns not already picked
	if len(distractorPatterns) < 3 {
		// Build shuffled list of all patterns
		shuffled := make([]*patterns.Pattern, len(patterns.All))
		for i := range patterns.All {
			shuffled[i] = &patterns.All[i]
		}
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		for _, pat := range shuffled {
			if len(distractorPatterns) >= 3 {
				break
			}
			if !picked[pat.Slug] {
				distractorPatterns = append(distractorPatterns, pat)
				picked[pat.Slug] = true
			}
		}
	}

	// Build 4 options: primary + 3 distractors
	allOptions := []*patterns.Pattern{primary}
	allOptions = append(allOptions, distractorPatterns[:3]...)

	// Shuffle options
	rand.Shuffle(len(allOptions), func(i, j int) {
		allOptions[i], allOptions[j] = allOptions[j], allOptions[i]
	})

	var options [4]Option
	answerIdx := 0
	for i, pat := range allOptions {
		options[i] = Option{
			Label: optionLabels[i],
			Name:  pat.Name,
			Slug:  pat.Slug,
		}
		if pat.Slug == primary.Slug {
			answerIdx = i
		}
	}

	return &Question{
		Problem:   p,
		Primary:   primary,
		Secondary: secondary,
		Options:   options,
		Answer:    answerIdx,
	}
}

// FilterProblems filters problems by difficulty and/or tag slug.
// Problems without a primary pattern are excluded.
func FilterProblems(problems []leetcode.Problem, difficulty, tagSlug string) []leetcode.Problem {
	var result []leetcode.Problem

	for _, p := range problems {
		// Must have a primary pattern
		hasPrimary := false
		for _, tag := range p.TopicTags {
			if patterns.IsPatternSlug(tag.Slug) {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			continue
		}

		// Filter by difficulty (case-insensitive)
		if difficulty != "" && !strings.EqualFold(p.Difficulty, difficulty) {
			continue
		}

		// Filter by tagSlug (any tag, not just primary)
		if tagSlug != "" {
			found := false
			for _, tag := range p.TopicTags {
				if tag.Slug == tagSlug {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		result = append(result, p)
	}

	return result
}

// NewSession creates a new quiz session with the given problems and count.
func NewSession(problems []leetcode.Problem, count int) *Session {
	// Shuffle problems (Go 1.20+ auto-seeds the global rand source)
	shuffled := make([]leetcode.Problem, len(problems))
	copy(shuffled, problems)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	var questions []Question
	for _, p := range shuffled {
		if len(questions) >= count {
			break
		}
		q := buildQuestion(p)
		if q != nil {
			questions = append(questions, *q)
		}
	}

	return &Session{
		Questions: questions,
	}
}
