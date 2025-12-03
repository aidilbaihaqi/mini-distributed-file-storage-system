# Mini DFS - Linux/WSL Setup Guide

## 🐧 Untuk Pengguna Linux/WSL

Karena Anda menggunakan **WSL (Windows Subsystem for Linux)**, gunakan script `.sh` (shell script) bukan `.bat` (Windows batch).

---

## 📋 Prerequisites

### 1. Install Dependencies

```bash
# Update package list
sudo apt update

# Install MySQL
sudo apt install mysql-server

# Install Python & pip
sudo apt install python3 python3-pip python3-venv

# Install Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install tmux (untuk menjalankan multiple services)
sudo apt install tmux

# Install curl
sudo apt install curl
```

### 2. Setup MySQL

```bash
# Start MySQL
sudo service mysql start

# Login ke MySQL
sudo mysql

# Buat database dan user
CREATE DATABASE dfs_meta;
CREATE USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';
GRANT ALL PRIVILEGES ON dfs_meta.* TO 'dfs_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;

# Import schema
cd ~/projects/dfs/server
mysql -u dfs_user -padmin123 dfs_meta < naming-service/schema.sql
```

### 3. Install Python Dependencies

```bash
cd ~/projects/dfs/server/storage-node/sn-1
pip3 install -r requirements.txt

cd ../sn-2
pip3 install -r requirements.txt

cd ../sn-3
pip3 install -r requirements.txt
```

### 4. Install Go Dependencies

```bash
cd ~/projects/dfs/server/naming-service
go mod tidy
```

---

## ✅ Verify Setup

```bash
cd ~/projects/dfs/server
./verify-setup.sh
```

Script ini akan check:
- MySQL installed & running
- Python installed
- Go installed
- Database connection
- Database tables
- Python dependencies
- Go dependencies

---

## 🚀 Menjalankan Services

### Option 1: Menggunakan tmux (Recommended)

```bash
cd ~/projects/dfs/server
./start-all.sh
```

Script ini akan:
1. Membuat tmux session bernama `mini-dfs`
2. Menjalankan 4 services di 4 windows berbeda:
   - Window 0: Naming Service (port 8080)
   - Window 1: Storage Node 1 (port 8001)
   - Window 2: Storage Node 2 (port 8002)
   - Window 3: Storage Node 3 (port 8003)

**Tmux Commands:**
```bash
# Attach ke tmux session
tmux attach -t mini-dfs

# Switch between windows
Ctrl+B then 0  # Naming Service
Ctrl+B then 1  # Storage Node 1
Ctrl+B then 2  # Storage Node 2
Ctrl+B then 3  # Storage Node 3

# Detach dari tmux (services tetap running)
Ctrl+B then D

# Stop all services
./stop-all.sh
```

### Option 2: Manual (4 Terminal Windows)

**Terminal 1 - Naming Service:**
```bash
cd ~/projects/dfs/server/naming-service
export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"
go run main.go
```

**Terminal 2 - Storage Node 1:**
```bash
cd ~/projects/dfs/server/storage-node/sn-1
uvicorn main:app --host 0.0.0.0 --port 8001
```

**Terminal 3 - Storage Node 2:**
```bash
cd ~/projects/dfs/server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002
```

**Terminal 4 - Storage Node 3:**
```bash
cd ~/projects/dfs/server/storage-node/sn-3
uvicorn main:app --host 0.0.0.0 --port 8003
```

---

## 🧪 Testing

### 1. Check Status

```bash
cd ~/projects/dfs/server
./check-status.sh
```

### 2. Test Upload

```bash
cd ~/projects/dfs/server
./test-upload.sh
```

### 3. Verify Replication

```bash
# Check files in each node
ls -la storage-node/sn-1/uploads/
ls -la storage-node/sn-2/uploads/
ls -la storage-node/sn-3/uploads/
```

### 4. Test Fault Tolerance

```bash
# Stop node-2
tmux attach -t mini-dfs
# Press Ctrl+B then 2 (go to window 2)
# Press Ctrl+C (stop node-2)
# Press Ctrl+B then D (detach)

# Upload file
curl -X POST http://localhost:8001/files -F "file=@test.txt"

# Check replication queue
curl http://localhost:8080/replication-queue?status=PENDING

# Start node-2 again
tmux attach -t mini-dfs
# Press Ctrl+B then 2
# Press Up Arrow (to get previous command)
# Press Enter (to restart uvicorn)
# Press Ctrl+B then D

# Wait 30 seconds for auto-recovery
# Or trigger manual recovery
curl -X POST http://localhost:8080/nodes/node-2/recover
```

---

## 🔍 Monitoring

### Check All Services

```bash
curl http://localhost:8080/health  # Naming Service
curl http://localhost:8001/health  # Storage Node 1
curl http://localhost:8002/health  # Storage Node 2
curl http://localhost:8003/health  # Storage Node 3
```

### List Files

```bash
curl http://localhost:8080/files
```

### Check Replication Queue

```bash
curl http://localhost:8080/replication-queue
```

### Check Nodes Status

```bash
curl http://localhost:8080/nodes
```

---

## 🧹 Cleanup

```bash
cd ~/projects/dfs/server
./cleanup.sh
```

Script ini akan:
- Delete semua uploaded files
- Reset database
- Delete test files

---

## 🛑 Stop Services

```bash
cd ~/projects/dfs/server
./stop-all.sh
```

---

## 📝 Available Scripts

### Linux Scripts (.sh)
- `start-all.sh` - Start all services in tmux
- `stop-all.sh` - Stop all services
- `check-status.sh` - Check status of all services
- `test-upload.sh` - Quick upload test
- `verify-setup.sh` - Verify setup requirements
- `cleanup.sh` - Clean uploaded files and reset database

### Windows Scripts (.bat) - Untuk Windows Native
- `start-all.bat` - Start all services (Windows CMD)
- `check-status.bat` - Check status (Windows CMD)
- `test-upload.bat` - Upload test (Windows CMD)
- Dan lain-lain...

**Note:** Karena Anda menggunakan WSL, gunakan script `.sh` bukan `.bat`

---

## 🐛 Troubleshooting

### MySQL tidak bisa connect

```bash
# Start MySQL service
sudo service mysql start

# Check status
sudo service mysql status

# Reset password jika perlu
sudo mysql
ALTER USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';
FLUSH PRIVILEGES;
```

### Port sudah digunakan

```bash
# Check process on port
lsof -i :8001

# Kill process
kill -9 <PID>
```

### Python module not found

```bash
cd ~/projects/dfs/server/storage-node/sn-1
pip3 install httpx fastapi uvicorn
```

### Go module error

```bash
cd ~/projects/dfs/server/naming-service
go mod tidy
go mod download
```

---

## 💡 Tips

### 1. Auto-start MySQL on WSL boot

```bash
echo "sudo service mysql start" >> ~/.bashrc
```

### 2. Alias untuk quick commands

```bash
# Add to ~/.bashrc
alias dfs-start='cd ~/projects/dfs/server && ./start-all.sh'
alias dfs-stop='cd ~/projects/dfs/server && ./stop-all.sh'
alias dfs-status='cd ~/projects/dfs/server && ./check-status.sh'
alias dfs-test='cd ~/projects/dfs/server && ./test-upload.sh'

# Reload bashrc
source ~/.bashrc

# Now you can use:
dfs-start
dfs-status
dfs-test
dfs-stop
```

### 3. View logs in tmux

```bash
# Attach to tmux
tmux attach -t mini-dfs

# Switch windows to see logs
Ctrl+B then 0  # Naming Service logs
Ctrl+B then 1  # Storage Node 1 logs
Ctrl+B then 2  # Storage Node 2 logs
Ctrl+B then 3  # Storage Node 3 logs

# Scroll in tmux
Ctrl+B then [  # Enter scroll mode
Use arrow keys or Page Up/Down
Press Q to exit scroll mode
```

---

## 🎯 Quick Start Summary

```bash
# 1. Verify setup
cd ~/projects/dfs/server
./verify-setup.sh

# 2. Start all services
./start-all.sh

# 3. Test upload
./test-upload.sh

# 4. Check status
./check-status.sh

# 5. Stop services
./stop-all.sh
```

---

## 📚 Documentation

- **README_LINUX.md** (this file) - Linux/WSL guide
- **QUICK_START_ID.md** - Quick start (general)
- **TESTING_REPLICATION.md** - Testing scenarios
- **TROUBLESHOOTING.md** - Common issues
- **IMPLEMENTATION_SUMMARY.md** - Technical overview

---

**Environment:** WSL (Windows Subsystem for Linux) / Ubuntu  
**Last Updated:** December 3, 2025
