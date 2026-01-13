# Verificar Estado Real de Conexión MQTT

## Problema Original
Aunque el log mostraba `[info] Connected to MQTT broker`, el dashboard siempre mostraba MQTT como desconectado (X roja).

## Causa
El endpoint `/api/status` solo marcaba las conexiones como `true` cuando se ejecutaba un poll exitoso. Como no había datos nuevos en SQL Server, nunca se ejecutaba el poll correctamente, así que nunca se marcaba como conectado.

## Solución Implementada

### 1. Método `UpdateConnectionStats()`
Se agregó un nuevo método en `poller.go` que verifica el estado **real** de cada conexión:

```go
func (p *Poller) UpdateConnectionStats() {
    p.statsMu.Lock()
    defer p.statsMu.Unlock()

    if p.mqttPub != nil {
        p.stats.MQTTConnected = p.mqttPub.IsConnected()  // Verifica estado real
    } else {
        p.stats.MQTTConnected = false
    }

    if p.akvaClient != nil {
        p.stats.SQLConnected = p.akvaClient.IsConnected()
    } else {
        p.stats.SQLConnected = false
    }

    if p.mongoRepo != nil {
        p.stats.MongoConnected = p.mongoRepo.IsConnected()
    } else {
        p.stats.MongoConnected = false
    }
}
```

### 2. Llamada en `Poll()`
Cada vez que se intenta hacer un poll, primero se actualiza el estado de las conexiones:

```go
func (p *Poller) Poll(ctx context.Context) error {
    // Actualizar estado real de conexiones ANTES de intentar operaciones
    p.UpdateConnectionStats()
    
    // Luego validar que los clientes existen
    if p.akvaClient == nil {
        return fmt.Errorf("not connected to SQL Server (Akva)")
    }
    // ...
}
```

### 3. Métodos de Verificación en Clientes

#### MQTT (mqtt/client.go)
```go
func (c *Client) IsConnected() bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if c.client == nil {
        return false
    }
    return c.client.IsConnected()  // Paho MQTT verifica estado real
}
```

#### SQL Server (akva/client.go)
```go
func (c *Client) IsConnected() bool {
    // Verifica si la conexión a la BD está activa
}
```

#### MongoDB (mongo/repository.go)
```go
func (r *Repository) IsConnected() bool {
    return r.client.IsConnected()
}
```

## Cómo Verificar

### Opción 1: Dashboard
1. Abre http://localhost:3001
2. Verifica la sección "Connections" en el panel principal
3. MQTT debería mostrar ✓ si está realmente conectado

### Opción 2: API Directa
```bash
curl http://localhost:8080/api/status \
  -H "Authorization: Basic YWRtaW46YWRtaW4="  # admin:admin en base64
```

Respuesta:
```json
{
  "workerRunning": true,
  "lastFechaHora": "2026-01-12T23:38:40Z",
  "eventsToday": 0,
  "ingestionRate": 0,
  "totalEvents": 0,
  "connections": {
    "sqlServer": false,      // ❌ No conectado (localhost:1433 no disponible)
    "mqtt": true,            // ✅ Conectado (mqtt.vmsfish.com:8883)
    "mongodb": false          // ❌ No conectado (localhost:27017 no disponible)
  },
  "uptimeSeconds": 45
}
```

### Opción 3: Logs
Observa los logs cuando el worker intenta hacer polls:

```
2026/01/12 23:38:40 worker.go:346: [info] Connected to MQTT broker
2026/01/12 23:38:45 worker.go:346: [error] Poll error: not connected
2026/01/12 23:38:50 worker.go:346: [error] Poll error: not connected
```

El `[info] Connected to MQTT broker` confirma que MQTT se conectó exitosamente.

## Estado en Tu Caso

Tu servidor está en este estado:
- ✅ **MQTT**: Conectado a `mqtt.vmsfish.com:8883`
- ❌ **SQL Server**: No disponible (localhost:1433 está rechazando conexiones)
- ❌ **MongoDB**: No disponible (localhost:27017 está rechazando conexiones)
- ❌ **Worker**: No puede hacer polls (falta SQL Server y MongoDB)

Esto es normal si no tienes Docker corriendo con esos servicios.

## Para Tener Todo Conectado

Si quieres probar con servicios reales:

```bash
# Desde la carpeta raíz del proyecto
docker-compose up -d

# Esto levanta:
# - SQL Server (puerto 1433)
# - MongoDB (puerto 27017)
# - MQTT (ya está en la nube)
```

## Cambios Realizados

- **Archivo**: `backend/internal/poller/poller.go`
- **Commit**: `9ab88a7`
- **Métodos nuevos**:
  - `UpdateConnectionStats()` - Verifica estado real de conexiones
  - Mejora del método `Poll()` - Llama UpdateConnectionStats antes de intentar operaciones

## Próximas Mejoras

1. Agregar endpoint `/api/test/connections` para forzar verificación inmediata
2. Agregar logs más detallados de por qué cada conexión falla
3. Implementar reconexión automática si se detecta una desconexión
4. Agregar webhooks para alertas de cambios de estado
