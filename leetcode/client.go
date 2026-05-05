package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/conallob/coding-interview-pop-quiz/internal/config"
)

// Tag represents a topic tag on a LeetCode problem.
type Tag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Problem represents a LeetCode problem.
type Problem struct {
	QuestionID string `json:"questionId"`
	Title      string `json:"title"`
	TitleSlug  string `json:"titleSlug"`
	Difficulty string `json:"difficulty"`
	TopicTags  []Tag  `json:"topicTags"`
	Content    string `json:"content,omitempty"`
}

// Client is a LeetCode API client.
type Client struct {
	creds      *config.Credentials
	httpClient *http.Client
}

// New creates a new LeetCode client.
func New(creds *config.Credentials) *Client {
	return &Client{
		creds:      creds,
		httpClient: &http.Client{},
	}
}

const graphqlEndpoint = "https://leetcode.com/graphql"

func (c *Client) buildRequest(body []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", graphqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://leetcode.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	cookie := "LEETCODE_SESSION=" + c.creds.Session
	if c.creds.CSRF != "" {
		cookie += "; csrftoken=" + c.creds.CSRF
		req.Header.Set("x-csrftoken", c.creds.CSRF)
	}
	req.Header.Set("Cookie", cookie)

	return req, nil
}

// FetchAllProblems fetches all problems from LeetCode using pagination.
func (c *Client) FetchAllProblems() ([]Problem, error) {
	const pageSize = 500
	var allProblems []Problem
	skip := 0

	for {
		query := map[string]interface{}{
			"query": `query problemList($limit: Int, $skip: Int) {
  problemsetQuestionList: questionList(
    categorySlug: ""
    limit: $limit
    skip: $skip
    filters: {}
  ) {
    total: totalNum
    questions: data {
      questionId title titleSlug difficulty
      topicTags { name slug }
    }
  }
}`,
			"variables": map[string]interface{}{
				"limit": pageSize,
				"skip":  skip,
			},
		}

		bodyBytes, err := json.Marshal(query)
		if err != nil {
			return nil, fmt.Errorf("marshal query: %w", err)
		}

		req, err := c.buildRequest(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
		}

		var result struct {
			Data struct {
				ProblemsetQuestionList struct {
					Total     int       `json:"total"`
					Questions []Problem `json:"questions"`
				} `json:"problemsetQuestionList"`
			} `json:"data"`
		}

		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		list := result.Data.ProblemsetQuestionList
		allProblems = append(allProblems, list.Questions...)
		skip += len(list.Questions)

		if skip >= list.Total || len(list.Questions) == 0 {
			break
		}
	}

	return allProblems, nil
}

// FetchContent fetches the raw HTML content for a problem.
func (c *Client) FetchContent(slug string) (string, error) {
	query := map[string]interface{}{
		"query": `query questionContent($titleSlug: String!) {
  question(titleSlug: $titleSlug) { content }
}`,
		"variables": map[string]interface{}{
			"titleSlug": slug,
		},
	}

	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return "", fmt.Errorf("marshal query: %w", err)
	}

	req, err := c.buildRequest(bodyBytes)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Question struct {
				Content string `json:"content"`
			} `json:"question"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return result.Data.Question.Content, nil
}

// Package-level compiled regexps for HTMLToText.
var (
	rePreBlock     = regexp.MustCompile(`(?is)<pre>(.*?)</pre>`)
	reInnerTags    = regexp.MustCompile(`<[^>]+>`)
	reBlockEnds    = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</h[1-6]>`)
	reListItem     = regexp.MustCompile(`(?i)<li>`)
	reAllTags      = regexp.MustCompile(`<[^>]+>`)
	reMultiNewline = regexp.MustCompile(`\n{3,}`)
)

// HTMLToText converts HTML problem descriptions to plain text suitable for terminal display.
func HTMLToText(h string) string {
	// Step 1: Extract <pre> blocks, strip inner tags, HTML-unescape, replace with placeholders.
	var preBlocks []string
	result := rePreBlock.ReplaceAllStringFunc(h, func(match string) string {
		// Extract inner content
		inner := rePreBlock.FindStringSubmatch(match)
		if len(inner) < 2 {
			return match
		}
		// Strip inner tags
		stripped := reInnerTags.ReplaceAllString(inner[1], "")
		// HTML-unescape
		unescaped := html.UnescapeString(stripped)
		idx := len(preBlocks)
		preBlocks = append(preBlocks, unescaped)
		return fmt.Sprintf("\x00PRE%d\x00", idx)
	})

	// Step 2: Replace block-end tags with newlines.
	result = reBlockEnds.ReplaceAllString(result, "\n")

	// Step 3: Replace <li> with bullet.
	result = reListItem.ReplaceAllString(result, "\n• ")

	// Step 4: Strip all remaining tags.
	result = reAllTags.ReplaceAllString(result, "")

	// Step 5: HTML-unescape the result.
	result = html.UnescapeString(result)

	// Step 6: Restore pre blocks wrapped in triple-backtick fences.
	for i, block := range preBlocks {
		placeholder := fmt.Sprintf("\x00PRE%d\x00", i)
		result = strings.ReplaceAll(result, placeholder, "\n```\n"+strings.TrimSpace(block)+"\n```\n")
	}

	// Step 7: Collapse 3+ consecutive newlines to 2, trim whitespace.
	result = reMultiNewline.ReplaceAllString(result, "\n\n")
	result = strings.TrimSpace(result)

	return result
}
