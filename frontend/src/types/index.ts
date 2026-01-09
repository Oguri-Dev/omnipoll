// Status types
export interface Status {
  workerRunning: boolean
  lastFechaHora: string
  eventsToday: number
  ingestionRate: number
  totalEvents: number
  connections: {
    sqlServer: boolean
    mqtt: boolean
    mongodb: boolean
  }
}

// Configuration types
export interface Config {
  sqlServer: SQLServerConfig
  mqtt: MQTTConfig
  mongodb: MongoDBConfig
  polling: PollingConfig
}

export interface SQLServerConfig {
  host: string
  port: number
  database: string
  user: string
  password: string
}

export interface MQTTConfig {
  broker: string
  port: number
  topic: string
  clientId: string
  user: string
  password: string
  qos: 0 | 1
}

export interface MongoDBConfig {
  uri: string
  database: string
  collection: string
}

export interface PollingConfig {
  intervalMs: number
  batchSize: number
}

// Log types
export interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error'
  message: string
}
