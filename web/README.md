# Miners Scheduler Web UI

A minimal React TypeScript web application for monitoring and managing the Avalon miners scheduler.

## Features

- Real-time health monitoring
- Scheduler status display
- Price information and action recommendations
- Discovered miners list
- System information
- Auto-refresh every 10 seconds

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

## Production Build

Build the application for production:

```bash
npm run build
```

The built files will be in the `dist` directory, which the Go backend serves at the root path.

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
- `GET /api/status` - Detailed status with miners and price data

## Technologies

- **React 18** - UI library
- **TypeScript 5** - Type-safe JavaScript
- **Vite 5** - Fast build tool and dev server
- **CSS Variables** - Modern styling approach

## Configuration

The Vite configuration includes:

- Development server on port 3000
- API proxy to backend on port 8080
- Production build output to `dist/`
- Source maps enabled

## Customization

### Changing the API URL

Edit `vite.config.ts` to change the backend API URL:

```typescript
proxy: {
  '/api': {
    target: 'http://your-backend:port',
    changeOrigin: true
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
  /* ... */
}
```

## License

Same as parent project