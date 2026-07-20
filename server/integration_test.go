package server

import (
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/conallob/coding-interview-pattern-drill/cache"
	"github.com/conallob/coding-interview-pattern-drill/leetcode"
)

// withCapturedListener runs fn (which is expected to eventually call
// Run(args) and never return normally) in a goroutine, intercepting the
// listener Run() binds via onListen. It returns the listener and a channel
// that's closed once Run() itself returns.
func withCapturedListener(t *testing.T, args []string) (net.Listener, <-chan struct{}) {
	t.Helper()

	listenerCh := make(chan net.Listener, 1)
	old := onListen
	onListen = func(l net.Listener) { listenerCh <- l }
	t.Cleanup(func() { onListen = old })

	done := make(chan struct{})
	go func() {
		defer close(done)
		Run(args)
	}()

	select {
	case listener := <-listenerCh:
		return listener, done
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Run() to start listening")
		return nil, nil // unreachable, t.Fatal halts the goroutine
	}
}

// stopAndWait closes the listener from outside — the only way to unblock
// Run()'s call to http.Serve — and waits for Run() to actually return,
// exercising the "Server stopped: ..." error branch along the way.
func stopAndWait(t *testing.T, listener net.Listener, done <-chan struct{}) {
	t.Helper()
	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close: %v", err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Run() to return after listener close")
	}
}

func TestRunIntegrationServesRealHTTP(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// --port 0 asks the OS for a free ephemeral port; Run() finds it on its
	// first attempt through the port-search loop.
	listener, done := withCapturedListener(t, []string{"--port", "0", "--no-open"})

	// Run() prints "http://localhost:<requested-port>", which is wrong when
	// --port 0 is used since the OS — not Run() — picks the real port; dial
	// the listener's actual bound address instead.
	addr := listener.Addr().String()

	resp, err := http.Get("http://" + addr + "/version")
	if err != nil {
		t.Fatalf("GET /version: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /version status = %d, want 200", resp.StatusCode)
	}

	resp2, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", resp2.StatusCode)
	}

	stopAndWait(t, listener, done)
}

func TestRunIntegrationRetriesOnPortConflict(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Occupy a real port first so Run()'s first attempt hits "address
	// already in use" and has to retry on port+1.
	blocker, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	defer func() { _ = blocker.Close() }()

	_, portStr, err := net.SplitHostPort(blocker.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi: %v", err)
	}

	listener, done := withCapturedListener(t, []string{"--port", strconv.Itoa(port), "--no-open"})

	if listener.Addr().String() == blocker.Addr().String() {
		t.Error("expected Run() to bind a different port after the conflict, got the same one")
	}

	stopAndWait(t, listener, done)
}

func TestRunIntegrationRefreshCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LEETCODE_SESSION", "sess")
	t.Setenv("LEETCODE_CSRF", "")

	withMockGraphQL(t, problemsGraphQLHandler([]leetcode.Problem{{TitleSlug: "refreshed"}}))

	// The refresh-cache fetch happens synchronously before Run() calls
	// onListen, so by the time we receive the listener the cache write has
	// already completed.
	listener, done := withCapturedListener(t, []string{"--port", "0", "--no-open", "--refresh-cache"})

	cached, err := cache.LoadProblems()
	if err != nil {
		t.Fatalf("LoadProblems: %v", err)
	}
	if len(cached) != 1 || cached[0].TitleSlug != "refreshed" {
		t.Errorf("cache after --refresh-cache = %v, want the refreshed problem", cached)
	}

	stopAndWait(t, listener, done)
}
