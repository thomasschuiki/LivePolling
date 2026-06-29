# LivePolling

A real-time polling application using WebSockets.

## Quick Start

```bash
go run server.go
```

Open http://localhost:8080/ for the voter client or http://localhost:8080/admin/ for the admin interface.

Default admin credentials: `admin` / `admin`

## Features

- Real-time polling with WebSocket communication
- Admin interface for creating questions and viewing live statistics
- Mobile-responsive voter client
- Vote tracking (prevents double voting)

## Tech Stack

- Go server with `nhooyr.io/websocket`
- Vanilla JavaScript frontend
- No database (in-memory state)
