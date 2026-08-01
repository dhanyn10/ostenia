package service

import (
	"fmt"
	"io"
	"ostenia/internal/backend/interfaces"
	"strings"
	"unicode/utf16"
)

// GetResourceUsage gathers real-time CPU, RAM, and Disk metrics from remote machines.
// Spins up a lightweight concurrent session executing combined command strings so interactive terminal
// performance is completely unaffected.
func (m *SSHManager) GetResourceUsage(sessionID string) (interfaces.ResourceUsage, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return interfaces.ResourceUsage{}, fmt.Errorf("session not found")
	}

	sess, err := conn.Client.NewSession()
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	// Combined commands to parse core system stats
	command := `cat /proc/stat; sleep 0.1; cat /proc/stat; cat /proc/meminfo; df -m /`

	err = sess.Run(command)
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	return parseResourceUsage(decodeMaybeUTF16(outputBytes)), nil
}

// decodeMaybeUTF16 dynamically converts UTF-16LE binary payloads into standard UTF-8 strings.
func decodeMaybeUTF16(b []byte) string {
	if len(b) < 2 {
		return string(b)
	}

	isUTF16 := false
	if b[0] == 0xFF && b[1] == 0xFE {
		isUTF16 = true
	} else if len(b) >= 4 && b[1] == 0x00 && b[3] == 0x00 {
		isUTF16 = true
	}

	if !isUTF16 {
		return string(b)
	}

	startIdx := 0
	if b[0] == 0xFF && b[1] == 0xFE {
		startIdx = 2
	}

	bytesToDecode := b[startIdx:]
	if len(bytesToDecode)%2 != 0 {
		bytesToDecode = bytesToDecode[:len(bytesToDecode)-1]
	}

	u16 := make([]uint16, len(bytesToDecode)/2)
	for i := 0; i < len(u16); i++ {
		u16[i] = uint16(bytesToDecode[2*i]) | (uint16(bytesToDecode[2*i+1]) << 8)
	}
	return string(utf16.Decode(u16))
}

// parseResourceUsage processes standard Unix '/proc/stat', '/proc/meminfo', and 'df' command dumps.
func parseResourceUsage(output string) interfaces.ResourceUsage {
	var usage interfaces.ResourceUsage
	lines := strings.Split(output, "\n")

	var cpuTicks [][]float64
	var memTotal, memAvailable, memFree, buffers, cached float64
	var diskTotal, diskUsed float64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		// Parse CPU stats from consecutive /proc/stat checks
		if strings.HasPrefix(line, "cpu ") && len(fields) >= 5 {
			var user, nice, system, idle, iowait, irq, softirq, steal float64
			fmt.Sscanf(fields[1], "%f", &user)
			fmt.Sscanf(fields[2], "%f", &nice)
			fmt.Sscanf(fields[3], "%f", &system)
			fmt.Sscanf(fields[4], "%f", &idle)
			if len(fields) >= 6 {
				fmt.Sscanf(fields[5], "%f", &iowait)
			}
			if len(fields) >= 7 {
				fmt.Sscanf(fields[6], "%f", &irq)
			}
			if len(fields) >= 8 {
				fmt.Sscanf(fields[7], "%f", &softirq)
			}
			if len(fields) >= 9 {
				fmt.Sscanf(fields[8], "%f", &steal)
			}

			total := user + nice + system + idle + iowait + irq + softirq + steal
			idleVal := idle + iowait
			cpuTicks = append(cpuTicks, []float64{total, idleVal})
		}

		// Parse memory thresholds
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %f kB", &memTotal)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %f kB", &memAvailable)
		} else if strings.HasPrefix(line, "MemFree:") {
			fmt.Sscanf(line, "MemFree: %f kB", &memFree)
		} else if strings.HasPrefix(line, "Buffers:") {
			fmt.Sscanf(line, "Buffers: %f kB", &buffers)
		} else if strings.HasPrefix(line, "Cached:") {
			fmt.Sscanf(line, "Cached: %f kB", &cached)
		}

		// Parse mount file systems ('df')
		if len(fields) >= 5 && fields[len(fields)-1] == "/" {
			if len(fields) >= 6 {
				var tot, usd float64
				fmt.Sscanf(fields[1], "%f", &tot)
				fmt.Sscanf(fields[2], "%f", &usd)
				if tot > 0 {
					diskTotal = tot
					diskUsed = usd
				}
			} else if len(fields) == 5 {
				var tot, usd float64
				fmt.Sscanf(fields[0], "%f", &tot)
				fmt.Sscanf(fields[1], "%f", &usd)
				if tot > 0 {
					diskTotal = tot
					diskUsed = usd
				}
			}
		}
	}

	// Calculate overall CPU utilization percentage
	if len(cpuTicks) >= 2 {
		diffTotal := cpuTicks[1][0] - cpuTicks[0][0]
		diffIdle := cpuTicks[1][1] - cpuTicks[0][1]
		if diffTotal > 0 {
			usage.CPU = ((diffTotal - diffIdle) / diffTotal) * 100
		}
	}

	// Calculate RAM utilization percentages
	if memTotal > 0 {
		usage.MemTotal = memTotal / 1024
		var used float64
		if memAvailable > 0 {
			used = memTotal - memAvailable
		} else {
			used = memTotal - memFree - buffers - cached
		}
		usage.MemUsed = used / 1024
		usage.Mem = (used / memTotal) * 100
	}

	// Calculate overall disk usage percentage
	if diskTotal > 0 {
		usage.DiskTotal = diskTotal
		usage.DiskUsed = diskUsed
		usage.Disk = (diskUsed / diskTotal) * 100
	}

	return usage
}
