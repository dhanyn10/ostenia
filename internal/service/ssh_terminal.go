package service

import (
	"context"
	"fmt"
	"io"
	"ostenia/internal/backend/interfaces"

	"golang.org/x/crypto/ssh"
)

// startTerminal creates an interactive shell session, requests a virtual PTY terminal,
// hooks up stdin/stdout piping, and starts the background stdout listener.
func (m *SSHManager) startTerminal(ctx context.Context, conn *SSHConnection) {
	fmt.Printf("[SSH] Requesting new SSH session for terminal (SessionID: %s)...\n", conn.SessionID)
	sshSession, err := conn.Client.NewSession()
	if err != nil {
		fmt.Printf("[SSH] Failed to create SSH session: %v\n", err)
		return
	}
	m.mu.Lock()
	conn.PTY = sshSession
	m.mu.Unlock()

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		fmt.Printf("[SSH] Failed to get stdout pipe: %v\n", err)
		return
	}
	stdin, err := sshSession.StdinPipe()
	if err != nil {
		fmt.Printf("[SSH] Failed to get stdin pipe: %v\n", err)
		return
	}
	m.mu.Lock()
	conn.Shell = stdin
	m.mu.Unlock()

	if err := m.setupPTY(sshSession); err != nil {
		fmt.Printf("[SSH] Failed to setup terminal PTY: %v\n", err)
		return
	}

	// To prevent initial WSL interactive shell hangs or connection delays, write an initial newline.
	if conn.IsWSL && conn.Shell != nil {
		_, _ = conn.Shell.Write([]byte("\n"))
	}

	exitChan := make(chan struct{})
	// Read and broadcast stdout outputs to the Wails frontend.
	go m.processTerminalOutput(ctx, conn, stdout, exitChan)
	// Monitor terminal termination or disconnect events.
	go m.handleTerminalExit(ctx, conn, sshSession, exitChan)
}

// setupPTY requests a standard pseudo-terminal (PTY) wrapper with basic echo/canonical rules configured.
func (m *SSHManager) setupPTY(sshSession interfaces.SSHSession) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // Enable echoing of characters
		ssh.TTY_OP_ISPEED: 14400, // Input speed
		ssh.TTY_OP_OSPEED: 14400, // Output speed
		ssh.ICANON:        0,     // Disable canonical mode to handle input byte-by-byte
	}

	if err := sshSession.RequestPty("xterm-256color", 50, 200, modes); err != nil {
		return err
	}

	return sshSession.Shell()
}

// processTerminalOutput continuously reads from the shell stdout pipe and broadcasts content
// directly to the frontend React view via Wails event emissions.
func (m *SSHManager) processTerminalOutput(ctx context.Context, conn *SSHConnection, stdout io.Reader, exitChan chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			if m.runtime != nil {
				// Emit SSH chunk output to corresponding session terminal
				m.runtime.EventsEmit(ctx, "ssh_output", map[string]interface{}{
					"sessionId": conn.SessionID,
					"data":      string(buf[:n]),
				})
			}
		}
		if err != nil {
			// Trigger exit signaling when read loop encounters EOF or socket disconnect
			select {
			case <-exitChan:
			default:
				close(exitChan)
			}
			break
		}
	}
}

// handleTerminalExit handles graceful session cleanup upon termination, context cancellation,
// or unexpected connection dropouts.
func (m *SSHManager) handleTerminalExit(ctx context.Context, conn *SSHConnection, sshSession interfaces.SSHSession, exitChan chan struct{}) {
	select {
	case <-ctx.Done():
		fmt.Printf("[SSH] Context done, closing terminal session %s\n", conn.SessionID)
		sshSession.Close()
	case <-exitChan:
		fmt.Printf("[SSH] Terminal output channel closed, disconnecting terminal session %s\n", conn.SessionID)
		if m.runtime != nil {
			m.runtime.EventsEmit(ctx, "ssh_disconnected", conn.SessionID)
		}
		m.mu.Lock()
		delete(m.connections, conn.SessionID)
		m.mu.Unlock()
	}
}

// ResizeTerminal updates the width/height dimensions of the remote PTY terminal session.
// Enforces hard guards to maintain terminal stability.
func (m *SSHManager) ResizeTerminal(sessionID string, cols, rows int) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.PTY == nil {
		return fmt.Errorf("session not connected")
	}

	// Logging to debug dimension issues in the sandbox.
	fmt.Printf("[SSH] Resizing session %s to Cols: %d, Rows: %d\n", sessionID, cols, rows)

	// HARD GUARD: Never allow dimensions to drop below minimal levels to prevent erratic wrapping.
	if cols < 100 {
		cols = 100
	}
	if rows < 10 {
		rows = 10
	}

	return conn.PTY.WindowChange(rows, cols)
}
