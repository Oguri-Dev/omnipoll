import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'
import { Terminal, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react'

interface LogEntry {
  timestamp: string
  level: string
  message: string
}

interface LogsResponse {
  success: boolean
  data: LogEntry[]
  page: number
  pages: number
  total: number
  limit: number
}

export default function Logs() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(100)
  const [level, setLevel] = useState('')

  const {
    data: response,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['logs', page, pageSize, level],
    queryFn: () => api.getLogs(level || undefined, page, pageSize),
    refetchInterval: 3000,
  })

  const logs = response?.data || []
  const totalPages = response?.pages || 1
  const total = response?.total || 0

  const levelColors: Record<string, string> = {
    ERROR: 'text-red-400',
    WARN: 'text-yellow-400',
    INFO: 'text-green-400',
    DEBUG: 'text-blue-400',
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Terminal className="w-6 h-6" />
          Logs
        </h1>
        <button onClick={() => refetch()} className="p-2 rounded hover:bg-gray-100">
          <RefreshCw className="w-5 h-5" />
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 flex gap-4 items-end">
        <div className="flex-1">
          <label className="block text-sm font-medium mb-2">Log Level</label>
          <select
            value={level}
            onChange={(e) => {
              setLevel(e.target.value)
              setPage(1)
            }}
            className="border rounded px-3 py-2 w-full"
          >
            <option value="">All Levels</option>
            <option value="ERROR">ERROR</option>
            <option value="WARN">WARN</option>
            <option value="INFO">INFO</option>
            <option value="DEBUG">DEBUG</option>
          </select>
        </div>

        <div className="flex-1">
          <label className="block text-sm font-medium mb-2">Page Size</label>
          <select
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value))
              setPage(1)
            }}
            className="border rounded px-3 py-2 w-full"
          >
            <option value={50}>50 items</option>
            <option value={100}>100 items</option>
            <option value={200}>200 items</option>
            <option value={500}>500 items</option>
          </select>
        </div>
      </div>

      {/* Results info */}
      <div className="bg-blue-50 rounded-lg p-4">
        <p className="text-sm">
          Showing <strong>{logs.length}</strong> of <strong>{total}</strong> log entries
          {total > 0 && ` (Page ${page} of ${totalPages})`}
        </p>
      </div>

      {/* Loading state */}
      {isLoading && (
        <div className="flex items-center justify-center h-64">
          <div className="text-gray-500">Loading logs...</div>
        </div>
      )}

      {/* Logs terminal view */}
      <div className="bg-gray-900 rounded-lg p-4 h-[500px] overflow-auto font-mono text-sm border border-gray-700">
        {!isLoading && logs.length > 0 ? (
          logs.map((log: LogEntry, index: number) => (
            <div
              key={index}
              className={`py-1 hover:bg-gray-800 px-2 ${levelColors[log.level] || 'text-white'}`}
            >
              <span className="text-gray-500">{new Date(log.timestamp).toLocaleTimeString()}</span>{' '}
              <span className="uppercase font-semibold">[{log.level}]</span>{' '}
              <span className="text-gray-300">{log.message}</span>
            </div>
          ))
        ) : !isLoading ? (
          <div className="text-gray-500">No logs found</div>
        ) : null}
      </div>

      {/* Pagination */}
      {!isLoading && totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="p-2 rounded border hover:bg-gray-100 disabled:opacity-50"
          >
            <ChevronLeft className="w-5 h-5" />
          </button>
          <span className="text-sm">
            Page {page} of {totalPages}
          </span>
          <button
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page === totalPages}
            className="p-2 rounded border hover:bg-gray-100 disabled:opacity-50"
          >
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>
      )}
    </div>
  )
}
