export default function Panel({ title, badge, children }:{ title: string; badge?: string; children: React.ReactNode }){
  return (
    <div className="p-4 rounded-lg bg-neutral-900 border border-neutral-800">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-lg font-semibold text-white">{title}</h2>
        {badge && <span className="px-2 py-0.5 rounded text-xs bg-green-800 text-green-200">{badge}</span>}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 text-sm text-neutral-200">
        {children}
      </div>
    </div>
  )
}