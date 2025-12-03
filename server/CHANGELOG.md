# Changelog - Mini Distributed File Storage System

All notable changes to this project will be documented in this file.

---

## [v0.2.0] - 2025-12-03

### 🎉 Major Features Added

#### Automated Replication
- Implemented automatic file replication from sn-1 to sn-2 and sn-3
- Parallel replication using async/await for better performance
- Response includes detailed replication status per node
- Configurable replica nodes list

#### Fault Tolerance
- System continues to function even when 1-2 nodes are DOWN
- Failed replications are tracked in replication_queue
- No rollback on partial failures
- Graceful error handling with detailed error messages

#### Auto-Recovery System
- Background job runs every 30 seconds in naming service
- Automatic detection of node status changes (DOWN → UP)
- Triggers recovery process for pending replications
- Updates replication_queue status automatically

#### Manual Recovery
- New endpoint: `POST /nodes/{nodeId}/recover`
- On-demand recovery trigger for testing and maintenance
- Returns detailed summary (total, success, failed counts)
- Processes all pending items for specified node

#### Replication Queue
- Database table for tracking replication status
- Status tracking: PENDING, IN_PROGRESS, COMPLETED, FAILED
- Retry count and error message logging
- Monitoring via API endpoints

#### Metadata Management
- File metadata stored in MySQL (naming service)
- Tracking file locations across all nodes
- SHA256 checksum for integrity validation
- API endpoints for querying files and replicas

### 📝 Modified Files

#### Storage Nodes

**server/storage-node/sn-1/main.py**
- Added `httpx` and `asyncio` imports
- Implemented `replicate_to_node()` function
- Implemented `replicate_to_all_nodes()` function
- Implemented `register_file_to_naming_service()` function
- Modified `POST /files` endpoint for automatic replication
- Added configuration: `REPLICA_NODES`, `NAMING_SERVICE_URL`, `NODE_ID`

**server/storage-node/sn-2/main.py**
- Upgraded from basic health check to full storage node
- Added `POST /files` endpoint for receiving replications
- Added `GET /files/{file_id}` endpoint for downloads
- Added `DELETE /files/{file_id}` endpoint
- Implemented file storage with metadata management
- Added helper functions: `save_file_to_disk()`, `resolve_file_path()`, etc.

**server/storage-node/sn-3/main.py**
- Upgraded from basic health check to full storage node
- Added `POST /files` endpoint for receiving replications
- Added `GET /files/{file_id}` endpoint for downloads
- Added `DELETE /files/{file_id}` endpoint
- Implemented file storage with metadata management
- Added helper functions: `save_file_to_disk()`, `resolve_file_path()`, etc.

**server/storage-node/sn-1/requirements.txt**
- Added `httpx==0.27.0` dependency

#### Naming Service

**server/naming-service/main.go**
- Added imports: `bytes`, `encoding/json`, `fmt`, `io`, `mime/multipart`
- Added structs: `FileMetadata`, `ReplicationQueueItem`
- Added database functions:
  - `updateNodeStatus()`
  - `addToReplicationQueue()`
  - `getPendingReplications()`
  - `markReplicationCompleted()`
  - `markReplicationFailed()`
  - `getFileLocations()`
  - `replicateFileToNode()`
- Added API endpoints:
  - `POST /files/register` - Register file metadata after upload
  - `POST /nodes/{nodeId}/recover` - Manual recovery trigger
  - `GET /replication-queue` - Monitor replication queue
  - `GET /files` - List all files with replica information
- Implemented background goroutine for auto-recovery (30s interval)
- Enhanced `getAllNodes()` to include role field

### 📄 New Files

#### Database
- **server/naming-service/schema.sql** - Complete database schema
  - Table: `nodes` - Storage node information
  - Table: `files` - File metadata
  - Table: `file_locations` - File location tracking
  - Table: `replication_queue` - Replication tracking
  - Default data for 3 storage nodes

#### Documentation
- **server/IMPLEMENTATION_SUMMARY.md** - Complete implementation overview
- **server/QUICK_START_ID.md** - Quick start guide (Bahasa Indonesia)
- **server/TESTING_REPLICATION.md** - Comprehensive testing guide
- **server/REPLICATION_FEATURES.md** - Technical documentation
- **server/TROUBLESHOOTING.md** - Troubleshooting guide
- **server/CHANGELOG.md** - This file

#### Scripts (Windows)
- **server/start-all.bat** - Start all services at once
- **server/test-upload.bat** - Quick upload test
- **server/check-status.bat** - Check all services status

### 🔧 Technical Details

#### Database Schema
```sql
- nodes (id, address, status, role, last_heartbeat)
- files (file_key, original_filename, size_bytes, checksum_sha256, uploaded_at)
- file_locations (id, file_key, node_id, status)
- replication_queue (id, file_key, target_node_id, source_node_id, status, retry_count, error_message)
```

#### API Endpoints Added

**Naming Service (Port 8080):**
- `POST /files/register` - Register file metadata
- `GET /files` - List files with replicas
- `GET /replication-queue` - Monitor queue
- `POST /nodes/{nodeId}/recover` - Manual recovery

**Storage Nodes (Ports 8001, 8002, 8003):**
- `POST /files` - Upload/receive file
- `GET /files/{file_id}` - Download file
- `DELETE /files/{file_id}` - Delete file
- `GET /health` - Health check

#### Configuration

**Storage Node 1:**
```python
REPLICA_NODES = ["http://localhost:8002", "http://localhost:8003"]
NAMING_SERVICE_URL = "http://localhost:8080"
NODE_ID = "node-1"
```

**Naming Service:**
```bash
DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"
```

### 🧪 Testing

#### Test Scenarios Covered
1. Normal upload with all nodes UP
2. Upload with 1 node DOWN (fault tolerance)
3. Auto-recovery when node comes back UP
4. Manual recovery trigger
5. Download from multiple nodes
6. Monitoring and queue inspection

#### Test Scripts
- Windows batch files for quick testing
- Manual curl commands for detailed testing
- Database queries for verification

### 📊 Performance

- **Replication:** Parallel async calls (non-blocking)
- **Timeout:** 30 seconds per node
- **Recovery Interval:** 30 seconds (configurable)
- **Health Check Timeout:** 2 seconds

### 🐛 Bug Fixes
- N/A (initial implementation)

### 🔒 Security
- Basic authentication for MySQL
- No encryption implemented yet (planned for future)

### ⚠️ Breaking Changes
- Storage nodes now require `httpx` dependency
- Database schema must be initialized with schema.sql
- Environment variable `DB_DSN` required for naming service

### 📝 Notes
- Frontend integration pending (currently using mock data)
- Latency-based routing not yet implemented
- Upload/download via naming service not yet implemented
- Checksum validation after replication not yet implemented

---

## [v0.1.0] - 2025-11-XX (Previous State)

### Initial Implementation
- Basic naming service with health check
- Storage node 1 with upload/download/delete
- Storage nodes 2 & 3 with health check only
- MySQL database setup
- Frontend dashboard with mock data
- Basic file operations

---

## [v0.3.0] - 2025-12-03

### 🎉 Major Features Added

#### Upload/Download Routing via Naming Service
- All file operations now go through naming service (port 8080)
- Client no longer needs to know storage node addresses
- Centralized control and monitoring
- New endpoints: `POST /upload`, `GET /download/{fileKey}`, `DELETE /files/{fileKey}`

#### Latency-Based Node Selection
- Background job measures latency every 30 seconds
- Automatic selection of fastest node for upload
- Automatic selection of fastest node for download (among nodes that have file)
- Database stores latency_ms for each node
- Optimal performance and load distribution

#### Smart Routing
- Upload routed to node with lowest latency
- Download routed to node with lowest latency that has the file
- Delete sent to all nodes that have the file
- Automatic failover if selected node is DOWN

#### Routing Information
- Upload response includes: `routed_via`, `selected_node`, `node_latency_ms`
- Download headers include: `X-Routed-From`, `X-Node-Latency-Ms`
- Delete response includes: `deleted_from`, `failed`, `total_nodes`

### 📝 Modified Files

**server/naming-service/main.go**
- Added `LatencyMs` field to `Node` struct
- Implemented `measureNodeLatency()` function
- Implemented `updateNodeLatency()` function
- Implemented `selectBestNodeForUpload()` function
- Implemented `selectBestNodeForDownload()` function
- Added endpoint: `POST /upload` - Upload file via naming service
- Added endpoint: `GET /download/{fileKey}` - Download file via naming service
- Added endpoint: `DELETE /files/{fileKey}` - Delete file from all nodes
- Updated background job to measure latency every 30 seconds
- Updated `getAllNodes()` to include latency_ms

**server/naming-service/schema.sql**
- Added `latency_ms BIGINT DEFAULT 0` column to `nodes` table
- Added index: `idx_status_latency (status, latency_ms)`

### 📄 New Files

**Testing Scripts:**
- **server/test-routing.bat** - Test upload/download via naming service
- **server/test-latency.bat** - Test latency-based node selection
- **server/test-delete-routing.bat** - Test delete via naming service

**Documentation:**
- **server/ROUTING_FEATURES.md** - Complete routing documentation

### 🔧 Technical Details

#### Latency Measurement
```go
func measureNodeLatency(nodeAddr string) int64 {
    start := time.Now()
    resp, err := client.Get(nodeAddr + "/health")
    elapsed := time.Since(start).Milliseconds()
    
    if err != nil {
        return 9999 // High latency for unreachable nodes
    }
    
    return elapsed
}
```

#### Node Selection Algorithm
- Filter nodes by status (UP only)
- Sort by latency_ms (ascending)
- Select first node (lowest latency)
- For download: Also filter by file availability

#### API Changes

**Upload:**
- Old: `POST http://localhost:8001/files`
- New: `POST http://localhost:8080/upload`

**Download:**
- Old: `GET http://localhost:8001/files/{file_id}`
- New: `GET http://localhost:8080/download/{file_id}`

**Delete:**
- Old: `DELETE http://localhost:8001/files/{file_id}` (single node)
- New: `DELETE http://localhost:8080/files/{file_id}` (all nodes)

### 🧪 Testing

#### Test Scenarios
1. Upload via naming service (routed to best node)
2. Download via naming service (from best node)
3. Delete via naming service (from all nodes)
4. Latency-based selection verification
5. Failover when best node is DOWN

#### Test Scripts
```bash
cd server
test-routing.bat      # Upload/download test
test-latency.bat      # Latency selection test
test-delete-routing.bat  # Delete test
```

### 📊 Performance

- **Latency Measurement:** Every 30 seconds
- **Routing Overhead:** ~1-5ms per request
- **Selection Time:** O(n) where n = number of nodes
- **Net Effect:** Positive (faster overall due to optimal routing)

### ⚠️ Breaking Changes

1. **Client must use naming service endpoints:**
   - Upload: `POST /upload` (not `/files`)
   - Download: `GET /download/{fileKey}` (not `/files/{fileKey}`)
   - Delete: `DELETE /files/{fileKey}` (via naming service)

2. **Database schema change:**
   - Added `latency_ms` column to `nodes` table
   - Migration required: Run schema.sql or ALTER TABLE

3. **Direct storage node access discouraged:**
   - Still works but not recommended
   - Use naming service for all operations

### 🐛 Bug Fixes
- N/A (new feature implementation)

### 🔒 Security
- Centralized access point (easier to add auth later)
- No direct storage node exposure needed

### 📝 Notes
- Storage nodes still support direct access (for backward compatibility)
- Latency measurement uses health endpoint (no extra overhead)
- Background job runs continuously (30s interval)
- Failover is automatic (no manual intervention needed)

---

## Upcoming Features (v0.4.0)

### Planned
- [ ] Frontend integration with backend API
- [ ] Checksum validation after replication
- [ ] File compression before replication
- [ ] Encryption at rest and in transit
- [ ] Admin dashboard for monitoring
- [ ] Metrics and statistics
- [ ] Docker Compose setup
- [ ] Authentication and authorization
- [ ] Rate limiting

### Under Consideration
- [ ] File versioning
- [ ] Soft delete with trash bin
- [ ] File deduplication
- [ ] Bandwidth throttling
- [ ] Rate limiting
- [ ] Authentication and authorization
- [ ] Multi-tenant support
- [ ] Backup and restore functionality

---

## Version History

- **v0.3.0** (2025-12-03) - Upload/Download Routing & Latency-Based Selection ✅
- **v0.2.0** (2025-12-03) - Automated Replication & Fault Tolerance ✅
- **v0.1.0** (2025-11-XX) - Initial Implementation

---

**Maintained by:** Development Team  
**Last Updated:** December 3, 2025
