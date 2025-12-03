# Quick Start - Automated Replication & Fault Tolerance

Panduan cepat untuk menjalankan dan testing fitur automated replication dan fault tolerance.

---

## 📋 Persiapan

### 1. Setup Database

```bash
# Masuk ke MySQL
mysql -u root -p

# Buat database dan user (jika belum)
CREATE DATABASE dfs_meta;
CREATE USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';
GRANT ALL PRIVILEGES ON dfs_meta.* TO 'dfs_user'@'localhost';
FLUSH PRIVILEGES;
exit;

# Import schema
cd server/naming-service
mysql -u dfs_user -padmin123 dfs_meta < schema.sql
```

### 2. Install Dependencies

**Storage Node 1:**
```bash
cd server/storage-node/sn-1
pip install httpx
# atau
pip install -r requirements.txt
```

**Storage Node 2 & 3:**
```bash
cd server/storage-node/sn-2
pip install -r requirements.txt

cd ../sn-3
pip install -r requirements.txt
```

**Naming Service:**
```bash
cd server/naming-service
go mod tidy
```

---

## 🚀 Menjalankan Services

### Opsi 1: Menggunakan Batch Script (Windows)

```bash
cd server
start-all.bat
```

Script ini akan membuka 4 terminal window untuk:
- Naming Service (port 8080)
- Storage Node 1 (port 8001)
- Storage Node 2 (port 8002)
- Storage Node 3 (port 8003)

### Opsi 2: Manual (Recommended untuk Development)

**Terminal 1 - Naming Service:**
```bash
cd server/naming-service
set DB_DSN=dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true
go run main.go
```

**Terminal 2 - Storage Node 1:**
```bash
cd server/storage-node/sn-1
uvicorn main:app --host 0.0.0.0 --port 8001
```

**Terminal 3 - Storage Node 2:**
```bash
cd server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002
```

**Terminal 4 - Storage Node 3:**
```bash
cd server/storage-node/sn-3
uvicorn main:app --host 0.0.0.0 --port 8003
```

---

## ✅ Verifikasi Services Running

```bash
cd server
check-status.bat
```

Atau manual:
```bash
# Naming service
curl http://localhost:8080/health

# Storage nodes
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health
```

---

## 🧪 Testing

### Test 1: Upload Normal (Semua Node UP)

```bash
cd server
test-upload.bat
```

Atau manual:
```bash
# Buat test file
echo "Test file content" > test.txt

# Upload
curl -X POST http://localhost:8001/files -F "file=@test.txt"
```

**Expected Response:**
```json
{
  "success": true,
  "file_id": "abc-123-def",
  "replication": {
    "total_nodes": 2,
    "successful": 2,
    "failed": 0
  }
}
```

**Verifikasi:**
```bash
# File harus ada di 3 folder
dir server\storage-node\sn-1\uploads
dir server\storage-node\sn-2\uploads
dir server\storage-node\sn-3\uploads
```

### Test 2: Fault Tolerance (1 Node DOWN)

1. **Stop node-2** (CTRL+C di terminal node-2)

2. **Upload file baru:**
```bash
echo "Test file 2" > test2.txt
curl -X POST http://localhost:8001/files -F "file=@test2.txt"
```

**Expected Response:**
```json
{
  "success": true,
  "file_id": "xyz-456",
  "replication": {
    "successful": 1,
    "failed": 1
  }
}
```

3. **Cek replication queue:**
```bash
curl http://localhost:8080/replication-queue
```

Akan muncul entry untuk node-2 dengan status PENDING.

### Test 3: Auto-Recovery

1. **Hidupkan kembali node-2:**
```bash
cd server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002
```

2. **Tunggu 30 detik** - Lihat log naming service:
```
Node node-2 status changed: DOWN -> UP
🔄 Triggering recovery for node node-2
✅ Auto-recovered xyz-456 to node-2
```

3. **Verifikasi file sudah ada:**
```bash
dir server\storage-node\sn-2\uploads
```

### Test 4: Manual Recovery

Jika tidak ingin menunggu auto-recovery:

```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```

**Response:**
```json
{
  "message": "recovery completed",
  "total": 1,
  "success": 1,
  "failed": 0
}
```

---

## 📊 Monitoring

### Cek Status Nodes
```bash
curl http://localhost:8080/nodes
```

### List Semua Files
```bash
curl http://localhost:8080/files
```

### Monitor Replication Queue
```bash
# Semua queue
curl http://localhost:8080/replication-queue

# Filter by status
curl http://localhost:8080/replication-queue?status=PENDING
curl http://localhost:8080/replication-queue?status=COMPLETED

# Filter by node
curl http://localhost:8080/replication-queue?node_id=node-2
```

### Download File
```bash
# Dari node manapun
curl -O http://localhost:8001/files/abc-123-def
curl -O http://localhost:8002/files/abc-123-def
curl -O http://localhost:8003/files/abc-123-def
```

---

## 🎯 Skenario Testing Lengkap

### Skenario 1: Upload → Replikasi → Download

```bash
# 1. Upload
curl -X POST http://localhost:8001/files -F "file=@test.txt"
# Catat file_id dari response

# 2. Cek metadata
curl http://localhost:8080/files

# 3. Download dari berbagai node
curl -O http://localhost:8001/files/{file_id}
curl -O http://localhost:8002/files/{file_id}
curl -O http://localhost:8003/files/{file_id}
```

### Skenario 2: Fault Tolerance → Recovery

```bash
# 1. Stop node-2 dan node-3
# CTRL+C di kedua terminal

# 2. Upload file
curl -X POST http://localhost:8001/files -F "file=@test.txt"

# 3. Cek queue
curl http://localhost:8080/replication-queue

# 4. Hidupkan node-2
cd server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002

# 5. Tunggu 30 detik atau trigger manual
curl -X POST http://localhost:8080/nodes/node-2/recover

# 6. Hidupkan node-3
cd server/storage-node/sn-3
uvicorn main:app --host 0.0.0.0 --port 8003

# 7. Recovery node-3
curl -X POST http://localhost:8080/nodes/node-3/recover

# 8. Verifikasi semua file ada
dir server\storage-node\sn-1\uploads
dir server\storage-node\sn-2\uploads
dir server\storage-node\sn-3\uploads
```

### Skenario 3: Multiple Files Upload

```bash
# Upload beberapa file sekaligus
for i in 1 2 3 4 5; do
  echo "Test file $i" > test$i.txt
  curl -X POST http://localhost:8001/files -F "file=@test$i.txt"
  sleep 1
done

# Cek semua files
curl http://localhost:8080/files
```

---

## 🔍 Troubleshooting

### Error: "connection refused"
- Pastikan service sudah running
- Cek port tidak bentrok dengan aplikasi lain
- Verify dengan: `netstat -an | findstr "8001"`

### Error: "database connection failed"
- Pastikan MySQL running: `net start MySQL80`
- Cek credentials di DB_DSN
- Test koneksi: `mysql -u dfs_user -padmin123 dfs_meta`

### File tidak tereplikasi
- Cek log sn-1 untuk error detail
- Verify node target UP: `curl http://localhost:8002/health`
- Cek replication queue: `curl http://localhost:8080/replication-queue`

### Auto-recovery tidak jalan
- Pastikan naming service running
- Tunggu minimal 30 detik
- Trigger manual jika perlu: `curl -X POST http://localhost:8080/nodes/node-2/recover`

---

## 📝 Catatan Penting

1. **Port yang digunakan:**
   - 8080: Naming Service
   - 8001: Storage Node 1 (Main)
   - 8002: Storage Node 2 (Replica)
   - 8003: Storage Node 3 (Backup)
   - 3306: MySQL

2. **File storage location:**
   - `server/storage-node/sn-1/uploads/`
   - `server/storage-node/sn-2/uploads/`
   - `server/storage-node/sn-3/uploads/`

3. **Auto-recovery interval:** 30 detik (configurable di main.go)

4. **Replication timeout:** 30 detik per node

5. **Database:** MySQL 8.0+ required

---

## ✅ Checklist Testing

- [ ] Semua services running
- [ ] Upload file berhasil
- [ ] File tereplikasi ke 3 node
- [ ] Metadata tersimpan di database
- [ ] Stop 1 node, upload tetap berhasil
- [ ] File masuk replication queue
- [ ] Node kembali UP, auto-recovery jalan
- [ ] File ter-sync ke node yang tadi DOWN
- [ ] Download dari semua node berhasil
- [ ] Monitoring endpoints berfungsi

---

## 🎉 Selesai!

Jika semua checklist di atas berhasil, berarti implementasi automated replication dan fault tolerance sudah berfungsi dengan baik!

**Next Steps:**
1. Testing dengan file berbagai ukuran (small, medium, large)
2. Testing dengan multiple concurrent uploads
3. Testing dengan 2 node DOWN sekaligus
4. Implementasi routing via naming service
5. Integrasi dengan frontend

Untuk detail lebih lanjut, lihat:
- `TESTING_REPLICATION.md` - Testing guide lengkap
- `REPLICATION_FEATURES.md` - Technical documentation
