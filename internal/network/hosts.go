package network

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var hostsPathOverride string

func GetHostsPath() string {
	if hostsPathOverride != "" {
		return hostsPathOverride
	}
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

func AddHost(ip string, hostname string) error {
	hostsPath := GetHostsPath()

	// Check if already exists
	f, err := os.Open(hostsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tag := "#WailsManaged"
	lineToAdd := fmt.Sprintf("%s  %s %s", ip, hostname, tag)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), hostname) {
			return nil // Already exists
		}
	}

	// Append to file
	f, err = os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("\n" + lineToAdd)
	return err
}

func RemoveManagedHosts() error {
	hostsPath := GetHostsPath()
	f, err := os.Open(hostsPath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if !strings.Contains(scanner.Text(), "#WailsManaged") {
			lines = append(lines, scanner.Text())
		}
	}
	f.Close()

	return os.WriteFile(hostsPath, []byte(strings.Join(lines, "\n")), 0644)
}
