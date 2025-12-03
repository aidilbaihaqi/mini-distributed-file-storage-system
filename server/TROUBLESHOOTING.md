# Troubleshooting Guide - Mini DFS

Panduan mengatasi masalah umum pada Mini Distributed File Storage System.

---

## 🔧 Common Issues

### 1. Database Connection Failed

**Error:**
```
gagal buka koneksi ke MySQL: Error 1045: Access denied
```

**Solutions:**

a) Verify MySQL is running:
```bash
# Windows
net start MySQL80

# Linux
sudo systemctl status mysql
```

b) Check credentials:
```bash
mysql -u dfs_user -padmin123 dfs_meta
```

c) Recreate user if needed:
```sql
DROP USER 'dfs_user'@'localhost';
CREATE USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';
GRANT ALL PRIVILEGES ON dfs_meta.* TO 'dfs_user'@'localhost';
FLUSH PRIVILEGES;
```

d) Verify DSN format:
```bash
# Correct format
export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"

# Windows
set DB_DSN=dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true
```

---

### 2. Port Already in Use

**Error:**
```
Error: [Errno 10048] Only one usage of each socket address
```

**Solutions:**

a) Check which process is using the port:
```bash
# Windows
netstat -ano | findstr "8001"

# Linux
lsof -i :8001
```

b) Kill the process:
```bash
# Windows
taskkill /PID <PID> /F

# Linux
kill -9 <PID>
```

c) Use different port:
```bash
uvicorn main:app --port 8011
```

---

### 3. File Not Replicated

**Symptoms:**
- Upload berhasil tapi file tidak ada di sn-2 atau sn-3
- Replication status shows "failed"

**Diagnosis:**

a) Check if target nodes are running:
```bash
curl http://localhost:8002/health
curl http://localhost:8003/health
```

b) Check replication queue:
```bash
curl http://localhost:8080/replication-queue?status=PENDING
```

c) Check sn-1 logs for error messages

**Solutions:**

a) If nodes are DOWN, start them:
```bash
cd server/storage-node/sn-2
uvicorn main:app --port 8002
```

b) Trigger manual recovery:
```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```

c) Check network connectivity:
```bash
curl -v http://localhost:8002/health
```

---

### 4. Auto-Recovery Not Working

**Symptoms:**
- Node kembali UP tapi file tidak ter-sync
- Replication queue tetap PENDING

**Diagnosis:**

a) Check naming service is running:
```bash
curl http://localhost:8080/health
```

b) Check naming service logs for errors

c) Verify background job is running (should see log every 30s)

**Solutions:**

a) Restart naming service:
```bash
cd server/naming-service
set DB_DSN=dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true
go run main.go
```

b) Trigger manual recovery:
```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```

c) Check if source node has the file:
```bash
# Get file_key from replication_queue
curl http://localhost:8080/replication-queue

# Try download from source node
curl http://localhost:8001/files/{file_key}
```

---

### 5. Module Not Found Error (Python)

**Error:**
```
ModuleNotFoundError: No module named 'httpx'
```

**Solutions:**

a) Install missing module:
```bash
pip install httpx
```

b) Install all requirements:
```bash
cd server/storage-node/sn-1
pip install -r requirements.txt
```

c) Use virtual environment:
```bash
python -m venv venv
venv\Scripts\activate  # Windows
source venv/bin/activate  # Linux
pip install -r requirements.txt
```

---

### 6. Go Module Error

**Error:**
```
cannot find package "github.com/gin-gonic/gin"
```

**Solutions:**

a) Download dependencies:
```bash
cd server/naming-service
go mod tidy
go mod download
```

b) If go.mod is missing:
```bash
go mod init naming-service
go get github.com/gin-gonic/gin
go get github.com/go-sql-driver/mysql
```

---

### 7. File Upload Timeout

**Symptoms:**
- Upload takes too long
- Connection timeout error

**Diagnosis:**

a) Check file size:
```bash
# Large files may take longer
```

b) Check network latency:
```bash
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8001/health
```

**Solutions:**

a) Increase timeout in sn-1:
```python
# In main.py
async with httpx.AsyncClient(timeout=60.0) as client:  # Increase from 30
```

b) Upload smaller files for testing

c) Check disk space:
```bash
# Windows
dir server\storage-node\sn-1\uploads

# Linux
df -h
```

---

### 8. Database Table Not Found

**Error:**
```
Error 1146: Table 'dfs_meta.replication_queue' doesn't exist
```

**Solutions:**

a) Run schema.sql:
```bash
cd server/naming-service
mysql -u dfs_user -padmin123 dfs_meta < schema.sql
```

b) Verify tables exist:
```bash
mysql -u dfs_user -padmin123 dfs_meta -e "SHOW TABLES;"
```

c) Check table structure:
```bash
mysql -u dfs_user -padmin123 dfs_meta -e "DESCRIBE replication_queue;"
```

---

### 9. CORS Error (Future Frontend Integration)

**Error:**
```
Access to fetch blocked by CORS policy
```

**Solutions:**

a) Add CORS middleware to FastAPI:
```python
from fastapi.middleware.cors import CORSMiddleware

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
```

b) Add CORS to Gin:
```go
import "github.com/gin-contrib/cors"

r.Use(cors.Default())
```

---

### 10. File Download Returns 404

**Symptoms:**
- Upload berhasil
- Download gagal dengan 404

**Diagnosis:**

a) Check if file exists:
```bash
dir server\storage-node\sn-1\uploads
```

b) Verify file_id is correct:
```bash
curl http://localhost:8080/files
```

c) Check metadata:
```bash
# Look for .meta.json file
dir server\storage-node\sn-1\uploads\*.meta.json
```

**Solutions:**

a) Use correct file_id from upload response

b) Check file permissions:
```bash
# Linux
ls -la server/storage-node/sn-1/uploads/
```

c) Verify resolve_file_path logic in main.py

---

## 🔍 Debugging Tips

### Enable Verbose Logging

**FastAPI:**
```bash
uvicorn main:app --port 8001 --log-level debug
```

**Go:**
```go
// Add more log statements
log.Printf("Debug: %+v\n", variable)
```

### Check Process Status

```bash
# Windows
tasklist | findstr "python"
tasklist | findstr "go"

# Linux
ps aux | grep python
ps aux | grep go
```

### Monitor Network Traffic

```bash
# Windows
netstat -an | findstr "800"

# Linux
netstat -tulpn | grep 800
```

### Check Disk Space

```bash
# Windows
dir server\storage-node\sn-1\uploads

# Linux
du -sh server/storage-node/*/uploads/
```

### Database Queries for Debugging

```sql
-- Check nodes status
SELECT * FROM nodes;

-- Check files
SELECT * FROM files ORDER BY uploaded_at DESC LIMIT 10;

-- Check file locations
SELECT f.file_key, f.original_filename, fl.node_id, fl.status
FROM files f
LEFT JOIN file_locations fl ON f.file_key = fl.file_key;

-- Check replication queue
SELECT * FROM replication_queue WHERE status = 'PENDING';

-- Count replicas per file
SELECT f.file_key, f.original_filename, COUNT(fl.node_id) as replica_count
FROM files f
LEFT JOIN file_locations fl ON f.file_key = fl.file_key
WHERE fl.status = 'ACTIVE'
GROUP BY f.file_key;
```

---

## 🧪 Testing Commands

### Quick Health Check All Services

```bash
@echo off
echo Checking all services...
curl -s http://localhost:8080/health && echo [OK] Naming Service || echo [FAIL] Naming Service
curl -s http://localhost:8001/health && echo [OK] Storage Node 1 || echo [FAIL] Storage Node 1
curl -s http://localhost:8002/health && echo [OK] Storage Node 2 || echo [FAIL] Storage Node 2
curl -s http://localhost:8003/health && echo [OK] Storage Node 3 || echo [FAIL] Storage Node 3
```

### Test Upload and Verify

```bash
# Upload
curl -X POST http://localhost:8001/files -F "file=@test.txt" > response.json

# Extract file_id (manual)
# Then verify on all nodes
curl http://localhost:8001/files/{file_id}
curl http://localhost:8002/files/{file_id}
curl http://localhost:8003/files/{file_id}
```

### Simulate Node Failure

```bash
# Stop node-2 (CTRL+C in terminal)

# Upload file
curl -X POST http://localhost:8001/files -F "file=@test.txt"

# Check queue
curl http://localhost:8080/replication-queue?status=PENDING

# Start node-2 again
cd server/storage-node/sn-2
uvicorn main:app --port 8002

# Wait 30s or trigger recovery
curl -X POST http://localhost:8080/nodes/node-2/recover

# Verify
curl http://localhost:8080/replication-queue?status=COMPLETED
```

---

## 📞 Getting Help

### Check Logs

1. **Naming Service logs** - Terminal output
2. **Storage Node logs** - Terminal output
3. **MySQL logs** - Check MySQL error log

### Collect Debug Info

```bash
# System info
systeminfo  # Windows
uname -a    # Linux

# Service status
curl http://localhost:8080/health
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health

# Database status
mysql -u dfs_user -padmin123 dfs_meta -e "SELECT * FROM nodes;"
mysql -u dfs_user -padmin123 dfs_meta -e "SELECT COUNT(*) FROM files;"
mysql -u dfs_user -padmin123 dfs_meta -e "SELECT * FROM replication_queue WHERE status='PENDING';"

# File counts
dir server\storage-node\sn-1\uploads | find /c ".meta.json"
dir server\storage-node\sn-2\uploads | find /c ".meta.json"
dir server\storage-node\sn-3\uploads | find /c ".meta.json"
```

---

## ✅ Verification Checklist

Before reporting an issue, verify:

- [ ] MySQL is running
- [ ] Database `dfs_meta` exists
- [ ] Tables created (run schema.sql)
- [ ] All dependencies installed (httpx, gin, etc)
- [ ] No port conflicts
- [ ] All services started successfully
- [ ] Can access health endpoints
- [ ] Firewall not blocking ports
- [ ] Sufficient disk space
- [ ] Correct file permissions

---

## 🔄 Reset Everything

If all else fails, reset to clean state:

```bash
# 1. Stop all services (CTRL+C in all terminals)

# 2. Clear uploads
del /Q server\storage-node\sn-1\uploads\*.*
del /Q server\storage-node\sn-2\uploads\*.*
del /Q server\storage-node\sn-3\uploads\*.*

# 3. Reset database
mysql -u dfs_user -padmin123 dfs_meta -e "DROP TABLE IF EXISTS replication_queue, file_locations, files, nodes;"
mysql -u dfs_user -padmin123 dfs_meta < server/naming-service/schema.sql

# 4. Restart all services
cd server
start-all.bat
```

---

## 📚 Additional Resources

- **QUICK_START_ID.md** - Setup guide
- **TESTING_REPLICATION.md** - Testing scenarios
- **REPLICATION_FEATURES.md** - Technical details
- **IMPLEMENTATION_SUMMARY.md** - Overview

---

**Last Updated:** December 3, 2025
