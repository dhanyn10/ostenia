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
