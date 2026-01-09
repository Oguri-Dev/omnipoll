import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'

export default function Logs() {
  const { data: logs, isLoading } = useQuery({
    queryKey: ['logs'],
    queryFn: api.getLogs,
    refetchInterval: 3000,
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Logs</h1>

      <div className="bg-gray-900 rounded-lg p-4 h-[600px] overflow-auto font-mono text-sm">
        {logs?.map((log: any, index: number) => (
          <div
            key={index}
            className={`py-1 ${
              log.level === 'error'
                ? 'text-red-400'
                : log.level === 'warn'
                ? 'text-yellow-400'
                : 'text-green-400'
            }`}
          >
            <span className="text-gray-500">{log.timestamp}</span>{' '}
            <span className="uppercase">[{log.level}]</span> {log.message}
          </div>
        ))}
      </div>
    </div>
  )
}
