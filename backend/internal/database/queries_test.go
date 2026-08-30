package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"getreleased/internal/release"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSaveAndDeleteRepository(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	repoID, err := db.SaveRepository(ctx, &release.Repository{
		Owner:       "octocat",
		Name:        "hello-world",
		FullName:    "octocat/hello-world",
		Description: "demo",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if _, err := db.SaveReleasesBatch(ctx, []release.Release{
		{RepositoryID: repoID, TagName: "v1.0.0", Name: "v1.0.0"},
	}); err != nil {
		t.Fatalf("save releases: %v", err)
	}

	tag, err := db.GetOrCreateTag(ctx, "go", "category")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if err := db.TagRepository(ctx, repoID, tag.ID); err != nil {
		t.Fatalf("tag repo: %v", err)
	}

	if err := db.DeleteRepository(ctx, repoID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	count, err := db.CountRepositories(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 repos, got %d", count)
	}

	rc, err := db.CountReleases(ctx)
	if err != nil {
		t.Fatalf("count releases: %v", err)
	}
	if rc != 0 {
		t.Errorf("expected 0 releases after cascade, got %d", rc)
	}
}

func TestDeleteRepositoryNotFound(t *testing.T) {
	db := testDB(t)
	err := db.DeleteRepository(context.Background(), 9999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateRepository(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		id         int64
		desc       string
		remark     string
		wantErr    error
		wantDesc   string
		wantRemark string
		setup      bool
	}{
		{name: "existing", id: 1, desc: "updated", remark: "note", wantErr: nil, wantDesc: "updated", wantRemark: "note", setup: true},
		{name: "not_found", id: 9999, desc: "x", remark: "", wantErr: sql.ErrNoRows, setup: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup {
				id, err := db.SaveRepository(ctx, &release.Repository{
					Owner: "o", Name: "n", FullName: "o/n", Description: "old",
				})
				if err != nil {
					t.Fatal(err)
				}
				tc.id = id
			}
			err := db.UpdateRepository(ctx, tc.id, tc.desc, tc.remark)
			if err != tc.wantErr {
				t.Errorf("expected %v, got %v", tc.wantErr, err)
			}
			if tc.setup {
				repo, err := db.GetRepositoryByFullName(ctx, "o/n")
				if err != nil {
					t.Fatal(err)
				}
				if repo.Description != tc.wantDesc {
					t.Errorf("expected %q, got %q", tc.wantDesc, repo.Description)
				}
				if repo.Remark != tc.wantRemark {
					t.Errorf("expected remark %q, got %q", tc.wantRemark, repo.Remark)
				}
			}
		})
	}
}

func TestDeleteTag(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	tag, err := db.GetOrCreateTag(ctx, "rust", "category")
	if err != nil {
		t.Fatal(err)
	}
	repoID, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "o", Name: "n", FullName: "o/n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.TagRepository(ctx, repoID, tag.ID); err != nil {
		t.Fatal(err)
	}

	if err := db.DeleteTag(ctx, tag.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	tags, err := db.ListTags(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}

	repo, err := db.GetRepositoryByFullName(ctx, "o/n")
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.Tags) != 0 {
		t.Errorf("expected 0 repo tags after cascade, got %d", len(repo.Tags))
	}
}

func TestDeleteTagNotFound(t *testing.T) {
	db := testDB(t)
	err := db.DeleteTag(context.Background(), 9999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateTag(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	tag, err := db.GetOrCreateTag(ctx, "old-name", "category")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateTag(ctx, tag.ID, "new-name", "category"); err != nil {
		t.Fatalf("update: %v", err)
	}

	tags, err := db.ListTags(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Name != "new-name" || tags[0].Type != "category" {
		t.Errorf("expected new-name/category, got %+v", tags)
	}

	if err := db.UpdateTag(ctx, 9999, "x", ""); err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestSetRepositoryTagIDs(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	repoID, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "o", Name: "n", FullName: "o/n",
	})
	if err != nil {
		t.Fatal(err)
	}
	t1, _ := db.GetOrCreateTag(ctx, "t1", "category")
	t2, _ := db.GetOrCreateTag(ctx, "t2", "category")

	if err := db.SetRepositoryTagIDs(ctx, repoID, []int64{t1.ID, t2.ID}); err != nil {
		t.Fatalf("set: %v", err)
	}
	repo, _ := db.GetRepositoryByFullName(ctx, "o/n")
	if len(repo.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(repo.Tags))
	}

	if err := db.SetRepositoryTagIDs(ctx, repoID, []int64{t1.ID}); err != nil {
		t.Fatalf("set one: %v", err)
	}
	repo, _ = db.GetRepositoryByFullName(ctx, "o/n")
	if len(repo.Tags) != 1 || repo.Tags[0].Name != "t1" {
		t.Errorf("expected 1 tag t1, got %+v", repo.Tags)
	}

	if err := db.SetRepositoryTagIDs(ctx, repoID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	repo, _ = db.GetRepositoryByFullName(ctx, "o/n")
	if len(repo.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(repo.Tags))
	}
}

func TestSeedRepositories(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	seeds := []RepoSeed{
		{Owner: "a", Name: "b"},
		{Owner: "c", Name: "d"},
	}
	if err := db.SeedRepositories(ctx, seeds); err != nil {
		t.Fatal(err)
	}
	count, _ := db.CountRepositories(ctx)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}

	if err := db.SeedRepositories(ctx, seeds); err != nil {
		t.Fatal(err)
	}
	count, _ = db.CountRepositories(ctx)
	if count != 2 {
		t.Errorf("expected 2 after idempotent seed, got %d", count)
	}
}

func TestSaveRepositoryLogoPath(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const logo = "assets/images/repos/octocat.png"
	if _, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "o", Name: "n", FullName: "o/n", LogoPath: logo,
	}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.GetRepositoryByFullName(ctx, "o/n")
	if err != nil {
		t.Fatal(err)
	}
	if repo.LogoPath != logo {
		t.Errorf("expected %q, got %q", logo, repo.LogoPath)
	}

	if _, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "o", Name: "n", FullName: "o/n", Description: "updated", LogoPath: "",
	}); err != nil {
		t.Fatal(err)
	}
	repo, err = db.GetRepositoryByFullName(ctx, "o/n")
	if err != nil {
		t.Fatal(err)
	}
	if repo.LogoPath != logo {
		t.Errorf("empty logo_path should not overwrite existing, got %q", repo.LogoPath)
	}
	if repo.Description != "updated" {
		t.Errorf("expected description updated, got %q", repo.Description)
	}
}

func TestListRepositoriesLogoPath(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	if _, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "a", Name: "b", FullName: "a/b", LogoPath: "assets/images/repos/a.png",
	}); err != nil {
		t.Fatal(err)
	}
	repos, err := db.ListRepositories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].LogoPath != "assets/images/repos/a.png" {
		t.Errorf("unexpected repos: %+v", repos)
	}
}

func TestSaveRepositoryMetadata(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	if _, err := db.SaveRepository(ctx, &release.Repository{
		Owner: "o", Name: "n", FullName: "o/n",
		Stars: 42, Language: "Go", IsArchived: true, IsPrivate: false,
	}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.GetRepositoryByFullName(ctx, "o/n")
	if err != nil {
		t.Fatal(err)
	}
	if repo.Stars != 42 || repo.Language != "Go" || !repo.IsArchived || repo.IsPrivate {
		t.Errorf("unexpected metadata: %+v", repo)
	}
}

func TestUsersCRUD(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	count, err := db.CountUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 users, got %d", count)
	}

	if err := db.CreateUser(ctx, "admin", "$2a$10$hash", "admin"); err != nil {
		t.Fatal(err)
	}
	count, _ = db.CountUsers(ctx)
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}

	user, err := db.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil || user.Username != "admin" || user.Role != "admin" {
		t.Errorf("unexpected user: %+v", user)
	}

	missing, err := db.GetUserByUsername(ctx, "nobody")
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Errorf("expected nil for missing user, got %+v", missing)
	}
}

func TestSaveReleasesWithAssets(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	repoID, err := db.SaveRepository(ctx, &release.Repository{Owner: "o", Name: "n", FullName: "o/n"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.SaveReleasesBatch(ctx, []release.Release{
		{
			RepositoryID: repoID, TagName: "v1.0.0", Name: "v1.0.0",
			TarballURL: "https://api.github.com/tarball",
			ZipballURL: "https://api.github.com/zipball",
			Assets: []release.ReleaseAsset{
				{Name: "app.tar.gz", Size: 1024, DownloadURL: "https://x/app.tar.gz", ContentType: "application/gzip"},
				{Name: "app.zip", Size: 2048, DownloadURL: "https://x/app.zip", ContentType: "application/zip"},
			},
		},
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := db.GetReleasesByRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 release, got %d", len(got))
	}
	r := got[0]
	if r.TarballURL != "https://api.github.com/tarball" {
		t.Errorf("tarball: %q", r.TarballURL)
	}
	if r.ZipballURL != "https://api.github.com/zipball" {
		t.Errorf("zipball: %q", r.ZipballURL)
	}
	if len(r.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(r.Assets))
	}
	if r.Assets[0].Name != "app.tar.gz" || r.Assets[1].Name != "app.zip" {
		t.Errorf("assets order: %+v", r.Assets)
	}

	if _, err := db.SaveReleasesBatch(ctx, []release.Release{
		{RepositoryID: repoID, TagName: "v1.0.0", Name: "v1.0.0", Body: "updated", Assets: []release.ReleaseAsset{
			{Name: "single.bin", Size: 10, DownloadURL: "https://x/single.bin"},
		}},
	}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	got, _ = db.GetReleasesByRepository(ctx, repoID)
	if len(got[0].Assets) != 1 || got[0].Assets[0].Name != "single.bin" {
		t.Errorf("expected single.bin after upsert, got %+v", got[0].Assets)
	}

	all, err := db.ListReleases(ctx)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 1 || len(all[0].Assets) != 1 {
		t.Errorf("ListReleases assets mismatch: %+v", all)
	}
}

func TestSaveReleasesBatchDiff(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	repoID, err := db.SaveRepository(ctx, &release.Repository{Owner: "o", Name: "n", FullName: "o/n"})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	rel := release.Release{
		RepositoryID: repoID, TagName: "v1.0.0", Name: "v1.0.0",
		Body:        "initial body",
		HTMLURL:     "https://example.com/v1.0.0",
		PublishedAt: now,
	}

	changed, err := db.SaveReleasesBatch(ctx, []release.Release{rel})
	if err != nil {
		t.Fatalf("first save: %v", err)
	}
	if !changed {
		t.Error("first save should report changed=true")
	}

	changed, err = db.SaveReleasesBatch(ctx, []release.Release{rel})
	if err != nil {
		t.Fatalf("identical save: %v", err)
	}
	if changed {
		t.Error("identical save should report changed=false")
	}

	rel.Body = "updated body"
	changed, err = db.SaveReleasesBatch(ctx, []release.Release{rel})
	if err != nil {
		t.Fatalf("modified save: %v", err)
	}
	if !changed {
		t.Error("modified save should report changed=true")
	}

	emptyChanged, err := db.SaveReleasesBatch(ctx, nil)
	if err != nil {
		t.Fatalf("empty save: %v", err)
	}
	if emptyChanged {
		t.Error("empty save should report changed=false")
	}
}

func TestGetRepositoryAndMeta(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	if _, err := db.GetRepository(ctx, "o", "n"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	repoID, err := db.SaveRepository(ctx, &release.Repository{Owner: "o", Name: "n", FullName: "o/n", Stars: 10})
	if err != nil {
		t.Fatal(err)
	}
	_ = repoID

	got, err := db.GetRepository(ctx, "o", "n")
	if err != nil {
		t.Fatalf("GetRepository: %v", err)
	}
	if got.Stars != 10 {
		t.Errorf("Stars = %d, want 10", got.Stars)
	}

	now := time.Now()
	meta := &release.Repository{
		FullName:      "o/n",
		Stars:         42,
		Language:      "Go",
		PushedAt:      now,
		ETag:          "etag-1",
		LastModified:  "Wed, 01 Jan 2025 00:00:00 GMT",
		LastCheckedAt: now,
	}
	if err := db.UpdateRepoMeta(ctx, meta); err != nil {
		t.Fatalf("UpdateRepoMeta: %v", err)
	}

	got, err = db.GetRepository(ctx, "o", "n")
	if err != nil {
		t.Fatalf("GetRepository after meta: %v", err)
	}
	if got.Stars != 42 {
		t.Errorf("Stars = %d, want 42", got.Stars)
	}
	if got.ETag != "etag-1" {
		t.Errorf("ETag = %q, want etag-1", got.ETag)
	}
	if got.Language != "Go" {
		t.Errorf("Language = %q, want Go", got.Language)
	}
}
