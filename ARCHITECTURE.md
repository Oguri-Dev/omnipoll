# Omnipoll - Arquitectura de APIs CRUD Implementadas

## Diagrama de Flujo de APIs

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENTE (Frontend/REST)                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      ADMIN HTTP SERVER (8080)                    │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              HTTP MIDDLEWARE STACK                       │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  1. Authentication (withAuth)       [HTTP Basic Auth]   │  │
│  │  2. CORS (withCORS)                  [Allow All]         │  │
│  │  3. Logging (withLogging)           [Request/Response]  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              ↓                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              CUSTOM ROUTER (router.go)                  │  │
│  │  - /api/events         → handleEventsRoute              │  │
│  │  - /api/events/:id     → handleEventByID                │  │
│  │  - /api/events/batch   → handleEventsBatch             │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              HTTP MUXER (http.ServeMux)                │  │
│  │  - /api/status         → handleStatus                  │  │
│  │  - /api/config         → handleConfig                  │  │
│  │  - /api/logs           → handleLogsImproved            │  │
│  │  - /api/worker/*       → handleWorker*                 │  │
│  │  - /api/test/*         → handleTest*                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    HANDLERS (event_handlers.go)                  │
│                                                                   │
│  ├─ handleEventsGet()     [GET /api/events]                    │
│  │  └─ QueryEvents() → Poller → Worker                         │
│  │                                                              │
│  ├─ handleEventGetByID()  [GET /api/events/:id]                │
│  │  └─ GetEventByID() → Poller → Worker                        │
│  │                                                              │
│  ├─ handleEventUpdate()   [PUT /api/events/:id]                │
│  │  └─ UpdateEvent() → Poller → Worker                         │
│  │                                                              │
│  ├─ handleEventDelete()   [DELETE /api/events/:id]             │
│  │  └─ DeleteEvent() → Poller → Worker                         │
│  │                                                              │
│  └─ handleEventsBatch()   [DELETE /api/events/batch]           │
│     └─ DeleteEventsBatch() → Poller → Worker                   │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                  WORKER (poller/worker.go)                      │
│                                                                   │
│  Orquestador Central                                            │
│  ├─ QueryEvents()         ──→ MongoRepo.QueryEvents()          │
│  ├─ GetEventByID()        ──→ MongoRepo.GetByID()              │
│  ├─ UpdateEvent()         ──→ MongoRepo.UpdateByID()           │
│  ├─ DeleteEvent()         ──→ MongoRepo.DeleteByID()           │
│  └─ DeleteEventsBatch()   ──→ MongoRepo.DeleteByFilter()       │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│             REPOSITORY PATTERN (mongo/repository.go)             │
│                                                                   │
│  Data Access Layer                                              │
│  ├─ QueryEvents(opts)       [Con paginación/filtros]           │
│  ├─ GetByID(id)             [Obtener por ID]                   │
│  ├─ UpdateByID(id, data)    [Actualizar documento]             │
│  ├─ DeleteByID(id)          [Eliminar documento]               │
│  └─ DeleteByFilter(src, dt) [Eliminar múltiples]              │
│                                                                   │
│  Métodos de Soporte                                             │
│  ├─ Insert()                [Insertar evento]                  │
│  ├─ InsertBatch()           [Insertar múltiples]               │
│  ├─ GetRecentEvents()       [Obtener recientes]                │
│  ├─ CountEvents()           [Contar total]                     │
│  └─ IsConnected()           [Estado de conexión]               │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    MONGODB COLLECTION                            │
│                                                                   │
│  historical_events                                              │
│  ├─ _id (índice primario)                                      │
│  ├─ source                                                     │
│  ├─ fechaHora                                                  │
│  ├─ unitName                                                   │
│  ├─ payload (documento flexible)                               │
│  └─ ingestedAt                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Estructura de Datos

### MongoDB Document (HistoricalEvent)
```json
{
  "_id": "Akva:12345",
  "source": "Akva",
  "fechaHora": "2024-01-12T10:30:00Z",
  "unitName": "Tank-A",
  "payload": {
    "name": "Feeding-001",
    "dia": "2024-01-12",
    "inicio": "10:30:00",
    "fin": "10:35:00",
    "dif": 300,
    "amountGrams": 5000.0,
    "pelletFishMin": 150.5,
    "fishCount": 3200.0,
    "pesoProm": 450.0,
    "biomasa": 1440000.0,
    "pelletPK": 2.5,
    "feedName": "Premium Feed",
    "siloName": "Silo-1",
    "doserName": "Doser-A",
    "gramsPerSec": 150.0,
    "kgTonMin": 9.0,
    "marca": 1
  },
  "ingestedAt": "2024-01-12T10:30:01Z"
}
```

### QueryOptions (Filtrado y Paginación)
```go
type QueryOptions struct {
  Page      int        // 1-based page number
  PageSize  int        // Items per page (max 500)
  StartDate *time.Time // Filter: fechaHora >= startDate
  EndDate   *time.Time // Filter: fechaHora <= endDate
  Source    string     // Filter: exact match on source
  UnitName  string     // Filter: case-insensitive regex
  SortBy    string     // "fechaHora" or "ingestedAt"
  SortOrder int        // 1 ascending, -1 descending
}
```

### API Response Format
```json
{
  "success": true,
  "data": [ /* array de resultados */ ],
  "page": 1,
  "pages": 10,
  "total": 500,
  "limit": 50
}
```

## Flujos de Casos de Uso

### 1. Listar Eventos con Filtros
```
Cliente: GET /api/events?source=Akva&unitName=Tank&page=2&pageSize=50
   ↓
middleware: Auth + CORS + Logging
   ↓
handleEventsGet(): Extrae parámetros de query
   ↓
Worker.QueryEvents(QueryOptions)
   ↓
MongoRepo.QueryEvents()
   ↓
MongoDB.Find() con filtros y paginación
   ↓
Resultado: QueryResult { Data[], Page, Total, TotalPages }
   ↓
Cliente: JSON con datos paginados
```

### 2. Obtener Evento Individual
```
Cliente: GET /api/events/Akva:12345
   ↓
middleware: Auth + CORS + Logging
   ↓
handleEventGetByID(id)
   ↓
Worker.GetEventByID(id)
   ↓
MongoRepo.GetByID()
   ↓
MongoDB.FindOne() por _id
   ↓
Cliente: JSON con evento o 404 si no existe
```

### 3. Actualizar Evento
```
Cliente: PUT /api/events/Akva:12345
Body: { payload: { biomasa: 1600.0 } }
   ↓
middleware: Auth + CORS + Logging
   ↓
handleEventUpdate(id, data)
   ↓
Worker.UpdateEvent(id, data)
   ↓
MongoRepo.UpdateByID()
   ↓
MongoDB.FindOneAndUpdate()
   ↓
Cliente: JSON con evento actualizado
```

### 4. Eliminar Evento
```
Cliente: DELETE /api/events/Akva:12345
   ↓
middleware: Auth + CORS + Logging
   ↓
handleEventDelete(id)
   ↓
Worker.DeleteEvent(id)
   ↓
MongoRepo.DeleteByID()
   ↓
MongoDB.DeleteOne()
   ↓
Cliente: { success: true, message: "..." }
```

### 5. Batch Delete
```
Cliente: DELETE /api/events/batch
Body: { source: "Akva", beforeDate: "2024-01-01T00:00:00Z" }
   ↓
middleware: Auth + CORS + Logging
   ↓
handleEventsBatch(source, beforeDate)
   ↓
Worker.DeleteEventsBatch()
   ↓
MongoRepo.DeleteByFilter()
   ↓
MongoDB.DeleteMany() con filtro
   ↓
Cliente: { success: true, deleted: 450 }
```

## Tabla de Endpoints

| Método | Endpoint | Handler | CRUD | Descripción |
|--------|----------|---------|------|-------------|
| GET | /api/events | handleEventsGet | READ | Listar con paginación/filtros |
| GET | /api/events/:id | handleEventGetByID | READ | Obtener uno por ID |
| PUT | /api/events/:id | handleEventUpdate | UPDATE | Actualizar evento |
| DELETE | /api/events/:id | handleEventDelete | DELETE | Eliminar evento |
| DELETE | /api/events/batch | handleEventsBatch | DELETE | Batch delete |
| GET | /api/config | handleConfig | READ | Obtener configuración |
| PUT | /api/config | handleConfig | UPDATE | Actualizar configuración |
| GET | /api/logs | handleLogsImproved | READ | Listar logs con filtros |
| GET | /api/status | handleStatus | READ | Estado del sistema |

## Índices de MongoDB Recomendados

```javascript
// Crear índices para optimizar queries
db.historical_events.createIndex({ "fechaHora": -1 })
db.historical_events.createIndex({ "source": 1 })
db.historical_events.createIndex({ "unitName": 1 })
db.historical_events.createIndex({ "ingestedAt": -1 })
db.historical_events.createIndex({ "source": 1, "fechaHora": -1 })
```

## Seguridad

- ✅ HTTP Basic Authentication en todos los endpoints
- ✅ Contraseñas ocultas en respuestas
- ✅ CORS configurable
- ✅ Validación de entrada (paginación max, etc)
- ⏳ Rate limiting (TODO)
- ⏳ Audit logging (TODO)

## Performance

- ✅ Paginación (no cargar todo en memoria)
- ✅ Filtrado en base de datos (no en aplicación)
- ✅ Límite de page size (máximo 500)
- ✅ Índices recomendados para queries comunes
- ⏳ Caching (TODO)
- ⏳ Compression (TODO)

