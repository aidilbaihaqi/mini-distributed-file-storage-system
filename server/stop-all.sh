#!/bin/bash

echo "========================================"
echo "Stopping Mini DFS Services..."
echo "========================================"
echo ""

# Kill tmux session
tmux kill-session -t mini-dfs 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✓ All services stopped"
else
    echo "No running services found"
fi

# Also kill any remaining processes on the ports
echo ""
echo "Cleaning up any remaining processes..."

# Kill processes on ports
for port in 8080 8001 8002 8003; do
    pid=$(lsof -ti:$port 2>/dev/null)
    if [ ! -z "$pid" ]; then
        kill -9 $pid 2>/dev/null
        echo "✓ Killed process on port $port"
    fi
done

echo ""
echo "========================================"
echo "All services stopped!"
echo "========================================"
echo ""
