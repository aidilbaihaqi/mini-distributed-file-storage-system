export function formatBytes(n: number): string {
  const u = ['B','KB','MB','GB','TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++ }
  return `${v.toFixed(v < 10 ? 2 : 0)} ${u[i]}`
}

export function ext(filename: string): string {
  const i = filename.lastIndexOf('.')
  return i >= 0 ? filename.slice(i + 1).toLowerCase() : ''
}

export function kindByExt(e: string): string {
  if (!e) return 'unknown'
  const images = ['png','jpg','jpeg','gif','bmp','webp','tiff']
  const docs = ['pdf','doc','docx','xls','xlsx','csv','ppt','pptx','txt','md']
  const audio = ['mp3','wav','ogg','aac','flac']
  const video = ['mp4','mkv','avi','mov','webm']
  const archive = ['zip','rar','tar','gz','7z']
  const binary = ['exe','bin','iso','dmg','apk','wasm']
  if (images.includes(e)) return 'image'
  if (docs.includes(e)) return 'document'
  if (audio.includes(e)) return 'audio'
  if (video.includes(e)) return 'video'
  if (archive.includes(e)) return 'archive'
  if (binary.includes(e)) return 'binary'
  return 'unknown'
}

export function iconForKind(k: string): string {
  switch (k) {
    case 'image': return '🖼️'
    case 'document': return '📄'
    case 'audio': return '🎵'
    case 'video': return '🎬'
    case 'archive': return '🗜️'
    case 'binary': return '💾'
    default: return '📁'
  }
}