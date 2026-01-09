#!/bin/bash
set -e

echo "🚀 Omnipoll - Script de Despliegue en Docker"
echo "============================================"
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Verificar Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker no está instalado${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose no está instalado${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker y Docker Compose detectados${NC}"
echo ""

# Verificar .env
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  Archivo .env no encontrado. Creando plantilla...${NC}"
    cat > .env << 'EOF'
# Master key para encriptar credenciales (mínimo 32 caracteres)
OMNIPOLL_MASTER_KEY=change-this-to-a-secure-random-key-32chars

# Configuración SQL Server (Akva)
SQL_SERVER_HOST=host.docker.internal
SQL_SERVER_PORT=1433
SQL_SERVER_DATABASE=FTFeeding
SQL_SERVER_USER=sa
SQL_SERVER_PASSWORD=change-me
EOF
    echo -e "${YELLOW}📝 Por favor edita el archivo .env con tus credenciales reales${NC}"
    echo -e "${YELLOW}   nano .env${NC}"
    exit 0
fi

echo -e "${GREEN}✅ Archivo .env encontrado${NC}"

# Verificar frontend build
if [ ! -d "frontend/dist" ]; then
    echo -e "${YELLOW}📦 Frontend no está construido. Construyendo...${NC}"
    cd frontend
    if [ ! -d "node_modules" ]; then
        echo "   Instalando dependencias..."
        npm install
    fi
    echo "   Construyendo frontend..."
    npm run build
    cd ..
    echo -e "${GREEN}✅ Frontend construido${NC}"
else
    echo -e "${GREEN}✅ Frontend ya está construido${NC}"
fi

# Copiar frontend al backend
echo "📁 Copiando frontend build al backend..."
mkdir -p backend/web
cp -r frontend/dist backend/web/
echo -e "${GREEN}✅ Frontend copiado${NC}"
echo ""

# Verificar configuración
if [ ! -f "backend/data/config.yaml" ]; then
    echo -e "${YELLOW}⚠️  config.yaml no encontrado. Creando configuración por defecto...${NC}"
    mkdir -p backend/data
    cat > backend/data/config.yaml << 'EOF'
sqlServer:
  host: host.docker.internal
  port: 1433
  database: FTFeeding
  user: sa
  password: ""
mqtt:
  broker: mosquitto
  port: 1883
  topic: ftfeeding/akva/detalle
  clientId: omnipoll-worker
  user: ""
  password: ""
  qos: 1
mongodb:
  uri: mongodb://mongodb:27017
  database: omnipoll
  collection: historical_events
polling:
  intervalMs: 5000
  batchSize: 100
admin:
  host: 0.0.0.0
  port: 8080
  username: admin
  password: "admin123"
EOF
    echo -e "${YELLOW}📝 Por favor edita backend/data/config.yaml con tus credenciales${NC}"
fi

echo -e "${GREEN}✅ Configuración lista${NC}"
echo ""

# Construir imágenes
echo "🔨 Construyendo imágenes Docker..."
docker-compose build

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error al construir las imágenes${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Imágenes construidas exitosamente${NC}"
echo ""

# Levantar servicios
echo "🚀 Levantando servicios..."
docker-compose up -d

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error al levantar los servicios${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Servicios iniciados${NC}"
echo ""

# Esperar a que los servicios estén listos
echo "⏳ Esperando que los servicios estén listos..."
sleep 5

# Verificar estado
echo ""
echo "📊 Estado de los servicios:"
docker-compose ps
echo ""

# Obtener IP del servidor
SERVER_IP=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}✅ Despliegue completado exitosamente!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo "🌐 Admin Panel disponible en:"
echo -e "   ${GREEN}http://localhost:8080${NC}"
echo -e "   ${GREEN}http://${SERVER_IP}:8080${NC}"
echo ""
echo "📊 MQTT Broker:"
echo -e "   Broker: ${GREEN}${SERVER_IP}:1883${NC}"
echo -e "   Topic: ${GREEN}ftfeeding/akva/detalle${NC}"
echo ""
echo "🗄️  MongoDB:"
echo -e "   URI: ${GREEN}mongodb://${SERVER_IP}:27017${NC}"
echo -e "   Database: ${GREEN}omnipoll${NC}"
echo ""
echo "📝 Comandos útiles:"
echo "   Ver logs:        docker-compose logs -f omnipoll"
echo "   Detener:         docker-compose down"
echo "   Reiniciar:       docker-compose restart omnipoll"
echo "   Ver estado:      docker-compose ps"
echo ""
echo "🔐 Credenciales por defecto:"
echo "   Usuario: admin"
echo "   Contraseña: admin123 (cambiar en producción)"
echo ""
echo -e "${YELLOW}⚠️  Recuerda configurar las credenciales reales de SQL Server en:${NC}"
echo "   backend/data/config.yaml"
echo ""

# Mostrar logs iniciales
echo "📋 Logs iniciales (Ctrl+C para salir):"
echo ""
docker-compose logs -f omnipoll
