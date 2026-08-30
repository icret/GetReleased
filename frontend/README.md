# GetReleased Frontend

Next.js 16 App Router 静态导出（SSG）前端。构建期逐仓库预渲染完整静态 HTML，由 Nginx 纯静态托管。

## 技术栈

- Next.js 16（App Router，`output: 'export'` 静态导出）
- React 19 + TypeScript
- Tailwind CSS v4 + shadcn/ui（Base UI）
- Vitest（单元测试）/ Oxlint / Prettier

## 开发

```bash
pnpm dev
```

## 构建

构建依赖 `public/data/*.json` 存在（由 Go exporter 写出）。构建期通过 `generateStaticParams` 读取全量仓库，逐个预渲染静态 HTML 到 `out/`。

```bash
pnpm build
```

产物 `out/` 包含：

- `index.html` — 首页（含全部仓库正文）
- `repository/{owner}/{name}.html` — 各仓库 Release 页
- `404.html`、`robots.txt`、`sitemap.xml`
- `_next/static/` — 带 hash 的 JS/CSS

本地预览产物：

```bash
pnpm preview
```

## 测试 / Lint

```bash
pnpm test      # vitest
pnpm lint      # oxlint
pnpm format    # prettier
```

## 站点域名

`sitemap.xml` / `robots.txt` / Open Graph 中的站点域名取自环境变量 `SITE_URL`，未设置时回退占位符 `https://getreleased.example.com`。**部署前必须设置真实域名**，否则 SEO 资产 URL 错误。

```bash
SITE_URL=https://your-domain.com pnpm build
```

## 部署顺序（固定）

数据更新需重新构建，顺序不可颠倒：

1. **Go tracker** 导出最新 JSON 到 `frontend/public/data/`
2. **`pnpm build`** 读取 JSON 生成 `out/`
3. **同步 `out/`** 至 Nginx `root`（`/var/www/getreleased`）

> 先 build 后导出会导致静态页数据过期。
