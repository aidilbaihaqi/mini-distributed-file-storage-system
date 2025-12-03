-- Schema untuk Mini Distributed File Storage System

-- Tabel nodes: menyimpan informasi storage nodes
CREATE TABLE IF NOT EXISTS nodes (
    id VARCHAR(50) PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    status ENUM('UP', 'DOWN') DEFAULT 'DOWN',
    role ENUM('MAIN', 'REPLICA', 'BACKUP') DEFAULT 'REPLICA',
    latency_ms BIGINT DEFAULT 0,
    last_heartbeat TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status_latency (status, latency_ms)
);

-- Tabel files: metadata global file
CREATE TABLE IF NOT EXISTS files (
    file_key VARCHAR(100) PRIMARY KEY,
    original_filename VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel file_locations: lokasi file pada node
CREATE TABLE IF NOT EXISTS file_locations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_key VARCHAR(100) NOT NULL,
    node_id VARCHAR(50) NOT NULL,
    status ENUM('ACTIVE', 'DELETED') DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_key) REFERENCES files(file_key) ON DELETE CASCADE,
    FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE,
    UNIQUE KEY unique_file_node (file_key, node_id)
);

-- Tabel replication_queue: backlog replikasi ketika node DOWN
CREATE TABLE IF NOT EXISTS replication_queue (
    id INT AUTO_INCREMENT PRIMARY KEY,
    file_key VARCHAR(100) NOT NULL,
    target_node_id VARCHAR(50) NOT NULL,
    source_node_id VARCHAR(50) NOT NULL,
    status ENUM('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED') DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    last_attempt TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    error_message TEXT,
    FOREIGN KEY (file_key) REFERENCES files(file_key) ON DELETE CASCADE,
    FOREIGN KEY (target_node_id) REFERENCES nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (source_node_id) REFERENCES nodes(id) ON DELETE CASCADE,
    INDEX idx_status (status),
    INDEX idx_target_node (target_node_id, status)
);

-- Insert default nodes
INSERT INTO nodes (id, address, status, role) VALUES
    ('node-1', 'http://localhost:8001', 'UP', 'MAIN'),
    ('node-2', 'http://localhost:8002', 'UP', 'REPLICA'),
    ('node-3', 'http://localhost:8003', 'UP', 'BACKUP')
ON DUPLICATE KEY UPDATE 
    address = VALUES(address),
    role = VALUES(role);
