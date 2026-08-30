package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	gh "github.com/google/go-github/v69/github"
)

const rateLimitThreshold = 100

type tokenState struct {
	client    *gh.Client
	remaining atomic.Int64
	resetAt   atomic.Int64
}

type Client struct {
	tokens  []*tokenState
	counter atomic.Int64
}

func NewClient(tokens []string) *Client {
	if len(tokens) == 0 {
		tokens = []string{""}
	}
	pool := make([]*tokenState, len(tokens))
	for i, t := range tokens {
		c := gh.NewClient(nil)
		if t != "" {
			c = c.WithAuthToken(t)
		}
		pool[i] = &tokenState{client: c}
		pool[i].remaining.Store(5000)
	}
	return &Client{tokens: pool}
}

func (c *Client) nextClient() (*tokenState, error) {
	now := time.Now().Unix()
	for i := 0; i < len(c.tokens); i++ {
		idx := int(c.counter.Add(1)) % len(c.tokens)
		t := c.tokens[idx]
		if resetAt := t.resetAt.Load(); resetAt > 0 && now > resetAt {
			t.remaining.Store(5000)
			t.resetAt.Store(0)
		}
		if t.remaining.Load() > rateLimitThreshold {
			return t, nil
		}
	}
	return nil, errors.New("all tokens exhausted or cooling down")
}

func (c *Client) updateRemaining(t *tokenState, resp *gh.Response) {
	if resp != nil && resp.Rate.Remaining > 0 {
		t.remaining.Store(int64(resp.Rate.Remaining))
		t.resetAt.Store(resp.Rate.Reset.Unix())
	}
}

func (c *Client) markCooldown(t *tokenState, resp *gh.Response) {
	if resp != nil && resp.Rate.Reset.Unix() > 0 {
		t.remaining.Store(0)
		t.resetAt.Store(resp.Rate.Reset.Unix())
	} else {
		t.remaining.Store(0)
		t.resetAt.Store(time.Now().Add(time.Minute).Unix())
	}
}

type FetchResult struct {
	Repo         *gh.Repository
	ETag         string
	LastModified string
	Modified     bool
}

func (c *Client) FetchRepository(ctx context.Context, owner, repo, etag, lastModified string) (*FetchResult, error) {
	var lastErr error
	for i := 0; i < len(c.tokens); i++ {
		t, err := c.nextClient()
		if err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, err
		}
		req, err := t.client.NewRequest(http.MethodGet, fmt.Sprintf("repos/%s/%s", owner, repo), nil)
		if err != nil {
			return nil, err
		}
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		if lastModified != "" {
			req.Header.Set("If-Modified-Since", lastModified)
		}

		var repoObj gh.Repository
		resp, err := t.client.Do(ctx, req, &repoObj)
		if resp != nil && resp.StatusCode == http.StatusNotModified {
			result := &FetchResult{
				ETag:         resp.Header.Get("ETag"),
				LastModified: resp.Header.Get("Last-Modified"),
				Modified:     false,
			}
			c.updateRemaining(t, resp)
			return result, nil
		}
		if err != nil {
			if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) {
				c.markCooldown(t, resp)
				lastErr = err
				slog.WarnContext(ctx, "token rate limited, cooldown and try next", "repo", owner+"/"+repo, "status", resp.StatusCode, "reset_at", time.Unix(t.resetAt.Load(), 0))
				continue
			}
			return nil, err
		}

		result := &FetchResult{
			ETag:         resp.Header.Get("ETag"),
			LastModified: resp.Header.Get("Last-Modified"),
			Modified:     resp.StatusCode != http.StatusNotModified,
		}
		if result.Modified {
			result.Repo = &repoObj
		}
		c.updateRemaining(t, resp)
		return result, nil
	}
	return nil, lastErr
}

func (c *Client) FetchReleases(ctx context.Context, owner, repo string) ([]*gh.RepositoryRelease, error) {
	var lastErr error
	for i := 0; i < len(c.tokens); i++ {
		t, err := c.nextClient()
		if err != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, err
		}
		releases, resp, err := t.client.Repositories.ListReleases(ctx, owner, repo, &gh.ListOptions{PerPage: 100})
		if err == nil {
			c.updateRemaining(t, resp)
			return releases, nil
		}
		lastErr = err
		if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) {
			c.markCooldown(t, resp)
			slog.WarnContext(ctx, "token rate limited, cooldown and try next", "repo", owner+"/"+repo, "status", resp.StatusCode, "reset_at", time.Unix(t.resetAt.Load(), 0))
			continue
		}
		c.updateRemaining(t, resp)
		return nil, err
	}
	return nil, lastErr
}

func (c *Client) RateLimitRemaining(ctx context.Context) (int, error) {
	now := time.Now().Unix()
	total := 0
	queried := 0
	for _, t := range c.tokens {
		if resetAt := t.resetAt.Load(); resetAt > 0 && now > resetAt {
			t.remaining.Store(5000)
			t.resetAt.Store(0)
		}
		if r := t.remaining.Load(); r > 0 {
			total += int(r)
			queried++
			continue
		}
		rl, _, err := t.client.RateLimit.Get(ctx)
		if err != nil {
			continue
		}
		core := rl.GetCore()
		if core == nil {
			continue
		}
		t.remaining.Store(int64(core.Remaining))
		t.resetAt.Store(core.Reset.Unix())
		total += core.Remaining
		queried++
	}
	if queried == 0 {
		return 0, errors.New("all token rate limit queries failed")
	}
	return total, nil
}

type Fetcher interface {
	FetchRepository(ctx context.Context, owner, repo, etag, lastModified string) (*FetchResult, error)
	FetchReleases(ctx context.Context, owner, repo string) ([]*gh.RepositoryRelease, error)
	RateLimitRemaining(ctx context.Context) (int, error)
}
