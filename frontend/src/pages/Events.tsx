import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'
import { Database, RefreshCw } from 'lucide-react'

interface Event {
  _id: string
  source: string
  fechaHora: string
  unitName: string
  payload: {
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
  }
  ingestedAt: string
}

export default function Events() {
  const {
    data: events,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['events'],
    queryFn: api.getEvents,
    refetchInterval: 5000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-500">Cargando eventos...</div>
      </div>
    )
  }

  const eventList: Event[] = events || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Database size={28} className="text-blue-600" />
          <div>
            <h1 className="text-2xl font-bold">Eventos</h1>
            <p className="text-sm text-gray-500">Últimos 100 eventos ingeridos desde Akva</p>
          </div>
        </div>
        <button
          onClick={() => refetch()}
          className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          <RefreshCw size={18} />
          Refrescar
        </button>
      </div>

      {eventList.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-12 text-center">
          <Database size={64} className="mx-auto text-gray-300 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No hay eventos todavía</h3>
          <p className="text-gray-500">
            Los eventos aparecerán aquí cuando el worker comience a ingerir datos desde Akva
          </p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Fecha/Hora
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Unidad
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Nombre
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Gramos
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Biomasa
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Peces
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Peso Prom
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Feed
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Silo
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Ingerido
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {eventList.map((event: Event) => (
                  <tr key={event._id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                      {new Date(event.fechaHora).toLocaleString('es-CL')}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                      {event.unitName}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {event.payload?.name || '-'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900">
                      {event.payload?.amountGrams?.toFixed(2) || '0'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900">
                      {event.payload?.biomasa?.toFixed(2) || '0'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900">
                      {event.payload?.fishCount || '0'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-right text-gray-900">
                      {event.payload?.pesoProm?.toFixed(3) || '0'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {event.payload?.feedName || '-'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-600">
                      {event.payload?.siloName || '-'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">
                      {new Date(event.ingestedAt).toLocaleTimeString('es-CL')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="bg-gray-50 px-4 py-3 border-t border-gray-200">
            <p className="text-sm text-gray-700">
              Mostrando <span className="font-medium">{eventList.length}</span> eventos
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
