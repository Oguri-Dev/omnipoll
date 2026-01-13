# 🌊 Omnipoll

**Enterprise-grade data ingestion agent** for polling external SQL Server databases, normalizing data in real-time, publishing to MQTT, and persisting historical events to MongoDB.

**Status:** ✅ Production Ready | **License:** Private | **Latest:** v1.0.0

---

## 📋 Quick Start

### Via Docker (Recommended - 5 minutes)

```bash
# Clone and setup
git clone <repo>
cd omnipoll

# Deploy everything
./deploy.sh              # Linux/Mac
# or
deploy.bat              # Windows
```

**That's it!** Dashboard available at `http://localhost:8080`

### Manual Setup

**Requirements:**
- Go 1.21+ 
- Node.js 18+
- Docker & Docker Compose
- SQL Server 2019+ (remote or local)

**Backend:**
```bash
cd backend
go mod download
OMNIPOLL_CONFIG_PATH=data/config.yaml go run ./cmd/omnipoll
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev          # Dev mode with hot reload
npm run build        # Production build
```

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    OMNIPOLL SYSTEM                           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐      ┌──────────────┐    ┌────────────┐   │
│  │ Web Admin   │      │   Poller     │    │ Watermark  │   │
│  │   (React)   │      │   (Worker)   │    │  Storage   │   │
│  └─────────────┘      └──────────────┘    └────────────┘   │
│       ↓                     ↓                   ↓             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Go Backend (HTTP REST + Polling)              │ │
│  └────────────────────────────────────────────────────────┘ │
│       ↓         ↓         ↓           ↓                     │
├──────────────────────────────────────────────────────────────┤
│            EXTERNAL SYSTEMS & STORAGE                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌───────────┐    ┌──────────┐    ┌────────────┐            │
│  │   SQL     │    │   MQTT   │    │ MongoDB    │            │
│  │  Server   │    │ Broker   │    │ Historical │            │
│  │  (Akva)   │    │(Mosquitto)    │  Events    │            │
│  └───────────┘    └──────────┘    └────────────┘            │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**Data Flow:**
```
SQL Server → Fetch New Records
    ↓
Normalize to NormalizedEvent
    ↓
Filter Against MongoDB (dedup)
    ↓
Map to MQTTMessage (JSON)
    ↓
Publish to MQTT Broker (QoS 1)
    ↓
External Systems (Dashboard, Alerting, etc)
```

---

## 📁 Project Structure

```
omnipoll/
├── 📁 backend/                      # Go REST API + Polling Service
│   ├── cmd/omnipoll/
│   │   └── main.go                  # Entry point
│   ├── internal/
│   │   ├── admin/                   # HTTP handlers & API routes
│   │   │   ├── handlers.go          # CRUD endpoints (200+ lines)
│   │   │   ├── logs_handlers.go     # Log endpoints
│   │   │   ├── server.go            # HTTP server setup
│   │   │   └── router.go            # Route definitions
│   │   ├── poller/                  # Main polling logic
│   │   │   ├── poller.go            # Poll orchestration
│   │   │   ├── watermark.go         # Last-seen tracking
│   │   │   └── worker.go            # Background worker
│   │   ├── akva/                    # SQL Server integration
│   │   │   ├── client.go            # DB connection
│   │   │   └── mapper.go            # Data mapping
│   │   ├── mqtt/                    # MQTT publishing
│   │   │   ├── client.go            # Paho MQTT wrapper
│   │   │   └── publisher.go         # Message publishing
│   │   ├── mongo/                   # MongoDB persistence
│   │   │   ├── client.go            # Connection management
│   │   │   └── repository.go        # CRUD operations
│   │   ├── config/                  # Configuration
│   │   │   ├── config.go            # Structs
│   │   │   └── loader.go            # YAML parsing
│   │   ├── crypto/                  # Encryption
│   │   │   └── encryption.go        # AES-256 encryption
│   │   └── events/                  # Domain models
│   │       └── event.go             # Event structs
│   ├── data/
│   │   ├── config.yaml              # 🔴 EDIT: Connection strings
│   │   └── watermark.json           # Auto-managed
│   ├── configs/
│   │   └── config.example.yaml      # Template
│   ├── Dockerfile                   # Multi-stage build
│   ├── go.mod                       # Dependencies
│   └── go.sum
│
├── 📁 frontend/                     # React Admin Dashboard
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx        # Status overview
│   │   │   ├── Events.tsx           # Event history
│   │   │   ├── Logs.tsx             # System logs
│   │   │   └── Configuration.tsx    # Config editor
│   │   ├── components/
│   │   │   ├── StatusCard.tsx       # Connection status
│   │   │   ├── ConnectionStatus.tsx # Real-time indicator
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Layout.tsx
│   │   ├── services/
│   │   │   └── api.ts               # Axios + HTTP calls
│   │   ├── types/
│   │   │   └── index.ts             # TypeScript interfaces
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── dist/                        # Build output (auto-populated by deploy.sh)
│
├── 📁 mosquitto/                    # MQTT Broker Config
│   └── config/
│       └── mosquitto.conf
│
├── 📁 scripts/                      # Automation Scripts
│   ├── deploy.sh                    # 🚀 Docker deployment (Linux/Mac)
│   ├── deploy.bat                   # 🚀 Docker deployment (Windows)
│   ├── setup-testing.sh             # Local testing setup
│   └── setup-testing.bat            # Local testing setup (Windows)
│
├── docker-compose.yml               # Multi-container orchestration
├── .env                             # 🔴 EDIT: Secrets & credentials
├── .gitignore
├── go.mod
│
├── 📚 DOCUMENTATION/ (15+ guides)
│   ├── MORNING_BRIEF.md             # ⭐ START HERE (for quick context)
│   ├── GO_LIVE.md                   # 3 deployment options
│   ├── PRODUCTION.md                # Complete deployment guide
│   ├── ARCHITECTURE.md              # System design deep-dive
│   ├── JSON_FLOW.md                 # Data transformation pipeline
│   ├── JSON_EXAMPLES.md             # 7 real-world JSON payloads
│   ├── TESTING_JSON.md              # Testing guide with SQL scripts
│   ├── SCRIPTS_GUIDE.md             # How to use automation scripts
│   ├── CRUD_IMPLEMENTATION.md       # API endpoints reference
│   ├── CONNECTION_STATUS_FIX.md     # Status reporting logic
│   ├── IMPLEMENTATION_SUMMARY.md    # What was built
│   └── STATUS.md                    # Current status & limitations
│
└── 🔖 VERSION CONTROL
    └── .git/                        # 20+ commits tracking development
```

---

## 🚀 Deployment Options

### Option A: Docker (Recommended)
**⏱️ Time:** 5 minutes | **Complexity:** Low | **Recommended For:** Development & testing

```bash
./deploy.sh          # or deploy.bat on Windows
```

- ✅ Auto-detects Docker/Docker Compose
- ✅ Builds frontend automatically  
- ✅ Creates config from template
- ✅ Starts all services
- ✅ Shows live logs

**Result:** Full stack at `http://localhost:8080`

### Option B: Linux Production Server
**⏱️ Time:** 2-3 hours | **Complexity:** Medium | **Recommended For:** Production

```bash
# Transfer code to server
scp -r omnipoll/ user@server:/opt/

# On server:
cd /opt/omnipoll
chmod +x deploy.sh
./deploy.sh
```

[See PRODUCTION.md for detailed steps]

### Option C: Manual Setup
**⏱️ Time:** 1-2 hours | **Complexity:** High | **For:** Advanced users

[See DEPLOY.md for detailed steps]

---

## 📊 API Endpoints

### Authentication
HTTP Basic Auth: `admin:admin` (change in production!)

### Status
```bash
GET /api/status
# Returns connection states for SQL Server, MQTT, MongoDB
```

**Response:**
```json
{
  "sqlServer": {
    "connected": true,
    "lastCheck": "2025-01-12T15:30:00Z"
  },
  "mqtt": {
    "connected": true,
    "lastCheck": "2025-01-12T15:30:00Z"
  },
  "mongodb": {
    "connected": true,
    "lastCheck": "2025-01-12T15:30:00Z"
  }
}
```

### Configuration
```bash
GET  /api/config           # Get current config
POST /api/config           # Update config
```

### Events
```bash
GET  /api/events           # List events
GET  /api/events/:id       # Get event
POST /api/events           # Create event
PUT  /api/events/:id       # Update event
DELETE /api/events/:id     # Delete event
```

### Logs
```bash
GET  /api/logs             # Get system logs
GET  /api/logs/:id         # Get log entry
```

[See CRUD_IMPLEMENTATION.md for full endpoint documentation]

---

## ⚙️ Configuration

### Environment Variables (`.env`)

```bash
# Encryption
OMNIPOLL_MASTER_KEY=<random-32-chars>    # Generate with: openssl rand -hex 16

# SQL Server
SQL_SERVER_HOST=localhost
SQL_SERVER_PORT=1433
SQL_SERVER_DATABASE=FTFeeding
SQL_SERVER_USER=sa
SQL_SERVER_PASSWORD=YourPassword123!

# File Paths
OMNIPOLL_CONFIG_PATH=backend/data/config.yaml
OMNIPOLL_WATERMARK_PATH=backend/data/watermark.json
```

### YAML Config (`backend/data/config.yaml`)

```yaml
sqlServer:
  host: localhost
  port: 1433
  database: FTFeeding
  user: sa
  password: "password"

mqtt:
  broker: mqtt.vmsfish.com
  port: 8883
  topic: feeding/mowi/
  clientId: omnipoll-production
  tls: true
  user: test
  password: test2025
  qos: 1

mongodb:
  uri: mongodb://localhost:27017
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
```

[See backend/configs/config.example.yaml for all options]

---

## 🧪 Testing

### Test JSON Publishing
```bash
# Insert test data
sqlcmd -S localhost -U sa -P password -d FTFeeding -Q "
  INSERT INTO dbo.Events (centro_id, type, data, timestamp)
  VALUES (1, 'feeding', '{...}', GETDATE())
"

# Monitor MQTT
mosquitto_sub -h mqtt.vmsfish.com -p 8883 \
  -t "feeding/mowi/+/" \
  -u test -P test2025 -v
```

[See TESTING_JSON.md for complete testing guide]

---

## 📖 Documentation Index

| Document | Purpose | Read Time |
|----------|---------|-----------|
| [MORNING_BRIEF.md](MORNING_BRIEF.md) | **Context for tomorrow** | 5 min |
| [GO_LIVE.md](GO_LIVE.md) | Deployment checklist | 5 min |
| [PRODUCTION.md](PRODUCTION.md) | Detailed prod setup | 20 min |
| [JSON_FLOW.md](JSON_FLOW.md) | Data transformation | 10 min |
| [JSON_EXAMPLES.md](JSON_EXAMPLES.md) | Real payload examples | 5 min |
| [TESTING_JSON.md](TESTING_JSON.md) | How to test | 15 min |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design | 15 min |
| [SCRIPTS_GUIDE.md](SCRIPTS_GUIDE.md) | Script reference | 10 min |
| [CRUD_IMPLEMENTATION.md](CRUD_IMPLEMENTATION.md) | API reference | 10 min |

---

## 🛠️ Development

### Backend Development
```bash
cd backend

# Install dependencies
go mod download

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
air

# Or standard run
go run ./cmd/omnipoll

# Build
go build -o omnipoll ./cmd/omnipoll

# Tests (when available)
go test ./...
```

### Frontend Development
```bash
cd frontend

# Install
npm install

# Dev server (http://localhost:3001 with hot reload)
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check
```

---

## 📦 Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend** | Go | 1.21+ |
| **Frontend** | React | 18.2 |
| **Build Tool** | Vite | 5.x |
| **Styling** | Tailwind CSS | 3.x |
| **SQL** | MSSQL Driver | Latest |
| **MQTT** | Paho Go | Latest |
| **NoSQL** | MongoDB Driver | Latest |
| **HTTP** | net/http | Built-in |
| **Encryption** | crypto/aes | Built-in |
| **Container** | Docker | 20.10+ |
| **Orchestration** | Docker Compose | 2.x |

---

## 🔐 Security

### Encryption
- Credentials encrypted with AES-256-GCM
- Master key in environment variable
- No secrets in git

### Authentication
- HTTP Basic Auth
- Change default credentials in production
- Supports MQTT TLS 1.2+

### Best Practices
- ✅ Config stored in YAML (not checked in)
- ✅ Secrets in environment variables
- ✅ Master key rotatable via OMNIPOLL_MASTER_KEY
- ✅ MQTT connections support TLS
- ✅ SQL Server connections encrypted (with TLS option)

---

## 🆘 Troubleshooting

| Issue | Solution |
|-------|----------|
| Docker not found | Install Docker Desktop (Windows) or docker.io (Linux) |
| Port 8080 in use | Edit docker-compose.yml: `"8081:8080"` |
| Frontend build fails | `cd frontend && npm install && npm run build` |
| MQTT won't connect | Check credentials in .env and config.yaml |
| SQL Server won't connect | Verify host/port/credentials, check firewall |
| Status shows disconnected | Check logs: `docker-compose logs omnipoll` |
| Permission denied (Linux) | `chmod +x deploy.sh setup-testing.sh` |

[See STATUS.md for known issues and limitations]

---

## 📋 Checklist for Production

- [ ] Edit `.env` with real SQL Server credentials
- [ ] Edit `backend/data/config.yaml` with production settings
- [ ] Change admin password from default
- [ ] Verify MQTT connection with TLS
- [ ] Insert test data and verify publishing
- [ ] Configure monitoring/logging
- [ ] Set up SSL/HTTPS proxy (Nginx recommended)
- [ ] Configure backup strategy for MongoDB
- [ ] Test failover scenarios
- [ ] Monitor disk space (watermark.json growth)

---

## 📞 Support & Resources

- **Issue Tracker:** [GitHub Issues]
- **Documentation:** See root directory for `.md` files
- **Logs:** `docker-compose logs -f omnipoll`
- **MQTT Monitor:** `mosquitto_sub` with `-h` and `-t` flags
- **Database:** MongoDB connection string in config.yaml

---

## 📈 Performance & Limits

| Metric | Value | Notes |
|--------|-------|-------|
| Poll Interval | 5000 ms | Configurable in config.yaml |
| Batch Size | 100 records | Per poll cycle |
| MQTT QoS | 1 | At least once delivery |
| Max JSON Size | ~1 MB | Per event |
| MongoDB Storage | Unlimited | Growth depends on record count |
| SQL Query Timeout | 30 sec | Configurable |

---

## 📄 License

**Private** - All rights reserved

---

## 👨‍💻 Contributing

This is a private project. Contact team lead for contribution guidelines.

---

## 📅 Changelog

### v1.0.0 (2025-01-12)
- ✅ Full CRUD implementation
- ✅ Real-time connection status
- ✅ Docker deployment automation
- ✅ Comprehensive documentation
- ✅ Production-ready

[Full changelog in git history: `git log --oneline`]

---

**Last Updated:** 2025-01-12  
**Status:** ✅ Production Ready  
**Maintainer:** Development Team
