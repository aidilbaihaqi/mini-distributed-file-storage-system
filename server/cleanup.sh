#!/bin/bash

echo "========================================"
echo "Mini DFS - Cleanup Script"
echo "========================================"
echo ""
echo "WARNING: This will delete all uploaded files and reset database!"
echo ""
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Cleanup cancelled."
    exit 0
fi

echo ""
echo "Cleaning up..."
echo ""

# Delete uploaded files
echo "[1/4] Deleting uploaded files from storage nodes..."
rm -f storage-node/sn-1/uploads/* 2>/dev/null
rm -f storage-node/sn-2/uploads/* 2>/dev/null
rm -f storage-node/sn-3/uploads/* 2>/dev/null
echo "   - Storage Node 1: Cleaned"
echo "   - Storage Node 2: Cleaned"
echo "   - Storage Node 3: Cleaned"

# Delete test files
echo ""
echo "[2/4] Deleting test files..."
rm -f test*.txt 2>/dev/null
rm -f response.json 2>/dev/null
echo "   - Test files deleted"

# Reset database
echo ""
echo "[3/4] Resetting database..."
mysql -u dfs_user -padmin123 dfs_meta -e "DROP TABLE IF EXISTS replication_queue, file_locations, files, nodes;" 2>/dev/null
mysql -u dfs_user -padmin123 dfs_meta < naming-service/schema.sql 2>/dev/null
echo "   - Database reset complete"

# Verify
echo ""
echo "[4/4] Verifying cleanup..."
echo ""
echo "Storage Node 1 uploads:"
ls -la storage-node/sn-1/uploads/ 2>/dev/null | wc -l
echo ""
echo "Storage Node 2 uploads:"
ls -la storage-node/sn-2/uploads/ 2>/dev/null | wc -l
echo ""
echo "Storage Node 3 uploads:"
ls -la storage-node/sn-3/uploads/ 2>/dev/null | wc -l
echo ""

echo "========================================"
echo "Cleanup Complete!"
echo "========================================"
echo ""
echo "All uploaded files deleted"
echo "Database reset to initial state"
echo "Test files removed"
echo ""
echo "You can now start fresh testing."
echo ""
