# Installation Guide (Non-Docker, Linux Fresh Install)
Panduan lengkap instalasi proyek Mini Distributed File Storage System pada Linux dari kondisi **kosong total**, hingga semua service dapat dijalankan secara manual tanpa Docker.

---

# 1. 🧰 Install Tools Dasar Linux

```bash
sudo apt update
sudo apt install -y wget curl git unzip vim build-essential
```

---

# 2. ⚙️ Install Python 3 + Pip

```bash
sudo apt install -y python3 python3-pip python3-venv
```

Cek:
```
python3 --version
pip3 --version
```

---

# 3. ⚙️ Install Node.js (Next.js)

```bash
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt install -y nodejs
```

Cek:
```
node -v
npm -v
```

---

# 4. ⚙️ Install Go (Gin Service)

```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

Cek:
```
go version
```

---

# 5. 🗄 Install MySQL

```bash
sudo apt install -y mysql-server
sudo systemctl enable mysql
sudo systemctl start mysql
```

Masuk:

```bash
sudo mysql
```

Buat DB:

```sql
CREATE DATABASE dfs_meta;
CREATE USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';
GRANT ALL PRIVILEGES ON dfs_meta.* TO 'dfs_user'@'localhost';
FLUSH PRIVILEGES;
```

---

# 6. 📦 Clone Repository

```bash
git clone https://github.com/USERNAME/mini-dfs.git
cd mini-dfs
```

---

# 7. ▶️ Jalankan Naming Service (Go)

```bash
cd server/naming-service
go mod tidy
export DB_DSN="dfs_user:admin123@tcp(localhost:3306)/dfs_meta?parseTime=true"
go run main.go
```

Endpoint:
```
http://localhost:8080
```

---

# 8. ▶️ Jalankan Storage Nodes (FastAPI)

Setiap node memiliki folder:

- sn-1  
- sn-2  
- sn-3  

## Node 1
```bash
cd server/storage-node/sn-1
pip3 install -r requirements.txt
python3 app.py
```

## Node 2
```bash
cd ../sn-2
pip3 install -r requirements.txt
python3 app.py
```

## Node 3
```bash
cd ../sn-3
pip3 install -r requirements.txt
python3 app.py
```

---

# 9. ▶️ Jalankan Frontend (Next.js)

```bash
cd client
npm install
npm run dev
```

Akses:
```
http://localhost:3000
```

---

# 10. 🌐 Test

### Upload:
```bash
curl -X POST http://localhost:8080/files/upload   -F "file=@/home/user/test.jpg"
```

### Download:
```bash
curl -O http://localhost:8080/files/<FILE_KEY>
```

### Cek node:
```
http://localhost:8080/nodes
```

---

# 11. 💥 Simulasi Node Failure

Stop salah satu node:
```
CTRL+C
```

Download tetap berjalan memakai node lain.

---

# 12. 🔄 Recovery

Hidupkan kembali node:
```
python3 app.py
```

Sistem akan melakukan re-sync otomatis.

---

# 13. 🧹 Jalankan di Background (opsional)

```bash
nohup go run main.go &
nohup python3 app.py &
nohup npm run dev &
```

---

# 14. 🎉 Instalasi Selesai!
Semua service aplikasi dapat berjalan tanpa Docker:

- Naming Service → port 8080  
- Storage Node → port 8001, 8002, 8003  
- Frontend → port 3000  
