# Guía de Testing: Verificación de JSON en MQTT

## 1. Requisitos Previos

### ✅ Ya Tienes
- Backend compilado y corriendo en `localhost:8080`
- Frontend en `http://localhost:3001`
- MQTT conectado a `mqtt.vmsfish.com:8883`
- Configuración valida en `config.yaml`

### ❌ Necesitas Para Testing Completo
- SQL Server con datos (Docker o local)
- MongoDB para deduplicación (Docker o local)

---

## 2. Testing sin SQL Server (Verificar Estructura de Código)

### Test 2.1: Verificar que los métodos existen

```bash
# Terminal backend
cd f:\vscode\omnipoll\backend

# Buscar métodos de publicación
grep -n "PublishBatch\|Publish\|json.Marshal" internal/mqtt/publisher.go
grep -n "buildDynamicTopic\|cleanJaula" internal/mqtt/publisher.go
grep -n "ToNormalizedEvent\|ToNormalizedEvents" internal/akva/mapper.go
```

**Resultado esperado:**
```
publisher.go:97: func (p *Publisher) PublishBatch(evts []events.NormalizedEvent) error
publisher.go:44: func (p *Publisher) buildDynamicTopic(centerName string) string
publisher.go:51: func (p *Publisher) cleanJaula(unitName string) string
publisher.go:83: payload, err := json.Marshal(msg)
mapper.go:20: func ToNormalizedEvent(record DetalleAlimentacion) events.NormalizedEvent
mapper.go:50: func ToNormalizedEvents(records []DetalleAlimentacion) []events.NormalizedEvent
```

### Test 2.2: Verificar estructura JSON compilada

```bash
cd f:\vscode\omnipoll\backend

# Ver campos del MQTTMessage
grep -A 12 "type MQTTMessage struct" internal/mqtt/publisher.go
```

**Resultado esperado:**
```go
type MQTTMessage struct {
	TimeStampAkva       string  `json:"TimeStampAkva"`
	TimeStampIngresado  string  `json:"TimeStampIngresado"`
	Jaula               string  `json:"Jaula"`
	Centro              string  `json:"Centro"`
	Gramos              float64 `json:"Gramos"`
	Biomasa             float64 `json:"Biomasa"`
	Peces               float64 `json:"Peces"`
	PesoPromedio        float64 `json:"PesoPromedio"`
	Alimento            string  `json:"Alimento"`
	Silo                string  `json:"Silo"`
}
```

---

## 3. Testing con Datos de Prueba (Requiere SQL Server)

### Step 3.1: Levantar servicios con Docker

```bash
cd f:\vscode\omnipoll

# Iniciar SQL Server y MongoDB
docker-compose up -d mssql mongodb

# Verificar que están corriendo
docker ps
```

**Esperado:**
```
CONTAINER ID   IMAGE                      STATUS
abc123...      mcr.microsoft.com/mssql... Up 2 minutes
def456...      mongo:latest               Up 2 minutes
```

### Step 3.2: Esperar a que SQL Server esté listo

```bash
# El primer arranque toma 2-3 minutos
sleep 180

# Verificar que está listo
docker logs mssql | grep "Recovery completed"
```

### Step 3.3: Insertar datos de prueba

```bash
# Conectar a SQL Server (usar Azure Data Studio o sqlcmd)
sqlcmd -S localhost,1433 -U sa -P AdminPassword123!

# O con herramienta GUI:
# Host: localhost
# Port: 1433
# User: sa
# Password: AdminPassword123!
# Database: FTFeeding
```

**Ejecutar script:**
```sql
USE FTFeeding

-- Ver tabla existente
SELECT TOP 5 * FROM [dbo].[DetalleAlimentacion]

-- Insertar datos de prueba
INSERT INTO [dbo].[DetalleAlimentacion] (
    ID, Name, UnitName, FechaHora, AmountGrams, FishCount, PesoProm, 
    Biomasa, FeedName, SiloName, Dia, Inicio, Fin, Dif, 
    PelletFishMin, PelletPK, GramsPerSec, KgTonMin, Marca, DoserName
) VALUES (
    'TEST001', 
    'Mowi-Test-Farm', 
    'Jaula-001', 
    GETUTCDATE(), 
    125.5, 
    1000, 
    125.4, 
    125400, 
    'Premium 4.5mm', 
    'Silo-A', 
    CAST(GETUTCDATE() AS DATE),
    '08:00',
    '09:00',
    1,
    1.5,
    1.2,
    2.1,
    10.5,
    0,
    'Doser-1'
)

-- Verificar inserción
SELECT TOP 1 * FROM [dbo].[DetalleAlimentacion] ORDER BY FechaHora DESC
```

### Step 3.4: Observar logs del backend

```bash
# Terminal backend (si está corriendo)
# Deberías ver:

2026/01/12 14:30:00 worker.go:346: [info] Connected to MQTT broker
2026/01/12 14:30:05 worker.go:346: [info] Connected to MongoDB
2026/01/12 14:30:05 worker.go:346: [info] Connected to SQL Server
2026/01/12 14:30:05 worker.go:346: [info] Worker started

# Luego en el siguiente poll (cada 5 segundos):
2026/01/12 14:30:10 poller.go:82: Fetched 1 new records from Akva
2026/01/12 14:30:10 poller.go:95: Attempting to publish 1 changed events to MQTT
2026/01/12 14:30:10 poller.go:101: Published 1 changed events to MQTT
```

### Step 3.5: Verificar que el JSON se envió a MQTT

**Opción A: Con mosquitto_sub**

```bash
# Terminal nueva
mosquitto_sub -h mqtt.vmsfish.com -p 8883 \
  -t "feeding/mowi/+/" \
  -u test \
  -P test2025 \
  -v
```

**Esperado:**
```
feeding/mowi/mowitestfarm/ {"TimeStampAkva":"2026-01-12T14:30:10Z","TimeStampIngresado":"2026-01-12T14:30:10.234Z","Jaula":"001","Centro":"Mowi-Test-Farm","Gramos":125.5,"Biomasa":125400,"Peces":1000,"PesoPromedio":125.4,"Alimento":"Premium 4.5mm","Silo":"Silo-A"}
```

**Opción B: Con MQTT Explorer**
1. Abrir MQTT Explorer
2. Conectar a mqtt.vmsfish.com:8883
3. Usuario: test, Contraseña: test2025
4. Subscribe: `feeding/mowi/#`
5. Ver mensajes en tiempo real

---

## 4. Verificación de Transformación

### Test 4.1: Verificar Topic dinámico

**Entrada SQL:**
- Centro: "Mowi-Test-Farm"

**Topic esperado:**
```
feeding/mowi/mowitestfarm/
```

**Verificación:**
```bash
# En el JSON que recibes por MQTT, busca el topic:
feeding/mowi/mowitestfarm/  # ✅ Correcto

# Verificar que:
# - Se convirtió a minúsculas (Mowi → mowi)
# - Se removieron espacios y caracteres especiales
# - El patrón es: feeding/mowi/{normalized_center}/
```

### Test 4.2: Verificar mapeo de campos

**Entrada SQL:**
```
Name: "Mowi-Test-Farm"
UnitName: "Jaula-001"
AmountGrams: 125.5
FishCount: 1000
PesoProm: 125.4
Biomasa: 125400
FeedName: "Premium 4.5mm"
SiloName: "Silo-A"
```

**Salida JSON MQTT:**
```json
{
  "Centro": "Mowi-Test-Farm",     // ← Name
  "Jaula": "001",                 // ← UnitName (limpiado - solo números)
  "Gramos": 125.5,                // ← AmountGrams
  "Peces": 1000,                  // ← FishCount
  "PesoPromedio": 125.4,          // ← PesoProm
  "Biomasa": 125400,              // ← Biomasa
  "Alimento": "Premium 4.5mm",    // ← FeedName
  "Silo": "Silo-A"                // ← SiloName
}
```

**Verificación:**
```bash
# Buscar en el JSON publicado:
jq . <<< '{"Centro":"Mowi-Test-Farm","Jaula":"001",...}'

# Verificar que:
# - Centro es exactamente igual a Name
# - Jaula es UnitName con solo números (Jaula-001 → 001)
# - Gramos es AmountGrams
# - Peces es FishCount
# - etc.
```

---

## 5. Testing de Deduplicación

### Test 5.1: Mismo evento dos veces

**Inserción 1:**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (...) 
VALUES ('TEST002', 'Mowi-Dedup-Test', 'Jaula-002', GETUTCDATE(), 100, 800, 125, 100000, 'Standard', 'Silo-B', ...)
```

**Resultado esperado:**
```
✅ Se publica a MQTT
```

**Inserción 2 (idéntico):**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (...) 
VALUES ('TEST002', 'Mowi-Dedup-Test', 'Jaula-002', GETUTCDATE(), 100, 800, 125, 100000, 'Standard', 'Silo-B', ...)
```

**Resultado esperado:**
```
❌ NO se publica (ya existe en MongoDB con los mismos valores)
```

**Verificación:**
```bash
# En MQTT Explorer no deberías ver dos eventos idénticos
# Solo uno en topic: feeding/mowi/mowideduptest/

# En logs del backend:
# Primer evento: "Published 1 changed events to MQTT"
# Segundo evento: "No changes detected (fetched 1 records)"
```

### Test 5.2: Mismo evento con cambio de Biomasa

**Inserción 1:**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (..., Biomasa: 125400, ...) 
VALUES ('TEST003', ...)
```

**Inserción 2 (con Biomasa diferente):**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (..., Biomasa: 125200, ...) 
VALUES ('TEST003', ...)  -- Misma ID, Biomasa bajó
```

**Resultado esperado:**
```
✅ Se publica segunda vez (cambió Biomasa)
```

**Verificación:**
```bash
# En MQTT Explorer deberías ver 2 mensajes en el mismo topic:
feeding/mowi/xxx/ 
  - {"Biomasa": 125400, ...}  # Primero
  - {"Biomasa": 125200, ...}  # Segundo (distinto)
```

---

## 6. Testing de Manejo de Errores

### Test 6.1: JSON con caracteres especiales

**SQL Insert:**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (
    ..., Name: 'Mowi™ Farm © 2024', FeedName: 'Premium "Premium Plus" 4.5mm', ...
)
```

**Resultado esperado:**
```
✅ Se publica sin errores
```

**Verificación:**
```bash
# El JSON debe ser válido y escapado correctamente:
{
  "Centro": "Mowi Farm  2024",           // Sin caracteres especiales
  "Alimento": "Premium \"Premium Plus\" 4.5mm",  // Escapado para JSON
  ...
}
```

### Test 6.2: Valores NULL

**SQL Insert (con NULLs):**
```sql
INSERT INTO [dbo].[DetalleAlimentacion] (
    ..., FeedName: NULL, SiloName: NULL, ...
)
```

**Resultado esperado:**
```
✅ Se publica (los campos NULL se mapean a strings vacíos)
```

**Verificación:**
```bash
# En el JSON deberían aparecer como strings vacíos:
{
  "Alimento": "",  // NULL → ""
  "Silo": "",      // NULL → ""
  ...
}
```

---

## 7. Verificación de Rendimiento

### Test 7.1: Batch de 100 eventos

```sql
-- Insertar 100 eventos
DECLARE @i INT = 0
WHILE @i < 100
BEGIN
  INSERT INTO [dbo].[DetalleAlimentacion] (...)
  VALUES ('TEST-' + CAST(@i AS VARCHAR(10)), 'Mowi-Perf-Test', 'Jaula-' + CAST(@i AS VARCHAR(10)), ...)
  SET @i = @i + 1
END
```

**Resultado esperado:**
```
✅ Se publican 100 eventos en < 5 segundos
```

**Verificación:**
```bash
# Logs del backend:
2026/01/12 15:00:00 poller.go:82: Fetched 100 new records from Akva
2026/01/12 15:00:02 poller.go:95: Attempting to publish 100 changed events to MQTT
2026/01/12 15:00:04 poller.go:101: Published 100 changed events to MQTT

# Latencia: ~2-4 segundos para 100 eventos es OK
```

---

## 8. Checklist Final

- [ ] Backend compila sin errores
- [ ] Backend conecta a MQTT (`[info] Connected to MQTT broker`)
- [ ] Frontend muestra MQTT como "conectado" (✅ verde)
- [ ] SQL Server tiene datos de prueba
- [ ] Datos fluyen desde SQL → Logger de Backend
- [ ] JSONs se publican a MQTT
- [ ] Puedo recibir JSONs con mosquitto_sub
- [ ] JSONs tienen estructura correcta
- [ ] Topics dinámicos se crean correctamente
- [ ] Deduplicación funciona
- [ ] Manejo de errores es correcto
- [ ] Rendimiento es aceptable

---

## 9. Comandos Rápidos para Testing

```bash
# Verificar compilación
cd f:\vscode\omnipoll\backend
go build -o omnipoll.exe ./cmd/omnipoll

# Iniciar backend
.\omnipoll.exe

# Iniciar frontend (en otra terminal)
cd f:\vscode\omnipoll\frontend
npm run dev

# Verificar MQTT en tiempo real
mosquitto_sub -h mqtt.vmsfish.com -p 8883 -t "feeding/mowi/+/" -u test -P test2025 -v

# Buscar JSONs en logs
grep "json.Marshal\|Publish\|changed events" logs.txt
```

---

**Estado:** Listo para testing  
**Última actualización:** 2026-01-12
