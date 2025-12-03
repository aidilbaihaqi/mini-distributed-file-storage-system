import type { NodeInfo } from "@/lib/types"
import NodeBadge from "./NodeBadge"

export default function NodeTable({ nodes, onRefresh, loading, error }:{ nodes: NodeInfo[]; onRefresh?: ()=>void; loading?: boolean; error?: string|null }) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Nodes</h2>
        <button className="px-3 py-1.5 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50" onClick={onRefresh} disabled={loading}>
          {loading ? 'Refreshing…' : 'Refresh status'}
        </button>
      </div>
      {error && <div className="text-red-600">{error}</div>}
      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-left border-b">
              <th className="p-2">ID</th>
              <th className="p-2">Address</th>
              <th className="p-2">Status</th>
              <th className="p-2">Last Heartbeat</th>
            </tr>
          </thead>
          <tbody>
            {nodes.map(n => (
              <tr key={n.id} className="border-b">
                <td className="p-2 font-mono">{n.id}</td>
                <td className="p-2">{n.address}</td>
                <td className="p-2"><NodeBadge n={n} /></td>
                <td className="p-2">{new Date(n.lastHeartbeat).toLocaleString()}</td>
              </tr>
            ))}
            {nodes.length === 0 && <tr><td className="p-2" colSpan={4}>No nodes</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}