import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 10000, // 10 second timeout
  auth: {
    username: 'admin',
    password: 'admin123',
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
  getLogs: async () => {
    const { data } = await client.get('/logs')
    return data
  },

  // Events
  getEvents: async () => {
    const { data } = await client.get('/events')
    return data
  },
}
