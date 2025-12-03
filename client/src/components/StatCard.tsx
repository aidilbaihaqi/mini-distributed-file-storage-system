export default function StatCard({ label, value }:{ label: string; value: React.ReactNode }){
  return (
    <div className="rounded-md bg-neutral-800 p-3">
      <div className="text-xs text-neutral-400">{label}</div>
      <div className="text-xl font-semibold text-white">{value}</div>
    </div>
  )
}