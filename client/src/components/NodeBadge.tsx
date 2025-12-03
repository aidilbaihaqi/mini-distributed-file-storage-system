import type { NodeInfo } from "@/lib/types"

export default function NodeBadge({ n }: { n: NodeInfo }) {
  const s = n.status === 'UP' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
  const r = n.role === 'MAIN' ? 'bg-blue-100 text-blue-700' : n.role === 'BACKUP' ? 'bg-yellow-100 text-yellow-700' : n.role === 'REPLICA' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-700'
  return (
    <div className="flex gap-2">
      <span className={`px-2 py-0.5 rounded text-xs ${s}`}>{n.status}</span>
      <span className={`px-2 py-0.5 rounded text-xs ${r}`}>{n.role ?? 'UNKNOWN'}</span>
    </div>
  )
}