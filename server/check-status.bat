@echo off
echo Checking Mini DFS Status...
echo.

echo ========================================
echo Naming Service Health:
echo ========================================
curl -s http://localhost:8080/health
echo.
echo.

echo ========================================
echo All Nodes Status:
echo ========================================
curl -s http://localhost:8080/nodes
echo.
echo.

echo ========================================
echo Node Health Check:
echo ========================================
curl -s http://localhost:8080/nodes/check
echo.
echo.

echo ========================================
echo Files List:
echo ========================================
curl -s http://localhost:8080/files
echo.
echo.

echo ========================================
echo Replication Queue:
echo ========================================
curl -s http://localhost:8080/replication-queue
echo.
echo.

pause
