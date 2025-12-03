@echo off
echo ========================================
echo Testing Upload/Download via Naming Service
echo ========================================
echo.

REM Create test file if not exists
if not exist test-routing.txt (
    echo This is a test file for routing via naming service > test-routing.txt
    echo Test file created: test-routing.txt
)

echo [1/4] Uploading file via naming service...
echo.
curl -X POST http://localhost:8080/upload -F "file=@test-routing.txt" > upload-response.json
type upload-response.json
echo.
echo.

REM Extract file_id manually (you need to look at the response)
echo [2/4] Please check upload-response.json for file_id
echo.
pause

set /p FILE_ID="Enter file_id from response: "

echo.
echo [3/4] Downloading file via naming service...
echo.
curl -v http://localhost:8080/download/%FILE_ID% -o downloaded-file.txt
echo.
echo.

echo [4/4] Verifying downloaded file...
echo.
type downloaded-file.txt
echo.
echo.

echo ========================================
echo Testing Complete!
echo ========================================
echo.
echo Check the following:
echo - upload-response.json for upload details
echo - downloaded-file.txt for downloaded content
echo - Response headers show X-Routed-From and X-Node-Latency-Ms
echo.
pause
