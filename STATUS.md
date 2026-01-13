# Omnipoll - Estado Actual del Proyecto

## ✅ COMPLETADO

### Backend
- ✅ Todos los CRUDs implementados (Eventos, Configuración, Logs)
- ✅ Autenticación funcionando (admin/admin)
- ✅ Persistencia de configuración en YAML
- ✅ Validación y preservación de datos sensibles
- ✅ API REST funcional
- ✅ Compilación exitosa

### Frontend
- ✅ Dashboard con conexión a API
- ✅ Página de Eventos con paginación y filtros
- ✅ Página de Logs con búsqueda
- ✅ Página de Configuración con tabs para cada sección
- ✅ Interfaz responsiva con Tailwind CSS
- ✅ Desarrollo con Vite hot-reload

### Documentación
- ✅ ARCHITECTURE.md - Diagrama de la arquitectura
- ✅ CRUD_IMPLEMENTATION.md - Documentación de endpoints
- ✅ TESTING_GUIDE.md - Guía de pruebas
- ✅ FRONTEND_SETUP.md - Setup del frontend
- ✅ DEPLOY.md - Instrucciones de deploy con Docker
- ✅ README.md - Documentación principal

## ⚠️ LIMITACIONES CONOCIDAS

### Hot-Reload de Configuración (DESHABILITADO)
**Status:** Deshabilitado temporalmente debido a race condition

**Qué funciona:**
- Los cambios de configuración se guardan en `config.yaml`
- El frontend puede ver los nuevos valores al refrescar
- La API retorna la configuración actualizada

**Qué NO funciona:**
- El backend no reconecta automáticamente a MQTT/SQL Server cuando cambia la config
- Se requiere reiniciar el servidor para usar los nuevos parámetros de conexión

**Por qué está deshabilitado:**
- El intento de hot-reload tenía una race condition
- Cuando se intenta recargar la configuración, otras solicitudes pueden acceder a clientes en estado de cambio (nil)
- Resultaba en panics cuando se alcanzaban esos clientes

**Plan de Mejora:**
Implementar hot-reload con sincronización adecuada usando:
- `sync.atomic.Pointer` para cambios atómicos
- `sync.Cond` para coordinar requests in-flight
- Canales para esperar que las solicitudes actuales terminen
- O: Reconexión lazy al detectar conexión rota

## 🔄 FLUJO ACTUAL DE USO

### 1. Cambiar MQTT
1. Usuario va a Configuration → MQTT
2. Cambia broker, puerto, topic, etc.
3. Click en "Save Configuration"
4. ✅ Datos se guardan en config.yaml
5. ⚠️ Backend sigue conectado al broker anterior
6. ❌ **SOLUCIÓN ACTUAL:** Reiniciar el servidor

### 2. Cambiar SQL Server
Mismo flujo que MQTT.

### 3. Cambiar MongoDB
Mismo flujo que MQTT.

## 🚀 PARA PRODUCCIÓN

1. **Implementar hot-reload adecuadamente** (ver Plan de Mejora)
2. **Habilitar cifrado de contraseñas en config.yaml**
   - Actualmente deshabilitado en desarrollo
   - Descomentar en `backend/internal/config/loader.go` líneas 147-171
3. **Usar variables de entorno para credenciales**
4. **Agregar HTTPS/TLS**
5. **Agregar logs más detallados**
6. **Pruebas automatizadas**

## 📋 ÚLTIMOS CAMBIOS

### Commit más reciente
```
Fix: Disable hot-reload feature due to race condition
- Commented out ReloadConfig() method in worker.go
- Removed ReloadConfig call from config PUT handler
- Configuration changes persist but require restart
```

### Cambios anteriores clave
- Disabled encryption in development (passwords in plain text in dev)
- Fixed config response structure (backend returns config directly)
- Fixed frontend authentication (changed admin123 → admin)
- Improved password preservation in config updates
- Added MQTT test connection endpoint

## 🧪 TESTING

### Backend
```bash
cd backend
go build -o omnipoll.exe ./cmd/omnipoll
./omnipoll.exe
```

### Frontend
```bash
cd frontend
npm install
npm run dev
# Accede a http://localhost:3001
```

### Credenciales de Prueba
- **Usuario:** admin
- **Contraseña:** admin

## 📊 SERVICIOS EXTERNOS

**Actualmente en desarrollo sin servicios:**
- MongoDB: No disponible (requiere Docker)
- SQL Server: No disponible (requiere Docker)
- MQTT: Se intenta conectar a mqtt.vmsfish.com:8883

Para ejecutar con servicios, ver `DEPLOY.md` para instrucciones de Docker Compose.

## ✨ CARACTERÍSTICAS FUTURAS

1. Hot-reload de configuración con sincronización adecuada
2. Websokets para live updates de eventos/logs
3. Autenticación JWT en lugar de HTTP Basic Auth
4. Multi-usuario con roles (admin, user, readonly)
5. Dashboard con gráficos y estadísticas
6. Exportación de datos (CSV, JSON)
7. Webhooks para eventos críticos
8. Alertas en tiempo real

---

**Última actualización:** 2026-12-01
