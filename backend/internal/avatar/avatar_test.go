package avatar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadEmptyURL(t *testing.T) {
	dl := NewDownloader(t.TempDir())
	webPath, err := dl.Download(context.Background(), "o", "")
	if err != nil {
		t.Fatal(err)
	}
	if webPath != "" {
		t.Errorf("expected empty, got %q", webPath)
	}
}

func TestDownloadExistsSkip(t *testing.T) {
	dir := t.TempDir()
	dl := NewDownloader(dir)

	path := filepath.Join(dir, "octocat.png")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	webPath, err := dl.Download(context.Background(), "octocat", "https://example.com/x.png")
	if err != nil {
		t.Fatal(err)
	}
	if webPath != "assets/images/repos/octocat.png" {
		t.Errorf("got %q", webPath)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "existing" {
		t.Error("existing file was overwritten")
	}
}

func TestDownloadHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-png-data"))
	}))
	defer server.Close()

	dir := t.TempDir()
	dl := NewDownloader(dir)

	webPath, err := dl.Download(context.Background(), "octocat", server.URL+"/avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if webPath != "assets/images/repos/octocat.png" {
		t.Errorf("got %q", webPath)
	}
	data, err := os.ReadFile(filepath.Join(dir, "octocat.png"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fake-png-data" {
		t.Error("wrong file content")
	}

	webPath2, err := dl.Download(context.Background(), "octocat", server.URL+"/avatar.png")
	if err != nil {
		t.Fatal(err)
	}
	if webPath2 != webPath {
		t.Errorf("second download should skip, got %q", webPath2)
	}
}

func TestDownloadNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dl := NewDownloader(t.TempDir())
	if _, err := dl.Download(context.Background(), "o", server.URL+"/x"); err == nil {
		t.Error("expected error for 404")
	}
}

func TestSanitize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"octocat", "octocat"},
		{"ClashX-Pro", "ClashX-Pro"},
		{"a_b", "a_b"},
		{"中文", "__"},
		{"a.b", "a_b"},
		{"", ""},
	}
	for _, c := range cases {
		if got := sanitize(c.in); got != c.want {
			t.Errorf("sanitize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
