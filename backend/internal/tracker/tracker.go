package tracker

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	"getreleased/internal/avatar"
	"getreleased/internal/database"
	"getreleased/internal/github"
	"getreleased/internal/release"

	gh "github.com/google/go-github/v69/github"
	"golang.org/x/sync/errgroup"
)

type RepositoryStore interface {
	ListRepositories(ctx context.Context) ([]release.Repository, error)
	SaveRepository(ctx context.Context, r *release.Repository) (int64, error)
	SaveReleasesBatch(ctx context.Context, releases []release.Release) (bool, error)
	GetRepository(ctx context.Context, owner, name string) (*release.Repository, error)
	UpdateRepoMeta(ctx context.Context, r *release.Repository) error
}

type Tracker struct {
	store    RepositoryStore
	ghClient github.Fetcher
	avatar   *avatar.Downloader
}

func New(store RepositoryStore, ghClient github.Fetcher, avatarDL *avatar.Downloader) *Tracker {
	return &Tracker{store: store, ghClient: ghClient, avatar: avatarDL}
}

func (t *Tracker) Track(ctx context.Context) (bool, error) {
	repos, err := t.store.ListRepositories(ctx)
	if err != nil {
		return false, err
	}

	var dirty atomic.Bool
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	for _, repo := range repos {
		repo := repo
		g.Go(func() error {
			if gctx.Err() != nil {
				return gctx.Err()
			}
			changed, err := t.trackOne(gctx, repo.Owner, repo.Name)
			if err != nil {
				slog.ErrorContext(gctx, "track repo", "repo", repo.FullName, "err", err)
				return nil
			}
			if changed {
				dirty.Store(true)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return false, err
	}
	return dirty.Load(), nil
}

func (t *Tracker) TrackOne(ctx context.Context, owner, name string) (bool, error) {
	return t.trackOne(ctx, owner, name)
}

func (t *Tracker) trackOne(ctx context.Context, owner, name string) (bool, error) {
	existing, err := t.store.GetRepository(ctx, owner, name)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return false, err
	}

	var etag, lastModified string
	if existing != nil {
		etag = existing.ETag
		lastModified = existing.LastModified
	}

	result, err := t.ghClient.FetchRepository(ctx, owner, name, etag, lastModified)
	if err != nil {
		return false, err
	}
	if !result.Modified {
		return false, nil
	}

	pushedAt := result.Repo.GetPushedAt().Time
	if existing != nil && !existing.LastCheckedAt.IsZero() && !pushedAt.After(existing.LastCheckedAt) {
		meta := &release.Repository{
			FullName:      owner + "/" + name,
			Stars:         int(result.Repo.GetStargazersCount()),
			Language:      result.Repo.GetLanguage(),
			IsArchived:    result.Repo.GetArchived(),
			IsPrivate:     result.Repo.GetPrivate(),
			PushedAt:      pushedAt,
			ETag:          result.ETag,
			LastModified:  result.LastModified,
			LastCheckedAt: time.Now(),
		}
		if err := t.store.UpdateRepoMeta(ctx, meta); err != nil {
			return false, err
		}
		return false, nil
	}

	logoPath := ""
	if t.avatar != nil {
		avatarURL := result.Repo.GetOwner().GetAvatarURL()
		if lp, err := t.avatar.Download(ctx, owner, avatarURL); err != nil {
			slog.WarnContext(ctx, "avatar download", "owner", owner, "err", err)
		} else {
			logoPath = lp
		}
	}

	releases, err := t.ghClient.FetchReleases(ctx, owner, name)
	if err != nil {
		return false, err
	}

	repo := &release.Repository{
		Owner:         owner,
		Name:          name,
		FullName:      owner + "/" + name,
		Description:   result.Repo.GetDescription(),
		LogoPath:      logoPath,
		Stars:         int(result.Repo.GetStargazersCount()),
		Language:      result.Repo.GetLanguage(),
		IsArchived:    result.Repo.GetArchived(),
		IsPrivate:     result.Repo.GetPrivate(),
		PushedAt:      pushedAt,
		ETag:          result.ETag,
		LastModified:  result.LastModified,
		LastCheckedAt: time.Now(),
	}
	if latest := pickLatestRelease(releases); latest != nil {
		repo.LatestVersion = latest.GetTagName()
		repo.LatestReleaseURL = latest.GetHTMLURL()
		repo.LatestReleaseDate = latest.GetPublishedAt().Time
	}

	repoID, err := t.store.SaveRepository(ctx, repo)
	if err != nil {
		return false, err
	}

	batch := make([]release.Release, len(releases))
	for i, r := range releases {
		batch[i] = toRelease(repoID, r)
	}
	changed, err := t.store.SaveReleasesBatch(ctx, batch)
	if err != nil {
		return false, err
	}
	return changed, nil
}

func pickLatestRelease(releases []*gh.RepositoryRelease) *gh.RepositoryRelease {
	var latest *gh.RepositoryRelease
	for _, r := range releases {
		if latest == nil || r.GetPublishedAt().Time.After(latest.GetPublishedAt().Time) {
			latest = r
		}
	}
	return latest
}

func toRelease(repoID int64, r *gh.RepositoryRelease) release.Release {
	assets := make([]release.ReleaseAsset, len(r.Assets))
	for i, a := range r.Assets {
		assets[i] = release.ReleaseAsset{
			Name:        a.GetName(),
			Size:        int64(a.GetSize()),
			DownloadURL: a.GetBrowserDownloadURL(),
			ContentType: a.GetContentType(),
		}
	}
	return release.Release{
		RepositoryID: repoID,
		TagName:      r.GetTagName(),
		Name:         r.GetName(),
		Body:         r.GetBody(),
		HTMLURL:      r.GetHTMLURL(),
		TarballURL:   r.GetTarballURL(),
		ZipballURL:   r.GetZipballURL(),
		PublishedAt:  r.GetPublishedAt().Time,
		IsPrerelease: r.GetPrerelease(),
		Assets:       assets,
	}
}
