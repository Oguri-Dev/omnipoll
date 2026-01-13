import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'
import { Database, RefreshCw, Trash2, Eye, ChevronLeft, ChevronRight } from 'lucide-react'

interface Event {
  _id: string
  Source: string
  FechaHora: string
  UnitName: string
  Payload?: {
    name?: string // Centro
    amountGrams?: number // Gramos
    pelletFishMin?: number
    fishCount?: number // Peces
    pesoProm?: number // Peso Promedio
    biomasa?: number
    pelletPK?: number
    feedName?: string // Alimento
    siloName?: string // Silo
    doserName?: string // Dosificador
    gramsPerSec?: number // Gramos por segundo
    kgTonMin?: number
    marca?: number
    dia?: string
    inicio?: string
    fin?: string
    dif?: number // Duración en segundos
  }
  IngestedAt: string
}

export default function Events() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [filters, setFilters] = useState({
    centro: '',
    jaula: '',
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
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <input
            type="text"
            name="centro"
            placeholder="Centro (e.g., Huelden YC 2025)"
            value={filters.centro}
            onChange={handleFilterChange}
            className="border rounded px-3 py-2"
          />
          <input
            type="text"
            name="jaula"
            placeholder="Jaula (e.g., 101)"
            value={filters.jaula}
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
            <table className="w-full text-xs">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Centro</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Jaula</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Fecha/Hora</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Día</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Inicio</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Fin</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Duración</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Gramos</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Pellet/Fish/Min</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Peces</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Peso Prom</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Biomasa</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Pellet PK</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Alimento</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Silo</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Dosificador</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Gramos/Seg</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Kg Ton/Min</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Marca</th>
                  <th className="px-2 py-2 text-left text-xs font-medium text-gray-700">Actions</th>
                </tr>
              </thead>
              <tbody>
                {events.map((event: Event) => (
                  <tr key={event._id} className="border-b hover:bg-gray-50">
                    <td className="px-2 py-2 text-xs">{event.Payload?.name || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.UnitName}</td>
                    <td className="px-2 py-2 text-xs">{new Date(event.FechaHora).toLocaleString()}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.dia || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.inicio || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.fin || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.dif ? `${event.Payload.dif}s` : '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.amountGrams?.toFixed(2) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.pelletFishMin?.toFixed(4) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.fishCount?.toFixed(0) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.pesoProm?.toFixed(2) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.biomasa?.toFixed(0) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.pelletPK?.toFixed(3) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.feedName || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.siloName || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.doserName || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.gramsPerSec?.toFixed(2) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.kgTonMin?.toFixed(6) || '-'}</td>
                    <td className="px-2 py-2 text-xs">{event.Payload?.marca || 0}</td>
                    <td className="px-2 py-2 text-xs\">
                      <div className="flex gap-1">
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
                  <p className="font-mono text-xs break-all">{selectedEvent._id}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Centro</p>
                  <p className="font-semibold">{selectedEvent.Payload?.name || '-'}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Jaula</p>
                  <p className="font-semibold">{selectedEvent.UnitName}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Fecha/Hora</p>
                  <p className="font-semibold">
                    {new Date(selectedEvent.FechaHora).toLocaleString()}
                  </p>
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="font-semibold mb-3">Datos de Alimentación</h3>
                <div className="grid grid-cols-2 gap-3">
                  <div className="text-sm">
                    <p className="text-gray-600">Gramos</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.amountGrams?.toFixed(2) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Peces</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.fishCount?.toFixed(0) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Biomasa (g)</p>
                    <p className="font-mono">{selectedEvent.Payload?.biomasa?.toFixed(0) || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Peso Promedio (g)</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.pesoProm?.toFixed(2) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Pellet/Fish/Min</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.pelletFishMin?.toFixed(4) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Pellet PK</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.pelletPK?.toFixed(3) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Gramos/Seg</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.gramsPerSec?.toFixed(2) || '-'}
                    </p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Kg Ton/Min</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.kgTonMin?.toFixed(6) || '-'}
                    </p>
                  </div>
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="font-semibold mb-3">Equipamiento</h3>
                <div className="grid grid-cols-2 gap-3">
                  <div className="text-sm">
                    <p className="text-gray-600">Alimento</p>
                    <p className="font-semibold">{selectedEvent.Payload?.feedName || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Silo</p>
                    <p className="font-semibold">{selectedEvent.Payload?.siloName || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Dosificador</p>
                    <p className="font-semibold">{selectedEvent.Payload?.doserName || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Marca</p>
                    <p className="font-mono">{selectedEvent.Payload?.marca || 0}</p>
                  </div>
                </div>
              </div>

              <div className="border-t pt-4">
                <h3 className="font-semibold mb-3">Tiempos</h3>
                <div className="grid grid-cols-3 gap-3">
                  <div className="text-sm">
                    <p className="text-gray-600">Inicio</p>
                    <p className="font-mono">{selectedEvent.Payload?.inicio || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Fin</p>
                    <p className="font-mono">{selectedEvent.Payload?.fin || '-'}</p>
                  </div>
                  <div className="text-sm">
                    <p className="text-gray-600">Duración</p>
                    <p className="font-mono">
                      {selectedEvent.Payload?.dif ? `${selectedEvent.Payload.dif}s` : '-'}
                    </p>
                  </div>
                </div>
              </div>

              <div className="border-t pt-4">
                <p className="text-xs text-gray-500">
                  Ingresado: {new Date(selectedEvent.IngestedAt).toLocaleString()}
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
