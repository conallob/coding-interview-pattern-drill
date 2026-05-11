package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/config"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
	"github.com/conallob/coding-interview-pattern-drill/patterns"
	"github.com/conallob/coding-interview-pattern-drill/quiz"
)

// ANSI colour constants
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// splitCSV splits a comma-separated string into trimmed, lowercased tokens,
// dropping any empty entries. Returns nil for an empty input string.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(strings.ToLower(part)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// Run is the entrypoint for the CLI subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("pattern-drill", flag.ExitOnError)
	difficulty := fs.String("difficulty", "", "Difficulties to include, comma-separated: easy,medium,hard")
	tag        := fs.String("tag", "", "Pattern slugs to include, comma-separated (see --list-tags)")
	count := fs.Int("count", 10, "Number of questions per session")
	listTags := fs.Bool("list-tags", false, "List all 18 patterns and exit")
	refreshCache := fs.Bool("refresh-cache", false, "Ignore cached problems and re-fetch")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing flags:", err)
		os.Exit(1)
	}

	if *listTags {
		printPatternTable()
		return
	}

	// Get credentials
	creds, err := config.Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, colorRed+"Error loading credentials:"+colorReset, err)
		os.Exit(1)
	}
	if creds == nil {
		fmt.Fprintln(os.Stderr, colorRed+"No credentials found."+colorReset)
		fmt.Fprintln(os.Stderr, "Set LEETCODE_SESSION env var, or run with `serve` subcommand to configure via browser.")
		fmt.Fprintln(os.Stderr, "  export LEETCODE_SESSION=<your-session-cookie>")
		os.Exit(1)
	}

	// Load problems
	var problems []leetcode.Problem
	if !*refreshCache {
		problems, err = cache.LoadProblems()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cache read error:", err)
		}
	}
	if problems == nil {
		fmt.Fprintln(os.Stderr, colorCyan+"Fetching problems from LeetCode..."+colorReset)
		client := leetcode.New(creds)
		problems, err = client.FetchAllProblems()
		if err != nil {
			fmt.Fprintln(os.Stderr, colorRed+"Failed to fetch problems:"+colorReset, err)
			os.Exit(1)
		}
		if err := cache.SaveProblems(problems); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not save cache:", err)
		}
	}

	// Filter & build session
	filtered := quiz.FilterProblems(problems, splitCSV(*difficulty), splitCSV(*tag))
	if len(filtered) == 0 {
		fmt.Fprintln(os.Stderr, colorRed+"No problems match the given filters."+colorReset)
		os.Exit(1)
	}

	session := quiz.NewSession(filtered, *count)
	if len(session.Questions) == 0 {
		fmt.Fprintln(os.Stderr, colorRed+"Could not build any questions from the filtered problems."+colorReset)
		os.Exit(1)
	}

	// Load content cache
	contentCache, err := cache.LoadContent()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: could not load content cache:", err)
		contentCache = make(map[string]string)
	}

	client := leetcode.New(creds)
	reader := bufio.NewReader(os.Stdin)
	total := len(session.Questions)

	for !session.Done() {
		q := session.Current()
		slug := q.Problem.TitleSlug

		// Fetch content if not cached
		if _, ok := contentCache[slug]; !ok {
			fmt.Fprintf(os.Stderr, colorCyan+"Fetching description for %q..."+colorReset+"\n", q.Problem.Title)
			html, err := client.FetchContent(slug)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Warning: could not fetch content:", err)
			} else {
				contentCache[slug] = html
				if err := cache.SaveContent(contentCache); err != nil {
					fmt.Fprintln(os.Stderr, "Warning: could not save content cache:", err)
				}
			}
		}

		// Display question
		displayQuestion(session, q, total, contentCache)

		// Read answer
		fmt.Print("(a/b/c/d or q to quit) ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nRead error:", err)
			break
		}
		line = strings.TrimSpace(strings.ToLower(line))

		if line == "q" {
			fmt.Println("\nQuitting. Final score:", session.Score, "/", session.Index)
			return
		}

		choice := -1
		switch line {
		case "a":
			choice = 0
		case "b":
			choice = 1
		case "c":
			choice = 2
		case "d":
			choice = 3
		default:
			fmt.Println(colorYellow + "Please enter a, b, c, or d." + colorReset)
			continue
		}

		correct := session.Submit(choice)
		displayResult(correct, choice, q)
		fmt.Println()
	}

	// Final score
	total = len(session.Questions)
	pct := 0
	if total > 0 {
		pct = session.Score * 100 / total
	}
	fmt.Printf(colorBold+"\n═══ Quiz Complete! ═══\n"+colorReset)
	fmt.Printf("Score: %s%d / %d (%d%%)%s\n", colorBold, session.Score, total, pct, colorReset)
	switch {
	case pct >= 80:
		fmt.Println(colorGreen + "Excellent!" + colorReset)
	case pct >= 60:
		fmt.Println(colorYellow + "Good job!" + colorReset)
	default:
		fmt.Println(colorRed + "Keep practising!" + colorReset)
	}
}

func printPatternTable() {
	fmt.Printf("%-25s  %s\n", "Name", "Slug")
	fmt.Println(strings.Repeat("─", 60))
	for _, p := range patterns.All {
		fmt.Printf("%-25s  %s\n", p.Name, p.Slug)
	}
}

func difficultyColor(d string) string {
	switch strings.ToLower(d) {
	case "easy":
		return colorGreen
	case "medium":
		return colorYellow
	case "hard":
		return colorRed
	}
	return ""
}

func displayQuestion(s *quiz.Session, q *quiz.Question, total int, contentCache map[string]string) {
	fmt.Println(colorCyan + strings.Repeat("─", 45) + colorReset)
	dc := difficultyColor(q.Problem.Difficulty)
	fmt.Printf("Question %d of %d   %s%s%s\n\n", s.Index+1, total, dc, q.Problem.Difficulty, colorReset)
	fmt.Println(colorBold + q.Problem.Title + colorReset)
	fmt.Println()

	if html, ok := contentCache[q.Problem.TitleSlug]; ok && html != "" {
		text := leetcode.HTMLToText(html)
		fmt.Println(text)
	}

	fmt.Println()
	fmt.Println("What algorithmic pattern does this problem use?")
	fmt.Println()
	for _, opt := range q.Options {
		fmt.Printf("  [%s] %s\n", opt.Label, opt.Name)
	}
	fmt.Println()
}

func displayResult(correct bool, choice int, q *quiz.Question) {
	if correct {
		fmt.Printf(colorGreen+"✓ Correct! The answer is %s."+colorReset+"\n", q.Primary.Name)
	} else {
		fmt.Printf(colorRed+"✗ Incorrect. The answer was [%s] %s."+colorReset+"\n",
			q.Options[q.Answer].Label, q.Primary.Name)
		fmt.Printf("  You chose: [%s] %s\n", q.Options[choice].Label, q.Options[choice].Name)
	}

	if len(q.Secondary) > 0 {
		names := make([]string, len(q.Secondary))
		for i, p := range q.Secondary {
			names[i] = p.Name
		}
		fmt.Printf("  Also tagged: %s\n", strings.Join(names, ", "))
	}
}
