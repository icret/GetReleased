# Tracker 扩展性整改方案

> 目标：支撑几百到上千仓库的 Release 追踪，新 release 发现延迟 < 60min，GitHub 配额用量 < 20%，build 链路按需触发。
> 约束：遵守 AGENTS.md（禁止过度设计、禁止提前实现未来功能、依赖最少、复用成熟库）。
> 现状：本地开发阶段，追踪几十个仓库，尚未上线。

---

## 一、现状分析（代码事实）

| # | 现状 | 代码位置 | 问题 |
|---|---|---|---|
| 1 | 每仓库每轮 **2 个 API 请求**（`FetchRepository` + `FetchReleases`） | `github/client.go:22,28` | 1000 仓库 = 2000 req/轮 |
| 2 | **串行追踪**，for 循环顺序执行 | `tracker/tracker.go:36-43` | 1000 仓库 × ~300ms × 2 ≈ 10 分钟/轮 |
| 3 | **无条件请求**，`ListReleases(ctx, owner, repo, nil)` 传 nil | `github/client.go:22` | 没用 ETag/If-Modified-Since，304 本可不计费 |
| 4 | **无增量跳过**，每轮全量拉所有仓库 | `tracker/tracker.go:36-43` | 没用 `pushed_at` 判断仓库是否变化 |
| 5 | **FetchReleases 只拿默认第一页 30 条** | `github/client.go:22` | 超 30 个 release 的仓库丢历史 release |
| 6 | **每轮 track 后立即全量 export** | `cmd/tracker/main.go:79-81` | 无变化也重写 `repositories.json` + `releases.json` |
| 7 | **exporter 全量重写两个大 JSON** | `exporter/exporter.go:13-31` | 1000 仓库时 JSON 几十 MB，每轮全量重写 |
| 8 | **SaveReleasesBatch 无 diff 检测** | `database/queries.go:124-136` | release 没变化也逐条 upsert |
| 9 | **schema 无 ETag/last_modified/pushed_at 字段** | `database/schema.sql:1-23` | 只有 `last_checked_at`，无法支撑条件请求与增量跳过 |
| 10 | **无速率限制感知** | `github/client.go` 全文 | 不读 `X-RateLimit-Remaining`，无退避 |
| 11 | **scheduler 固定间隔无抖动** | `scheduler/scheduler.go:19-29` | 单实例无影响，多实例会同步打 GitHub |
| 12 | **单 token 配置** | `.env:2` 只有 `GITHUB_TOKEN` | 无多 token 轮询/故障转移，配额上限锁死 5000 req/h |

---

## 二、1000 仓库场景下的崩溃点（量化）

假设：1000 仓库，每仓库平均 30 个 release，GitHub token 已配置（5000 req/h），60 min 间隔。

| 维度 | 当前实现 | 1000 仓库表现 | 限额 | 状态 |
|---|---|---|---|---|
| GitHub 配额 | 2000 req/轮 × 1 轮/h | 2000 req/h | 5000 req/h | **占 40%，余量紧** |
| 追踪耗时 | 串行 10 分钟/轮 | 10 min | 60 min 间隔 | 占 1/6，可接受但无并发余量 |
| export 耗时 | 全量重写几十 MB JSON | ~1-2 s | - | 不是瓶颈 |
| Next.js build | `generateStaticParams` 1000 页 | 10-30 min | - | **超过追踪间隔，部署链路阻塞** |
| 无效 IO | 绝大多数轮次无新 release | 全量重写 + 重 build | - | **纯浪费** |

**结论**：1000 仓库 60 min 间隔下，配额占 40% 看似有余量，但考虑突发批量追踪（手动触发、首次全量）、GitHub 二级速率限制（30 req/min）、token 失效等场景，余量并不充裕。追踪耗时和 build 链路是更确定的瓶颈。

---

## 三、整改目标（量化）

| 指标 | 目标 | 当前 |
|---|---|---|
| 支撑仓库数 | ≥ 1000 | ~50 |
| 新 release 发现延迟 | < 60 min | 60 min（但 1000 仓库时追踪耗时 10 min，实际延迟 > 60 min） |
| GitHub 配额占用 | < 20%（多 token 后 < 10%） | 1000 仓库时 40%（单 token） |
| 追踪一轮耗时 | < 2 min | 1000 仓库时 10 min |
| export 触发 | 仅 release 真变化时 | 每轮全量 |
| build 触发 | 仅 JSON 变化时 | 每轮（部署流程） |

---

## 四、整改方案（分阶段）

按收益/成本排序，每阶段独立可交付，逐步推进。

### Phase 1：修 bug + 基础优化（不破坏 schema，0 迁移）

**目标**：修历史 release 丢失 bug，并发追踪降到 2 min，按需 export 消除无效 build，多 token 扩容配额。

| 改动 | 文件 | 内容 |
|---|---|---|
| FetchReleases 设 `PerPage: 100` | `github/client.go:22` | `ListReleases(ctx, owner, repo, &gh.ListOptions{PerPage: 100})`，第一页从 30 扩到 100，对 release 追踪场景够用（新 release 总在前 100）；全量历史需分页遍历，暂不实现 |
| 多 token 配额感知轮询 | `github/client.go` + `.env` | `GITHUB_TOKENS`（逗号分隔，优先于 `GITHUB_TOKEN`），配额感知 round-robin（跳过快耗尽的 token），403/429 冷却切下一个，详见 §五 |
| 并发追踪 | `tracker/tracker.go:36-43` | `errgroup` + semaphore 限并发 8，详见 §五 |
| 按需 export（dirty 标记） | `tracker/tracker.go` + `cmd/tracker/main.go` | track 返回是否有变化，无变化跳过 export，详见 §五 |
| 速率限制感知 | `github/client.go` | 暴露 `RateLimitRemaining()`，tracker 每轮开始前检查，<500 时跳过本轮并告警 |

**收益**：追踪耗时 10 min → 1-2 min，build 频次降 90%+，3 token 配额上限 15000 req/h（1000 仓库 60min 间隔仅用 13%）。
**成本**：~150 行改动，不破坏 schema。

### Phase 2：条件请求（schema 变更，1 次迁移）

**目标**：304 Not Modified 不计费，可放心高频检查，配额占用降 90%+。

| 改动 | 文件 | 内容 |
|---|---|---|
| schema 加 `etag` + `last_modified` | `database/schema.sql` | `repositories` 表加两列，详见 §六 |
| model 加字段 | `release/model.go` | `Repository` 加 `ETag` + `LastModified` |
| client 带 If-None-Match | `github/client.go` | `FetchRepository`/`FetchReleases` 接受 etag 参数，设置 `req.Header`，处理 304 |
| tracker 读写 etag | `tracker/tracker.go` | 请求前读 DB 的 etag，请求后写回 |
| 间隔可降 | `.env` | `TRACK_INTERVAL=30m`（304 不计费，高频检查安全） |

**收益**：1000 仓库每轮只有 ~5-20 个真变化仓库扣配额，配额占用从 40% 降到 <3%。
**成本**：~200 行改动 + 1 次 schema 迁移。

### Phase 3：增量跳过（schema 变更，1 次迁移）

**目标**：仓库没 push 就跳过 FetchReleases，省一半请求 + 减少无效写。

| 改动 | 文件 | 内容 |
|---|---|---|
| schema 加 `pushed_at` | `database/schema.sql` | `repositories` 表加 `pushed_at DATETIME` |
| tracker 跳过逻辑 | `tracker/tracker.go` | `FetchRepository` 返回 `pushed_at`，若 ≤ `last_checked_at` 则跳过 `FetchReleases`，只更新元数据 |
| SaveReleasesBatch 加 diff | `database/queries.go:124-136` | upsert 前比较 body/html_url/published_at，无变化跳过 |

**收益**：API 请求再减一半，SQLite 写入减少。
**成本**：~150 行改动 + 1 次 schema 迁移。

### Phase 4：架构解耦（可选，规模到 2000+ 再考虑）

**目标**：tracker 只写 SQLite，export + build 由独立信号触发。

当前 `cmd/tracker/main.go` 把 track + export 绑在一个 task 里。规模更大时可拆：
- tracker 进程：只追踪 + 写 SQLite，写完 touch 一个 `data/.dirty` 标记文件。
- exporter + build：由 systemd timer 或 CI 检测 `.dirty` 存在时触发 `deploy.sh` 的 export+build 段，完成后删除 `.dirty`。

**本方案暂不实施**，Phase 1-3 已足够支撑 1000 仓库。符合 AGENTS.md "禁止过度设计"。

---

## 五、详细代码改动清单

### 5.1 Phase 1 改动

#### `.env` / `.env.example` — 多 token 配置

```bash
# GitHub API Token（单个，向后兼容）
GITHUB_TOKEN=

# GitHub API Tokens（多个，逗号分隔，优先于 GITHUB_TOKEN）
# 多 token 配额感知轮询（跳过快耗尽的 token）+ 403/429 冷却切下一个
# 3 token × 5000 req/h = 15000 req/h，1000 仓库 60min 间隔仅用 13%
GITHUB_TOKENS=
```

> 加载优先级：`GITHUB_TOKENS` 非空则用 `GITHUB_TOKENS`，否则回退 `GITHUB_TOKEN`（向后兼容）。

#### `backend/internal/github/client.go` — 多 token 配额感知轮询 + 分页 + 速率限制

```go
package github

import (
    "context"
    "errors"
    "log/slog"
    "sync/atomic"
    "time"

    gh "github.com/google/go-github/v69/github"
)

const rateLimitThreshold = 100 // 剩余配额低于此值则跳过该 token

type tokenState struct {
    client    *gh.Client
    remaining atomic.Int64 // 从 GitHub 响应头实时更新
    resetAt   atomic.Int64 // Unix 秒，配额恢复时刻；冷却期内不选中
}

type Client struct {
    tokens  []*tokenState
    counter atomic.Int64 // round-robin 计数器
}

// NewClient 接受多 token，空切片则用未认证客户端（受 60 req/h 限制）。
// 单 token 时 pool 退化为单元素，所有逻辑正常工作。
func NewClient(tokens []string) *Client {
    if len(tokens) == 0 {
        tokens = []string{""}
    }
    pool := make([]*tokenState, len(tokens))
    for i, t := range tokens {
        c := gh.NewClient(nil)
        if t != "" {
            c = c.WithAuthToken(t)
        }
        pool[i] = &tokenState{client: c}
        pool[i].remaining.Store(5000) // 初始假设满配额
    }
    return &Client{tokens: pool}
}

// nextClient 配额感知轮询：round-robin 跳过 remaining < 阈值 或冷却中的 token。
func (c *Client) nextClient() (*tokenState, error) {
    now := time.Now().Unix()
    for i := 0; i < len(c.tokens); i++ {
        idx := int(c.counter.Add(1)) % len(c.tokens)
        t := c.tokens[idx]
        // 冷却到期，重置配额估计
        if resetAt := t.resetAt.Load(); resetAt > 0 && now > resetAt {
            t.remaining.Store(5000)
            t.resetAt.Store(0)
        }
        if t.remaining.Load() > rateLimitThreshold {
            return t, nil
        }
    }
    return nil, errors.New("all tokens exhausted or cooling down")
}

// updateRemaining 从 GitHub 响应头更新 token 配额状态。
// go-github 自动解析 X-RateLimit-Remaining / X-RateLimit-Reset 到 resp.Rate。
func (c *Client) updateRemaining(t *tokenState, resp *gh.Response) {
    if resp != nil && resp.Rate.Remaining > 0 {
        t.remaining.Store(int64(resp.Rate.Remaining))
        t.resetAt.Store(resp.Rate.Reset.Unix())
    }
}

// markCooldown 403/429 后标记 token 冷却到 reset 时刻。
func (c *Client) markCooldown(t *tokenState, resp *gh.Response) {
    if resp != nil && resp.Rate.Reset.Unix() > 0 {
        t.remaining.Store(0)
        t.resetAt.Store(resp.Rate.Reset.Unix())
    } else {
        t.remaining.Store(0) // 响应头无 reset 信息，保守冷却 1 分钟
        t.resetAt.Store(time.Now().Add(time.Minute).Unix())
    }
}

func (c *Client) FetchReleases(ctx context.Context, owner, repo string) ([]*gh.RepositoryRelease, error) {
    var lastErr error
    for i := 0; i < len(c.tokens); i++ {
        t, err := c.nextClient()
        if err != nil {
            return nil, lastErr
        }
        releases, resp, err := t.client.Repositories.ListReleases(ctx, owner, repo, &gh.ListOptions{PerPage: 100})
        if err == nil {
            c.updateRemaining(t, resp)
            return releases, nil
        }
        lastErr = err
        if resp != nil && (resp.StatusCode == 403 || resp.StatusCode == 429) {
            c.markCooldown(t, resp)
            slog.WarnContext(ctx, "token rate limited, cooldown and try next", "repo", owner+"/"+repo, "status", resp.StatusCode, "reset_at", time.Unix(t.resetAt.Load(), 0))
            continue
        }
        c.updateRemaining(t, resp)
        return nil, err // 非速率限制错误不重试
    }
    return nil, lastErr
}

func (c *Client) FetchRepository(ctx context.Context, owner, repo string) (*gh.Repository, error) {
    // 同 FetchReleases 的配额感知轮询 + 故障转移模式
    // ... 详见 FetchReleases ...
}

// RateLimitRemaining 返回所有未冷却 token 的剩余配额总和。
// 单 token 时返回该 token 的 remaining；多 token 时返回总和，避免单 token 耗尽误判整轮跳过。
func (c *Client) RateLimitRemaining(ctx context.Context) (int, error) {
    now := time.Now().Unix()
    total := 0
    queried := 0
    for _, t := range c.tokens {
        // 冷却到期，重置
        if resetAt := t.resetAt.Load(); resetAt > 0 && now > resetAt {
            t.remaining.Store(5000)
            t.resetAt.Store(0)
        }
        // 优先用响应头缓存的 remaining（零额外请求）
        if r := t.remaining.Load(); r > 0 {
            total += int(r)
            queried++
            continue
        }
        // 缓存为空才查 API
        rl, _, err := t.client.RateLimits(ctx)
        if err != nil {
            continue
        }
        remaining := int64(rl.GetCore().GetRemaining())
        t.remaining.Store(remaining)
        t.resetAt.Store(rl.GetCore().GetReset().Unix())
        total += int(remaining)
        queried++
    }
    if queried == 0 {
        return 0, errors.New("all token rate limit queries failed")
    }
    return total, nil
}
```

> `Fetcher` 接口同步加 `RateLimitRemaining` 方法。
> `cmd/tracker/main.go` 加载 token：优先 `GITHUB_TOKENS`（逗号分隔），回退 `GITHUB_TOKEN`。

#### `backend/internal/tracker/tracker.go` — 并发 + dirty 标记

```go
import "golang.org/x/sync/errgroup"

func (t *Tracker) Track(ctx context.Context) (bool, error) {
    repos, err := t.store.ListRepositories(ctx)
    if err != nil {
        return false, err
    }

    var dirty atomic.Bool
    g, gctx := errgroup.WithContext(ctx)
    g.SetLimit(8) // 并发上限

    for _, repo := range repos {
        repo := repo
        g.Go(func() error {
            if gctx.Err() != nil {
                return gctx.Err()
            }
            changed, err := t.trackOne(gctx, repo.Owner, repo.Name)
            if err != nil {
                slog.ErrorContext(gctx, "track repo", "repo", repo.FullName, "err", err)
                return nil // 单仓库失败不阻断整轮
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

// trackOne 返回 (changed, error)，changed=true 表示 release 有变化
func (t *Tracker) trackOne(ctx context.Context, owner, name string) (bool, error) {
    // ... 现有逻辑 ...
    // changed = 比较新旧 release 列表的 tag_name + published_at 集合是否不同
    // 详见 SaveReleasesBatch 的 diff 逻辑（Phase 3）
}
```

> 新增依赖：`golang.org/x/sync`（errgroup）。这是 Go 官方扩展库，符合"复用成熟库"。
> 多 token 用 `sync/atomic` 标准库，无新依赖。

#### `backend/cmd/tracker/main.go` — 按需 export + token 加载

```go
// token 加载
var tokens []string
if v := os.Getenv("GITHUB_TOKENS"); v != "" {
    tokens = strings.Split(v, ",")
} else if v := os.Getenv("GITHUB_TOKEN"); v != "" {
    tokens = []string{v}
}
ghClient := github.NewClient(tokens)

task := func(ctx context.Context) {
    // 速率限制检查
    if remaining, _ := ghClient.RateLimitRemaining(ctx); remaining < 500 {
        slog.WarnContext(ctx, "rate limit low, skip track", "remaining", remaining)
        return
    }

    dirty, err := trk.Track(ctx)
    if err != nil {
        slog.ErrorContext(ctx, "track", "err", err)
    }
    if !dirty {
        slog.InfoContext(ctx, "track done, no change, skip export")
        return
    }
    if err := trk.Export(ctx, exportDir); err != nil {
        slog.ErrorContext(ctx, "export", "err", err)
    }
}
```

#### `backend/internal/tracker/tracker.go` — `TrackOne` 同步改签名

`TrackOne` 也要返回 `(bool, error)`。

#### `backend/internal/api/handler_actions.go` — server 调用点同步改

server 的 `handleTrack`（`handler_actions.go:46`）调用 `a.trk.Track(ctx)`，`handleSync` 调用 `a.trk.TrackOne(...)`，两处都要适配新签名：

```go
// handleTrack (handler_actions.go:44-55)
go func() {
    ctx := context.WithoutCancel(r.Context())
    dirty, err := a.trk.Track(ctx)
    a.trackMu.Lock()
    a.trackState.running = false
    a.trackState.finishedAt = time.Now()
    a.trackState.err = ""
    if err != nil {
        a.trackState.err = err.Error()
    }
    a.trackMu.Unlock()
    // dirty 信息可记入 trackState 供 status 端点返回
}()
```

> server 的 `handleTrack` 只追踪不导出（AGENTS.md API 表），dirty 返回值可用于 status 展示，不触发 export。

### 5.2 Phase 2 改动

#### `backend/internal/database/schema.sql` — 加条件请求字段

```sql
-- 在 repositories 表内追加（项目未上线，直接改 schema.sql 重建 DB，见 §六）
    etag          TEXT,
    last_modified TEXT,
```

#### `backend/internal/release/model.go` — Repository 加字段

```go
type Repository struct {
    // ... 现有字段 ...
    ETag         string `json:"-" db:"etag"`
    LastModified string `json:"-" db:"last_modified"`
}
```

> `json:"-"` 不导出到前端 JSON，这是追踪内部缓存字段。

#### `backend/internal/github/client.go` — 条件请求

```go
type FetchResult struct {
    Repo      *gh.Repository
    ETag      string
    Modified  bool // false 表示 304 Not Modified
}

// FetchRepository 在 Phase 1 配额感知轮询基础上，加条件请求头 + 304 处理。
// 与 FetchReleases 的差异仅在请求头和 304 解析，故障转移循环相同。
func (c *Client) FetchRepository(ctx context.Context, owner, repo, etag, lastModified string) (*FetchResult, error) {
    var lastErr error
    for i := 0; i < len(c.tokens); i++ {
        t, err := c.nextClient()
        if err != nil {
            return nil, lastErr
        }
        req, err := t.client.NewRequest("GET", fmt.Sprintf("repos/%s/%s", owner, repo), nil)
        if err != nil {
            return nil, err
        }
        if etag != "" {
            req.Header.Set("If-None-Match", etag)
        }
        if lastModified != "" {
            req.Header.Set("If-Modified-Since", lastModified)
        }

        resp, err := t.client.Do(ctx, req)
        if err != nil {
            return nil, err
        }

        if resp.StatusCode == 403 || resp.StatusCode == 429 {
            c.markCooldown(t, resp)
            lastErr = fmt.Errorf("rate limited: %d", resp.StatusCode)
            resp.Body.Close()
            continue
        }

        result := &FetchResult{
            ETag:     resp.Header.Get("ETag"),
            Modified: resp.StatusCode != http.StatusNotModified,
        }
        if result.Modified {
            if err := json.NewDecoder(resp.Body).Decode(&result.Repo); err != nil {
                resp.Body.Close()
                return nil, err
            }
        }
        resp.Body.Close()
        c.updateRemaining(t, resp)
        return result, nil
    }
    return nil, lastErr
}
```

> `FetchReleases` 同理。`Fetcher` 接口同步更新。
> 304 时 `result.Repo == nil`，tracker 跳过该仓库的后续处理。

#### `backend/internal/tracker/tracker.go` — 读写 etag

```go
func (t *Tracker) trackOne(ctx context.Context, owner, name string) (bool, error) {
    repo, err := t.store.GetRepository(ctx, owner, name) // 新增：读现有 etag
    if err != nil && !errors.Is(err, ErrNotFound) {
        return false, err
    }

    result, err := t.ghClient.FetchRepository(ctx, owner, name, repo.ETag, repo.LastModified)
    if err != nil {
        return false, err
    }
    if !result.Modified {
        return false, nil // 304，无变化
    }
    // ... 现有处理逻辑，SaveRepository 时写入 result.ETag + result.LastModified ...
}
```

### 5.3 Phase 3 改动

#### `backend/internal/database/schema.sql` — 加 pushed_at

```sql
    pushed_at     DATETIME,
```

#### `backend/internal/tracker/tracker.go` — pushed_at 跳过

```go
func (t *Tracker) trackOne(ctx context.Context, owner, name string) (bool, error) {
    existing, err := t.store.GetRepository(ctx, owner, name)
    // ...

    result, err := t.ghClient.FetchRepository(ctx, owner, name, existing.ETag, existing.LastModified)
    // ...
    if !result.Modified {
        return false, nil
    }

    pushedAt := result.Repo.GetPushedAt().Time
    // 仓库自上次追踪后没 push，release 不会变，跳过 FetchReleases
    if !existing.LastCheckedAt.IsZero() && !pushedAt.After(existing.LastCheckedAt) {
        if err := t.store.UpdateRepoMeta(ctx, repo); err != nil { // 只更新 stars/etag/pushed_at
            return false, err
        }
        return false, nil
    }
    // ... 继续 FetchReleases ...
}
```

#### `backend/internal/database/queries.go` — SaveReleasesBatch 加 diff

```go
func (d *DB) SaveReleasesBatch(ctx context.Context, releases []release.Release) (bool, error) {
    if len(releases) == 0 {
        return false, nil
    }
    var changed bool
    err := WithTransaction(ctx, d, func(tx *sqlx.Tx) error {
        for i := range releases {
            // 先查现有记录
            var existing release.Release
            err := tx.GetContext(ctx, &existing,
                `SELECT tag_name, name, body, html_url, published_at, is_prerelease
                 FROM releases WHERE repository_id = ? AND tag_name = ?`,
                releases[i].RepositoryID, releases[i].TagName)
            if err == nil && releaseEqual(existing, releases[i]) {
                continue // 无变化跳过
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
```

> `RepositoryStore` 接口的 `SaveReleasesBatch` 返回值从 `error` 改为 `(bool, error)`。

### 5.4 一次做 1+2+3 时的合并要点

**`trackOne` 最终形态**（Phase 1+2+3 合并，三段的 `// ... 现有逻辑 ...` 合为此版）：

```go
func (t *Tracker) trackOne(ctx context.Context, owner, name string) (bool, error) {
    // Phase 2: 读现有 etag
    existing, err := t.store.GetRepository(ctx, owner, name)
    if err != nil && !errors.Is(err, database.ErrNotFound) {
        return false, err
    }

    // Phase 2: 条件请求
    result, err := t.ghClient.FetchRepository(ctx, owner, name, existing.ETag, existing.LastModified)
    if err != nil {
        return false, err
    }
    if !result.Modified {
        return false, nil // 304，无变化
    }

    // Phase 3: pushed_at 跳过
    pushedAt := result.Repo.GetPushedAt().Time
    if !existing.LastCheckedAt.IsZero() && !pushedAt.After(existing.LastCheckedAt) {
        if err := t.store.UpdateRepoMeta(ctx, /* stars/etag/pushed_at/last_checked_at */); err != nil {
            return false, err
        }
        return false, nil
    }

    // Phase 1: 现有逻辑（FetchReleases + SaveRepository + SaveReleasesBatch）
    releases, err := t.ghClient.FetchReleases(ctx, owner, name)
    if err != nil {
        return false, err
    }
    repoID, err := t.store.SaveRepository(ctx, /* 含 result.ETag + result.LastModified + pushedAt */)
    if err != nil {
        return false, err
    }
    batch := toReleaseBatch(repoID, releases)

    // Phase 3: changed 来自 SaveReleasesBatch 的 diff 返回值，不在此处自己比较
    changed, err := t.store.SaveReleasesBatch(ctx, batch)
    if err != nil {
        return false, err
    }
    return changed, nil
}
```

**需新增到 `RepositoryStore` 接口的方法**（现有接口只有 `ListRepositories`/`SaveRepository`/`SaveReleasesBatch`）：

| 方法 | 用途 | Phase |
|---|---|---|
| `GetRepository(ctx, owner, name) (*release.Repository, error)` | 读单个仓库（含 etag/last_modified/pushed_at/last_checked_at） | 2 |
| `UpdateRepoMeta(ctx, ...) error` | 只更新元数据（stars/etag/pushed_at/last_checked_at），不触碰 release | 3 |

> `database.ErrNotFound`：若现有代码无此哨兵错误，新增 `var ErrNotFound = errors.New("not found")` 到 `database` 包，`GetRepository` 查无记录时返回它。

**schema 以 §六完整版为准**：§5.2/§5.3 的 schema 片段仅说明新增字段，实际改动只改 §六的 `schema.sql` 一次。

---

## 六、schema 迁移

Phase 2 + Phase 3 合并为一次迁移（全新项目，未上线，直接改 schema.sql + 重建 DB）。

### `backend/internal/database/schema.sql` — repositories 表最终形态

```sql
CREATE TABLE IF NOT EXISTS repositories (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    owner                TEXT NOT NULL,
    name                 TEXT NOT NULL,
    full_name            TEXT NOT NULL UNIQUE,
    description          TEXT,
    logo_path            TEXT,
    -- GitHub 元数据
    stars                INTEGER NOT NULL DEFAULT 0,
    language             TEXT,
    is_archived          INTEGER NOT NULL DEFAULT 0,
    is_private           INTEGER NOT NULL DEFAULT 0,
    pushed_at            DATETIME,                    -- 新增：GitHub 仓库 pushed_at，用于增量跳过
    -- 最新 release 缓存
    latest_version       TEXT,
    latest_release_url   TEXT,
    latest_release_date  DATETIME,
    -- 条件请求缓存
    etag                 TEXT,                        -- 新增：HTTP ETag
    last_modified        TEXT,                        -- 新增：HTTP Last-Modified
    -- 运维/审计
    last_checked_at      DATETIME,
    remark               TEXT,
    created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

> 新增 3 个字段：`pushed_at`、`etag`、`last_modified`。均允许 NULL，首次追踪时为空。
> 因项目未上线，无需写 ALTER TABLE 迁移脚本，直接改 schema.sql 重建 DB 即可。

---

## 七、新增依赖

| 依赖 | 用途 | Phase | 是否符合 AGENTS.md |
|---|---|---|---|
| `golang.org/x/sync` | errgroup 并发控制 | 1 | ✓ Go 官方扩展库，成熟稳定 |

无其他新增依赖。ETag/条件请求用 `net/http` 标准库，多 token 轮询用 `sync/atomic` 标准库，不引入第三方。

---

## 八、验证方式

### 8.1 单元测试

| 模块 | 测试 | 文件 |
|---|---|---|
| tracker | 并发追踪不丢仓库、dirty 标记正确 | `tracker/tracker_test.go` |
| github client | 多 token 配额感知轮询、403/429 冷却转移、304 处理、ETag 回传、分页 | `github/client_test.go`（新增） |
| queries | SaveReleasesBatch diff 检测、pushed_at 跳过 | `database/queries_test.go` |

### 8.2 集成验证

1. **50 仓库回归**：现有 `repositories.json` seed，跑一轮追踪，确认 release 数据与改造前一致。
2. **1000 仓库压测**：mock GitHub API（用 `httptest` 返回固定 release），验证：
   - 追踪一轮 < 2 min
   - 无变化轮次不触发 export
   - 配额检查正确跳过
   - 多 token 配额感知轮询跳过低配额 token、403 冷却转移生效
3. **配额监控**：tracker 日志输出每轮 `rate_limit_remaining`（多 token 时输出总和），确认 <500 时跳过。

### 8.3 构建验证

```bash
cd backend
go build ./cmd/tracker && go build ./cmd/server
go test ./...
golangci-lint run
gofmt -l . && goimports -l .
```

---

## 九、不做的事（遵守 AGENTS.md "禁止过度设计"）

| 项 | 不做理由 |
|---|---|
| 消息队列（NATS/Redis Stream） | 1000 仓库规模 errgroup + semaphore 足够，引入 MQ 是过度设计 |
| Redis 缓存层 | SQLite + ETag 已是缓存，再加 Redis 多一层数据源，违反"SQLite 是唯一核心数据源" |
| 分片追踪（多 tracker 实例） | 1000 仓库单实例 + 并发 8 足够，10000+ 再考虑 |
| 改 SQLite 为 Postgres | SQLite WAL 模式读写并发足够，AGENTS.md 锁定 SQLite |
| 增量 export（分片 JSON） | 全量重写两个 JSON 文件在 1000 仓库时仅 ~1-2s，不是瓶颈；分片会改前端读取逻辑，规模不匹配 |
| GraphQL API | REST + ETag 已满足，GraphQL 复杂度更高且 go-github REST 是官方首选 |
| Webhook 替代轮询 | 需要公网回调地址，本地开发阶段不具备；轮询 + ETag 在 1000 仓库规模足够 |
| FetchReleases 全量分页遍历 | release 追踪只关心新 release（总在前 100 条），全量历史对当前功能无价值；超 100 个 release 的仓库罕见，需要时再加 |

---

## 十、实施顺序与交付物

| 阶段 | 交付物 | 验收标准 |
|---|---|---|
| Phase 1 | 分页修复 + 多 token + 并发 + 按需 export + 配额检查 | 50 仓库追踪 < 10s，无变化轮次不 export，3 token 配额分摊均匀 |
| Phase 2 | ETag 条件请求 + schema 迁移 | 1000 仓库配额占用 < 3%，304 命中率 > 90% |
| Phase 3 | pushed_at 跳过 + release diff | API 请求再减一半，SQLite 写入减少 |

Phase 1 可立即开始，不破坏 schema。Phase 2 + 3 合并一次 schema 变更（项目未上线，直接改 schema.sql 重建 DB）。

Phase 4（架构解耦）待规模到 2000+ 仓库且 build 链路成为瓶颈时再评估。

