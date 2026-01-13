# Guía de Producción: Testing y Despliegue de Omnipoll

## Opciones de Despliegue

```
┌─────────────────────────────────────────────────────────────────┐
│                   OMNIPOLL - Opciones de Deployment             │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  OPCIÓN 1: Local Development (Windows)                          │
│  ├─ Backend: localhost:8080                                     │
│  ├─ Frontend: localhost:3001                                    │
│  ├─ MQTT: mqtt.vmsfish.com:8883 (nube)                         │
│  ├─ SQL/Mongo: Via Docker Compose local                        │
│  └─ Ideal para: Desarrollo, testing                            │
│                                                                   │
│  OPCIÓN 2: Docker Local (Windows/Mac/Linux)                    │
│  ├─ Backend: Docker container :8080                             │
│  ├─ MongoDB: Docker container :27017                           │
│  ├─ MQTT: Docker o nube :1883/8883                            │
│  ├─ SQL: Servidor remoto                                       │
│  └─ Ideal para: Pre-producción, testing completo              │
│                                                                   │
│  OPCIÓN 3: Linux Production Server                              │
│  ├─ Backend: /app/omnipoll en servidor Linux                   │
│  ├─ MongoDB: Container Linux                                   │
│  ├─ MQTT: Mosquitto o nube                                     │
│  ├─ SQL: Servidor remoto de producción                         │
│  ├─ Nginx: Proxy reverso, HTTPS                               │
│  └─ Ideal para: Producción real                                │
│                                                                   │
│  OPCIÓN 4: Cloud (AWS/Azure/GCP)                               │
│  ├─ Kubernetes o ECS                                           │
│  ├─ Managed MongoDB (Atlas)                                    │
│  ├─ CloudSQL para SQL Server                                   │
│  ├─ Managed MQTT (IoT Hub)                                     │
│  └─ Ideal para: Escalabilidad, alta disponibilidad             │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔷 OPCIÓN 1: Testing Local (Recomendado Primero)

### Estado Actual
```
✅ Backend compilado y corriendo
✅ Frontend en desarrollo (npm run dev)
✅ MQTT conectado a nube (mqtt.vmsfish.com:8883)
❌ SQL Server no disponible
❌ MongoDB no disponible
```

### Pasos para Testing Local Completo

#### 1.1 Levantar Servicios con Docker

```bash
cd f:\vscode\omnipoll

# Iniciar SQL Server + MongoDB con Docker Compose
docker-compose up -d mssql mongodb

# O solo MongoDB (si SQL está en otro lado)
docker-compose up -d mongodb
```

**Verificar estado:**
```bash
docker ps
```

Expected:
```
CONTAINER ID   IMAGE                      STATUS
abc123         mcr.microsoft.com/mssql... Up 2 minutes
def456         mongo:latest               Up 2 minutes
```

#### 1.2 Esperar a que SQL Server inicie (2-3 minutos)

```bash
docker logs mssql | tail -20
# Buscar: "Recovery completed" o "SQL Server started successfully"
```

#### 1.3 Verificar Backend detecta las conexiones

**En terminal backend:**
```
2026/01/12 14:30:00 worker.go:346: [info] Watermark loaded
2026/01/12 14:30:00 worker.go:346: [info] Connected to MQTT broker
2026/01/12 14:30:05 worker.go:346: [info] Connected to MongoDB
2026/01/12 14:30:05 worker.go:346: [info] Connected to SQL Server
```

Si los servicios están disponibles, debería conectar automáticamente.

#### 1.4 Insertar Datos de Prueba en SQL Server

```bash
# Opción A: SQL Management Studio (GUI)
# Conectar a: localhost,1433
# User: sa
# Password: AdminPassword123!
# Database: FTFeeding

# Opción B: SQLCMD (línea de comando)
sqlcmd -S localhost,1433 -U sa -P AdminPassword123!
```

**Script SQL:**
```sql
USE FTFeeding

-- Insertar evento de prueba
INSERT INTO [dbo].[DetalleAlimentacion] (
    ID, Name, UnitName, FechaHora, AmountGrams, FishCount, PesoProm, 
    Biomasa, FeedName, SiloName, Dia, Inicio, Fin, Dif, 
    PelletFishMin, PelletPK, GramsPerSec, KgTonMin, Marca, DoserName
) VALUES (
    'PROD-001', 
    'Centro-Principal', 
    'Jaula-100', 
    GETUTCDATE(), 
    250.5, 
    2000, 
    125.25, 
    250500, 
    'Premium 4.5mm', 
    'Silo-Principal', 
    CAST(GETUTCDATE() AS DATE),
    '08:00',
    '09:00',
    1,
    1.5,
    1.2,
    2.1,
    10.5,
    0,
    'Doser-Main'
)

-- Verificar
SELECT TOP 5 * FROM [dbo].[DetalleAlimentacion] ORDER BY FechaHora DESC
```

#### 1.5 Monitorear Logs en Tiempo Real

**Terminal 1: Backend**
```bash
cd f:\vscode\omnipoll\backend
.\omnipoll.exe
```

Buscar logs de polling:
```
2026/01/12 14:35:00 poller.go:82: Fetched 1 new records from Akva
2026/01/12 14:35:00 poller.go:95: Attempting to publish 1 changed events to MQTT
2026/01/12 14:35:00 poller.go:101: Published 1 changed events to MQTT
```

#### 1.6 Verificar JSON en MQTT

**Terminal 2: MQTT Subscriber**
```bash
mosquitto_sub -h mqtt.vmsfish.com -p 8883 \
  -t "feeding/mowi/+/" \
  -u test \
  -P test2025 \
  -v
```

**Output esperado:**
```
feeding/mowi/centroprincipal/ {"TimeStampAkva":"2026-01-12T14:35:00Z",...}
```

---

## 🟢 OPCIÓN 2: Docker Local Completo

### Caso de Uso
Tienes SQL Server remoto, quieres testear todo en Docker antes de producción.

### 2.1 Preparar Frontend Build

```bash
cd f:\vscode\omnipoll\frontend

# Instalar y buildar
npm install
npm run build

# Resultado: frontend/dist/ generado
```

### 2.2 Copiar dist al Backend

```bash
# Desde raíz del proyecto
mkdir -p backend/web
cp -r frontend/dist backend/web/dist

# Verificar
ls -la backend/web/dist/index.html
```

### 2.3 Actualizar config.yaml

Editar `backend/data/config.yaml`:

```yaml
sqlServer:
  host: 192.168.1.100        # IP real del SQL Server
  port: 1433
  database: FTFeeding
  user: sa
  password: 'AdminPassword123!'

mqtt:
  broker: mosquitto           # Nombre del servicio Docker
  port: 1883
  topic: feeding/mowi/
  clientId: omnipoll-production
  user: ''
  password: ''
  qos: 1

mongodb:
  uri: mongodb://mongodb:27017    # Nombre del servicio Docker
  database: omnipoll
  collection: historical_events

polling:
  intervalMs: 5000
  batchSize: 100

admin:
  host: 0.0.0.0              # Escuchar en todas las interfaces
  port: 8080
  username: admin
  password: 'admin'          # Cambiar después
```

### 2.4 Actualizar Dockerfile (si es necesario)

```dockerfile
# backend/Dockerfile - Verificar que sirve el frontend

FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o omnipoll ./cmd/omnipoll

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/omnipoll .
COPY data/ data/
COPY web/dist/ web/dist/           # ← Incluye frontend

EXPOSE 8080
CMD ["./omnipoll"]
```

### 2.5 Levantar Stack Completo

```bash
cd f:\vscode\omnipoll

# Iniciar todos los servicios
docker-compose up -d

# Verificar
docker ps
docker logs omnipoll
docker logs mongodb
docker logs mosquitto
```

### 2.6 Verificar Acceso

```bash
# Backend + Frontend (servido por backend)
http://localhost:8080

# API Status
curl -u admin:admin http://localhost:8080/api/status

# MongoDB
docker exec -it omnipoll_mongodb_1 mongosh
  > use omnipoll
  > db.historical_events.find().count()

# MQTT
docker logs omnipoll | grep -i mqtt
```

---

## 🟠 OPCIÓN 3: Servidor Linux Production

### Requisitos
- Servidor Linux (Ubuntu 20.04+)
- Docker + Docker Compose instalados
- Acceso SSH
- Dominio + SSL (opcional pero recomendado)

### 3.1 Preparar en Windows

```bash
# 1. Build frontend en Windows
cd frontend
npm run build

# 2. Copiar todo al backend
mkdir backend/web
cp -r frontend/dist backend/web/

# 3. Crear tarball para transferir
tar czf omnipoll.tar.gz \
  --exclude=node_modules \
  --exclude=.git \
  --exclude=dist \
  .

# O usar scp/rsync para transferir
scp omnipoll.tar.gz usuario@servidor:/home/usuario/
```

### 3.2 Configurar en Servidor Linux

```bash
ssh usuario@servidor

# Descomprimir
cd /home/usuario
tar xzf omnipoll.tar.gz
cd omnipoll

# Crear .env
cat > .env << EOF
OMNIPOLL_MASTER_KEY=tu-clave-secreta-aqui
SQL_SERVER_HOST=192.168.1.100
SQL_SERVER_USER=sa
SQL_SERVER_PASSWORD=tu-password
EOF

# Actualizar config.yaml
vi backend/data/config.yaml
# Cambiar credenciales SQL, MQTT, etc.
```

### 3.3 Levantar en Producción

```bash
# Iniciar stack completo
docker-compose up -d

# Verificar logs
docker logs -f omnipoll

# Verificar acceso
curl -u admin:admin http://localhost:8080/api/status
```

### 3.4 Configurar Nginx (Reverse Proxy + HTTPS)

```bash
# Instalar Nginx
sudo apt-get install nginx

# Crear config
sudo tee /etc/nginx/sites-available/omnipoll << EOF
server {
    listen 80;
    server_name omnipoll.tu-dominio.com;
    
    # Redirigir a HTTPS
    return 301 https://\$server_name\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name omnipoll.tu-dominio.com;
    
    ssl_certificate /etc/letsencrypt/live/omnipoll.tu-dominio.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/omnipoll.tu-dominio.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Habilitar
sudo ln -s /etc/nginx/sites-available/omnipoll /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# SSL con Let's Encrypt
sudo apt-get install certbot python3-certbot-nginx
sudo certbot certonly --nginx -d omnipoll.tu-dominio.com
```

### 3.5 Monitoreo

```bash
# Ver logs en tiempo real
docker logs -f omnipoll

# Monitoreo de recursos
docker stats omnipoll

# Backup de datos
docker run --rm -v omnipoll_data:/data \
  -v /backup:/backup \
  alpine tar czf /backup/omnipoll-$(date +%Y%m%d).tar.gz -C /data .
```

---

## 🔴 OPCIÓN 4: Cloud Deployment (AWS/Azure/GCP)

### Arquitectura Recomendada

```
┌─────────────────────────────────────────────────────┐
│  Internet                                            │
│  ↓ (HTTPS)                                          │
│  CloudFront / CDN                                    │
│  ↓                                                   │
│  ALB (Application Load Balancer)                    │
│  ↓                                                   │
│  ECS / Kubernetes Cluster                          │
│  ├─ Omnipoll Pod (replicas)                        │
│  └─ (Auto-scaling)                                 │
│  ↓                                                   │
│  RDS (SQL Server) + CloudSQL                       │
│  MongoDB Atlas (managed)                            │
│  EventBridge / IoT Hub (MQTT)                       │
│  CloudWatch / Datadog (monitoring)                  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Pasos AWS ECS (Ejemplo)

```bash
# 1. Build y push a ECR
aws ecr create-repository --repository-name omnipoll
docker build -t omnipoll:latest backend/
docker tag omnipoll:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/omnipoll:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/omnipoll:latest

# 2. Crear task definition en ECS
# Ver: aws/ecs-task-definition.json

# 3. Crear service en ECS
aws ecs create-service \
  --cluster omnipoll \
  --service-name omnipoll-api \
  --task-definition omnipoll:1 \
  --desired-count 2 \
  --load-balancers targetGroupArn=arn:aws:...,containerName=omnipoll,containerPort=8080

# 4. Configurar auto-scaling
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --resource-id service/omnipoll/omnipoll-api \
  --scalable-dimension ecs:service:DesiredCount \
  --min-capacity 2 \
  --max-capacity 10
```

---

## ✅ Checklist Pre-Producción

- [ ] Backend compila sin errores
- [ ] Frontend buildea correctamente
- [ ] Todas las conexiones funcionan (SQL, Mongo, MQTT)
- [ ] JSONs se publican correctamente a MQTT
- [ ] Dashboard muestra datos en tiempo real
- [ ] Logs están configurados y funcionan
- [ ] config.yaml tiene credenciales de producción
- [ ] Contraseña admin cambió de 'admin'
- [ ] Nginx/proxy reverso configurado
- [ ] SSL/HTTPS habilitado
- [ ] Backups configurados
- [ ] Monitoreo y alertas activas
- [ ] Load testing realizado (100+ eventos/segundo)
- [ ] Plan de rollback definido
- [ ] Documentación de operaciones actualizada

---

## 🚀 Opciones Recomendadas por Caso

### Para Testing Rápido (Esta Semana)
→ **OPCIÓN 1: Local Development**
- Lo que tienes ahora
- + Docker para SQL/Mongo
- 30 minutos para estar funcionando

### Para Pre-Producción (Este Mes)
→ **OPCIÓN 2: Docker Local Completo**
- Stack completo en Docker
- Simula producción sin infraestructura
- 1-2 horas de setup

### Para Producción Inmediata
→ **OPCIÓN 3: Linux Server + Nginx**
- Servidor dedicado o VM
- Simple de mantener
- Costo bajo
- 2-3 horas de setup

### Para Alta Escala
→ **OPCIÓN 4: Cloud (AWS/Azure/GCP)**
- Auto-scaling automático
- Managed services (no mantener)
- Más caro pero muy confiable
- 1 semana de setup

---

## 📊 Comparativa de Opciones

| Aspecto | Opción 1 | Opción 2 | Opción 3 | Opción 4 |
|---------|----------|----------|----------|----------|
| **Setup Time** | 30 min | 1-2 h | 2-3 h | 1 semana |
| **Costo** | $0 | $0 | $20-50/mes | $100-500/mes |
| **Escalabilidad** | No | Limitada | Manual | Automática |
| **HA/Redundancy** | No | No | Manual | Automática |
| **Monitoring** | Basic | Basic | Manual | Incluido |
| **SSL/HTTPS** | No | No | Sí | Sí |
| **Ideal Para** | Desarrollo | Pre-prod | Pequeña escala | Producción |

---

**Recomendación Actual:**

Empezar con **OPCIÓN 2 (Docker Local)** para validar todo el flujo con datos reales, luego pasar a **OPCIÓN 3 (Linux Server)** para producción si la infraestructura es simple, o **OPCIÓN 4 (Cloud)** si necesitas escalabilidad.

¿Cuál opción prefieres para empezar?
