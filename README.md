# coding-interview-pattern-drill

I loathe LeetCode-style interviews — both because of how artificial they are and how common it is to reinvent the wheel instead of using a standard library. But they're ubiquitous, so here's a tool to make practicing less painful.

`coding-interview-pattern-drill` trains pattern recognition: given a problem description, identify the correct algorithmic approach *before* writing any code. That's the skill that actually matters in a timed interview.

Problems are pulled from the LeetCode API and presented as multiple-choice questions. Wrong options come from a hand-tuned confusion map (the patterns most commonly mistaken for the right one), not random noise.

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew tap conallob/taps
brew install coding-interview-pattern-drill
```

### Go install

```sh
go install github.com/conallob/coding-interview-pop-quiz/cmd/coding-interview-pattern-drill@latest
```

### Build from source

```sh
git clone https://github.com/conallob/coding-interview-pop-quiz.git
cd coding-interview-pop-quiz
go build -o coding-interview-pattern-drill ./cmd/coding-interview-pattern-drill
```

---

## Authentication

The tool fetches problems from the LeetCode API, which requires a valid session cookie.

1. Log in to [leetcode.com](https://leetcode.com) in your browser
2. Open DevTools (`F12` or `⌘⌥I`)
3. Go to **Application → Cookies → https://leetcode.com**
4. Copy the value of the `LEETCODE_SESSION` cookie

You can supply it in two ways:

**Environment variable** (CLI mode):
```sh
export LEETCODE_SESSION=<paste value here>
export LEETCODE_CSRF=<csrftoken value, if required>
```

**Browser UI** (server mode): paste it into the settings form — no terminal required. See [HTTP Server mode](#http-server-mode) below.

Credentials saved via the browser UI are persisted to `~/.config/pattern-drill/credentials.json` (mode `0600`) and survive restarts.

---

## CLI mode

Run a quiz session directly in your terminal.

```sh
coding-interview-pattern-drill [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--difficulty easy\|medium\|hard` | *(all)* | Filter problems by difficulty |
| `--tag <slug>` | *(all patterns)* | Drill a single pattern (see `--list-tags`) |
| `--count N` | `10` | Number of problems to quiz |
| `--list-tags` | — | Print all 18 pattern slugs and exit |
| `--refresh-cache` | — | Force a fresh fetch from LeetCode, ignoring the local cache |

### Examples

```sh
# Quick 10-question session across all difficulties
coding-interview-pattern-drill

# 20 medium problems only
coding-interview-pattern-drill --difficulty medium --count 20

# Drill dynamic programming specifically
coding-interview-pattern-drill --tag dynamic-programming

# See all available pattern slugs
coding-interview-pattern-drill --list-tags

# Force refresh the problem list
coding-interview-pattern-drill --refresh-cache
```

### Quiz flow

```
─────────────────────────────────────────────────────────────────
Question 3 of 10   Medium

Longest Substring Without Repeating Characters

Given a string s, find the length of the longest substring
without repeating characters.

  Input: s = "abcabcbb"
  Output: 3

What algorithmic pattern does this problem use?

  [A] Dynamic Programming
  [B] Sliding Window
  [C] Hash Map
  [D] Two Pointers

(a/b/c/d or q to quit)
```

After answering:

```
✓ Correct! The answer is Sliding Window.
  Also tagged: Hash Map
```

Type `q` at any prompt to quit early. A final score is shown at the end of each session.

---

## HTTP server mode

Starts a local web server and opens the quiz in your browser. Cookie management happens entirely in the browser — no environment variables needed.

```sh
coding-interview-pattern-drill serve [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--port N` | `7777` | Port to listen on |
| `--no-open` | — | Don't auto-open the browser |

### Examples

```sh
# Start on the default port and open the browser
coding-interview-pattern-drill serve

# Use a custom port without opening the browser
coding-interview-pattern-drill serve --port 8080 --no-open
```

Then open `http://localhost:7777` (or your chosen port) if it didn't open automatically.

### First-time setup in the browser

1. Click the **⚙ Settings** gear icon
2. Paste your `LEETCODE_SESSION` cookie value into the field
3. Click **Save** — credentials are stored locally and reused on restart

### Quiz configuration

From the main screen, select:
- **Difficulty**: All / Easy / Medium / Hard
- **Pattern**: All Patterns, or any of the 18 specific patterns
- **Questions**: how many to include in the session (default 10)

Then hit **Start Quiz**. Use the `A` `B` `C` `D` keys to answer, and `Enter` or `N` to advance after each result.

---

## Supported patterns

The tool recognises 18 algorithmic patterns. Use the slug with `--tag` in CLI mode.

| Pattern | Slug |
|---|---|
| Two Pointers | `two-pointers` |
| Sliding Window | `sliding-window` |
| Binary Search | `binary-search` |
| Dynamic Programming | `dynamic-programming` |
| Greedy | `greedy` |
| DFS | `depth-first-search` |
| BFS | `breadth-first-search` |
| Backtracking | `backtracking` |
| Heap / Priority Queue | `heap-priority-queue` |
| Hash Map | `hash-table` |
| Stack | `stack` |
| Monotonic Stack | `monotonic-stack` |
| Topological Sort | `topological-sort` |
| Union Find | `union-find` |
| Trie | `trie` |
| Divide and Conquer | `divide-and-conquer` |
| Prefix Sum | `prefix-sum` |
| Bit Manipulation | `bit-manipulation` |

---

## Caching

Problems are cached locally to avoid hammering the LeetCode API:

| Cache | Location | TTL |
|---|---|---|
| Problem list | `~/.cache/pattern-drill/problems.json` | 24 hours |
| Problem content | `~/.cache/pattern-drill/content.json` | No expiry |

Use `--refresh-cache` (CLI) or the **Refresh Cache** button (browser UI) to force a fresh fetch.

---

## Releasing

Tag a commit and push — GitHub Actions handles the rest:

```sh
git tag v1.2.3
git push origin v1.2.3
```

The release workflow builds binaries for Linux, macOS (universal), and Windows, publishes a GitHub release, and opens a PR against [conallob/homebrew-taps](https://github.com/conallob/homebrew-taps) to update the formula.

Requires a `HOMEBREW_TAP_GITHUB_TOKEN` secret (a PAT with write access to the tap repo) configured in the repository settings.
