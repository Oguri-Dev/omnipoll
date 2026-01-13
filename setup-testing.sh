#!/bin/bash
# Omnipoll Quick Setup Script for Testing
# Usage: ./setup-testing.sh

set -e

echo "=================================================="
echo "Omnipoll - Testing Setup Script"
echo "=================================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check Docker
echo -e "${YELLOW}[1/5]${NC} Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker not found. Please install Docker first.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker is installed${NC}"

# Check Docker Compose
echo -e "${YELLOW}[2/5]${NC} Checking Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose not found. Please install Docker Compose first.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker Compose is installed${NC}"

# Build Backend
echo -e "${YELLOW}[3/5]${NC} Building backend..."
cd backend
if command -v go &> /dev/null; then
    go build -o omnipoll ./cmd/omnipoll
    echo -e "${GREEN}✓ Backend built successfully${NC}"
else
    echo -e "${YELLOW}⚠ Go not installed. Backend will be built in Docker.${NC}"
fi
cd ..

# Start Services
echo -e "${YELLOW}[4/5]${NC} Starting Docker services..."
docker-compose up -d mongodb mosquitto

echo -e "${GREEN}✓ MongoDB and Mosquitto started${NC}"
echo ""

# Wait for services
echo -e "${YELLOW}[5/5]${NC} Waiting for services to be ready (30 seconds)..."
sleep 30

# Check services
echo ""
echo "=================================================="
echo "Service Status"
echo "=================================================="

if docker ps | grep -q mongodb; then
    echo -e "${GREEN}✓ MongoDB${NC} running on localhost:27017"
else
    echo -e "${RED}✗ MongoDB${NC} not running"
fi

if docker ps | grep -q mosquitto; then
    echo -e "${GREEN}✓ Mosquitto${NC} running on localhost:1883"
else
    echo -e "${RED}✗ Mosquitto${NC} not running"
fi

echo ""
echo "=================================================="
echo "Next Steps"
echo "=================================================="
echo ""
echo "1. Start the backend:"
echo -e "   ${YELLOW}cd backend${NC}"
echo -e "   ${YELLOW}./omnipoll.exe${NC} (Windows) or ${YELLOW}./omnipoll${NC} (Linux/Mac)"
echo ""
echo "2. Start the frontend (new terminal):"
echo -e "   ${YELLOW}cd frontend${NC}"
echo -e "   ${YELLOW}npm install${NC}"
echo -e "   ${YELLOW}npm run dev${NC}"
echo ""
echo "3. Insert test data into SQL Server (with SQL Server running):"
echo -e "   ${YELLOW}See TESTING_JSON.md for SQL scripts${NC}"
echo ""
echo "4. Monitor MQTT in another terminal:"
echo -e "   ${YELLOW}mosquitto_sub -h mqtt.vmsfish.com -p 8883 -t 'feeding/mowi/+/' -u test -P test2025 -v${NC}"
echo ""
echo "5. Access dashboard:"
echo -e "   ${YELLOW}http://localhost:3001${NC}"
echo -e "   ${YELLOW}API Status: http://localhost:8080/api/status${NC}"
echo ""
echo "=================================================="
echo "Configuration"
echo "=================================================="
echo ""
echo "Update backend/data/config.yaml with your credentials:"
echo "  - SQL Server host/user/password"
echo "  - MQTT broker (if using local Mosquitto)"
echo "  - MongoDB URI (if using Docker Mosquitto)"
echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
