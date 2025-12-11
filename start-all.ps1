# Start All DFS Services
# Jalankan dengan: .\start-all.ps1

$rootPath = (Get-Location).Path

Write-Host "🚀 Starting Distributed File Storage System..." -ForegroundColor Cyan

# Start Naming Service (Go)
Write-Host "Starting Naming Service..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$rootPath\server\naming-service'; go run main.go"

Start-Sleep -Seconds 2

# Start Storage Node 1
Write-Host "Starting Storage Node 1..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$rootPath\server\storage-node\sn-1'; & .\venv\Scripts\Activate.ps1; uvicorn main:app --port 8001 --host 0.0.0.0"

# Start Storage Node 2
Write-Host "Starting Storage Node 2..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$rootPath\server\storage-node\sn-2'; & .\venv\Scripts\Activate.ps1; uvicorn main:app --port 8002 --host 0.0.0.0"

# Start Storage Node 3
Write-Host "Starting Storage Node 3..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$rootPath\server\storage-node\sn-3'; & .\venv\Scripts\Activate.ps1; uvicorn main:app --port 8003 --host 0.0.0.0"

Start-Sleep -Seconds 2

# Start Client (Next.js)
Write-Host "Starting Client..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$rootPath\client'; npm run dev"

Write-Host ""
Write-Host "✅ All services started!" -ForegroundColor Green
Write-Host ""
Write-Host "Services running at:" -ForegroundColor Cyan
Write-Host "  - Naming Service: http://localhost:8080"
Write-Host "  - Storage Node 1: http://localhost:8001"
Write-Host "  - Storage Node 2: http://localhost:8002"
Write-Host "  - Storage Node 3: http://localhost:8003"
Write-Host