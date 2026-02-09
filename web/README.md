# Energy Management System Web UI

A minimal React TypeScript web application for monitoring and managing the Energy Management System (EMS).

## Features

- Real-time energy flow monitoring
- Solar (PV) production tracking
- Battery state of charge (SOC) and power flow
- Grid import/export visualization
- Electricity price information and trends
- Controllable load status and management
- System health monitoring
- Manual device control
- Historical data charts
- Auto-refresh with WebSocket support

## Prerequisites

- Node.js 18+ or npm/yarn
- The Go backend server running on port 8080

## Installation

```bash
npm install
```

## Development

Run the development server with hot reload:

```bash
npm run dev
```

The application will be available at `http://localhost:3000`. API requests will be proxied to the backend at `http://localhost:8080`.

### Demo Mode

For development or demonstration purposes without a backend server, you can run the application in demo mode. This mode uses realistic mock data that simulates all backend functionality:

```bash
npm run dev:demo
```

In demo mode:
- No backend connection is required
- Realistic simulated data updates every 10 seconds
- Solar power, battery, grid, and device data reflect realistic daily patterns
- Electricity prices vary based on time of day
- MPC decisions are generated for the next 24 hours in 15-minute intervals (96 decisions)
- All UI components work exactly as they would with a real backend

## Production Build

Build the application for production:

```bash
npm run build
```

The built files will be in the `dist` directory, which the Go backend serves at the root path.

### Demo Build

Build a standalone demo version that doesn't require a backend:

```bash
npm run build:demo
```

The demo build will be in the `dist-demo` directory. You can preview it with:

```bash
npm run preview:demo
```

The demo build is useful for:
- Demonstrations and presentations
- UI/UX testing without backend setup
- Showcasing the application to stakeholders
- Development when backend is unavailable
- Static hosting scenarios

## Project Structure

```
web/
├── src/
│   ├── App.tsx          # Main application component
│   ├── App.css          # Application styles
│   ├── main.tsx         # React entry point
│   ├── index.css        # Global styles
│   └── vite-env.d.ts    # Vite type definitions
├── index.html           # HTML template
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
├── tsconfig.node.json   # TypeScript config for Vite
└── vite.config.ts       # Vite configuration
```

## API Endpoints

The web application consumes the following API endpoints:

- `GET /api/health` - Health check endpoint
- `GET /api/ready` - Readiness check endpoint
- `GET /api/status` - Detailed system status with energy flows, prices, and device data
- `GET /api/pv` - Current PV production and battery state
- `GET /api/devices` - List of controllable devices
- `POST /api/device/:id/control` - Manual device control
- `WebSocket /ws` - Real-time updates

## Dashboard Sections

### Energy Overview
- **Solar Production**: Real-time PV power generation
- **Battery Status**: Current SOC, charge/discharge power
- **Grid Status**: Import/export power and daily totals
- **Total Load**: Current power consumption

### Price Information
- **Current Price**: Real-time electricity market price
- **Price Limit**: Configured activation threshold
- **Action Recommendation**: Buy/sell/hold based on price
- **Price Forecast**: Next 24-hour price trend

### Device Management
- **Device List**: All discovered controllable loads
- **Status**: Current state (active/standby) and mode
- **Power Consumption**: Real-time and cumulative
- **Manual Control**: Override automatic management
- **Thermal Status**: Temperature and fan speed monitoring

### System Information
- **EMS Status**: Overall system health
- **Last Update**: Timestamp of latest data
- **Optimization**: MPC schedule status
- **Alerts**: Any warnings or errors

## Technologies

- **React 18** - UI library
- **TypeScript 5** - Type-safe JavaScript
- **Vite 5** - Fast build tool and dev server
- **Chart.js** (optional) - Data visualization
- **CSS Variables** - Modern styling approach

## Configuration

The Vite configuration includes:

- Development server on port 3000
- API proxy to backend on port 8080 (disabled in demo mode)
- WebSocket proxy for real-time updates (mock data in demo mode)
- Production build output to `dist/` (or `dist-demo/` for demo builds)
- Source maps enabled
- Conditional compilation for demo mode via `__DEMO_MODE__` flag

## Customization

### Changing the API URL

Edit `vite.config.ts` to change the backend API URL:

```typescript
proxy: {
  '/api': {
    target: 'http://your-backend:port',
    changeOrigin: true
  },
  '/ws': {
    target: 'ws://your-backend:port',
    ws: true
  }
}
```

### Styling

The application uses CSS variables defined in `App.css`. Modify the `:root` selector to customize colors:

```css
:root {
  --color-primary: #2563eb;
  --color-success: #16a34a;
  --color-warning: #ea580c;
  --color-error: #dc2626;
  --color-solar: #f59e0b;
  --color-battery: #10b981;
  --color-grid: #6366f1;
  /* ... */
}
```

### Refresh Interval

By default, the dashboard refreshes every 10 seconds. To change this, modify the interval in `App.tsx`:

```typescript
const REFRESH_INTERVAL = 10000; // milliseconds
```

## Features in Detail

### Real-time Monitoring
The dashboard connects to the backend via WebSocket for real-time updates of:
- Energy production and consumption
- Battery charge/discharge
- Grid import/export
- Device status changes
- Price updates

### Manual Control
Users can manually override automatic control:
- Activate/deactivate devices
- Change device operating modes (Eco/Standard/Super)
- Force battery charge/discharge
- Set custom price limits

### Data Visualization
Historical data can be visualized with charts showing:
- Energy production vs consumption over time
- Battery SOC trends
- Price fluctuations
- Cost savings

## Troubleshooting

### Cannot connect to backend
- Verify the backend is running on port 8080
- Check firewall settings
- Ensure CORS is properly configured

### No real-time updates
- Check WebSocket connection in browser console
- Verify WebSocket support in your environment
- Check proxy configuration in `vite.config.ts`

### Build errors
- Ensure Node.js version is 18 or higher
- Clear `node_modules` and reinstall: `rm -rf node_modules && npm install`
- Check for TypeScript errors: `npm run type-check`

## Development Tips

- Use browser developer tools to monitor API calls
- Enable React DevTools for component inspection
- Check console for WebSocket connection status
- Use network tab to debug API issues
- Use demo mode for frontend-only development
- Mock data generator can be customized in `src/utils/mockData.ts`

## License

Same as parent project (Energy Management System)