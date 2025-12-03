import { useMemo, useState } from 'react'
import { ext, kindByExt, iconForKind, formatBytes } from '@/lib/files'
import type { FileItem } from '@/lib/types'

export default function FileExplorer({ files, onDelete, viewDefault='grid' }:{ files: FileItem[]; onDelete?: (id:string)=>void; viewDefault?: 'grid'|'list' }){
  const [view, setView] = useState<'grid'|'list'>(viewDefault)
  const [error] = useState<string|null>(null)

  const items = useMemo(()=>files.map(f=>({
    ...f,
    ext: ext(f.filename),
    kind: kindByExt(ext(f.filename)),
  })),[files])

  function remove(id:string){ onDelete?.(id) }

  if (view === 'grid') {
    return (
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Files</h2>
          <div className="flex gap-2">
            <button className={`px-3 py-1.5 rounded ${view==='grid'?'bg-gray-800 text-white':'bg-gray-200'}`} onClick={()=>setView('grid')}>Grid</button>
            <button className={`px-3 py-1.5 rounded ${view==='list'?'bg-gray-800 text-white':'bg-gray-200'}`} onClick={()=>setView('list')}>List</button>
          </div>
        </div>
        {error && <div className="text-red-600">{error}</div>}
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {items.map(it=> (
            <div key={it.id} className="p-3 bg-white rounded shadow flex flex-col gap-2">
              <div className="text-3xl">{iconForKind(it.kind)}</div>
              <div className="text-sm font-medium truncate" title={it.filename}>{it.filename}</div>
              <div className="text-xs text-gray-500">{formatBytes(it.size)}</div>
              <div className="flex gap-2">
                <a className="px-2 py-1 rounded bg-blue-600 text-white text-xs" href={`#`}>Download</a>
                <button className="px-2 py-1 rounded bg-red-600 text-white text-xs" onClick={()=>remove(it.id)}>Delete</button>
              </div>
            </div>
          ))}
          {items.length===0 && <div className="text-sm text-gray-500">Belum ada file</div>}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Files</h2>
        <div className="flex gap-2">
          <button className={`px-3 py-1.5 rounded ${view==='grid'?'bg-gray-200':'bg-gray-800 text-white'}`} onClick={()=>setView('grid')}>Grid</button>
          <button className={`px-3 py-1.5 rounded ${view==='list'?'bg-gray-800 text-white':'bg-gray-200'}`} onClick={()=>setView('list')}>List</button>
        </div>
      </div>
      {error && <div className="text-red-600">{error}</div>}
      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-left border-b">
              <th className="p-2">Nama</th>
              <th className="p-2">Tipe</th>
              <th className="p-2">Ukuran</th>
              <th className="p-2">Dibuat</th>
              <th className="p-2">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {items.map(it=> (
              <tr key={it.id} className="border-b">
                <td className="p-2">{it.filename}</td>
                <td className="p-2">{it.kind}</td>
                <td className="p-2">{formatBytes(it.size)}</td>
                <td className="p-2">{new Date(it.createdAt).toLocaleString()}</td>
                <td className="p-2">
                  <div className="flex gap-2">
                    <a className="px-2 py-1 rounded bg-blue-600 text-white text-xs" href={`/api/files/${it.id}/download`}>Download</a>
                    <button className="px-2 py-1 rounded bg-red-600 text-white text-xs" onClick={()=>remove(it.id)}>Delete</button>
                  </div>
                </td>
              </tr>
            ))}
            {items.length===0 && <tr><td className="p-2" colSpan={5}>Belum ada file</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}