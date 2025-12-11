# Stop All DFS Services
Write-Host "🛑 Stopping all DFS services..." -ForegroundColor Yellow

# Stop processes by port
$ports = @(8080, 8001, 8002, 8003, 3000)

foreach ($port in $ports) {
    $process = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | 
               Select-Object -ExpandProperty OwningProcess -ErrorAction SilentlyContinue
    
    if ($process) {
        Stop-Process -Id $process -Force -ErrorAction SilentlyContinue
        Write-Host "  Stopped process on port $port" -ForegroundColor Gray
    }
}

Write-Host "✅ All services stopped!" -ForegroundColor Green
