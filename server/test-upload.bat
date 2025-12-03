@echo off
echo Testing file upload to Mini DFS...
echo.

REM Create a test file if it doesn't exist
if not exist test.txt (
    echo This is a test file for Mini DFS > test.txt
    echo Test file created: test.txt
)

echo Uploading test.txt to Storage Node 1...
echo.

curl -X POST http://localhost:8001/files -F "file=@test.txt"

echo.
echo.
echo Upload complete! Check the response above.
echo.
echo To verify replication, check:
echo - server\storage-node\sn-1\uploads\
echo - server\storage-node\sn-2\uploads\
echo - server\storage-node\sn-3\uploads\
echo.
pause
