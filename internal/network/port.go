package network

import (
	"fmt"
	"net"
	"time"
)

// GetAvailablePort mencari port pertama yang tersedia dari daftar port yang diberikan
func GetAvailablePort(ports []int) (int, error) {
	for _, port := range ports {
		if isPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found in range")
}

func isPortAvailable(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return true // Port tersedia
	}
	conn.Close()
	return false // Port sedang digunakan
}
