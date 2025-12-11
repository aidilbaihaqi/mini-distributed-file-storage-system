# Mini Distributed File Storage System - Docker Setup

## Quick Start

### 1. Build dan jalankan semua services
```bash
docker-compose up --build
```

### 2. Jalankan di background
```bash
docker-compose up -d --build
```

### 3. Lihat logs
```bash
docker-compose logs -f
```

### 4. Stop semua services
```bash
docker-compose down
```

### 5. Stop dan hapus volumes (reset data)
```bash
docker-compose down -v
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| MySQL | 3306 | Database metadata |
| Naming Service | 8080 | Koordinator & API Gateway |
| Storage Node 1 | 8001 | Storage node |
| Storage Node 2 | 8002 | Storage node |
| Storage Node 3 | 8003 | Storage node |
| Client | 3000 | Web Dashboard |

## Akses

- Dashboard: http://localhost:3000
- Naming Service API: http://localhost:8080
- Storage Node 1: http://localhost:8001
- Storage Node 2: http://localhost:8002
- Storage Node 3: http://localhost:8003

## Test Fault Tolerance

### Matikan satu node
```bash
docker-compose stop storage-node-1
```

### Upload file (akan masuk ke node-2 dan node-3)
Upload via dashboard di http://localhost:3000

### Nyalakan kembali node
```bash
docker-compose start storage-node-1
```
File akan otomatis di-sync dari replication queue.

## Scaling (Tambah Node)

Edit `docker-compose.yml` dan tambahkan service baru, lalu update database:
```sql
INSERT INTO nodes (id, address, status, role) VALUES
('node-4', 'http://storage-node-4:8000', 'DOWN', 'REPLICA');
```
