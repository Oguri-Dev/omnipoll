# Frontend Setup Guide

## Prerequisites

- Node.js 18+ (download from nodejs.org)
- npm (comes with Node.js)
- Backend Omnipoll running on `http://localhost:8080`

## Installation

### Windows (with PowerShell)

```powershell
cd frontend

# If npm is not in PATH, use full path:
& "C:\Program Files\nodejs\npm.cmd" install
& "C:\Program Files\nodejs\npm.cmd" run build
```

### Linux/Mac

```bash
cd frontend
npm install
npm run build
```

## Running Development Server

```bash
cd frontend
npm run dev
```

The development server will start at `http://localhost:5173` (or another available port).

## Available Scripts

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm run preview` - Preview production build locally
- `npm run lint` - Run ESLint to check code quality

## Project Structure

```
frontend/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── ConnectionStatus.tsx
│   │   ├── Header.tsx
│   │   ├── Layout.tsx
│   │   ├── Sidebar.tsx
│   │   └── StatusCard.tsx
│   ├── pages/               # Page components
│   │   ├── Dashboard.tsx   # System status and worker controls
│   │   ├── Events.tsx      # Event management (CRUD)
│   │   ├── Configuration.tsx # Configuration management
│   │   └── Logs.tsx        # Log viewer with filtering
│   ├── services/            # API client
│   │   └── api.ts          # API communication
│   ├── types/               # TypeScript type definitions
│   │   └── index.ts
│   ├── App.tsx             # Main app component
│   ├── main.tsx            # Entry point
│   └── index.css            # Global styles
├── public/                  # Static assets
├── package.json             # Dependencies
├── vite.config.ts          # Vite configuration
├── tailwind.config.js      # Tailwind CSS configuration
└── tsconfig.json           # TypeScript configuration
```

## Features Implemented

### Dashboard Page

- Real-time system status
- Connection status for SQL Server, MQTT, MongoDB
- Worker control buttons (Start, Stop, Reset Watermark)
- Uptime statistics
- Ingestion rate monitoring

### Events Page

- **List Events**: Paginated, filterable event list
  - Filter by source, unit name, date range
  - Configurable page size (10-250 items)
  - Sort by timestamp or ingestion date
- **View Details**: Modal with complete event payload
- **Delete Event**: Individual event deletion
- **Batch Operations**: Ready for batch delete

### Configuration Page

- Tabbed interface for different config sections
- **SQL Server**: Host, port, database, user, password
- **MQTT**: Broker, port, topic, client ID, user, password, QoS
- **MongoDB**: URI, database, collection
- **Polling**: Interval and batch size
- Connection testing for each service
- Real-time save feedback

### Logs Page

- **Log Viewer**: Terminal-style log display
- **Filtering**: Filter by log level (ERROR, WARN, INFO, DEBUG)
- **Pagination**: Configurable page size (50-500 entries)
- **Color Coding**: Different colors for different log levels
- **Real-time Refresh**: Auto-refreshes every 3 seconds

## API Integration

The frontend communicates with the backend API using the `api.ts` client:

### Endpoints

#### Status & Worker

- `GET /api/status` - Get system status
- `POST /api/worker/start` - Start worker
- `POST /api/worker/stop` - Stop worker
- `POST /api/watermark/reset` - Reset watermark

#### Configuration

- `GET /api/config` - Get configuration
- `PUT /api/config` - Update configuration
- `POST /api/test/sqlserver` - Test SQL Server connection
- `POST /api/test/mqtt` - Test MQTT connection
- `POST /api/test/mongodb` - Test MongoDB connection

#### Events

- `GET /api/events` - List events with filters
  - Query params: `page`, `pageSize`, `source`, `unitName`, `startDate`, `endDate`, `sortBy`, `sortOrder`
- `GET /api/events/:id` - Get event by ID
- `PUT /api/events/:id` - Update event
- `DELETE /api/events/:id` - Delete event
- `DELETE /api/events/batch` - Batch delete events

#### Logs

- `GET /api/logs` - Get logs
  - Query params: `level`, `page`, `pageSize`

## Authentication

All API requests use HTTP Basic Authentication:

- Username: `admin`
- Password: `admin123` (change in `frontend/src/services/api.ts`)

To change credentials, edit the `api.ts` file:

```typescript
const client = axios.create({
  baseURL: '/api',
  auth: {
    username: 'admin',
    password: 'your-password-here',
  },
})
```

## Building for Production

```bash
npm run build
```

This creates an optimized build in the `dist/` folder. The backend will serve this automatically if the build files are in `backend/web/dist/`.

## Deployment

### With Docker

The Docker setup expects the frontend to be built and copied to `backend/web/dist/`:

```bash
# From project root
npm run build
mkdir -p backend/web
cp -r frontend/dist backend/web/

docker-compose up -d
```

### Manual Deployment

1. Build the frontend: `npm run build`
2. Copy `dist/` contents to `backend/web/dist/`
3. Start the backend server

The backend will serve the frontend at `http://localhost:8080/`

## Troubleshooting

### Node.js not found

Ensure Node.js is installed and in your system PATH.

### npm install fails

- Delete `node_modules` and `package-lock.json`
- Try again: `npm install`

### Port 5173 already in use

Vite will automatically use the next available port. Check the terminal output for the actual URL.

### API connection errors

- Ensure the backend is running on `http://localhost:8080`
- Check that credentials in `api.ts` are correct
- Verify CORS is enabled in the backend (it should be by default)

### Styles not applying

- Make sure Tailwind CSS is properly configured
- Rebuild: `npm run build`

## Development Tips

### Hot Module Replacement (HMR)

When running `npm run dev`, changes to files are reflected immediately in the browser without requiring a full refresh.

### TypeScript

The project uses TypeScript for type safety. Check `src/types/index.ts` for custom types.

### React Query

API data fetching is handled by React Query (@tanstack/react-query) which provides:

- Automatic caching
- Background refetching
- Stale data handling
- Error handling

### Tailwind CSS

Styling uses Tailwind CSS utility classes. See `tailwind.config.js` for customization.

## Performance Optimization

The frontend includes several optimizations:

- Code splitting via Vite
- Component lazy loading for pages
- React Query caching
- Automatic pagination to prevent loading too many items

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

Requires ES2020+ support for JavaScript features used.

---

For issues or questions, check the main project documentation in `../README.md` or `../CRUD_IMPLEMENTATION.md`.
