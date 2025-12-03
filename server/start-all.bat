@echo off
echo Starting Mini DFS Services...
echo.

REM Start Naming Service
echo [1/4] Starting Naming Service on port 8080...
start "Naming Service" cmd /k "cd naming-service && set DB_DSN=dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true && go run main.go"
timeout /t 2 /nobreak >nul

REM Start Storage Node 1
echo [2/4] Starting Storage Node 1 on port 8001...
start "Storage Node 1" cmd /k "cd storage-node\sn-1 && uvicorn main:app --host 0.0.0.0 --port 8001"
timeout /t 2 /nobreak >nul

REM Start Storage Node 2
echo [3/4] Starting Storage Node 2 on port 8002...
start "Storage Node 2" cmd /k "cd storage-node\sn-2 && uvicorn main:app --host 0.0.0.0 --port 8002"
timeout /t 2 /nobreak >nul

REM Start Storage Node 3
echo [4/4] Starting Storage Node 3 on port 8003...
start "Storage Node 3" cmd /k "cd storage-node\sn-3 && uvicorn main:app --host 0.0.0.0 --port 8003"

echo.
echo ========================================
echo All services started!
echo ========================================
echo.
echo Naming Service:    http://localhost:8080
echo Storage Node 1:    http://localhost:8001
echo Storage Node 2:    http://localhost:8002
echo Storage Node 3:    http://localhost:8003
echo.
echo Press any key to exit...
pause >nul
