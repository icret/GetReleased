$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
Set-Location $repoRoot

Write-Host "==> 构建后端 (go build)..." -ForegroundColor Cyan
Push-Location backend
go build -o bin/server.exe ./cmd/server
go build -o bin/tracker.exe ./cmd/tracker
Pop-Location

Write-Host "==> 启动 server (http://localhost:8080)..." -ForegroundColor Cyan
$server = Start-Process -FilePath "backend\bin\server.exe" -WorkingDirectory $repoRoot -PassThru

Write-Host "==> 启动 tracker..." -ForegroundColor Cyan
$tracker = Start-Process -FilePath "backend\bin\tracker.exe" -WorkingDirectory $repoRoot -PassThru

Write-Host "==> 启动 frontend dev (http://localhost:3000)..." -ForegroundColor Cyan
Write-Host "按 Ctrl+C 退出，将自动停止 server 与 tracker" -ForegroundColor Yellow
Write-Host ""

try {
    Push-Location frontend
    pnpm dev
}
finally {
    Pop-Location
    foreach ($p in @($server, $tracker)) {
        if ($p -and -not $p.HasExited) {
            Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
        }
    }
}