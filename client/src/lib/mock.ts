import type { FileItem, NodeInfo, SystemHealth } from './types'

export const mockNodes: NodeInfo[] = [
  {
    id: 'node-1',
    address: 'http://localhost:8001',
    status: 'UP',
    role: 'MAIN',
    lastHeartbeat: new Date().toISOString(),
  },
  {
    id: 'node-2',
    address: 'http://localhost:8002',
    status: 'UP',
    role: 'REPLICA',
    lastHeartbeat: new Date().toISOString(),
  },
  {
    id: 'node-3',
    address: 'http://localhost:8003',
    status: 'DOWN',
    role: 'BACKUP',
    lastHeartbeat: new Date().toISOString(),
  },
]

export const mockHealth: SystemHealth & { uptime: string; version: string; requests: number; baseUrl: string } = {
  status: 'ok',
  upNodes: 2,
  downNodes: 1,
  totalFiles: 7,
  uptime: '03:25:45',
  version: 'v0.1.0',
  requests: 42,
  baseUrl: 'http://localhost:8080',
}

export const mockFiles: FileItem[] = [
  { id: 'f1', filename: 'design.png', size: 234567, createdAt: new Date().toISOString(), replicas: ['node-1','node-2'] },
  { id: 'f2', filename: 'report.pdf', size: 834234, createdAt: new Date().toISOString(), replicas: ['node-2'] },
  { id: 'f3', filename: 'music.mp3', size: 5234623, createdAt: new Date().toISOString(), replicas: ['node-1'] },
]

export const mockLogs: string[] = [
  '[INFO] Naming service started at http://localhost:8080',
  '[INFO] Registered node node-1 (MAIN) http://localhost:8001',
  '[INFO] Registered node node-2 (REPLICA) http://localhost:8002',
  '[WARN] node-3 heartbeat missed (status DOWN)',
  '[INFO] File upload design.png replicas: node-1,node-2',
  '[INFO] Requests processed: 42',
]