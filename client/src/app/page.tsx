"use client"
import { useEffect, useMemo, useState } from "react"
import type { NodeInfo, FileItem } from "@/lib/types"
import NodeTable from "@/components/NodeTable"
import UploadDropzone from "@/components/UploadDropzone"
import FileExplorer from "@/components/FileExplorer"
import Panel from "@/components/Panel"
import StatCard from "@/components/StatCard"
import { mockHealth, mockNodes, mockFiles } from "@/lib/mock"

export default function Home() {
  const [nodes, setNodes] = useState<NodeInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string|null>(null)
  const [health, setHealth] = useState("unknown")
  const [refreshToken, setRefreshToken] = useState(0)
  const [files, setFiles] = useState<FileItem[]>(mockFiles)

  useEffect(()=>{
    setNodes(mockNodes)
    setHealth(mockHealth.status)
  },[])

  const upDown = useMemo(()=>({
    up: nodes.filter(n=>n.status==='UP').length,
    down: nodes.filter(n=>n.status!=='UP').length,
  }),[nodes])

  async function refresh(){
    setLoading(true); setError(null)
    try { setNodes(prev=>prev.map(n=> ({ ...n, lastHeartbeat: new Date().toISOString() }))) } catch(e:any){ setError(e.message) } finally { setLoading(false) }
  }

  function onFilesAdded(add: File[]){
    const now = Date.now()
    const newItems: FileItem[] = add.map((f,i)=>({
      id: `dummy-${now}-${i}`,
      filename: f.name,
      size: f.size,
      createdAt: new Date().toISOString(),
      replicas: ['node-1'],
    }))
    setFiles(prev=>[...newItems, ...prev])
  }

  function onDeleteFile(id: string){
    setFiles(prev=>prev.filter(f=>f.id!==id))
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
          <StatCard label="Uptime" value={mockHealth.uptime} />
          <StatCard label="Version" value={mockHealth.version} />
          <StatCard label="Requests" value={mockHealth.requests} />
          <StatCard label="Base URL" value={<span className="font-mono">{mockHealth.baseUrl}</span>} />
        </Panel>

        <Panel title="Cluster Status" badge={`Naming: ${health}`}>
          <StatCard label="Total Nodes" value={nodes.length} />
          <StatCard label="UP" value={<span className="text-green-400">{upDown.up}</span>} />
          <StatCard label="DOWN" value={<span className="text-red-400">{upDown.down}</span>} />
          <StatCard label="Main" value={nodes.filter(n=>n.role==='MAIN').length} />
          <StatCard label="Backup" value={nodes.filter(n=>n.role==='BACKUP').length} />
          <StatCard label="Replica" value={nodes.filter(n=>n.role==='REPLICA').length} />
          <StatCard label="Total Files" value={mockHealth.totalFiles ?? '—'} />
          <StatCard label="Requests" value={mockHealth.requests} />
          <StatCard label="Uptime" value={mockHealth.uptime} />
          <StatCard label="Version" value={mockHealth.version} />
        </Panel>

        <div className="p-4 rounded-lg bg-neutral-900 border border-neutral-800">
          <NodeTable nodes={nodes} onRefresh={refresh} loading={loading} error={error} />
        </div>

        <div className="p-4 rounded-lg bg-neutral-900 border border-neutral-800 space-y-4">
          <UploadDropzone onUploaded={()=>setRefreshToken(t=>t+1)} onFilesAdded={onFilesAdded} />
          <FileExplorer files={files} onDelete={onDeleteFile} />
        </div>
      </div>
    </div>
  )
}
