# Veranda

A productivity CLI tool for timers, stopwatches, and focus sessions with ambient sound support.

## Features

- **⏱️ Timers**: Create countdown timers with millisecond precision
- **⏱️ Stopwatches**: Track elapsed time with lap support
- **🎵 Ambient Sound**: Auto-playing background sounds during focus sessions
- **🖥️ TUI**: Beautiful terminal interface with live updates
- **🔔 Notifications**: Desktop notifications on timer completion
- **📝 Journal**: Log your thoughts and progress
- **🌐 HTTP API**: Daemon server for cross-process communication

## Installation

```bash
go install github.com/lightt77/veranda/cmd/veranda@latest
```

Or build from source:

```bash
git clone https://github.com/lightt77/veranda.git
cd veranda
go build -o veranda ./cmd/veranda
```

## Quick Start

### 1. Start the Daemon

The daemon handles all background operations:

```bash
veranda daemon
```

### 2. Create a Timer

Quick syntax - creates and starts immediately:

```bash
veranda timer 25m "Focus Time"
```

### 3. Launch TUI

Interactive terminal interface:

```bash
veranda tui
```

## Commands

### Daemon

```bash
veranda daemon              # Start the daemon server
```

### Timers

```bash
veranda timer 25m "Pomodoro"     # Create and start 25-minute timer
veranda timer 1h "Deep Work"     # Create 1-hour timer
veranda timer 90s "Quick Break"  # Create 90-second timer

veranda timer list               # List all timers
veranda timer start [id]         # Start a timer
veranda timer pause [id]         # Pause a timer
veranda timer stop [id]          # Stop a timer (alias for pause)
veranda timer delete [id]        # Delete a timer
```

### Stopwatches

```bash
veranda stopwatch start "Workout"   # Start a stopwatch
veranda stopwatch list              # List all stopwatches
veranda stopwatch stop [id]         # Stop a stopwatch
veranda stopwatch lap [id]          # Record a lap
veranda stopwatch reset [id]        # Reset a stopwatch
veranda stopwatch delete [id]       # Delete a stopwatch
```

### Ambient Sound

```bash
veranda ambient play          # Start ambient sound playback
veranda ambient stop          # Stop ambient sound playback
veranda ambient status        # Check ambient sound status
veranda ambient volume 0.5    # Set volume to 50%
```

### Journal

```bash
veranda journal write "Completed the project!"   # Write entry
veranda journal list                             # List entries
veranda journal recent 5                         # Show 5 recent entries
```

### TUI

```bash
veranda tui                   # Launch interactive TUI
```

## TUI Controls

```
tab / →         Switch tabs (Timers / Stopwatches)
shift+tab / ←   Switch tabs reverse
t               Create quick timer (Timers tab)
s               Create quick stopwatch (Stopwatches tab)
r               Refresh data
q / ctrl+c      Quit
```

## Configuration

### Directories

- **Data**: `~/.veranda/data/` - SQLite database
- **Logs**: `~/.veranda/logs/` - Log files
- **Sounds**: `~/.veranda/sounds/` - Ambient sound MP3 files

### Ambient Sounds

Add MP3 files to `~/.veranda/sounds/` for ambient playback:

```bash
mkdir -p ~/.veranda/sounds
cp your-sounds/*.mp3 ~/.veranda/sounds/
```

Ambient sound automatically:
- Starts when you start a timer or stopwatch
- Stops when all timers/stopwatches are stopped
- Shuffles through all MP3 files
- Fades in (3s) and out (5s)

## HTTP API

The daemon exposes a REST API on `http://localhost:17342`:

### Health

```bash
curl http://localhost:17342/health
```

### Timers

```bash
# List timers
curl http://localhost:17342/timers

# Create timer
curl -X POST http://localhost:17342/timers \
  -H "Content-Type: application/json" \
  -d '{"label":"Test","duration_ms":1500000}'

# Start timer
curl -X POST http://localhost:17342/timers/1234567890/start
```

### Stopwatches

```bash
# List stopwatches
curl http://localhost:17342/stopwatches

# Create and start stopwatch
curl -X POST http://localhost:17342/stopwatches \
  -H "Content-Type: application/json" \
  -d '{"label":"Workout","start":true}'
```

### Settings

```bash
# Get all settings
curl http://localhost:17342/settings

# Set setting
curl -X PUT http://localhost:17342/settings/my_setting \
  -H "Content-Type: application/json" \
  -d '{"value":"my_value"}'
```

See `internal/daemon/handlers.go` for all endpoints.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │  timer  │ │stopwatch│ │ ambient │ │ journal │          │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘          │
│       └───────────┴───────────┴───────────┘                │
│                          │                                  │
│                    ┌─────┴─────┐                           │
│                    │   tui     │                           │
│                    └─────┬─────┘                           │
└──────────────────────────┼──────────────────────────────────┘
                           │
                    ┌──────┴──────┐
                    │Daemon Client│
                    └──────┬──────┘
                           │ HTTP
                    ┌──────┴──────┐
                    │Daemon Server│
                    └──────┬──────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
  ┌────┴────┐      ┌──────┴──────┐     ┌──────┴──────┐
  │ Services│      │  Ambient    │     │Notification │
  └────┬────┘      │   Audio     │     │   Service   │
       │           └─────────────┘     └─────────────┘
  ┌────┴────┐
  │Repository│
  └────┬────┘
       │
  ┌────┴────┐
  │ SQLite  │
  └─────────┘
```

## Project Structure

```
veranda/
├── cmd/
│   ├── commands/          # Cobra CLI commands
│   │   ├── root.go        # Root command
│   │   ├── timer.go       # Timer commands
│   │   ├── stopwatch.go   # Stopwatch commands
│   │   ├── daemon.go      # Daemon command
│   │   ├── ambient.go     # Ambient sound commands
│   │   ├── journal.go     # Journal commands
│   │   └── tui.go         # TUI command
│   └── veranda/
│       └── main.go        # Entry point
├── internal/
│   ├── audio/             # Ambient sound player
│   ├── config/            # Configuration
│   ├── daemon/            # HTTP server and client
│   ├── db/                # Database layer
│   ├── models/            # Domain models
│   ├── notify/            # Desktop notifications
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic
│   └── tui/               # Bubble Tea TUI
└── pkg/
    └── api/               # Public API types
```

## Development

### Run Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test ./... -cover
```

### Build

```bash
go build -o veranda ./cmd/veranda
```

## Dependencies

- **SQLite**: `modernc.org/sqlite` (pure Go)
- **Audio**: `github.com/gopxl/beep/v2`
- **TUI**: `github.com/charmbracelet/bubbletea`
- **HTTP**: `github.com/go-chi/chi/v5`
- **CLI**: `github.com/spf13/cobra`
- **Notifications**: `github.com/gen2brain/beeep`

## Platform Support

- ✅ macOS
- ✅ Linux
- ✅ Windows

Notifications and ambient sound work on all supported platforms.

## License

MIT

## Roadmap

- [x] Timers and stopwatches
- [x] Ambient sound with fade in/out
- [x] Desktop notifications
- [x] HTTP API
- [x] TUI with live updates
- [ ] Scheduled tasks (cron-like)
- [ ] Statistics and reports
- [ ] Config file support
- [ ] Plugin system
