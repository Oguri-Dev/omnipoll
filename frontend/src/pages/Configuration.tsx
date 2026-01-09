import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'

export default function Configuration() {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'sqlserver' | 'mqtt' | 'mongodb' | 'polling'>(
    'sqlserver'
  )

  const { data: config, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: api.getConfig,
  })

  const saveConfig = useMutation({
    mutationFn: api.saveConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config'] })
      alert('Configuration saved!')
    },
  })

  const testConnection = useMutation({
    mutationFn: api.testConnection,
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  const tabs = [
    { id: 'sqlserver', label: 'SQL Server' },
    { id: 'mqtt', label: 'MQTT' },
    { id: 'mongodb', label: 'MongoDB' },
    { id: 'polling', label: 'Polling' },
  ] as const

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Configuration</h1>

      {/* Tabs */}
      <div className="border-b">
        <nav className="flex gap-4">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-2 px-4 border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="bg-white rounded-lg shadow p-6">
        {activeTab === 'sqlserver' && (
          <SQLServerConfig
            config={config?.sqlServer}
            onSave={saveConfig.mutate}
            onTest={() => testConnection.mutate('sqlserver')}
          />
        )}
        {activeTab === 'mqtt' && (
          <MQTTConfig
            config={config?.mqtt}
            onSave={saveConfig.mutate}
            onTest={() => testConnection.mutate('mqtt')}
          />
        )}
        {activeTab === 'mongodb' && (
          <MongoDBConfig
            config={config?.mongodb}
            onSave={saveConfig.mutate}
            onTest={() => testConnection.mutate('mongodb')}
          />
        )}
        {activeTab === 'polling' && (
          <PollingConfig config={config?.polling} onSave={saveConfig.mutate} />
        )}
      </div>
    </div>
  )
}

// Config form components (simplified)
function SQLServerConfig({ config, onTest }: any) {
  return (
    <form className="space-y-4">
      <h3 className="text-lg font-medium">SQL Server Configuration</h3>
      <div className="grid grid-cols-2 gap-4">
        <input
          type="text"
          placeholder="Host"
          defaultValue={config?.host}
          className="border rounded p-2"
        />
        <input
          type="number"
          placeholder="Port"
          defaultValue={config?.port}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="Database"
          defaultValue={config?.database}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="User"
          defaultValue={config?.user}
          className="border rounded p-2"
        />
        <input type="password" placeholder="Password" className="border rounded p-2 col-span-2" />
      </div>
      <div className="flex gap-4">
        <button type="button" onClick={onTest} className="px-4 py-2 bg-gray-500 text-white rounded">
          Test Connection
        </button>
        <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded">
          Save
        </button>
      </div>
    </form>
  )
}

function MQTTConfig({ config, onTest }: any) {
  return (
    <form className="space-y-4">
      <h3 className="text-lg font-medium">MQTT Configuration</h3>
      <div className="grid grid-cols-2 gap-4">
        <input
          type="text"
          placeholder="Broker"
          defaultValue={config?.broker}
          className="border rounded p-2"
        />
        <input
          type="number"
          placeholder="Port"
          defaultValue={config?.port}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="Topic"
          defaultValue={config?.topic}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="Client ID"
          defaultValue={config?.clientId}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="User"
          defaultValue={config?.user}
          className="border rounded p-2"
        />
        <input type="password" placeholder="Password" className="border rounded p-2" />
        <select defaultValue={config?.qos} className="border rounded p-2">
          <option value={0}>QoS 0</option>
          <option value={1}>QoS 1</option>
        </select>
      </div>
      <div className="flex gap-4">
        <button type="button" onClick={onTest} className="px-4 py-2 bg-gray-500 text-white rounded">
          Test Connection
        </button>
        <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded">
          Save
        </button>
      </div>
    </form>
  )
}

function MongoDBConfig({ config, onTest }: any) {
  return (
    <form className="space-y-4">
      <h3 className="text-lg font-medium">MongoDB Configuration</h3>
      <div className="grid grid-cols-1 gap-4">
        <input
          type="text"
          placeholder="URI"
          defaultValue={config?.uri}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="Database"
          defaultValue={config?.database}
          className="border rounded p-2"
        />
        <input
          type="text"
          placeholder="Collection"
          defaultValue={config?.collection}
          className="border rounded p-2"
        />
      </div>
      <div className="flex gap-4">
        <button type="button" onClick={onTest} className="px-4 py-2 bg-gray-500 text-white rounded">
          Test Connection
        </button>
        <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded">
          Save
        </button>
      </div>
    </form>
  )
}

function PollingConfig({ config }: any) {
  return (
    <form className="space-y-4">
      <h3 className="text-lg font-medium">Polling Configuration</h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm text-gray-600 mb-1">Interval (ms)</label>
          <input
            type="number"
            defaultValue={config?.intervalMs}
            className="border rounded p-2 w-full"
          />
        </div>
        <div>
          <label className="block text-sm text-gray-600 mb-1">Batch Size</label>
          <input
            type="number"
            defaultValue={config?.batchSize}
            className="border rounded p-2 w-full"
          />
        </div>
      </div>
      <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded">
        Save
      </button>
    </form>
  )
}
