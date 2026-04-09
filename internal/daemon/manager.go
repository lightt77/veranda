// Package daemon provides utilities for managing the daemon process.
package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Manager handles daemon lifecycle (auto-start, monitoring)
type Manager struct {
	port      int
	client    *Client
	isRunning bool
}

// NewManager creates a new daemon manager
func NewManager(port int) *Manager {
	return &Manager{
		port:   port,
		client: NewClient(port),
	}
}

// EnsureRunning checks if daemon is running, starts it if not
// Returns nil if daemon is now running (was already running or started successfully)
func (m *Manager) EnsureRunning() error {
	if m.client.IsRunning() {
		return nil
	}

	// Daemon not running, start it
	return m.startDaemon()
}

// startDaemon starts the daemon process in the background
func (m *Manager) startDaemon() error {
	// Get the executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start daemon in background
	cmd := exec.Command(exePath, "daemon")

	// Detach from parent
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start but don't wait
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Wait a moment for daemon to initialize
	time.Sleep(800 * time.Millisecond)

	// Verify it started
	if !m.client.IsRunning() {
		return fmt.Errorf("daemon failed to start")
	}

	return nil
}

// IsRunning returns whether the daemon is running
func (m *Manager) IsRunning() bool {
	return m.client.IsRunning()
}
