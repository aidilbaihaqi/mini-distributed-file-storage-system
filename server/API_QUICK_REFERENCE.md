# API Quick Reference - Mini DFS

Cheat sheet untuk API endpoints. **Semua request harus melalui Naming Service (Port 8080).**

---

## ⚠️ Important

```
✅ BENAR:  curl http://localhost:8080/upload -F "file=@test.txt"
❌ SALAH:  curl http://localhost:8001/files -F "file=@test.txt"

Client harus mengakses melalui Naming Service (8080), 
BUKAN langsung ke Storage Nodes (8001, 8002, 8003)
```

---

## 🚀 Quick Commands

### Health Check
```bash
curl http://localhost:8080/health
```

### Upload File ⭐
```bash
curl -X POST http://localhost:8080/upload -F "file=@myfile.txt"
```

### Download File ⭐
```bash
curl -O -J http://localhost:8080/download/{FILE_KEY}
```

### Delete File ⭐
```bash
curl -X DELETE http://localhost:8080/files/{FILE_KEY}
```

### List Files
```bash
curl http://localhost:8080/files
```

### Check Nodes
```bash
curl http://localhost:8080/nodes
```

### Check Replication Queue
```bash
curl http://localhost:8080/replication-queue
```

---

## 📊 All Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/upload` | **Upload file** |
| GET | `/download/{key}` | **Download file** |
| DELETE | `/files/{key}` | **Delete file** |
| GET | `/files` | List all files |
| GET | `/nodes` | List all nodes |
| GET | `/nodes/check` | Health check nodes |
| GET | `/replication-queue` | Monitor queue |
| POST | `/nodes/{id}/recover` | Manual recovery |

---

## 🔧 Common Workflows

### Upload & Download
```bash
# 1. Upload
curl -X POST http://localhost:8080/upload -F "file=@test.txt" > response.json

# 2. Get file_key
cat response.json | grep file_id

# 3. Download
curl -O -J http://localhost:8080/download/{FILE_KEY}
```

### Test Fault Tolerance
```bash
# 1. Stop node-2 (CTRL+C)

# 2. Upload file
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# 3. Check queue
curl http://localhost:8080/replication-queue?status=PENDING

# 4. Start node-2 again

# 5. Trigger recovery
curl -X POST http://localhost:8080/nodes/node-2/recover
```

### Monitor System
```bash
curl http://localhost:8080/health
curl http://localhost:8080/nodes
curl http://localhost:8080/files
curl http://localhost:8080/replication-queue
```

---

## 🎯 Response Examples

### Upload Success
```json
{
  "success": true,
  "file_id": "abc-123-def",
  "routed_via": "naming-service",
  "selected_node": "node-1",
  "node_latency_ms": 5,
  "replication": {
    "successful": 2,
    "failed": 0
  }
}
```

### Files List
```json
{
  "files": [
    {
      "file_key": "abc-123",
      "original_filename": "test.txt",
      "replicas": ["node-1", "node-2", "node-3"]
    }
  ],
  "count": 1
}
```

### Nodes List
```json
[
  {
    "id": "node-1",
    "status": "UP",
    "latency_ms": 5
  }
]
```

---

## 💡 Tips

### Pretty Print JSON
```bash
curl http://localhost:8080/files | jq
```

### Save Response
```bash
curl http://localhost:8080/files > files.json
```

### Check Response Time
```bash
curl -w "\nTime: %{time_total}s\n" http://localhost:8080/health
```

---

## 🔍 Filter Replication Queue

```bash
# By status
curl http://localhost:8080/replication-queue?status=PENDING
curl http://localhost:8080/replication-queue?status=COMPLETED

# By node
curl http://localhost:8080/replication-queue?node_id=node-2

# Both
curl "http://localhost:8080/replication-queue?node_id=node-2&status=PENDING"
```

---

## 📝 Notes

- Replace `{FILE_KEY}` dengan file_id dari response upload
- Semua request melalui port **8080** (Naming Service)
- Storage nodes (8001, 8002, 8003) adalah **internal API**
- Auto-recovery berjalan setiap 30 detik

---

**Base URL:** `http://localhost:8080`  
**Last Updated:** December 3, 2025
