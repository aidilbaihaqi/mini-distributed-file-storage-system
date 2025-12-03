# Setup Guide untuk WSL (Windows Subsystem for Linux)

## 🎯 Jawaban Singkat

**Q: File `.bat` dijalankan di terminal Linux atau Windows?**  
**A: File `.bat` untuk Windows CMD/PowerShell. Untuk WSL/Linux, gunakan file `.sh`**

---

## 📌 Environment Anda

Berdasarkan path `\\wsl.localhost\Ubuntu\home\gristof\projects\dfs`, Anda menggunakan:
- **WSL (Windows Subsystem for Linux)**
- **Ubuntu** di dalam Windows

---

## 🔧 Cara Setup

### 1. Buka WSL Terminal

```bash
# Dari Windows, buka WSL terminal
wsl

# Atau buka Ubuntu app dari Start Menu
```

### 2. Navigate ke Project

```bash
cd ~/projects/dfs/server
```

### 3. Make Scripts Executable

```bash
# Buat semua .sh files executable
chmod +x *.sh
```

### 4. Verify Setup

```bash
./verify-setup.sh
```

---

## 🚀 Menjalankan Services

### Gunakan Script `.sh` (bukan `.bat`)

```bash
cd ~/projects/dfs/server

# Start all services
./start-all.sh

# Check status
./check-status.sh

# Test upload
./test-upload.sh

# Stop all services
./stop-all.sh
```

---

## 📁 File Structure

```
server/
├── *.sh          ← Untuk Linux/WSL (gunakan ini!)
├── *.bat         ← Untuk Windows CMD (jangan gunakan di WSL)
├── naming-service/
├── storage-node/
└── ...
```

---

## 🔄 Perbedaan .bat vs .sh

### Windows (.bat)
```batch
@echo off
echo "Hello"
pause
```
- Dijalankan di: Windows CMD atau PowerShell
- Syntax: Windows batch script
- Tidak bisa di WSL/Linux

### Linux (.sh)
```bash
#!/bin/bash
echo "Hello"
read -p "Press enter..."
```
- Dijalankan di: WSL, Linux, macOS
- Syntax: Bash shell script
- Bisa di WSL ✓

---

## ✅ Quick Start (WSL)

```bash
# 1. Masuk ke WSL
wsl

# 2. Go to project
cd ~/projects/dfs/server

# 3. Make executable
chmod +x *.sh

# 4. Verify
./verify-setup.sh

# 5. Start services
./start-all.sh

# 6. Test (di terminal baru atau detach dari tmux)
./test-upload.sh

# 7. Stop
./stop-all.sh
```

---

## 🎯 Recommended Workflow

### Development di WSL:

1. **Edit code** - Bisa dari Windows (VS Code, Kiro, dll)
2. **Run services** - Di WSL terminal (gunakan `.sh`)
3. **Test** - Di WSL terminal atau Windows browser

### Contoh:

```bash
# Terminal 1 (WSL) - Start services
cd ~/projects/dfs/server
./start-all.sh

# Terminal 2 (WSL) - Testing
cd ~/projects/dfs/server
./test-upload.sh
curl http://localhost:8080/files

# Browser (Windows) - Access API
http://localhost:8080/health
http://localhost:8001/health
```

---

## 💡 Tips

### 1. Alias untuk kemudahan

Tambahkan ke `~/.bashrc`:

```bash
# DFS aliases
alias dfs='cd ~/projects/dfs/server'
alias dfs-start='cd ~/projects/dfs/server && ./start-all.sh'
alias dfs-stop='cd ~/projects/dfs/server && ./stop-all.sh'
alias dfs-status='cd ~/projects/dfs/server && ./check-status.sh'
alias dfs-test='cd ~/projects/dfs/server && ./test-upload.sh'
```

Reload:
```bash
source ~/.bashrc
```

Sekarang bisa:
```bash
dfs-start   # Start all
dfs-status  # Check status
dfs-test    # Test upload
dfs-stop    # Stop all
```

### 2. Auto-start MySQL

```bash
# Add to ~/.bashrc
if ! service mysql status > /dev/null 2>&1; then
    sudo service mysql start
fi
```

### 3. Tmux cheatsheet

```bash
# Attach to session
tmux attach -t mini-dfs

# Switch windows
Ctrl+B then 0/1/2/3

# Detach (keep running)
Ctrl+B then D

# Kill session
tmux kill-session -t mini-dfs
```

---

## 🐛 Common Issues

### "Permission denied" saat run .sh

```bash
chmod +x *.sh
```

### "command not found: ./start-all.sh"

```bash
# Pastikan di directory yang benar
pwd  # Should show: /home/gristof/projects/dfs/server

# Check file exists
ls -la start-all.sh
```

### MySQL tidak running

```bash
sudo service mysql start
sudo service mysql status
```

### Port already in use

```bash
# Check what's using the port
lsof -i :8001

# Kill it
kill -9 <PID>

# Or use stop-all.sh
./stop-all.sh
```

---

## 📚 Documentation Files

**Untuk WSL/Linux:**
- ✅ **README_LINUX.md** - Linux setup guide
- ✅ **SETUP_WSL.md** (this file) - WSL specific guide
- ✅ All `.sh` scripts

**Untuk Windows Native:**
- ⚠️ **start-all.bat** dan `.bat` lainnya
- Hanya jika run di Windows CMD, bukan WSL

**General:**
- **QUICK_START_ID.md** - Quick start (Bahasa Indonesia)
- **TESTING_REPLICATION.md** - Testing guide
- **TROUBLESHOOTING.md** - Common issues
- **IMPLEMENTATION_SUMMARY.md** - Technical overview

---

## ✅ Checklist

Sebelum mulai testing:

- [ ] WSL terminal terbuka
- [ ] Navigate ke `~/projects/dfs/server`
- [ ] Run `chmod +x *.sh`
- [ ] Run `./verify-setup.sh` - semua check passed
- [ ] MySQL running (`sudo service mysql start`)
- [ ] Database schema loaded
- [ ] Python dependencies installed
- [ ] Go dependencies installed

Jika semua ✓, jalankan:

```bash
./start-all.sh
```

---

**Environment:** WSL Ubuntu  
**Use:** `.sh` scripts (NOT `.bat`)  
**Last Updated:** December 3, 2025
