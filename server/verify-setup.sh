#!/bin/bash

echo "========================================"
echo "Mini DFS - Setup Verification"
echo "========================================"
echo ""

ERRORS=0

# Check MySQL
echo "[1/8] Checking MySQL..."
if command -v mysql &> /dev/null; then
    echo "   [OK] MySQL installed"
else
    echo "   [FAIL] MySQL not found"
    ((ERRORS++))
fi

# Check Python
echo ""
echo "[2/8] Checking Python..."
if command -v python3 &> /dev/null; then
    echo "   [OK] Python installed"
else
    echo "   [FAIL] Python not found"
    ((ERRORS++))
fi

# Check Go
echo ""
echo "[3/8] Checking Go..."
if command -v go &> /dev/null; then
    echo "   [OK] Go installed"
else
    echo "   [FAIL] Go not found"
    ((ERRORS++))
fi

# Check curl
echo ""
echo "[4/8] Checking curl..."
if command -v curl &> /dev/null; then
    echo "   [OK] curl installed"
else
    echo "   [FAIL] curl not found"
    ((ERRORS++))
fi

# Check database
echo ""
echo "[5/8] Checking database connection..."
if mysql -u dfs_user -padmin123 dfs_meta -e "SELECT 1;" &> /dev/null; then
    echo "   [OK] Database connection successful"
else
    echo "   [FAIL] Cannot connect to database dfs_meta"
    echo "          Run: mysql -u root -p"
    echo "          Then: CREATE DATABASE dfs_meta;"
    echo "                CREATE USER 'dfs_user'@'localhost' IDENTIFIED BY 'admin123';"
    echo "                GRANT ALL PRIVILEGES ON dfs_meta.* TO 'dfs_user'@'localhost';"
    ((ERRORS++))
fi

# Check database tables
echo ""
echo "[6/8] Checking database tables..."
if mysql -u dfs_user -padmin123 dfs_meta -e "SHOW TABLES LIKE 'nodes';" 2>/dev/null | grep -q "nodes"; then
    echo "   [OK] Database tables exist"
else
    echo "   [FAIL] Database tables not found"
    echo "          Run: mysql -u dfs_user -padmin123 dfs_meta < naming-service/schema.sql"
    ((ERRORS++))
fi

# Check Python dependencies
echo ""
echo "[7/8] Checking Python dependencies..."
if python3 -c "import fastapi, httpx" &> /dev/null; then
    echo "   [OK] Python dependencies installed"
else
    echo "   [FAIL] Python dependencies missing"
    echo "          Run: cd storage-node/sn-1 && pip3 install -r requirements.txt"
    ((ERRORS++))
fi

# Check Go dependencies
echo ""
echo "[8/8] Checking Go dependencies..."
if [ -f "naming-service/go.mod" ]; then
    echo "   [OK] Go module initialized"
else
    echo "   [FAIL] Go module not initialized"
    echo "          Run: cd naming-service && go mod tidy"
    ((ERRORS++))
fi

echo ""
echo "========================================"
echo "Verification Summary"
echo "========================================"
echo ""

if [ $ERRORS -eq 0 ]; then
    echo "[SUCCESS] All checks passed! ✓"
    echo ""
    echo "You can now start the services:"
    echo "   cd server"
    echo "   ./start-all.sh"
else
    echo "[FAILED] $ERRORS check(s) failed! ✗"
    echo ""
    echo "Please fix the issues above before starting services."
    echo "See INSTALLATION.md for detailed setup instructions."
fi

echo ""
