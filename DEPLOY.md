# Guía de Despliegue en Docker (Linux)

## Prerrequisitos

- Docker y Docker Compose instalados en el servidor Linux
- Acceso de red al SQL Server (Akva)
- Puertos disponibles: 8080 (admin panel), 1883 (MQTT), 27017 (MongoDB)

## Pasos de Despliegue

### 1. Clonar o transferir el proyecto al servidor Linux

```bash
# Opción A: Clonar desde repositorio
git clone <tu-repo> omnipoll
cd omnipoll

# Opción B: Transferir con rsync (desde Windows)
rsync -avz --exclude 'node_modules' --exclude '.git' \
  /c/Users/Andres/Documents/ReactWork/Omnipoll/ \
  usuario@servidor-linux:/home/usuario/omnipoll/
```

### 2. Construir el frontend

En el servidor Linux o en Windows antes de transferir:

```bash
cd frontend
npm install
npm run build
```

Esto genera `frontend/dist/` que el Dockerfile del backend copiará a `web/dist/`.

### 3. Mover el build del frontend al backend

```bash
# Desde la raíz del proyecto
mkdir -p backend/web
cp -r frontend/dist backend/web/
```

### 4. Configurar variables de entorno

Crear archivo `.env` en la raíz del proyecto:

```bash
# .env
OMNIPOLL_MASTER_KEY=tu-clave-secreta-de-32-caracteres-minimo

# SQL Server Akva (ajustar según tu red)
SQL_SERVER_HOST=ip-servidor-akva
SQL_SERVER_PORT=1433
SQL_SERVER_DATABASE=FTFeeding
SQL_SERVER_USER=tu-usuario
SQL_SERVER_PASSWORD=tu-password
```

### 5. Configurar config.yaml

Editar `backend/data/config.yaml` con las credenciales reales:

```yaml
sqlServer:
  host: ${SQL_SERVER_HOST:-172.17.0.1} # IP del host Docker o servidor Akva
  port: 1433
  database: FTFeeding
  user: sa
  password: 'TuPasswordAqui' # Se encriptará automáticamente

mqtt:
  broker: mosquitto # Nombre del servicio en docker-compose
  port: 1883
  topic: ftfeeding/akva/detalle
  clientId: omnipoll-worker
  user: ''
  password: ''
  qos: 1

mongodb:
  uri: mongodb://mongodb:27017 # Nombre del servicio en docker-compose
  database: omnipoll
  collection: historical_events

polling:
  intervalMs: 5000 # Polling cada 5 segundos
  batchSize: 100

admin:
  host: 0.0.0.0 # Escuchar en todas las interfaces
  port: 8080
  username: admin
  password: 'admin123' # Cambiar en producción
```

### 6. Actualizar docker-compose.yml

Asegúrate de que el docker-compose.yml tenga la red correcta:

```yaml
version: '3.8'

services:
  omnipoll:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - '8080:8080'
    environment:
      - OMNIPOLL_MASTER_KEY=${OMNIPOLL_MASTER_KEY}
      - OMNIPOLL_CONFIG_PATH=/app/data/config.yaml
      - OMNIPOLL_WATERMARK_PATH=/app/data/watermark.json
    volumes:
      - omnipoll_data:/app/data
      - ./backend/data:/app/data # Montar config local
    depends_on:
      - mongodb
      - mosquitto
    restart: unless-stopped
    extra_hosts:
      - 'host.docker.internal:host-gateway' # Para acceder al SQL Server del host

  mongodb:
    image: mongo:7
    ports:
      - '27017:27017'
    volumes:
      - mongodb_data:/data/db
    restart: unless-stopped

  mosquitto:
    image: eclipse-mosquitto:2
    ports:
      - '1883:1883'
      - '9001:9001'
    volumes:
      - ./mosquitto/config:/mosquitto/config
      - mosquitto_data:/mosquitto/data
      - mosquitto_log:/mosquitto/log
    restart: unless-stopped

volumes:
  omnipoll_data:
  mongodb_data:
  mosquitto_data:
  mosquitto_log:
```

### 7. Construir y levantar los servicios

```bash
# Construir las imágenes
docker-compose build

# Levantar todos los servicios
docker-compose up -d

# Ver logs en tiempo real
docker-compose logs -f omnipoll

# Verificar estado
docker-compose ps
```

### 8. Verificar el despliegue

1. **Admin Panel**: Abre http://ip-servidor:8080 en tu navegador

   - Usuario: `admin`
   - Contraseña: `admin123` (o la que configuraste)

2. **Verificar conexiones**: En el dashboard deberías ver:

   - MongoDB: ✅ Connected
   - MQTT: ✅ Connected
   - SQL Server: ✅ Connected (si la config es correcta)

3. **Ver logs del worker**:

   ```bash
   docker-compose logs -f omnipoll
   ```

4. **Verificar eventos ingeridos**:
   - Ve a la página "Events" en el admin panel
   - Deberías ver los registros de TB_DetalleAlimentacion

### 9. Iniciar el polling

Desde el admin panel:

1. Ve al Dashboard
2. Haz clic en "Start Worker"
3. Observa los logs para verificar la ingesta de datos

## Comandos Útiles

```bash
# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (resetear datos)
docker-compose down -v

# Reiniciar solo Omnipoll
docker-compose restart omnipoll

# Ver logs de un servicio específico
docker-compose logs -f mongodb
docker-compose logs -f mosquitto

# Ejecutar comandos dentro del contenedor
docker-compose exec omnipoll sh

# Ver configuración actual
docker-compose exec omnipoll cat /app/data/config.yaml

# Ver watermark actual
docker-compose exec omnipoll cat /app/data/watermark.json

# Backup de MongoDB
docker-compose exec mongodb mongodump --out /data/backup

# Restaurar MongoDB
docker-compose exec mongodb mongorestore /data/backup
```

## Acceso desde SQL Server en Host Docker

Si el SQL Server corre en el mismo servidor Linux pero fuera de Docker:

1. Usar `host.docker.internal` en el config.yaml (ya configurado en docker-compose)
2. O usar la IP del bridge de Docker (generalmente 172.17.0.1)

```yaml
sqlServer:
  host: host.docker.internal # o 172.17.0.1
```

## Acceso desde SQL Server en Otra Máquina

Si el SQL Server está en otra máquina de la red:

```yaml
sqlServer:
  host: 192.168.x.x # IP del servidor Akva
  port: 1433
```

Asegúrate de que:

- El firewall del servidor Akva permita conexiones al puerto 1433
- SQL Server esté configurado para aceptar conexiones remotas
- El usuario tenga permisos de lectura en `TB_DetalleAlimentacion`

## Seguridad en Producción

1. **Cambiar contraseñas por defecto**:

   ```yaml
   admin:
     username: admin
     password: 'contraseña-segura-aqui'
   ```

2. **Usar OMNIPOLL_MASTER_KEY fuerte**:

   ```bash
   # Generar clave aleatoria de 32 bytes
   openssl rand -hex 32
   ```

3. **Restringir acceso al puerto 8080**:

   ```bash
   # Solo desde red local
   iptables -A INPUT -p tcp --dport 8080 -s 192.168.0.0/16 -j ACCEPT
   iptables -A INPUT -p tcp --dport 8080 -j DROP
   ```

4. **Usar HTTPS** (opcional):
   - Configurar un reverse proxy (nginx/traefik)
   - Agregar certificados SSL/TLS

## Troubleshooting

### Error: Cannot connect to SQL Server

```bash
# Verificar conectividad desde el contenedor
docker-compose exec omnipoll ping host.docker.internal
docker-compose exec omnipoll nc -zv host.docker.internal 1433
```

### Error: Permission denied writing config

```bash
# Ajustar permisos del directorio data
sudo chown -R 1000:1000 backend/data
```

### Frontend no carga

```bash
# Verificar que web/dist existe en el contenedor
docker-compose exec omnipoll ls -la /app/web/dist

# Reconstruir con el frontend
cd frontend && npm run build && cd ..
cp -r frontend/dist backend/web/
docker-compose build --no-cache omnipoll
docker-compose up -d omnipoll
```

### MongoDB sin datos

```bash
# Verificar colecciones
docker-compose exec mongodb mongosh omnipoll --eval "db.historical_events.countDocuments()"

# Ver eventos recientes
docker-compose exec mongodb mongosh omnipoll --eval "db.historical_events.find().limit(5).pretty()"
```

## Monitoreo

### Ver métricas del worker

Endpoint: `GET http://localhost:8080/api/status`

```json
{
  "worker": {
    "running": true,
    "lastPoll": "2026-01-09T11:52:00Z",
    "stats": {
      "totalPolls": 150,
      "totalRecordsFetched": 1234,
      "totalRecordsPublished": 1234,
      "totalRecordsPersisted": 1234,
      "lastError": ""
    }
  },
  "connections": {
    "sqlServer": true,
    "mqtt": true,
    "mongodb": true
  }
}
```

### Logs estructurados

Todos los logs incluyen timestamp, nivel y contexto:

```
2026/01/09 11:51:49 worker.go:241: [info] Connected to MQTT broker
2026/01/09 11:52:00 poller.go:85: [info] Fetched 10 new records from Akva
2026/01/09 11:52:00 poller.go:105: [info] Published 10 events to MQTT
2026/01/09 11:52:00 poller.go:115: [info] Persisted 10 events to MongoDB
```

## Actualización del Sistema

```bash
# 1. Detener servicios
docker-compose down

# 2. Actualizar código
git pull

# 3. Reconstruir frontend
cd frontend && npm install && npm run build && cd ..
cp -r frontend/dist backend/web/

# 4. Reconstruir imágenes
docker-compose build

# 5. Levantar servicios
docker-compose up -d

# 6. Verificar logs
docker-compose logs -f omnipoll
```
