# 项目结构

> 目录索引（目录级，不列具体文件）。详细模块边界见下文。

## 顶层布局

```text
backend/
frontend/
deploy/
```

## 目录索引

```text
GetReleased/
│
├── backend/                   # Go Tracker + Server（追踪 + 导出 + 管理 API）
│   ├── cmd/
│   │   ├── tracker/           # 追踪进程入口
│   │   └── server/            # 管理 API 进程入口
│   ├── internal/
│   │   ├── github/            # GitHub API 客户端
│   │   ├── release/           # 模型、解析、版本比较
│   │   ├── scheduler/         # 定时调度
│   │   ├── database/          # SQLite 封装 + CRUD
│   │   ├── exporter/          # SQLite → JSON 导出
│   │   ├── avatar/            # 头像下载 + 本地缓存（存在即跳过）
│   │   ├── tracker/           # 追踪编排逻辑（tracker 与 server 共用）
│   │   ├── auth/              # bcrypt + JWT + 鉴权中间件（不掺业务逻辑）
│   │   ├── logging/           # slog 结构化日志 + requestID 中间件
│   │   └── api/               # HTTP handler + 路由 + 响应封装
│   ├── config/                # 追踪仓库列表（repositories.json，降级为 seed）
│   ├── migrations/            # 预留迁移文件
│   ├── data/                  # tracker.db（运行时生成，.gitignore）
│   ├── .golangci.yml          # golangci-lint 配置
│   ├── go.mod
│   └── go.sum
│
├── frontend/                  # Next.js + React + TypeScript（SSG 静态导出）
│   ├── public/
│   │   ├── data/              # JSON 数据（exporter 写入）
│   │   ├── assets/images/repos/  # 仓库头像（avatar 包下载，按 owner.png 命名）
│   │   ├── favicon.svg
│   │   └── icons.svg
│   ├── src/
│   │   ├── app/               # App Router 路由与页面
│   │   │   ├── admin/         # 管理后台（layout + repositories/tags/users 子页）
│   │   │   ├── login/         # 管理员登录
│   │   │   ├── repository/[owner]/[name]/page.tsx
│   │   │   ├── layout.tsx     # 根布局
│   │   │   ├── page.tsx       # 首页
│   │   │   ├── not-found.tsx
│   │   │   ├── robots.ts
│   │   │   ├── sitemap.ts
│   │   │   └── globals.css    # Tailwind v4 + 主题变量
│   │   ├── components/        # 通用组件（含 ui/ shadcn）
│   │   ├── data/              # 构建期 JSON 读取
│   │   ├── features/          # 功能模块（home/、releases/、repositories/、dashboard/、admin/）
│   │   ├── hooks/             # 自定义 hooks
│   │   ├── types/             # 类型定义
│   │   └── lib/               # 工具函数
│   ├── .oxlintrc.json         # oxlint 配置

│   ├── .prettierrc.json       # prettier 配置
│   ├── .prettierignore
│   ├── components.json        # shadcn/ui 配置
│   ├── next.config.ts
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── postcss.config.mjs
│   ├── tsconfig.json
│   ├── vitest.config.ts
│   └── README.md
│
├── deploy/
│   ├── nginx/                 # Nginx 配置
│   ├── systemd/               # systemd 服务单元
│   └── deploy.sh              # 一键部署 (login→export→build→rsync 同步)
│
├── docs/
│   └── archive/              # 已完成的计划归档
│
├── .env.example
├── .gitignore
├── dev.ps1                    # 一键启动本地全链路开发 (server + tracker + frontend dev)
├── lefthook.yml               # Git hooks 配置 (pre-commit lint + pre-push test)
└── README.md
```

## 模块边界（严格遵守）

后端各目录职责单一，互不越界：

* **`backend/internal/github/`** — 只负责 GitHub API，不掺数据库逻辑。
* **`backend/internal/database/`** — 只负责 SQLite，不掺 GitHub API。
* **`backend/internal/exporter/`** — 只负责 `SQLite → JSON` 导出。
* **`backend/internal/avatar/`** — 只负责头像下载与本地缓存，不掺 GitHub API/DB 逻辑。
* **`backend/internal/tracker/`** — 追踪编排逻辑，tracker 与 server 共用。
* **`backend/internal/auth/`** — bcrypt + JWT + 鉴权中间件，不掺业务逻辑。
* **`backend/internal/api/`** — HTTP handler + 路由 + 响应封装；经 database 包读写 SQLite、经 tracker 包编排追踪、经 exporter 包导出、经 github 包访问 GitHub API。

目的：AI 修改某一部分时不需要理解整个项目。
