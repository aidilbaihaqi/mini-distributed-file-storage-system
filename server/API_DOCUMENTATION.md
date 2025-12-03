# API Documentation - Mini Distributed File Storage System

Dokumentasi lengkap untuk semua API endpoints yang tersedia di Mini DFS.

---

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Base URLs](#base-urls)
3. [Public APIs (Naming Service)](#public-apis-naming-service)
4. [Internal APIs (Storage Nodes)](#internal-apis-storage-nodes)
5. [Error Responses](#error-responses)
6. [Examples](#examples)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         CLIENT                               │
│              (Frontend / curl / Postman)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP Requests
                         ▼
              ┌──────────────────────┐
              │   NAMING SERVICE     │  ← Gateway (Port 8080)
              │   (Public API)       │
              │                      │
              │  • /upload           │
              │  • /download/:key    │
              │  • /files/:key       │
              │  • /files            │
              │  • /nodes            │
              │  • /health           │
              └──────────┬───────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ Storage Node 1 │ │ Storage Node 2 │ │ Storage Node 3 │
│ (Internal API) │ │ (Internal API) │ │ (Internal API) │
│   Port 8001    │ │   Port 8002    │ │   Port 8003    │
└────────────────┘ └────────────────┘ └────────────────┘
```

### ⚠️ Important Notes

1. **Client harus mengakses melalui Naming Service (Port 8080)**
2. **Storage Nodes (8001, 8002, 8003) adalah internal API** - tidak boleh diakses langsung oleh client
3. **Naming Service bertindak sebagai Gateway** - routing, load balancing, dan fault tolerance

---

## Base URLs

### Public API (Untuk Client)

| Service | URL | Description |
|---------|-----|-------------|
| **Naming Service** | `http://localhost:8080` | Gateway - Semua request client harus melalui sini |

### Internal API (Hanya untuk Naming Service)

| Service | URL | Description |
|---------|-----|-------------|
| Storage Node 1 | `http://localhost:8001` | Internal - Main storage |
| Storage Node 2 | `http://localhost:8002` | Internal - Replica storage |
| Storage Node 3 | `http://localhost:8003` | Internal - Backup storage |

---

## Public APIs (Naming Service)

Base URL: `http://localhost:8080`

**Semua request dari client harus menggunakan endpoints ini.**

---

### 1. Health Check

Check if naming service is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "UP",
  "service": "naming-service",
  "hostname": "server-name"
}
```

**Example:**
```bash
curl http://localhost:8080/health
```

---

### 2. Upload File ⭐

Upload file melalui naming service. File akan otomatis di-route ke node terbaik dan direplikasi.

**Endpoint:** `POST /upload`

**Request:**
- Method: `POST`
- Content-Type: `multipart/form-data`
- Body: Form data dengan field `file`

**Response:**
```json
{
  "success": true,
  "file_id": "abc-123-def-456",
  "stored_name": "abc-123-def-456.pdf",
  "original_filename": "document.pdf",
  "size_bytes": 1024000,
  "checksum_sha256": "a1b2c3d4e5f6...",
  "node": "main",
  "replication": {
    "total_nodes": 2,
    "successful": 2,
    "failed": 0,
    "details": [
      {"node": "node-2", "success": true},
      {"node": "node-3", "success": true}
    ]
  },
  "routed_via": "naming-service",
  "selected_node": "node-1",
  "node_latency_ms": 5
}
```

**Example:**
```bash
# Upload file
curl -X POST http://localhost:8080/upload -F "file=@document.pdf"

# Upload dengan output ke file
curl -X POST http://localhost:8080/upload -F "file=@document.pdf" > response.json
```

**Notes:**
- Naming service akan memilih node dengan latency terendah
- File otomatis direplikasi ke semua node yang UP
- Jika ada node DOWN, file masuk replication queue untuk recovery nanti

---

### 3. Download File ⭐

Download file melalui naming service. Request akan di-route ke node terbaik yang memiliki file.

**Endpoint:** `GET /download/{file_key}`

**Path Parameters:**
- `file_key` - File ID yang didapat dari response upload

**Response:**
- Content-Type: Sesuai tipe file
- Content-Disposition: `attachment; filename="original_filename.ext"`
- Body: Binary file data
- Headers:
  - `X-Routed-From`: Node ID yang melayani request
  - `X-Node-Latency-Ms`: Latency node dalam milliseconds

**Example:**
```bash
# Download file
curl -O -J http://localhost:8080/download/abc-123-def-456

# Download dengan nama custom
curl -o myfile.pdf http://localhost:8080/download/abc-123-def-456
```

**Notes:**
- Naming service akan memilih node dengan latency terendah yang memiliki file
- Jika node utama DOWN, otomatis failover ke node lain
- Response header menunjukkan node mana yang melayani request

---

### 4. Delete File ⭐

Hapus file dari semua nodes melalui naming service.

**Endpoint:** `DELETE /files/{file_key}`

**Path Parameters:**
- `file_key` - File ID yang akan dihapus

**Response:**
```json
{
  "success": true,
  "file_key": "abc-123-def-456",
  "deleted_from": 3,
  "failed": 0,
  "total_nodes": 3
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/files/abc-123-def-456
```

**Notes:**
- File akan dihapus dari SEMUA nodes yang memiliki file tersebut
- Response menunjukkan berapa node yang berhasil dihapus

---

### 5. List All Files

Get list of all files dengan informasi replicas.

**Endpoint:** `GET /files`

**Response:**
```json
{
  "files": [
    {
      "file_key": "abc-123-def",
      "original_filename": "document.pdf",
      "size_bytes": 1024000,
      "checksum_sha256": "a1b2c3d4...",
      "uploaded_at": "2025-12-03T10:30:00Z",
      "replicas": ["node-1", "node-2", "node-3"]
    }
  ],
  "count": 1
}
```

**Example:**
```bash
curl http://localhost:8080/files
```

---

### 6. List All Nodes

Get list of all registered storage nodes dengan status dan latency.

**Endpoint:** `GET /nodes`

**Response:**
```json
[
  {
    "id": "node-1",
    "address": "http://localhost:8001",
    "status": "UP",
    "role": "MAIN",
    "last_heartbeat": "2025-12-03T10:30:00Z",
    "latency_ms": 5
  },
  {
    "id": "node-2",
    "address": "http://localhost:8002",
    "status": "UP",
    "role": "REPLICA",
    "last_heartbeat": "2025-12-03T10:30:00Z",
    "latency_ms": 8
  },
  {
    "id": "node-3",
    "address": "http://localhost:8003",
    "status": "DOWN",
    "role": "BACKUP",
    "last_heartbeat": "2025-12-03T10:25:00Z",
    "latency_ms": 9999
  }
]
```

**Example:**
```bash
curl http://localhost:8080/nodes
```

---

### 7. Health Check All Nodes

Perform real-time health check pada semua storage nodes.

**Endpoint:** `GET /nodes/check`

**Response:**
```json
{
  "checked_at": "2025-12-03T10:30:00Z",
  "nodes": [
    {"id": "node-1", "address": "http://localhost:8001", "status": "UP"},
    {"id": "node-2", "address": "http://localhost:8002", "status": "UP"},
    {"id": "node-3", "address": "http://localhost:8003", "status": "DOWN"}
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/nodes/check
```

---

### 8. Get Replication Queue

Monitor status replication queue.

**Endpoint:** `GET /replication-queue`

**Query Parameters:**
- `node_id` (optional) - Filter by target node ID
- `status` (optional) - Filter by status: `PENDING`, `IN_PROGRESS`, `COMPLETED`, `FAILED`

**Response:**
```json
{
  "items": [
    {
      "id": 1,
      "file_key": "abc-123-def",
      "target_node_id": "node-2",
      "source_node_id": "node-1",
      "status": "PENDING",
      "retry_count": 0,
      "last_attempt": null,
      "created_at": "2025-12-03T10:30:00Z"
    }
  ],
  "count": 1
}
```

**Examples:**
```bash
# All queue items
curl http://localhost:8080/replication-queue

# Filter by status
curl http://localhost:8080/replication-queue?status=PENDING

# Filter by node
curl http://localhost:8080/replication-queue?node_id=node-2

# Both filters
curl "http://localhost:8080/replication-queue?node_id=node-2&status=PENDING"
```

---

### 9. Manual Recovery

Trigger manual recovery untuk node tertentu.

**Endpoint:** `POST /nodes/{nodeId}/recover`

**Path Parameters:**
- `nodeId` - Node ID yang akan di-recover (e.g., "node-2")

**Response:**
```json
{
  "message": "recovery completed",
  "total": 5,
  "success": 4,
  "failed": 1,
  "pending_items": [...]
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```

---

## Internal APIs (Storage Nodes)

⚠️ **WARNING: Endpoints ini hanya untuk internal use oleh Naming Service. Client tidak boleh mengakses langsung!**

### Storage Node Endpoints (Internal)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/files` | Receive file from naming service |
| GET | `/files/{id}` | Serve file to naming service |
| DELETE | `/files/{id}` | Delete file |

**Note:** Dalam production, storage nodes sebaiknya:
- Hanya bisa diakses dari internal network
- Menggunakan firewall untuk block akses dari luar
- Atau menggunakan authentication token dari naming service

---

## Error Responses

### Common Error Codes

| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid input |
| 404 | Not Found - File atau resource tidak ditemukan |
| 500 | Internal Server Error - Server error |
| 503 | Service Unavailable - Tidak ada node yang tersedia |

### Error Response Format

```json
{
  "error": "Error message description"
}
```

### Examples

**File Not Found (404):**
```json
{
  "error": "file not found or no available nodes"
}
```

**No Available Nodes (503):**
```json
{
  "error": "no available nodes"
}
```

**Upload Failed (500):**
```json
{
  "error": "gagal upload ke node: connection refused"
}
```

---

## Examples

### Complete Upload & Download Flow

```bash
# 1. Upload file via naming service
curl -X POST http://localhost:8080/upload \
  -F "file=@document.pdf" \
  > upload_response.json

# 2. Extract file_key from response
FILE_KEY=$(cat upload_response.json | grep -o '"file_id":"[^"]*"' | cut -d'"' -f4)
echo "File Key: $FILE_KEY"

# 3. Check file in system
curl http://localhost:8080/files

# 4. Download file via naming service
curl -O -J http://localhost:8080/download/$FILE_KEY

# 5. Delete file via naming service
curl -X DELETE http://localhost:8080/files/$FILE_KEY
```

---

### Fault Tolerance Test

```bash
# 1. Upload file (all nodes UP)
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# 2. Stop node-2 (in another terminal)
# CTRL+C on node-2 terminal

# 3. Upload another file (node-2 DOWN)
curl -X POST http://localhost:8080/upload -F "file=@test2.txt"

# 4. Check replication queue
curl http://localhost:8080/replication-queue?status=PENDING

# 5. Start node-2 again
# uvicorn main:app --port 8002

# 6. Wait 30 seconds for auto-recovery
# Or trigger manual recovery
curl -X POST http://localhost:8080/nodes/node-2/recover

# 7. Verify queue is processed
curl http://localhost:8080/replication-queue?status=COMPLETED
```

---

### Monitoring Workflow

```bash
# Check all services
curl http://localhost:8080/health

# Check nodes status with latency
curl http://localhost:8080/nodes

# Real-time health check
curl http://localhost:8080/nodes/check

# List all files
curl http://localhost:8080/files

# Check replication queue
curl http://localhost:8080/replication-queue
```

---

### Batch Upload

```bash
# Upload multiple files
for file in *.txt; do
  echo "Uploading $file..."
  curl -X POST http://localhost:8080/upload -F "file=@$file"
  sleep 1
done

# Check total files
curl http://localhost:8080/files | grep -o '"count":[0-9]*'
```

---

## API Summary

### Public Endpoints (Client Use)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check naming service |
| POST | `/upload` | **Upload file** |
| GET | `/download/{key}` | **Download file** |
| DELETE | `/files/{key}` | **Delete file** |
| GET | `/files` | List all files |
| GET | `/nodes` | List all nodes |
| GET | `/nodes/check` | Health check all nodes |
| GET | `/replication-queue` | Monitor replication queue |
| POST | `/nodes/{id}/recover` | Manual recovery |

### Key Points

1. ✅ **Upload:** `POST http://localhost:8080/upload`
2. ✅ **Download:** `GET http://localhost:8080/download/{file_key}`
3. ✅ **Delete:** `DELETE http://localhost:8080/files/{file_key}`
4. ❌ **Jangan akses storage nodes langsung** (8001, 8002, 8003)

---

## Response Headers

### Download Response Headers

| Header | Description |
|--------|-------------|
| `Content-Type` | MIME type file |
| `Content-Disposition` | Filename untuk download |
| `X-Routed-From` | Node ID yang melayani request |
| `X-Node-Latency-Ms` | Latency node dalam ms |

---

## Rate Limits

Currently no rate limits. Recommended for production:
- Upload: 10 requests/minute per IP
- Download: 100 requests/minute per IP
- API calls: 60 requests/minute per IP

---

## Authentication

Currently no authentication. Recommended for production:
- API Key authentication
- JWT tokens
- OAuth 2.0

---

## CORS

Untuk frontend integration, tambahkan CORS middleware di naming service:

```go
import "github.com/gin-contrib/cors"

r.Use(cors.Default())
```

---

## Versioning

Current API Version: **v0.2.0**

---

**Last Updated:** December 3, 2025  
**API Version:** v0.2.0  
**Status:** Production Ready (Backend Only)
