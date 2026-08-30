# 项目目标

开发一个用于追踪 GitHub 等开源软件 Release 的 Web 平台。

## 核心功能

以 GitHub 为主要数据源：

1. 管理需要追踪的 GitHub Repository，支持标签（多对多）与按标签筛选。
2. Go Tracker 定时通过 GitHub API 获取 Release 数据。
3. 数据保存到 SQLite。
4. Go 将 SQLite 导出为前端可直接读取的 JSON。
5. React + TypeScript 前端展示项目、Release、版本、发布时间、Release Notes。
6. Nginx 提供前端静态文件与 JSON。
7. 管理员后台：密码登录、仓库/标签 CRUD、手动触发追踪，Go Server 提供 Admin API。

### 数据模型

权威定义见 `backend/internal/database/schema.sql`，共 6 张表：

| 表 | 用途 | 关键字段 |
|---|---|---|
| `repositories` | 仓库主表 | `owner`+`name`（唯一 `full_name`）、`stars`、`language`、`pushed_at`、`latest_version`/`latest_release_date`（缓存）、`etag`/`last_modified`（条件请求缓存，`json:"-"` 不导出） |
| `tags` | 标签 | `name`（唯一）、`type` |
| `repository_tags` | 仓库-标签多对多 | 联合主键 `(repository_id, tag_id)`，级联删除 |
| `releases` | release 记录 | `repository_id`+`tag_name`（唯一），`body`、`html_url`、`tarball_url`/`zipball_url`、`published_at`、`is_prerelease` |
| `release_assets` | release 附加资产（下载件元数据，不存二进制） | `release_id`（级联删除）、`name`、`size`、`download_url`、`content_type` |
| `users` | 管理员账号 | `username`（唯一）、`password_hash`（bcrypt）、`role` |

Go 结构体映射见 `backend/internal/release/model.go`（`Repository` / `Tag` / `Release` / `ReleaseAsset` / `User`）。

### Admin API 端点

权威定义见 `backend/internal/api/router.go`，JWT 鉴权（`/api/admin/*`）：

| 方法 | 路径 | 用途 |
|---|---|---|
| POST | `/api/login` | 登录获取 JWT |
| GET / POST | `/api/admin/repositories` | 列出 / 新增仓库 |
| PUT / DELETE | `/api/admin/repositories/{id}` | 更新 / 删除仓库 |
| PUT | `/api/admin/repositories/{id}/tags` | 设置仓库标签 |
| POST | `/api/admin/repositories/{id}/sync` | 单仓库同步 |
| GET / POST | `/api/admin/tags` | 列出 / 新增标签 |
| PUT / DELETE | `/api/admin/tags/{id}` | 更新 / 删除标签 |
| POST | `/api/admin/track` | 手动触发全量追踪（仅追踪，不导出） |
| POST | `/api/admin/export` | 导出 JSON（仅导出，不追踪） |
| GET | `/api/admin/track/status` | 追踪状态 |
| GET | `/api/admin/stats` | 统计 |
| GET / POST | `/api/admin/users` | 列出 / 新增用户 |
| PUT | `/api/admin/users/{id}/password` | 重置密码 |
| DELETE | `/api/admin/users/{id}` | 删除用户 |

### 前端路由

权威定义见 `frontend/src/app/`，Next.js App Router 静态导出：

| 路由 | 渲染方式 | 用途 |
|---|---|---|
| `/` | SSG | 首页（仓库列表） |
| `/repository/[owner]/[name]` | SSG（`generateStaticParams` 逐仓库预渲染） | 仓库详情 + Release 列表 |
| `/login` | SSG | 管理员登录 |
| `/admin` | SSG（客户端 fetch） | 管理后台首页 |
| `/admin/repositories` | SSG（客户端 fetch） | 仓库管理 |
| `/admin/tags` | SSG（客户端 fetch） | 标签管理 |
| `/admin/users` | SSG（客户端 fetch） | 用户管理 |

## 后续扩展（禁止提前实现）

* GitLab / Gitea 等其他代码托管平台
* 用户中心、REST API、通知功能

## 核心设计原则

* 一律 AI 驱动开发。
* 必须优先复用成熟、稳定、经过验证的第三方库，禁止重复造轮子。
* 必须保持代码简单、模块清晰、依赖最少。
* 禁止过度设计，禁止提前实现未来功能。
* Go 负责数据追踪、处理与管理 API；React 必须只负责 UI 展示。
* SQLite 是唯一核心数据源；JSON 是前端唯一公开读取的数据格式。

### AI 工作模式

* 修改文件、目录、模块边界前，先读取 `docs/project_structure.md` 与相邻文件，遵循现有约定。
* 修改业务逻辑前向用户确认；纯格式化/补注释/修 typo 可直接执行。
* 不主动 `git commit` / `git push`，除非用户明确要求。
* 新增依赖前检查"暂缓启用"清单，确认不属于未到时机的功能。
* 优先复用 `frontend/src/lib/`、`frontend/src/components/`、`backend/internal/tracker/` 中已有工具，禁止重复造轮子。
* 完成代码后必须运行构建/测试/lint（见下文"构建与验证命令"），不交付未验证代码。

## 技术栈

**一律使用最新稳定版本。具体版本以以下文件为权威来源（项目初始化后由 AI 维护）：**

* Go — `backend/go.mod`
* Node.js / 前端依赖 — `frontend/package.json`（`engines`、`packageManager` 字段）

### 后端

* **Go**
* **SQLite**
* **modernc.org/sqlite** — SQLite Driver
* **jmoiron/sqlx** — SQLite 查询封装（结构体扫描、事务、批量写入）
* **google/go-github** — GitHub API
* **Masterminds/semver** — 语义版本比较
* **golang-jwt/jwt/v5** — JWT 签发/校验（admin 鉴权）
* **golang.org/x/crypto** — bcrypt（管理员密码哈希）
* **github.com/google/uuid** — task_id 与 requestID 生成
* **golang.org/x/sync** — errgroup 并发追踪（`errgroup.SetLimit` 限并发 8）

### 前端

* **pnpm** — 包管理器（**一律使用 pnpm，禁止 npm / yarn**）
* **Next.js** — App Router 静态导出（`output: 'export'`），构建期逐仓库预渲染静态 HTML
* **React**
* **TypeScript**
* **Tailwind CSS**
* **shadcn/ui** — UI 组件（底层 `@base-ui/react`，见 [Shadcn][1]）
* **Lucide React** — 图标
* **react-markdown** — Release Notes 渲染
* **remark-gfm** — GitHub Flavored Markdown

### 数据层

* **公开页面** — 构建期 `fs` 读取 JSON（SSG 数据注入 `public/data/*.json`），无运行时 `fetch`
* **管理后台** — 运行时 `fetch` 调用 Go Server Admin API（JWT 鉴权）

### Web / 部署

* **Nginx**
* **systemd**

### Lint / 格式化 / 测试

* **gofmt** / **goimports** — Go 格式化（内置）
* **golangci-lint** — Go 静态检查（已启用 errcheck + revive，配置见 `backend/.golangci.yml`）
* **go test** — Go 测试（内置）
* **oxlint** — 前端 lint
* **prettier** — 前端格式化
* **vitest** — 前端单元测试

### 暂缓启用（未到时机不引入）

* **Recharts** — 做统计图时再用
* **Motion** — 需要动画时再用
* **TanStack Query v5** — 正式 API / 复杂数据请求时再用 ([TanStack][2])

### AI 开发辅助（非运行时依赖）

* **shadcn/ui Skills**
* **shadcn MCP Server**
* **Motion AI Kit / Skills**

## 环境变量

权威来源为 `.env.example`（复制为 `.env` 填写）。按用途分组：

| 变量 | 作用 | 必填 | 默认/说明 |
|---|---|---|---|
| `GITHUB_TOKEN` | GitHub API Token（单个，向后兼容） | 推荐 | 空则受未认证速率限制（60 req/h）；`GITHUB_TOKENS` 非空时优先使用 `GITHUB_TOKENS` |
| `GITHUB_TOKENS` | GitHub API Tokens（多个，逗号分隔） | 否 | 多 token 配额感知轮询（跳过低配额 token）+ 403/429 冷却切下一个；3 token × 5000 req/h = 15000 req/h |
| `TRACK_INTERVAL` | 追踪间隔（分钟） | 否 | `30` |
| `REPOS_FILE` | 仓库列表 JSON（仅首次 seed） | 否 | `./backend/config/repositories.json` |
| `DB_PATH` | SQLite 路径 | 否 | `./backend/data/tracker.db` |
| `EXPORT_DIR` | JSON 输出目录 | 否 | `./frontend/public/data` |
| `ADMIN_USERNAME` | 管理员用户名（仅首次 seed） | 否 | `admin` |
| `ADMIN_PASSWORD` | 管理员密码明文（首次 seed 后 bcrypt hash 入库） | 首次启动必填 | DB 已有 admin 时可空 |
| `JWT_SECRET` | JWT 签名密钥 | 是 | 空则拒绝启动；32 字节以上为最佳实践 |
| `SERVER_ADDR` | API server 监听地址 | 否 | `:8080` |
| `SERVER_DEV` | 开发模式 CORS | 否 | `false`；`true` 时允许 `localhost:3000` |
| `NEXT_PUBLIC_API_BASE_URL` | 前端 admin fetch baseURL | 否 | `http://localhost:8080` |
| `SITE_URL` | 站点正式域名 | 部署必填 | 未设置回退 `https://getreleased.example.com`，注入 sitemap/robots/OG |

## 构建与验证命令

### 后端

```bash
cd backend
go build ./cmd/tracker          # 追踪进程
go build ./cmd/server           # 管理 API 进程
go test ./...                   # 全部测试
golangci-lint run               # 静态检查
gofmt -l . && goimports -l .    # 格式检查（无输出即通过）
```

### 前端

```bash
cd frontend
pnpm install                    # 安装依赖
pnpm dev                        # 开发服务器
pnpm build                      # 生产构建（tsc --noEmit && next build）
pnpm test                       # 单元测试（vitest run）
pnpm lint                       # 代码检查（oxlint src）
pnpm format                     # 格式化（prettier --write .）
```

### 本地全链路

**一键启动**（Windows）：`./dev.ps1` — 构建后端 + 启动 server/tracker + frontend dev，Ctrl+C 自动清理。

**手动分步**：

1. 启动 server：`ADMIN_PASSWORD=xxx JWT_SECRET=yyy ./backend/bin/server`
2. 启动 tracker：`./backend/bin/tracker`
3. 开发前端：`cd frontend && pnpm dev`
4. 访问 `http://localhost:3000/admin` 进入管理后台

## 代码风格

### Go

* 格式化：`gofmt` + `goimports`（配置见 `backend/.golangci.yml`）。
* 命名：类型 `PascalCase`，方法/变量 `camelCase`，包名小写单词。
* struct 字段对齐（gofmt 自动处理），`json`/`db` tag 成对出现。
* 错误处理分层：内部包透传底层错误（`return err`），API/CLI 边界包装上下文（`fmt.Errorf("xxx: %w", err)`）；包装一律用 `%w` 保留错误链，顶层用 `slog.ErrorContext` + `os.Exit(1)`。
* 模块边界：见 `docs/project_structure.md` "模块边界"小节，各 `internal/` 子包职责单一，互不越界。

### 前端

* 格式化：`prettier`（配置见 `frontend/.prettierrc.json`：`semi: false`、`singleQuote: true`、`printWidth: 300`、`trailingComma: all`、`endOfLine: lf`）。
* Lint：`oxlint`（配置见 `frontend/.oxlintrc.json`）。
* 命名：组件 `PascalCase`（如 `RepositoryCard`、`ReleaseCard`），工具函数 `camelCase`，类型 `PascalCase`。
* 文件名：组件文件 `PascalCase.tsx`，工具/类型 `kebab-case.ts` 或 `index.ts`，文档 `snake_case.md`。
* 非必要不自写 CSS，优先用 Tailwind 工具类与已有 `globals.css`。
* 一律使用 pnpm，禁止 npm / yarn。

## 测试约定

### Go

* 测试文件 `*_test.go` 与源码**同目录**，包名相同（白盒测试）。
* 运行：`cd backend && go test ./...`。
* 命名：测试函数 `TestXxx`，表驱动用 `tests := []struct{...}` + `t.Run(tt.name, ...)`。

### 前端

* 测试文件 `*.test.ts` / `*.test.tsx`。
  * 组件测试：`__tests__/*.test.tsx`（如 `src/features/admin/__tests__/`）。
  * 工具/hooks 测试：与源码同目录（如 `src/lib/utils.test.ts`）。
* 运行：`cd frontend && pnpm test`（vitest）。
* 渲染测试用 `@testing-library/react` + `jsdom`。

### 通用

* 修改代码后必须同步更新相关测试，不交付未测试代码。
* 新增公开函数/组件时补对应测试。

## Git 工作流

* 主干分支 `master`，开发直接在 `master` 或短生命周期分支上进行。
* 提交信息简洁，聚焦 **why** 而非 what（如 `fix: tracker 漏写 latest_release_date 缓存`）。
* **不主动 commit / push**，除非用户明确要求。
* **lefthook** 已启用：pre-commit 按 staged 文件分流跑 lint（backend `golangci-lint`、frontend `oxlint`+`prettier --check`），pre-push 跑 test（`go test ./...`、`pnpm test`）。配置见 `lefthook.yml`。
* 禁止 `--no-verify` 跳过 hook，禁止 `--force` 推送 master。
* 提交前确保 `go test ./...`、`golangci-lint run`、`pnpm test`、`pnpm lint` 全部通过。

## 部署流程

数据更新需重新构建，顺序固定（不可颠倒）：

1. Go tracker 导出最新 JSON 到 `frontend/public/data/`
2. `pnpm build` 读取 JSON 生成静态产物 `out/`
3. 同步 `out/` 至 Nginx `root`

**一键部署**（Linux）：`ADMIN_USERNAME=xxx ADMIN_PASSWORD=yyy NGINX_ROOT=/var/www/getreleased ./deploy/deploy.sh` — login→export→build→rsync，固化上述顺序。

> 先 build 后导出会导致静态页数据过期。站点域名通过 `SITE_URL` 环境变量注入 sitemap/robots/OG，部署前必须设置真实域名。

Go Server（Admin API）作为独立 systemd 服务常驻运行，Nginx 通过 `/api/` 反代至 server。

## 项目结构

**AI 在创建文件、修改目录、涉及模块边界时，必须先读取 [docs/project_structure.md](docs/project_structure.md)。**

**核心约束**：顶层固定为 `backend/`、`frontend/`、`deploy/`，边界清晰，禁止提前实现未来功能。

## 常见陷阱

* **GitHub API 速率限制**：无 token 60 req/h，有 token 5000 req/h；批量追踪需设置 `GITHUB_TOKEN` 或 `GITHUB_TOKENS`（多 token 配额感知轮询，3 token 可达 15000 req/h）。tracker 每轮开始前检查配额，<500 跳过本轮。
* **SQLite 并发**：tracker 写入时 server 可读不可写；WAL 模式下读写可并发，但双写需避免。
* **JWT_SECRET**：空值拒绝启动；长度不足 32 字节虽可运行但不符最佳实践。
* **SITE_URL**：未设置时回退占位符 `https://getreleased.example.com`，线上 sitemap/robots/OG 域名错误，部署前必须设置。
* **部署顺序**：必须先导出 JSON 后 build，颠倒会导致静态页数据过期。
* **静态导出限制**：`output: 'export'` 下不支持运行时服务端 API；admin 页面通过客户端 `fetch` 调用 Go Server，不走 SSG 数据注入。
* **lefthook 未生效**：hook 不触发通常是 lefthook 未安装，需先安装二进制（`go install` 或从 [releases](https://github.com/evilmartians/lefthook/releases) 下载）再 `lefthook install`。

[1]: https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default "July 2026 - Base UI as the Default - shadcn/ui"
[2]: https://tanstack.com/query/latest/docs/framework/react "React | TanStack Query React Docs"
