package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/config"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
	"github.com/conallob/coding-interview-pattern-drill/quiz"
)

// withMockGraphQL points leetcode.GraphQLEndpoint at a test server for the
// duration of the test, restoring the real endpoint afterwards.
func withMockGraphQL(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	original := leetcode.GraphQLEndpoint
	leetcode.GraphQLEndpoint = srv.URL
	t.Cleanup(func() { leetcode.GraphQLEndpoint = original })
}

func problemsGraphQLHandler(problems []leetcode.Problem) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"problemsetQuestionList": map[string]interface{}{
					"total":     len(problems),
					"questions": problems,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func newTestApp() *App {
	return &App{
		phase:        "idle",
		contentCache: make(map[string]string),
	}
}

func doJSON(t *testing.T, handler http.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(method, "/", reader)
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestIsAddrInUse(t *testing.T) {
	if isAddrInUse(nil) {
		t.Error("isAddrInUse(nil) = true, want false")
	}
	if !isAddrInUse(&addrError{"listen tcp :7777: bind: address already in use"}) {
		t.Error("expected true for address-in-use error")
	}
	if isAddrInUse(&addrError{"some other error"}) {
		t.Error("expected false for unrelated error")
	}
}

type addrError struct{ msg string }

func (e *addrError) Error() string { return e.msg }

func TestHandleSettingsGetUnset(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleSettings, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]bool
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["set"] {
		t.Error("expected set=false with no credentials")
	}
}

func TestHandleSettingsPostSaves(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	app := newTestApp()
	rec := doJSON(t, app.handleSettings, http.MethodPost, `{"session":"abc","csrf":"xyz"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if app.creds == nil || app.creds.Session != "abc" {
		t.Errorf("app.creds not updated: %+v", app.creds)
	}
}

func TestHandleSettingsPostInvalidJSON(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleSettings, http.MethodPost, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleSettingsMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleSettings, http.MethodDelete, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleTagsReturnsAllPatterns(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleTags, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var items []map[string]string
	json.Unmarshal(rec.Body.Bytes(), &items)
	if len(items) == 0 {
		t.Error("expected non-empty tag list")
	}
}

func TestHandleTagsMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleTags, http.MethodPost, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleQuizStartNoCredentials(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizStart, http.MethodPost, `{}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestHandleQuizStartMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizStart, http.MethodGet, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleQuizStartUsesCachedProblems(t *testing.T) {
	app := newTestApp()
	app.creds = &config.Credentials{Session: "sess"}
	app.allProblems = []leetcode.Problem{
		{TitleSlug: "a", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
	}

	rec := doJSON(t, app.handleQuizStart, http.MethodPost, `{"count":1}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if app.phase != "question" {
		t.Errorf("phase = %q, want question", app.phase)
	}
	if app.session == nil || len(app.session.Questions) != 1 {
		t.Errorf("expected 1 question, got session=%+v", app.session)
	}
}

func TestHandleQuizStartInvalidJSON(t *testing.T) {
	app := newTestApp()
	app.creds = &config.Credentials{Session: "sess"}
	rec := doJSON(t, app.handleQuizStart, http.MethodPost, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleQuizStateIdle(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizState, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["phase"] != "idle" {
		t.Errorf("phase = %v, want idle", body["phase"])
	}
}

func TestHandleQuizStateQuestionPhase(t *testing.T) {
	app := newTestApp()
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "Problem A", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
	}
	app.session = quiz.NewSession(problems, 1)
	app.phase = "question"
	app.contentCache["a"] = "<p>desc</p>"

	rec := doJSON(t, app.handleQuizState, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["question"] == nil {
		t.Error("expected question payload, got nil")
	}
}

func TestHandleQuizStateMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizState, http.MethodPost, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleQuizAnswerNoActiveQuestion(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizAnswer, http.MethodPost, `{"choice":0}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleQuizAnswerCorrectFlow(t *testing.T) {
	app := newTestApp()
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "Problem A", Difficulty: "Easy", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
	}
	app.session = quiz.NewSession(problems, 1)
	app.phase = "question"

	answer := app.session.Current().Answer
	rec := doJSON(t, app.handleQuizAnswer, http.MethodPost, `{"choice":`+strconv.Itoa(answer)+`}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if app.phase != "done" {
		t.Errorf("phase = %q, want done (single-question session)", app.phase)
	}
}

func TestHandleQuizAnswerInvalidJSON(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizAnswer, http.MethodPost, `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleQuizNextAdvancesFromResult(t *testing.T) {
	app := newTestApp()
	problems := []leetcode.Problem{
		{TitleSlug: "a", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
		{TitleSlug: "b", TopicTags: []leetcode.Tag{{Slug: "sliding-window"}}},
	}
	app.session = quiz.NewSession(problems, 2)
	app.session.Submit(0)
	app.phase = "result"

	rec := doJSON(t, app.handleQuizNext, http.MethodPost, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if app.phase != "question" {
		t.Errorf("phase = %q, want question", app.phase)
	}
}

func TestHandleQuizNextMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleQuizNext, http.MethodGet, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleCacheRefreshNoCredentials(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleCacheRefresh, http.MethodPost, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestHandleCacheRefreshMethodNotAllowed(t *testing.T) {
	app := newTestApp()
	rec := doJSON(t, app.handleCacheRefresh, http.MethodGet, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHandleVersion(t *testing.T) {
	rec := doJSON(t, handleVersion, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	json.Unmarshal(rec.Body.Bytes(), &body)
	if _, ok := body["version"]; !ok {
		t.Error("expected version field in response")
	}
}

func TestHandleVersionMethodNotAllowed(t *testing.T) {
	rec := doJSON(t, handleVersion, http.MethodPost, "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

// ── loadProblems ─────────────────────────────────────────────────────────────

func TestLoadProblemsReturnsInMemoryCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	app := newTestApp()
	app.allProblems = []leetcode.Problem{{TitleSlug: "in-memory"}}

	got, err := app.loadProblems(false, &config.Credentials{Session: "sess"})
	if err != nil {
		t.Fatalf("loadProblems error: %v", err)
	}
	if len(got) != 1 || got[0].TitleSlug != "in-memory" {
		t.Errorf("loadProblems() = %v, want the in-memory slice untouched", got)
	}
}

func TestLoadProblemsReadsFromDiskCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if err := cache.SaveProblems([]leetcode.Problem{{TitleSlug: "on-disk"}}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}

	app := newTestApp()
	got, err := app.loadProblems(false, &config.Credentials{Session: "sess"})
	if err != nil {
		t.Fatalf("loadProblems error: %v", err)
	}
	if len(got) != 1 || got[0].TitleSlug != "on-disk" {
		t.Errorf("loadProblems() = %v, want the disk-cached slice", got)
	}
	if len(app.allProblems) != 1 {
		t.Error("expected loadProblems to populate app.allProblems from disk cache")
	}
}

func TestLoadProblemsFetchesWhenCacheEmpty(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	withMockGraphQL(t, problemsGraphQLHandler([]leetcode.Problem{{TitleSlug: "fetched"}}))

	app := newTestApp()
	got, err := app.loadProblems(false, &config.Credentials{Session: "sess"})
	if err != nil {
		t.Fatalf("loadProblems error: %v", err)
	}
	if len(got) != 1 || got[0].TitleSlug != "fetched" {
		t.Errorf("loadProblems() = %v, want freshly-fetched slice", got)
	}

	cached, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems (disk) error: %v", err)
	}
	if len(cached) != 1 {
		t.Error("expected loadProblems to persist the fetched result to disk cache")
	}
}

func TestLoadProblemsRefreshIgnoresCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if err := cache.SaveProblems([]leetcode.Problem{{TitleSlug: "stale"}}); err != nil {
		t.Fatalf("SaveProblems: %v", err)
	}
	withMockGraphQL(t, problemsGraphQLHandler([]leetcode.Problem{{TitleSlug: "refreshed"}}))

	app := newTestApp()
	got, err := app.loadProblems(true, &config.Credentials{Session: "sess"})
	if err != nil {
		t.Fatalf("loadProblems error: %v", err)
	}
	if len(got) != 1 || got[0].TitleSlug != "refreshed" {
		t.Errorf("loadProblems(refresh=true) = %v, want the refreshed slice, not the stale cache", got)
	}
}

func TestLoadProblemsFetchError(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	withMockGraphQL(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	app := newTestApp()
	_, err := app.loadProblems(false, &config.Credentials{Session: "sess"})
	if err == nil {
		t.Fatal("loadProblems() expected error on fetch failure, got nil")
	}
}

// ── handleCacheRefresh ───────────────────────────────────────────────────────

func TestHandleCacheRefreshSuccess(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	withMockGraphQL(t, problemsGraphQLHandler([]leetcode.Problem{{TitleSlug: "refreshed"}}))

	app := newTestApp()
	app.creds = &config.Credentials{Session: "sess"}

	rec := doJSON(t, app.handleCacheRefresh, http.MethodPost, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if len(app.allProblems) != 1 || app.allProblems[0].TitleSlug != "refreshed" {
		t.Errorf("app.allProblems = %v, want refreshed slice", app.allProblems)
	}
}

func TestHandleCacheRefreshFetchError(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	withMockGraphQL(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	app := newTestApp()
	app.creds = &config.Credentials{Session: "sess"}

	rec := doJSON(t, app.handleCacheRefresh, http.MethodPost, "")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

// ── handleQuizState edge cases ───────────────────────────────────────────────

func TestHandleQuizStateResultPhaseShowsPreviousQuestion(t *testing.T) {
	app := newTestApp()
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "A", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
		{TitleSlug: "b", Title: "B", TopicTags: []leetcode.Tag{{Slug: "sliding-window"}}},
	}
	app.session = quiz.NewSession(problems, 2)
	app.session.Submit(0)
	app.phase = "result"

	rec := doJSON(t, app.handleQuizState, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["question"] == nil {
		t.Error("expected the just-answered question to be shown during result phase")
	}
}

func TestHandleQuizStateFetchesMissingContent(t *testing.T) {
	withMockGraphQL(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"question": map[string]interface{}{"content": "<p>fetched content</p>"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	app := newTestApp()
	app.creds = &config.Credentials{Session: "sess"}
	problems := []leetcode.Problem{
		{TitleSlug: "a", Title: "A", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
	}
	app.session = quiz.NewSession(problems, 1)
	app.phase = "question"

	rec := doJSON(t, app.handleQuizState, http.MethodGet, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if app.contentCache["a"] != "<p>fetched content</p>" {
		t.Errorf("contentCache[\"a\"] = %q, want fetched content", app.contentCache["a"])
	}
}

// ── handleQuizAnswer edge cases ──────────────────────────────────────────────

func TestHandleQuizAnswerNoCurrentQuestion(t *testing.T) {
	app := newTestApp()
	problems := []leetcode.Problem{
		{TitleSlug: "a", TopicTags: []leetcode.Tag{{Slug: "two-pointers"}}},
	}
	app.session = quiz.NewSession(problems, 1)
	// Force an inconsistent state that shouldn't occur in normal operation:
	// phase says "question" but the session is already exhausted, so
	// session.Current() returns nil.
	app.session.Index = len(app.session.Questions)
	app.phase = "question"

	rec := doJSON(t, app.handleQuizAnswer, http.MethodPost, `{"choice":0}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

