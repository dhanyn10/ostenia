package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// PortOwner contains details about the process listening on a port
type PortOwner struct {
	Port        int    `json:"port"`
	InUse       bool   `json:"inUse"`
	PID         int    `json:"pid"`
	ProcessName string `json:"processName"`
	ExecPath    string `json:"execPath"`
}

// GetAvailablePort finds the first available port from the given list of candidate ports
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
		return true // Port available
	}
	conn.Close()
	return false // Port currently in use
}

// GetPortOwner checks whether the specified port is in use and identifies its owner
func GetPortOwner(port int) (*PortOwner, error) {
	if isPortAvailable(port) {
		return &PortOwner{Port: port, InUse: false}, nil
	}

	pid, procName, execPath, err := findPortOwnerOS(port)
	if err != nil {
		return &PortOwner{Port: port, InUse: true, ProcessName: "Unknown"}, nil
	}

	return &PortOwner{
		Port:        port,
		InUse:       true,
		PID:         pid,
		ProcessName: procName,
		ExecPath:    execPath,
	}, nil
}

func parsePortOwnerFromNetstat(output string, targetPort int) (int, error) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[3] == "LISTENING" {
			localAddr := fields[1]
			lastColon := strings.LastIndex(localAddr, ":")
			if lastColon != -1 {
				p, err := strconv.Atoi(localAddr[lastColon+1:])
				if err == nil && p == targetPort {
					pid, err := strconv.Atoi(fields[len(fields)-1])
					if err == nil && pid > 0 {
						return pid, nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("no process found listening on port %d", targetPort)
}
