# 🚀 START HERE - Mini DFS Backend

## ✅ Implementation Complete!

Backend untuk **Automated Replication** dan **Fault Tolerance** sudah selesai diimplementasikan.

---

## 📌 Anda Menggunakan WSL (Windows Subsystem for Linux)

Berdasarkan environment Anda, gunakan **script `.sh`** bukan `.bat`

---

## 🎯 Quick Start (3 Langkah)

### 1️⃣ Buka WSL Terminal

```bash
# Dari Windows, ketik di terminal:
wsl

# Atau buka Ubuntu app dari Start Menu
```

### 2️⃣ Setup (Sekali Saja)

```bash
# Navigate ke project
cd ~/projects/dfs/server

# Make scripts executable
chmod +x *.sh

# Verify setup
./verify-setup.sh
```

Jika ada yang FAIL, ikuti instruksi yang muncul.

### 3️⃣ Start & Test

```bash
# Start all services (dalam tmux)
./start-all.sh

# Test upload (di terminal baru atau detach dari tmux)
./test-upload.sh

# Check status
./check-status.sh

# Stop services
./stop-all.sh
```

---

## 📚 Dokumentasi

### Untuk Anda (WSL User):
1. **SETUP_WSL.md** ⭐ - Setup guide khusus WSL
2. **README_LINUX.md** - Linux/WSL commands
3. **TESTING_REPLICATION.md** - Testing scenarios

### API Documentation:
4. **API_DOCUMENTATION.md** ⭐ - Complete API reference
5. **Mini-DFS.postman_collection.json** - Postman collection

### General:
6. **IMPLEMENTATION_SUMMARY.md** - Technical overview
7. **QUICK_START_ID.md** - Quick start (Bahasa Indonesia)
8. **TROUBLESHOOTING.md** - Common issues

---

## 🔧 Available Scripts

### Linux/WSL Scripts (Gunakan Ini!)
```bash
./start-all.sh       # Start all services in tmux
./stop-all.sh        # Stop all services
./check-status.sh    # Check status
./test-upload.sh     # Quick upload test
./verify-setup.sh    # Verify requirements
./cleanup.sh         # Clean & reset
```

### Windows Scripts (Jangan Gunakan di WSL)
```
start-all.bat        # Untuk Windows CMD
check-status.bat     # Untuk Windows CMD
test-upload.bat      # Untuk Windows CMD
```

---

## 🧪 Testing Scenarios

### 1. Normal Upload (All Nodes UP)
```bash
./test-upload.sh
```
Expected: File di 3 node

### 2. Fault Tolerance (1 Node DOWN)
```bash
# Stop node-2
tmux attach -t mini-dfs
# Ctrl+B then 2, then Ctrl+C
# Ctrl+B then D

# Upload via Naming Service
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# Check queue
curl http://localhost:8080/replication-queue?status=PENDING
```
Expected: File di 2 node, 1 masuk queue

### 3. Auto-Recovery
```bash
# Start node-2 kembali
tmux attach -t mini-dfs
# Ctrl+B then 2
# Up Arrow, Enter (restart uvicorn)
# Ctrl+B then D

# Wait 30 seconds
# Check logs
tmux attach -t mini-dfs
# Ctrl+B then 0 (naming service logs)
```
Expected: "Auto-recovered ... to node-2"

### 4. Manual Recovery
```bash
curl -X POST http://localhost:8080/nodes/node-2/recover
```
Expected: {"success": X, "failed": 0}

---

## 🎯 Tmux Commands

```bash
# Attach to session
tmux attach -t mini-dfs

# Switch windows
Ctrl+B then 0  # Naming Service
Ctrl+B then 1  # Storage Node 1
Ctrl+B then 2  # Storage Node 2
Ctrl+B then 3  # Storage Node 3

# Detach (keep services running)
Ctrl+B then D

# Kill session
tmux kill-session -t mini-dfs
# Or use: ./stop-all.sh
```

---

## 🔍 API Endpoints (via Naming Service)

```bash
# Health check
curl http://localhost:8080/health

# Upload file ⭐
curl -X POST http://localhost:8080/upload -F "file=@test.txt"

# Download file ⭐
curl -O -J http://localhost:8080/download/{FILE_KEY}

# Delete file ⭐
curl -X DELETE http://localhost:8080/files/{FILE_KEY}

# List files
curl http://localhost:8080/files

# Check nodes (with latency)
curl http://localhost:8080/nodes

# Check queue
curl http://localhost:8080/replication-queue

# Manual recovery
curl -X POST http://localhost:8080/nodes/node-2/recover
```

⚠️ **Semua request harus melalui port 8080 (Naming Service)**

---

## 🐛 Troubleshooting

### "Permission denied"
```bash
chmod +x *.sh
```

### "MySQL connection failed"
```bash
sudo service mysql start
```

### "Port already in use"
```bash
./stop-all.sh
```

### Lihat log error
```bash
tmux attach -t mini-dfs
# Switch ke window yang error
# Ctrl+B then [  (scroll mode)
# Q to exit
```

---

## ✅ Verification Checklist

Sebelum testing, pastikan:

- [ ] WSL terminal terbuka
- [ ] Di directory: `~/projects/dfs/server`
- [ ] Scripts executable: `chmod +x *.sh`
- [ ] Setup verified: `./verify-setup.sh` passed
- [ ] MySQL running
- [ ] Database schema loaded
- [ ] Dependencies installed

Jika semua ✓:

```bash
./start-all.sh
```

---

## 🎉 What's Implemented

✅ **Gateway Routing** - Semua request melalui Naming Service (8080)  
✅ **Latency-based Selection** - Pilih node dengan latency terendah  
✅ **Automated Replication** - Upload otomatis direplikasi ke semua node  
✅ **Fault Tolerance** - Upload tetap berhasil meski node DOWN  
✅ **Auto-Recovery** - Node kembali UP, file auto-sync (30s)  
✅ **Manual Recovery** - Trigger recovery on-demand  
✅ **Replication Queue** - Database tracking untuk replikasi  
✅ **Metadata Management** - File metadata di MySQL  
✅ **Monitoring APIs** - Check status, files, queue, nodes  

---

## ⏳ Not Implemented Yet

⏳ Frontend integration  
⏳ Authentication & Authorization  
⏳ File versioning  

---

## 💡 Pro Tips

### 1. Create Aliases

Add to `~/.bashrc`:
```bash
alias dfs='cd ~/projects/dfs/server'
alias dfs-start='cd ~/projects/dfs/server && ./start-all.sh'
alias dfs-stop='cd ~/projects/dfs/server && ./stop-all.sh'
alias dfs-status='cd ~/projects/dfs/server && ./check-status.sh'
alias dfs-test='cd ~/projects/dfs/server && ./test-upload.sh'
```

Then:
```bash
source ~/.bashrc
dfs-start  # Quick start!
```

### 2. Auto-start MySQL

Add to `~/.bashrc`:
```bash
sudo service mysql start 2>/dev/null
```

### 3. View Logs

```bash
# Attach to tmux
tmux attach -t mini-dfs

# Switch windows to see logs
Ctrl+B then 0/1/2/3

# Scroll logs
Ctrl+B then [
# Use arrows or Page Up/Down
# Press Q to exit
```

---

## 📞 Need Help?

1. **Setup issues:** See **SETUP_WSL.md**
2. **Testing guide:** See **TESTING_REPLICATION.md**
3. **Common errors:** See **TROUBLESHOOTING.md**
4. **Technical details:** See **IMPLEMENTATION_SUMMARY.md**

---

## 🚀 Ready to Start?

```bash
# 1. Open WSL
wsl

# 2. Go to project
cd ~/projects/dfs/server

# 3. Make executable
chmod +x *.sh

# 4. Verify
./verify-setup.sh

# 5. Start!
./start-all.sh

# 6. Test!
./test-upload.sh

# 7. Check!
./check-status.sh
```

---

**Environment:** WSL Ubuntu  
**Status:** ✅ Ready for Testing  
**Last Updated:** December 3, 2025

---

**🎯 NEXT STEP:** Run `./verify-setup.sh` untuk memulai!
