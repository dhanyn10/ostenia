package network

import (
	"net"
	"strconv"
	"testing"
)

func TestGetAvailablePort(t *testing.T) {
	t.Run("Find available port", func(t *testing.T) {
		// Listen on a random port to make it unavailable
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to listen: %v", err)
		}
		defer ln.Close()

		_, portStr, _ := net.SplitHostPort(ln.Addr().String())
		usedPort, _ := strconv.Atoi(portStr)

		// Try to find available port from list [usedPort, someOtherPort]
		ports := []int{usedPort, usedPort + 1}
		got, err := GetAvailablePort(ports)
		if err != nil {
			t.Fatalf("GetAvailablePort() error = %v", err)
		}

		if got == usedPort {
			t.Errorf("GetAvailablePort() returned used port %v", got)
		}
	})

	t.Run("No available ports", func(t *testing.T) {
		ln, _ := net.Listen("tcp", "127.0.0.1:0")
		defer ln.Close()
		_, portStr, _ := net.SplitHostPort(ln.Addr().String())
		usedPort, _ := strconv.Atoi(portStr)

		ports := []int{usedPort}
		_, err := GetAvailablePort(ports)
		if err == nil {
			t.Errorf("GetAvailablePort() expected error when no ports available")
		}
	})
}

func TestGetPortOwner_AvailablePort(t *testing.T) {
	// Pick an available port by asking the OS for one, then closing it
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	ln.Close()

	owner, err := GetPortOwner(port)
	if err != nil {
		t.Fatalf("GetPortOwner() error = %v", err)
	}
	if owner.InUse {
		t.Errorf("Expected port %d to not be in use", port)
	}
}

func TestGetPortOwner_OccupiedPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	owner, err := GetPortOwner(port)
	if err != nil {
		t.Fatalf("GetPortOwner() error = %v", err)
	}
	if !owner.InUse {
		t.Errorf("Expected port %d to be marked in use", port)
	}
}

func TestParsePortOwnerFromNetstat(t *testing.T) {
	sampleNetstat := `
  Proto  Local Address          Foreign Address        State           PID
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1840
  TCP    0.0.0.0:3306           0.0.0.0:0              LISTENING       16912
  TCP    0.0.0.0:33060          0.0.0.0:0              LISTENING       9999
  TCP    [::]:80                [::]:0                 LISTENING       4
  TCP    127.0.0.1:3306         127.0.0.1:50000        ESTABLISHED     16912
`

	t.Run("Match port 3306", func(t *testing.T) {
		pid, err := parsePortOwnerFromNetstat(sampleNetstat, 3306)
		if err != nil {
			t.Fatalf("Expected match, got error: %v", err)
		}
		if pid != 16912 {
			t.Errorf("Expected PID 16912, got %d", pid)
		}
	})

	t.Run("Match port 33060 (no prefix collision with 3306)", func(t *testing.T) {
		pid, err := parsePortOwnerFromNetstat(sampleNetstat, 33060)
		if err != nil {
			t.Fatalf("Expected match, got error: %v", err)
		}
		if pid != 9999 {
			t.Errorf("Expected PID 9999, got %d", pid)
		}
	})

	t.Run("Match IPv6 port 80", func(t *testing.T) {
		pid, err := parsePortOwnerFromNetstat(sampleNetstat, 80)
		if err != nil {
			t.Fatalf("Expected match, got error: %v", err)
		}
		if pid != 4 {
			t.Errorf("Expected PID 4, got %d", pid)
		}
	})

	t.Run("Port not listening", func(t *testing.T) {
		_, err := parsePortOwnerFromNetstat(sampleNetstat, 9000)
		if err == nil {
			t.Errorf("Expected error for non-listening port, got nil")
		}
	})
}
