# Scripts Disponibles en Omnipoll

## 🚀 Scripts de Deploy (RECOMENDADO)

### `deploy.sh` (Linux/Mac) o `deploy.bat` (Windows)

**¿Qué hace?**

1. ✅ Verifica que Docker está instalado
2. ✅ Crea archivo `.env` si no existe
3. ✅ Build frontend automáticamente (npm install + npm run build)
4. ✅ Copia frontend build al backend (`backend/web/dist/`)
5. ✅ Crea `config.yaml` por defecto si no existe
6. ✅ Construye imágenes Docker (omnipoll, mongodb, mosquitto)
7. ✅ Levanta todos los servicios (`docker-compose up -d`)
8. ✅ Muestra logs en tiempo real

**Uso:**

```bash
# Windows
deploy.bat

# Linux/Mac
./deploy.sh
```

**Resultado: Stack completo en 5 minutos** ⚡

---

## 🧪 Scripts de Testing

### `setup-testing.sh` (Linux/Mac) o `setup-testing.bat` (Windows)

**¿Qué hace?**

1. ✅ Verifica Docker
2. ✅ Levanta MongoDB + Mosquitto
3. ✅ Espera a que servicios estén listos
4. ✅ Muestra instrucciones para siguiente paso

**Uso:**

```bash
# Windows
setup-testing.bat

# Linux/Mac
./setup-testing.sh
```

**Para qué sirve?**

- Testing local sin build de frontend
- Verificar backend conecta a servicios
- Insertar datos SQL y ver flujo MQTT

---

## 🔨 Scripts de Backend

### `backend/scripts/build.sh` (Linux/Mac)

**¿Qué hace?**

1. ✅ Compila backend con Go
2. ✅ Genera binario ejecutable

**Uso:**

```bash
cd backend/scripts
./build.sh
```

---

## 📊 Comparativa de Scripts

| Script                | Sistema   | Tiempo | Para Qué        | Comando                    |
| --------------------- | --------- | ------ | --------------- | -------------------------- |
| **deploy.sh**         | Linux/Mac | 5 min  | Deploy completo | `./deploy.sh`              |
| **deploy.bat**        | Windows   | 5 min  | Deploy completo | `deploy.bat`               |
| **setup-testing.sh**  | Linux/Mac | 2 min  | Testing local   | `./setup-testing.sh`       |
| **setup-testing.bat** | Windows   | 2 min  | Testing local   | `setup-testing.bat`        |
| **build.sh**          | Linux/Mac | 30 seg | Build backend   | `backend/scripts/build.sh` |

---

## 🎯 Flujo Recomendado

### Día 1: Validación Rápida

```
1. Ejecutar setup-testing.sh/bat
2. Insertar datos SQL
3. Verificar datos en MQTT
```

### Día 2: Deploy Completo

```
1. Ejecutar deploy.sh/bat
2. Verificar servicios corriendo
3. Testing exhaustivo
```

### Día 3: Producción

```
1. Transferir a servidor Linux
2. Ejecutar deploy.sh en servidor
3. Configurar Nginx + SSL
4. Go live
```

---

## ⚙️ Configuración Pre-Deploy

### Editar antes de ejecutar `deploy.sh`:

**`.env`** (Credenciales de SQL Server)

```bash
OMNIPOLL_MASTER_KEY=tu-clave-de-32-caracteres
SQL_SERVER_HOST=ip-de-tu-sql-server
SQL_SERVER_USER=sa
SQL_SERVER_PASSWORD=tu-password
```

**`backend/data/config.yaml`** (Conexiones)

```yaml
sqlServer:
  host: tu-servidor-sql
  user: sa
  password: tu-password

mqtt:
  broker: mosquitto # o tu-servidor-mqtt
  port: 1883

admin:
  password: cambiar-en-produccion
```

---

## 🆘 Troubleshooting Scripts

### "Docker no encontrado"

```bash
# Instalar Docker
# Windows: https://www.docker.com/products/docker-desktop
# Linux: curl -fsSL https://get.docker.com | sudo sh
```

### "Error en build frontend"

```bash
# El script intenta instalar npm automáticamente
# Si falla, instalar Node.js manualmente
node --version   # debe ser v16+
npm --version    # debe ser v8+
```

### "Permisos denegados en Linux"

```bash
# Hacer script ejecutable
chmod +x deploy.sh
chmod +x setup-testing.sh
```

### "Puerto 8080 ya en uso"

```bash
# Cambiar puerto en docker-compose.yml
# Buscar "8080:8080" y cambiar a "8081:8080"
```

---

## 📝 Ejemplos de Uso

### Caso 1: Deploy rápido en Windows

```
1. Abrir PowerShell / CMD
2. cd f:\vscode\omnipoll
3. .\deploy.bat
4. Esperar 5 minutos
5. Acceder a http://localhost:8080
```

### Caso 2: Testing local en Linux

```
1. cd ~/omnipoll
2. ./setup-testing.sh
3. Insertar datos SQL
4. Verificar en MQTT
```

### Caso 3: Deploy en servidor Linux

```
1. SSH al servidor: ssh usuario@servidor
2. cd /home/usuario/omnipoll
3. ./deploy.sh
4. Configurar Nginx (opcional)
5. Listo en producción
```

---

## 🔒 Seguridad

**Antes de ejecutar en producción:**

1. ✅ Editar `.env` con credenciales reales
2. ✅ Editar `config.yaml` con credenciales de SQL Server
3. ✅ Cambiar contraseña admin en `config.yaml`
4. ✅ Habilitar encriptación en `loader.go` (descomentar)
5. ✅ Configurar SSL/HTTPS en Nginx

---

## 📚 Documentación Relacionada

- **GO_LIVE.md** - Guía rápida de opciones
- **PRODUCTION.md** - Guía completa de deployment
- **TESTING_JSON.md** - Testing con datos reales
- **DEPLOY.md** - Setup manual detallado

---

## ✨ Características de los Scripts

- ✅ Manejo de errores
- ✅ Colores en output (en Linux/Mac)
- ✅ Logs informativos
- ✅ Automatizan tareas repetitivas
- ✅ Reducen chances de errores manuales
- ✅ Funcionan en desarrollo y producción
- ✅ Documentados internamente (comentarios)

---

**Última actualización:** 2026-01-12  
**Estado:** ✅ Todos los scripts testeados y funcionando
