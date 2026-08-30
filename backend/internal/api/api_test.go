package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	gh "github.com/google/go-github/v69/github"

	"getreleased/internal/auth"
	"getreleased/internal/database"
	"getreleased/internal/github"
	"getreleased/internal/release"
	"getreleased/internal/tracker"
)

func setupTestServer(t *testing.T) (*httptest.Server, *database.DB, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.OpenWithMaxConns(dbPath, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	authSvc, err := auth.NewService(db, "testsecret", 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := authSvc.EnsureAdminSeed(context.Background(), "admin", "testpass"); err != nil {
		t.Fatal(err)
	}

	api := New(db, tracker.New(db, nil, nil), nil, authSvc, filepath.Join(dir, "data"))
	server := httptest.NewServer(api.Router(false))
	t.Cleanup(server.Close)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "testpass"})
	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	return server, db, loginResp.Data.Token
}

func authedGet(server *httptest.Server, token, path string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, server.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return http.DefaultClient.Do(req)
}

func authedRequest(server *httptest.Server, token, method, path string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			panic(err)
		}
	}
	req, _ := http.NewRequest(method, server.URL+path, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func decodeData(resp *http.Response, v any) {
	defer func() { _ = resp.Body.Close() }()
	var wrapper struct {
		Data any `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&wrapper)
	if v != nil {
		tmp, _ := json.Marshal(wrapper.Data)
		_ = json.Unmarshal(tmp, v)
	}
}

func decodeError(resp *http.Response) string {
	defer func() { _ = resp.Body.Close() }()
	var wrapper struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&wrapper)
	return wrapper.Error
}

func TestLoginWrongPassword(t *testing.T) {
	server, _, _ := setupTestServer(t)
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestLoginWrongUsername(t *testing.T) {
	server, _, _ := setupTestServer(t)
	body, _ := json.Marshal(map[string]string{"username": "nobody", "password": "testpass"})
	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestAdminWithoutToken(t *testing.T) {
	server, _, _ := setupTestServer(t)
	resp, err := http.Get(server.URL + "/api/admin/repositories")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestTagCRUD(t *testing.T) {
	server, _, token := setupTestServer(t)

	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/tags", map[string]string{"name": "go", "type": "category"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	var tag release.Tag
	decodeData(resp, &tag)
	if tag.Name != "go" || tag.Type != "category" {
		t.Errorf("expected go/category, got %s/%s", tag.Name, tag.Type)
	}

	resp, err = authedRequest(server, token, http.MethodPost, "/api/admin/tags", map[string]string{"name": "go"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("duplicate: expected 409, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = authedRequest(server, token, http.MethodPut, "/api/admin/tags/"+itoa(tag.ID), map[string]string{"name": "golang", "type": "platform"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	var updated release.Tag
	decodeData(resp, &updated)
	if updated.Type != "platform" {
		t.Errorf("expected type platform, got %s", updated.Type)
	}
	_ = resp.Body.Close()

	resp, err = authedGet(server, token, "/api/admin/tags")
	if err != nil {
		t.Fatal(err)
	}
	var tags []release.Tag
	decodeData(resp, &tags)
	if len(tags) != 1 || tags[0].Name != "golang" || tags[0].Type != "platform" {
		t.Errorf("expected [golang/platform], got %+v", tags)
	}

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/tags/"+itoa(tag.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/tags/"+itoa(tag.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("delete missing: expected 404, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestRepositoryUpdateAndDelete(t *testing.T) {
	server, db, token := setupTestServer(t)
	ctx := context.Background()

	repoID, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "octocat", Name: "hello", FullName: "octocat/hello", Description: "old",
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := authedRequest(server, token, http.MethodPut, "/api/admin/repositories/"+itoa(repoID), map[string]string{"description": "new", "remark": "manual note"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	repo, _ := db.GetRepositoryByFullName(ctx, "octocat/hello")
	if repo.Description != "new" {
		t.Errorf("expected new, got %s", repo.Description)
	}
	if repo.Remark != "manual note" {
		t.Errorf("expected remark manual note, got %q", repo.Remark)
	}

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/repositories/"+itoa(repoID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	count, _ := db.CountRepositories(ctx)
	if count != 0 {
		t.Errorf("expected 0 repos, got %d", count)
	}
}

func TestSetRepositoryTags(t *testing.T) {
	server, db, token := setupTestServer(t)
	ctx := context.Background()

	repoID, _ := db.SaveRepository(ctx, &release.Repository{Owner: "o", Name: "n", FullName: "o/n"})
	tag1, _ := db.CreateTag(ctx, "t1", "category")
	tag2, _ := db.CreateTag(ctx, "t2", "platform")

	resp, err := authedRequest(server, token, http.MethodPut, "/api/admin/repositories/"+itoa(repoID)+"/tags", map[string]any{"tag_ids": []int64{tag1.ID, tag2.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("set tags: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	repo, _ := db.GetRepositoryByFullName(ctx, "o/n")
	if len(repo.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(repo.Tags))
	}

	resp, err = authedRequest(server, token, http.MethodPut, "/api/admin/repositories/"+itoa(repoID)+"/tags", map[string]any{"tag_ids": []int64{}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("clear tags: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	repo, _ = db.GetRepositoryByFullName(ctx, "o/n")
	if len(repo.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(repo.Tags))
	}
}

func TestTrackStatus(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedGet(server, token, "/api/admin/track/status")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var status struct {
		Running bool `json:"running"`
	}
	decodeData(resp, &status)
	if status.Running {
		t.Error("expected not running")
	}
}

func TestInvalidID(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedRequest(server, token, http.MethodDelete, "/api/admin/repositories/abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSyncRepositoryNotFound(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/repositories/999999/sync", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", resp.StatusCode, decodeError(resp))
	}
}

func TestExport(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/export", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	var result struct {
		Exported bool `json:"exported"`
	}
	decodeData(resp, &result)
	if !result.Exported {
		t.Error("expected exported true")
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

func TestListTagsEmptyReturnsArray(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedGet(server, token, "/api/admin/tags")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if string(wrapper.Data) != "[]" {
		t.Errorf("expected [], got %s", string(wrapper.Data))
	}
}

func TestListRepositoriesEmptyReturnsArray(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedGet(server, token, "/api/admin/repositories")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if string(wrapper.Data) != "[]" {
		t.Errorf("expected [], got %s", string(wrapper.Data))
	}
}

func loginAs(server *httptest.Server, username, password string) string {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		panic(err)
	}
	return loginResp.Data.Token
}

func TestUserListAndCreate(t *testing.T) {
	server, _, token := setupTestServer(t)

	resp, err := authedGet(server, token, "/api/admin/users")
	if err != nil {
		t.Fatal(err)
	}
	var users []release.User
	decodeData(resp, &users)
	if len(users) != 1 || users[0].Username != "admin" {
		t.Fatalf("expected [admin], got %+v", users)
	}

	resp, err = authedRequest(server, token, http.MethodPost, "/api/admin/users", map[string]string{"username": "alice", "password": "alicepass"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	var created release.User
	decodeData(resp, &created)
	if created.Username != "alice" || created.ID == 0 {
		t.Errorf("unexpected created user: %+v", created)
	}

	resp, err = authedRequest(server, token, http.MethodPost, "/api/admin/users", map[string]string{"username": "alice", "password": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("duplicate: expected 409, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = authedRequest(server, token, http.MethodPost, "/api/admin/users", map[string]string{"username": "", "password": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty username: expected 400, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestResetPassword(t *testing.T) {
	server, _, token := setupTestServer(t)

	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/users", map[string]string{"username": "bob", "password": "bobpass"})
	if err != nil {
		t.Fatal(err)
	}
	var bob release.User
	decodeData(resp, &bob)

	resp, err = authedRequest(server, token, http.MethodPut, "/api/admin/users/"+itoa(bob.ID)+"/password", map[string]string{"password": "newpass"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	token2 := loginAs(server, "bob", "newpass")
	if token2 == "" {
		t.Error("expected login with new password to succeed")
	}

	resp, err = authedRequest(server, token, http.MethodPut, "/api/admin/users/999999/password", map[string]string{"password": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("reset missing: expected 404, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = authedRequest(server, token, http.MethodPut, "/api/admin/users/"+itoa(bob.ID)+"/password", map[string]string{"password": ""})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty password: expected 400, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestDeleteUser(t *testing.T) {
	server, _, token := setupTestServer(t)

	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/users", map[string]string{"username": "carol", "password": "carolpass"})
	if err != nil {
		t.Fatal(err)
	}
	var carol release.User
	decodeData(resp, &carol)

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/users/"+itoa(carol.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/users/"+itoa(carol.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("delete missing: expected 404, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestDeleteUserSelf(t *testing.T) {
	server, _, token := setupTestServer(t)
	resp, err := authedGet(server, token, "/api/admin/users")
	if err != nil {
		t.Fatal(err)
	}
	var users []release.User
	decodeData(resp, &users)
	adminID := users[0].ID

	resp, err = authedRequest(server, token, http.MethodDelete, "/api/admin/users/"+itoa(adminID), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("delete self: expected 400, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	_ = resp.Body.Close()
}

type fakeGHClient struct {
	repo     *gh.Repository
	releases []*gh.RepositoryRelease
	err      error
}

func (f *fakeGHClient) FetchRepository(ctx context.Context, owner, repo, etag, lastModified string) (*github.FetchResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &github.FetchResult{Repo: f.repo, Modified: true}, nil
}

func (f *fakeGHClient) FetchReleases(ctx context.Context, owner, repo string) ([]*gh.RepositoryRelease, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.releases, nil
}

func (f *fakeGHClient) RateLimitRemaining(ctx context.Context) (int, error) {
	return 5000, nil
}

func TestCreateRepositoryExports(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.OpenWithMaxConns(dbPath, 5, 5)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}

	authSvc, err := auth.NewService(db, "testsecret", 12*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := authSvc.EnsureAdminSeed(context.Background(), "admin", "testpass"); err != nil {
		t.Fatal(err)
	}

	fakeGH := &fakeGHClient{
		repo: &gh.Repository{
			FullName: gh.Ptr("octocat/hello"),
			Owner:    &gh.User{Login: gh.Ptr("octocat")},
			Name:     gh.Ptr("hello"),
		},
	}
	trk := tracker.New(db, fakeGH, nil)
	exportDir := filepath.Join(dir, "data")
	api := New(db, trk, fakeGH, authSvc, exportDir)
	server := httptest.NewServer(api.Router(false))
	t.Cleanup(server.Close)

	token := loginAs(server, "admin", "testpass")

	resp, err := authedRequest(server, token, http.MethodPost, "/api/admin/repositories", map[string]string{"owner": "octocat", "name": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create: expected 200, got %d: %s", resp.StatusCode, decodeError(resp))
	}
	var repo release.Repository
	decodeData(resp, &repo)
	if repo.FullName != "octocat/hello" {
		t.Errorf("expected octocat/hello, got %s", repo.FullName)
	}

	entries, err := os.ReadDir(exportDir)
	if err != nil {
		t.Fatalf("read export dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected exported JSON files after create")
	}
}
