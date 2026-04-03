import { useEffect, useState } from 'react'
import { useAuth } from '../App'

interface CaptureSession {
  id: string
  player: string
  startedAt: string
  endedAt?: string
}

interface Props {
  wsRef: React.RefObject<WebSocket | null>
  recording: boolean
  onClose: () => void
  onViewCapture: (id: string) => void
}

export default function CaptureModal({ wsRef, recording, onClose, onViewCapture }: Props) {
  const { user } = useAuth()
  const [captures, setCaptures] = useState<CaptureSession[]>([])
  const [loading, setLoading] = useState(true)

  const authHeaders = (): Record<string, string> => {
    if (!user?.token) return {}
    return { Authorization: `Bearer ${user.token}` }
  }

  useEffect(() => {
    fetch('/api/captures', { headers: authHeaders() })
      .then(r => r.ok ? r.json() : [])
      .then((data: CaptureSession[]) => { setCaptures(data || []); setLoading(false) })
  }, [recording])

  const startCapture = () => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'start_capture' }))
    }
  }

  const stopCapture = () => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'stop_capture' }))
    }
  }

  const formatDate = (d: string) => {
    const date = new Date(d)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  const formatDuration = (start: string, end?: string) => {
    const s = new Date(start).getTime()
    const e = end ? new Date(end).getTime() : Date.now()
    const mins = Math.round((e - s) / 60000)
    if (mins < 1) return '<1 min'
    if (mins < 60) return `${mins} min`
    return `${Math.floor(mins / 60)}h ${mins % 60}m`
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-[#111] border border-[#444] rounded-lg w-full max-w-lg max-h-[80vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between p-4 border-b border-[#333]">
          <h2 className="text-amber-400 font-mono font-bold">Session Capture</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white text-lg">&times;</button>
        </div>

        <div className="p-4 border-b border-[#333]">
          {recording ? (
            <div className="flex items-center gap-3">
              <div className="w-3 h-3 bg-red-500 rounded-full animate-pulse" />
              <span className="text-red-400 font-mono text-sm">Recording...</span>
              <button
                onClick={stopCapture}
                className="ml-auto px-4 py-1.5 bg-red-700 text-white rounded text-xs font-mono hover:bg-red-600"
              >
                Stop Recording
              </button>
            </div>
          ) : (
            <button
              onClick={startCapture}
              className="w-full py-2 bg-amber-700 text-white rounded text-sm font-mono hover:bg-amber-600"
            >
              Start Recording
            </button>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-4">
          <h3 className="text-gray-400 text-xs uppercase mb-2 font-mono">Previous Captures</h3>
          {loading ? (
            <div className="text-gray-600 text-center text-xs py-4">Loading...</div>
          ) : captures.length === 0 ? (
            <div className="text-gray-600 text-center text-xs py-4">No captures yet</div>
          ) : (
            <div className="space-y-1">
              {captures.map(c => (
                <button
                  key={c.id}
                  onClick={() => onViewCapture(c.id)}
                  className="w-full text-left px-3 py-2 bg-[#0a0a0a] border border-[#222] rounded hover:border-amber-500 text-xs font-mono"
                >
                  <div className="flex justify-between items-center">
                    <span className="text-gray-300">{c.player}</span>
                    <span className="text-gray-600">{formatDuration(c.startedAt, c.endedAt)}</span>
                  </div>
                  <div className="text-gray-600 mt-0.5">{formatDate(c.startedAt)}</div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
