"use client"
import { useEffect, useState } from 'react'
import { mockLogs } from '@/lib/mock'

export default function LogsPage(){
  const [logs, setLogs] = useState<string[]>([])
  const [error, setError] = useState<string|null>(null)
  useEffect(()=>{ try{ setLogs(mockLogs) } catch(e:any){ setError(e.message) } },[])
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6 space-y-4">
        <h1 className="text-2xl font-bold">Logs</h1>
        {error && <div className="text-red-600">{error}</div>}
        <div className="p-4 bg-white rounded shadow">
          <pre className="text-sm whitespace-pre-wrap">{logs.join('\n')}</pre>
        </div>
      </div>
    </div>
  )
}