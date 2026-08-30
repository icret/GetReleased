# GetReleased

追踪 GitHub 等开源软件 Release 的 Web 平台。

## 架构

- `backend/` — Go Tracker（追踪 GitHub Release 写入 SQLite 导出 JSON）+ Server（管理 API）
- `frontend/` — Next.js + React + TypeScript 前端（SSG 静态导出），展示 Release 信息 + 管理后台
- `deploy/` — Nginx 与 systemd 部署配置

详见 [docs/project_structure.md](docs/project_structure.md)。

## 开发

### 环境变量

复制 `.env.example` 为 `.env` 并填写：

- `GITHUB_TOKEN` — GitHub API Token
- `ADMIN_PASSWORD` — 管理员密码（明文，启动时 bcrypt hash）
- `JWT_SECRET` — JWT 签名密钥

### 后端

```bash
cd backend
go build ./cmd/tracker     # 追踪进程
go build ./cmd/server      # 管理 API 进程
go test ./...              # 测试
golangci-lint run          # 静态检查
```

### 前端

```bash
cd frontend
pnpm install
pnpm dev      # 开发服务器
pnpm build    # 生产构建
pnpm test     # 单元测试
pnpm lint     # 代码检查
```

### 本地全链路

1. 启动 server：`ADMIN_PASSWORD=xxx JWT_SECRET=yyy ./backend/bin/server`
2. 启动 tracker：`./backend/bin/tracker`
3. 开发前端：`cd frontend && pnpm dev`
4. 访问 `http://localhost:3000/admin` 进入管理后台

## 部署

数据更新需重新构建，顺序固定（不可颠倒）：

1. Go tracker 导出最新 JSON 到 `frontend/public/data/`
2. `pnpm build` 读取 JSON 生成静态产物 `out/`
3. 同步 `out/` 至 Nginx `root`

详见 [AGENTS.md](AGENTS.md) 部署流程。
