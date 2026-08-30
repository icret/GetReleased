package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v69/github"
)

func newTestClient(t *testing.T, tokens []string, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient(tokens)
	u, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	for _, ts := range c.tokens {
		ts.client.BaseURL = u
	}
	return c
}

func rateHeaders(w http.ResponseWriter, remaining int, reset time.Time) {
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
	w.Header().Set("X-RateLimit-Limit", "5000")
}

func writeRepoJSON(w http.ResponseWriter, fullName string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"full_name":        fullName,
		"name":             strings.Split(fullName, "/")[1],
		"owner":            map[string]string{"login": strings.Split(fullName, "/")[0]},
		"stargazers_count": 100,
		"pushed_at":        "2025-01-01T00:00:00Z",
	})
}

func TestFetchRepository_ETagReturned(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"abc123"`)
		w.Header().Set("Last-Modified", "Wed, 01 Jan 2025 00:00:00 GMT")
		rateHeaders(w, 5000, time.Now().Add(time.Hour))
		w.WriteHeader(http.StatusOK)
		writeRepoJSON(w, "o/n")
	})
	c := newTestClient(t, []string{"token1"}, handler)

	result, err := c.FetchRepository(context.Background(), "o", "n", "", "")
	if err != nil {
		t.Fatalf("FetchRepository: %v", err)
	}
	if !result.Modified {
		t.Error("expected Modified=true")
	}
	if result.ETag != `"abc123"` {
		t.Errorf("ETag = %q, want %q", result.ETag, `"abc123"`)
	}
	if result.LastModified != "Wed, 01 Jan 2025 00:00:00 GMT" {
		t.Errorf("LastModified = %q", result.LastModified)
	}
	if result.Repo == nil {
		t.Error("expected Repo non-nil")
	}
}

func TestFetchRepository_304NotModified(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") != `"abc123"` {
			t.Errorf("If-None-Match = %q, want %q", r.Header.Get("If-None-Match"), `"abc123"`)
		}
		rateHeaders(w, 5000, time.Now().Add(time.Hour))
		w.WriteHeader(http.StatusNotModified)
	})
	c := newTestClient(t, []string{"token1"}, handler)

	result, err := c.FetchRepository(context.Background(), "o", "n", `"abc123"`, "")
	if err != nil {
		t.Fatalf("FetchRepository: %v", err)
	}
	if result.Modified {
		t.Error("expected Modified=false on 304")
	}
	if result.Repo != nil {
		t.Error("expected Repo nil on 304")
	}
}

func TestFetchRepository_403CooldownFallback(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.Contains(auth, "token2") {
			rateHeaders(w, 0, time.Now().Add(time.Hour))
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "rate limit exceeded"})
		} else {
			rateHeaders(w, 5000, time.Now().Add(time.Hour))
			w.WriteHeader(http.StatusOK)
			writeRepoJSON(w, "o/n")
		}
	})
	c := newTestClient(t, []string{"token1", "token2"}, handler)

	result, err := c.FetchRepository(context.Background(), "o", "n", "", "")
	if err != nil {
		t.Fatalf("FetchRepository: %v", err)
	}
	if !result.Modified {
		t.Error("expected Modified=true after fallback")
	}
	if c.tokens[1].remaining.Load() != 0 {
		t.Error("token2 should be in cooldown (remaining=0)")
	}
}

func TestFetchRepository_LowRemainingSkipsToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.Contains(auth, "token1") {
			t.Error("token1 should be skipped due to low remaining")
		}
		rateHeaders(w, 5000, time.Now().Add(time.Hour))
		w.WriteHeader(http.StatusOK)
		writeRepoJSON(w, "o/n")
	})
	c := newTestClient(t, []string{"token1", "token2"}, handler)
	c.tokens[0].remaining.Store(50)

	result, err := c.FetchRepository(context.Background(), "o", "n", "", "")
	if err != nil {
		t.Fatalf("FetchRepository: %v", err)
	}
	if !result.Modified {
		t.Error("expected Modified=true")
	}
}

func TestRateLimitRemaining_Sum(t *testing.T) {
	c := NewClient([]string{"token1", "token2"})
	c.tokens[0].remaining.Store(1000)
	c.tokens[1].remaining.Store(2000)

	total, err := c.RateLimitRemaining(context.Background())
	if err != nil {
		t.Fatalf("RateLimitRemaining: %v", err)
	}
	if total != 3000 {
		t.Errorf("total = %d, want 3000", total)
	}
}

func TestRateLimitRemaining_CooldownReset(t *testing.T) {
	c := NewClient([]string{"token1"})
	c.tokens[0].remaining.Store(0)
	c.tokens[0].resetAt.Store(time.Now().Add(-time.Minute).Unix())

	total, err := c.RateLimitRemaining(context.Background())
	if err != nil {
		t.Fatalf("RateLimitRemaining: %v", err)
	}
	if total != 5000 {
		t.Errorf("total = %d, want 5000 after cooldown reset", total)
	}
}

func TestFetchReleases_PerPage100(t *testing.T) {
	var perPage string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		perPage = r.URL.Query().Get("per_page")
		rateHeaders(w, 5000, time.Now().Add(time.Hour))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]gh.RepositoryRelease{})
	})
	c := newTestClient(t, []string{"token1"}, handler)

	_, err := c.FetchReleases(context.Background(), "o", "n")
	if err != nil {
		t.Fatalf("FetchReleases: %v", err)
	}
	if perPage != "100" {
		t.Errorf("per_page = %q, want 100", perPage)
	}
}

func TestNewClient_EmptyTokensUsesUnauthenticated(t *testing.T) {
	c := NewClient(nil)
	if len(c.tokens) != 1 {
		t.Fatalf("expected 1 token state, got %d", len(c.tokens))
	}
	if c.tokens[0].remaining.Load() != 5000 {
		t.Errorf("expected initial remaining 5000, got %d", c.tokens[0].remaining.Load())
	}
}

func TestNextClient_AllExhausted(t *testing.T) {
	c := NewClient([]string{"token1"})
	c.tokens[0].remaining.Store(50)

	_, err := c.nextClient()
	if err == nil {
		t.Error("expected error when all tokens exhausted")
	}
}
