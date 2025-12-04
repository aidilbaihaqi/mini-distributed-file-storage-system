package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
    "mime/multipart"
    "mime"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Node struct {
	ID            string     `json:"id"`
	Address       string     `json:"address"`
	Status        string     `json:"status"`
	Role          string     `json:"role"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	LatencyMs     int64      `json:"latency_ms,omitempty"`
}

type NodeStatus struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Status  string `json:"status"` // UP / DOWN
}

type FileMetadata struct {
	FileKey          string   `json:"file_key"`
	OriginalFilename string   `json:"original_filename"`
	SizeBytes        int64    `json:"size_bytes"`
	ChecksumSHA256   string   `json:"checksum_sha256"`
	UploadedAt       string   `json:"uploaded_at"`
	Replicas         []string `json:"replicas"`
}

type ReplicationQueueItem struct {
	ID            int       `json:"id"`
	FileKey       string    `json:"file_key"`
	TargetNodeID  string    `json:"target_node_id"`
	SourceNodeID  string    `json:"source_node_id"`
	Status        string    `json:"status"`
	RetryCount    int       `json:"retry_count"`
	LastAttempt   *time.Time `json:"last_attempt,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

var db *sql.DB

func initDB() {
	// sesuaikan username/password/database dengan yang tadi dibuat
	dsn := "dfs_user:admin123@tcp(127.0.0.1:3306)/dfs_meta?parseTime=true"

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("gagal buka koneksi ke MySQL: %v", err)
	}

	// cek koneksi
	if err := db.Ping(); err != nil {
		log.Fatalf("gagal ping MySQL: %v", err)
	}

	log.Println("✅ Terhubung ke MySQL dfs_meta")
}

func getAllNodes() ([]Node, error) {
	rows, err := db.Query(`
        SELECT id, address, status, role, last_heartbeat, COALESCE(latency_ms, 0) as latency_ms
        FROM nodes
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Address, &n.Status, &n.Role, &n.LastHeartbeat, &n.LatencyMs); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	return nodes, rows.Err()
}

func updateNodeStatus(nodeID string, status string) error {
	_, err := db.Exec(`
		UPDATE nodes 
		SET status = ?, last_heartbeat = NOW() 
		WHERE id = ?
	`, status, nodeID)
	return err
}

func updateNodeLatency(nodeID string, latencyMs int64) error {
	_, err := db.Exec(`
		UPDATE nodes 
		SET latency_ms = ?
		WHERE id = ?
	`, latencyMs, nodeID)
	return err
}

func measureNodeLatency(nodeAddr string) int64 {
	client := &http.Client{Timeout: 2 * time.Second}
	
	start := time.Now()
	resp, err := client.Get(nodeAddr + "/health")
	elapsed := time.Since(start).Milliseconds()
	
	if err != nil {
		return 9999 // Return high latency if node is unreachable
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 9999
	}
	
	return elapsed
}

func selectBestNodeForUpload(nodes []Node) *Node {
	// Upload harus selalu ke MAIN node karena hanya MAIN yang handle replication
	// Jika MAIN down, pilih node lain dengan latency terendah sebagai fallback
	
	var mainNode *Node
	var fallbackNode *Node
	lowestLatency := int64(9999)
	
	for i := range nodes {
		node := &nodes[i]
		if node.Status == "UP" {
			// Prioritas 1: MAIN node
			if node.Role == "MAIN" {
				mainNode = node
			}
			// Fallback: node dengan latency terendah
			if node.LatencyMs < lowestLatency {
				lowestLatency = node.LatencyMs
				fallbackNode = node
			}
		}
	}
	
	// Return MAIN node jika UP, otherwise fallback
	if mainNode != nil {
		return mainNode
	}
	return fallbackNode
}

func selectBestNodeForDownload(fileKey string, nodes []Node) *Node {
	// Get nodes that have the file
	nodeIDs, err := getFileLocations(fileKey)
	if err != nil || len(nodeIDs) == 0 {
		return nil
	}
	
	// Create map for quick lookup
	hasFileMap := make(map[string]bool)
	for _, nodeID := range nodeIDs {
		hasFileMap[nodeID] = true
	}
	
	// Find best node among those that have the file
	var bestNode *Node
	lowestLatency := int64(9999)
	
	for i := range nodes {
		node := &nodes[i]
		if node.Status == "UP" && hasFileMap[node.ID] {
			if node.LatencyMs < lowestLatency {
				lowestLatency = node.LatencyMs
				bestNode = node
			}
		}
	}
	
	return bestNode
}

func addToReplicationQueue(fileKey, targetNodeID, sourceNodeID string) error {
	_, err := db.Exec(`
		INSERT INTO replication_queue (file_key, target_node_id, source_node_id, status)
		VALUES (?, ?, ?, 'PENDING')
	`, fileKey, targetNodeID, sourceNodeID)
	return err
}

func getPendingReplications(targetNodeID string) ([]ReplicationQueueItem, error) {
	rows, err := db.Query(`
		SELECT id, file_key, target_node_id, source_node_id, status, retry_count, last_attempt, created_at
		FROM replication_queue
		WHERE target_node_id = ? AND status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 100
	`, targetNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ReplicationQueueItem
	for rows.Next() {
		var item ReplicationQueueItem
		if err := rows.Scan(&item.ID, &item.FileKey, &item.TargetNodeID, &item.SourceNodeID, 
			&item.Status, &item.RetryCount, &item.LastAttempt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func markReplicationCompleted(queueID int) error {
	_, err := db.Exec(`
		UPDATE replication_queue 
		SET status = 'COMPLETED', completed_at = NOW()
		WHERE id = ?
	`, queueID)
	return err
}

func markReplicationFailed(queueID int, errorMsg string) error {
	_, err := db.Exec(`
		UPDATE replication_queue 
		SET status = 'FAILED', retry_count = retry_count + 1, last_attempt = NOW(), error_message = ?
		WHERE id = ?
	`, errorMsg, queueID)
	return err
}

func getFileLocations(fileKey string) ([]string, error) {
	rows, err := db.Query(`
		SELECT node_id FROM file_locations 
		WHERE file_key = ? AND status = 'ACTIVE'
	`, fileKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodeIDs []string
	for rows.Next() {
		var nodeID string
		if err := rows.Scan(&nodeID); err != nil {
			return nil, err
		}
		nodeIDs = append(nodeIDs, nodeID)
	}

	return nodeIDs, rows.Err()
}

func replicateFileToNode(fileKey, sourceNodeAddr, targetNodeAddr string) error {
	client := &http.Client{Timeout: 30 * time.Second}

	// Download dari source node
	resp, err := client.Get(fmt.Sprintf("%s/files/%s", sourceNodeAddr, fileKey))
	if err != nil {
		return fmt.Errorf("gagal download dari source: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("source node return status %d", resp.StatusCode)
	}

	// Baca file content
	fileContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gagal baca file: %v", err)
	}

	// Upload ke target node
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
    // Ambil filename dari header jika ada
    filename := fileKey
    if cd := resp.Header.Get("Content-Disposition"); cd != "" {
        if _, params, err := mime.ParseMediaType(cd); err == nil {
            if fn, ok := params["filename"]; ok && fn != "" {
                filename = fn
            }
        }
    }

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("gagal create form file: %v", err)
	}

	if _, err := part.Write(fileContent); err != nil {
		return fmt.Errorf("gagal write file content: %v", err)
	}

	writer.Close()

    req, err := http.NewRequest("POST", fmt.Sprintf("%s/files?file_id=%s", targetNodeAddr, fileKey), body)
	if err != nil {
		return fmt.Errorf("gagal create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal upload ke target: %v", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != 200 {
		return fmt.Errorf("target node return status %d", uploadResp.StatusCode)
	}

	return nil
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	// Health naming service sendiri
	r.GET("/health", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.JSON(http.StatusOK, gin.H{
			"status":   "UP",
			"service":  "naming-service",
			"hostname": hostname,
		})
	})

	// List node dari database
	r.GET("/nodes", func(c *gin.Context) {
		nodes, err := getAllNodes()
		if err != nil {
			log.Println("error ambil nodes:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data nodes"})
			return
		}
		c.JSON(http.StatusOK, nodes)
	})

	// Health check semua node (ping /health ke tiap storage node)
	r.GET("/nodes/check", func(c *gin.Context) {
		client := &http.Client{
			Timeout: 2 * time.Second,
		}

		nodes, err := getAllNodes()
		if err != nil {
			log.Println("error ambil nodes:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data nodes"})
			return
		}

		statuses := make([]NodeStatus, 0, len(nodes))

		for _, n := range nodes {
			resp, err := client.Get(n.Address + "/health")

			nodeStatus := NodeStatus{
				ID:      n.ID,
				Address: n.Address,
				Status:  "DOWN",
			}

			if err == nil && resp.StatusCode == http.StatusOK {
				nodeStatus.Status = "UP"
			}

			statuses = append(statuses, nodeStatus)
		}

		c.JSON(http.StatusOK, gin.H{
			"checked_at": time.Now().Format(time.RFC3339),
			"nodes":      statuses,
		})
	})

	// Endpoint untuk register file metadata setelah upload
	r.POST("/files/register", func(c *gin.Context) {
		var req struct {
			FileKey          string   `json:"file_key"`
			OriginalFilename string   `json:"original_filename"`
			SizeBytes        int64    `json:"size_bytes"`
			ChecksumSHA256   string   `json:"checksum_sha256"`
			NodeID           string   `json:"node_id"`
			FailedNodes      []string `json:"failed_nodes"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Simpan metadata file
		_, err := db.Exec(`
			INSERT INTO files (file_key, original_filename, size_bytes, checksum_sha256)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE 
				original_filename = VALUES(original_filename),
				size_bytes = VALUES(size_bytes),
				checksum_sha256 = VALUES(checksum_sha256)
		`, req.FileKey, req.OriginalFilename, req.SizeBytes, req.ChecksumSHA256)

		if err != nil {
			log.Println("error insert file metadata:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal simpan metadata"})
			return
		}

		// Simpan lokasi file
		_, err = db.Exec(`
			INSERT INTO file_locations (file_key, node_id, status)
			VALUES (?, ?, 'ACTIVE')
			ON DUPLICATE KEY UPDATE status = 'ACTIVE'
		`, req.FileKey, req.NodeID)

		if err != nil {
			log.Println("error insert file location:", err)
		}

		// Tambahkan ke replication queue untuk node yang gagal
		for _, failedNodeID := range req.FailedNodes {
			if err := addToReplicationQueue(req.FileKey, failedNodeID, req.NodeID); err != nil {
				log.Printf("error add to replication queue for node %s: %v\n", failedNodeID, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "file metadata registered",
		})
	})

	// Endpoint untuk recovery - sync file yang pending ke node yang baru UP
	r.POST("/nodes/:nodeId/recover", func(c *gin.Context) {
		nodeID := c.Param("nodeId")

		// Ambil pending replications untuk node ini
		items, err := getPendingReplications(nodeID)
		if err != nil {
			log.Println("error get pending replications:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil pending replications"})
			return
		}

		if len(items) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "no pending replications",
				"count":   0,
			})
			return
		}

		// Ambil info node target dan source
		nodes, err := getAllNodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil nodes"})
			return
		}

		nodeMap := make(map[string]string)
		for _, n := range nodes {
			nodeMap[n.ID] = n.Address
		}

		// Proses setiap item
		successCount := 0
		failCount := 0

		for _, item := range items {
			sourceAddr, sourceOk := nodeMap[item.SourceNodeID]
			targetAddr, targetOk := nodeMap[item.TargetNodeID]

			if !sourceOk || !targetOk {
				markReplicationFailed(item.ID, "node not found")
				failCount++
				continue
			}

			// Lakukan replikasi
			if err := replicateFileToNode(item.FileKey, sourceAddr, targetAddr); err != nil {
				log.Printf("replication failed for queue %d: %v\n", item.ID, err)
				markReplicationFailed(item.ID, err.Error())
				failCount++
			} else {
				markReplicationCompleted(item.ID)
				
				// Update file_locations
				db.Exec(`
					INSERT INTO file_locations (file_key, node_id, status)
					VALUES (?, ?, 'ACTIVE')
					ON DUPLICATE KEY UPDATE status = 'ACTIVE'
				`, item.FileKey, item.TargetNodeID)
				
				successCount++
				log.Printf("✅ Replicated %s to %s\n", item.FileKey, item.TargetNodeID)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "recovery completed",
			"total":         len(items),
			"success":       successCount,
			"failed":        failCount,
			"pending_items": items,
		})
	})

	// Endpoint untuk melihat replication queue
	r.GET("/replication-queue", func(c *gin.Context) {
		nodeID := c.Query("node_id")
		status := c.Query("status")

		query := "SELECT id, file_key, target_node_id, source_node_id, status, retry_count, last_attempt, created_at FROM replication_queue WHERE 1=1"
		args := []interface{}{}

		if nodeID != "" {
			query += " AND target_node_id = ?"
			args = append(args, nodeID)
		}

		if status != "" {
			query += " AND status = ?"
			args = append(args, status)
		}

		query += " ORDER BY created_at DESC LIMIT 100"

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal query queue"})
			return
		}
		defer rows.Close()

		var items []ReplicationQueueItem
		for rows.Next() {
			var item ReplicationQueueItem
			if err := rows.Scan(&item.ID, &item.FileKey, &item.TargetNodeID, &item.SourceNodeID,
				&item.Status, &item.RetryCount, &item.LastAttempt, &item.CreatedAt); err != nil {
				continue
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"items": items,
			"count": len(items),
		})
	})

	// Endpoint untuk upload file via naming service
	r.POST("/upload", func(c *gin.Context) {
		// Get file from request
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
			return
		}

		// Get all nodes
		nodes, err := getAllNodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil nodes"})
			return
		}

		// Select best node based on latency
		bestNode := selectBestNodeForUpload(nodes)
		if bestNode == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available nodes"})
			return
		}

		log.Printf("📤 Routing upload to %s (latency: %dms)\n", bestNode.ID, bestNode.LatencyMs)

		// Forward file to selected node
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal baca file"})
			return
		}
		defer fileContent.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal create form"})
			return
		}

		if _, err := io.Copy(part, fileContent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal copy file"})
			return
		}
		writer.Close()

		// Send to storage node
		client := &http.Client{Timeout: 60 * time.Second}
		req, err := http.NewRequest("POST", bestNode.Address+"/files", body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal create request"})
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gagal upload ke node: %v", err)})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("node return status %d", resp.StatusCode)})
			return
		}

		// Parse response from storage node
		var uploadResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal parse response"})
			return
		}

		// Add routing info to response
		uploadResp["routed_via"] = "naming-service"
		uploadResp["selected_node"] = bestNode.ID
		uploadResp["node_latency_ms"] = bestNode.LatencyMs

		c.JSON(http.StatusOK, uploadResp)
	})

	// Endpoint untuk download file via naming service
	r.GET("/download/:fileKey", func(c *gin.Context) {
		fileKey := c.Param("fileKey")

		// Get all nodes
		nodes, err := getAllNodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil nodes"})
			return
		}

		// Select best node that has the file
		bestNode := selectBestNodeForDownload(fileKey, nodes)
		if bestNode == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found or no available nodes"})
			return
		}

		log.Printf("📥 Routing download from %s (latency: %dms)\n", bestNode.ID, bestNode.LatencyMs)

		// Forward request to selected node
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Get(fmt.Sprintf("%s/files/%s", bestNode.Address, fileKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gagal download dari node: %v", err)})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		// Copy headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Add custom header to indicate routing
		c.Header("X-Routed-From", bestNode.ID)
		c.Header("X-Node-Latency-Ms", fmt.Sprintf("%d", bestNode.LatencyMs))

		// Stream file to client
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	})

	// Endpoint untuk delete file via naming service
	r.DELETE("/files/:fileKey", func(c *gin.Context) {
		fileKey := c.Param("fileKey")

		log.Printf("🗑️ Delete request for file: %s\n", fileKey)

		// Get all nodes that have the file
		nodeIDs, err := getFileLocations(fileKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil file locations"})
			return
		}

		// Jika tidak ada di file_locations, coba cari di semua nodes
		if len(nodeIDs) == 0 {
			// Coba delete dari semua nodes (fallback)
			nodeIDs = []string{"node-1", "node-2", "node-3"}
			log.Printf("⚠️ File %s not found in file_locations, trying all nodes\n", fileKey)
		}

		// Get all nodes info
		nodes, err := getAllNodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal ambil nodes"})
			return
		}

		nodeMap := make(map[string]string)
		for _, n := range nodes {
			nodeMap[n.ID] = n.Address
		}

		// Delete from all nodes
		client := &http.Client{Timeout: 10 * time.Second}
		successCount := 0
		failCount := 0
		deletedNodes := []string{}

		for _, nodeID := range nodeIDs {
			nodeAddr, ok := nodeMap[nodeID]
			if !ok {
				log.Printf("⚠️ Node %s not found in nodeMap\n", nodeID)
				failCount++
				continue
			}

			req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files/%s", nodeAddr, fileKey), nil)
			if err != nil {
				log.Printf("❌ Failed to create delete request for %s: %v\n", nodeID, err)
				failCount++
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("❌ Failed to delete from %s: %v\n", nodeID, err)
				failCount++
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == 200 {
				successCount++
				deletedNodes = append(deletedNodes, nodeID)
				log.Printf("✅ Deleted from %s\n", nodeID)
				
				// Update file_locations
				db.Exec(`
					UPDATE file_locations 
					SET status = 'DELETED' 
					WHERE file_key = ? AND node_id = ?
				`, fileKey, nodeID)
			} else if resp.StatusCode == 404 {
				// File not found on this node, not an error
				log.Printf("ℹ️ File not found on %s (already deleted or never existed)\n", nodeID)
			} else {
				log.Printf("❌ Delete from %s returned status %d\n", nodeID, resp.StatusCode)
				failCount++
			}
		}

		// Delete from replication_queue
		result, err := db.Exec(`DELETE FROM replication_queue WHERE file_key = ?`, fileKey)
		if err != nil {
			log.Printf("⚠️ Failed to delete from replication_queue: %v\n", err)
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				log.Printf("✅ Deleted %d entries from replication_queue\n", rowsAffected)
			}
		}

		// Delete from file_locations
		db.Exec(`DELETE FROM file_locations WHERE file_key = ?`, fileKey)

		// Delete from files table
		db.Exec(`DELETE FROM files WHERE file_key = ?`, fileKey)

		log.Printf("🗑️ Delete completed for %s: %d success, %d failed\n", fileKey, successCount, failCount)

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"file_key":      fileKey,
			"deleted_from":  successCount,
			"failed":        failCount,
			"total_nodes":   len(nodeIDs),
			"deleted_nodes": deletedNodes,
			"message":       "File deleted from all nodes and database",
		})
	})

	// Endpoint untuk list files
	r.GET("/files", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT f.file_key, f.original_filename, f.size_bytes, f.checksum_sha256, f.uploaded_at
			FROM files f
			ORDER BY f.uploaded_at DESC
			LIMIT 100
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal query files"})
			return
		}
		defer rows.Close()

		var files []FileMetadata
		for rows.Next() {
			var f FileMetadata
			if err := rows.Scan(&f.FileKey, &f.OriginalFilename, &f.SizeBytes, &f.ChecksumSHA256, &f.UploadedAt); err != nil {
				continue
			}

			// Ambil replicas
			replicas, _ := getFileLocations(f.FileKey)
			f.Replicas = replicas

			files = append(files, f)
		}

		c.JSON(http.StatusOK, gin.H{
			"files": files,
			"count": len(files),
		})
	})

	// Background job untuk auto-recovery dan latency measurement (cek setiap 30 detik)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			nodes, err := getAllNodes()
			if err != nil {
				continue
			}

			client := &http.Client{Timeout: 2 * time.Second}

			for _, node := range nodes {
				// Measure latency
				latency := measureNodeLatency(node.Address)
				updateNodeLatency(node.ID, latency)

				// Cek health
				resp, err := client.Get(node.Address + "/health")
				newStatus := "DOWN"
				if err == nil && resp.StatusCode == http.StatusOK {
					newStatus = "UP"
					resp.Body.Close()
				}

				// Update status jika berubah
				if node.Status != newStatus {
					log.Printf("Node %s status changed: %s -> %s (latency: %dms)\n", node.ID, node.Status, newStatus, latency)
					updateNodeStatus(node.ID, newStatus)

					// Jika node baru UP, trigger recovery
					if newStatus == "UP" {
						log.Printf("🔄 Triggering recovery for node %s\n", node.ID)
						
						// Panggil recovery endpoint secara internal
						items, err := getPendingReplications(node.ID)
						if err != nil || len(items) == 0 {
							continue
						}

						nodeMap := make(map[string]string)
						for _, n := range nodes {
							nodeMap[n.ID] = n.Address
						}

						for _, item := range items {
							sourceAddr, sourceOk := nodeMap[item.SourceNodeID]
							targetAddr := node.Address

							if !sourceOk {
								continue
							}

							if err := replicateFileToNode(item.FileKey, sourceAddr, targetAddr); err != nil {
								markReplicationFailed(item.ID, err.Error())
							} else {
								markReplicationCompleted(item.ID)
								db.Exec(`
									INSERT INTO file_locations (file_key, node_id, status)
									VALUES (?, ?, 'ACTIVE')
									ON DUPLICATE KEY UPDATE status = 'ACTIVE'
								`, item.FileKey, item.TargetNodeID)
								log.Printf("✅ Auto-recovered %s to %s\n", item.FileKey, node.ID)
							}
						}
					}
				}
			}
		}
	}()

	log.Println("🚀 Naming service berjalan di :8080")
	log.Println("📊 Auto-recovery background job started")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}

