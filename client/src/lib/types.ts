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