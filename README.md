# coding-interview-pattern-drill

[![Test](https://github.com/conallob/coding-interview-pattern-drill/workflows/Test/badge.svg)](https://github.com/conallob/coding-interview-pattern-drill/actions)
[![codecov](https://codecov.io/gh/conallob/coding-interview-pattern-drill/branch/main/graph/badge.svg)](https://codecov.io/gh/conallob/coding-interview-pattern-drill)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-support-yellow.svg?style=flat&logo=buy-me-a-coffee)](https://www.buymeacoffee.com/conallob)

## Why this exists

LeetCode-style interviews are a poor proxy for real engineering ability. They favour people who've memorised solutions over people who can reason through novel problems, and they routinely ask you to reimplement things that exist perfectly well in the standard library. They're also ubiquitous and unavoidable, so here we are.

The biggest source of wasted time in LeetCode prep is misidentifying the pattern. You read a problem, reach for the wrong tool — BFS when you needed Dijkstra, DP when it's just Greedy — and burn 20 minutes going nowhere. The actual coding, once you know the approach, usually takes 10 minutes.

`coding-interview-pattern-drill` targets that gap. Given a real problem description, pick the right algorithmic pattern from four options before writing a single line of code. Wrong options are drawn from a hand-tuned confusion map — the patterns most commonly mistaken for the correct one — so the choices are meaningfully hard rather than obviously wrong.

The goal is to build fast, confident pattern recognition so that in an interview you spend your time on implementation, not on figuring out where to start.

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
go install github.com/conallob/coding-interview-pattern-drill/cmd/coding-interview-pattern-drill@latest
```

### Build from source

```sh
git clone https://github.com/conallob/coding-interview-pattern-drill.git
cd coding-interview-pattern-drill
go build -o coding-interview-pattern-drill ./cmd/coding-interview-pattern-drill
```

---

## Authentication

The tool fetches problems from the LeetCode API, which requires a valid session cookie.

### Quickest: `login` command

```sh
coding-interview-pattern-drill login
```

This opens two browser tabs: one to LeetCode's own login page (so you can sign in however your account works — password, Google, GitHub, or any other SSO/OIDC provider) and one to the tool's local settings form. Once you're logged in to LeetCode, copy the `LEETCODE_SESSION` cookie value from DevTools (`F12` → **Application → Cookies → https://leetcode.com**) and paste it into the settings form. Credentials are saved to disk (see [Where credentials are stored](#where-credentials-are-stored)) and reused on every future run.

### Manual

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

**Browser UI** (server mode): the settings form includes a **Log in to LeetCode →** button that opens LeetCode's login page in a new tab, then paste the cookie into the form — no terminal required. See [HTTP Server mode](#http-server-mode) below.

Credentials saved via the browser UI are persisted to disk (see [Where credentials are stored](#where-credentials-are-stored)) and survive restarts.

### Where credentials are stored

The tool uses Go's [`os.UserConfigDir()`](https://pkg.go.dev/os#UserConfigDir), which follows each OS's native convention rather than a single fixed path:

| Platform | Path |
| --- | --- |
| macOS | `~/Library/Application Support/pattern-drill/credentials.json` |
| Linux | `$XDG_CONFIG_HOME/pattern-drill/credentials.json` (falls back to `~/.config/pattern-drill/credentials.json` if `XDG_CONFIG_HOME` is unset) |
| Windows | `%AppData%\pattern-drill\credentials.json` |

Check the path for your platform above. The file is written with mode `0600` (owner read/write only).

---

## CLI mode

Run a quiz session directly in your terminal.

```sh
coding-interview-pattern-drill [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--difficulty <list>` | *(all)* | Comma-separated difficulties to include: `easy,medium,hard` |
| `--tag <list>` | *(all patterns)* | Comma-separated pattern slugs to include (see `--list-tags`) |
| `--count N` | `10` | Number of problems to quiz |
| `--list-tags` | — | Print all 18 pattern slugs and exit |
| `--refresh-cache` | — | Force a fresh fetch from LeetCode, ignoring the local cache |

### Examples

```sh
# Quick 10-question session across all difficulties and patterns
coding-interview-pattern-drill

# 20 medium problems only
coding-interview-pattern-drill --difficulty medium --count 20

# Easy and hard, but not medium
coding-interview-pattern-drill --difficulty easy,hard

# Drill two commonly-confused patterns together
coding-interview-pattern-drill --tag dynamic-programming,backtracking

# Medium/hard problems from graph-related patterns only
coding-interview-pattern-drill --difficulty medium,hard --tag depth-first-search,breadth-first-search

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

