package network

import (
	"net"
	"strconv"
	"testing"
)

func TestGetAvailablePort(t *testing.T) {
	// Temukan port yang benar-benar terbuka
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	usedPort, _ := strconv.Atoi(portStr)

	// Test GetAvailablePort dengan usedPort (harus gagal jika hanya usedPort yang diberikan)
	_, err = GetAvailablePort([]int{usedPort})
	if err == nil {
		t.Errorf("Expected error for used port %d, got nil", usedPort)
	}

	// Test GetAvailablePort dengan port yang kemungkinan besar tersedia
	// Kita cari port lain yang tersedia
	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	_, portStr2, _ := net.SplitHostPort(ln2.Addr().String())
	freePort, _ := strconv.Atoi(portStr2)
	ln2.Close() // Sekarang freePort tersedia

	gotPort, err := GetAvailablePort([]int{freePort})
	if err != nil {
		t.Errorf("Expected no error for free port %d, got %v", freePort, err)
	}
	if gotPort != freePort {
		t.Errorf("Expected port %d, got %d", freePort, gotPort)
	}
}
