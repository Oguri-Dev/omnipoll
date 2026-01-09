import { CheckCircle, XCircle, Loader } from 'lucide-react'

interface ConnectionStatusProps {
  name: string
  connected: boolean
  loading?: boolean
  onTest?: () => void
}

export default function ConnectionStatus({
  name,
  connected,
  loading,
  onTest,
}: ConnectionStatusProps) {
  return (
    <div className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
      <div className="flex items-center gap-3">
        {loading ? (
          <Loader className="animate-spin text-gray-400" size={20} />
        ) : connected ? (
          <CheckCircle className="text-green-500" size={20} />
        ) : (
          <XCircle className="text-red-500" size={20} />
        )}
        <span className="font-medium">{name}</span>
      </div>
      {onTest && (
        <button
          onClick={onTest}
          disabled={loading}
          className="px-3 py-1 text-sm bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
        >
          Test
        </button>
      )}
    </div>
  )
}
