-- 仓库主表
CREATE TABLE IF NOT EXISTS repositories (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    owner                TEXT NOT NULL,
    name                 TEXT NOT NULL,
    full_name            TEXT NOT NULL UNIQUE,
    description          TEXT,
    logo_path            TEXT,
    -- GitHub 元数据（追踪时刷新）
    stars                INTEGER NOT NULL DEFAULT 0,
    language             TEXT,
    is_archived          INTEGER NOT NULL DEFAULT 0,
    is_private           INTEGER NOT NULL DEFAULT 0,
    pushed_at            DATETIME,
    -- 最新 release 缓存（追踪时刷新，避免列表页 JOIN releases）
    latest_version       TEXT,
    latest_release_url   TEXT,
    latest_release_date  DATETIME,
    -- 条件请求缓存（追踪内部，不导出到前端 JSON）
    etag                 TEXT,
    last_modified        TEXT,
    -- 运维/审计
    last_checked_at      DATETIME,
    remark               TEXT,
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 标签
CREATE TABLE IF NOT EXISTS tags (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'category'
);

-- 仓库-标签多对多
CREATE TABLE IF NOT EXISTS repository_tags (
    repository_id INTEGER NOT NULL,
    tag_id        INTEGER NOT NULL,
    PRIMARY KEY (repository_id, tag_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id)        REFERENCES tags(id)        ON DELETE CASCADE
);

-- release 记录
CREATE TABLE IF NOT EXISTS releases (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    tag_name      TEXT NOT NULL,
    name          TEXT,
    body          TEXT,
    html_url      TEXT,
    tarball_url   TEXT NOT NULL DEFAULT '',
    zipball_url   TEXT NOT NULL DEFAULT '',
    published_at  DATETIME,
    is_prerelease INTEGER NOT NULL DEFAULT 0,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    UNIQUE(repository_id, tag_name)
);

-- release 附加资产（二进制下载件，如 .deb/.exe/.dmg）
CREATE TABLE IF NOT EXISTS release_assets (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    release_id    INTEGER NOT NULL,
    name          TEXT NOT NULL,
    size          INTEGER NOT NULL DEFAULT 0,
    download_url  TEXT NOT NULL,
    content_type  TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (release_id) REFERENCES releases(id) ON DELETE CASCADE
);

-- 管理员账号
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'admin',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_releases_published_at  ON releases(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_releases_repo_date     ON releases(repository_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_repository_tags_tag_id ON repository_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_repositories_language  ON repositories(language);
CREATE INDEX IF NOT EXISTS idx_repositories_stars     ON repositories(stars DESC);
CREATE INDEX IF NOT EXISTS idx_release_assets_release_id ON release_assets(release_id);
