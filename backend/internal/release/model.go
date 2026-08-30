package release

import "time"

type Repository struct {
	ID                 int64     `json:"id" db:"id"`
	Owner              string    `json:"owner" db:"owner"`
	Name               string    `json:"name" db:"name"`
	FullName           string    `json:"full_name" db:"full_name"`
	Description        string    `json:"description" db:"description"`
	LogoPath           string    `json:"logo_path" db:"logo_path"`
	Stars              int       `json:"stars" db:"stars"`
	Language           string    `json:"language" db:"language"`
	IsArchived         bool      `json:"is_archived" db:"is_archived"`
	IsPrivate          bool      `json:"is_private" db:"is_private"`
	PushedAt           time.Time `json:"pushed_at" db:"pushed_at"`
	LatestVersion      string    `json:"latest_version" db:"latest_version"`
	LatestReleaseURL   string    `json:"latest_release_url" db:"latest_release_url"`
	LatestReleaseDate  time.Time `json:"latest_release_date" db:"latest_release_date"`
	ETag               string    `json:"-" db:"etag"`
	LastModified       string    `json:"-" db:"last_modified"`
	LastCheckedAt      time.Time `json:"last_checked_at" db:"last_checked_at"`
	Remark             string    `json:"remark" db:"remark"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	Tags               []Tag     `json:"tags,omitempty"`
	ReleaseCount       int       `json:"release_count" db:"-"`
	LatestIsPrerelease bool      `json:"latest_is_prerelease,omitempty" db:"-"`
}

type Tag struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Type string `json:"type" db:"type"`
}

type Release struct {
	ID           int64          `json:"id" db:"id"`
	RepositoryID int64          `json:"repository_id" db:"repository_id"`
	TagName      string         `json:"tag_name" db:"tag_name"`
	Name         string         `json:"name" db:"name"`
	Body         string         `json:"body" db:"body"`
	HTMLURL      string         `json:"html_url" db:"html_url"`
	TarballURL   string         `json:"tarball_url" db:"tarball_url"`
	ZipballURL   string         `json:"zipball_url" db:"zipball_url"`
	PublishedAt  time.Time      `json:"published_at" db:"published_at"`
	IsPrerelease bool           `json:"is_prerelease" db:"is_prerelease"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	Assets       []ReleaseAsset `json:"assets,omitempty" db:"-"`
}

// ReleaseAsset release 附加资产（二进制下载件）。
type ReleaseAsset struct {
	ID          int64     `json:"id" db:"id"`
	ReleaseID   int64     `json:"release_id" db:"release_id"`
	Name        string    `json:"name" db:"name"`
	Size        int64     `json:"size" db:"size"`
	DownloadURL string    `json:"download_url" db:"download_url"`
	ContentType string    `json:"content_type" db:"content_type"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// User 管理员账号
type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
