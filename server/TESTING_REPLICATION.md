# Testing Guide - Automated Replication & Fault Tolerance

Panduan untuk testing fitur automated replication dan fault tolerance pada Mini DFS.

---

## Prerequisites

1. MySQL sudah running dan database `dfs_meta` sudah dibuat
2. Jalankan schema.sql untuk membuat tabel:

```bash
cd server/naming-service
mysql -u dfs_user -padmin123 dfs_meta < schema.sql
```

3. Install dependencies untuk sn-1:

```bash
cd server/storage-node/sn-1
pip install httpx
# atau
pip install -r requirements.txt
```

---

## Skenario 1: Normal Upload dengan Semua Node UP

### Langkah:

1. **Start semua services:**

Terminal 1 - Naming Service:
```bash
cd server/naming-service
export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"
go run main.go
```

Terminal 2 - Storage Node 1 (Main):
```bash
cd server/storage-node/sn-1
uvicorn main:app --host 0.0.0.0 --port 8001
```

Terminal 3 - Storage Node 2 (Replica):
```bash
cd server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002
```

Terminal 4 - Storage Node 3 (Backup):
```bash
cd server/storage-node/sn-3
uvicorn main:app --host 0.0.0.0 --port 8003
```

2. **Upload file ke sn-1:**

```bash
curl -X POST http://localhost:8001/files \
  -F "file=@test.jpg"
```

3. **Expected Result:**

Response akan menunjukkan:
```json
{
  "success": true,
  "file_id": "abc-123-def",
  "replication": {
    "total_nodes": 2,
    "successful": 2,
    "failed": 0,
    "details": [
      {"node": "node-2", "success": true},
      {"node": "node-3", "success": true}
    ]
  }
}
```

4. **Verifikasi file ada di semua node:**

```bash
# Cek di node-1
ls server/storage-node/sn-1/uploads/

# Cek di node-2
ls server/storage-node/sn-2/uploads/

# Cek di node-3
ls server/storage-node/sn-3/uploads/
```

5. **Cek metadata di naming service:**

```bash
curl http://localhost:8080/files
```

---

## Skenario 2: Upload dengan 1 Node DOWN (Fault Tolerance)

### Langkah:

1. **Stop node-2:**

Tekan `CTRL+C` di terminal node-2

2. **Upload file baru:**

```bash
curl -X POST http://localhost:8001/files \
  -F "file=@test2.jpg"
```

3. **Expected Result:**

Response akan menunjukkan node-2 gagal:
```json
{
  "success": true,
  "file_id": "xyz-456-abc",
  "replication": {
    "total_nodes": 2,
    "successful": 1,
    "failed": 1,
    "details": [
      {"node": "node-2", "success": false, "error": "..."},
      {"node": "node-3", "success": true}
    ]
  }
}
```

4. **Cek replication queue:**

```bash
curl http://localhost:8080/replication-queue
```

Akan muncul entry untuk node-2 dengan status PENDING.

5. **Verifikasi file:**

```bash
# File ada di node-1 dan node-3
ls server/storage-node/sn-1/uploads/
ls server/storage-node/sn-3/uploads/

# File TIDAK ada di node-2 (karena DOWN)
ls server/storage-node/sn-2/uploads/
```

---

## Skenario 3: Auto-Recovery saat Node Kembali UP

### Langkah:

1. **Hidupkan kembali node-2:**

```bash
cd server/storage-node/sn-2
uvicorn main:app --host 0.0.0.0 --port 8002
```

2. **Tunggu 30 detik** (background job akan auto-detect dan trigger recovery)

3. **Cek log naming service:**

Akan muncul log seperti:
```
Node node-2 status changed: DOWN -> UP
🔄 Triggering recovery for node node-2
✅ Auto-recovered xyz-456-abc to node-2
```

4. **Verifikasi file sudah ada di node-2:**

```bash
ls server/storage-node/sn-2/uploads/
```

File yang tadi gagal sekarang sudah ada!

5. **Cek replication queue:**

```bash
curl http://localhost:8080/replication-queue?status=COMPLETED
```

Status akan berubah menjadi COMPLETED.

---

## Skenario 4: Manual Recovery Trigger

Jika tidak ingin menunggu auto-recovery, bisa trigger manual:

```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```

Response:
```json
{
  "message": "recovery completed",
  "total": 3,
  "success": 3,
  "failed": 0
}
```

---

## Skenario 5: Download File dari Node Manapun

File bisa didownload dari node manapun yang memiliki file tersebut:

```bash
# Download dari node-1
curl -O http://localhost:8001/files/abc-123-def

# Download dari node-2
curl -O http://localhost:8002/files/abc-123-def

# Download dari node-3
curl -O http://localhost:8003/files/abc-123-def
```

---

## Monitoring Endpoints

### 1. Cek status semua nodes:
```bash
curl http://localhost:8080/nodes
```

### 2. Health check nodes:
```bash
curl http://localhost:8080/nodes/check
```

### 3. List semua files:
```bash
curl http://localhost:8080/files
```

### 4. Cek replication queue:
```bash
# Semua queue
curl http://localhost:8080/replication-queue

# Filter by node
curl http://localhost:8080/replication-queue?node_id=node-2

# Filter by status
curl http://localhost:8080/replication-queue?status=PENDING
```

---

## Troubleshooting

### File tidak tereplikasi:

1. Cek log sn-1 untuk error message
2. Cek apakah node target benar-benar UP: `curl http://localhost:8002/health`
3. Cek replication queue untuk detail error

### Auto-recovery tidak jalan:

1. Pastikan naming service running
2. Cek log naming service untuk error
3. Trigger manual recovery: `curl -X POST http://localhost:8080/nodes/node-2/recover`

### Database error:

1. Pastikan MySQL running: `systemctl status mysql`
2. Cek koneksi: `mysql -u dfs_user -padmin123 dfs_meta`
3. Jalankan ulang schema.sql jika perlu

---

## Expected Behavior Summary

✅ **Upload dengan semua node UP**: File tersimpan di 3 node
✅ **Upload dengan 1 node DOWN**: File tersimpan di 2 node, 1 masuk queue
✅ **Node kembali UP**: Auto-recovery dalam 30 detik
✅ **Manual recovery**: Bisa trigger kapan saja
✅ **Download**: Bisa dari node manapun yang punya file
✅ **Fault tolerance**: Sistem tetap jalan meski 1-2 node mati

---

## Next Steps

Setelah testing backend berhasil, langkah selanjutnya:

1. Implementasi routing di naming service (upload/download via naming service)
2. Latency-based node selection
3. Integrasi frontend dengan API backend
4. Load balancing untuk download
