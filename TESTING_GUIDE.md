# 🧪 Guía de Pruebas - CRUDs API

## Requisitos

- Backend corriendo en `http://localhost:8080`
- Usuario: `admin`, Contraseña: (desde config)
- Herramienta: curl, Postman, o similar

## Ejemplos con curl

### Autenticación Básica

Todos los comandos necesitan autenticación:
```bash
# Reemplaza USERNAME:PASSWORD con tus credenciales
curl -u admin:password http://localhost:8080/api/...
```

---

## 📍 Eventos CRUD

### 1. Listar Eventos (GET /api/events)

```bash
# Obtener página 1 con 50 items
curl -u admin:password "http://localhost:8080/api/events"

# Con paginación específica
curl -u admin:password "http://localhost:8080/api/events?page=2&pageSize=100"

# Filtrar por rango de fechas
curl -u admin:password "http://localhost:8080/api/events?startDate=2024-01-01T00:00:00Z&endDate=2024-01-31T23:59:59Z"

# Filtrar por fuente
curl -u admin:password "http://localhost:8080/api/events?source=Akva"

# Filtrar por unidad
curl -u admin:password "http://localhost:8080/api/events?unitName=Tank-A"

# Combinado: múltiples filtros
curl -u admin:password "http://localhost:8080/api/events?source=Akva&unitName=Tank&page=1&pageSize=50&sortBy=fechaHora&sortOrder=-1"
```

**Respuesta esperada:**
```json
{
  "success": true,
  "data": [
    {
      "_id": "Akva:12345",
      "source": "Akva",
      "fechaHora": "2024-01-12T10:30:00Z",
      "unitName": "Tank-A",
      "payload": { ... },
      "ingestedAt": "2024-01-12T10:30:01Z"
    }
  ],
  "page": 1,
  "pages": 5,
  "total": 234,
  "limit": 50
}
```

### 2. Obtener Evento Individual (GET /api/events/:id)

```bash
# Obtener un evento específico por ID
curl -u admin:password "http://localhost:8080/api/events/Akva:12345"
```

**Respuesta esperada:**
```json
{
  "success": true,
  "data": {
    "_id": "Akva:12345",
    "source": "Akva",
    "fechaHora": "2024-01-12T10:30:00Z",
    "unitName": "Tank-A",
    "payload": {
      "biomasa": 1500.5,
      "pesoProm": 450.0,
      "fishCount": 3200,
      ...
    },
    "ingestedAt": "2024-01-12T10:30:01Z"
  }
}
```

### 3. Actualizar Evento (PUT /api/events/:id)

```bash
# Actualizar específicos campos de un evento
curl -X PUT -u admin:password \
  -H "Content-Type: application/json" \
  -d '{
    "payload": {
      "biomasa": 1600.0,
      "pesoProm": 475.5
    }
  }' \
  "http://localhost:8080/api/events/Akva:12345"
```

**Respuesta:** El evento actualizado

### 4. Eliminar Evento (DELETE /api/events/:id)

```bash
# Eliminar un evento específico
curl -X DELETE -u admin:password "http://localhost:8080/api/events/Akva:12345"
```

**Respuesta esperada:**
```json
{
  "success": true,
  "data": {
    "message": "Event deleted successfully"
  }
}
```

### 5. Batch Delete (DELETE /api/events/batch)

```bash
# Eliminar todos los eventos de una fuente anteriores a cierta fecha
curl -X DELETE -u admin:password \
  -H "Content-Type: application/json" \
  -d '{
    "source": "Akva",
    "beforeDate": "2024-01-01T00:00:00Z"
  }' \
  "http://localhost:8080/api/events/batch"
```

**Respuesta esperada:**
```json
{
  "success": true,
  "data": {
    "message": "Batch delete completed",
    "deleted": 450
  }
}
```

---

## ⚙️ Configuración CRUD

### 1. Obtener Configuración (GET /api/config)

```bash
curl -u admin:password "http://localhost:8080/api/config"
```

**Respuesta:** Config actual (contraseñas ocultas como "********")

### 2. Actualizar Configuración (PUT /api/config)

```bash
# Actualizar configuración
curl -X PUT -u admin:password \
  -H "Content-Type: application/json" \
  -d '{
    "polling": {
      "intervalMs": 10000,
      "batchSize": 200
    },
    "mqtt": {
      "qos": 2
    }
  }' \
  "http://localhost:8080/api/config"
```

**Nota:** Deja contraseñas como "********" para preservar valores actuales

---

## 📋 Logs CRUD

### 1. Obtener Logs (GET /api/logs)

```bash
# Obtener todos los logs
curl -u admin:password "http://localhost:8080/api/logs"

# Con paginación
curl -u admin:password "http://localhost:8080/api/logs?page=1&pageSize=100"

# Filtrar por nivel
curl -u admin:password "http://localhost:8080/api/logs?level=ERROR"

# Combinado
curl -u admin:password "http://localhost:8080/api/logs?level=WARN&page=1&pageSize=50"
```

**Respuesta esperada:**
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
      "message": "Failed to connect to MQTT"
    }
  ],
  "page": 1,
  "pages": 3,
  "total": 250,
  "limit": 100
}
```

---

## 🧬 Scripting en PowerShell

```powershell
# Función auxiliar para curl con auth
function Call-OmnipollAPI {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body,
        [string]$Username = "admin",
        [string]$Password = "password"
    )
    
    $auth = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes("$($Username):$($Password)"))
    $headers = @{
        "Authorization" = "Basic $auth"
        "Content-Type" = "application/json"
    }
    
    $url = "http://localhost:8080$Endpoint"
    
    if ($Body) {
        Invoke-RestMethod -Uri $url -Method $Method -Headers $headers -Body ($Body | ConvertTo-Json)
    } else {
        Invoke-RestMethod -Uri $url -Method $Method -Headers $headers
    }
}

# Uso:
# Call-OmnipollAPI -Endpoint "/api/events?page=1"
# Call-OmnipollAPI -Method "DELETE" -Endpoint "/api/events/Akva:12345"
```

---

## ✔️ Checklist de Pruebas

- [ ] GET /api/events - Listar eventos
- [ ] GET /api/events/:id - Obtener evento por ID (requiere ID válido)
- [ ] PUT /api/events/:id - Actualizar evento
- [ ] DELETE /api/events/:id - Eliminar evento
- [ ] DELETE /api/events/batch - Batch delete
- [ ] GET /api/config - Obtener configuración
- [ ] PUT /api/config - Actualizar configuración
- [ ] GET /api/logs - Obtener logs
- [ ] GET /api/logs?level=ERROR - Filtrar logs por nivel

---

## 🐛 Troubleshooting

### Error 401 (Unauthorized)
- Verifica credenciales de usuario
- Asegúrate de usar `curl -u username:password`

### Error 404 (Not Found)
- Verifica que la URL sea correcta
- Comprueba el ID del evento si es una operación específica

### Error 500 (Internal Server Error)
- Revisa los logs del servidor
- Verifica que MongoDB/SQL Server estén conectados

### Respuesta vacía o timeout
- Asegúrate de que el servidor esté corriendo
- Verifica la URL base (puerto 8080)

---

## 📊 Ejemplo de Flujo Completo

```bash
#!/bin/bash
USER="admin"
PASS="password"
BASE_URL="http://localhost:8080"

# 1. Listar eventos
echo "=== Listando eventos ==="
curl -s -u $USER:$PASS "$BASE_URL/api/events?pageSize=10" | jq .

# 2. Obtener primer evento
echo "=== Obteniendo primer evento ==="
EVENT_ID=$(curl -s -u $USER:$PASS "$BASE_URL/api/events?pageSize=1" | jq -r '.data[0]._id')
curl -s -u $USER:$PASS "$BASE_URL/api/events/$EVENT_ID" | jq .

# 3. Actualizar el evento
echo "=== Actualizando evento ==="
curl -s -X PUT -u $USER:$PASS \
  -H "Content-Type: application/json" \
  -d '{"payload": {"testField": "testValue"}}' \
  "$BASE_URL/api/events/$EVENT_ID" | jq .

# 4. Obtener logs de error
echo "=== Obteniendo logs de error ==="
curl -s -u $USER:$PASS "$BASE_URL/api/logs?level=ERROR&pageSize=5" | jq .

echo "=== Prueba completada ==="
```

