import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 10000, // 10 second timeout
  auth: {
    username: 'admin',
    password: 'admin',
  },
})

export const api = {
  // Status
  getStatus: async () => {
    const { data } = await client.get('/status')
    return data
  },

  // Worker controls
  startWorker: async () => {
    const { data } = await client.post('/worker/start')
    return data
  },

  stopWorker: async () => {
    const { data } = await client.post('/worker/stop')
    return data
  },

  resetWatermark: async () => {
    const { data } = await client.post('/watermark/reset')
    return data
  },

  // Configuration
  getConfig: async () => {
    const { data } = await client.get('/config')
    return data
  },

  saveConfig: async (config: any) => {
    const { data } = await client.put('/config', config)
    return data
  },

  testConnection: async (type: 'sqlserver' | 'mqtt' | 'mongodb') => {
    const { data } = await client.post(`/test/${type}`)
    return data
  },

  // Logs
  getLogs: async (level?: string, page?: number, pageSize?: number) => {
    const { data } = await client.get('/logs', {
      params: { level, page, pageSize },
    })
    return data
  },

  // Events
  getEvents: async (page?: number, pageSize?: number, filters?: any) => {
    const { data } = await client.get('/events', {
      params: { page, pageSize, ...filters },
    })
    return data
  },

  getEventById: async (id: string) => {
    const { data } = await client.get(`/events/${id}`)
    return data
  },

  updateEvent: async (id: string, payload: any) => {
    const { data } = await client.put(`/events/${id}`, payload)
    return data
  },

  deleteEvent: async (id: string) => {
    const { data } = await client.delete(`/events/${id}`)
    return data
  },

  deleteEventsBatch: async (source?: string, beforeDate?: string) => {
    const { data } = await client.delete('/events/batch', {
      data: { source, beforeDate },
    })
    return data
  },
}
