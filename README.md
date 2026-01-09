# Omnipoll

Data ingestion agent for polling external SQL Server (Akva), normalizing data, publishing to MQTT, and persisting to MongoDB.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Omnipoll                            │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Admin Panel  │  │   Worker     │  │    Watermark     │  │
│  │   (HTTP)     │  │  (Poller)    │  │   Persistence    │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                     External Systems                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  SQL Server  │  │    MQTT      │  │     MongoDB      │  │
│  │   (Akva)     │  │  (Mosquitto) │  │   (Historical)   │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
Omnipoll/
├── backend/                 # Go backend
│   ├── cmd/omnipoll/       # Entry point
│   ├── internal/
│   │   ├── config/         # Configuration management
│   │   ├── poller/         # Polling worker + watermark
│   │   ├── akva/           # SQL Server client
│   │   ├── mqtt/           # MQTT publisher
│   │   ├── mongo/          # MongoDB repository
│   │   ├── admin/          # HTTP admin API
│   │   ├── crypto/         # Credential encryption
│   │   └── events/         # Event models
│   ├── configs/            # Example configs
│   ├── scripts/            # Build scripts
│   ├── web/                # Frontend build output
│   ├── Dockerfile
│   └── go.mod
├── frontend/               # React admin panel
│   ├── src/
│   │   ├── components/     # UI components
│   │   ├── pages/          # Route pages
│   │   ├── services/       # API client
│   │   └── types/          # TypeScript types
│   ├── package.json
│   └── vite.config.ts
├── mosquitto/              # Mosquitto config
├── docker-compose.yml
└── README.md
```

## Development

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/omnipoll
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Docker

```bash
docker-compose up -d
```

## Configuration

See `backend/configs/config.example.yaml` for configuration options.

Environment variables:

- `OMNIPOLL_MASTER_KEY`: Master key for credential encryption
- `OMNIPOLL_CONFIG_PATH`: Path to config file
- `OMNIPOLL_WATERMARK_PATH`: Path to watermark file

## License

Private
