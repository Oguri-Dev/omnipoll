# Verificación del Flujo de Creación de JSON para MQTT

## Resumen del Flujo

```
SQL Server (Akva)
       ↓ FetchNewRecords()
Registros DetalleAlimentacion
       ↓ ToNormalizedEvents() → Mapper
NormalizedEvent[] (formato interno)
       ↓ FilterChangedEvents() (opcional)
Filtered Events
       ↓ PublishBatch() → Publisher
JSON + Topic MQTT
       ↓ client.Publish()
MQTT Broker (mqtt.vmsfish.com:8883)
```

---

## 1. Origen de Datos: SQL Server (Akva)

### Localización

📁 `backend/internal/akva/client.go`

### Función de Lectura

```go
func (c *Client) FetchNewRecords(ctx context.Context, lastFechaHora time.Time, ...) ([]DetalleAlimentacion, error)
```

### Campos Extraídos (DetalleAlimentacion)

```go
type DetalleAlimentacion struct {
    ID            string      // Identificador único
    Name          string      // Centro/Granja
    UnitName      string      // Jaula/Unidad
    FechaHora     time.Time   // Timestamp del evento
    Dia           time.Time   // Fecha del evento
    Inicio        string      // Hora de inicio
    Fin           string      // Hora de fin
    Dif           int         // Diferencia
    AmountGrams   float64     // Gramos de alimento
    PelletFishMin float64     // Pellets por pez y minuto
    FishCount     float64     // Cantidad de peces
    PesoProm      float64     // Peso promedio
    Biomasa       float64     // Biomasa total
    PelletPK      float64     // Pellets por kg
    FeedName      string      // Nombre del alimento
    SiloName      string      // Nombre del silo
    DoserName     string      // Dosificador
    GramsPerSec   float64     // Gramos/segundo
    KgTonMin      float64     // Kg/ton/min
    Marca         int         // Marca/Flag
}
```

---

## 2. Transformación: Mapper

### Localización

📁 `backend/internal/akva/mapper.go`

### Función Principal

```go
func ToNormalizedEvent(record DetalleAlimentacion) events.NormalizedEvent
```

### Procesos

1. **Sanitización UTF-8** - Limpia caracteres inválidos
2. **Conversión de Fechas** - A formato RFC3339
3. **Mapeo de Campos** - Normaliza nombres de campos a camelCase

### Estructura Normalizada (NormalizedEvent)

```go
type NormalizedEvent struct {
    ID            string    `json:"id"`                  // akva:123
    Source        string    `json:"source"`              // "akva"
    Name          string    `json:"name"`                // Centro
    UnitName      string    `json:"unitName"`            // Jaula
    FechaHora     string    `json:"fechaHora"`           // 2026-01-12T23:38:40Z
    Dia           string    `json:"dia"`                 // 2026-01-12
    Inicio        string    `json:"inicio"`
    Fin           string    `json:"fin"`
    Dif           int       `json:"dif"`
    AmountGrams   float64   `json:"amountGrams"`         // 125.5
    PelletFishMin float64   `json:"pelletFishMin"`
    FishCount     float64   `json:"fishCount"`           // 1000
    PesoProm      float64   `json:"pesoProm"`            // 125.4
    Biomasa       float64   `json:"biomasa"`             // 125400
    PelletPK      float64   `json:"pelletPK"`
    FeedName      string    `json:"feedName"`            // "Premium"
    SiloName      string    `json:"siloName"`            // "Silo-A"
    DoserName     string    `json:"doserName"`
    GramsPerSec   float64   `json:"gramsPerSec"`
    KgTonMin      float64   `json:"kgTonMin"`
    Marca         int       `json:"marca"`
    IngestedAt    time.Time `json:"ingestedAt"`          // Timestamp ingesta
}
```

---

## 3. Filtrado (Opcional)

### Localización

📁 `backend/internal/poller/poller.go` - `filterChangedEvents()`

### Propósito

Comparar eventos nuevos con MongoDB y solo publicar si hay cambios en campos business-critical:

- `AmountGrams` (cantidad de alimento)
- `FishCount` (cantidad de peces)
- `PesoProm` (peso promedio)
- `Biomasa` (biomasa)
- `FeedName` (alimento)

### Lógica

```
FOR cada evento nuevo:
  IF evento NO existe en MongoDB → PUBLICAR
  IF evento existe pero CAMBIÓ biomasa/cantidad → PUBLICAR
  IF evento igual → NO PUBLICAR (ya se publicó antes)
```

---

## 4. Publicación a MQTT

### Localización

📁 `backend/internal/mqtt/publisher.go`

### Función Principal

```go
func (p *Publisher) PublishBatch(evts []events.NormalizedEvent) error
```

### Transformación a MQTTMessage

**Entrada:** NormalizedEvent

```json
{
  "id": "akva:12345",
  "source": "akva",
  "name": "Centro-Mowi",
  "unitName": "Jaula-05",
  "fechaHora": "2026-01-12T23:38:40Z",
  "amountGrams": 125.5,
  "fishCount": 1000,
  "pesoProm": 125.4,
  "biomasa": 125400,
  "feedName": "Premium 4.5mm",
  "siloName": "Silo-A"
}
```

**Transformación:**

```go
msg := MQTTMessage{
    TimeStampAkva:      "2026-01-12T23:38:40Z",          // FechaHora
    TimeStampIngresado: "2026-01-12T23:38:41.123Z",      // IngestedAt
    Jaula:              "05",                             // cleanJaula(UnitName) - solo números
    Centro:             "Centro-Mowi",                    // Name
    Gramos:             125.5,                            // AmountGrams
    Biomasa:            125400,                           // Biomasa
    Peces:              1000,                             // FishCount
    PesoPromedio:       125.4,                            // PesoProm
    Alimento:           "Premium 4.5mm",                  // FeedName
    Silo:               "Silo-A",                         // SiloName
}
```

**Salida JSON:**

```json
{
  "TimeStampAkva": "2026-01-12T23:38:40Z",
  "TimeStampIngresado": "2026-01-12T23:38:41.123Z",
  "Jaula": "05",
  "Centro": "Centro-Mowi",
  "Gramos": 125.5,
  "Biomasa": 125400,
  "Peces": 1000,
  "PesoPromedio": 125.4,
  "Alimento": "Premium 4.5mm",
  "Silo": "Silo-A"
}
```

### Cálculo de Topic Dinámico

**Función:** `buildDynamicTopic()`

```go
// Input: "Centro-Mowi"
// 1. Lowercase: "centro-mowi"
// 2. Replace spaces: "centro-mowi" (sin espacios)
// 3. Remove special chars: "centromowi" (sin - ni caracteres especiales)
// 4. Output: "feeding/mowi/centromowi/"
```

**Topic Final:** `feeding/mowi/centromowi/`

### Publicación

```go
token := client.Publish(
    topic,           // "feeding/mowi/centromowi/"
    cfg.QoS,         // QoS = 1 (de config.yaml)
    false,           // retain = false
    payload,         // JSON marshalled
)
token.WaitTimeout(10 * time.Second)  // Espera hasta 10 segundos
```

---

## 5. Flujo Completo con Ejemplo

### Entrada desde SQL Server

```
ID: A001
Name: Centro-Mowi
UnitName: Jaula-05
FechaHora: 2026-01-12 23:38:40 UTC
AmountGrams: 125.5
FishCount: 1000
PesoProm: 125.4
Biomasa: 125400
FeedName: Premium 4.5mm
SiloName: Silo-A
```

### Paso 1: Mapper → NormalizedEvent

```json
{
  "id": "akva:A001",
  "source": "akva",
  "name": "Centro-Mowi",
  "unitName": "Jaula-05",
  "fechaHora": "2026-01-12T23:38:40Z",
  "dia": "2026-01-12",
  "amountGrams": 125.5,
  "fishCount": 1000,
  "pesoProm": 125.4,
  "biomasa": 125400,
  "feedName": "Premium 4.5mm",
  "siloName": "Silo-A",
  "ingestedAt": "2026-01-12T23:38:41.123456Z"
}
```

### Paso 2: Filtrado (opcional)

- ¿Existe en MongoDB? NO → PUBLICAR
- ¿Cambió biomasa? → PUBLICAR
- ¿Cambió cantidad? → PUBLICAR

### Paso 3: Publisher → MQTTMessage

```json
{
  "TimeStampAkva": "2026-01-12T23:38:40Z",
  "TimeStampIngresado": "2026-01-12T23:38:41.123456Z",
  "Jaula": "05",
  "Centro": "Centro-Mowi",
  "Gramos": 125.5,
  "Biomasa": 125400,
  "Peces": 1000,
  "PesoPromedio": 125.4,
  "Alimento": "Premium 4.5mm",
  "Silo": "Silo-A"
}
```

### Paso 4: Publicación a MQTT

```
Topic: feeding/mowi/centromowi/
QoS: 1
Payload: {"TimeStampAkva": "2026-01-12T23:38:40Z", ...}
Broker: mqtt.vmsfish.com:8883
```

---

## 6. Características del JSON para MQTT

### ✅ Lo que SÍ incluye

- Timestamps (Akva e Ingesta)
- Centro y Jaula (identificadores)
- Datos nutricionales (Gramos, Biomasa, Peces, PesoPromedio)
- Alimento y Silo
- QoS 1 (garantía de al menos una entrega)

### ❌ Lo que NO incluye

- ID interno (no se envía)
- Source (no se envía)
- Campos internos (Dif, PelletFishMin, Marca, etc.)
- Información de MongoDB (ID del documento)

---

## 7. Validación de JSON

### Paso 1: json.Marshal() en Publisher

```go
payload, err := json.Marshal(msg)
if err != nil {
    return fmt.Errorf("failed to marshal event: %w", err)
}
```

✅ Valida tipos y estructura antes de enviar

### Paso 2: Publicación

```go
token := client.Publish(topic, cfg.QoS, false, payload)
if token.WaitTimeout(10 * time.Second) && token.Error() != nil {
    return fmt.Errorf("failed to publish message: %w", token.Error())
}
```

✅ Espera confirmación del broker (QoS 1)

---

## 8. Estado Actual en Tu Sistema

### ✅ Configurado Correctamente

- MQTT conectado a `mqtt.vmsfish.com:8883`
- JSON marshalling implementado
- Topics dinámicos basados en Centro
- QoS 1 configurado

### ❌ No Puede Funcionar Todavía

- **SQL Server no disponible** → No hay registros para leer
- **MongoDB no disponible** → No puede filtrar cambios
- **Poll error: not connected** → Sin SQL + MongoDB, no hay nothing que procesar

---

## 9. Cómo Verificar (Cuando tengas Servicios)

### Opción 1: Ver logs de publicación

```bash
# En los logs del backend deberías ver:
2026/01/12 23:38:50 poller.go:95: Attempting to publish 5 changed events to MQTT
2026/01/12 23:38:50 poller.go:101: Published 5 changed events to MQTT (fetched 12 total)
```

### Opción 2: Suscribirse a MQTT (con cliente externo)

```bash
mosquitto_sub -h mqtt.vmsfish.com -p 8883 -t "feeding/mowi/+/" --cafile ca.crt -u test -P test2025
```

### Opción 3: Usar MQTT Explorer (GUI)

- Host: mqtt.vmsfish.com
- Port: 8883
- Username: test
- Password: test2025
- Subscribe to: `feeding/mowi/#`

---

## 10. Resumen de Archivos Involucrados

| Archivo             | Responsabilidad                                 |
| ------------------- | ----------------------------------------------- |
| `akva/client.go`    | Lee de SQL Server                               |
| `akva/mapper.go`    | Convierte DetalleAlimentacion → NormalizedEvent |
| `events/event.go`   | Define estructura de NormalizedEvent            |
| `mqtt/publisher.go` | Convierte NormalizedEvent → MQTTMessage + JSON  |
| `poller/poller.go`  | Orquesta todo (fetch → filter → publish)        |
| `config.yaml`       | Configuración (broker, credenciales, QoS)       |

---

## 11. Próximos Pasos

Para probar el flujo completo necesitas:

1. **Levantar servicios:**
   ```bash
   docker-compose up -d
   ```
2. **Insertar datos de prueba en SQL Server:**

   ```sql
   INSERT INTO Akva...DetalleAlimentacion
   VALUES (...)
   ```

3. **Ver logs de publicación:**

   ```bash
   # Terminal del backend mostrará los publishes
   tail -f logs.txt
   ```

4. **Verificar en MQTT:**
   ```bash
   mosquitto_sub -h mqtt.vmsfish.com -p 8883 \
     -t "feeding/mowi/+/" \
     -u test -P test2025
   ```

---

**Última actualización:** 2026-01-12
