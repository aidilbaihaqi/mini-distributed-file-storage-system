# Mini Distributed File Storage System  
Sistem penyimpanan file terdistribusi dengan **replication**, **fault tolerance**, dan **smart-node detection**, dikembangkan sebagai proyek akhir mata kuliah Sistem Terdistribusi.

Project ini menggunakan pendekatan arsitektur multi-service:

- **server/** → seluruh backend (Naming Service, Storage Nodes, Database)
- **client/** → Web dashboard (Next.js + React)

---

## 🚀 Fitur Utama

### 🔹 Distributed Storage Architecture
- File disimpan di 3 node penyimpanan:
  - `sn-1` → Main Storage Node  
  - `sn-2`, `sn-3` → Replica Storage Nodes  

### 🔹 Automated Replication
- Setiap upload → direplikasi otomatis ke 2 node cadangan
- Metadata tersimpan di MySQL

### 🔹 Smart Node Detection (Latency-based)
Naming service memilih node terbaik berdasarkan:
1. Status UP
2. Latency terendah
3. Ketersediaan file

### 🔹 Fault Tolerance
Jika 1 node mati:
- Upload & download tetap berjalan
- Sistem memilih node lain secara otomatis

### 🔹 Recovery System
Saat node kembali UP:
- Sistem membaca `replication_queue`
- File disinkronisasi ulang

### 🔹 Dashboard Monitoring
Frontend menampilkan:
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
Lihat **INSTALLATION.md** untuk panduan lengkap.

---

## 🧪 Pengujian

### Test upload:
```
curl -X POST http://localhost:8080/files/upload -F "file=@test.jpg"
```

### Test download:
```
curl -O http://localhost:8080/files/<FILE_KEY>
```

---

## 👥 Pengembang
- Backend Gin / FastAPI  
- DevOps  
- Frontend Next.js  
- Database & Replication Logic  

---

## 📝 Lisensi
Bebas digunakan untuk pembelajaran dan tugas akademik.
