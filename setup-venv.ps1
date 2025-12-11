# Setup Virtual Environments untuk semua Storage Nodes
Write-Host "Setting up virtual environments..." -ForegroundColor Cyan

$nodes = @("sn-1", "sn-2", "sn-3")
$rootPath = (Get-Location).Path

foreach ($node in $nodes) {
    $nodePath = "$rootPath\server\storage-node\$node"
    Write-Host "Setting up $node..." -ForegroundColor Yellow
    
    Set-Location $nodePath
    
    # Buat venv
    python -m venv venv
    
    # Aktivasi dan install dependencies
    & .\venv\Scripts\Activate.ps1
    pip install -r requirements.txt
    deactivate
    
    Write-Host "$node done!" -ForegroundColor Green
}

Set-Location $rootPath
Write-Host ""
Write-Host "All virtual environments ready!" -ForegroundColor Green
