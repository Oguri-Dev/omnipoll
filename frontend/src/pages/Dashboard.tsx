import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Database, Radio, Activity, Play, Square, RotateCcw } from 'lucide-react'
import StatusCard from '../components/StatusCard'
import ConnectionStatus from '../components/ConnectionStatus'
import { api } from '../services/api'

export default function Dashboard() {
  const queryClient = useQueryClient()

  const {
    data: status,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['status'],
    queryFn: api.getStatus,
    refetchInterval: 5000,
    retry: 2,
    staleTime: 3000,
  })

  const startWorker = useMutation({
    mutationFn: api.startWorker,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['status'] }),
  })

  const stopWorker = useMutation({
    mutationFn: api.stopWorker,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['status'] }),
  })

  const resetWatermark = useMutation({
    mutationFn: api.resetWatermark,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['status'] }),
  })

  if (isLoading) {
    return <div className="p-4">Loading...</div>
  }

  if (isError) {
    return (
      <div className="p-4">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          Error loading status: {error instanceof Error ? error.message : 'Unknown error'}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {/* Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatusCard
          title="Last FechaHora"
          value={status?.lastFechaHora || 'N/A'}
          icon={<Activity size={24} />}
          status="neutral"
        />
        <StatusCard
          title="Events Today"
          value={status?.eventsToday || 0}
          icon={<Radio size={24} />}
          status="success"
        />
        <StatusCard
          title="Ingestion Rate"
          value={`${status?.ingestionRate || 0}/min`}
          icon={<Activity size={24} />}
          status="success"
        />
        <StatusCard
          title="Total Events"
          value={status?.totalEvents || 0}
          icon={<Database size={24} />}
          status="neutral"
        />
      </div>

      {/* Connections */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Connections</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <ConnectionStatus
            name="SQL Server (Akva)"
            connected={status?.connections?.sqlServer || false}
          />
          <ConnectionStatus name="MQTT Broker" connected={status?.connections?.mqtt || false} />
          <ConnectionStatus name="MongoDB" connected={status?.connections?.mongodb || false} />
        </div>
      </div>

      {/* Worker Controls */}
      <div>
        <h2 className="text-lg font-semibold mb-4">Worker Controls</h2>
        <div className="flex gap-4">
          <button
            onClick={() => startWorker.mutate()}
            disabled={status?.workerRunning || startWorker.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
          >
            <Play size={20} />
            Start Worker
          </button>
          <button
            onClick={() => stopWorker.mutate()}
            disabled={!status?.workerRunning || stopWorker.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
          >
            <Square size={20} />
            Stop Worker
          </button>
          <button
            onClick={() => {
              if (
                confirm(
                  'Are you sure you want to reset the watermark? This will re-process all data.'
                )
              ) {
                resetWatermark.mutate()
              }
            }}
            disabled={status?.workerRunning || resetWatermark.isPending}
            className="flex items-center gap-2 px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600 disabled:opacity-50"
          >
            <RotateCcw size={20} />
            Reset Watermark
          </button>
        </div>
      </div>
    </div>
  )
}
