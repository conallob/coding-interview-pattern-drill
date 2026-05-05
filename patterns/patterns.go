package patterns

// Pattern represents an algorithmic pattern used in coding problems.
type Pattern struct {
	Name        string
	Slug        string
	Distractors []string // confusion-map: slugs of commonly-mistaken alternatives
}

// All contains all 18 supported patterns.
var All = []Pattern{
	{
		Name:        "Two Pointers",
		Slug:        "two-pointers",
		Distractors: []string{"sliding-window", "binary-search", "hash-table"},
	},
	{
		Name:        "Sliding Window",
		Slug:        "sliding-window",
		Distractors: []string{"two-pointers", "dynamic-programming", "prefix-sum"},
	},
	{
		Name:        "Binary Search",
		Slug:        "binary-search",
		Distractors: []string{"two-pointers", "divide-and-conquer", "greedy"},
	},
	{
		Name:        "Dynamic Programming",
		Slug:        "dynamic-programming",
		Distractors: []string{"greedy", "backtracking", "divide-and-conquer"},
	},
	{
		Name:        "Greedy",
		Slug:        "greedy",
		Distractors: []string{"dynamic-programming", "two-pointers", "binary-search"},
	},
	{
		Name:        "DFS",
		Slug:        "depth-first-search",
		Distractors: []string{"breadth-first-search", "backtracking", "topological-sort"},
	},
	{
		Name:        "BFS",
		Slug:        "breadth-first-search",
		Distractors: []string{"depth-first-search", "greedy", "topological-sort"},
	},
	{
		Name:        "Backtracking",
		Slug:        "backtracking",
		Distractors: []string{"depth-first-search", "dynamic-programming", "greedy"},
	},
	{
		Name:        "Heap / Priority Queue",
		Slug:        "heap-priority-queue",
		Distractors: []string{"greedy", "breadth-first-search", "sliding-window"},
	},
	{
		Name:        "Hash Map",
		Slug:        "hash-table",
		Distractors: []string{"two-pointers", "prefix-sum", "sliding-window"},
	},
	{
		Name:        "Stack",
		Slug:        "stack",
		Distractors: []string{"monotonic-stack", "depth-first-search", "greedy"},
	},
	{
		Name:        "Monotonic Stack",
		Slug:        "monotonic-stack",
		Distractors: []string{"stack", "sliding-window", "greedy"},
	},
	{
		Name:        "Topological Sort",
		Slug:        "topological-sort",
		Distractors: []string{"breadth-first-search", "depth-first-search", "union-find"},
	},
	{
		Name:        "Union Find",
		Slug:        "union-find",
		Distractors: []string{"depth-first-search", "breadth-first-search", "topological-sort"},
	},
	{
		Name:        "Trie",
		Slug:        "trie",
		Distractors: []string{"hash-table", "depth-first-search", "backtracking"},
	},
	{
		Name:        "Divide and Conquer",
		Slug:        "divide-and-conquer",
		Distractors: []string{"dynamic-programming", "binary-search", "backtracking"},
	},
	{
		Name:        "Prefix Sum",
		Slug:        "prefix-sum",
		Distractors: []string{"sliding-window", "hash-table", "dynamic-programming"},
	},
	{
		Name:        "Bit Manipulation",
		Slug:        "bit-manipulation",
		Distractors: []string{"greedy", "dynamic-programming", "hash-table"},
	},
}

// BySlug maps pattern slugs to their Pattern structs.
var BySlug map[string]*Pattern

func init() {
	BySlug = make(map[string]*Pattern, len(All))
	for i := range All {
		BySlug[All[i].Slug] = &All[i]
	}
}

// IsPatternSlug returns true if the given slug corresponds to a known pattern.
func IsPatternSlug(slug string) bool {
	_, ok := BySlug[slug]
	return ok
}
