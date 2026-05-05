package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/conallob/coding-interview-pop-quiz/internal/leetcode"
)

const ttl = 24 * time.Hour

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "pattern-drill"), nil
}

func ensureCacheDir() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// problemsWrapper wraps problems with a fetch timestamp for TTL checking.
type problemsWrapper struct {
	FetchedAt time.Time         `json:"fetchedAt"`
	Problems  []leetcode.Problem `json:"problems"`
}

// LoadProblems loads problems from cache. Returns nil, nil if absent or TTL expired (24h).
func LoadProblems() ([]leetcode.Problem, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "problems.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var wrapper problemsWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	// Check TTL
	if time.Since(wrapper.FetchedAt) > ttl {
		return nil, nil
	}

	return wrapper.Problems, nil
}

// SaveProblems saves problems to cache with the current timestamp.
func SaveProblems(problems []leetcode.Problem) error {
	dir, err := ensureCacheDir()
	if err != nil {
		return err
	}

	wrapper := problemsWrapper{
		FetchedAt: time.Now(),
		Problems:  problems,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "problems.json"), data, 0600)
}

// contentWrapper wraps content cache.
type contentWrapper struct {
	Content map[string]string `json:"content"`
}

// LoadContent loads the content cache (slug→HTML). Never expires. Returns empty map if absent.
func LoadContent() (map[string]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return make(map[string]string), err
	}

	path := filepath.Join(dir, "content.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string), nil
		}
		return make(map[string]string), err
	}

	var wrapper contentWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return make(map[string]string), err
	}

	if wrapper.Content == nil {
		return make(map[string]string), nil
	}
	return wrapper.Content, nil
}

// SaveContent saves the content cache to disk.
func SaveContent(contents map[string]string) error {
	dir, err := ensureCacheDir()
	if err != nil {
		return err
	}

	wrapper := contentWrapper{Content: contents}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "content.json"), data, 0600)
}

// Clear removes problems.json and content.json from the cache directory.
func Clear() error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	var errs []error
	for _, name := range []string{"problems.json", "content.json"} {
		path := filepath.Join(dir, name)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
