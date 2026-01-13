import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'
import { Database, RefreshCw, Trash2, Eye, ChevronLeft, ChevronRight } from 'lucide-react'

interface Event {
  _id: string
  source: string
  fechaHora: string
  unitName: string
  name?: string
  amountGrams?: number
  pelletFishMin?: number
  fishCount?: number
  pesoProm?: number
  biomasa?: number
  pelletPK?: number
  feedName?: string
  siloName?: string
  doserName?: string
  gramsPerSec?: number
  kgTonMin?: number
  marca?: number
  dia?: string
  inicio?: string
  fin?: string
  dif?: number
  ingestedAt: string
}

export default function Events() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [filters, setFilters] = useState({
    source: '',
    unitName: '',
    startDate: '',
    endDate: '',
  })
  const [selectedEvent, setSelectedEvent] = useState<Event | null>(null)
  const [showModal, setShowModal] = useState(false)

  const {
    data: response,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['events', page, pageSize, filters],
    queryFn: () => api.getEvents(page, pageSize, filters),
    refetchInterval: 5000,
  })

  const deleteEventMutation = useMutation({
    mutationFn: api.deleteEvent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      setSelectedEvent(null)
      setShowModal(false)
    },
  })

  const handleFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFilters((prev) => ({ ...prev, [name]: value }))
    setPage(1)
  }

  const handleDeleteEvent = async (eventId: string) => {
    if (window.confirm('Are you sure you want to delete this event?')) {
      deleteEventMutation.mutate(eventId)
    }
  }

  const handleViewDetails = (event: Event) => {
    setSelectedEvent(event)
    setShowModal(true)
  }

  const events = response?.data || []
  const totalPages = response?.pages || 1
  const total = response?.total || 0

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Database className="w-6 h-6" />
          Events
        </h1>
        <button onClick={() => refetch()} className="p-2 rounded hover:bg-gray-100">
          <RefreshCw className="w-5 h-5" />
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 space-y-4">
        <h2 className="font-semibold">Filters</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <input
            type="text"
            name="source"
            placeholder="Source (e.g., Akva)"
            value={filters.source}
            onChange={handleFilterChange}
            className="border rounded px-3 py-2"
          />
          <input
            type="text"
            name="unitName"
            placeholder="Unit Name"
            value={filters.unitName}
            onChange={handleFilterChange}
            className="border rounded px-3 py-2"
          />
          <input
            type="date"
            name="startDate"
            value={filters.startDate}
            onChange={handleFilterChange}
            className="border rounded px-3 py-2"
          />
          <input
            type="date"
            name="endDate"
            value={filters.endDate}
            onChange={handleFilterChange}
            className="border rounded px-3 py-2"
          />
        </div>
        <div className="flex gap-2">
          <select
            value={pageSize}
            onChange={(e) => {
              setPageSize(Number(e.target.value))
              setPage(1)
            }}
            className="border rounded px-3 py-2"
          >
            <option value={10}>10 items</option>
            <option value={50}>50 items</option>
            <option value={100}>100 items</option>
            <option value={250}>250 items</option>
          </select>
        </div>
      </div>

      {/* Results info */}
      <div className="bg-blue-50 rounded-lg p-4">
        <p className="text-sm">
          Showing <strong>{events.length}</strong> of <strong>{total}</strong> events
          {total > 0 && ` (Page ${page} of ${totalPages})`}
        </p>
      </div>

      {/* Error state */}
      {isError && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          Error loading events: {error instanceof Error ? error.message : 'Unknown error'}
        </div>
      )}

      {/* Loading state */}
      {isLoading && (
        <div className="flex items-center justify-center h-64">
          <div className="text-gray-500">Loading events...</div>
        </div>
      )}

      {/* Events table */}
      {!isLoading && events.length > 0 && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">Source</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">Unit</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">
                    Fecha/Hora
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">Biomasa</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">Feed</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-700">Actions</th>
                </tr>
              </thead>
              <tbody>
                {events.map((event: Event) => (
                  <tr key={event._id} className="border-b hover:bg-gray-50">
                    <td className="px-6 py-3 text-sm">{event.source}</td>
                    <td className="px-6 py-3 text-sm">{event.unitName}</td>
                    <td className="px-6 py-3 text-sm">
                      {new Date(event.fechaHora).toLocaleString()}
                    </td>
                    <td className="px-6 py-3 text-sm">
                      {event.biomasa ? event.biomasa.toFixed(0) : '-'}
                    </td>
                    <td className="px-6 py-3 text-sm">{event.feedName || '-'}</td>
                    <td className="px-6 py-3 text-sm">
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleViewDetails(event)}
                          className="text-blue-600 hover:text-blue-800 p-1"
                          title="View details"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteEvent(event._id)}
                          disabled={deleteEventMutation.isPending}
                          className="text-red-600 hover:text-red-800 p-1 disabled:opacity-50"
                          title="Delete event"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!isLoading && events.length === 0 && (
        <div className="bg-gray-50 rounded-lg p-8 text-center">
          <Database className="w-12 h-12 text-gray-400 mx-auto mb-2" />
          <p className="text-gray-500">No events found</p>
        </div>
      )}

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

      {/* Detail Modal */}
      {showModal && selectedEvent && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 max-h-screen overflow-y-auto">
            <div className="flex justify-between items-start mb-4">
              <h2 className="text-xl font-bold">Event Details</h2>
              <button
                onClick={() => setShowModal(false)}
                className="text-gray-500 hover:text-gray-700 text-2xl"
              >
                ×
              </button>
            </div>

            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-gray-600">ID</p>
                  <p className="font-mono text-sm">{selectedEvent._id}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Source</p>
                  <p className="font-semibold">{selectedEvent.source}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Unit Name</p>
                  <p className="font-semibold">{selectedEvent.unitName}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Fecha/Hora</p>
                  <p className="font-semibold">
                    {new Date(selectedEvent.fechaHora).toLocaleString()}
                  </p>
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="font-semibold mb-3">Event Data</h3>
                <div className="grid grid-cols-2 gap-3">
                  {Object.entries(selectedEvent)
                    .filter(([key]) => !['_id', 'source', 'unitName', 'fechaHora', 'ingestedAt'].includes(key))
                    .map(([key, value]) => (
                      <div key={key} className="text-sm">
                        <p className="text-gray-600">{key}</p>
                        <p className="font-mono">{String(value)}</p>
                      </div>
                    ))}
                </div>
              </div>

              <div className="border-t pt-4">
                <p className="text-xs text-gray-500">
                  Ingested at: {new Date(selectedEvent.ingestedAt).toLocaleString()}
                </p>
              </div>
            </div>

            <div className="flex gap-2 mt-6">
              <button
                onClick={() => setShowModal(false)}
                className="flex-1 px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded"
              >
                Close
              </button>
              <button
                onClick={() => {
                  handleDeleteEvent(selectedEvent._id)
                }}
                disabled={deleteEventMutation.isPending}
                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded disabled:opacity-50"
              >
                {deleteEventMutation.isPending ? 'Deleting...' : 'Delete Event'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
