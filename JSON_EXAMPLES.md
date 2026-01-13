# Ejemplos de JSON en MQTT - Omnipoll

## Escenario 1: Evento Simple de Alimentación

### Entrada desde SQL Server
```
Centro: Mowi Feeding Farm
Jaula: Unit-A-005
Fecha: 2026-01-12 10:30:00
Gramos: 150.75
Peces: 5000
Peso Promedio: 100.5g
Biomasa: 502500g
Alimento: Premium 4.5mm
Silo: Main-Silo-1
```

### Topic MQTT
```
feeding/mowi/mowifeedingfarm/
```

### JSON Payload (Pretty Print)
```json
{
  "TimeStampAkva": "2026-01-12T10:30:00Z",
  "TimeStampIngresado": "2026-01-12T10:30:02.456Z",
  "Jaula": "005",
  "Centro": "Mowi Feeding Farm",
  "Gramos": 150.75,
  "Biomasa": 502500,
  "Peces": 5000,
  "PesoPromedio": 100.5,
  "Alimento": "Premium 4.5mm",
  "Silo": "Main-Silo-1"
}
```

### JSON Payload (Compact - Lo que se envía por MQTT)
```
{"TimeStampAkva":"2026-01-12T10:30:00Z","TimeStampIngresado":"2026-01-12T10:30:02.456Z","Jaula":"005","Centro":"Mowi Feeding Farm","Gramos":150.75,"Biomasa":502500,"Peces":5000,"PesoPromedio":100.5,"Alimento":"Premium 4.5mm","Silo":"Main-Silo-1"}
```

**Tamaño:** ~200 bytes

---

## Escenario 2: Múltiples Jaulas en Batch

### Evento 1
```json
{
  "TimeStampAkva": "2026-01-12T10:30:00Z",
  "TimeStampIngresado": "2026-01-12T10:30:02.456Z",
  "Jaula": "001",
  "Centro": "Mowi-North",
  "Gramos": 120,
  "Biomasa": 480000,
  "Peces": 4000,
  "PesoPromedio": 120,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-A"
}
```

### Evento 2
```json
{
  "TimeStampAkva": "2026-01-12T10:30:05Z",
  "TimeStampIngresado": "2026-01-12T10:30:07.789Z",
  "Jaula": "002",
  "Centro": "Mowi-North",
  "Gramos": 135,
  "Biomasa": 495000,
  "Peces": 4100,
  "PesoPromedio": 120.7,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-A"
}
```

### Evento 3 (Centro diferente)
```json
{
  "TimeStampAkva": "2026-01-12T10:30:10Z",
  "TimeStampIngresado": "2026-01-12T10:30:12.234Z",
  "Jaula": "045",
  "Centro": "Mowi-South",
  "Gramos": 200,
  "Biomasa": 750000,
  "Peces": 6000,
  "PesoPromedio": 125,
  "Alimento": "Premium 4.5mm",
  "Silo": "Silo-C"
}
```

### Topics Publicados
```
feeding/mowi/mowinog/      # Evento 1 y 2
feeding/mowi/mowisouth/    # Evento 3
```

---

## Escenario 3: Cambio de Alimento

### Antes
```json
{
  "TimeStampAkva": "2026-01-12T09:00:00Z",
  "TimeStampIngresado": "2026-01-12T09:00:01.123Z",
  "Jaula": "010",
  "Centro": "Mowi-East",
  "Gramos": 100,
  "Biomasa": 400000,
  "Peces": 3200,
  "PesoPromedio": 125,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-B"
}
```

### Después (mismo evento con alimento diferente)
```json
{
  "TimeStampAkva": "2026-01-12T10:00:00Z",
  "TimeStampIngresado": "2026-01-12T10:00:02.456Z",
  "Jaula": "010",
  "Centro": "Mowi-East",
  "Gramos": 100,
  "Biomasa": 400000,
  "Peces": 3200,
  "PesoPromedio": 125,
  "Alimento": "Premium 4.5mm",    // ← CAMBIÓ
  "Silo": "Silo-B"
}
```

**Resultado:** Se publica porque cambió el campo `Alimento`

---

## Escenario 4: Eventos Duplicados (No se Publican)

### Primer evento (se publica)
```json
{
  "TimeStampAkva": "2026-01-12T11:00:00Z",
  "TimeStampIngresado": "2026-01-12T11:00:01.000Z",
  "Jaula": "020",
  "Centro": "Mowi-West",
  "Gramos": 175,
  "Biomasa": 525000,
  "Peces": 4200,
  "PesoPromedio": 125,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-D"
}
```

### Segundo evento idéntico (NO se publica)
```json
{
  "TimeStampAkva": "2026-01-12T11:00:00Z",
  "TimeStampIngresado": "2026-01-12T11:05:00.000Z",  // ← Diferente timestamp
  "Jaula": "020",
  "Centro": "Mowi-West",
  "Gramos": 175,                                      // ← Igual
  "Biomasa": 525000,                                  // ← Igual
  "Peces": 4200,                                      // ← Igual
  "PesoPromedio": 125,                                // ← Igual
  "Alimento": "Standard 3.5mm",                       // ← Igual
  "Silo": "Silo-D"
}
```

**Resultado:** NO se publica (filtrado por deduplicación)

---

## Escenario 5: Variación de Biomasa

### Evento 1
```json
{
  "TimeStampAkva": "2026-01-12T12:00:00Z",
  "TimeStampIngresado": "2026-01-12T12:00:01.000Z",
  "Jaula": "030",
  "Centro": "Mowi-Central",
  "Gramos": 180,
  "Biomasa": 540000,      // ← 540 kg
  "Peces": 4320,
  "PesoPromedio": 125,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-E"
}
```

### Evento 2 (biomasa cambió por mortandad)
```json
{
  "TimeStampAkva": "2026-01-12T13:00:00Z",
  "TimeStampIngresado": "2026-01-12T13:00:01.000Z",
  "Jaula": "030",
  "Centro": "Mowi-Central",
  "Gramos": 180,
  "Biomasa": 535000,      // ← 535 kg (bajó 5 kg)
  "Peces": 4300,          // ← Bajó 20 peces
  "PesoPromedio": 124.4,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-E"
}
```

**Resultado:** SE publica (cambió Biomasa y FishCount)

---

## Escenario 6: Caracteres Especiales Sanitizados

### Entrada SQL Server (con caracteres problemáticos)
```
Centro: Mowi™ Feeding® Farm © 2024
Jaula: Jaula - #5 (ñ)
Alimento: Alimento "Premium" 4.5mm [EXTRA]
Silo: Silo @ Main
```

### JSON después de sanitización
```json
{
  "TimeStampAkva": "2026-01-12T14:00:00Z",
  "TimeStampIngresado": "2026-01-12T14:00:01.000Z",
  "Jaula": "5",                          // ← Solo números
  "Centro": "Mowi Feeding Farm",         // ← Sin símbolos
  "Gramos": 160,
  "Biomasa": 480000,
  "Peces": 3000,
  "PesoPromedio": 160,
  "Alimento": "Alimento \"Premium\" 4.5mm [EXTRA]",  // ← Escapado para JSON
  "Silo": "Silo @ Main"
}
```

### Topic calculado
```
feeding/mowi/mowifeedingfarm/   // Sin caracteres especiales
```

---

## Escenario 7: Timestamps en Diferentes Zonas Horarias

### Entrada SQL Server (zona local)
```
FechaHora: 2026-01-12 15:30:00 (hora local Chile, UTC-3)
```

### Conversión en Mapper
```go
FechaHora: record.FechaHora.UTC().Format(time.RFC3339)
// Resultado: "2026-01-12T18:30:00Z"  ← UTC
```

### JSON publicado
```json
{
  "TimeStampAkva": "2026-01-12T18:30:00Z",           // ← UTC
  "TimeStampIngresado": "2026-01-12T18:30:02.789Z",  // ← UTC
  "Jaula": "050",
  "Centro": "Mowi-UTC-Test",
  "Gramos": 145,
  "Biomasa": 435000,
  "Peces": 3000,
  "PesoPromedio": 145,
  "Alimento": "Standard 3.5mm",
  "Silo": "Silo-F"
}
```

---

## Cómo Leer los JSONs en Tiempo Real

### 1. Con mosquitto_sub (línea de comando)
```bash
mosquitto_sub -h mqtt.vmsfish.com -p 8883 \
  -t "feeding/mowi/+/" \
  -u test \
  -P test2025 \
  -v  # Mostrar topic también
```

**Output:**
```
feeding/mowi/mowifeedingfarm/ {"TimeStampAkva":"2026-01-12T10:30:00Z","TimeStampIngresado":"2026-01-12T10:30:02.456Z","Jaula":"005","Centro":"Mowi Feeding Farm","Gramos":150.75,...}
feeding/mowi/mowinog/ {"TimeStampAkva":"2026-01-12T10:30:05Z","TimeStampIngresado":"2026-01-12T10:30:07.789Z","Jaula":"001","Centro":"Mowi-North","Gramos":120,...}
```

### 2. Con MQTT Explorer (GUI)
1. Instalar: https://mqtt-explorer.com/
2. Conectar:
   - Broker: `mqtt.vmsfish.com`
   - Port: `8883`
   - User: `test`
   - Password: `test2025`
3. Subscribe: `feeding/mowi/#`
4. Ver JSONs en el panel derecho (con pretty print)

### 3. Con Node-RED
1. Agregar nodo "mqtt in"
2. Configurar broker: mqtt.vmsfish.com:8883
3. Topic: `feeding/mowi/#`
4. Conectar a "debug" para ver JSONs
5. Los JSONs se parsean automáticamente

### 4. Logear en el backend
Agregar log en `publisher.go`:
```go
log.Printf("Publishing to %s: %s", topic, string(payload))
```

---

## Tamaños Esperados

| Componente | Tamaño |
|-----------|--------|
| JSON payload (típico) | 150-250 bytes |
| Header MQTT | ~20 bytes |
| Total por mensaje | 170-270 bytes |
| 1000 eventos/hora | ~200-270 KB |
| 10000 eventos/hora | ~2-2.7 MB |

---

## Validación de JSON

El JSON se valida en dos puntos:

### 1. Marshalling en Publisher
```go
payload, err := json.Marshal(msg)
if err != nil {
    log.Printf("ERROR: Failed to marshal: %v", err)
    return err
}
// Si hay error, el evento NO se publica
```

### 2. Publicación a MQTT
```go
token := client.Publish(topic, cfg.QoS, false, payload)
if token.WaitTimeout(10 * time.Second) && token.Error() != nil {
    log.Printf("ERROR: Failed to publish: %v", token.Error())
    return err
}
```

---

## Resumen

| Aspecto | Detalle |
|--------|--------|
| **Formato** | JSON válido, UTF-8 |
| **Topic** | Dinámico por Centro |
| **Payload** | MQTTMessage con 10 campos |
| **QoS** | 1 (al menos una entrega) |
| **Tamaño típico** | 150-250 bytes |
| **Frecuencia** | Cada poll (configurable) |
| **Filtrado** | Deduplicación por cambios |
| **Sanitización** | UTF-8 y caracteres especiales |
| **Timestamps** | RFC3339 en UTC |
| **Validación** | En Marshal + Publish |

---

**Última actualización:** 2026-01-12
**Estado del Sistema:** Listo para producción (requiere servicios)
