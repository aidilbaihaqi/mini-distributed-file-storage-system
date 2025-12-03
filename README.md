# Mini Distributed File Storage System  
Sistem penyimpanan file terdistribusi dengan **replication**, **fault tolerance**, dan **smart-node detection**, dikembangkan sebagai proyek akhir mata kuliah Sistem Terdistribusi.

Project ini menggunakan pendekatan arsitektur multi-service:

- **server/** → seluruh backend (Naming Service, Storage Nodes, Database)
- **client/** → Web dashboard (Next.js + React)

---

## 🚀 Fitur Utama

### ✅ Distributed Storage Architecture
- File disimpan di 3 node penyimpanan:
  - `sn-1` → Main Storage Node (Port 8001)
  - `sn-2` → Replica Storage Node (Port 8002)
  - `sn-3` → Backup Storage Node (Port 8003)

### ✅ Automated Replication (IMPLEMENTED)
- Setiap upload ke sn-1 → **otomatis direplikasi** ke sn-2 dan sn-3
- Replikasi berjalan **parallel** menggunakan async/await
- Metadata tersimpan di MySQL
- Response mencakup status replikasi per node

### ✅ Fault Tolerance (IMPLEMENTED)
- Upload tetap berhasil meski 1-2 node DOWN
- File yang gagal direplikasi masuk **replication_queue**
- Sistem tidak rollback jika ada node yang gagal
- Tracking lengkap di database

### ✅ Auto-Recovery System (IMPLEMENTED)
- Background job berjalan setiap 30 detik
- Otomatis detect node yang kembali UP
- Trigger recovery untuk sync file yang pending
- Update status di replication_queue (PENDING → COMPLETED)

### ✅ Manual Recovery (IMPLEMENTED)
- Endpoint untuk trigger recovery on-demand
- Berguna untuk testing dan maintenance
- API: `POST /nodes/{nodeId}/recover`

### ✅ Replication Queue (IMPLEMENTED)
- Database table untuk tracking replikasi
- Status: PENDING, IN_PROGRESS, COMPLETED, FAILED
- Retry count dan error message
- Monitoring via API

### ✅ Metadata Management (IMPLEMENTED)
- File metadata di MySQL (naming service)
- Tracking lokasi file di setiap node
- Checksum SHA256 untuk validasi
- API untuk list files dengan info replicas

### 🔹 Smart Node Detection (Latency-based) - PLANNED
Naming service memilih node terbaik berdasarkan:
1. Status UP
2. Latency terendah
3. Ketersediaan file

### 🔹 Dashboard Monitoring - PARTIAL
Frontend menampilkan (masih mock data):
- Status node (UP/DOWN)
- Latency node
- File explorer
- Statistik replikasi
- Log aktivitas

---

## 🧩 Struktur Direktori

```
root/
│
├── client/
│   ├── public/
│   ├── src/
│   ├── package.json
│   └── Dockerfile
│
└── server/
    ├── docker-compose.yml
    │
    ├── naming-service/
    │   ├── main.go
    │   ├── go.mod
    │   └── Dockerfile
    │
    └── storage-node/
        ├── sn-1/
        ├── sn-2/
        └── sn-3/
```

---

## 📦 Teknologi

| Layer | Teknologi |
|------|-----------|
| Naming Service | Go (Gin) |
| Storage Nodes | Python FastAPI |
| Database | MySQL 8 |
| Frontend | Next.js + React |
| DevOps | Docker Compose (opsional) |

---

## 🗄 Database Metadata

### Tabel utama:

#### `nodes`
Menyimpan info node:
- status
- latency
- heartbeat

#### `files`
Metadata global file:
- file_key
- nama asli
- ukuran

#### `file_locations`
Lokasi file pada node.

#### `replication_queue`
Backlog replikasi ketika node DOWN.

---

## ▶️ Menjalankan Aplikasi

### Quick Start (Backend Only)

1. **Setup Database:**
```bash
mysql -u dfs_user -padmin123 dfs_meta < server/naming-service/schema.sql
```

2. **Start All Services (Windows):**
```bash
cd server
start-all.bat
```

3. **Test Upload:**
```bash
cd server
test-upload.bat
```

4. **Check Status:**
```bash
cd server
check-status.bat
```

### Manual Start

Lihat **server/QUICK_START_ID.md** untuk panduan lengkap.

Untuk instalasi dari awal, lihat **INSTALLATION.md**.

---

## 🧪 Pengujian

### Test upload (via naming service dengan latency-based routing):
```bash
curl -X POST http://localhost:8080/upload -F "file=@test.jpg"
```

### Test download (via naming service):
```bash
curl -O http://localhost:8080/download/{FILE_ID}
```

### Test delete (dari semua node):
```bash
curl -X DELETE http://localhost:8080/files/{FILE_ID}
```

### Test latency-based selection:
```bash
# Check node latencies
curl http://localhost:8080/nodes

# Upload will route to node with lowest latency
curl -X POST http://localhost:8080/upload -F "file=@test.jpg"
```

### Test fault tolerance:
```bash
# Stop node-2 (CTRL+C)
# Upload file (will route to other node)
curl -X POST http://localhost:8080/upload -F "file=@test.jpg"
# Check replication queue
curl http://localhost:8080/replication-queue
```

### Test auto-recovery:
```bash
# Start node-2 kembali
# Wait 30 seconds
# Check logs: "Auto-recovered ... to node-2"
```

### Monitoring:
```bash
# List all files with replicas
curl http://localhost:8080/files

# Check nodes status with latency
curl http://localhost:8080/nodes

# Monitor replication queue
curl http://localhost:8080/replication-queue
```

### Testing Scripts (Windows):
```bash
cd server
test-routing.bat        # Test upload/download routing
test-latency.bat        # Test latency-based selection
test-delete-routing.bat # Test delete from all nodes
```

Lihat **server/TESTING_REPLICATION.md** dan **server/ROUTING_FEATURES.md** untuk dokumentasi lengkap.

---

## 📚 Dokumentasi

### Backend (Implemented)
- **server/IMPLEMENTATION_SUMMARY.md** - Overview implementasi lengkap
- **server/ROUTING_FEATURES.md** - Routing & latency-based selection
- **server/REPLICATION_FEATURES.md** - Automated replication
- **server/QUICK_START_ID.md** - Quick start guide (Bahasa Indonesia)
- **server/TESTING_REPLICATION.md** - Testing guide dengan 5 skenario
- **server/TROUBLESHOOTING.md** - Troubleshooting guide
- **server/CHANGELOG.md** - Version history
- **server/naming-service/schema.sql** - Database schema

### Frontend (Partial - Mock Data)
- **client/README.md** - Frontend documentation

---

## 🎯 Implementation Status

### ✅ Completed (Backend)
- [x] Automated replication (sn-1 → sn-2, sn-3)
- [x] Fault tolerance (upload tetap berhasil meski node DOWN)
- [x] Replication queue (tracking di database)
- [x] Auto-recovery (background job 30s interval)
- [x] Manual recovery (API endpoint)
- [x] Metadata management (MySQL)
- [x] Upload/download routing via naming service
- [x] Latency-based node selection
- [x] Smart routing dengan automatic failover
- [x] Monitoring endpoints
- [x] Testing scripts (Windows batch files)
- [x] Complete documentation

### ✅ Upload/Download Routing (IMPLEMENTED)
- Semua operasi file via naming service (port 8080)
- Client tidak perlu tahu alamat storage node
- Centralized control dan monitoring
- API: `POST /upload`, `GET /download/{fileKey}`, `DELETE /files/{fileKey}`

### ✅ Latency-Based Selection (IMPLEMENTED)
- Background job ukur latency setiap 30 detik
- Automatic pilih node tercepat untuk upload
- Automatic pilih node tercepat untuk download
- Database simpan latency_ms per node
- Optimal performance dan load distribution

### ⏳ Pending (Next Phase)
- [ ] Frontend integration dengan backend API
- [ ] Checksum validation setelah replikasi
- [ ] File compression
- [ ] Encryption

---

## 👥 Pengembang
- Backend Gin / FastAPI ✅
- Database & Replication Logic ✅
- DevOps & Testing Scripts ✅
- Frontend Next.js (partial - mock data)

---

## 📝 Lisensi
Bebas digunakan untuk pembelajaran dan tugas akademik.

---

## 🚀 Quick Links

- [Quick Start Guide](server/QUICK_START_ID.md)
- [Testing Guide](server/TESTING_REPLICATION.md)
- [Implementation Summary](server/IMPLEMENTATION_SUMMARY.md)
- [Troubleshooting](server/TROUBLESHOOTING.md)
