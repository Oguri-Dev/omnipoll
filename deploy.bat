@echo off
REM Omnipoll - Script de Despliegue en Docker para Windows
REM Usage: .\deploy.bat

cls
echo.
echo 🚀 Omnipoll - Script de Despliegue en Docker
echo ============================================
echo.

setlocal enabledelayedexpansion

REM Colores simulados con títulos
REM Green = OK messages
REM Red = Errors  
REM Yellow = Warnings

REM Verificar Docker
echo Verificando Docker...
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker no está instalado
    pause
    exit /b 1
)
echo ✅ Docker detectado
echo.

REM Verificar Docker Compose
echo Verificando Docker Compose...
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose no está instalado
    pause
    exit /b 1
)
echo ✅ Docker Compose detectado
echo.

REM Verificar .env
if not exist .env (
    echo ⚠️  Archivo .env no encontrado. Creando plantilla...
    (
        echo # Master key para encriptar credenciales (minimo 32 caracteres^)
        echo OMNIPOLL_MASTER_KEY=change-this-to-a-secure-random-key-32chars
        echo.
        echo # Configuracion SQL Server (Akva^)
        echo SQL_SERVER_HOST=host.docker.internal
        echo SQL_SERVER_PORT=1433
        echo SQL_SERVER_DATABASE=FTFeeding
        echo SQL_SERVER_USER=sa
        echo SQL_SERVER_PASSWORD=change-me
    ) > .env
    
    echo 📝 Por favor edita el archivo .env con tus credenciales reales
    echo.
    echo Abriendo .env en editor...
    notepad .env
    pause
    exit /b 0
)

echo ✅ Archivo .env encontrado
echo.

REM Verificar frontend build
if not exist "frontend\dist" (
    echo 📦 Frontend no está construido. Construyendo...
    cd frontend
    
    if not exist "node_modules" (
        echo    Instalando dependencias...
        call npm install
    )
    
    echo    Construyendo frontend...
    call npm run build
    
    if errorlevel 1 (
        echo ❌ Error al construir frontend
        cd ..
        pause
        exit /b 1
    )
    
    cd ..
    echo ✅ Frontend construido
) else (
    echo ✅ Frontend ya está construido
)
echo.

REM Copiar frontend al backend
echo Copiando frontend build al backend...
if not exist "backend\web" mkdir backend\web
xcopy /E /I /Y frontend\dist backend\web\dist >nul
echo ✅ Frontend copiado
echo.

REM Verificar configuración
if not exist "backend\data\config.yaml" (
    echo ⚠️  config.yaml no encontrado. Creando configuracion por defecto...
    if not exist "backend\data" mkdir backend\data
    
    (
        echo sqlServer:
        echo   host: host.docker.internal
        echo   port: 1433
        echo   database: FTFeeding
        echo   user: sa
        echo   password: ""
        echo mqtt:
        echo   broker: mosquitto
        echo   port: 1883
        echo   topic: ftfeeding/akva/detalle
        echo   clientId: omnipoll-worker
        echo   user: ""
        echo   password: ""
        echo   qos: 1
        echo mongodb:
        echo   uri: mongodb://mongodb:27017
        echo   database: omnipoll
        echo   collection: historical_events
        echo polling:
        echo   intervalMs: 5000
        echo   batchSize: 100
        echo admin:
        echo   host: 0.0.0.0
        echo   port: 8080
        echo   username: admin
        echo   password: "admin123"
    ) > backend\data\config.yaml
    
    echo 📝 Por favor edita backend\data\config.yaml con tus credenciales
)

echo ✅ Configuracion lista
echo.

REM Construir imágenes
echo Construyendo imágenes Docker...
docker-compose build

if errorlevel 1 (
    echo ❌ Error al construir las imágenes
    pause
    exit /b 1
)

echo ✅ Imágenes construidas exitosamente
echo.

REM Levantar servicios
echo Levantando servicios...
docker-compose up -d

if errorlevel 1 (
    echo ❌ Error al levantar los servicios
    pause
    exit /b 1
)

echo ✅ Servicios iniciados
echo.

REM Esperar a que los servicios estén listos
echo Esperando que los servicios estén listos (5 segundos)...
timeout /t 5 /nobreak
echo.

REM Verificar estado
echo Estado de los servicios:
docker-compose ps
echo.

REM Mostrar información
echo.
echo ============================================
echo ✅ Despliegue completado exitosamente!
echo ============================================
echo.
echo 🌐 Admin Panel disponible en:
echo    http://localhost:8080
echo.
echo 📊 MQTT Broker:
echo    Broker: localhost:1883
echo    Topic: ftfeeding/akva/detalle
echo.
echo 🗄️  MongoDB:
echo    URI: mongodb://localhost:27017
echo    Database: omnipoll
echo.
echo 📝 Comandos útiles:
echo    Ver logs:        docker-compose logs -f omnipoll
echo    Detener:         docker-compose down
echo    Reiniciar:       docker-compose restart omnipoll
echo    Ver estado:      docker-compose ps
echo.
echo 🔐 Credenciales por defecto:
echo    Usuario: admin
echo    Contraseña: admin123 ^(cambiar en producción^)
echo.
echo ⚠️  Recuerda configurar las credenciales reales de SQL Server en:
echo    backend\data\config.yaml
echo.
echo Mostrando logs iniciales (Ctrl+C para salir)...
echo.
docker-compose logs -f omnipoll

pause
