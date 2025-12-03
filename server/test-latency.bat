@echo off
echo ========================================
echo Testing Latency-Based Node Selection
echo ========================================
echo.

echo [1/3] Checking current node latencies...
echo.
curl -s http://localhost:8080/nodes | jq ".[] | {id, status, latency_ms}"
echo.
echo.

echo [2/3] Uploading file (should route to lowest latency node)...
echo.
echo Test file for latency testing > test-latency.txt
curl -X POST http://localhost:8080/upload -F "file=@test-latency.txt"
echo.
echo.

echo [3/3] Checking which node was selected...
echo Look for "selected_node" and "node_latency_ms" in the response above
echo.
echo.

echo ========================================
echo Latency Test Complete!
echo ========================================
echo.
echo The file should have been routed to the node with lowest latency.
echo.
pause
