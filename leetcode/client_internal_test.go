package leetcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/config"
)

// withMockEndpoint points GraphQLEndpoint at a test server for the duration
// of the test, restoring the real endpoint afterwards.
func withMockEndpoint(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	original := GraphQLEndpoint
	GraphQLEndpoint = srv.URL
	t.Cleanup(func() { GraphQLEndpoint = original })
}

func TestBuildRequestSetsCookieAndCSRF(t *testing.T) {
	c := New(&config.Credentials{Session: "sess123", CSRF: "csrf456"})
	req, err := c.buildRequest([]byte(`{}`))
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}

	cookie := req.Header.Get("Cookie")
	if !strings.Contains(cookie, "LEETCODE_SESSION=sess123") {
		t.Errorf("Cookie header missing session: %q", cookie)
	}
	if !strings.Contains(cookie, "csrftoken=csrf456") {
		t.Errorf("Cookie header missing csrftoken: %q", cookie)
	}
	if got := req.Header.Get("x-csrftoken"); got != "csrf456" {
		t.Errorf("x-csrftoken header = %q, want csrf456", got)
	}
}

func TestBuildRequestOmitsCSRFWhenEmpty(t *testing.T) {
	c := New(&config.Credentials{Session: "sess123"})
	req, err := c.buildRequest([]byte(`{}`))
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}
	if got := req.Header.Get("x-csrftoken"); got != "" {
		t.Errorf("x-csrftoken header = %q, want empty", got)
	}
	if strings.Contains(req.Header.Get("Cookie"), "csrftoken=") {
		t.Errorf("Cookie header should not contain csrftoken: %q", req.Header.Get("Cookie"))
	}
}

func TestFetchAllProblemsSinglePage(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"problemsetQuestionList": map[string]interface{}{
					"total": 2,
					"questions": []Problem{
						{QuestionID: "1", Title: "Two Sum", TitleSlug: "two-sum", Difficulty: "Easy"},
						{QuestionID: "2", Title: "Add Two Numbers", TitleSlug: "add-two-numbers", Difficulty: "Medium"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := New(&config.Credentials{Session: "sess"})
	problems, err := c.FetchAllProblems()
	if err != nil {
		t.Fatalf("FetchAllProblems error: %v", err)
	}
	if len(problems) != 2 {
		t.Fatalf("got %d problems, want 2", len(problems))
	}
	if problems[0].TitleSlug != "two-sum" {
		t.Errorf("problems[0].TitleSlug = %q, want two-sum", problems[0].TitleSlug)
	}
}

func TestFetchAllProblemsPaginates(t *testing.T) {
	pages := [][]Problem{
		{{QuestionID: "1", TitleSlug: "a"}, {QuestionID: "2", TitleSlug: "b"}},
		{{QuestionID: "3", TitleSlug: "c"}},
	}
	call := 0

	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		var page []Problem
		if call < len(pages) {
			page = pages[call]
		}
		call++
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"problemsetQuestionList": map[string]interface{}{
					"total":     3,
					"questions": page,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := New(&config.Credentials{Session: "sess"})
	problems, err := c.FetchAllProblems()
	if err != nil {
		t.Fatalf("FetchAllProblems error: %v", err)
	}
	if len(problems) != 3 {
		t.Fatalf("got %d problems, want 3 across pages", len(problems))
	}
	if call != 2 {
		t.Errorf("expected 2 requests for pagination, got %d", call)
	}
}

func TestFetchAllProblemsStopsOnEmptyPage(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"problemsetQuestionList": map[string]interface{}{
					"total":     100, // lies about the total; empty page should still stop the loop
					"questions": []Problem{},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := New(&config.Credentials{Session: "sess"})
	problems, err := c.FetchAllProblems()
	if err != nil {
		t.Fatalf("FetchAllProblems error: %v", err)
	}
	if len(problems) != 0 {
		t.Errorf("got %d problems, want 0", len(problems))
	}
}

func TestFetchAllProblemsErrorStatus(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("bad session"))
	})

	c := New(&config.Credentials{Session: "expired"})
	_, err := c.FetchAllProblems()
	if err == nil {
		t.Fatal("expected error on non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestFetchAllProblemsMalformedJSON(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})

	c := New(&config.Credentials{Session: "sess"})
	_, err := c.FetchAllProblems()
	if err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}

func TestFetchContentSuccess(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"question": map[string]interface{}{
					"content": "<p>Given an array...</p>",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	c := New(&config.Credentials{Session: "sess"})
	content, err := c.FetchContent("two-sum")
	if err != nil {
		t.Fatalf("FetchContent error: %v", err)
	}
	if content != "<p>Given an array...</p>" {
		t.Errorf("FetchContent() = %q, unexpected content", content)
	}
}

func TestFetchContentErrorStatus(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := New(&config.Credentials{Session: "sess"})
	_, err := c.FetchContent("two-sum")
	if err == nil {
		t.Fatal("expected error on 500 status, got nil")
	}
}

func TestFetchContentMalformedJSON(t *testing.T) {
	withMockEndpoint(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{not valid"))
	})

	c := New(&config.Credentials{Session: "sess"})
	_, err := c.FetchContent("two-sum")
	if err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}
