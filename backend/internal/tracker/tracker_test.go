package tracker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"getreleased/internal/database"
	"getreleased/internal/github"
	"getreleased/internal/release"

	gh "github.com/google/go-github/v69/github"
)

type fakeStore struct {
	repos        []release.Repository
	existing     *release.Repository
	listErr      error
	savedRepo    *release.Repository
	savedBatch   []release.Release
	saveErr      error
	batchErr     error
	batchChanged bool
	updatedMeta  atomic.Int64
	getErr       error
}

func (f *fakeStore) ListRepositories(ctx context.Context) ([]release.Repository, error) {
	return f.repos, f.listErr
}

func (f *fakeStore) SaveRepository(ctx context.Context, r *release.Repository) (int64, error) {
	f.savedRepo = r
	if f.saveErr != nil {
		return 0, f.saveErr
	}
	return 1, nil
}

func (f *fakeStore) SaveReleasesBatch(ctx context.Context, releases []release.Release) (bool, error) {
	f.savedBatch = releases
	return f.batchChanged, f.batchErr
}

func (f *fakeStore) GetRepository(ctx context.Context, owner, name string) (*release.Repository, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.existing == nil {
		return nil, database.ErrNotFound
	}
	return f.existing, nil
}

func (f *fakeStore) UpdateRepoMeta(ctx context.Context, r *release.Repository) error {
	f.updatedMeta.Add(1)
	return nil
}

type fakeFetcher struct {
	repo      *gh.Repository
	releases  []*gh.RepositoryRelease
	modified  bool
	etag      string
	lastMod   string
	repoErr   error
	relErr    error
	rateLimit int
}

func (f *fakeFetcher) FetchRepository(ctx context.Context, owner, repo, etag, lastModified string) (*github.FetchResult, error) {
	if f.repoErr != nil {
		return nil, f.repoErr
	}
	modified := f.modified
	if !modified {
		return &github.FetchResult{Modified: false}, nil
	}
	return &github.FetchResult{Repo: f.repo, Modified: true, ETag: f.etag, LastModified: f.lastMod}, nil
}

func (f *fakeFetcher) FetchReleases(ctx context.Context, owner, repo string) ([]*gh.RepositoryRelease, error) {
	return f.releases, f.relErr
}

func (f *fakeFetcher) RateLimitRemaining(ctx context.Context) (int, error) {
	if f.rateLimit == 0 {
		return 5000, nil
	}
	return f.rateLimit, nil
}

func newRepo() *gh.Repository {
	return &gh.Repository{
		Name:            gh.Ptr("n"),
		Owner:           &gh.User{Login: gh.Ptr("o")},
		Description:     gh.Ptr("desc"),
		StargazersCount: gh.Ptr(100),
		Language:        gh.Ptr("Go"),
		PushedAt:        &gh.Timestamp{Time: time.Now()},
	}
}

func TestTrack_NoRepos(t *testing.T) {
	store := &fakeStore{repos: nil}
	trk := New(store, &fakeFetcher{}, nil)
	dirty, err := trk.Track(context.Background())
	if err != nil {
		t.Fatalf("Track: %v", err)
	}
	if dirty {
		t.Error("expected dirty=false for no repos")
	}
}

func TestTrack_ListError(t *testing.T) {
	store := &fakeStore{listErr: errors.New("db down")}
	trk := New(store, &fakeFetcher{}, nil)
	if _, err := trk.Track(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackOne_FetchRepoError(t *testing.T) {
	store := &fakeStore{}
	fetcher := &fakeFetcher{repoErr: errors.New("not found")}
	trk := New(store, fetcher, nil)
	if _, err := trk.TrackOne(context.Background(), "o", "n"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackOne_FetchReleasesError(t *testing.T) {
	store := &fakeStore{}
	fetcher := &fakeFetcher{
		repo:     newRepo(),
		modified: true,
		relErr:   errors.New("rate limit"),
	}
	trk := New(store, fetcher, nil)
	if _, err := trk.TrackOne(context.Background(), "o", "n"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackOne_SaveRepoError(t *testing.T) {
	store := &fakeStore{saveErr: errors.New("db write fail")}
	fetcher := &fakeFetcher{repo: newRepo(), modified: true}
	trk := New(store, fetcher, nil)
	if _, err := trk.TrackOne(context.Background(), "o", "n"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTrackOne_Success(t *testing.T) {
	store := &fakeStore{batchChanged: true}
	now := time.Now()
	fetcher := &fakeFetcher{
		repo:     newRepo(),
		modified: true,
		releases: []*gh.RepositoryRelease{
			{
				TagName:     gh.Ptr("v1.0.0"),
				Name:        gh.Ptr("Release 1"),
				PublishedAt: &gh.Timestamp{Time: now},
				HTMLURL:     gh.Ptr("https://example.com/v1.0.0"),
			},
		},
	}
	trk := New(store, fetcher, nil)
	dirty, err := trk.TrackOne(context.Background(), "o", "n")
	if err != nil {
		t.Fatalf("TrackOne: %v", err)
	}
	if !dirty {
		t.Error("expected dirty=true on new release")
	}
	if store.savedRepo == nil {
		t.Fatal("SaveRepository not called")
	}
	if store.savedRepo.LatestVersion != "v1.0.0" {
		t.Errorf("LatestVersion = %q, want v1.0.0", store.savedRepo.LatestVersion)
	}
	if store.savedRepo.Stars != 100 {
		t.Errorf("Stars = %d, want 100", store.savedRepo.Stars)
	}
	if len(store.savedBatch) != 1 {
		t.Fatalf("savedBatch len = %d, want 1", len(store.savedBatch))
	}
	if store.savedBatch[0].TagName != "v1.0.0" {
		t.Errorf("batch TagName = %q, want v1.0.0", store.savedBatch[0].TagName)
	}
}

func TestTrack_ContinuesOnError(t *testing.T) {
	store := &fakeStore{
		repos: []release.Repository{
			{Owner: "o1", Name: "n1"},
			{Owner: "o2", Name: "n2"},
		},
	}
	fetcher := &fakeFetcher{repoErr: errors.New("fail")}
	trk := New(store, fetcher, nil)
	if _, err := trk.Track(context.Background()); err != nil {
		t.Fatalf("Track should not fail on individual repo error: %v", err)
	}
}

func TestTrack_ContextCancel(t *testing.T) {
	store := &fakeStore{
		repos: []release.Repository{{Owner: "o", Name: "n"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fetcher := &fakeFetcher{repo: newRepo(), modified: true}
	trk := New(store, fetcher, nil)
	if _, err := trk.Track(ctx); err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestTrackOne_NotModified304(t *testing.T) {
	store := &fakeStore{}
	fetcher := &fakeFetcher{modified: false}
	trk := New(store, fetcher, nil)
	dirty, err := trk.TrackOne(context.Background(), "o", "n")
	if err != nil {
		t.Fatalf("TrackOne 304: %v", err)
	}
	if dirty {
		t.Error("expected dirty=false on 304")
	}
	if store.savedRepo != nil {
		t.Error("SaveRepository should not be called on 304")
	}
}

func TestTrackOne_PushedAtSkip(t *testing.T) {
	oldPushed := time.Now().Add(-2 * time.Hour)
	oldCheck := time.Now().Add(-1 * time.Hour)
	store := &fakeStore{
		existing: &release.Repository{
			Owner:         "o",
			Name:          "n",
			FullName:      "o/n",
			LastCheckedAt: oldCheck,
		},
	}
	repo := newRepo()
	repo.PushedAt = &gh.Timestamp{Time: oldPushed}
	fetcher := &fakeFetcher{repo: repo, modified: true, relErr: errors.New("should not fetch releases")}
	trk := New(store, fetcher, nil)
	dirty, err := trk.TrackOne(context.Background(), "o", "n")
	if err != nil {
		t.Fatalf("TrackOne pushed_at skip: %v", err)
	}
	if dirty {
		t.Error("expected dirty=false when pushed_at <= last_checked_at")
	}
	if store.updatedMeta.Load() != 1 {
		t.Error("UpdateRepoMeta should be called once")
	}
	if store.savedRepo != nil {
		t.Error("SaveRepository should not be called on pushed_at skip")
	}
}

func TestTrack_DirtyFlag(t *testing.T) {
	store := &fakeStore{
		repos: []release.Repository{
			{Owner: "o", Name: "n"},
		},
		batchChanged: true,
	}
	now := time.Now()
	fetcher := &fakeFetcher{
		repo:     newRepo(),
		modified: true,
		releases: []*gh.RepositoryRelease{
			{TagName: gh.Ptr("v1.0.0"), PublishedAt: &gh.Timestamp{Time: now}},
		},
	}
	trk := New(store, fetcher, nil)
	dirty, err := trk.Track(context.Background())
	if err != nil {
		t.Fatalf("Track: %v", err)
	}
	if !dirty {
		t.Error("expected dirty=true when batch changed")
	}
}

func TestTrack_NoDirty(t *testing.T) {
	store := &fakeStore{
		repos: []release.Repository{
			{Owner: "o", Name: "n"},
		},
		batchChanged: false,
	}
	now := time.Now()
	fetcher := &fakeFetcher{
		repo:     newRepo(),
		modified: true,
		releases: []*gh.RepositoryRelease{
			{TagName: gh.Ptr("v1.0.0"), PublishedAt: &gh.Timestamp{Time: now}},
		},
	}
	trk := New(store, fetcher, nil)
	dirty, err := trk.Track(context.Background())
	if err != nil {
		t.Fatalf("Track: %v", err)
	}
	if dirty {
		t.Error("expected dirty=false when batch unchanged")
	}
}

func TestTrack_Concurrent(t *testing.T) {
	const n = 20
	repos := make([]release.Repository, n)
	for i := range repos {
		repos[i] = release.Repository{Owner: "o", Name: "n"}
	}
	store := &fakeStore{repos: repos, batchChanged: true}
	now := time.Now()
	fetcher := &fakeFetcher{
		repo:     newRepo(),
		modified: true,
		releases: []*gh.RepositoryRelease{
			{TagName: gh.Ptr("v1.0.0"), PublishedAt: &gh.Timestamp{Time: now}},
		},
	}
	trk := New(store, fetcher, nil)
	dirty, err := trk.Track(context.Background())
	if err != nil {
		t.Fatalf("Track: %v", err)
	}
	if !dirty {
		t.Error("expected dirty=true")
	}
}
