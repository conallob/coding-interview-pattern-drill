package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/conallob/coding-interview-pop-quiz/internal/cache"
	"github.com/conallob/coding-interview-pop-quiz/internal/config"
	"github.com/conallob/coding-interview-pop-quiz/internal/leetcode"
	"github.com/conallob/coding-interview-pop-quiz/internal/patterns"
	"github.com/conallob/coding-interview-pop-quiz/internal/quiz"
)

// resultPayload is the JSON payload for a quiz answer result.
type resultPayload struct {
	Correct           bool              `json:"correct"`
	ChoiceIndex       int               `json:"choiceIndex"`
	AnswerIndex       int               `json:"answerIndex"`
	PrimaryPattern    patternInfo       `json:"primaryPattern"`
	SecondaryPatterns []patternInfo     `json:"secondaryPatterns"`
}

type patternInfo struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// App holds server state.
type App struct {
	mu           sync.Mutex
	creds        *config.Credentials
	allProblems  []leetcode.Problem
	contentCache map[string]string
	session      *quiz.Session
	phase        string // "idle" | "question" | "result" | "done"
	lastResult   *resultPayload
}

// Run starts the web server.
func Run(args []string) {
	fs := flag.NewFlagSet("pattern-drill serve", flag.ExitOnError)
	port := fs.Int("port", 7777, "Port to listen on")
	noOpen := fs.Bool("no-open", false, "Don't open browser automatically")
	refreshCache := fs.Bool("refresh-cache", false, "Refresh problem cache on startup")
	fs.Parse(args)

	app := &App{
		phase:        "idle",
		contentCache: make(map[string]string),
	}

	// Load credentials at startup
	creds, err := config.Get()
	if err == nil {
		app.creds = creds
	}

	// Load content cache at startup
	cc, err := cache.LoadContent()
	if err == nil {
		app.contentCache = cc
	}

	// If refresh-cache requested, clear and re-fetch
	if *refreshCache && app.creds != nil {
		cache.Clear()
		client := leetcode.New(app.creds)
		problems, err := client.FetchAllProblems()
		if err == nil {
			app.allProblems = problems
			cache.SaveProblems(problems)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, _ := staticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	mux.HandleFunc("/api/settings", app.handleSettings)
	mux.HandleFunc("/api/tags", app.handleTags)
	mux.HandleFunc("/api/quiz/start", app.handleQuizStart)
	mux.HandleFunc("/api/quiz/state", app.handleQuizState)
	mux.HandleFunc("/api/quiz/answer", app.handleQuizAnswer)
	mux.HandleFunc("/api/quiz/next", app.handleQuizNext)
	mux.HandleFunc("/api/cache/refresh", app.handleCacheRefresh)

	// Find a free port
	basePort := *port
	var listener net.Listener
	for i := 0; i < 10; i++ {
		addr := fmt.Sprintf(":%d", basePort+i)
		l, err := net.Listen("tcp", addr)
		if err != nil {
			if isAddrInUse(err) {
				continue
			}
			fmt.Println("Error starting server:", err)
			return
		}
		listener = l
		basePort = basePort + i
		break
	}
	if listener == nil {
		fmt.Println("Could not find a free port after 10 attempts")
		return
	}

	url := fmt.Sprintf("http://localhost:%d", basePort)
	fmt.Println("Pattern Drill server running at", url)

	if !*noOpen {
		openBrowser(url)
	}

	http.Serve(listener, mux)
}

func isAddrInUse(err error) bool {
	return err != nil && strings.Contains(err.Error(), "address already in use")
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /api/settings  → {"set": bool}
// POST /api/settings → body: {"session":"...","csrf":"..."} → save, return {"ok":true}
func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		set := a.creds != nil && a.creds.Session != ""
		a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]bool{"set": set})

	case http.MethodPost:
		var body struct {
			Session string `json:"session"`
			CSRF    string `json:"csrf"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		creds := &config.Credentials{
			Session: body.Session,
			CSRF:    body.CSRF,
		}
		if err := config.Save(creds); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save credentials")
			return
		}
		a.mu.Lock()
		a.creds = creds
		// Reload content cache
		cc, _ := cache.LoadContent()
		if cc != nil {
			a.contentCache = cc
		}
		a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// GET /api/tags → [{"name":"...","slug":"..."}, ...]
func (a *App) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	type tagItem struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	items := make([]tagItem, len(patterns.All))
	for i, p := range patterns.All {
		items[i] = tagItem{Name: p.Name, Slug: p.Slug}
	}
	writeJSON(w, http.StatusOK, items)
}

// POST /api/quiz/start
func (a *App) handleQuizStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	a.mu.Lock()
	creds := a.creds
	a.mu.Unlock()

	if creds == nil || creds.Session == "" {
		writeError(w, http.StatusUnauthorized, "credentials not set")
		return
	}

	var body struct {
		Difficulty string `json:"difficulty"`
		Tag        string `json:"tag"`
		Count      int    `json:"count"`
		Refresh    bool   `json:"refresh"`
	}
	body.Count = 10
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Count <= 0 {
		body.Count = 10
	}

	problems, err := a.loadProblems(body.Refresh, creds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load problems: "+err.Error())
		return
	}

	filtered := quiz.FilterProblems(problems, body.Difficulty, body.Tag)
	session := quiz.NewSession(filtered, body.Count)

	a.mu.Lock()
	a.session = session
	a.phase = "question"
	a.lastResult = nil
	a.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"total": len(session.Questions),
	})
}

// loadProblems tries cache first (unless refresh=true), then fetches from API.
func (a *App) loadProblems(refresh bool, creds *config.Credentials) ([]leetcode.Problem, error) {
	a.mu.Lock()
	if len(a.allProblems) > 0 && !refresh {
		problems := a.allProblems
		a.mu.Unlock()
		return problems, nil
	}
	a.mu.Unlock()

	if !refresh {
		cached, err := cache.LoadProblems()
		if err == nil && cached != nil {
			a.mu.Lock()
			a.allProblems = cached
			a.mu.Unlock()
			return cached, nil
		}
	} else {
		cache.Clear()
	}

	client := leetcode.New(creds)
	problems, err := client.FetchAllProblems()
	if err != nil {
		return nil, err
	}
	cache.SaveProblems(problems)

	a.mu.Lock()
	a.allProblems = problems
	a.mu.Unlock()

	return problems, nil
}

// GET /api/quiz/state
func (a *App) handleQuizState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	a.mu.Lock()
	phase := a.phase
	session := a.session
	lastResult := a.lastResult
	creds := a.creds
	a.mu.Unlock()

	type optionJSON struct {
		Label string `json:"label"`
		Name  string `json:"name"`
		Slug  string `json:"slug"`
	}
	type questionJSON struct {
		Title      string       `json:"title"`
		Difficulty string       `json:"difficulty"`
		Content    string       `json:"content"`
		Options    []optionJSON `json:"options"`
	}

	resp := map[string]interface{}{
		"phase":    phase,
		"index":    0,
		"total":    0,
		"score":    0,
		"question": nil,
		"result":   nil,
	}

	if session != nil {
		resp["index"] = session.Index
		resp["total"] = len(session.Questions)
		resp["score"] = session.Score
	}

	if (phase == "question" || phase == "result") && session != nil {
		var q *quiz.Question
		if phase == "question" {
			q = session.Current()
		} else if phase == "result" && session.Index > 0 {
			// Show the question that was just answered
			prev := session.Questions[session.Index-1]
			q = &prev
		}

		if q != nil {
			slug := q.Problem.TitleSlug

			// Fetch content if missing (for question phase)
			if phase == "question" {
				a.mu.Lock()
				_, hasContent := a.contentCache[slug]
				a.mu.Unlock()

				if !hasContent && creds != nil {
					client := leetcode.New(creds)
					html, err := client.FetchContent(slug)
					if err == nil {
						a.mu.Lock()
						a.contentCache[slug] = html
						cc := a.contentCache
						a.mu.Unlock()
						cache.SaveContent(cc)
					}
				}
			}

			a.mu.Lock()
			content := a.contentCache[slug]
			a.mu.Unlock()

			opts := make([]optionJSON, 4)
			for i, opt := range q.Options {
				opts[i] = optionJSON{Label: opt.Label, Name: opt.Name, Slug: opt.Slug}
			}

			resp["question"] = questionJSON{
				Title:      q.Problem.Title,
				Difficulty: q.Problem.Difficulty,
				Content:    content,
				Options:    opts,
			}
		}
	}

	if phase == "result" && lastResult != nil {
		resp["result"] = lastResult
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /api/quiz/answer → {"choice": 0}
func (a *App) handleQuizAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Choice int `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.session == nil || a.phase != "question" {
		writeError(w, http.StatusBadRequest, "no active question")
		return
	}

	q := a.session.Current()
	if q == nil {
		writeError(w, http.StatusBadRequest, "no current question")
		return
	}

	answerIdx := q.Answer
	primary := patternInfo{Name: q.Primary.Name, Slug: q.Primary.Slug}
	secondary := make([]patternInfo, len(q.Secondary))
	for i, p := range q.Secondary {
		secondary[i] = patternInfo{Name: p.Name, Slug: p.Slug}
	}

	correct := a.session.Submit(body.Choice)

	result := &resultPayload{
		Correct:           correct,
		ChoiceIndex:       body.Choice,
		AnswerIndex:       answerIdx,
		PrimaryPattern:    primary,
		SecondaryPatterns: secondary,
	}
	a.lastResult = result

	if a.session.Done() {
		a.phase = "done"
	} else {
		a.phase = "result"
	}

	writeJSON(w, http.StatusOK, result)
}

// POST /api/quiz/next
func (a *App) handleQuizNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.phase == "result" {
		if a.session != nil && a.session.Done() {
			a.phase = "done"
		} else {
			a.phase = "question"
		}
		a.lastResult = nil
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/cache/refresh
func (a *App) handleCacheRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	a.mu.Lock()
	creds := a.creds
	a.mu.Unlock()

	if creds == nil || creds.Session == "" {
		writeError(w, http.StatusUnauthorized, "credentials not set")
		return
	}

	if err := cache.Clear(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear cache: "+err.Error())
		return
	}

	client := leetcode.New(creds)
	problems, err := client.FetchAllProblems()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch problems: "+err.Error())
		return
	}

	cache.SaveProblems(problems)

	a.mu.Lock()
	a.allProblems = problems
	a.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
