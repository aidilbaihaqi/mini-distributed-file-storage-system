package main

import (
	"database/sql"
	"log"
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
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
}

type NodeStatus struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Status  string `json:"status"` // UP / DOWN
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
        SELECT id, address, status, last_heartbeat
        FROM nodes
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Address, &n.Status, &n.LastHeartbeat); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	return nodes, rows.Err()
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

	log.Println("🚀 Naming service berjalan di :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}

