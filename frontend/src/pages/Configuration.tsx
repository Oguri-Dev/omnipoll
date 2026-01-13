import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../services/api'
import { Settings, CheckCircle, AlertCircle } from 'lucide-react'

interface ConfigData {
  sqlServer?: Record<string, any>
  mqtt?: Record<string, any>
  mongodb?: Record<string, any>
  polling?: Record<string, any>
  admin?: Record<string, any>
}

export default function Configuration() {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'sqlserver' | 'mqtt' | 'mongodb' | 'polling'>('sqlserver')
  const [testResult, setTestResult] = useState<{ type?: string; success?: boolean; message?: string } | null>(null)
  const [successMessage, setSuccessMessage] = useState('')

  const { data: response, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: api.getConfig,
  })

  const config = response?.data || {}

  const saveConfig = useMutation({
    mutationFn: api.saveConfig,
    onSuccess: () => {
      setSuccessMessage('Configuration saved successfully!')
      setTimeout(() => setSuccessMessage(''), 3000)
      queryClient.invalidateQueries({ queryKey: ['config'] })
    },
    onError: (error: any) => {
      alert('Error saving configuration: ' + (error.response?.data?.error || error.message))
    },
  })

  const testConnection = useMutation({
    mutationFn: (type: 'sqlserver' | 'mqtt' | 'mongodb') => api.testConnection(type),
    onSuccess: (data, type) => {
      setTestResult({
        type,
        success: data.connected,
        message: data.error || 'Connection successful',
      })
      setTimeout(() => setTestResult(null), 3000)
    },
    onError: (error: any, type) => {
      setTestResult({
        type,
        success: false,
        message: error.response?.data?.error || error.message,
      })
      setTimeout(() => setTestResult(null), 3000)
    },
  })

  if (isLoading) {
    return <div className="p-4">Loading configuration...</div>
  }

  const tabs = [
    { id: 'sqlserver' as const, label: 'SQL Server' },
    { id: 'mqtt' as const, label: 'MQTT' },
    { id: 'mongodb' as const, label: 'MongoDB' },
    { id: 'polling' as const, label: 'Polling' },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Settings className="w-6 h-6" />
        <h1 className="text-2xl font-bold">Configuration</h1>
      </div>

      {/* Success message */}
      {successMessage && (
        <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded flex items-center gap-2">
          <CheckCircle className="w-5 h-5" />
          {successMessage}
        </div>
      )}

      {/* Test result */}
      {testResult && (
        <div className={`border px-4 py-3 rounded flex items-center gap-2 ${testResult.success ? 'bg-green-100 border-green-400 text-green-700' : 'bg-red-100 border-red-400 text-red-700'}`}>
          {testResult.success ? <CheckCircle className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
          <span>{testResult.type} - {testResult.message}</span>
        </div>
      )}

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
          <ConfigForm
            title="SQL Server Configuration"
            config={config.sqlServer || {}}
            fields={[
              { key: 'host', label: 'Host', type: 'text' },
              { key: 'port', label: 'Port', type: 'number' },
              { key: 'database', label: 'Database', type: 'text' },
              { key: 'user', label: 'User', type: 'text' },
              { key: 'password', label: 'Password', type: 'password' },
            ]}
            onSave={(data) => {
              saveConfig.mutate({ ...config, sqlServer: data })
            }}
            onTest={() => testConnection.mutate('sqlserver')}
            isTesting={testConnection.isPending}
            isSaving={saveConfig.isPending}
          />
        )}

        {activeTab === 'mqtt' && (
          <ConfigForm
            title="MQTT Configuration"
            config={config.mqtt || {}}
            fields={[
              { key: 'broker', label: 'Broker', type: 'text' },
              { key: 'port', label: 'Port', type: 'number' },
              { key: 'topic', label: 'Topic', type: 'text' },
              { key: 'clientId', label: 'Client ID', type: 'text' },
              { key: 'user', label: 'User', type: 'text' },
              { key: 'password', label: 'Password', type: 'password' },
              { key: 'qos', label: 'QoS', type: 'select', options: [0, 1, 2] },
            ]}
            onSave={(data) => {
              saveConfig.mutate({ ...config, mqtt: data })
            }}
            onTest={() => testConnection.mutate('mqtt')}
            isTesting={testConnection.isPending}
            isSaving={saveConfig.isPending}
          />
        )}

        {activeTab === 'mongodb' && (
          <ConfigForm
            title="MongoDB Configuration"
            config={config.mongodb || {}}
            fields={[
              { key: 'uri', label: 'URI', type: 'text' },
              { key: 'database', label: 'Database', type: 'text' },
              { key: 'collection', label: 'Collection', type: 'text' },
            ]}
            onSave={(data) => {
              saveConfig.mutate({ ...config, mongodb: data })
            }}
            onTest={() => testConnection.mutate('mongodb')}
            isTesting={testConnection.isPending}
            isSaving={saveConfig.isPending}
          />
        )}

        {activeTab === 'polling' && (
          <ConfigForm
            title="Polling Configuration"
            config={config.polling || {}}
            fields={[
              { key: 'intervalMs', label: 'Interval (ms)', type: 'number' },
              { key: 'batchSize', label: 'Batch Size', type: 'number' },
            ]}
            onSave={(data) => {
              saveConfig.mutate({ ...config, polling: data })
            }}
            isSaving={saveConfig.isPending}
            showTest={false}
          />
        )}
      </div>
    </div>
  )
}

function ConfigForm({
  title,
  config,
  fields,
  onSave,
  onTest,
  isTesting,
  isSaving,
  showTest = true,
}: {
  title: string
  config: Record<string, any>
  fields: Array<{
    key: string
    label: string
    type: string
    options?: any[]
  }>
  onSave: (data: Record<string, any>) => void
  onTest?: () => void
  isTesting?: boolean
  isSaving?: boolean
  showTest?: boolean
}) {
  const [formData, setFormData] = useState(config)

  const handleChange = (key: string, value: any) => {
    setFormData(prev => ({ ...prev, [key]: value }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSave(formData)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h3 className="text-lg font-medium">{title}</h3>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {fields.map((field) => (
          <div key={field.key}>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {field.label}
            </label>
            {field.type === 'select' ? (
              <select
                value={formData[field.key] || ''}
                onChange={(e) => handleChange(field.key, Number(e.target.value))}
                className="w-full border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select...</option>
                {field.options?.map((opt) => (
                  <option key={opt} value={opt}>
                    {opt}
                  </option>
                ))}
              </select>
            ) : (
              <input
                type={field.type}
                value={formData[field.key] || ''}
                onChange={(e) => handleChange(field.key, field.type === 'number' ? Number(e.target.value) : e.target.value)}
                placeholder={field.label}
                className="w-full border rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            )}
          </div>
        ))}
      </div>

      <div className="flex gap-4 pt-4">
        {showTest && onTest && (
          <button
            type="button"
            onClick={onTest}
            disabled={isTesting}
            className="px-4 py-2 bg-gray-500 hover:bg-gray-600 text-white rounded disabled:opacity-50"
          >
            {isTesting ? 'Testing...' : 'Test Connection'}
          </button>
        )}
        <button
          type="submit"
          disabled={isSaving}
          className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded disabled:opacity-50"
        >
          {isSaving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>
    </form>
  )
}
