#!/usr/bin/env bash
set -euo pipefail

: "${ADMIN_USERNAME:?需设置 ADMIN_USERNAME}"
: "${ADMIN_PASSWORD:?需设置 ADMIN_PASSWORD}"
: "${NGINX_ROOT:?需设置 NGINX_ROOT (nginx 静态根目录绝对路径)}"
: "${API_BASE_URL:=http://localhost:8080}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "==> 1/3 登录并导出 JSON..."
TOKEN="$(curl -fsSL -X POST "$API_BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" | jq -r '.token')"
curl -fsSL -X POST "$API_BASE_URL/api/admin/export" \
  -H "Authorization: Bearer $TOKEN"

echo "==> 2/3 构建前端 (pnpm build)..."
(cd frontend && pnpm build)

echo "==> 3/3 同步 out/ 至 nginx root ($NGINX_ROOT)..."
rsync -av --delete frontend/out/ "$NGINX_ROOT/"

echo "==> 部署完成"