import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'

export default function Header() {
  const { data: status } = useQuery({
    queryKey: ['status'],
    queryFn: api.getStatus,
    refetchInterval: 5000,
  })

  return (
    <header className="bg-white border-b px-6 py-4 flex justify-between items-center">
      <div>
        <h2 className="text-lg font-semibold">Control Panel</h2>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-500">Worker:</span>
          <span
            className={`px-2 py-1 rounded text-xs font-medium ${
              status?.workerRunning ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`}
          >
            {status?.workerRunning ? 'Running' : 'Stopped'}
          </span>
        </div>
      </div>
    </header>
  )
}
