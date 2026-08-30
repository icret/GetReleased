package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"getreleased/internal/release"
)

var ErrDuplicate = errors.New("duplicate")

var ErrNotFound = errors.New("not found")

func IsUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func NormalizeTagType(t string) string {
	switch t {
	case "category", "platform":
		return t
	default:
		return "category"
	}
}

func (d *DB) CreateTag(ctx context.Context, name, tagType string) (*release.Tag, error) {
	t := NormalizeTagType(tagType)
	res, err := d.conn.ExecContext(ctx, `INSERT INTO tags (name, type) VALUES (?, ?)`, name, t)
	if err != nil {
		if IsUniqueConstraintError(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &release.Tag{ID: id, Name: name, Type: t}, nil
}

func (d *DB) SaveRepository(ctx context.Context, r *release.Repository) (int64, error) {
	var id int64
	err := d.conn.QueryRowContext(ctx,
		`INSERT INTO repositories (
		   owner, name, full_name, description, logo_path,
		   stars, language, is_archived, is_private, pushed_at,
		   latest_version, latest_release_url, latest_release_date,
		   etag, last_modified, last_checked_at, updated_at
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(full_name) DO UPDATE SET
		   description         = excluded.description,
		   logo_path           = COALESCE(NULLIF(excluded.logo_path, ''), repositories.logo_path),
		   stars               = excluded.stars,
		   language            = excluded.language,
		   is_archived         = excluded.is_archived,
		   is_private          = excluded.is_private,
		   pushed_at           = excluded.pushed_at,
		   latest_version      = excluded.latest_version,
		   latest_release_url  = excluded.latest_release_url,
		   latest_release_date = excluded.latest_release_date,
		   etag                = excluded.etag,
		   last_modified       = excluded.last_modified,
		   last_checked_at     = excluded.last_checked_at,
		   updated_at          = CURRENT_TIMESTAMP
		 RETURNING id`,
		r.Owner, r.Name, r.FullName, r.Description, r.LogoPath,
		r.Stars, r.Language, r.IsArchived, r.IsPrivate, r.PushedAt,
		r.LatestVersion, r.LatestReleaseURL, r.LatestReleaseDate,
		r.ETag, r.LastModified, r.LastCheckedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *DB) GetRepository(ctx context.Context, owner, name string) (*release.Repository, error) {
	repo, err := d.GetRepositoryByFullName(ctx, owner+"/"+name)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, ErrNotFound
	}
	return repo, nil
}

func (d *DB) UpdateRepoMeta(ctx context.Context, r *release.Repository) error {
	res, err := d.conn.ExecContext(ctx,
		`UPDATE repositories SET
		   stars = ?, language = ?, is_archived = ?, is_private = ?,
		   pushed_at = ?, etag = ?, last_modified = ?,
		   last_checked_at = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE full_name = ?`,
		r.Stars, r.Language, r.IsArchived, r.IsPrivate,
		r.PushedAt, r.ETag, r.LastModified, r.LastCheckedAt, r.FullName)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

const releaseUpsertSQL = `INSERT INTO releases (repository_id, tag_name, name, body, html_url, tarball_url, zipball_url, published_at, is_prerelease)
 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
 ON CONFLICT(repository_id, tag_name) DO UPDATE SET
   name         = excluded.name,
   body         = excluded.body,
   html_url     = excluded.html_url,
   tarball_url  = excluded.tarball_url,
   zipball_url  = excluded.zipball_url,
   published_at = excluded.published_at,
   is_prerelease = excluded.is_prerelease
 RETURNING id`

// saveReleaseInTx upsert 一条 release 并同步其 assets（先删后插），返回 release id。
func saveReleaseInTx(ctx context.Context, tx *sqlx.Tx, r *release.Release) (int64, error) {
	var id int64
	if err := tx.QueryRowxContext(ctx, releaseUpsertSQL,
		r.RepositoryID, r.TagName, r.Name, r.Body, r.HTMLURL, r.TarballURL, r.ZipballURL, r.PublishedAt, r.IsPrerelease,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("upsert release: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM release_assets WHERE release_id = ?`, id); err != nil {
		return 0, fmt.Errorf("clear assets: %w", err)
	}
	if len(r.Assets) == 0 {
		return id, nil
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO release_assets (release_id, name, size, download_url, content_type) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare asset stmt: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	for _, a := range r.Assets {
		if _, err := stmt.ExecContext(ctx, id, a.Name, a.Size, a.DownloadURL, a.ContentType); err != nil {
			return 0, fmt.Errorf("insert asset: %w", err)
		}
	}
	return id, nil
}

func (d *DB) SaveRelease(ctx context.Context, r *release.Release) error {
	return WithTransaction(ctx, d, func(tx *sqlx.Tx) error {
		_, err := saveReleaseInTx(ctx, tx, r)
		return err
	})
}

func (d *DB) SaveReleasesBatch(ctx context.Context, releases []release.Release) (bool, error) {
	if len(releases) == 0 {
		return false, nil
	}
	var changed bool
	err := WithTransaction(ctx, d, func(tx *sqlx.Tx) error {
		for i := range releases {
			var existing release.Release
			err := tx.GetContext(ctx, &existing,
				`SELECT tag_name, name, body, html_url, published_at, is_prerelease
				 FROM releases WHERE repository_id = ? AND tag_name = ?`,
				releases[i].RepositoryID, releases[i].TagName)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("query release %s: %w", releases[i].TagName, err)
			}
			if err == nil && releaseEqual(existing, releases[i]) {
				continue
			}
			changed = true
			if _, err := saveReleaseInTx(ctx, tx, &releases[i]); err != nil {
				return fmt.Errorf("save release %s: %w", releases[i].TagName, err)
			}
		}
		return nil
	})
	return changed, err
}

func releaseEqual(a, b release.Release) bool {
	return a.Name == b.Name && a.Body == b.Body && a.HTMLURL == b.HTMLURL &&
		a.PublishedAt.Equal(b.PublishedAt) && a.IsPrerelease == b.IsPrerelease
}

// attachAssets 批量加载并挂载 release_assets，避免 N+1。
func (d *DB) attachAssets(ctx context.Context, releases []release.Release) error {
	if len(releases) == 0 {
		return nil
	}
	ids := make([]int64, len(releases))
	indexByID := make(map[int64]int, len(releases))
	for i := range releases {
		ids[i] = releases[i].ID
		indexByID[releases[i].ID] = i
	}
	query, args, err := sqlx.In(
		`SELECT id, release_id, name, size, download_url, content_type, created_at
		 FROM release_assets WHERE release_id IN (?) ORDER BY release_id, name`, ids)
	if err != nil {
		return fmt.Errorf("build assets query: %w", err)
	}
	var assets []release.ReleaseAsset
	if err := d.conn.SelectContext(ctx, &assets, d.conn.Rebind(query), args...); err != nil {
		return fmt.Errorf("select assets: %w", err)
	}
	for _, a := range assets {
		if idx, ok := indexByID[a.ReleaseID]; ok {
			releases[idx].Assets = append(releases[idx].Assets, a)
		}
	}
	return nil
}

type RepoSeed struct {
	Owner string
	Name  string
}

func (d *DB) SeedRepositories(ctx context.Context, seeds []RepoSeed) error {
	if len(seeds) == 0 {
		return nil
	}
	return WithTransaction(ctx, d, func(tx *sqlx.Tx) error {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO repositories (owner, name, full_name) VALUES (?, ?, ?)
			 ON CONFLICT(full_name) DO NOTHING`)
		if err != nil {
			return fmt.Errorf("prepare seed stmt: %w", err)
		}
		defer func() { _ = stmt.Close() }()
		for _, s := range seeds {
			if _, err := stmt.ExecContext(ctx, s.Owner, s.Name, s.Owner+"/"+s.Name); err != nil {
				return fmt.Errorf("exec seed: %w", err)
			}
		}
		return nil
	})
}

func (d *DB) ListReleases(ctx context.Context) ([]release.Release, error) {
	var releases []release.Release
	err := d.conn.SelectContext(ctx, &releases,
		`SELECT id, repository_id, tag_name, name, body, html_url, tarball_url, zipball_url, published_at, is_prerelease, created_at
		 FROM releases ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	if err := d.attachAssets(ctx, releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (d *DB) ListRecentReleases(ctx context.Context, limit int) ([]release.Release, error) {
	var releases []release.Release
	err := d.conn.SelectContext(ctx, &releases,
		`SELECT id, repository_id, tag_name, name, body, html_url, tarball_url, zipball_url, published_at, is_prerelease, created_at
		 FROM releases ORDER BY published_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	if err := d.attachAssets(ctx, releases); err != nil {
		return nil, err
	}
	return releases, nil
}

type repoRow struct {
	ID                int64          `db:"id"`
	Owner             string         `db:"owner"`
	Name              string         `db:"name"`
	FullName          string         `db:"full_name"`
	Description       sql.NullString `db:"description"`
	LogoPath          sql.NullString `db:"logo_path"`
	Stars             int            `db:"stars"`
	Language          sql.NullString `db:"language"`
	IsArchived        int            `db:"is_archived"`
	IsPrivate         int            `db:"is_private"`
	PushedAt          sql.NullTime   `db:"pushed_at"`
	LatestVersion     sql.NullString `db:"latest_version"`
	LatestReleaseURL  sql.NullString `db:"latest_release_url"`
	LatestReleaseDate sql.NullTime   `db:"latest_release_date"`
	ETag              sql.NullString `db:"etag"`
	LastModified      sql.NullString `db:"last_modified"`
	LastCheckedAt     sql.NullTime   `db:"last_checked_at"`
	Remark            sql.NullString `db:"remark"`
	CreatedAt         string         `db:"created_at"`
	UpdatedAt         string         `db:"updated_at"`
	TagNames          sql.NullString `db:"tag_names"`
	TagTypes          sql.NullString `db:"tag_types"`
	TagIDs            sql.NullString `db:"tag_ids"`
}

const repoSelectColumns = `r.id, r.owner, r.name, r.full_name, r.description, r.logo_path,
       r.stars, r.language, r.is_archived, r.is_private, r.pushed_at,
       r.latest_version, r.latest_release_url, r.latest_release_date,
       r.etag, r.last_modified, r.last_checked_at, r.remark, r.created_at, r.updated_at,
       GROUP_CONCAT(t.id ORDER BY t.id) AS tag_ids,
       GROUP_CONCAT(t.name ORDER BY t.id) AS tag_names,
       GROUP_CONCAT(t.type ORDER BY t.id) AS tag_types`

const repoJoins = `FROM repositories r
LEFT JOIN repository_tags rt ON r.id = rt.repository_id
LEFT JOIN tags t ON rt.tag_id = t.id`

func (d *DB) scanRepositories(ctx context.Context, query string, args ...any) ([]release.Repository, error) {
	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []release.Repository
	for rows.Next() {
		var row repoRow
		if err := rows.Scan(
			&row.ID, &row.Owner, &row.Name, &row.FullName,
			&row.Description, &row.LogoPath,
			&row.Stars, &row.Language, &row.IsArchived, &row.IsPrivate, &row.PushedAt,
			&row.LatestVersion, &row.LatestReleaseURL, &row.LatestReleaseDate,
			&row.ETag, &row.LastModified, &row.LastCheckedAt, &row.Remark, &row.CreatedAt, &row.UpdatedAt,
			&row.TagIDs, &row.TagNames, &row.TagTypes,
		); err != nil {
			return nil, err
		}
		result = append(result, row.toRepository())
	}
	return result, rows.Err()
}

func (d *DB) ListRepositories(ctx context.Context) ([]release.Repository, error) {
	return d.scanRepositories(ctx, `SELECT `+repoSelectColumns+`
	`+repoJoins+`
	GROUP BY r.id
	ORDER BY r.full_name`)
}

func (r *repoRow) toRepository() release.Repository {
	repo := release.Repository{
		ID:               r.ID,
		Owner:            r.Owner,
		Name:             r.Name,
		FullName:         r.FullName,
		Description:      r.Description.String,
		LogoPath:         r.LogoPath.String,
		Stars:            r.Stars,
		Language:         r.Language.String,
		IsArchived:       r.IsArchived != 0,
		IsPrivate:        r.IsPrivate != 0,
		LatestVersion:    r.LatestVersion.String,
		LatestReleaseURL: r.LatestReleaseURL.String,
		Remark:           r.Remark.String,
	}
	if t, err := time.Parse("2006-01-02 15:04:05", r.CreatedAt); err == nil {
		repo.CreatedAt = t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", r.UpdatedAt); err == nil {
		repo.UpdatedAt = t
	}
	if r.LatestReleaseDate.Valid {
		repo.LatestReleaseDate = r.LatestReleaseDate.Time
	}
	if r.PushedAt.Valid {
		repo.PushedAt = r.PushedAt.Time
	}
	if r.LastCheckedAt.Valid {
		repo.LastCheckedAt = r.LastCheckedAt.Time
	}
	repo.ETag = r.ETag.String
	repo.LastModified = r.LastModified.String
	if r.TagNames.Valid && r.TagNames.String != "" {
		ids := splitCSV(r.TagIDs.String)
		names := splitCSV(r.TagNames.String)
		types := splitCSV(r.TagTypes.String)
		for i := range names {
			repo.Tags = append(repo.Tags, release.Tag{
				ID:   mustParseInt64(ids[i]),
				Name: names[i],
				Type: types[i],
			})
		}
	}
	return repo
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func mustParseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func (d *DB) GetRepositoryByFullName(ctx context.Context, fullName string) (*release.Repository, error) {
	repos, err := d.scanRepositories(ctx, `SELECT `+repoSelectColumns+`
	`+repoJoins+`
	WHERE r.full_name = ?
	GROUP BY r.id`, fullName)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, nil
	}
	return &repos[0], nil
}

func (d *DB) GetRepositoryByID(ctx context.Context, id int64) (*release.Repository, error) {
	repos, err := d.scanRepositories(ctx, `SELECT `+repoSelectColumns+`
	`+repoJoins+`
	WHERE r.id = ?
	GROUP BY r.id`, id)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 {
		return nil, nil
	}
	return &repos[0], nil
}

func (d *DB) GetReleasesByRepository(ctx context.Context, repoID int64) ([]release.Release, error) {
	var releases []release.Release
	err := d.conn.SelectContext(ctx, &releases,
		`SELECT id, repository_id, tag_name, name, body, html_url, tarball_url, zipball_url, published_at, is_prerelease, created_at
		 FROM releases WHERE repository_id = ? ORDER BY published_at DESC`,
		repoID)
	if err != nil {
		return nil, err
	}
	if err := d.attachAssets(ctx, releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (d *DB) CountReleases(ctx context.Context) (int, error) {
	var count int
	err := d.conn.GetContext(ctx, &count, `SELECT COUNT(*) FROM releases`)
	return count, err
}

func (d *DB) CountRepositories(ctx context.Context) (int, error) {
	var count int
	err := d.conn.GetContext(ctx, &count, `SELECT COUNT(*) FROM repositories`)
	return count, err
}

func (d *DB) ListTags(ctx context.Context) ([]release.Tag, error) {
	var tags []release.Tag
	err := d.conn.SelectContext(ctx, &tags,
		`SELECT id, name, type FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (d *DB) GetOrCreateTag(ctx context.Context, name, tagType string) (*release.Tag, error) {
	t := NormalizeTagType(tagType)
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO tags (name, type) VALUES (?, ?) ON CONFLICT(name) DO NOTHING`,
		name, t)
	if err != nil {
		return nil, err
	}
	var tag release.Tag
	err = d.conn.GetContext(ctx, &tag,
		`SELECT id, name, type FROM tags WHERE name = ?`, name)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) TagRepository(ctx context.Context, repoID, tagID int64) error {
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO repository_tags (repository_id, tag_id) VALUES (?, ?)
		 ON CONFLICT DO NOTHING`,
		repoID, tagID)
	return err
}

func (d *DB) DeleteRepository(ctx context.Context, id int64) error {
	res, err := d.conn.ExecContext(ctx, `DELETE FROM repositories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) UpdateRepository(ctx context.Context, id int64, description, remark string) error {
	res, err := d.conn.ExecContext(ctx,
		`UPDATE repositories SET description = ?, remark = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		description, remark, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) DeleteTag(ctx context.Context, id int64) error {
	res, err := d.conn.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) UpdateTag(ctx context.Context, id int64, name, tagType string) error {
	t := NormalizeTagType(tagType)
	res, err := d.conn.ExecContext(ctx, `UPDATE tags SET name = ?, type = ? WHERE id = ?`, name, t, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) SetRepositoryTagIDs(ctx context.Context, repoID int64, tagIDs []int64) error {
	return WithTransaction(ctx, d, func(tx *sqlx.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM repository_tags WHERE repository_id = ?`, repoID); err != nil {
			return fmt.Errorf("clear tags: %w", err)
		}
		for _, tagID := range tagIDs {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO repository_tags (repository_id, tag_id) VALUES (?, ?)`, repoID, tagID); err != nil {
				return fmt.Errorf("link tag: %w", err)
			}
		}
		return nil
	})
}

func (d *DB) GetUserByUsername(ctx context.Context, username string) (*release.User, error) {
	var u release.User
	err := d.conn.GetContext(ctx, &u,
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`,
		username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *DB) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := d.conn.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`)
	return count, err
}

func (d *DB) CreateUser(ctx context.Context, username, passwordHash, role string) error {
	_, err := d.conn.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, passwordHash, role)
	return err
}

func (d *DB) ListUsers(ctx context.Context) ([]release.User, error) {
	var users []release.User
	err := d.conn.SelectContext(ctx, &users,
		`SELECT id, username, password_hash, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (d *DB) GetUserByID(ctx context.Context, id int64) (*release.User, error) {
	var u release.User
	err := d.conn.GetContext(ctx, &u,
		`SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (d *DB) UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error {
	res, err := d.conn.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) DeleteUser(ctx context.Context, id int64) error {
	res, err := d.conn.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
