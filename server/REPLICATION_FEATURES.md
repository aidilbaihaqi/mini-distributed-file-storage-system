# Automated Replication & Fault Tolerance - Implementation Guide

Dokumentasi lengkap fitur automated replication dan fault tolerance yang sudah diimplementasikan.

---

## 🎯 Fitur yang Sudah Diimplementasikan

### ✅ 1. Automated Replication
- Setiap file yang diupload ke **sn-1 (Main Node)** otomatis direplikasi ke:
  - **sn-2 (Replica Node)**
  - **sn-3 (Backup Node)**
- Replikasi berjalan secara **parallel** menggunakan async/await
- Response upload mencakup status replikasi untuk setiap node

### ✅ 2. Fault Tolerance
- Jika salah satu node DOWN saat upload:
  - File tetap tersimpan di node yang UP
  - Node yang DOWN dicatat di **replication_queue**
  - Upload tetap berhasil (tidak gagal total)

### ✅ 3. Replication Queue
- Database table untuk tracking file yang gagal direplikasi
- Menyimpan informasi:
  - File yang perlu direplikasi
  - Target node yang DOWN
  - Source node yang memiliki file
  - Status: PENDING, IN_PROGRESS, COMPLETED, FAILED
  - Retry count dan error message

### ✅ 4. Auto-Recovery System
- Background job di naming service yang berjalan setiap 30 detik
- Otomatis mendeteksi node yang kembali UP
- Trigger recovery untuk sync file yang pending
- Update status di replication_queue

### ✅ 5. Manual Recovery Trigger
- Endpoint untuk trigger recovery secara manual
- Berguna untuk testing atau recovery on-demand
- Endpoint: `POST /nodes/{nodeId}/recover`

### ✅ 6. Metadata Management
- File metadata tersimpan di MySQL (naming service)
- Tracking lokasi file di setiap node
- Informasi checksum untuk validasi integritas

### ✅ 7. Monitoring Endpoints
- `/files` - List semua file dengan info replicas
- `/replication-queue` - Monitor queue status
- `/nodes` - Status semua nodes
- `/nodes/check` - Health check real-time

---

## 🏗️ Arsitektur

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Upload                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   Storage Node 1     │
              │   (Main - Port 8001) │
              └──────────┬───────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌────────────────┐ ┌────────────┐ ┌────────────────┐
│ Storage Node 2 │ │   Naming   │ │ Storage Node 3 │
│ (Replica-8002) │ │  Service   │ │ (Backup-8003)  │
└────────────────┘ │ (Port 8080)│ └────────────────┘
                   └──────┬─────┘
                          │
                   ┌──────▼──────┐
                   │    MySQL    │
                   │  (Metadata) │
                   └─────────────┘
```

---

## 📊 Database Schema

### Table: `nodes`
```sql
- id (PK)
- address
- status (UP/DOWN)
- role (MAIN/REPLICA/BACKUP)
- last_heartbeat
```

### Table: `files`
```sql
- file_key (PK)
- original_filename
- size_bytes
- checksum_sha256
- uploaded_at
```

### Table: `file_locations`
```sql
- id (PK)
- file_key (FK)
- node_id (FK)
- status (ACTIVE/DELETED)
```

### Table: `replication_queue`
```sql
- id (PK)
- file_key (FK)
- target_node_id (FK)
- source_node_id (FK)
- status (PENDING/IN_PROGRESS/COMPLETED/FAILED)
- retry_count
- last_attempt
- error_message
```

---

## 🔄 Flow Diagram

### Upload Flow (Normal - All Nodes UP)
```
1. Client → POST /files → sn-1
2. sn-1 → Save file locally
3. sn-1 → Replicate to sn-2 (parallel)
4. sn-1 → Replicate to sn-3 (parallel)
5. sn-1 → Register to naming service
6. naming service → Save metadata to MySQL
7. naming service → Save file_locations
8. sn-1 → Return response to client
```

### Upload Flow (Fault Tolerance - sn-2 DOWN)
```
1. Client → POST /files → sn-1
2. sn-1 → Save file locally ✅
3. sn-1 → Replicate to sn-2 ❌ (timeout/error)
4. sn-1 → Replicate to sn-3 ✅
5. sn-1 → Register to naming service
   - file_key, metadata
   - failed_nodes: ["node-2"]
6. naming service → Save metadata
7. naming service → Add to replication_queue
   - file_key, target: node-2, source: node-1
8. sn-1 → Return response (success with partial replication)
```

### Recovery Flow (Auto)
```
1. Background job (every 30s) → Check node health
2. Detect node-2: DOWN → UP
3. Query replication_queue for node-2
4. For each pending item:
   - Download file from source node
   - Upload to target node
   - Mark as COMPLETED
   - Update file_locations
```

### Recovery Flow (Manual)
```
1. Admin → POST /nodes/node-2/recover
2. Query replication_queue for node-2
3. Process all pending items
4. Return summary (success/failed count)
```

---

## 🔧 Technical Implementation

### Storage Node 1 (sn-1) - Python/FastAPI

**Key Functions:**
- `replicate_to_node()` - Replikasi ke 1 node
- `replicate_to_all_nodes()` - Replikasi parallel ke semua node
- `register_file_to_naming_service()` - Report hasil upload

**Dependencies:**
- `httpx` - Async HTTP client untuk replikasi

### Storage Node 2 & 3 - Python/FastAPI

**Endpoints:**
- `POST /files` - Terima file replikasi
- `GET /files/{file_id}` - Download file
- `DELETE /files/{file_id}` - Hapus file
- `GET /health` - Health check

### Naming Service - Go/Gin

**Key Functions:**
- `addToReplicationQueue()` - Tambah ke queue
- `getPendingReplications()` - Ambil pending items
- `replicateFileToNode()` - Replikasi file
- `markReplicationCompleted()` - Update status
- Background goroutine untuk auto-recovery

**Endpoints:**
- `POST /files/register` - Register file metadata
- `POST /nodes/{nodeId}/recover` - Manual recovery
- `GET /replication-queue` - Monitor queue
- `GET /files` - List files dengan replicas

---

## 🧪 Testing Scenarios

### Scenario 1: Normal Upload
```bash
# All nodes UP
curl -X POST http://localhost:8001/files -F "file=@test.jpg"

# Expected: File di 3 node, replication.successful = 2
```

### Scenario 2: Fault Tolerance
```bash
# Stop node-2
# Upload file
curl -X POST http://localhost:8001/files -F "file=@test2.jpg"

# Expected: File di 2 node, replication.failed = 1
# Check queue: curl http://localhost:8080/replication-queue
```

### Scenario 3: Auto Recovery
```bash
# Start node-2 kembali
# Wait 30 seconds
# Check logs: "Auto-recovered xyz to node-2"
# Verify: ls server/storage-node/sn-2/uploads/
```

### Scenario 4: Manual Recovery
```bash
curl -X POST http://localhost:8080/nodes/node-2/recover

# Expected: {"success": 3, "failed": 0}
```

---

## 📈 Performance Considerations

### Parallel Replication
- Menggunakan `asyncio.gather()` untuk replikasi parallel
- Timeout 30 detik per node
- Tidak blocking main upload process

### Background Job
- Interval 30 detik (configurable)
- Health check timeout 2 detik
- Batch processing untuk recovery

### Database Indexing
- Index pada `replication_queue.status`
- Index pada `replication_queue.target_node_id`
- Composite unique key pada `file_locations`

---

## 🚀 Next Steps

Fitur yang bisa ditambahkan:

1. **Smart Node Selection**
   - Latency-based routing
   - Load balancing

2. **Upload via Naming Service**
   - Client upload ke naming service
   - Naming service pilih node terbaik

3. **Download via Naming Service**
   - Automatic failover jika node DOWN
   - Load balancing untuk download

4. **Checksum Validation**
   - Verify file integrity setelah replikasi
   - Auto-retry jika checksum mismatch

5. **Monitoring Dashboard**
   - Real-time replication status
   - Node health visualization
   - Queue statistics

---

## 🐛 Troubleshooting

### File tidak tereplikasi
- Cek log sn-1 untuk error detail
- Verify node target UP: `curl http://localhost:8002/health`
- Check replication queue: `curl http://localhost:8080/replication-queue`

### Auto-recovery tidak jalan
- Pastikan naming service running
- Cek log untuk error message
- Trigger manual: `curl -X POST http://localhost:8080/nodes/node-2/recover`

### Database connection error
- Verify MySQL running
- Check credentials di DSN
- Run schema.sql untuk create tables

---

## 📝 Configuration

### Environment Variables

**Naming Service:**
```bash
export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"
```

**Storage Node 1:**
```python
REPLICA_NODES = [
    "http://localhost:8002",
    "http://localhost:8003",
]
NAMING_SERVICE_URL = "http://localhost:8080"
NODE_ID = "node-1"
```

### Ports
- Naming Service: 8080
- Storage Node 1: 8001
- Storage Node 2: 8002
- Storage Node 3: 8003
- MySQL: 3306

---

## ✅ Implementation Checklist

- [x] Automated replication dari sn-1 ke sn-2 dan sn-3
- [x] Fault tolerance saat node DOWN
- [x] Replication queue di database
- [x] Auto-recovery background job
- [x] Manual recovery endpoint
- [x] Metadata management di naming service
- [x] Monitoring endpoints
- [x] Testing scripts (Windows batch files)
- [x] Documentation
- [ ] Frontend integration (next phase)
- [ ] Latency-based routing (next phase)
- [ ] Upload/download via naming service (next phase)

---

Implementasi backend untuk automated replication dan fault tolerance sudah **COMPLETE** ✅
