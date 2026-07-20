package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/config"
)

func TestFromEnvNilWhenBothEmpty(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRF", "")
	if got := config.FromEnv(); got != nil {
		t.Errorf("FromEnv() = %+v, want nil when both env vars are empty", got)
	}
}

func TestFromEnvSessionOnly(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess123")
	t.Setenv("LEETCODE_CSRF", "")
	got := config.FromEnv()
	if got == nil {
		t.Fatal("FromEnv() = nil, want credentials")
	}
	if got.Session != "sess123" {
		t.Errorf("Session = %q, want %q", got.Session, "sess123")
	}
	if got.CSRF != "" {
		t.Errorf("CSRF = %q, want empty", got.CSRF)
	}
}

func TestFromEnvBoth(t *testing.T) {
	t.Setenv("LEETCODE_SESSION", "sess456")
	t.Setenv("LEETCODE_CSRF", "csrf789")
	got := config.FromEnv()
	if got == nil {
		t.Fatal("FromEnv() = nil, want credentials")
	}
	if got.Session != "sess456" {
		t.Errorf("Session = %q, want %q", got.Session, "sess456")
	}
	if got.CSRF != "csrf789" {
		t.Errorf("CSRF = %q, want %q", got.CSRF, "csrf789")
	}
}

func TestLoadNilWhenAbsent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got != nil {
		t.Errorf("Load() = %+v, want nil when no credentials saved", got)
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	want := &config.Credentials{Session: "my-session", CSRF: "my-csrf"}

	if err := config.Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got == nil {
		t.Fatal("Load() = nil after Save()")
	}
	if got.Session != want.Session {
		t.Errorf("Session = %q, want %q", got.Session, want.Session)
	}
	if got.CSRF != want.CSRF {
		t.Errorf("CSRF = %q, want %q", got.CSRF, want.CSRF)
	}
}

func TestSaveAndLoadWithoutCSRF(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	want := &config.Credentials{Session: "session-only"}

	if err := config.Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got == nil {
		t.Fatal("Load() = nil after Save()")
	}
	if got.CSRF != "" {
		t.Errorf("CSRF = %q, want empty for credentials saved without CSRF", got.CSRF)
	}
}

func TestGetPrefersEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("LEETCODE_SESSION", "env-session")
	t.Setenv("LEETCODE_CSRF", "")

	// Save different credentials to disk.
	_ = config.Save(&config.Credentials{Session: "disk-session"})

	got, err := config.Get()
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() = nil")
	}
	if got.Session != "env-session" {
		t.Errorf("Get() returned Session = %q, want env-session (env should take priority)", got.Session)
	}
}

func TestLoadMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	credDir := filepath.Join(dir, "pattern-drill")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(credDir, "credentials.json"), []byte("not valid json"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error on malformed JSON, got nil")
	}
}

func TestLoadErrorWhenNoHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() expected error when neither XDG_CONFIG_HOME nor HOME is set")
	}
}

func TestSaveErrorWhenNoHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	if err := config.Save(&config.Credentials{Session: "sess"}); err == nil {
		t.Fatal("Save() expected error when neither XDG_CONFIG_HOME nor HOME is set")
	}
}

func TestGetFallsBackToDisk(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LEETCODE_SESSION", "")
	t.Setenv("LEETCODE_CSRF", "")

	_ = config.Save(&config.Credentials{Session: "disk-session", CSRF: "disk-csrf"})

	got, err := config.Get()
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() = nil, want disk credentials")
	}
	if got.Session != "disk-session" {
		t.Errorf("Session = %q, want disk-session", got.Session)
	}
}
