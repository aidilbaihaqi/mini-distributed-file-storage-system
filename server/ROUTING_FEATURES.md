# Upload/Download Routing & Latency-Based Selection

Dokumentasi lengkap untuk fitur routing via naming service dan latency-based node selection.

---

## 🎯 Overview

Semua operasi file (upload, download, delete) sekarang **harus melalui naming service**. Naming service akan:

1. **Mengukur latency** setiap node secara berkala (30 detik)
2. **Memilih node terbaik** berdasarkan latency terendah
3. **Routing request** ke node yang dipilih
4. **Automatic failover** jika node DOWN

---

## 🚀 Fitur yang Diimplementasikan

### ✅ 1. Upload via Naming Service

**Endpoint:** `POST /upload`

**Cara Kerja:**
1. Client upload file ke naming service (port 8080)
2. Naming service query semua nodes dari database
3. Pilih node dengan latency terendah yang status UP
4. Forward file ke node terpilih
5. Node melakukan replikasi otomatis ke node lain
6. Return response dengan info routing

**Request:**
```bash
curl -X POST http://localhost:8080/upload -F "file=@test.txt"
```

**Response:**
```json
{
  "success": true,
  "file_id": "abc-123-def",
  "original_filename": "test.txt",
  "size_bytes": 1024,
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

### ✅ 2. Download via Naming Service

**Endpoint:** `GET /download/{fileKey}`

**Cara Kerja:**
1. Client request download ke naming service
2. Naming service query file_locations untuk cari node yang punya file
3. Pilih node dengan latency terendah dari yang punya file
4. Forward request ke node terpilih
5. Stream file ke client
6. Add custom headers untuk info routing

**Request:**
```bash
curl -O http://localhost:8080/download/abc-123-def
```

**Response Headers:**
```
X-Routed-From: node-2
X-Node-Latency-Ms: 3
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="test.txt"
```

### ✅ 3. Delete via Naming Service

**Endpoint:** `DELETE /files/{fileKey}`

**Cara Kerja:**
1. Client request delete ke naming service
2. Naming service query file_locations untuk semua node yang punya file
3. Send delete request ke **semua node** yang punya file
4. Update file_locations status menjadi 'DELETED'
5. Return summary (berapa node berhasil/gagal)

**Request:**
```bash
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

### ✅ 4. Latency-Based Node Selection

**Cara Kerja:**
1. Background job di naming service (interval 30 detik)
2. Ping setiap node dengan `GET /health`
3. Measure response time (latency)
4. Update database: `nodes.latency_ms`
5. Saat upload/download, pilih node dengan latency terendah

**Algorithm:**
```go
func selectBestNodeForUpload(nodes []Node) *Node {
    var bestNode *Node
    lowestLatency := int64(9999)
    
    for i := range nodes {
        node := &nodes[i]
        if node.Status == "UP" {
            if node.LatencyMs < lowestLatency {
                lowestLatency = node.LatencyMs
                bestNode = node
            }
        }
    }
    
    return bestNode
}
```

**For Download:**
- Filter nodes yang punya file (dari file_locations)
- Pilih yang latency terendah dari filtered nodes
- Automatic failover ke node lain jika node terpilih DOWN

### ✅ 5. Latency Monitoring

**Endpoint:** `GET /nodes`

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
    "status": "DOWN",
    "latency_ms": 9999
  }
]
```

---

## 📊 Database Schema Update

### Table: nodes

**New Column:**
```sql
latency_ms BIGINT DEFAULT 0
```

**New Index:**
```sql
INDEX idx_status_latency (status, latency_ms)
```

**Purpose:**
- Store measured latency for each node
- Optimize query for selecting best node
- Default 0 for new nodes
- 9999 for unreachable nodes

---

## 🔄 Flow Diagrams

### Upload Flow (via Naming Service)

```
Client
  ↓ POST /upload
Naming Service
  ↓ Query nodes (status=UP, ORDER BY latency_ms)
  ↓ Select best node (lowest latency)
  ↓ Forward file
Storage Node (e.g., node-2)
  ↓ Save file locally
  ↓ Replicate to other nodes
  ↓ Register metadata
Naming Service
  ↓ Return response with routing info
Client
```

### Download Flow (via Naming Service)

```
Client
  ↓ GET /download/{fileKey}
Naming Service
  ↓ Query file_locations (WHERE file_key = ?)
  ↓ Get nodes that have file
  ↓ Select best node (lowest latency)
  ↓ Forward request
Storage Node (e.g., node-3)
  ↓ Stream file
Naming Service
  ↓ Stream to client (with routing headers)
Client
```

### Delete Flow (via Naming Service)

```
Client
  ↓ DELETE /files/{fileKey}
Naming Service
  ↓ Query file_locations
  ↓ Get all nodes that have file
  ↓ Send DELETE to each node (parallel)
Storage Nodes
  ↓ Delete file locally
  ↓ Return success/fail
Naming Service
  ↓ Update file_locations (status=DELETED)
  ↓ Return summary
Client
```

### Latency Measurement (Background Job)

```
Background Job (every 30s)
  ↓ Get all nodes from database
  ↓ For each node:
      ├─ Measure latency (ping /health)
      ├─ Update nodes.latency_ms
      ├─ Check status (UP/DOWN)
      └─ Trigger recovery if needed
```

---

## 🧪 Testing

### Test 1: Upload via Naming Service

```bash
cd server
test-routing.bat
```

**Expected:**
- File uploaded via naming service
- Response shows `selected_node` with lowest latency
- File replicated to all nodes
- Can download from naming service

### Test 2: Latency-Based Selection

```bash
cd server
test-latency.bat
```

**Expected:**
- Shows current latency for all nodes
- Upload routed to node with lowest latency
- Response includes `node_latency_ms`

### Test 3: Download Routing

```bash
# Upload file first
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# Note the file_id from response
# Download via naming service
curl -v http://localhost:8080/download/{file_id} -o downloaded.txt

# Check headers
# Should see: X-Routed-From and X-Node-Latency-Ms
```

### Test 4: Delete via Naming Service

```bash
cd server
test-delete-routing.bat
```

**Expected:**
- File deleted from all nodes
- Response shows `deleted_from: 3`
- file_locations updated to DELETED

### Test 5: Failover

```bash
# Stop node with lowest latency
# Upload file
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# Should route to next best node
# Check response for selected_node
```

---

## 📝 API Endpoints Summary

### Naming Service (Port 8080)

**File Operations (NEW):**
- `POST /upload` - Upload file (routed to best node)
- `GET /download/{fileKey}` - Download file (from best node)
- `DELETE /files/{fileKey}` - Delete file (from all nodes)

**Management:**
- `GET /health` - Health check
- `GET /nodes` - List nodes with latency
- `GET /nodes/check` - Health check all nodes
- `GET /files` - List files with replicas
- `GET /replication-queue` - Monitor queue
- `POST /nodes/{nodeId}/recover` - Manual recovery
- `POST /files/register` - Register metadata (internal)

### Storage Nodes (Ports 8001, 8002, 8003)

**Direct Access (NOT RECOMMENDED):**
- `POST /files` - Upload (use naming service instead)
- `GET /files/{file_id}` - Download (use naming service instead)
- `DELETE /files/{file_id}` - Delete (use naming service instead)
- `GET /health` - Health check (OK for monitoring)

---

## ⚠️ Important Changes

### Breaking Changes

1. **Upload endpoint changed:**
   - Old: `POST http://localhost:8001/files`
   - New: `POST http://localhost:8080/upload`

2. **Download endpoint changed:**
   - Old: `GET http://localhost:8001/files/{file_id}`
   - New: `GET http://localhost:8080/download/{file_id}`

3. **Delete endpoint changed:**
   - Old: `DELETE http://localhost:8001/files/{file_id}`
   - New: `DELETE http://localhost:8080/files/{file_id}`

4. **Database schema:**
   - Added `latency_ms` column to `nodes` table
   - Added index on `(status, latency_ms)`

### Migration Steps

1. **Update database:**
```bash
mysql -u dfs_user -padmin123 dfs_meta < server/naming-service/schema.sql
```

Or manually:
```sql
ALTER TABLE nodes ADD COLUMN latency_ms BIGINT DEFAULT 0;
ALTER TABLE nodes ADD INDEX idx_status_latency (status, latency_ms);
```

2. **Update client code:**
- Change upload URL to naming service
- Change download URL to naming service
- Change delete URL to naming service

3. **Restart services:**
- Restart naming service (to load new code)
- Storage nodes don't need restart

---

## 🎯 Benefits

### 1. Centralized Control
- All file operations go through naming service
- Easier to implement access control
- Centralized logging and monitoring

### 2. Optimal Performance
- Automatic selection of fastest node
- Reduced latency for users
- Better resource utilization

### 3. High Availability
- Automatic failover if node DOWN
- Download from any available node
- No single point of failure (for download)

### 4. Load Balancing
- Distribute load based on latency
- Prevent overloading slow nodes
- Better overall system performance

### 5. Transparency
- Client doesn't need to know node addresses
- Routing info in response/headers
- Easy to add/remove nodes

---

## 🔍 Monitoring

### Check Node Latencies

```bash
curl http://localhost:8080/nodes | jq '.[] | {id, status, latency_ms}'
```

### Check Routing Logs

Look for these in naming service logs:
```
📤 Routing upload to node-2 (latency: 3ms)
📥 Routing download from node-1 (latency: 5ms)
```

### Verify Latency Measurement

```bash
# Check database
mysql -u dfs_user -padmin123 dfs_meta -e "SELECT id, status, latency_ms FROM nodes;"
```

---

## 🐛 Troubleshooting

### Upload fails with "no available nodes"

**Cause:** All nodes are DOWN or latency measurement not yet done

**Solution:**
1. Check nodes status: `curl http://localhost:8080/nodes`
2. Wait 30 seconds for latency measurement
3. Verify nodes are running: `curl http://localhost:8001/health`

### Download returns 404

**Cause:** File not found or no nodes have the file

**Solution:**
1. Check file exists: `curl http://localhost:8080/files`
2. Check file_locations: Query database
3. Verify at least one node with file is UP

### Latency always 9999

**Cause:** Node is unreachable or health endpoint failing

**Solution:**
1. Check node is running
2. Verify health endpoint: `curl http://localhost:8001/health`
3. Check firewall/network

### Wrong node selected

**Cause:** Latency measurement outdated or incorrect

**Solution:**
1. Wait for next measurement cycle (30s)
2. Check logs for latency measurement
3. Manually trigger: Restart naming service

---

## 📈 Performance Considerations

### Latency Measurement
- Interval: 30 seconds (configurable)
- Timeout: 2 seconds per node
- Overhead: Minimal (3 nodes × 2s = 6s max per cycle)

### Upload Routing
- Overhead: ~1-5ms (query + selection)
- Benefit: Route to fastest node
- Net effect: Positive (faster upload)

### Download Routing
- Overhead: ~1-5ms (query + selection)
- Benefit: Route to fastest node
- Net effect: Positive (faster download)

### Caching (Future Enhancement)
- Cache node latencies in memory
- Reduce database queries
- Update cache every 30s

---

## ✅ Implementation Checklist

- [x] Upload routing via naming service
- [x] Download routing via naming service
- [x] Delete routing via naming service
- [x] Latency measurement (background job)
- [x] Latency-based node selection (upload)
- [x] Latency-based node selection (download)
- [x] Database schema update (latency_ms column)
- [x] Response includes routing info
- [x] Headers include routing info (download)
- [x] Testing scripts
- [x] Documentation

---

## 🎉 Conclusion

Routing via naming service dan latency-based selection sudah **COMPLETE** dan **TESTED**.

**Key Features:**
- ✅ All file operations via naming service
- ✅ Automatic node selection based on latency
- ✅ Failover to next best node
- ✅ Centralized control and monitoring
- ✅ Optimal performance

**Ready for:**
- Frontend integration
- Production deployment
- Load testing

---

**Version:** v0.3.0  
**Last Updated:** December 3, 2025  
**Status:** ✅ COMPLETE
