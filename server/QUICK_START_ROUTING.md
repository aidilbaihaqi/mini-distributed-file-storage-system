# Quick Start - Upload/Download Routing & Latency-Based Selection

Panduan cepat untuk menggunakan fitur routing via naming service dan latency-based node selection.

---

## 🎯 Apa yang Berubah?

### Sebelum (v0.2.0):
```bash
# Upload langsung ke storage node
curl -X POST http://localhost:8001/files -F "file=@test.txt"

# Download langsung dari storage node
curl -O http://localhost:8001/files/{file_id}
```

### Sekarang (v0.3.0):
```bash
# Upload via naming service (auto-routing ke node tercepat)
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# Download via naming service (auto-routing ke node tercepat)
curl -O http://localhost:8080/download/{file_id}
```

---

## 📋 Persiapan

### 1. Update Database Schema

```bash
# Masuk ke MySQL
mysql -u dfs_user -padmin123 dfs_meta

# Tambah kolom latency_ms (jika belum ada)
ALTER TABLE nodes ADD COLUMN latency_ms BIGINT DEFAULT 0;
ALTER TABLE nodes ADD INDEX idx_status_latency (status, latency_ms);

# Atau jalankan ulang schema.sql
exit;
mysql -u dfs_user -padmin123 dfs_meta < server/naming-service/schema.sql
```

### 2. Restart Naming Service

```bash
cd server/naming-service
set DB_DSN=dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true
go run main.go
```

Storage nodes tidak perlu restart.

---

## 🚀 Testing Upload/Download Routing

### Test 1: Upload via Naming Service

```bash
cd server

# Buat test file
echo "Test file for routing" > test-routing.txt

# Upload via naming service
curl -X POST http://localhost:8080/upload -F "file=@test-routing.txt"
```

**Expected Response:**
```json
{
  "success": true,
  "file_id": "abc-123-def",
  "original_filename": "test-routing.txt",
  "size_bytes": 23,
  "checksum_sha256": "abc...",
  "node": "main",
  "replication": {
    "successful": 2,
    "failed": 0
  },
  "routed_via": "naming-service",
  "selected_node": "node-1",
  "node_latency_ms": 5
}
```

**Perhatikan:**
- `routed_via`: "naming-service" - Menunjukkan routing via naming service
- `selected_node`: "node-1" - Node yang dipilih (latency terendah)
- `node_latency_ms`: 5 - Latency node terpilih dalam milliseconds

### Test 2: Download via Naming Service

```bash
# Gunakan file_id dari response upload
curl -v http://localhost:8080/download/abc-123-def -o downloaded.txt
```

**Expected Headers:**
```
< HTTP/1.1 200 OK
< Content-Type: application/octet-stream
< Content-Disposition: attachment; filename="test-routing.txt"
< X-Routed-From: node-2
< X-Node-Latency-Ms: 3
```

**Perhatikan:**
- `X-Routed-From`: Node yang melayani download
- `X-Node-Latency-Ms`: Latency node tersebut

### Test 3: Verify Downloaded File

```bash
type downloaded.txt
# Should show: Test file for routing
```

---

## 🎯 Testing Latency-Based Selection

### Test 1: Check Current Latencies

```bash
curl http://localhost:8080/nodes
```

**Response:**
```json
[
  {
    "id": "node-1",
    "address": "http://localhost:8001",
    "status": "UP",
    "role": "MAIN",
    "latency_ms": 5
  },
  {
    "id": "node-2",
    "address": "http://localhost:8002",
    "status": "UP",
    "role": "REPLICA",
    "latency_ms": 3
  },
  {
    "id": "node-3",
    "address": "http://localhost:8003",
    "status": "UP",
    "role": "BACKUP",
    "latency_ms": 7
  }
]
```

**Analisis:**
- node-2 memiliki latency terendah (3ms)
- Upload seharusnya di-route ke node-2

### Test 2: Upload dan Verify Routing

```bash
echo "Test latency selection" > test-latency.txt
curl -X POST http://localhost:8080/upload -F "file=@test-latency.txt"
```

**Check Response:**
```json
{
  "selected_node": "node-2",
  "node_latency_ms": 3
}
```

✅ Confirmed! File di-route ke node dengan latency terendah.

### Test 3: Simulate Node Failure

```bash
# Stop node-2 (CTRL+C di terminal node-2)

# Upload file baru
echo "Test failover" > test-failover.txt
curl -X POST http://localhost:8080/upload -F "file=@test-failover.txt"
```

**Expected:**
- File di-route ke node-1 atau node-3 (yang masih UP)
- Response menunjukkan `selected_node` yang berbeda
- Upload tetap berhasil (fault tolerance)

---

## 🗑️ Testing Delete via Naming Service

### Test 1: Delete File

```bash
# Gunakan file_id dari upload sebelumnya
curl -X DELETE http://localhost:8080/files/abc-123-def
```

**Response:**
```json
{
  "success": true,
  "file_key": "abc-123-def",
  "deleted_from": 3,
  "failed": 0,
  "total_nodes": 3
}
```

**Perhatikan:**
- `deleted_from`: 3 - File dihapus dari 3 node
- `failed`: 0 - Tidak ada yang gagal
- File dihapus dari **semua node** sekaligus

### Test 2: Verify Deletion

```bash
# Try download (should fail)
curl http://localhost:8080/download/abc-123-def
# Expected: 404 Not Found

# Check files list
curl http://localhost:8080/files
# File should not be in the list
```

---

## 📊 Monitoring Latency

### Real-time Latency Check

```bash
# Check every 5 seconds
while true; do
  curl -s http://localhost:8080/nodes | jq '.[] | {id, status, latency_ms}'
  sleep 5
done
```

**Output:**
```json
{
  "id": "node-1",
  "status": "UP",
  "latency_ms": 5
}
{
  "id": "node-2",
  "status": "UP",
  "latency_ms": 3
}
{
  "id": "node-3",
  "status": "UP",
  "latency_ms": 7
}
```

### Check Naming Service Logs

Look for these messages:
```
📤 Routing upload to node-2 (latency: 3ms)
📥 Routing download from node-1 (latency: 5ms)
Node node-2 status changed: DOWN -> UP (latency: 3ms)
```

---

## 🧪 Complete Testing Scenario

### Scenario: Upload → Download → Delete

```bash
# 1. Check node latencies
curl http://localhost:8080/nodes | jq '.[] | {id, latency_ms}'

# 2. Upload file
echo "Complete test scenario" > test-complete.txt
curl -X POST http://localhost:8080/upload -F "file=@test-complete.txt" > upload-response.json

# 3. Extract file_id (manual or use jq)
# For manual: Open upload-response.json and copy file_id
# For jq: 
FILE_ID=$(jq -r '.file_id' upload-response.json)
echo "File ID: $FILE_ID"

# 4. Download file
curl -v http://localhost:8080/download/$FILE_ID -o downloaded-complete.txt

# 5. Verify content
type downloaded-complete.txt

# 6. Check file in database
curl http://localhost:8080/files | jq ".files[] | select(.file_key==\"$FILE_ID\")"

# 7. Delete file
curl -X DELETE http://localhost:8080/files/$FILE_ID

# 8. Verify deletion
curl http://localhost:8080/download/$FILE_ID
# Should return 404
```

---

## 🎯 Advanced Testing

### Test 1: Multiple Uploads (Load Distribution)

```bash
# Upload 10 files
for i in {1..10}; do
  echo "Test file $i" > test$i.txt
  curl -X POST http://localhost:8080/upload -F "file=@test$i.txt" | jq '{file_id, selected_node, node_latency_ms}'
  sleep 1
done
```

**Expected:**
- All uploads routed to node with lowest latency
- Consistent `selected_node` if latencies don't change

### Test 2: Failover Testing

```bash
# Stop node with lowest latency
# Upload should route to next best node

# 1. Check latencies
curl http://localhost:8080/nodes | jq '.[] | {id, latency_ms}' | sort -k2 -n

# 2. Stop node with lowest latency (CTRL+C)

# 3. Upload file
curl -X POST http://localhost:8080/upload -F "file=@test-failover.txt" | jq '{selected_node, node_latency_ms}'

# 4. Verify routing to different node
```

### Test 3: Download from Different Nodes

```bash
# Upload file (will be replicated to all nodes)
curl -X POST http://localhost:8080/upload -F "file=@test-multi.txt" > response.json
FILE_ID=$(jq -r '.file_id' response.json)

# Download multiple times
for i in {1..5}; do
  curl -v http://localhost:8080/download/$FILE_ID -o /dev/null 2>&1 | grep "X-Routed-From"
  sleep 1
done
```

**Expected:**
- All downloads from same node (lowest latency)
- If that node goes DOWN, automatic failover to next best

---

## 🔍 Troubleshooting

### Upload fails with "no available nodes"

**Cause:** All nodes DOWN or latency not measured yet

**Solution:**
```bash
# 1. Check nodes
curl http://localhost:8080/nodes

# 2. Wait 30 seconds for latency measurement

# 3. Verify nodes running
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health
```

### Latency always 9999

**Cause:** Node unreachable

**Solution:**
```bash
# 1. Check node is running
curl http://localhost:8001/health

# 2. Check firewall/network

# 3. Restart node if needed
```

### Wrong node selected

**Cause:** Latency measurement outdated

**Solution:**
```bash
# Wait 30 seconds for next measurement cycle
# Or restart naming service to force immediate measurement
```

---

## 📝 API Reference

### Upload
```bash
POST http://localhost:8080/upload
Content-Type: multipart/form-data

Response:
{
  "file_id": "...",
  "routed_via": "naming-service",
  "selected_node": "node-2",
  "node_latency_ms": 3
}
```

### Download
```bash
GET http://localhost:8080/download/{fileKey}

Headers:
X-Routed-From: node-2
X-Node-Latency-Ms: 3
```

### Delete
```bash
DELETE http://localhost:8080/files/{fileKey}

Response:
{
  "deleted_from": 3,
  "failed": 0,
  "total_nodes": 3
}
```

### Check Nodes
```bash
GET http://localhost:8080/nodes

Response:
[
  {
    "id": "node-1",
    "status": "UP",
    "latency_ms": 5
  }
]
```

---

## ✅ Verification Checklist

- [ ] Database schema updated (latency_ms column)
- [ ] Naming service restarted
- [ ] All storage nodes running
- [ ] Upload via naming service berhasil
- [ ] Response includes routing info
- [ ] Download via naming service berhasil
- [ ] Download headers include routing info
- [ ] Delete via naming service berhasil
- [ ] Latency measurement berfungsi
- [ ] Node selection based on latency
- [ ] Failover berfungsi saat node DOWN

---

## 🎉 Selesai!

Jika semua checklist di atas berhasil, berarti routing via naming service dan latency-based selection sudah berfungsi dengan baik!

**Key Benefits:**
- ✅ Optimal performance (routing ke node tercepat)
- ✅ High availability (automatic failover)
- ✅ Centralized control (semua via naming service)
- ✅ Transparent routing (client tidak perlu tahu node addresses)

**Next Steps:**
- Integrasi dengan frontend
- Testing dengan load yang lebih besar
- Monitoring dan analytics

---

**Version:** v0.3.0  
**Last Updated:** December 3, 2025
