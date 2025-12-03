@echo off
echo ========================================
echo Testing Delete via Naming Service
echo ========================================
echo.

set /p FILE_ID="Enter file_id to delete: "

echo.
echo Deleting file %FILE_ID% from all nodes...
echo.

curl -X DELETE http://localhost:8080/files/%FILE_ID%

echo.
echo.
echo ========================================
echo Delete Complete!
echo ========================================
echo.
echo Check the response for:
echo - deleted_from: number of nodes file was deleted from
echo - failed: number of nodes that failed
echo.
pause
