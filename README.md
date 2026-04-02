# Veranda

A lightweight, beautiful productivity app with timers, stopwatches, and todos. Built with Rust and Tauri.

![Veranda Icon](src-tauri/icons/icon-128x128.png)

## Features

- **Timers**: Create countdown timers with custom durations and labels
- **Stopwatches**: Track elapsed time with lap recording
- **Todos**: Manage tasks with projects, labels, and status tracking
- **Background Art**: Set custom background images from Pinterest URLs
- **Notifications**: Get notified when timers complete (with sound)
- **CLI Interface**: Control Veranda from the command line
- **Background Daemon**: Timers and stopwatches continue running even when the GUI is closed

## Installation

### macOS

1. Download the latest `.dmg` from [Releases](../../releases)
2. Open the DMG and drag Veranda to your Applications folder
3. Launch Veranda from Applications

### From Source

Requirements:
- Rust 1.77.2 or later
- Node.js 18 or later
- macOS 10.13+ (for macOS builds)

```bash
# Clone the repository
git clone https://github.com/yourusername/veranda.git
cd veranda

# Install dependencies
npm install

# Build the frontend
npm run build

# Build the app
cd src-tauri
cargo build --release
```

The built app will be at `src-tauri/target/release/veranda`.

## Usage

### GUI

Launch Veranda to open the main window with three tabs:
- **Timers**: Create and manage countdown timers
- **Stopwatches**: Track time and record laps
- **Todos**: Manage your tasks

Keyboard shortcuts:
- `Cmd+T`: Create new timer
- `Cmd+S`: Create new stopwatch
- `Cmd+D`: Create new todo

### CLI

Veranda can be controlled from the command line:

```bash
# List all timers
./veranda timer list

# Create a new timer
./veranda timer create -l "Pomodoro" -d 1500

# Start a timer
./veranda timer start -i <timer-id>

# Pause a timer
./veranda timer pause -i <timer-id>

# List stopwatches
./veranda stopwatch list

# Create and start a stopwatch
./veranda stopwatch create -l "Work Session"
./veranda stopwatch start -i <stopwatch-id>

# Manage todos
./veranda todo list
./veranda todo add -t "Buy groceries"
./veranda todo done -i <todo-id>

# Run daemon only (no GUI)
./veranda daemon
```

### Setting up CLI Alias

Add to your `.zshrc` or `.bashrc`:

```bash
alias veranda="/Applications/Veranda.app/Contents/MacOS/veranda"
```

Then reload your shell:

```bash
source ~/.zshrc  # or ~/.bashrc
```

## Data Storage

All data is stored in SQLite at `~/veranda-data/`:
- `veranda.db`: Main database with timers, stopwatches, and todos
- `images/`: Downloaded background images

This location is configurable by modifying `DATA_DIR` in `src-tauri/src/config.rs`.

## Development

### Project Structure

```
veranda/
├── src-ui/              # React frontend
│   ├── components/      # UI components
│   ├── stores/         # Zustand state management
│   └── hooks/          # Custom React hooks
├── src-tauri/          # Rust backend
│   ├── src/
│   │   ├── background/ # Background service daemon
│   │   ├── db/         # Database operations
│   │   ├── models/     # Data models
│   │   └── services/   # Image and sound services
│   └── icons/          # App icons
└── scripts/            # Build scripts
```

### Running in Development Mode

```bash
# Terminal 1: Start the dev server
npm run dev

# Terminal 2: Run the Tauri app
cd src-tauri
cargo tauri dev
```

### Building for Production

```bash
# Build frontend
npm run build

# Build the app bundle
cd src-tauri
cargo tauri build
```

The bundled app will be in `src-tauri/target/release/bundle/`.

## Technologies

- **Frontend**: React + TypeScript + Tailwind CSS + Zustand
- **Backend**: Rust + Tauri v2
- **Database**: SQLite via rusqlite
- **Notifications**: macOS Notification Center
- **Audio**: macOS System Sounds (via afplay)

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Tauri](https://tauri.app/)
- Icons generated with custom tooling
- Inspired by minimalist productivity apps
