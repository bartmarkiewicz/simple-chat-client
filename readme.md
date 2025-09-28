# Simple WebSocket Chat

A minimal real-time chat application with:
- Backend: Go 1.24 (WebSocket server)
- Frontend: TypeScript + Vue 3 (client UI)
- Dev tooling: Vite, ESLint, Prettier
- Tests: Playwright (e2e), Golang (unit)

## Features
- Realtime messaging via WebSockets
- Basic global chat room
- Join/leave notifications
- Lightweight

## Requirements
- Go 1.24+
- Node.js 22+ and npm
- Ports available: 8080 (backend), 5173 (frontend dev build server)

## Getting Started

### 1) Backend (Go)
- Run:
    - go mod
    - go run .
- The WebSocket endpoint will be available at:
    - ws://localhost:8080/web-socket

### 2) Frontend (Vue + TS)
- Install:
    - npm install
- Run:
    - npm run dev
- Open:
    - http://localhost:5173

## Production Build
- Frontend:
    - npm run build
    - Static output in dist/
- Backend options:
    - Serve the frontend separately via any static server/proxy
    - Or proxy requests from a reverse proxy (eg Nginx) to:
        - / -> frontend static host
        - /web-socket -> Go server at :8080

## Future Things

- Multiple rooms and private messaging
- Message persistence
- Auth and indicator of online users
- UI refinements
- Configurability of the web server and hardcoded ports
- Dockerfile

## Contributing

- Fork, create a feature branch, open PR.
- Keep code formatted (ESLint/Prettier) and tests passing.

## Contact

For issues or feature requests, open an issue in the repository.