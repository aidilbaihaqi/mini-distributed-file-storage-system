#!/bin/bash

echo "========================================"
echo "Starting Mini DFS Services..."
echo "========================================"
echo ""

# Check if tmux is installed
if ! command -v tmux &> /dev/null; then
    echo "tmux is not installed. Installing..."
    sudo apt-get update && sudo apt-get install -y tmux
fi

# Kill existing tmux session if exists
tmux kill-session -t mini-dfs 2>/dev/null

# Create new tmux session
tmux new-session -d -s mini-dfs

# Window 0: Naming Service
tmux rename-window -t mini-dfs:0 'Naming-Service'
tmux send-keys -t mini-dfs:0 'cd naming-service' C-m
tmux send-keys -t mini-dfs:0 'export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"' C-m
tmux send-keys -t mini-dfs:0 'go run main.go' C-m

sleep 2

# Window 1: Storage Node 1
tmux new-window -t mini-dfs:1 -n 'Storage-Node-1'
tmux send-keys -t mini-dfs:1 'cd storage-node/sn-1' C-m
tmux send-keys -t mini-dfs:1 'uvicorn main:app --host 0.0.0.0 --port 8001' C-m

sleep 2

# Window 2: Storage Node 2
tmux new-window -t mini-dfs:2 -n 'Storage-Node-2'
tmux send-keys -t mini-dfs:2 'cd storage-node/sn-2' C-m
tmux send-keys -t mini-dfs:2 'uvicorn main:app --host 0.0.0.0 --port 8002' C-m

sleep 2

# Window 3: Storage Node 3
tmux new-window -t mini-dfs:3 -n 'Storage-Node-3'
tmux send-keys -t mini-dfs:3 'cd storage-node/sn-3' C-m
tmux send-keys -t mini-dfs:3 'uvicorn main:app --host 0.0.0.0 --port 8003' C-m

echo ""
echo "========================================"
echo "All services started in tmux session!"
echo "========================================"
echo ""
echo "Naming Service:    http://localhost:8080"
echo "Storage Node 1:    http://localhost:8001"
echo "Storage Node 2:    http://localhost:8002"
echo "Storage Node 3:    http://localhost:8003"
echo ""
echo "To view services:"
echo "  tmux attach -t mini-dfs"
echo ""
echo "To switch between windows:"
echo "  Ctrl+B then 0/1/2/3"
echo ""
echo "To detach from tmux:"
echo "  Ctrl+B then D"
echo ""
echo "To stop all services:"
echo "  ./stop-all.sh"
echo ""
