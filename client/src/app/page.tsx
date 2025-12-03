"use client"
import { useEffect, useMemo, useState } from "react"
import type { NodeInfo, FileItem, ReplicationQueueItem } from "@/lib/types"
import NodeTable from "@/components/NodeTable"
import UploadDropzone from "@/components/UploadDropzone"
import FileExplorer from "@/components/FileExplorer"
import Panel from "@/components/Panel"
import StatCard from "@/components/StatCard"
import { getNodes, checkNodes, getHealth, listFiles, uploadFiles, deleteFile as apiDelete, getReplicationQueue } from "@/lib/api"

export default function Home() {
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string|null>(null)
  const [health, setHealth] = useState("unknown")
  const [refreshToken, setRefreshToken] = useState(0)
  const [files, setFiles] = useState<FileItem[]>([])
  const [queue, setQueue] = useState<ReplicationQueueItem[]>([])

  useEffect(()=>{
    ;(async()=>{
      try {
        const [n, h, f, q] = await Promise.all([getNodes(), getHealth(), listFiles(), getReplicationQueue({ status: 'PENDING' })])
        setNodes(n)
        setHealth(h.status)
        setFiles(f)
        setQueue(q)
      } catch(e:any){ setError(e.message) }
    })()
  },[])

  const upDown = useMemo(()=>({
    up: nodes.filter(n=>n.status==='UP').length,
    down: nodes.filter(n=>n.status!=='UP').length,
  }),[nodes])

  async function refresh(){
    setLoading(true); setError(null)
    try {
      const status = await checkNodes()
      setNodes(status)
      const f = await listFiles()
      setFiles(f)
      const q = await getReplicationQueue({ status: 'PENDING' })
      setQueue(q)
    } catch(e:any){ setError(e.message) } finally { setLoading(false) }
  }

  async function onFilesAdded(add: File[]){
    setLoading(true); setError(null)
    try {
      const uploaded = await uploadFiles(add)
      setFiles(prev=>[...uploaded, ...prev])
      setRefreshToken(t=>t+1)
    } catch(e:any){ setError(e.message) } finally { setLoading(false) }
  }

  async function onDeleteFile(id: string){
    setLoading(true); setError(null)
    try { await apiDelete(id); setFiles(prev=>prev.filter(f=>f.id!==id)) } catch(e:any){ setError(e.message) } finally { setLoading(false) }
  }

  return (
    <div className="min-h-screen bg-neutral-950 text-white">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">DFS Dashboard</h1>
          <div className="flex gap-2">
            <button className="px-3 py-1.5 rounded bg-blue-600">Check Nodes</button>
            <button className="px-3 py-1.5 rounded bg-blue-600" onClick={refresh} disabled={loading}>{loading?'Refreshing…':'Refresh'}</button>
          </div>
        </header>

        <Panel title="Naming Service" badge={health}>
          <StatCard label="Status" value={health} />
        </Panel>

        <Panel title="Cluster Status" badge={`Naming: ${health}`}>
          <StatCard label="Total Nodes" value={nodes.length} />
          <StatCard label="UP" value={<span className="text-green-400">{upDown.up}</span>} />
          <StatCard label="DOWN" value={<span className="text-red-400">{upDown.down}</span>} />
          <StatCard label="Main" value={nodes.filter(n=>n.role==='MAIN').length} />
          <StatCard label="Backup" value={nodes.filter(n=>n.role==='BACKUP').length} />
          <StatCard label="Replica" value={nodes.filter(n=>n.role==='REPLICA').length} />
          <StatCard label="Total Files" value={files.length} />
        </Panel>

        <div className="p-4 rounded-lg bg-neutral-900 border border-neutral-800">
          <NodeTable nodes={nodes} onRefresh={refresh} loading={loading} error={error} />
        </div>

        <div className="p-4 rounded-lg bg-neutral-900 border border-neutral-800 space-y-4">
          <UploadDropzone onUploaded={()=>setRefreshToken(t=>t+1)} onFilesAdded={onFilesAdded} />
          <FileExplorer files={files} onDelete={onDeleteFile} />
        </div>

        <Panel title="Replication Queue" badge={`items: ${queue.length}`}>
          <div className="p-4 rounded bg-neutral-900 border border-neutral-800">
            {queue.length === 0 ? (
              <div className="text-sm text-neutral-400">Tidak ada item pending</div>
            ) : (
              <ul className="text-sm font-mono space-y-1">
                {queue.map(q => (
                  <li key={q.id}>
                    [{q.status}] {q.file_key} → {q.target_node_id} from {q.source_node_id} @ {new Date(q.created_at).toLocaleString()}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </Panel>
      </div>
    </div>
  )
}
