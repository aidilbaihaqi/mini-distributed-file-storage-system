import type { FileItem, NodeInfo, SystemHealth, ReplicationQueueItem } from './types'

const base = '/api'

export async function getNodes(): Promise<NodeInfo[]> {
  const res = await fetch(`${base}/nodes`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed nodes ${res.status}`)
  const nodes = await res.json()
  return (nodes as any[]).map(n => ({
    id: n.id,
    address: n.address,
    status: (n.status ?? 'DOWN') as NodeInfo['status'],
    role: n.role,
    lastHeartbeat: n.last_heartbeat ? String(n.last_heartbeat) : new Date().toISOString(),
  }))
}

export async function checkNodes(): Promise<NodeInfo[]> {
  const res = await fetch(`${base}/nodes/check`, { method: 'GET', cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed check ${res.status}`)
  const data = await res.json()
  const arr = Array.isArray(data.nodes) ? data.nodes : []
  return arr.map((n: any) => ({
    id: n.id,
    address: n.address,
    status: (n.status ?? 'DOWN') as NodeInfo['status'],
    lastHeartbeat: new Date().toISOString(),
  }))
}

export async function getHealth(): Promise<SystemHealth> {
  const res = await fetch(`${base}/health`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed health ${res.status}`)
  const h = await res.json()
  const statusRaw = String(h.status ?? 'UNKNOWN')
  const status = statusRaw === 'UP' ? 'ok' : statusRaw === 'DOWN' ? 'down' : 'unknown'
  return {
    status,
    upNodes: 0,
    downNodes: 0,
    totalFiles: undefined,
  }
}

export async function listFiles(): Promise<FileItem[]> {
  const res = await fetch(`${base}/files`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed files ${res.status}`)
  const data = await res.json()
  const files = Array.isArray(data.files) ? data.files : data
  return (files as any[]).map(f => ({
    id: f.file_key ?? f.id,
    filename: f.original_filename ?? f.filename,
    size: f.size_bytes ?? f.size ?? 0,
    replicas: f.replicas ?? [],
    createdAt: f.uploaded_at ?? new Date().toISOString(),
    type: undefined,
  }))
}

export async function uploadFiles(files: File[], onProgress?: (p: number) => void): Promise<FileItem[]> {
  const out: FileItem[] = []
  for (let i = 0; i < files.length; i++) {
    const fd = new FormData()
    fd.append('file', files[i])
    const res = await fetch(`${base}/upload`, { method: 'POST', body: fd })
    if (!res.ok) throw new Error(`Failed upload ${res.status}`)
    const r = await res.json()
    out.push({
      id: r.file_id ?? r.fileKey ?? `unknown-${Date.now()}`,
      filename: r.original_filename ?? files[i].name,
      size: r.size_bytes ?? files[i].size ?? 0,
      createdAt: new Date().toISOString(),
      replicas: [],
    })
    if (onProgress) onProgress(Math.round(((i + 1) / files.length) * 100))
  }
  return out
}

export async function deleteFile(id: string): Promise<void> {
  const res = await fetch(`${base}/files/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`Failed delete ${res.status}`)
}

export async function apiBase(): Promise<string> {
  const env = (process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080') as string
  return env
}

export async function getReplicationQueue(params?: { nodeId?: string; status?: string }): Promise<ReplicationQueueItem[]> {
  const qs = new URLSearchParams()
  if (params?.nodeId) qs.set('node_id', params.nodeId)
  if (params?.status) qs.set('status', params.status)
  const url = `${base}/replication-queue${qs.toString() ? `?${qs.toString()}` : ''}`
  const res = await fetch(url, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed queue ${res.status}`)
  const data = await res.json()
  const items = Array.isArray(data.items) ? data.items : []
  return items.map((it: any) => ({
    id: Number(it.id),
    file_key: String(it.file_key),
    target_node_id: String(it.target_node_id),
    source_node_id: String(it.source_node_id),
    status: String(it.status),
    retry_count: Number(it.retry_count ?? 0),
    last_attempt: it.last_attempt ? String(it.last_attempt) : undefined,
    created_at: String(it.created_at),
  }))
}
