#!/bin/bash

echo "========================================"
echo "Testing file upload to Mini DFS"
echo "========================================"
echo ""
echo "⚠️  Upload melalui Naming Service (Port 8080)"
echo ""

# Create a test file if it doesn't exist
if [ ! -f test.txt ]; then
    echo "This is a test file for Mini DFS - $(date)" > test.txt
    echo "✓ Test file created: test.txt"
fi

echo ""
echo "Uploading test.txt via Naming Service..."
echo ""

curl -X POST http://localhost:8080/upload -F "file=@test.txt"

echo ""
echo ""
echo "========================================"
echo "Upload complete!"
echo "========================================"
echo ""
echo "Response menunjukkan:"
echo "- file_id: ID file untuk download/delete"
echo "- selected_node: Node yang dipilih (latency terendah)"
echo "- replication: Status replikasi ke node lain"
echo ""
echo "Untuk download file:"
echo "  curl -O -J http://localhost:8080/download/{FILE_ID}"
echo ""
echo "Untuk list semua files:"
echo "  curl http://localhost:8080/files"
echo ""
echo "Untuk verify di storage nodes:"
echo "  ls storage-node/sn-1/uploads/"
echo "  ls storage-node/sn-2/uploads/"
echo "  ls storage-node/sn-3/uploads/"
echo ""
