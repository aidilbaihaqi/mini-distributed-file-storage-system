import type { FileItem, NodeInfo, SystemHealth } from './types'

const base = '/api'

export async function getNodes(): Promise<NodeInfo[]> {
  const res = await fetch(`${base}/nodes`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed nodes ${res.status}`)
  return res.json()
}

export async function checkNodes(): Promise<NodeInfo[]> {
  const res = await fetch(`${base}/nodes/check`, { method: 'POST' })
  if (!res.ok) throw new Error(`Failed check ${res.status}`)
  return res.json()
}

export async function getHealth(): Promise<SystemHealth> {
  const res = await fetch(`${base}/health`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed health ${res.status}`)
  const h = await res.json()
  return {
    status: (h.status ?? 'unknown') as SystemHealth['status'],
    upNodes: h.upNodes ?? 0,
    downNodes: h.downNodes ?? 0,
    totalFiles: h.totalFiles ?? undefined,
  }
}

export async function listFiles(): Promise<FileItem[]> {
  const res = await fetch(`${base}/files`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed files ${res.status}`)
  return res.json()
}

export async function uploadFiles(files: File[], onProgress?: (p: number) => void): Promise<FileItem[]> {
  const fd = new FormData()
  for (const f of files) fd.append('files', f)
  const res = await fetch(`${base}/upload`, {
    method: 'POST',
    body: fd,
  })
  if (!res.ok) throw new Error(`Failed upload ${res.status}`)
  return res.json()
}

export async function deleteFile(id: string): Promise<void> {
  const res = await fetch(`${base}/files/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`Failed delete ${res.status}`)
}

export async function getLogs(): Promise<string[]> {
  const res = await fetch(`${base}/logs`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`Failed logs ${res.status}`)
  return res.json()
}