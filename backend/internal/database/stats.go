package database

import (
	"context"
	"time"
)

// StatsOverview 仪表盘概览计数。
type StatsOverview struct {
	RepositoryCount int `json:"repository_count"`
	ReleaseCount    int `json:"release_count"`
	TagCount        int `json:"tag_count"`
	UserCount       int `json:"user_count"`
	PrereleaseCount int `json:"prerelease_count"`
	ArchivedCount   int `json:"archived_count"`
	PrivateCount    int `json:"private_count"`
	UntaggedCount   int `json:"untagged_count"`
}

func (d *DB) StatsOverview(ctx context.Context) (StatsOverview, error) {
	var s StatsOverview
	var err error
	if s.RepositoryCount, err = d.CountRepositories(ctx); err != nil {
		return s, err
	}
	if s.ReleaseCount, err = d.CountReleases(ctx); err != nil {
		return s, err
	}
	if s.UserCount, err = d.CountUsers(ctx); err != nil {
		return s, err
	}
	if err = d.conn.GetContext(ctx, &s.TagCount, `SELECT COUNT(*) FROM tags`); err != nil {
		return s, err
	}
	if err = d.conn.GetContext(ctx, &s.PrereleaseCount, `SELECT COUNT(*) FROM releases WHERE is_prerelease = 1`); err != nil {
		return s, err
	}
	if err = d.conn.GetContext(ctx, &s.ArchivedCount, `SELECT COUNT(*) FROM repositories WHERE is_archived = 1`); err != nil {
		return s, err
	}
	if err = d.conn.GetContext(ctx, &s.PrivateCount, `SELECT COUNT(*) FROM repositories WHERE is_private = 1`); err != nil {
		return s, err
	}
	if err = d.conn.GetContext(ctx, &s.UntaggedCount,
		`SELECT COUNT(*) FROM repositories r WHERE NOT EXISTS(SELECT 1 FROM repository_tags rt WHERE rt.repository_id = r.id)`); err != nil {
		return s, err
	}
	return s, nil
}

// LanguageCount 按编程语言聚合的仓库数。
type LanguageCount struct {
	Language string `json:"language" db:"language"`
	Count    int    `json:"count" db:"count"`
}

func (d *DB) StatsLanguages(ctx context.Context) ([]LanguageCount, error) {
	var rows []LanguageCount
	err := d.conn.SelectContext(ctx, &rows,
		`SELECT COALESCE(NULLIF(language, ''), 'Unknown') AS language, COUNT(*) AS count
		 FROM repositories GROUP BY COALESCE(NULLIF(language, ''), 'Unknown') ORDER BY count DESC`)
	return rows, err
}

// TopRepository 按 Star 排序的仓库。
type TopRepository struct {
	ID            int64  `json:"id" db:"id"`
	FullName      string `json:"full_name" db:"full_name"`
	Stars         int    `json:"stars" db:"stars"`
	LatestVersion string `json:"latest_version" db:"latest_version"`
	Language      string `json:"language" db:"language"`
}

func (d *DB) StatsTopRepositories(ctx context.Context, limit int) ([]TopRepository, error) {
	var rows []TopRepository
	err := d.conn.SelectContext(ctx, &rows,
		`SELECT id, full_name, stars, COALESCE(latest_version, '') AS latest_version, COALESCE(language, '') AS language
		 FROM repositories ORDER BY stars DESC LIMIT ?`, limit)
	return rows, err
}

// RecentRelease 最近发布的 Release（带仓库全名）。
type RecentRelease struct {
	ID           int64     `json:"id" db:"id"`
	RepositoryID int64     `json:"repository_id" db:"repository_id"`
	FullName     string    `json:"full_name" db:"full_name"`
	TagName      string    `json:"tag_name" db:"tag_name"`
	Name         string    `json:"name" db:"name"`
	HTMLURL      string    `json:"html_url" db:"html_url"`
	PublishedAt  time.Time `json:"published_at" db:"published_at"`
	IsPrerelease bool      `json:"is_prerelease" db:"is_prerelease"`
}

func (d *DB) StatsRecentReleases(ctx context.Context, limit int) ([]RecentRelease, error) {
	var rows []RecentRelease
	err := d.conn.SelectContext(ctx, &rows,
		`SELECT r.id, r.repository_id, repo.full_name, r.tag_name,
		        COALESCE(r.name, '') AS name, COALESCE(r.html_url, '') AS html_url,
		        r.published_at, r.is_prerelease
		 FROM releases r JOIN repositories repo ON r.repository_id = repo.id
		 ORDER BY r.published_at DESC LIMIT ?`, limit)
	return rows, err
}

// ReleaseTrendPoint 按月聚合的发布数。
type ReleaseTrendPoint struct {
	Month string `json:"month" db:"month"`
	Count int    `json:"count" db:"count"`
}

// StatsReleaseTrend 返回最近 months 个月的发布趋势，缺失月份补 0。
func (d *DB) StatsReleaseTrend(ctx context.Context, months int) ([]ReleaseTrendPoint, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month()-time.Month(months)+1, 1, 0, 0, 0, 0, time.UTC)
	startKey := start.Format("2006-01")
	var rows []ReleaseTrendPoint
	err := d.conn.SelectContext(ctx, &rows,
		`SELECT substr(published_at, 1, 7) AS month, COUNT(*) AS count
		 FROM releases
		 WHERE published_at IS NOT NULL AND substr(published_at, 1, 7) >= ?
		 GROUP BY month ORDER BY month`, startKey)
	if err != nil {
		return nil, err
	}
	countMap := make(map[string]int, len(rows))
	for _, r := range rows {
		countMap[r.Month] = r.Count
	}
	trend := make([]ReleaseTrendPoint, 0, months)
	t := start
	for i := 0; i < months; i++ {
		key := t.Format("2006-01")
		trend = append(trend, ReleaseTrendPoint{Month: key, Count: countMap[key]})
		t = t.AddDate(0, 1, 0)
	}
	return trend, nil
}

// TagTypeCount 按标签类型聚合的数量。
type TagTypeCount struct {
	Type  string `json:"type" db:"type"`
	Count int    `json:"count" db:"count"`
}

func (d *DB) StatsTagTypes(ctx context.Context) ([]TagTypeCount, error) {
	var rows []TagTypeCount
	err := d.conn.SelectContext(ctx, &rows,
		`SELECT type, COUNT(*) AS count FROM tags GROUP BY type ORDER BY count DESC`)
	return rows, err
}
