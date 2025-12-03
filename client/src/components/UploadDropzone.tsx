import { useCallback, useState } from 'react'

export default function UploadDropzone({ onUploaded, onFilesAdded }:{ onUploaded?: ()=>void; onFilesAdded?: (files: File[])=>void }){
  const [hover, setHover] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string|null>(null)

  const handleFiles = useCallback(async (files: FileList | null) => {
    if (!files || files.length === 0) return
    setBusy(true); setError(null)
    try {
      const arr = Array.from(files)
      onFilesAdded?.(arr)
      await new Promise(r=>setTimeout(r, 500))
      onUploaded?.()
    } catch (e:any){
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }, [onUploaded])

  return (
    <div
      className={`border-2 border-dashed rounded p-6 text-center ${hover ? 'border-blue-500 bg-blue-50' : 'border-gray-300'}`}
      onDragOver={(e)=>{e.preventDefault(); setHover(true)}}
      onDragLeave={()=>setHover(false)}
      onDrop={(e)=>{e.preventDefault(); setHover(false); handleFiles(e.dataTransfer.files)}}
    >
      <input type="file" multiple onChange={(e)=>handleFiles(e.target.files)} className="hidden" id="fileInput" />
      <label htmlFor="fileInput" className="cursor-pointer inline-block px-3 py-2 rounded bg-blue-600 text-white">
        {busy ? 'Uploading…' : 'Pilih file atau drag & drop ke sini'}
      </label>
      {error && <div className="mt-2 text-red-600">{error}</div>}
    </div>
  )
}