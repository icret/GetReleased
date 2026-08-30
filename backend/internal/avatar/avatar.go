package avatar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	webPrefix   = "assets/images/repos/"
	maxBodySize = 10 << 20
)

type Downloader struct {
	dir    string
	client *http.Client
}

func NewDownloader(dir string) *Downloader {
	return &Downloader{dir: dir, client: &http.Client{Timeout: 30 * time.Second}}
}

func (d *Downloader) Download(ctx context.Context, owner, avatarURL string) (string, error) {
	if avatarURL == "" {
		return "", nil
	}
	safeName := sanitize(owner) + ".png"
	path := filepath.Join(d.dir, safeName)
	webPath := webPrefix + safeName

	if _, err := os.Stat(path); err == nil {
		return webPath, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("avatar request: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("avatar fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avatar status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return "", fmt.Errorf("avatar read: %w", err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("avatar empty body")
	}
	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		return "", fmt.Errorf("avatar mkdir: %w", err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("avatar write: %w", err)
	}
	return webPath, nil
}

func sanitize(owner string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, owner)
}
