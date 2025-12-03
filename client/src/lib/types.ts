export type NodeRole = 'MAIN' | 'BACKUP' | 'REPLICA'
export type NodeStatus = 'UP' | 'DOWN'

export interface NodeInfo {
  id: string
  address: string
  status: NodeStatus
  role?: NodeRole
  lastHeartbeat: string
}

export interface FileItem {
  id: string
  filename: string
  size: number
  replicas?: string[]
  createdAt: string
  type?: string
}

export interface SystemHealth {
  status: 'ok' | 'degraded' | 'down' | 'unknown'
  upNodes: number
  downNodes: number
  totalFiles?: number
}

export interface ReplicationQueueItem {
  id: number
  file_key: string
  target_node_id: string
  source_node_id: string
  status: string
  retry_count: number
  last_attempt?: string
  created_at: string
}
