# CRUD Implementation Summary

## Eventos CRUD (Completed ✅)

### Endpoints

#### GET /api/events
Retrieves a paginated list of events with filtering capabilities.

**Query Parameters:**
- `page` (int, default=1): Page number
- `pageSize` or `limit` (int, default=50, max=500): Items per page
- `startDate` (RFC3339): Filter events after this date
- `endDate` (RFC3339): Filter events before this date
- `source` (string): Filter by source (e.g., "Akva")
- `unitName` (string): Filter by unit name (case-insensitive)
- `sortBy` (string, default="ingestedAt"): Sort field
- `sortOrder` (int, default=-1): 1=ascending, -1=descending

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "_id": "Akva:123",
      "source": "Akva",
      "fechaHora": "2024-01-12T10:30:00Z",
      "unitName": "Tank-A",
      "payload": { ... },
      "ingestedAt": "2024-01-12T10:30:01Z"
    }
  ],
  "page": 1,
  "pages": 10,
  "total": 500,
  "limit": 50
}
```

#### GET /api/events/:id
Retrieves a single event by ID.

**Response:**
```json
{
  "success": true,
  "data": {
    "_id": "Akva:123",
    "source": "Akva",
    "fechaHora": "2024-01-12T10:30:00Z",
    "unitName": "Tank-A",
    "payload": { ... },
    "ingestedAt": "2024-01-12T10:30:01Z"
  }
}
```

#### PUT /api/events/:id
Updates a single event.

**Request Body:**
```json
{
  "payload": {
    "biomasa": 1500.5,
    "pesoProm": 450.0
  }
}
```

**Response:** Returns the updated event object

#### DELETE /api/events/:id
Deletes a single event.

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Event deleted successfully"
  }
}
```

#### DELETE /api/events/batch
Batch delete events matching criteria.

**Request Body:**
```json
{
  "source": "Akva",
  "beforeDate": "2024-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Batch delete completed",
    "deleted": 150
  }
}
```

## Configuración CRUD (Completed ✅)

### Endpoints

#### GET /api/config
Retrieves current configuration (passwords masked).

**Response:**
```json
{
  "success": true,
  "data": {
    "sqlServer": {
      "host": "localhost",
      "port": 1433,
      "database": "FTFeeding",
      "user": "sa",
      "password": "********"
    },
    "mqtt": { ... },
    "mongodb": { ... },
    "polling": { ... },
    "admin": { ... }
  }
}
```

#### PUT /api/config
Updates configuration. Passwords with value "********" are preserved from current config.

**Request Body:** Same structure as GET response

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

## Logs CRUD (Completed ✅)

### Endpoints

#### GET /api/logs
Retrieves logs with filtering and pagination.

**Query Parameters:**
- `page` (int, default=1): Page number
- `pageSize` or `limit` (int, default=100): Items per page
- `level` (string): Filter by log level (INFO, WARN, ERROR, DEBUG)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "timestamp": "2024-01-12T10:30:00Z",
      "level": "INFO",
      "message": "Connected to SQL Server"
    },
    {
      "timestamp": "2024-01-12T10:29:59Z",
      "level": "WARN",
      "message": "Failed to connect to MQTT: connection refused"
    }
  ],
  "page": 1,
  "pages": 5,
  "total": 450,
  "limit": 100
}
```

## Backend Changes

### New Files
- `internal/admin/responses.go` - Standard API response helpers
- `internal/admin/event_handlers.go` - Event CRUD handlers
- `internal/admin/logs_handlers.go` - Improved logs handler
- `internal/admin/router.go` - Custom router for ID-based routes

### Modified Files
- `internal/mongo/repository.go` - Added CRUD methods:
  - `GetByID()` - Fetch single event
  - `QueryEvents()` - Fetch with filtering/pagination
  - `UpdateByID()` - Update event
  - `DeleteByID()` - Delete single event
  - `DeleteByFilter()` - Batch delete

- `internal/poller/worker.go` - Added methods to expose CRUD operations:
  - `QueryEvents()`
  - `GetEventByID()`
  - `UpdateEvent()`
  - `DeleteEvent()`
  - `DeleteEventsBatch()`

- `internal/admin/server.go` - Updated routing to support event IDs
- `internal/admin/handlers.go` - Updated event handler routing

## API Response Format

All endpoints use consistent response format:

**Success Response:**
```json
{
  "success": true,
  "data": { ... },
  "page": 1,
  "pages": 10,
  "total": 500,
  "limit": 50
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Error message here"
}
```

## Authentication

All endpoints require HTTP Basic Authentication with:
- Username: `admin` (configurable)
- Password: From config (encrypted)

## Pagination

Supported across Events and Logs endpoints:
- Default page size: 50 (events), 100 (logs)
- Maximum page size: 500
- Returns total count and total pages

## Filtering

### Events
- **Date Range**: `startDate` and `endDate` (RFC3339 format)
- **Source**: Filter by data source
- **Unit Name**: Case-insensitive search
- **Sort**: By `fechaHora` or `ingestedAt`, ascending/descending

### Logs
- **Level**: Filter by log level (INFO, WARN, ERROR, DEBUG)
- **Pagination**: Page and pageSize

## Next Steps

1. **Frontend Integration** - Connect React pages to these endpoints
2. **Testing** - Add unit and integration tests
3. **Validation** - Add input validation for update operations
4. **Transactions** - Consider MongoDB transactions for batch operations
5. **Audit Logging** - Log who modified what and when
6. **Soft Deletes** - Consider soft delete for audit trail
