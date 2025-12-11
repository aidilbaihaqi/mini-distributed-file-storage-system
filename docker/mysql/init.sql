-- Initialize DFS Database

USE dfs_meta;

-- Tabel nodes
CREATE TABLE IF NOT EXISTS nodes (
    id VARCHAR(50) PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'DOWN',
    role VARCHAR(20) DEFAULT 'REPLICA',
    latency_ms INT DEFAULT 0,
    last_heartbeat DATETIME
);

-- Tabel files
CREATE TABLE IF NOT EXISTS files (
    file_key VARCHAR(255) PRIMARY KEY,
    original_filename VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64),
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tabel file_locations
CREATE TABLE IF NOT EXISTS file_locations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_key VARCHAR(255) NOT NULL,
    node_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_file_node (file_key, node_id)
);

-- Tabel replication_queue
CREATE TABLE IF NOT EXISTS replication_queue (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_key VARCHAR(255) NOT NULL,
    target_node_id VARCHAR(50) NOT NULL,
    source_node_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    last_attempt DATETIME,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Insert default nodes (menggunakan nama container Docker)
INSERT INTO nodes (id, address, status, role) VALUES
('node-1', 'http://storage-node-1:8000', 'DOWN', 'MAIN'),
('node-2', 'http://storage-node-2:8000', 'DOWN', 'BACKUP'),
('node-3', 'http://storage-node-3:8000', 'DOWN', 'REPLICA');
