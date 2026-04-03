import { useEffect, useState } from 'react'
import { useAuth } from '../App'

interface CaptureLine {
  time: string
  type: string
  text: string
}

interface CaptureData {
  id: string
  player: string
  startedAt: string
  endedAt?: string
  lines: CaptureLine[]
}

const lineColors: Record<string, string> = {
  input: 'text-cyan-400',
  output: 'text-gray-300',
  system: 'text-yellow-400',
  broadcast: 'text-green-400',
}

export default function CaptureViewer({ captureId, onBack }: { captureId: string; onBack: () => void }) {
  const { user } = useAuth()
  const [data, setData] = useState<CaptureData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch(`/api/captures/${captureId}`, {
      headers: user?.token ? { Authorization: `Bearer ${user.token}` } : {},
    })
      .then(r => r.ok ? r.json() : null)
      .then(d => { setData(d); setLoading(false) })
  }, [captureId])

  const download = () => {
    if (!user?.token) return
    window.open(`/api/captures/${captureId}/text?token=${user.token}`, '_blank')
  }

  if (loading) return <div className="flex items-center justify-center h-full text-gray-600 font-mono">Loading capture...</div>
  if (!data) return <div className="flex items-center justify-center h-full text-gray-600 font-mono">Capture not found</div>

  return (
    <div className="flex flex-col h-full font-mono text-sm">
      <div className="flex items-center gap-3 px-4 py-2 bg-[#111] border-b border-[#333]">
        <button onClick={onBack} className="text-gray-400 hover:text-white text-sm">&larr; Back</button>
        <h2 className="text-amber-400 font-bold">Session Capture: {data.player}</h2>
        <span className="text-gray-600 text-xs">{new Date(data.startedAt).toLocaleDateString()}</span>
        <span className="text-gray-600 text-xs">{data.lines?.length || 0} lines</span>
        <button
          onClick={download}
          className="ml-auto px-3 py-1 bg-[#222] border border-[#444] rounded text-xs text-gray-300 hover:border-amber-500"
        >
          Download .txt
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-4 bg-[#0a0a0a]">
        {data.lines?.map((line, i) => (
          <div key={i} className="py-0.5 text-xs">
            {line.type === 'input' ? (
              <span className="text-cyan-400">&gt; {line.text}</span>
            ) : (
              <span className={lineColors[line.type] || 'text-gray-300'}>{line.text}</span>
            )}
          </div>
        ))}
        {(!data.lines || data.lines.length === 0) && (
          <div className="text-gray-600 text-center py-8">Empty capture</div>
        )}
      </div>
    </div>
  )
}
