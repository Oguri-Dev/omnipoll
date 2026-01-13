@echo off
REM Omnipoll Quick Setup Script for Windows Testing
REM Usage: setup-testing.bat

cls
echo ==================================================
echo Omnipoll - Testing Setup Script (Windows)
echo ==================================================
echo.

REM Check Docker
echo [1/5] Checking Docker installation...
docker --version >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker not found. Please install Docker Desktop first.
    pause
    exit /b 1
)
echo OK: Docker is installed
echo.

REM Check Docker Compose
echo [2/5] Checking Docker Compose...
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo WARNING: Docker Compose not found. Docker Desktop should include it.
)
echo OK: Docker Compose is available
echo.

REM Check Go
echo [3/5] Building backend...
go version >nul 2>&1
if errorlevel 1 (
    echo WARNING: Go not installed. Backend will be built in Docker.
) else (
    echo Building with Go...
    cd backend
    go build -o omnipoll.exe .\cmd\omnipoll
    if errorlevel 1 (
        echo ERROR: Backend build failed
        pause
        exit /b 1
    )
    cd ..
    echo OK: Backend built successfully
)
echo.

REM Start Services
echo [4/5] Starting Docker services...
docker-compose up -d mongodb mosquitto

if errorlevel 1 (
    echo ERROR: Docker Compose failed to start services
    pause
    exit /b 1
)

echo OK: MongoDB and Mosquitto starting...
echo.

REM Wait for services
echo [5/5] Waiting for services to be ready (30 seconds)...
timeout /t 30 /nobreak
echo.

REM Check services
echo ==================================================
echo Service Status
echo ==================================================
echo.

docker ps | find "mongodb" >nul
if errorlevel 0 (
    echo [OK] MongoDB running on localhost:27017
) else (
    echo [FAIL] MongoDB not running
)

docker ps | find "mosquitto" >nul
if errorlevel 0 (
    echo [OK] Mosquitto running on localhost:1883
) else (
    echo [FAIL] Mosquitto not running
)

echo.
echo ==================================================
echo Next Steps
echo ==================================================
echo.
echo 1. Start the backend (new Terminal or PowerShell):
echo    cd backend
echo    .\omnipoll.exe
echo.
echo 2. Start the frontend (new Terminal):
echo    cd frontend
echo    npm install
echo    npm run dev
echo.
echo 3. Insert test data into SQL Server:
echo    See TESTING_JSON.md for SQL scripts
echo    Use SQL Server Management Studio or sqlcmd
echo.
echo 4. Monitor MQTT (new Terminal):
echo    mosquitto_sub -h mqtt.vmsfish.com -p 8883 -t "feeding/mowi/+/" -u test -P test2025 -v
echo.
echo 5. Access dashboard:
echo    Frontend: http://localhost:3001
echo    API Status: http://localhost:8080/api/status
echo.
echo ==================================================
echo Configuration
echo ==================================================
echo.
echo Update backend\data\config.yaml with your credentials:
echo   - SQL Server host/user/password
echo   - MQTT broker settings
echo   - MongoDB URI
echo.
echo Setup complete!
echo.
pause
